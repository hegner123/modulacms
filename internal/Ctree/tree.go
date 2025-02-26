package ctree

import (
	"fmt"

	mdb "github.com/hegner123/modulacms/db-sqlite"
	db "github.com/hegner123/modulacms/internal/Db"
	utility "github.com/hegner123/modulacms/internal/Utility"
)

type (
	Scanner func(parent, child mdb.ListAdminDatatypeByRouteIdRow) int
)

func Scan(parent, child mdb.ListAdminDatatypeByRouteIdRow) int {
	if parent.AdminDtID == child.ParentID.Int64 {
		return 1
	} else {
		return 0
	}
}

type Tree struct {
	fn   Scanner
	Root *Node
}

type Node struct {
	Val     mdb.ListAdminDatatypeByRouteIdRow
	Juniors []*Node
	Fields  []mdb.AdminFields
}

func NewTree(scanFunction Scanner) *Tree {
	return &Tree{
		fn: scanFunction,
	}
}

func (tree *Tree) Add(dt mdb.ListAdminDatatypeByRouteIdRow) {
	tree.Root = tree.Root.Add(tree.fn, dt)
}

func (parent *Node) Add(compare Scanner, child mdb.ListAdminDatatypeByRouteIdRow) *Node {
	if parent == nil {
		return &Node{Val: child}
	}
	switch r := compare(parent.Val, child); {
	case r == 1:
		parent.Juniors = append(parent.Juniors, &Node{Val: child})
	case r == 0:
		for i := 0; i < len(parent.Juniors); i++ {
			parent.Juniors[i].Add(compare, child)
		}
	}

	return parent
}

func (node *Node) PrintTree(level int) {
	if node == nil {
		return
	}
	indent := ""
	for i := 0; i < level; i++ {
		indent += "  "
	}
	fmt.Printf("%s%d (%s)\n", indent, node.Val.AdminDtID, node.Val.Label)
	for _, v := range node.Fields {
		fmt.Printf("%s%d (%s)\n", "|"+indent, v.AdminFieldID, v.Label)
	}
	for _, junior := range node.Juniors {
		junior.PrintTree(level + 1)
	}
}

func (node *Node) AddTreeFields(dbName string) {
	if node == nil {
		return
	}
	var res []int64
	rows := dbGetFields(res, int(node.Val.AdminDtID), "", "")
	if len(rows) > 0 {
		for _, fieldId := range rows {
			if dbName == "" {
				dbName = ""
			}
			dbc := db.GetDb(db.Database{})
			defer dbc.Connection.Close()
			field, err := db.GetAdminField(dbc.Connection, dbc.Context, fieldId)
			if err != nil {
				utility.LogError("failed to : ", err)
			}
			node.Fields = append(node.Fields, *field)
		}
	}
	for _, junior := range node.Juniors {
		junior.AddTreeFields(dbName)
	}
}

func DbGetChildren(res []int64, id int64, dbName string, message string) []int64 {
	if dbName == "" {
		dbName = ""
	}
	dbc := db.GetDb(db.Database{})
	rows, err := db.ListAdminDatatypeChildren(dbc.Connection, dbc.Context, id)
	if err != nil {
		utility.LogError("failed to : ", err)
	}
	if len(*rows) == 0 {
		return res
	} else {
		for i, row := range *rows {
			res = append(res, row.AdminDtID)
			l := fmt.Sprintf("Rows %d ", len(*rows))
			s := fmt.Sprintf("row index:%d, id %d, %s\n\n", i, id, l)
			res = DbGetChildren(res, int64(row.AdminDtID), dbName, s)
		}
		return res
	}
}

func dbGetFields(res []int64, id int, dbPath string, message string) []int64 {
	var dbName string
	if dbPath == "" {
		dbName = ""
	} else {
		dbName = dbPath
	}

	dbc := db.GetDb(db.Database{Src: dbName})
	defer dbc.Connection.Close()
	rows, err := db.ListAdminFieldByAdminDtId(dbc.Connection, dbc.Context, int64(id))
	if err != nil {
		fmt.Printf("failed to : %v", err)
	}
	for _, row := range *rows {
		res = append(res, row.AdminFieldID)
	}
	return res
}
