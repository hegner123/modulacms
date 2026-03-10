// drivergen generates MySQL and PostgreSQL wrapper methods from the SQLite
// (canonical) version in internal/db/*_custom.go files.
//
// Classification is receiver-based, not marker-based:
//
//   - func (d Database) ...            → SQLite (source of truth, replicated)
//   - func (d MysqlDatabase) ...       → MySQL  (dropped, regenerated)
//   - func (d PsqlDatabase) ...        → PostgreSQL (dropped, regenerated)
//   - type FooCmdMysql struct ...       → MySQL command type (preserved)
//   - type FooCmdPsql struct ...        → PostgreSQL command type (preserved)
//   - type FooCmd struct (no suffix)    → SQLite command type (preserved)
//   - func (c FooCmdMysql) ...          → MySQL command method (preserved)
//   - func (c FooCmdPsql) ...           → PostgreSQL command method (preserved)
//   - func (c FooCmd) ...               → SQLite command method (preserved)
//   - func standalone(...) ...          → shared (preserved)
//   - type Shared struct ...            → shared (preserved)
//
// Only Database-receiver methods are replicated. Everything else is preserved
// verbatim in its original position.
//
// Modes:
//
//	--mode driver  (default) Tri-database replication in _custom.go files
//	--mode admin   Generate admin variants from non-admin source files
//
// Usage:
//
//	drivergen [flags] [file ...]
//	drivergen -dir internal/db/
//	drivergen --mode admin
//	drivergen -verify
package main

import (
	"flag"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	var (
		mode   string
		dir    string
		dryRun bool
		verify bool
	)
	flag.StringVar(&mode, "mode", "driver", "Generation mode: driver (tri-database) or admin (admin from non-admin)")
	flag.StringVar(&dir, "dir", "", "Process all matching files in directory (driver mode: *_custom.go)")
	flag.BoolVar(&dryRun, "dry-run", false, "Print what would change without writing")
	flag.BoolVar(&verify, "verify", false, "Exit non-zero if generated output differs from file")
	flag.Parse()

	switch mode {
	case "driver":
		os.Exit(runDriverMode(dir, dryRun, verify, flag.Args()))
	case "admin":
		os.Exit(runAdminMode(dryRun, verify, flag.Args()))
	default:
		fatal("unknown mode %q: must be 'driver' or 'admin'", mode)
	}
}

// ---------------------------------------------------------------------------
// Driver mode (tri-database replication)
// ---------------------------------------------------------------------------

func runDriverMode(dir string, dryRun, verify bool, args []string) int {
	files := args
	if dir != "" {
		matches, err := filepath.Glob(filepath.Join(dir, "*_custom.go"))
		if err != nil {
			fatal("glob: %v", err)
		}
		files = append(files, matches...)
	}
	if len(files) == 0 {
		matches, err := filepath.Glob(filepath.Join("internal", "db", "*_custom.go"))
		if err != nil {
			fatal("glob: %v", err)
		}
		files = matches
	}

	hasError := false
	for _, path := range files {
		result, err := processFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "SKIP %s: %v\n", path, err)
			continue
		}
		if result == "" {
			continue
		}
		if writeOrVerify(path, result, dryRun, verify) {
			hasError = true
		}
	}
	if hasError {
		return 1
	}
	return 0
}

// ---------------------------------------------------------------------------
// Admin mode (generate admin files from non-admin source)
// ---------------------------------------------------------------------------

func runAdminMode(dryRun, verify bool, args []string) int {
	hasError := false

	for gi := range AdminEntityGroups {
		group := &AdminEntityGroups[gi]
		replacer := group.Replacer()

		for _, pair := range group.Pairs {
			// Filter by args if provided
			if len(args) > 0 && !containsPath(args, pair.Source) && !containsPath(args, pair.Target) {
				continue
			}

			result, err := processAdminPair(pair, replacer)
			if err != nil {
				fmt.Fprintf(os.Stderr, "SKIP %s → %s: %v\n", pair.Source, pair.Target, err)
				continue
			}
			if result == "" {
				continue
			}
			if writeOrVerify(pair.Target, result, dryRun, verify) {
				hasError = true
			}
		}
	}
	if hasError {
		return 1
	}
	return 0
}

