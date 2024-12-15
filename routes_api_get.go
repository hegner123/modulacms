package main

import (
	"encoding/json"
	"net/http"
	"strconv"
)

func apiGetAdminDatatype(w http.ResponseWriter, r *http.Request) error {
	db, ctx, err := getDb(Database{})
	if err != nil {
		return err
	}
	defer db.Close()
	err = r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return err
	}
	adminDatatypeIdInput := r.FormValue("id")
	if err != nil {
		return err
	}
	adminDatatypeId, err := strconv.ParseInt(adminDatatypeIdInput, 10, 64)
	fetchedAdminDatatype := dbGetAdminDatatypeById(db, ctx, adminDatatypeId)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(fetchedAdminDatatype)
	if err != nil {
		logError("failed to encode json ", err)
	}
	return nil
}

func apiGetAdminField(w http.ResponseWriter, r *http.Request) error {
	db, ctx, err := getDb(Database{})
	if err != nil {
		return err
	}
	defer db.Close()
	err = r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return err
	}
	adminFieldInput := r.FormValue("id")
	if err != nil {
		return err
	}
	adminFieldId, err := strconv.ParseInt(adminFieldInput, 10, 64)
	fetchedAdminField := dbGetAdminDatatypeById(db, ctx, adminFieldId)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(fetchedAdminField)
	if err != nil {
		logError("failed to encode json ", err)
	}
	return nil
}

func apiGetAdminRoute(w http.ResponseWriter, r *http.Request) error {
	db, ctx, err := getDb(Database{})
	if err != nil {
		return err
	}
	defer db.Close()
	err = r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return err
	}
	adminRouteSlug := r.FormValue("slug")
	if err != nil {
		return err
	}
	fetchedAdminRoute := dbGetAdminRoute(db, ctx, adminRouteSlug)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(fetchedAdminRoute)
	if err != nil {
		logError("failed to encode json ", err)
	}
	return nil
}

func apiGetDatatype(w http.ResponseWriter, r *http.Request) error {
	db, ctx, err := getDb(Database{})
	if err != nil {
		return err
	}
	defer db.Close()
	err = r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return err
	}
	datatypeId := r.FormValue("id")
	if err != nil {
		return err
	}
	intDatatypeId, err := strconv.ParseInt(datatypeId, 10, 64)
	if err != nil {
		logError("failed to : ", err)
	}
	fetchedDatatype := dbGetDatatype(db, ctx, intDatatypeId)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(fetchedDatatype)
	return nil
}

func apiGetField(w http.ResponseWriter, r *http.Request) error {
	db, ctx, err := getDb(Database{})
	if err != nil {
		return err
	}
	defer db.Close()
	err = r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return err
	}
	fieldId := r.FormValue("id")
	if err != nil {
		return err
	}
	intFieldId, err := strconv.ParseInt(fieldId, 10, 64)
	if err != nil {
		logError("failed to : ", err)
	}
	fetchedField := dbGetField(db, ctx, intFieldId)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(fetchedField)
	return nil
}

func apiGetMedia(w http.ResponseWriter, r *http.Request) error {
	db, ctx, err := getDb(Database{})
	if err != nil {
		return err
	}
	defer db.Close()
	err = r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return err
	}
	mediaId := r.FormValue("id")
	if err != nil {
		return err
	}
	intMediaId, err := strconv.ParseInt(mediaId, 10, 64)
	if err != nil {
		logError("failed to : ", err)
	}
	fetchedMedia := dbGetMedia(db, ctx, intMediaId)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(fetchedMedia)
	return nil
}

func apiGetMediaDimension(w http.ResponseWriter, r *http.Request) error {
	db, ctx, err := getDb(Database{})
	if err != nil {
		return err
	}
	defer db.Close()
	err = r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return err
	}
	mediaDimensionId := r.FormValue("id")
	if err != nil {
		return err
	}
	intMediaDimensionId, err := strconv.ParseInt(mediaDimensionId, 10, 64)
	if err != nil {
		logError("failed to : ", err)
	}
	fetchedMediaDimension := dbGetMediaDimension(db, ctx, intMediaDimensionId)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(fetchedMediaDimension)
	return nil
}

func apiGetRoute(w http.ResponseWriter, r *http.Request) error {
	db, ctx, err := getDb(Database{})
	if err != nil {
		return err
	}
	defer db.Close()
	err = r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return err
	}
	routeSlug := r.FormValue("slug")
	if err != nil {
		return err
	}
	fetchedRoute := dbGetRoute(db, ctx, routeSlug)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(fetchedRoute)
	return nil
}

func apiGetTable(w http.ResponseWriter, r *http.Request) error {
	db, ctx, err := getDb(Database{})
	if err != nil {
		return err
	}
	defer db.Close()
	err = r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return err
	}
	TableId := r.FormValue("id")
	if err != nil {
		return err
	}
	intTableId, err := strconv.ParseInt(TableId, 10, 64)
	if err != nil {
		logError("failed to : ", err)
	}
	fetchedTable := dbGetTable(db, ctx, intTableId)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(fetchedTable)
	return nil
}

func apiGetToken(w http.ResponseWriter, r *http.Request) error {
	db, ctx, err := getDb(Database{})
	if err != nil {
		return err
	}
	defer db.Close()
	err = r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return err
	}
	TokenId := r.FormValue("id")
	if err != nil {
		return err
	}
	intTokenId, err := strconv.ParseInt(TokenId, 10, 64)
	if err != nil {
		logError("failed to : ", err)
	}
	fetchedToken := dbGetToken(db, ctx, intTokenId)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(fetchedToken)
	return nil
}

func apiGetUser(w http.ResponseWriter, r *http.Request) error {
	db, ctx, err := getDb(Database{})
	if err != nil {
		return err
	}
	defer db.Close()
	err = r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return err
	}
	userId := r.FormValue("id")
	if err != nil {
		return err
	}
	intUserId, err := strconv.ParseInt(userId, 10, 64)
	if err != nil {
		logError("failed to : ", err)
	}
	fetchedUser, err := dbGetUser(db, ctx, intUserId)
	if err != nil {
		logError("failed to getUser: ", err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(fetchedUser)
	return nil
}
