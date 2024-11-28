package main

func checkInstallStatus() bool {
    var userExists, routeExists bool
	db, ctx, err := getDb(Database{DB: "modula.db"})
	if err != nil {
		logError("failed to get db", err)
	}
    defer db.Close()
    user,err := dbGetUser(db, ctx, 1)
    if err != nil { 
        logError("failed to retrive user record", err)
        return false
    }
    userExists = false
    if user.Email == "system@modulacms.com"{
        userExists = true
    }
	homeRes := dbGetAdminRoute(db, ctx, "/")
    routeExists = false
    if homeRes.Slug == "/"{
        routeExists = true
    }
    if userExists && routeExists {
        return true
    }
	return false
}
