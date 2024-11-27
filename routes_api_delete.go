package main

import (
	"encoding/json"
	"net/http"
	"strconv"
)

func apiDeleteAdminRoute(w http.ResponseWriter, r *http.Request) error {
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
	res := dbDeleteAdminRoute(db, ctx, adminRouteSlug)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(map[string]string{"res": res})
	if err != nil {
		logError("failed to : ", err)
	}
	return nil
}

func apiDeleteDataType(w http.ResponseWriter, r *http.Request) error {
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
	res := dbDeleteDataType(db, ctx, intDatatypeId)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(map[string]string{"res": res})
	if err != nil {
		logError("failed to : ", err)
	}
	return nil
}

func apiDeleteField(w http.ResponseWriter, r *http.Request) error {
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
	res := dbDeleteField(db, ctx, intFieldId)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(map[string]string{"res": res})
	if err != nil {
		logError("failed to : ", err)
	}
	return nil
}

func apiDeleteMedia(w http.ResponseWriter, r *http.Request) error {
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
	res := dbDeleteMedia(db, ctx, intMediaId)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(map[string]string{"res": res})
	if err != nil {
		logError("failed to : ", err)
	}
	return nil
}

func apiDeleteMediaDimension(w http.ResponseWriter, r *http.Request) error {
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
	res := dbDeleteMediaDimension(db, ctx, intMediaDimensionId)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(map[string]string{"res": res})
	if err != nil {
		logError("failed to : ", err)
	}
	return nil
}

func apiDeleteRoute(w http.ResponseWriter, r *http.Request) error {
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
	res := dbDeleteRoute(db, ctx, routeSlug)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(map[string]string{"res": res})
	if err != nil {
		logError("failed to : ", err)
	}
	return nil
}

func apiDeleteTable(w http.ResponseWriter, r *http.Request) error {
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
	tableId := r.FormValue("id")
	if err != nil {
		return err
	}
	intTableId, err := strconv.ParseInt(tableId, 10, 64)
	if err != nil {
		logError("failed to : ", err)
	}
	res := dbDeleteTable(db, ctx, intTableId)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(map[string]string{"res": res})
	if err != nil {
		logError("failed to : ", err)
	}
	return nil
}

func apiDeleteToken(w http.ResponseWriter, r *http.Request) error {
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
	res := dbDeleteToken(db, ctx, intTokenId)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(map[string]string{"res": res})
	if err != nil {
		logError("failed to : ", err)
	}
	return nil
}

func apiDeleteUser(w http.ResponseWriter, r *http.Request) error {
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
	res := dbDeleteUser(db, ctx, intUserId)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(map[string]string{"res": res})
    if err != nil { 
        logError("failed to : ", err)
    }
	return nil
}