// processAdminPair generates an admin file from a non-admin source file.
// It applies the entity group's substitution map, then compares each function
// block against the existing admin file to detect divergence.
func processAdminPair(pair FilePair, replacer *strings.Replacer) (string, error) {
	sourceData, err := os.ReadFile(pair.Source)
	if err != nil {
		return "", fmt.Errorf("read source: %w", err)
	}

	// Read existing target for divergence detection
	existingData, _ := os.ReadFile(pair.Target)

	sourceLines := strings.Split(string(sourceData), "\n")
	sourceBlocks := parseBlocks(sourceLines)

	var existingBlocks []block
	if len(existingData) > 0 {
		existingBlocks = parseBlocks(strings.Split(string(existingData), "\n"))
	}

	// Build index of existing target functions/methods by name for divergence detection.
	// In admin mode, we index ALL functions (not just methods with receivers),
	// because router handlers are package-level functions.
	existingByName := map[string]block{}
	for _, b := range existingBlocks {
		if b.kind == blockFunc && b.name != "" {
			existingByName[b.name] = b
		}
	}

	// Build set of method names present in existing target (for skip detection)
	existingMethodNames := map[string]bool{}
	for _, b := range existingBlocks {
		if b.kind == blockFunc && b.name != "" {
			existingMethodNames[b.name] = true
		}
	}

	// Transform: apply substitutions to each source block
	var output []string
	for _, b := range sourceBlocks {
		// Check for skip annotation in doc comments
		if hasAnnotation(b.lines, "drivergen:skip-admin") {
			continue
		}

		transformed := make([]string, len(b.lines))
		for i, line := range b.lines {
			transformed[i] = replacer.Replace(line)
		}
		// Strip drivergen annotations from output
		transformed = stripAnnotations(transformed)

		// For type blocks: skip if the type name didn't change after substitution.
		// This prevents duplicating shared types (like RecursiveDeleteResponse)
		// that don't contain entity-specific tokens.
		if b.kind == blockType && b.name != "" {
			transformedName := replacer.Replace(b.name)
			if transformedName == b.name {
				// Type name unchanged — it's shared, don't duplicate
				continue
			}
		}

		// For function blocks, check divergence against existing target
		if b.kind == blockFunc && b.name != "" {
			transformedName := replacer.Replace(b.name)

			if existing, ok := existingByName[transformedName]; ok {
				if !linesEqual(transformed, existing.lines) {
					// Divergent — use existing hand-written version
					output = append(output, existing.lines...)
					continue
				}
			} else if len(existingMethodNames) > 0 {
				// Method doesn't exist in the current target file.
				// If the target file exists and has methods, this is likely
				// a method that was intentionally omitted from the admin version.
				// Skip it to avoid generating broken code.
				continue
			}
		}

		output = append(output, transformed...)
	}

	// Append functions and types that exist in the target but were not
	// generated from the source. These are target-only declarations
	// (e.g., non-admin has RecursiveDeleteResponse that admin lacks).
	generatedNames := map[string]bool{}
	for _, b := range sourceBlocks {
		if b.name != "" {
			generatedNames[replacer.Replace(b.name)] = true
		}
	}
	for _, b := range existingBlocks {
		if b.name == "" {
			continue
		}
		if generatedNames[b.name] {
			continue
		}
		// This declaration exists in target but not in source — preserve it
		output = append(output, b.lines...)
	}

	result := strings.Join(output, "\n")

	// If the existing target has a different import block, preserve it.
	// This handles cases where divergent methods use different imports.
	if len(existingData) > 0 {
		result = preserveImportBlock(result, string(existingData))
	}

	formatted, fmtErr := format.Source([]byte(result))
	if fmtErr != nil {
		fmt.Fprintf(os.Stderr, "WARNING: gofmt failed for %s, writing unformatted: %v\n", pair.Target, fmtErr)
		return result, nil
	}
	return string(formatted), nil
}

