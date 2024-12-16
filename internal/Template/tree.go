
package mTemplate

import (
	"fmt"
	"runtime"

	mdb "github.com/hegner123/modulacms/db-sqlite"
)


type (
	Scanner func(parent, child mdb.AdminDatatypes) int
)

func Scan(parent, child mdb.AdminDatatypes) int {
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
	Val     mdb.AdminDatatypes
	Juniors []*Node
	Fields  []mdb.AdminFields
}

func NewTree(scanFunction Scanner) *Tree {
	return &Tree{
		fn: scanFunction,
	}
}

func (tree *Tree) Add(dt mdb.AdminDatatypes) {
	tree.Root = tree.Root.Add(tree.fn, dt)
}

func (parent *Node) Add(compare Scanner, child mdb.AdminDatatypes) *Node {
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
			db, ctx, err := getDb(Database{src: dbName})
			if err != nil {
				logError("failed to connect to db", err)
				_, file, line, _ := runtime.Caller(0)
				fmt.Printf("Current line number: %s:%d\n", file, line)
			}
			defer db.Close()
			field := dbGetAdminField(db, ctx, fieldId)
			node.Fields = append(node.Fields, field)
		}
	}
	for _, junior := range node.Juniors {
		junior.AddTreeFields(dbName)
	}
}

func dbGetChildren(res []int64, id int, dbName string, message string) []int64 {
	if dbName == "" {
		dbName = ""
	}
	db, ctx, err := getDb(Database{src: dbName})
	if err != nil {
		logError("failed to connect to db", err)
		_, file, line, _ := runtime.Caller(0)
		fmt.Printf("Current line number: %s:%d\n", file, line)
	}
	defer db.Close()
	queries := mdb.New(db)
	rows, err := queries.ListAdminDatatypeChildren(ctx, ni(id))
	if err != nil {
		logError("failed to query admin Datatypes ", err)
		_, file, line, _ := runtime.Caller(0)
		fmt.Printf("Current line number: %s:%d\n", file, line)
	}
	if len(rows) == 0 {
		return res
	} else {
		for i, row := range rows {
			res = append(res, row.AdminDtID)
			l := fmt.Sprintf("Rows %d ", len(rows))
			s := fmt.Sprintf("row index:%d, id %d, %s\n\n", i, id, l)
			res = dbGetChildren(res, int(row.AdminDtID), dbName, s)
		}
		return res
	}
}

func dbGetFields(res []int64, id int, dbName string, message string) []int64 {
	if dbName == "" {
		dbName = ""
	}
	db, ctx, err := getDb(Database{src: dbName})
	if err != nil {
		logError("failed to connect to db", err)
		_, file, line, _ := runtime.Caller(0)
		fmt.Printf("Current line number: %s:%d\n", file, line)
	}
	defer db.Close()
	rows := dbListAdminFieldByAdminDtId(db, ctx, int64(id))
	for _, row := range rows {
		res = append(res, row.AdminFieldID)
	}
	return res
}
