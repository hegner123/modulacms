package db
/*
import (
	"fmt"
	"testing"
)

func TestHistoryMap(t *testing.T) {
	a := AdminContentData{
		AdminContentDataID: 1,
		AdminRouteID:       1,
		AdminDatatypeID:    2,
		History: Ns(`[
              {
                "admin_content_data_id": 1,
                "admin_route_id": 10,
                "admin_datatype_id": 100,
                "date_created": "2024-01-01T00:00:00Z",
                "date_modified": "2024-01-02T00:00:00Z"
              },
              {
                "admin_content_data_id": 1,
                "admin_route_id": 11,
                "admin_datatype_id": 101,
                "date_created": "2024-02-01T00:00:00Z",
                "date_modified": "2024-02-02T00:00:00Z"
              }
            ]`,
		),
		DateCreated:  Ns("2024-02-01T00:00:00Z"),
		DateModified: Ns("2024-02-02T00:00:00Z"),
	}

	h, err := MapHistory(a)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(h)

}

func TestPopHistory(t *testing.T) {
	a := AdminContentData{
		AdminContentDataID: 1,
		AdminRouteID:       1,
		AdminDatatypeID:    2,
		History: Ns(`[
              {
                "admin_content_data_id": 1,
                "admin_route_id": 10,
                "admin_datatype_id": 100,
                "date_created": "2024-01-01T00:00:00Z",
                "date_modified": "2024-01-02T00:00:00Z"
              },
              {
                "admin_content_data_id": 1,
                "admin_route_id": 11,
                "admin_datatype_id": 101,
                "date_created": "2024-02-01T00:00:00Z",
                "date_modified": "2024-02-02T00:00:00Z"
              },
              {
                "admin_content_data_id": 1,
                "admin_route_id": 1,
                "admin_datatype_id": 2,
                "date_created": "2024-02-01T00:00:00Z",
                "date_modified": "2024-02-02T00:00:00Z"
              }
            ]`,
		),
		DateCreated:  Ns("2024-02-01T00:00:00Z"),
		DateModified: Ns("2024-02-02T00:00:00Z"),
	}
	_, err := PopHistory(a)
	if err != nil {
		t.Fatal(err)
	}

}
*/
