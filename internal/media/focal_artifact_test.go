package media

import (
	"database/sql"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"testing"

	db "github.com/hegner123/modulacms/internal/db"
)

// TestFocalPointArtifacts generates crop output files to test_artifacts/ for visual inspection.
// Run with: go test -v -run TestFocalPointArtifacts ./internal/media/
func TestFocalPointArtifacts(t *testing.T) {
	artifactDir := filepath.Join("..", "..", "test_artifacts")
	if err := os.MkdirAll(artifactDir, 0755); err != nil {
		t.Fatalf("create artifact dir: %v", err)
	}

	fixturePath := filepath.Join("..", "..", "test300x300.png")
	if _, err := os.Stat(fixturePath); err != nil {
		t.Skipf("fixture not found: %v", err)
	}

	dims := []db.MediaDimensions{
		{Width: sql.NullInt64{Int64: 150, Valid: true}, Height: sql.NullInt64{Int64: 150, Valid: true}},
		{Width: sql.NullInt64{Int64: 200, Valid: true}, Height: sql.NullInt64{Int64: 100, Valid: true}},
	}
	lister := &mockDimensionLister{dims: &dims}

	type testCase struct {
		label string
		fp    *image.Point
	}

	cases := []testCase{
		{"nil_center", nil},
		{"focal_0_0", &image.Point{X: 0, Y: 0}},
		{"focal_150_150", &image.Point{X: 150, Y: 150}},
		{"focal_299_299", &image.Point{X: 299, Y: 299}},
		{"focal_50_250", &image.Point{X: 50, Y: 250}},
		{"focal_250_50", &image.Point{X: 250, Y: 50}},
	}

	for _, tc := range cases {
		// Each case gets its own subdirectory
		subDir := filepath.Join(artifactDir, tc.label)
		if err := os.MkdirAll(subDir, 0755); err != nil {
			t.Fatalf("mkdir %s: %v", subDir, err)
		}

		// Copy fixture into subDir as source
		srcPath := filepath.Join(subDir, "test300x300.png")
		if err := copyFile(fixturePath, srcPath); err != nil {
			t.Fatalf("copy fixture for %s: %v", tc.label, err)
		}

		files, err := OptimizeUpload(srcPath, subDir, lister, tc.fp)
		if err != nil {
			t.Fatalf("OptimizeUpload %s: %v", tc.label, err)
		}

		fpStr := "nil"
		if tc.fp != nil {
			fpStr = fmt.Sprintf("(%d,%d)", tc.fp.X, tc.fp.Y)
		}
		t.Logf("=== %s (focal=%s) ===", tc.label, fpStr)
		for _, f := range *files {
			info, _ := os.Stat(f)
			t.Logf("  %s  (%d bytes)", filepath.Base(f), info.Size())
		}
	}

	t.Logf("\nArtifacts saved to: %s", artifactDir)
	t.Logf("Open in Finder: open %s", artifactDir)
}
