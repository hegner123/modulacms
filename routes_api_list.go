package main

import (
	"encoding/json"
	"net/http"
)

func apiListAdminFields(w http.ResponseWriter, r *http.Request) error {
	db, ctx, err := getDb(Database{})
	if err != nil {
		return err
	}
	defer db.Close()
	defer ctx.Done()
	fetchedAdminFields := dbListAdminFields(db, ctx)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(fetchedAdminFields)
	return nil
}

func apiListAdminDatatypes(w http.ResponseWriter, r *http.Request) error {
	db, ctx, err := getDb(Database{})
	if err != nil {
		return err
	}
	defer db.Close()
	defer ctx.Done()
	fetchedAdminDatatype := dbListAdminDatatypes(db, ctx)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(fetchedAdminDatatype)
	return nil
}

func apiListAdminRoutes(w http.ResponseWriter, r *http.Request) error {
	db, ctx, err := getDb(Database{})
	if err != nil {
		return err
	}
	defer db.Close()
	defer ctx.Done()
	fetchedAdminRoutes := dbListAdminRoute(db, ctx)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(fetchedAdminRoutes)
	return nil
}

func apiListDatatypes(w http.ResponseWriter, r *http.Request) error {
	db, ctx, err := getDb(Database{})
	if err != nil {
		return err
	}
	defer db.Close()
	defer ctx.Done()
	fetchedDatatypes := dbListDatatype(db, ctx)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(fetchedDatatypes)
	return nil
}

func apiListFields(w http.ResponseWriter, r *http.Request) error {
	db, ctx, err := getDb(Database{})
	if err != nil {
		return err
	}
	defer db.Close()
	defer ctx.Done()
	fetchedField := dbListField(db, ctx)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(fetchedField)
	return nil
}

func apiListMedia(w http.ResponseWriter, r *http.Request) error {
	db, ctx, err := getDb(Database{})
	if err != nil {
		return err
	}
	defer db.Close()
	defer ctx.Done()
	fetchedMedia := dbListMedia(db, ctx)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(fetchedMedia)
	return nil
}

func apiListMediaDimensions(w http.ResponseWriter, r *http.Request) error {
	db, ctx, err := getDb(Database{})
	if err != nil {
		return err
	}
	defer db.Close()
	defer ctx.Done()
	fetchedMediaDimension := dbListMediaDimension(db, ctx)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(fetchedMediaDimension)
	return nil
}

func apiListFieldsForRoute(w http.ResponseWriter, r *http.Request) error {
	/*
		var routeid int64
		db, ctx, err := getDb(Database{})
		if err != nil {
			return err
		}
		key := "routid"
		params, err := parseQueryParams(r)
		if err != nil {
			logError("failed to parseQueryParams : ", err)
			return err
		}

		if value, exists := params[key]; exists {
			fmt.Fprintf(w, "Key '%s' exists with value: %s\n", key, value)
			routeid, err = strconv.ParseInt(params[key], 10, 64)
		} else {
			fmt.Fprintf(w, "Key '%s' does not exist\n", key)
		}
		fs := dbJoinDatatypeByRoute(db, ctx, routeid)
		fields, err := json.Marshal(fs)
		if err != nil {
			logError("failed to Marshal : ", err)
			return err
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		err = json.NewEncoder(w).Encode(fields)
		return nil
	*/
	return nil
}

func apiListRoutes(w http.ResponseWriter, r *http.Request) error {
	db, ctx, err := getDb(Database{})
	if err != nil {
		return err
	}
	defer db.Close()
	defer ctx.Done()
	fetchedRoutes := dbListRoute(db, ctx)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(fetchedRoutes)
	return nil
}

func apiListTables(w http.ResponseWriter, r *http.Request) error {
	db, ctx, err := getDb(Database{})
	if err != nil {
		return err
	}
	defer db.Close()
	defer ctx.Done()
	fetchedTable := dbListTable(db, ctx)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(fetchedTable)
	return nil
}

func apiListUsers(w http.ResponseWriter, r *http.Request) error {
	db, ctx, err := getDb(Database{})
	if err != nil {
		return err
	}
	defer db.Close()
	defer ctx.Done()
	fetchedUsers := dbListUser(db, ctx)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(fetchedUsers)
	return nil
}
