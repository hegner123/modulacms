package development

import (
	"testing"

	"github.com/hegner123/modulacms/internal/cli"
	"github.com/hegner123/modulacms/internal/db"
)

type InsertAttempt struct {
	Node             db.ContentData
	Depth            int
	DatatypeName     string
	ExpectedToInsert bool
	ActuallyInserted bool
}

func calculateDepth(node db.ContentData, contentMap map[int64]db.ContentData) int {
	if !node.ParentID.Valid {
		return 0
	}

	parent, exists := contentMap[node.ParentID.Int64]
	if !exists {
		return 0
	}

	return 1 + calculateDepth(parent, contentMap)
}

func getDatatypeName(datatypeID int64, database *db.Database) (string, error) {
	datatype, err := database.GetDatatype(datatypeID)
	if err != nil {
		return "", err
	}
	return datatype.Label, nil
}

func TestTreeQueries(t *testing.T) {
	database, err := ConnectToDatabase("./modula.db")
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer HandleDatabaseCloseDeferErr(database)

	t.Run("ListAllContentData", func(t *testing.T) {

		contentData, err := database.ListContentData()
		if err != nil {
			LogError("ListContentData", err)
			t.Errorf("Failed to list content data: %v", err)
			return
		}

		LogResults("All Content Data", contentData)
		t.Logf("Found %d content data records", len(*contentData))
	})

	t.Run("BuildTreeStructure", func(t *testing.T) {
		contentData, err := database.ListContentData()
		if err != nil {
			LogError("ListContentData", err)
			t.Errorf("Failed to list content data: %v", err)
			return
		}
		res, err := database.GetRouteTreeByRouteID(1)
		if err != nil {
			return
		}
		db.LogRouteTree("Route Tree Query Result", res)

		if len(*contentData) == 0 {
			t.Skip("No content data available for tree building")
		}

		var rootNode *db.ContentData
		var datatypeDef *db.Datatypes
		for _, data := range *contentData {
			if !data.ParentID.Valid {
				rootNode = &data
				break
			}
		}

		if rootNode == nil {
			t.Skip("No root node found (node without parent)")
		}

		treeNode := cli.TreeNode{
			Node: rootNode,
		}
		datatypeDef, err = database.GetDatatype(treeNode.Node.DatatypeID)
		if err != nil {
			return
		}
		treeNode.NodeDatatype = *datatypeDef

		tree := cli.NewTreeRoot(treeNode)

		LogTreeStructure("Tree Root Node", tree)

		for _, data := range *contentData {
			if data.ParentID.Valid && data.ContentDataID != rootNode.ContentDataID {
				childNode := cli.TreeNode{
					Node: &data,
				}

				datatypeDef, err := database.GetDatatype(data.DatatypeID)
				if err == nil {
					childNode.NodeDatatype = *datatypeDef
				}

				inserted := tree.Insert(childNode, data.ParentID.Int64)
				if inserted {
					t.Logf("Successfully inserted node %d under parent %d",
						data.ContentDataID, data.ParentID.Int64)
				} else {
					t.Logf("Failed to insert node %d under parent %d",
						data.ContentDataID, data.ParentID.Int64)
				}
			}
		}

		LogTreeStructure("Complete Tree Structure", tree)
	})

	t.Run("QueryContentDataByRoute", func(t *testing.T) {
		contentData, err := database.ListContentData()
		if err != nil {
			LogError("ListContentData", err)
			t.Errorf("Failed to list content data: %v", err)
			return
		}

		if len(*contentData) == 0 {
			t.Skip("No content data available")
		}

		routeID := (*contentData)[0].RouteID
		routeData, err := database.ListContentDataByRoute(routeID)
		if err != nil {
			LogError("ListContentDataByRoute", err)
			t.Errorf("Failed to get content data by route: %v", err)
			return
		}

		LogResults("Content Data by Route", routeData)
		t.Logf("Found %d records for route %d", len(*routeData), routeID)
	})

	t.Run("TestTreeInitLogic", func(t *testing.T) {
		// Define expected init behavior
		expectedDepthLimits := map[string]int{
			"Page":        0,  // Root level
			"Navigation":  1,  // Always load for page structure
			"Hero":        1,  // Always load for page structure
			"Footer":      1,  // Always load for page structure
			"Container":   1,  // Load containers for layout
			"Row":         2,  // Load rows when container accessed
			"Column":      -1, // Don't auto-load, lazy load only
			"Card":        -1, // Don't auto-load, lazy load only
			"RichText":    -1, // Don't auto-load, lazy load only
			"Image":       -1, // Don't auto-load, lazy load only
			"Button":      -1, // Don't auto-load, lazy load only
			"CardHeader":  -1, // Don't auto-load, lazy load only
			"CardBody":    -1, // Don't auto-load, lazy load only
			"CardFooter":  -1, // Don't auto-load, lazy load only
			"Feature":     -1, // Don't auto-load, lazy load only
			"Testimonial": -1, // Don't auto-load, lazy load only
			"Gallery":     -1, // Don't auto-load, lazy load only
		}

		contentData, err := database.ListContentData()
		if err != nil {
			t.Fatalf("Failed to list content data: %v", err)
		}

		if len(*contentData) == 0 {
			t.Skip("No content data available for init logic testing")
		}

		// Track what gets inserted vs what should be deferred
		insertAttempts := make(map[int64]InsertAttempt)

		// Build lookup maps for efficient processing
		contentMap := make(map[int64]db.ContentData)
		for _, data := range *contentData {
			contentMap[data.ContentDataID] = data
		}

		for _, data := range *contentData {
			depth := calculateDepth(data, contentMap)
			datatypeName, err := getDatatypeName(data.DatatypeID, database)
			if err != nil {
				datatypeName = "Unknown"
			}

			expectedDepth, exists := expectedDepthLimits[datatypeName]
			var shouldInsert bool
			if !exists {
				shouldInsert = false // Unknown datatypes shouldn't auto-insert
			} else if expectedDepth == -1 {
				shouldInsert = false // -1 means lazy load only, don't auto-insert
			} else {
				shouldInsert = depth <= expectedDepth // Insert if within expected depth
			}

			insertAttempts[data.ContentDataID] = InsertAttempt{
				Node:             data,
				Depth:            depth,
				DatatypeName:     datatypeName,
				ExpectedToInsert: shouldInsert,
				ActuallyInserted: false,
			}
		}

		// Find root node and build tree
		var rootNode *db.ContentData
		for _, data := range *contentData {
			if !data.ParentID.Valid {
				rootNode = &data
				break
			}
		}

		if rootNode == nil {
			t.Skip("No root node found")
		}

		// Create root tree node with datatype info
		rootTreeNode := cli.TreeNode{
			Node: rootNode,
		}
		datatypeDef, err := database.GetDatatype(rootTreeNode.Node.DatatypeID)
		if err == nil {
			rootTreeNode.NodeDatatype = *datatypeDef
		}

		tree := cli.NewTreeRoot(rootTreeNode)
		insertAttempts[rootNode.ContentDataID] = InsertAttempt{
			Node:             *rootNode,
			Depth:            0,
			DatatypeName:     "Page",
			ExpectedToInsert: true,
			ActuallyInserted: true, // Root is always inserted
		}

		// Attempt to insert all other nodes
		for _, data := range *contentData {
			if data.ParentID.Valid && data.ContentDataID != rootNode.ContentDataID {
				childNode := cli.TreeNode{
					Node: &data,
				}

				datatypeDef, err := database.GetDatatype(data.DatatypeID)
				if err == nil {
					childNode.NodeDatatype = *datatypeDef
				}

				inserted := tree.Insert(childNode, data.ParentID.Int64)

				// Update the attempt record
				if attempt, exists := insertAttempts[data.ContentDataID]; exists {
					attempt.ActuallyInserted = inserted
					insertAttempts[data.ContentDataID] = attempt
				}
			}
		}

		for _, data := range *contentData {
			if data.ParentID.Valid && !insertAttempts[data.ContentDataID].ActuallyInserted {
				childNode := cli.TreeNode{
					Node: &data,
				}

				datatypeDef, err := database.GetDatatype(data.DatatypeID)
				if err == nil {
					childNode.NodeDatatype = *datatypeDef
				}

				var inserted bool
				if parentNode, exists := tree.NodeIndex[data.ParentID.Int64]; exists && parentNode != nil {
					inserted = parentNode.ShallowInsert(childNode, data.ParentID.Int64)
				} else {
					inserted = false
				}

				// Update the attempt record
				if attempt, exists := insertAttempts[data.ContentDataID]; exists {
					attempt.ActuallyInserted = inserted
					insertAttempts[data.ContentDataID] = attempt
				}
			}
		}
		// Analyze results
		correctPredictions := 0
		totalNodes := 0

		t.Logf("\n=== Init Logic Analysis ===")
		for id, attempt := range insertAttempts {
			totalNodes++

			if attempt.ExpectedToInsert == attempt.ActuallyInserted {
				correctPredictions++
			}

			status := "✓"
			if attempt.ExpectedToInsert != attempt.ActuallyInserted {
				status = "✗"
			}

			expectedStr := "DEFER"
			if attempt.ExpectedToInsert {
				expectedStr = "INSERT"
			}

			actualStr := "DEFERRED"
			if attempt.ActuallyInserted {
				actualStr = "INSERTED"
			}

			t.Logf("%s Node %d: %s (depth %d) - Expected: %s, Actual: %s",
				status, id, attempt.DatatypeName, attempt.Depth, expectedStr, actualStr)
		}

		accuracy := float64(correctPredictions) / float64(totalNodes) * 100
		t.Logf("\nInit Logic Accuracy: %.1f%% (%d/%d correct predictions)",
			accuracy, correctPredictions, totalNodes)

		if accuracy < 75.0 {
			t.Errorf("Init logic accuracy too low: %.1f%%. Expected at least 75%%", accuracy)
		}

		LogTreeStructure("Final Tree After Init Logic Test", tree)
	})
}
