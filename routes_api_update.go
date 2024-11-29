package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	mdb "github.com/hegner123/modulacms/db-sqlite"
)

func apiUpdateAdminRoute(w http.ResponseWriter, r *http.Request) error {
	db, ctx, err := getDb(Database{})
	if err != nil {
		logError("failed to get database: ", err)
	}
	err = r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
	}
	routeSlug := r.FormValue("slug")
	if err != nil {
		return err
	}
	updatedAdminRoute := mdb.UpdateAdminRouteParams{}
	jsonAdminRoute := formMapJson(r)
	json.Unmarshal(jsonAdminRoute, updatedAdminRoute)
	updatedAdminRoute.Slug_2 = routeSlug
	_ = dbUpdateAdminRoute(db, ctx, updatedAdminRoute)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	err = json.NewEncoder(w).Encode(updatedAdminRoute)
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	return nil
}

func apiUpdateDatatype(w http.ResponseWriter, r *http.Request) error {
	db, ctx, err := getDb(Database{})
	if err != nil {
		logError("failed to get database: ", err)
	}
	err = r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
	}
	datatypeId := r.FormValue("id")
	if err != nil {
		return err
	}
	intDatatypeId, err := strconv.ParseInt(datatypeId, 10, 64)
	if err != nil {
		logError("failed to : ", err)
		return err
	}
	updatedDatatype := mdb.UpdateDatatypeParams{}
	jsonDatatype := formMapJson(r)
	json.Unmarshal(jsonDatatype, updatedDatatype)
	updatedDatatype.DatatypeID = intDatatypeId
	_ = dbUpdateDatatype(db, ctx, updatedDatatype)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	err = json.NewEncoder(w).Encode(updatedDatatype)
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	return nil
}

func apiUpdateField(w http.ResponseWriter, r *http.Request) error {
	db, ctx, err := getDb(Database{})
	if err != nil {
		logError("failed to get database: ", err)
	}
	err = r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
	}
	fieldId := r.FormValue("id")
	if err != nil {
		return err
	}
	intFieldId, err := strconv.ParseInt(fieldId, 10, 64)
	if err != nil {
		logError("failed to : ", err)
		return err
	}
	updatedField := mdb.UpdateFieldParams{}
	jsonField := formMapJson(r)
	json.Unmarshal(jsonField, updatedField)
	updatedField.FieldID = intFieldId
	_ = dbUpdateField(db, ctx, updatedField)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	err = json.NewEncoder(w).Encode(updatedField)
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	return nil
}

func apiUpdateMedia(w http.ResponseWriter, r *http.Request) error {
	db, ctx, err := getDb(Database{})
	if err != nil {
		logError("failed to get database: ", err)
	}
	err = r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
	}
	mediaId := r.FormValue("id")
	if err != nil {
		return err
	}
	intMediaId, err := strconv.ParseInt(mediaId, 10, 64)
	if err != nil {
		logError("failed to : ", err)
		return err
	}
	updatedMedia := mdb.UpdateMediaParams{}
	jsonMedia := formMapJson(r)
	json.Unmarshal(jsonMedia, updatedMedia)
	updatedMedia.ID = intMediaId
	_ = dbUpdateMedia(db, ctx, updatedMedia)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	err = json.NewEncoder(w).Encode(updatedMedia)
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	return nil
}

func apiUpdateMediaDimension(w http.ResponseWriter, r *http.Request) error {
	db, ctx, err := getDb(Database{})
	if err != nil {
		logError("failed to get database: ", err)
	}
	err = r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
	}
	mediaDimensionId := r.FormValue("id")
	if err != nil {
		return err
	}
	intMediaDimensionId, err := strconv.ParseInt(mediaDimensionId, 10, 64)
	if err != nil {
		logError("failed to : ", err)
		return err
	}
	updatedMediaDimension := mdb.UpdateMediaDimensionParams{}
	jsonMediaDimension := formMapJson(r)
	json.Unmarshal(jsonMediaDimension, updatedMediaDimension)
	updatedMediaDimension.ID = intMediaDimensionId
	_ = dbUpdateMediaDimension(db, ctx, updatedMediaDimension)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	err = json.NewEncoder(w).Encode(updatedMediaDimension)
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	return nil
}

func apiUpdateRoute(w http.ResponseWriter, r *http.Request) error {
	db, ctx, err := getDb(Database{})
	if err != nil {
		logError("failed to get database: ", err)
	}
	err = r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
	}
	routeSlug := r.FormValue("slug")
	if err != nil {
		return err
	}
	updatedRoute := mdb.UpdateRouteParams{}
	jsonRoute := formMapJson(r)
	json.Unmarshal(jsonRoute, updatedRoute)
	updatedRoute.Slug_2 = routeSlug
	_ = dbUpdateRoute(db, ctx, updatedRoute)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	err = json.NewEncoder(w).Encode(updatedRoute)
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	return nil
}

func apiUpdateTables(w http.ResponseWriter, r *http.Request) error {
	db, ctx, err := getDb(Database{})
	if err != nil {
		logError("failed to get database: ", err)
	}
	err = r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
	}
	tableId := r.FormValue("id")
	if err != nil {
		return err
	}
	intTablesId, err := strconv.ParseInt(tableId, 10, 64)
	if err != nil {
		logError("failed to : ", err)
		return err
	}
	updatedTables := mdb.UpdateTableParams{}
	jsonTables := formMapJson(r)
	json.Unmarshal(jsonTables, updatedTables)
	updatedTables.ID = intTablesId
	_ = dbUpdateTable(db, ctx, updatedTables)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	err = json.NewEncoder(w).Encode(updatedTables)
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	return nil
}

func apiUpdateToken(w http.ResponseWriter, r *http.Request) error {
	db, ctx, err := getDb(Database{})
	if err != nil {
		logError("failed to get database: ", err)
	}
	err = r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
	}
	tokenId := r.FormValue("id")
	if err != nil {
		return err
	}
	intTokenId, err := strconv.ParseInt(tokenId, 10, 64)
	if err != nil {
		logError("failed to : ", err)
		return err
	}
	updatedToken := mdb.UpdateTokenParams{}
	jsonToken := formMapJson(r)
	json.Unmarshal(jsonToken, updatedToken)
	updatedToken.ID = intTokenId
	_ = dbUpdateToken(db, ctx, updatedToken)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	err = json.NewEncoder(w).Encode(updatedToken)
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	return nil
}

func apiUpdateUser(w http.ResponseWriter, r *http.Request) error {
	db, ctx, err := getDb(Database{})
	if err != nil {
		logError("failed to get database: ", err)
	}
	err = r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
	}
	userId := r.FormValue("id")
	if err != nil {
		return err
	}
	intUserId, err := strconv.ParseInt(userId, 10, 64)
	if err != nil {
		logError("failed to : ", err)
		return err
	}
	updatedUser := mdb.UpdateUserParams{}
	jsonUser := formMapJson(r)
	json.Unmarshal(jsonUser, updatedUser)
	updatedUser.UserID = intUserId
	_ = dbUpdateUser(db, ctx, updatedUser)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	err = json.NewEncoder(w).Encode(updatedUser)
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	return nil
}
