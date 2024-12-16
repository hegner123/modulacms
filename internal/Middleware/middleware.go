package media

import (
	"net/http"

	mdb "github.com/hegner123/modulacms/db-sqlite"
)

func UserIsAuth(u mdb.Users, r *http.Request) bool {
	db, ctx, err := getDb(Database{})
	if err != nil {
		logError("failed to get database ", err)
	}
	defer db.Close()
	tkn := dbGetTokenByUserId(db, ctx, u.UserID)
	if tkn.TokenType == "Access" && !tkn.Revoked.Bool {
		return true
	}
	return false
}
