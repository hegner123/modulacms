package permissions

import (
	"encoding/json"

	utility "github.com/hegner123/modulacms/internal/utility"
)

type Permission string

type Role struct {
	Label       string       `json:"label"`
	Permissions []Permission `json:"permissions"`
}

const (
	can_create_admindatatype  Permission = "can_create_admindatatype"
	can_read_admindatatype    Permission = "can_read_admindatatype"
	can_update_admindatatype  Permission = "can_update_admindatatype"
	can_delete_admindatatype  Permission = "can_delete_admindatatype"
	can_create_adminfield     Permission = "can_create_adminfield"
	can_read_adminfield       Permission = "can_read_adminfield"
	can_update_adminfield     Permission = "can_update_adminfield"
	can_delete_adminfield     Permission = "can_delete_adminfield"
	can_create_adminroute     Permission = "can_create_adminroute"
	can_read_adminroute       Permission = "can_read_adminroute"
	can_update_adminroute     Permission = "can_update_adminroute"
	can_delete_adminroute     Permission = "can_delete_adminroute"
	can_create_contentdata    Permission = "can_create_contentdata"
	can_read_contentdata      Permission = "can_read_contentdata"
	can_update_contentdata    Permission = "can_update_contentdata"
	can_delete_contentdata    Permission = "can_delete_contentdata"
	can_create_contentfield   Permission = "can_create_contentfield"
	can_read_contentfield     Permission = "can_read_contentfield"
	can_update_contentfield   Permission = "can_update_contentfield"
	can_delete_contentfield   Permission = "can_delete_contentfield"
	can_create_datatype       Permission = "can_create_datatype"
	can_read_datatype         Permission = "can_read_datatype"
	can_update_datatype       Permission = "can_update_datatype"
	can_delete_datatype       Permission = "can_delete_datatype"
	can_create_field          Permission = "can_create_field"
	can_read_field            Permission = "can_read_field"
	can_update_field          Permission = "can_update_field"
	can_delete_field          Permission = "can_delete_field"
	can_create_media          Permission = "can_create_media"
	can_read_media            Permission = "can_read_media"
	can_update_media          Permission = "can_update_media"
	can_delete_media          Permission = "can_delete_media"
	can_create_mediadimension Permission = "can_create_mediadimension"
	can_read_mediadimension   Permission = "can_read_mediadimension"
	can_update_mediadimension Permission = "can_update_mediadimension"
	can_delete_mediadimension Permission = "can_delete_mediadimension"
	can_create_route          Permission = "can_create_route"
	can_read_route            Permission = "can_read_route"
	can_update_route          Permission = "can_update_route"
	can_delete_route          Permission = "can_delete_route"
	can_create_table          Permission = "can_create_table"
	can_read_table            Permission = "can_read_table"
	can_update_table          Permission = "can_update_table"
	can_delete_table          Permission = "can_delete_table"
	can_create_user           Permission = "can_create_user"
	can_read_user             Permission = "can_read_user"
	can_update_user           Permission = "can_update_user"
	can_delete_user           Permission = "can_delete_user"
	can_read_own_user         Permission = "can_read_own_user"
	can_update_own_user       Permission = "can_update_own_user"
	can_delete_own_user       Permission = "can_delete_own_user"
)

var AdminRole Role = Role{
	Label: "Admin",
	Permissions: []Permission{
		can_create_admindatatype,
		can_read_admindatatype,
		can_update_admindatatype,
		can_delete_admindatatype,
		can_create_adminfield,
		can_read_adminfield,
		can_update_adminfield,
		can_delete_adminfield,
		can_create_adminroute,
		can_read_adminroute,
		can_update_adminroute,
		can_delete_adminroute,
		can_create_contentdata,
		can_read_contentdata,
		can_update_contentdata,
		can_delete_contentdata,
		can_create_contentfield,
		can_read_contentfield,
		can_update_contentfield,
		can_delete_contentfield,
		can_create_datatype,
		can_read_datatype,
		can_update_datatype,
		can_delete_datatype,
		can_create_field,
		can_read_field,
		can_update_field,
		can_delete_field,
		can_create_media,
		can_read_media,
		can_update_media,
		can_delete_media,
		can_create_mediadimension,
		can_read_mediadimension,
		can_update_mediadimension,
		can_delete_mediadimension,
		can_create_route,
		can_read_route,
		can_update_route,
		can_delete_route,
		can_create_table,
		can_read_table,
		can_update_table,
		can_delete_table,
		can_create_user,
		can_read_user,
		can_update_user,
		can_delete_user,
		can_read_own_user,
		can_update_own_user,
		can_delete_own_user,
	},
}

func (r Role) JSON() []byte {
	j, err := json.Marshal(AdminRole)
	if err != nil {
        utility.DefaultLogger.Error("",err)
	}
	return j
}
