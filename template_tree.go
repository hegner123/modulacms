package main

func BuildTemplateStructFromRouteId(id int64, dbPath string) (*Tree, error) {
	dbName := "test.db"
	db, ctx, err := getDb(Database{src: dbName})
	if err != nil {
		return nil, err
	}
	defer db.Close()

	global := dbGetAdminDatatypeGlobalId(db, ctx)
	rows := dbGetChildren(make([]int64, 0), int(global.AdminDtID), dbName, "origin")

	t1 := NewTree(Scan)

	for i := 0; i < len(rows); i++ {
		dt := dbGetAdminDatatypeById(db, ctx, rows[i])
		t1.Add(dt)
	}

	t1.Root.AddTreeFields("")
	return t1, nil
}