// preserveImportBlock replaces the import block in generated output with the
// one from the existing file. This avoids unused-import errors when divergent
// methods reference different packages than the source file.
func preserveImportBlock(generated, existing string) string {
	genImport := extractImportBlock(generated)
	existImport := extractImportBlock(existing)
	if genImport == "" || existImport == "" {
		return generated
	}
	if genImport == existImport {
		return generated
	}
	return strings.Replace(generated, genImport, existImport, 1)
}

// extractImportBlock extracts the first import(...) block from source.
func extractImportBlock(src string) string {
	start := strings.Index(src, "import (")
	if start == -1 {
		return ""
	}
	end := strings.Index(src[start:], ")")
	if end == -1 {
		return ""
	}
	return src[start : start+end+1]
}

// hasAnnotation checks if any line in the block contains the given annotation.
func hasAnnotation(lines []string, annotation string) bool {
	for _, l := range lines {
		if strings.Contains(l, annotation) {
			return true
		}
	}
	return false
}

// stripAnnotations removes drivergen annotation comments from output lines.
func stripAnnotations(lines []string) []string {
	result := make([]string, 0, len(lines))
	for _, l := range lines {
		if strings.Contains(l, "drivergen:skip-admin") {
			continue
		}
		// Strip inline annotations
		l = strings.Replace(l, "// drivergen:narrow", "", 1)
		l = strings.Replace(l, "// drivergen:widen", "", 1)
		result = append(result, l)
	}
	return result
}

