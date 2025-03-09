package auth

import (
	"fmt"
	"strconv"
	"time"

	config "github.com/hegner123/modulacms/internal/Config"
	db "github.com/hegner123/modulacms/internal/Db"
)

type TokenPackage struct {
	AccessToken  string
	RefreshToken string
}

func CreateSessionTokens(userId int64, c config.Config) (*TokenPackage, error) {
	d := db.ConfigDB(c)
	issued := fmt.Sprint(time.Now().Unix())
	expires := fmt.Sprint(time.Now().AddDate(0, 0, 7).Unix())
	s := db.CreateSessionParams{
		UserID:    userId,
		CreatedAt: db.Ns(issued),
		ExpiresAt: db.Ns(expires),
	}
	session, err := d.CreateSession(s)
	if err != nil {
		return nil, err
	}

	us, err := d.GetSessionsByUserId(userId)
	if err != nil {
		return nil, err
	}
	DeleteExpiredSessions(d, *us)

	r, err := GenerateRefreshToken(userId, session.SessionID, c)
	if err != nil {
		return nil, err
	}
	a, err := GenerateAccessToken(userId, session.SessionID, c)
	if err != nil {
		return nil, err
	}
	res := TokenPackage{
		AccessToken:  a,
		RefreshToken: r,
	}
	return &res, nil
}

func CheckSession(id int64, c config.Config) bool {
	d := db.ConfigDB(c)
	s, err := d.GetSession(id)
	if err != nil {
		return false
	}
	timestampInt, err := strconv.ParseInt(s.ExpiresAt.String, 10, 64)
	if err != nil {
		fmt.Println("Error parsing timestamp:", err)
		return false
	}
	t := time.Unix(timestampInt, 0)
   
    // time.now is not after session expiration
    // true false
	return !time.Now().After(t)
}

func DeleteExpiredSessions(d db.DbDriver, s []db.Sessions) {
	for _, v := range s {
		DeleteExpiredSession(d, v)
	}
}
func DeleteExpiredSession(d db.DbDriver, s db.Sessions) {
	timestampInt, err := strconv.ParseInt(s.ExpiresAt.String, 10, 64)
	if err != nil {
		fmt.Println("Error parsing timestamp:", err)
		return
	}
	t := time.Unix(timestampInt, 0)
	if time.Now().After(t) {
		err := d.DeleteSession(s.SessionID)
		if err != nil {
			return
		}
	}
}
