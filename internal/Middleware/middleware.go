package media

import (
	"net/http"

	mdb "github.com/hegner123/modulacms/db-sqlite"
	db "github.com/hegner123/modulacms/internal/Db"
	utility "github.com/hegner123/modulacms/internal/Utility"
)

func UserIsAuth(u mdb.Users, r *http.Request) bool {
	dbc := db.GetDb(db.Database{})
	tkn, err := db.GetTokenByUserId(dbc.Connection, dbc.Context, u.UserID)
	if err != nil {
		utility.LogError("failed to : ", err)
	}
	if tkn.TokenType == "Access" && !tkn.Revoked.Bool {
		return true
	}
	return false
}
