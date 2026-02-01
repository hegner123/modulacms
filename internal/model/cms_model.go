package model

import (
	"fmt"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
)

// CreateNode constructs a minimal Node from the given IDs and attaches it to
// the tree. If the root is empty, the node becomes the root. Otherwise, it
// searches for a parent node by parentID and appends the new node as a child.
//
// TODO: This function creates a Datatype with only DatatypeID and ContentDataID
// populated — all other fields (Label, Type, Slug, ParentID, AuthorID, dates,
// RouteID, etc.) are zero-valued. The resulting node is a skeleton that cannot
// be rendered or persisted meaningfully. Needs to accept full Datatype/Field
// data, or fetch missing data from the database.
func CreateNode(root Root, parentID, datatypeID, contentID string) (Root, error) {
	// Build a minimal Datatype with only the ID fields set.
	// All other fields (Label, Type, Slug, dates, etc.) remain zero-valued.
	datatype := Datatype{
		Info: db.DatatypeJSON{
			DatatypeID: datatypeID,
		},
		Content: db.ContentDataJSON{
			ContentDataID: contentID,
		},
	}

	node := NewNode(datatype)

	// If no root exists yet, set this node as the root.
	if root.Node == nil {
		return SetRootNode(root, &node), nil
	}

	// No parent specified, nothing to attach to.
	if parentID == "" {
		return root, nil
	}

	// Attempt to find the parent and append this node as a child.
	parentNode := root.Node.FindNodeByID(parentID)
	if parentNode == nil {
		return root, fmt.Errorf("CreateNode: parent node %s not found", parentID)
	}
	parentNode.AddChild(&node)

	return root, nil
}

// LoadPageContent is a stub intended to fetch content data and all related
// content fields from the database by content ID, then build a model tree
// using BuildTree.
func LoadPageContent(contentID string, config config.Config) (Root, error) {
	return Root{}, fmt.Errorf("LoadPageContent: not implemented")
}

// SavePageContent is a stub intended to decompose a Root tree back into
// database rows (content_data, content_fields, etc.) and persist them.
func SavePageContent(root Root, config config.Config) error {
	return fmt.Errorf("SavePageContent: not implemented")
}