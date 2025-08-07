package development

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/hegner123/modulacms/internal/cli"
	"github.com/hegner123/modulacms/internal/db"
	_ "github.com/mattn/go-sqlite3"
)

func ConnectToDatabase(dbPath string) (*db.Database, error) {
	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	database := &db.Database{
		Connection: conn,
		Context:    context.Background(),
	}

	return database, nil
}

func LogResults(title string, results any) {
	fmt.Printf("\n=== %s ===\n", title)
	switch results := results.(type) {
	case *[]db.ContentData:
		LogContentDataRows(*results)
	default:

		fmt.Printf("%+v\n", results)
	}
	fmt.Println("==================")
}

func LogContentDataRows(rows []db.ContentData) {
	for i, v := range rows {
		fmt.Printf("| ContentDataID: %+v ", v.ContentDataID)
		fmt.Printf(" RoutID: %+v ", v.RouteID)
		if i%2 == 1 {
			fmt.Print("|")
			fmt.Println()
		}
	}
	fmt.Println()
}

func LogError(operation string, err error) {
	log.Printf("Error in %s: %v", operation, err)
}

func LogTreeStructure(title string, tree *cli.TreeRoot) {
	fmt.Printf("\n=== %s ===\n", title)
	if tree == nil {
		fmt.Println("Tree is nil")
		fmt.Println("==================")
		return
	}

	fmt.Printf("Tree Index Size: %d nodes\n", len(tree.NodeIndex))
	fmt.Println("\nTree Structure:")
	LogTreeNode(tree.Root, "", true)
	fmt.Println("==================")
	fmt.Println()
}

func LogTreeNode(node *cli.TreeNode, prefix string, isLast bool) {
	if node == nil {
		return
	}

	connector := "├── "
	if isLast {
		connector = "└── "
	}

	fmt.Printf("%s%sID: %d | Route: %d | Datatype: %d",
		prefix, connector,
		node.Node.ContentDataID,
		node.Node.RouteID,
		node.Node.DatatypeID)

	if node.NodeDatatype.Label != "" {
		fmt.Printf(" (%s)", node.NodeDatatype.Label)
	}

	if node.Node.ParentID.Valid {
		fmt.Printf(" | Parent: %d", node.Node.ParentID.Int64)
	} else {
		fmt.Printf(" | ROOT")
	}
	fmt.Println()

	if node.Nodes != nil && len(*node.Nodes) > 0 {
		childPrefix := prefix
		if isLast {
			childPrefix += "    "
		} else {
			childPrefix += "│   "
		}

		children := *node.Nodes
		for i, child := range children {
			isLastChild := i == len(children)-1
			LogTreeNode(child, childPrefix, isLastChild)
		}
	}
}
