package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

func TestDatatypeJSON(t *testing.T) {
	parent := sql.NullInt64{
		Valid: false,
		Int64: 0,
	}
	tm := time.Now().Format(time.Stamp)
	h := sql.NullString{
		String: "",
		Valid:  false,
	}
	dnc := sql.NullString{
		String: tm,
		Valid:  true,
	}
	dnm := sql.NullString{
		String: tm,
		Valid:  true,
	}
	p := NullInt64{
		NullInt64: parent,
	}
	dc := NullString{
		NullString: dnc,
	}
	dm := NullString{
		NullString: dnm,
	}
	his := NullString{
		NullString: h,
	}
    j1 := Datatypes {
        DatatypeID: 1,
        ParentID: parent,
        Label: "Page",
        Type: "ROOT", 
        AuthorID: 1,
        DateCreated: dnc,
        DateModified: dnm,
    }
	b1, err := json.Marshal(j1)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(b1))

	j := DatatypeJSON{
		DatatypeID:   1,
		ParentID:     p,
		Label:        "Page",
		Type:         "ROOT",
        AuthorID:     1,
		DateCreated:  dc,
		DateModified: dm,
		History:      his,
	}

	b, err := json.Marshal(j)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(b))

}

func TestDatatypeUnmarshal(t *testing.T) {
	s := `{"datatype_id":1,"parent_id":null,"label":"Page","type":"ROOT","author_id":1,"date_created":"Apr  1 11:34:24","date_modified":"Apr  1 11:34:24","history":null}`
	j := DatatypeJSON{}
    j1:= Datatypes{}
	err := json.Unmarshal([]byte(s), &j1)
	if err != nil {
        fmt.Println(err.Error())
        t.Fail()
	}
	fmt.Println(j1)
	err = json.Unmarshal([]byte(s), &j)
	if err != nil {
        fmt.Println(err.Error())
        t.Fail()
	}
	fmt.Println(j)

}
