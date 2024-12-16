package main

import "fmt"

func checkInstallStatus(database string) bool {
	var userExists, routeExists bool
	db, ctx, err := getDb(Database{src: database})
	if err != nil {
		logError("failed to get db", err)
	}
	defer db.Close()
	userExists = false
	routeExists = false

	userCount := countUsers(db, ctx)
	fmt.Printf("userCount :%d\n", userCount)
	adminRoutes := countAdminRoutes(db, ctx)

	if userCount > 0 {
		userExists = true
	}
	if adminRoutes > 0 {
		routeExists = true
	}

	if userExists && routeExists {
		return true
	}
	return false
}