func containsPath(args []string, path string) bool {
	for _, a := range args {
		if a == path || filepath.Base(a) == filepath.Base(path) {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// Shared output helpers
// ---------------------------------------------------------------------------

func writeOrVerify(path, result string, dryRun, verify bool) (hadError bool) {
	if verify {
		existing, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "VERIFY FAIL: cannot read %s: %v\n", path, err)
			return true
		}
		if string(existing) != result {
			fmt.Fprintf(os.Stderr, "VERIFY FAIL: %s is out of date. Run 'just drivergen' to regenerate.\n", path)
			return true
		}
		fmt.Printf("VERIFY OK: %s\n", path)
		return false
	}

	if dryRun {
		fmt.Printf("Would write: %s\n", path)
		return false
	}

	if err := os.WriteFile(path, []byte(result), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR writing %s: %v\n", path, err)
		return true
	}
	fmt.Printf("Generated: %s\n", path)
	return false
}

// ---------------------------------------------------------------------------
// Driver affinity — every top-level declaration is classified
// ---------------------------------------------------------------------------

type driver int

const (
	driverShared driver = iota // standalone funcs, shared types
	driverSQLite               // func (d Database), type XxxCmd struct (no Mysql/Psql)
	driverMySQL                // func (d MysqlDatabase), type XxxCmdMysql, func (c XxxCmdMysql)
	driverPSQL                 // func (d PsqlDatabase), type XxxCmdPsql, func (c XxxCmdPsql)
)

// classifyFunc determines which driver a function belongs to based on its
// receiver type. Returns driverShared for standalone functions.
func classifyFunc(receiverType string) driver {
	switch receiverType {
	case "Database":
		return driverSQLite
	case "MysqlDatabase":
		return driverMySQL
	case "PsqlDatabase":
		return driverPSQL
	default:
		// Receiver is a command type or other struct. Check for Mysql/Psql suffix.
		if strings.HasSuffix(receiverType, "Mysql") {
			return driverMySQL
		}
		if strings.HasSuffix(receiverType, "Psql") {
			return driverPSQL
		}
		// Could be a SQLite command type (no suffix) or a shared helper receiver.
		// We treat bare Cmd types as SQLite-affiliated.
		if strings.HasSuffix(receiverType, "Cmd") {
			return driverSQLite
		}
		// Unknown receiver — treat as shared (safe fallback).
		return driverShared
	}
}

// classifyType determines which driver a type declaration belongs to based
// on its name.
func classifyType(typeName string) driver {
	if strings.HasSuffix(typeName, "Mysql") {
		return driverMySQL
	}
	if strings.HasSuffix(typeName, "Psql") {
		return driverPSQL
	}
	// No suffix — could be shared or SQLite command type.
	// Command types end in "Cmd"; everything else is shared.
	if strings.HasSuffix(typeName, "Cmd") {
		return driverSQLite
	}
	return driverShared
}

// ---------------------------------------------------------------------------
// Block — a top-level declaration or chunk of non-declaration lines
// ---------------------------------------------------------------------------

type blockKind int

const (
	blockOther blockKind = iota // blank lines, comments, section markers, etc.
	blockFunc                   // function/method declaration
	blockType                   // type declaration
)

type block struct {
	kind     blockKind
	driver   driver
	lines    []string
	receiver string // for blockFunc: the receiver type name ("Database", etc.)
	name     string // for blockFunc: method name; for blockType: type name
	isDBMethod bool  // true only for Database/MysqlDatabase/PsqlDatabase receiver methods
}

// ---------------------------------------------------------------------------
// File parsing — split into classified blocks
// ---------------------------------------------------------------------------

func parseBlocks(lines []string) []block {
	var blocks []block
	i := 0

	for i < len(lines) {
		trimmed := strings.TrimSpace(lines[i])

		// Function or doc-comment-before-function
		if strings.HasPrefix(trimmed, "func ") || (strings.HasPrefix(trimmed, "//") && peekFunc(lines, i)) {
			startIdx := i
			// Skip doc comments
			for i < len(lines) && strings.HasPrefix(strings.TrimSpace(lines[i]), "//") {
				i++
			}
			if i >= len(lines) || !strings.HasPrefix(strings.TrimSpace(lines[i]), "func ") {
				// Comments at end of file or not followed by func
				blocks = append(blocks, block{kind: blockOther, driver: driverShared, lines: copyLines(lines, startIdx, i)})
				continue
			}
			recvType := extractReceiver(lines[i])
			methodName := extractMethodName(lines[i])
			collectBlock(lines, i, &i)
			d := classifyFunc(recvType)
			isDB := recvType == "Database" || recvType == "MysqlDatabase" || recvType == "PsqlDatabase"
			blockLines := copyLines(lines, startIdx, i)
			// Methods that use audited commands cannot be mechanically replicated
			// because the command types differ structurally per driver.
			usesAudited := containsAuditedCall(blockLines)
			blocks = append(blocks, block{
				kind:       blockFunc,
				driver:     d,
				lines:      blockLines,
				receiver:   recvType,
				name:       methodName,
				isDBMethod: isDB && !usesAudited,
			})
			continue
		}

		// Doc-comment-before-type or standalone comment
		if strings.HasPrefix(trimmed, "//") {
			startIdx := i
			for i < len(lines) && strings.HasPrefix(strings.TrimSpace(lines[i]), "//") {
				i++
			}
			if i < len(lines) && strings.HasPrefix(strings.TrimSpace(lines[i]), "type ") {
				// Comment precedes type — collect together
				typeName := extractTypeName(strings.TrimSpace(lines[i]))
				collectBlock(lines, i, &i)
				d := classifyType(typeName)
				blocks = append(blocks, block{
					kind:   blockType,
					driver: d,
					lines:  copyLines(lines, startIdx, i),
					name:   typeName,
				})
				continue
			}
			// Standalone comment block
			blocks = append(blocks, block{kind: blockOther, driver: driverShared, lines: copyLines(lines, startIdx, i)})
			continue
		}

		// Type declaration without preceding comment
		if strings.HasPrefix(trimmed, "type ") {
			startIdx := i
			typeName := extractTypeName(trimmed)
			collectBlock(lines, i, &i)
			d := classifyType(typeName)
			blocks = append(blocks, block{
				kind:   blockType,
				driver: d,
				lines:  copyLines(lines, startIdx, i),
				name:   typeName,
			})
			continue
		}

		// Everything else
		blocks = append(blocks, block{kind: blockOther, driver: driverShared, lines: []string{lines[i]}})
		i++
	}

	return blocks
}

// linesEqual compares two line slices after trimming whitespace from each line.
// This allows formatting differences to not count as divergence.
func linesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if strings.TrimSpace(a[i]) != strings.TrimSpace(b[i]) {
			return false
		}
	}
	return true
}

