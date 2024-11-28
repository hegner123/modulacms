package main

func checkInstallStatus() bool {
	var userExists, routeExists bool
	db, ctx, err := getDb(Database{})
	if err != nil {
		logError("failed to get db", err)
	}
	defer db.Close()
	userExists = false
    routeExists = false

	user, err := dbGetUser(db, ctx, int64(1))
	if err != nil {
		logError("failed to confirm system user install", err)
	}
	homeRes := dbGetAdminRoute(db, ctx, "/")


	if user.Email == "system@modulacms.com" {
		userExists = true
	}
	if homeRes.Slug == "/" {
		routeExists = true
	}


	if userExists && routeExists {
		return true
	}
	return false
}
