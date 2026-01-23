package model

import (
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
)

// CreateRoot creates a new Root instance for the content model
func CreateRoot() Root {
	return NewRoot()
}

// CreateNode creates a new Node with datatype and adds it to the tree
func CreateNode(root Root, parentID, datatypeID, contentID string) Root {
	datatype := Datatype{
		Info: db.DatatypeJSON{
			DatatypeID: datatypeID,
		},
		Content: db.ContentDataJSON{
			ContentDataID: contentID,
		},
	}

	node := NewNode(datatype)

	// If root is empty, add as root node
	if root.Node == nil {
		return AddChild(root, &node)
	}

	// If parentID provided, find parent and add as child
	if parentID != "" {
		parentNode := root.Node.FindNodeByID(parentID)
		if parentNode != nil {
			parentNode.AddChild(&node)
		}
	}

	return root
}

// LoadPageContent loads a page by content ID and builds a model tree
func LoadPageContent(contentID int64, config config.Config) (Root, error) {
	// TODO: Implement loading page content from DB
	// This would fetch the content data and all related content fields
	// and build a model tree from it
	return NewRoot(), nil
}

// SavePageContent saves a page model tree to the database
func SavePageContent(root Root, config config.Config) error {
	// TODO: Implement saving the page content to DB
	// This would convert the model tree to database entries
	return nil
}