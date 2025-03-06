package query

import (
	"fmt"
	sq "github.com/Masterminds/squirrel"
)

type SqlLogicOperators int
type SqlMethods int

const (
	INSERT SqlMethods        = 1
	SELECT SqlMethods        = 2
	UPDATE SqlMethods        = 3
	DELETE SqlMethods        = 4
	AND    SqlLogicOperators = 0
	OR     SqlLogicOperators = 1
)

type SQLQuery struct {
	Method  SqlMethods
	Table   string
	Columns []string
	Values  []string
	Filter  []WhereKeyValue
	Err     error
}

type WhereKeyValue struct {
	Key   string
	Value string
	Logic *SqlLogicOperators
}

func Query() (*string, error) {
	sqs, _, _ := sq.Insert("users").Columns("username", "name", "email").Values("snappy", "bob", "bob@test.com").ToSql()
	return &sqs, nil
}

func ParseMethods(m SQLQuery) sq.InsertBuilder {
	var res sq.InsertBuilder

	switch m.Method {
	case INSERT:
		res = BuildInsert(m)
	case SELECT:
	case UPDATE:
	case DELETE:
	}

	return res
}

func BuildInsert(m SQLQuery) sq.InsertBuilder {
	sq := sq.Insert(m.Table).Columns("username")
	return sq
}

func BuildColumns(c []string) {

}

func ParseFilters(f []WhereKeyValue) string {
	var res string
	for _, v := range f {
		fmt.Println(v)
	}

	return res
}
