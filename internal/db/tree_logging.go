package db

import (
	"fmt"
	"strings"

	"github.com/hegner123/modulacms/internal/db/types"
)

///////////////////////////////
//LOGGING FUNCTIONS
//////////////////////////////

// LogRouteTree provides a readable format for GetRouteTreeByRouteID output
func LogRouteTree(title string, rows *[]GetRouteTreeByRouteIDRow)string {
	s := ""
	if rows == nil || len(*rows) == 0 {
		s += fmt.Sprintf("\n=== %s ===\nNo tree data found\n==================\n\n", title)
		return s
	}

	s += fmt.Sprintf("\n=== %s ===\n", title)

	// Group rows by ContentDataID for better readability
	nodeMap := make(map[types.ContentID][]GetRouteTreeByRouteIDRow)
	for _, row := range *rows {
		nodeMap[row.ContentDataID] = append(nodeMap[row.ContentDataID], row)
	}

	// Display each node with its fields
	for contentID, nodeRows := range nodeMap {
		if len(nodeRows) == 0 {
			continue
		}

		firstRow := nodeRows[0]
		parentStr := "ROOT"
		if firstRow.ParentID.Valid {
			parentStr = fmt.Sprintf("Parent: %s", firstRow.ParentID.ID)
		}

		s += fmt.Sprintf("┌─ Node ID: %s | %s | %s (%s)\n",
			contentID, parentStr, firstRow.DatatypeLabel, firstRow.DatatypeType)

		// Display all fields for this node
		for i, row := range nodeRows {
			valueStr := "<empty>"
			if row.FieldValue.Valid && row.FieldValue.String != "" {
				value := row.FieldValue.String
				if len(value) > 60 {
					value = value[:57] + "..."
				}
				// Replace newlines and tabs for cleaner display
				value = strings.ReplaceAll(value, "\n", "\\n")
				value = strings.ReplaceAll(value, "\t", "\\t")
				valueStr = value
			}

			connector := "├─"
			if i == len(nodeRows)-1 {
				connector = "└─"
			}

			s += fmt.Sprintf("│  %s %s (%s): %s\n",
				connector, row.FieldLabel, row.FieldType, valueStr)
		}
		fmt.Println("│")
	}

	s += fmt.Sprintf("==================\nTotal nodes: %d\n\n", len(nodeMap))
        return s
}

// LogRouteTreeCompact provides a more compact format for GetRouteTreeByRouteID output
func LogRouteTreeCompact(title string, rows *[]GetRouteTreeByRouteIDRow)string {
	s := ""
	if rows == nil || len(*rows) == 0 {
		s += fmt.Sprintf("\n=== %s ===\nNo tree data found\n==================\n\n", title)
		return s
	}

	s += fmt.Sprintf("\n=== %s ===\n", title)

	// Group rows by ContentDataID
	nodeMap := make(map[types.ContentID][]GetRouteTreeByRouteIDRow)
	for _, row := range *rows {
		nodeMap[row.ContentDataID] = append(nodeMap[row.ContentDataID], row)
	}

	// Display each node in compact format
	for contentID, nodeRows := range nodeMap {
		if len(nodeRows) == 0 {
			continue
		}

		firstRow := nodeRows[0]
		parentStr := "ROOT"
		if firstRow.ParentID.Valid {
			parentStr = fmt.Sprintf("P:%s", firstRow.ParentID.ID)
		}

		// Count non-empty fields
		fieldCount := 0
		for _, row := range nodeRows {
			if row.FieldValue.Valid && row.FieldValue.String != "" {
				fieldCount++
			}
		}

		s += fmt.Sprintf("• ID:%s %s %s (%d fields)\n",
			contentID, parentStr, firstRow.DatatypeLabel, fieldCount)
	}

	s += fmt.Sprintf("==================\nTotal: %d nodes\n\n", len(nodeMap))
        return s
}

// LogRouteTreeHierarchical displays the tree data in a hierarchical format showing parent-child relationships
func LogRouteTreeHierarchical(title string, rows *[]GetRouteTreeByRouteIDRow)string {
	s := ""
	if rows == nil || len(*rows) == 0 {
		s += fmt.Sprintf("\n=== %s ===\nNo tree data found\n==================\n\n", title)
		return s
	}

	s += fmt.Sprintf("\n=== %s ===\n", title)

	// Group rows by ContentDataID
	nodeMap := make(map[types.ContentID][]GetRouteTreeByRouteIDRow)
	for _, row := range *rows {
		nodeMap[row.ContentDataID] = append(nodeMap[row.ContentDataID], row)
	}

	// Find root nodes (those without parents)
	var rootNodes []types.ContentID
	childrenMap := make(map[types.ContentID][]types.ContentID)

	for contentID, nodeRows := range nodeMap {
		if len(nodeRows) == 0 {
			continue
		}

		firstRow := nodeRows[0]
		if !firstRow.ParentID.Valid {
			rootNodes = append(rootNodes, contentID)
		} else {
			parentID := firstRow.ParentID.ID
			childrenMap[parentID] = append(childrenMap[parentID], contentID)
		}
	}

	// Print hierarchy starting from root nodes
	var printNode func(nodeID types.ContentID, prefix string, isLast bool)
	printNode = func(nodeID types.ContentID, prefix string, isLast bool) {
		nodeRows, exists := nodeMap[nodeID]
		if !exists || len(nodeRows) == 0 {
			return
		}

		firstRow := nodeRows[0]
		connector := "├─"
		if isLast {
			connector = "└─"
		}

		// Count fields with values
		fieldCount := 0
		for _, row := range nodeRows {
			if row.FieldValue.Valid && row.FieldValue.String != "" {
				fieldCount++
			}
		}

		s += fmt.Sprintf("%s%s ID:%s %s (%d fields)\n",
			prefix, connector, nodeID, firstRow.DatatypeLabel, fieldCount)

		// Print children
		children := childrenMap[nodeID]
		for i, childID := range children {
			isLastChild := i == len(children)-1
			childPrefix := prefix
			if isLast {
				childPrefix += "    "
			} else {
				childPrefix += "│   "
			}
			printNode(childID, childPrefix, isLastChild)
		}
	}

	// Print all root nodes
	for i, rootID := range rootNodes {
		isLast := i == len(rootNodes)-1
		printNode(rootID, "", isLast)
	}

	s += fmt.Sprintf("==================\nTotal nodes: %d\n\n", len(nodeMap))
        return s
}

