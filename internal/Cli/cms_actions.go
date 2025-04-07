package cli

import (
	"fmt"

	config "github.com/hegner123/modulacms/internal/Config"
	db "github.com/hegner123/modulacms/internal/Db"
)

func CreateDatatypeDefinition(dt db.CreateDatatypeParams, f []db.CreateFieldParams, c config.Config) error {
	d := db.ConfigDB(c)
	datatype := d.CreateDatatype(dt)
	result := []db.Fields{}
	for _, v := range f {
		v.ParentID = db.Int64ToNullInt64(datatype.DatatypeID)
		r := d.CreateField(v)
		result = append(result, r)
	}
	if len(result) < len(f) {
		err := fmt.Errorf("RESULT LENGTH LESS THAN PASSED FIELDS")
		return err
	}
	return nil
}

func CreateDatatypeInstance(cd db.CreateContentDataParams, cf []db.CreateContentFieldParams, c config.Config) error {
	d := db.ConfigDB(c)
	datatype := d.CreateContentData(cd)
	result := []db.ContentFields{}
	for _, v := range cf {
		v.ContentDataID = datatype.ContentDataID
		r := d.CreateContentField(v)
		result = append(result, r)
	}
	if len(result) < len(cf) {
		err := fmt.Errorf("RESULT LENGTH LESS THAN PASSED CONTENT FIELDS")
		return err
	}
	return nil

}
