package mTemplate

import (
	"fmt"

	db "github.com/hegner123/modulacms/internal/Db"
)

func checkInstallStatus(database string) bool {

	var userExists, routeExists bool
	connectedDb := db.GetDb(db.Database{})
	defer connectedDb.Connection.Close()
	userExists = false
	routeExists = false

	userCount, err := db.CountUsers(connectedDb.Connection, connectedDb.Context)
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	fmt.Printf("userCount :%d\n", userCount)
	adminRoutes, err := db.CountAdminRoutes(connectedDb.Connection, connectedDb.Context)
	if err != nil {
		fmt.Printf("%v\n", err)
	}

	if *userCount > 0 {
		userExists = true
	}
	if *adminRoutes > 0 {
		routeExists = true
	}

	if userExists && routeExists {
		return true
	}
	return false
}