// containsAuditedCall returns true if any line in the block calls audited.Create,
// audited.Update, or audited.Delete. These methods use driver-specific command
// types and cannot be mechanically replicated.
func containsAuditedCall(lines []string) bool {
	for _, l := range lines {
		if strings.Contains(l, "audited.Create(") ||
			strings.Contains(l, "audited.Update(") ||
			strings.Contains(l, "audited.Delete(") {
			return true
		}
	}
	return false
}

// peekFunc scans past comment lines starting at idx to see if a func follows.
func peekFunc(lines []string, idx int) bool {
	j := idx
	for j < len(lines) && strings.HasPrefix(strings.TrimSpace(lines[j]), "//") {
		j++
	}
	return j < len(lines) && strings.HasPrefix(strings.TrimSpace(lines[j]), "func ")
}

// copyLines returns a copy of lines[start:end].
func copyLines(lines []string, start, end int) []string {
	if end > len(lines) {
		end = len(lines)
	}
	cp := make([]string, end-start)
	copy(cp, lines[start:end])
	return cp
}

// collectBlock advances past a brace-delimited block (func or type) starting
// at lines[startIdx]. Returns the block lines. Updates *pos to point past it.
// For single-line declarations without braces, returns just that line.
func collectBlock(lines []string, startIdx int, pos *int) []string {
	i := startIdx
	braceDepth := 0
	foundOpen := false

	for i < len(lines) {
		line := lines[i]
		opens := strings.Count(line, "{")
		closes := strings.Count(line, "}")
		braceDepth += opens - closes
		if opens > 0 {
			foundOpen = true
		}
		i++
		if foundOpen && braceDepth == 0 {
			break
		}
		// Single-line declaration without braces (e.g., type Foo string)
		if !foundOpen && i == startIdx+1 && !strings.HasSuffix(strings.TrimSpace(line), "{") {
			// Check next line for opening brace
			if i < len(lines) && strings.Contains(lines[i], "{") {
				continue // multi-line, keep going
			}
			break
		}
	}
	*pos = i
	return lines[startIdx:i]
}

// ---------------------------------------------------------------------------
// File processing
// ---------------------------------------------------------------------------

func processFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	content := string(data)
	lines := strings.Split(content, "\n")

	blocks := parseBlocks(lines)

	// Check if there are any Database-receiver methods to replicate
	hasSQLiteMethods := false
	for _, b := range blocks {
		if b.kind == blockFunc && b.isDBMethod && b.driver == driverSQLite {
			hasSQLiteMethods = true
			break
		}
	}
	if !hasSQLiteMethods {
		return "", nil // nothing to do
	}

	// Build index of existing MySQL/PSQL methods by name for divergence detection.
	existingMySQL := map[string]block{}
	existingPSQL := map[string]block{}
	for _, b := range blocks {
		if b.kind != blockFunc {
			continue
		}
		if b.receiver == "MysqlDatabase" {
			existingMySQL[b.name] = b
		}
		if b.receiver == "PsqlDatabase" {
			existingPSQL[b.name] = b
		}
	}

	// Build output:
	// 1. All shared and SQLite blocks in original order
	// 2. Generated MySQL methods (from SQLite Database-receiver methods)
	// 3. Generated PSQL methods (from SQLite Database-receiver methods)
	// 4. Preserved MySQL/PSQL non-Database-receiver blocks (command types, command methods)
	//
	// Note: MySQL/PSQL Database-receiver methods are DROPPED (regenerated).

	var output []string
	var sqliteDBMethods []block      // Database-receiver methods to replicate
	var preservedMySQL []block       // MySQL command types/methods (not regenerated)
	var preservedPSQL []block        // PSQL command types/methods (not regenerated)

	// divergent tracks methods where the existing MySQL/PSQL version doesn't
	// match what mechanical substitution would produce. These are preserved.
	divergentMySQL := map[string]bool{}
	divergentPSQL := map[string]bool{}

	for _, b := range blocks {
		switch {
		// Shared content: always keep in place
		case b.driver == driverShared:
			output = append(output, b.lines...)

		// SQLite: keep in place; collect DB methods for replication
		case b.driver == driverSQLite:
			output = append(output, b.lines...)
			if b.kind == blockFunc && b.isDBMethod {
				// Check if existing MySQL/PSQL version diverges from substitution
				if existing, ok := existingMySQL[b.name]; ok {
					generated := generateDriverLines(b.lines, "mysql")
					if !linesEqual(generated, existing.lines) {
						divergentMySQL[b.name] = true
					}
				}
				if existing, ok := existingPSQL[b.name]; ok {
					generated := generateDriverLines(b.lines, "psql")
					if !linesEqual(generated, existing.lines) {
						divergentPSQL[b.name] = true
					}
				}
				sqliteDBMethods = append(sqliteDBMethods, b)
			}

		// MySQL Database-receiver methods: drop if not divergent
		case b.driver == driverMySQL && b.isDBMethod:
			if divergentMySQL[b.name] {
				preservedMySQL = append(preservedMySQL, b)
			}
			// else: dropped, will be regenerated

		// MySQL non-DB blocks: preserve
		case b.driver == driverMySQL:
			preservedMySQL = append(preservedMySQL, b)

		// PSQL Database-receiver methods: drop if not divergent
		case b.driver == driverPSQL && b.isDBMethod:
			if divergentPSQL[b.name] {
				preservedPSQL = append(preservedPSQL, b)
			}
			// else: dropped, will be regenerated

		// PSQL non-DB blocks: preserve
		case b.driver == driverPSQL:
			preservedPSQL = append(preservedPSQL, b)

		// Fallback: keep as-is
		default:
			output = append(output, b.lines...)
		}
	}

	// Also drop section markers that were classified as shared (// MYSQL, // PSQL)
	// but actually belong to driver sections. We handle this by checking blockOther
	// lines for section marker patterns during output filtering.
	output = filterSectionMarkers(output, "MYSQL")
	output = filterSectionMarkers(output, "PSQL")

	// Emit generated MySQL section
	output = append(output, "", "// MYSQL", "")
	for _, b := range sqliteDBMethods {
		if divergentMySQL[b.name] {
			continue // preserved version used instead
		}
		generated := generateDriverLines(b.lines, "mysql")
		output = append(output, generated...)
		output = append(output, "")
	}
	// Append preserved MySQL content (divergent methods, command types, command methods)
	for _, b := range preservedMySQL {
		output = append(output, b.lines...)
	}

	// Emit generated PSQL section
	output = append(output, "", "// PSQL", "")
	for _, b := range sqliteDBMethods {
		if divergentPSQL[b.name] {
			continue // preserved version used instead
		}
		generated := generateDriverLines(b.lines, "psql")
		output = append(output, generated...)
		output = append(output, "")
	}
	// Append preserved PSQL content
	for _, b := range preservedPSQL {
		output = append(output, b.lines...)
	}

	result := strings.Join(output, "\n")
	formatted, err := format.Source([]byte(result))
	if err != nil {
		// Return unformatted — better than losing content
		fmt.Fprintf(os.Stderr, "WARNING: gofmt failed, writing unformatted: %v\n", err)
		return result, nil
	}
	return string(formatted), nil
}

// filterSectionMarkers removes standalone section marker lines (e.g., "// MYSQL")
// from the output. They'll be re-emitted in the correct position.
func filterSectionMarkers(lines []string, driverLabel string) []string {
	var result []string
	for _, l := range lines {
		trimmed := strings.TrimSpace(l)
		body := strings.TrimLeft(trimmed, "/ ")
		if strings.HasPrefix(trimmed, "//") && isSectionLabel(body, driverLabel) {
			continue
		}
		result = append(result, l)
	}
	return result
}

// isSectionLabel returns true if body (after stripping "// " prefix) is a
// bare section label like "MYSQL", "MYSQL QUERIES", "PSQL", "POSTGRES".
// Does NOT match method-specific markers like "MYSQL - MapFoo".
func isSectionLabel(body, label string) bool {
	if label == "PSQL" {
		if body == "PSQL" || body == "POSTGRES" || body == "PSQL QUERIES" || body == "POSTGRES QUERIES" {
			return true
		}
		return false
	}
	return body == label || body == label+" QUERIES"
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func extractReceiver(funcLine string) string {
	trimmed := strings.TrimSpace(funcLine)
	if !strings.HasPrefix(trimmed, "func (") {
		return ""
	}
	rest := trimmed[len("func ("):]
	paren := strings.Index(rest, ")")
	if paren == -1 {
		return ""
	}
	parts := strings.Fields(rest[:paren])
	if len(parts) < 2 {
		return ""
	}
	// Handle pointer receivers: "func (d *Database)"
	recv := parts[1]
	recv = strings.TrimPrefix(recv, "*")
	return recv
}

func extractMethodName(funcLine string) string {
	trimmed := strings.TrimSpace(funcLine)
	if !strings.HasPrefix(trimmed, "func ") {
		return ""
	}
	rest := trimmed[len("func "):]

	// Standalone function: func Name(...)
	// Method: func (recv Type) Name(...)
	if strings.HasPrefix(rest, "(") {
		// Method — skip past receiver: find matching ")"
		depth := 0
		i := 0
		for i < len(rest) {
			if rest[i] == '(' {
				depth++
			} else if rest[i] == ')' {
				depth--
				if depth == 0 {
					i++
					break
				}
			}
			i++
		}
		rest = strings.TrimSpace(rest[i:])
	}

	// rest now starts with "Name(" or "Name[T](" (generic)
	openParen := strings.IndexAny(rest, "([")
	if openParen == -1 {
		return rest
	}
	return rest[:openParen]
}

func extractTypeName(typeLine string) string {
	// "type FooCmd struct {" -> "FooCmd"
	trimmed := strings.TrimSpace(typeLine)
	if !strings.HasPrefix(trimmed, "type ") {
		return ""
	}
	rest := trimmed[len("type "):]
	space := strings.IndexAny(rest, " \t")
	if space == -1 {
		return rest
	}
	return rest[:space]
}

// ---------------------------------------------------------------------------
// Driver generation (string substitution)
// ---------------------------------------------------------------------------

type driverConfig struct {
	receiver     string // "MysqlDatabase" or "PsqlDatabase"
	pkg          string // "mdbm" or "mdbp"
	recorder     string // "MysqlRecorder" or "PsqlRecorder"
	commentLabel string // "MySQL" or "PostgreSQL"
	commentLower string // "mysql" or "psql"
	int32Cast    bool   // true for both MySQL and PostgreSQL
}

var drivers = map[string]driverConfig{
	"mysql": {
		receiver:     "MysqlDatabase",
		pkg:          "mdbm",
		recorder:     "MysqlRecorder",
		commentLabel: "MySQL",
		commentLower: "mysql",
		int32Cast:    true,
	},
	"psql": {
		receiver:     "PsqlDatabase",
		pkg:          "mdbp",
		recorder:     "PsqlRecorder",
		commentLabel: "PostgreSQL",
		commentLower: "psql",
		int32Cast:    true,
	},
}

// generateDriverLines transforms SQLite method lines into the target driver variant.
func generateDriverLines(sqliteLines []string, target string) []string {
	cfg := drivers[target]
	var result []string
	for _, line := range sqliteLines {
		result = append(result, transformLine(line, cfg))
	}
	return result
}

// transformLine applies all substitution rules to a single line.
func transformLine(line string, cfg driverConfig) string {
	out := line

	// 1. Receiver: "(d Database)" -> "(d MysqlDatabase)"
	out = strings.Replace(out, "(d Database)", "(d "+cfg.receiver+")", 1)

	// 2. sqlc package prefix: "mdb." -> "mdbm." (careful not to match mdbm./mdbp.)
	out = replaceSqlcPackage(out, cfg.pkg)

	// 3. Recorder: "SQLiteRecorder" -> "MysqlRecorder"
	out = strings.ReplaceAll(out, "SQLiteRecorder", cfg.recorder)

	// 4. Pagination int32 cast
	if cfg.int32Cast {
		out = applyInt32Casts(out)
	}

	// 5. Annotation-based casts
	out = applyAnnotationCasts(out, cfg.int32Cast)

	// 6. Doc comment labels
	out = strings.ReplaceAll(out, "(SQLite)", "("+cfg.commentLabel+")")
	out = strings.ReplaceAll(out, "(sqlite)", "("+cfg.commentLower+")")
	out = strings.ReplaceAll(out, " SQLite)", " "+cfg.commentLabel+")")
	out = strings.ReplaceAll(out, " SQLite.", " "+cfg.commentLabel+".")
	out = strings.ReplaceAll(out, "on SQLite", "on "+cfg.commentLabel)
	out = strings.ReplaceAll(out, "for SQLite", "for "+cfg.commentLabel)

	return out
}

// replaceSqlcPackage replaces "mdb." with the target package prefix,
// avoiding matches inside "mdbm." or "mdbp.".
func replaceSqlcPackage(line, targetPkg string) string {
	var result strings.Builder
	i := 0
	for i < len(line) {
		if i+4 <= len(line) && line[i:i+4] == "mdb." {
			// Check preceding char isn't part of a longer prefix
			if i > 0 && (line[i-1] >= 'a' && line[i-1] <= 'z') {
				result.WriteByte(line[i])
				i++
				continue
			}
			result.WriteString(targetPkg + ".")
			i += 4
		} else {
			result.WriteByte(line[i])
			i++
		}
	}
	return result.String()
}

// applyInt32Casts wraps pagination params with int32() for MySQL/PostgreSQL.
func applyInt32Casts(line string) string {
	trimmed := strings.TrimSpace(line)

	// Pattern: struct field assignment with params.Limit or params.Offset
	if !strings.Contains(trimmed, ":") {
		return line
	}
	if !strings.Contains(trimmed, "params.Limit") && !strings.Contains(trimmed, "params.Offset") {
		return line
	}
	if strings.Contains(trimmed, "int32(") {
		return line // already wrapped
	}

	colonIdx := strings.Index(trimmed, ":")
	if colonIdx == -1 {
		return line
	}
	value := strings.TrimSpace(trimmed[colonIdx+1:])
	value = strings.TrimSuffix(value, ",")

	if value == "params.Limit" || value == "params.Offset" {
		indent := line[:len(line)-len(strings.TrimLeft(line, " \t"))]
		field := strings.TrimSpace(trimmed[:colonIdx])
		return indent + field + ":  int32(" + value + "),"
	}

	return line
}

// applyAnnotationCasts handles // drivergen:narrow and // drivergen:widen annotations.
func applyAnnotationCasts(line string, int32Cast bool) string {
	if !int32Cast {
		return line
	}

	if strings.Contains(line, "// drivergen:narrow") {
		line = strings.Replace(line, "// drivergen:narrow", "", 1)
		line = wrapStructFieldValue(line, "int32")
	}
	if strings.Contains(line, "// drivergen:widen") {
		line = strings.Replace(line, "// drivergen:widen", "", 1)
		line = wrapStructFieldValue(line, "int64")
	}
	return line
}

// wrapStructFieldValue wraps the value in a struct field assignment with a cast.
func wrapStructFieldValue(line, cast string) string {
	trimmed := strings.TrimSpace(line)
	colonIdx := strings.Index(trimmed, ":")
	if colonIdx == -1 {
		return line
	}
	value := strings.TrimSpace(trimmed[colonIdx+1:])
	hasSuffix := strings.HasSuffix(value, ",")
	value = strings.TrimSuffix(value, ",")
	value = strings.TrimSpace(value)

	if strings.HasPrefix(value, cast+"(") {
		return line
	}

	indent := line[:len(line)-len(strings.TrimLeft(line, " \t"))]
	field := strings.TrimSpace(trimmed[:colonIdx])
	result := indent + field + ": " + cast + "(" + value + ")"
	if hasSuffix {
		result += ","
	}
	return result
}

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
