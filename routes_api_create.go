package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	mdb "github.com/hegner123/modulacms/db-sqlite"
)

func apiCreateAdminRoute(w http.ResponseWriter, r *http.Request) {
	db, ctx, err := getDb(Database{})
	if err != nil {
		logError("failed to get database: ", err)
	}
	err = r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
	}
	newAdminRoute := mdb.CreateAdminRouteParams{}
	jsonUser := formMapJson(r)
	json.Unmarshal(jsonUser, newAdminRoute)

	_ = dbCreateAdminRoute(db, ctx, newAdminRoute)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(newAdminRoute)
	if err != nil {
		fmt.Printf("%s\n", err)
	}
}

func apiCreateDatatype(w http.ResponseWriter, r *http.Request) {
	db, ctx, err := getDb(Database{})
	if err != nil {
		logError("failed to get database: ", err)
	}
	datatype := mdb.CreateDatatypeParams{}

	fmt.Printf("admin create datatype\n")
	err = r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
	}
	jsonBytes := formMapJson(r)
	err = json.Unmarshal(jsonBytes, &datatype)
	if err != nil {
		logError("failed to unmarshall", err)
	}
	newDatatype, err := dbCreateDataType(db, ctx, datatype)
	if err != nil {
		logError("failed to create", err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(newDatatype)
	if err != nil {
		fmt.Printf("%s\n", err)
	}
}

func apiCreateField(w http.ResponseWriter, r *http.Request) {
	db, ctx, err := getDb(Database{})
	if err != nil {
		logError("failed to get database: ", err)
	}
	field := mdb.CreateFieldParams{}

	fmt.Printf("admin create field\n")
	err = r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
	}
	jsonBytes := formMapJson(r)
	err = json.Unmarshal(jsonBytes, field)
    if err != nil { 
        logError("failed to unmarshall", err)
    }
	newField, err := dbCreateField(db, ctx, field)
    if err != nil { 
        logError("failed to create failed", err)
    }
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(newField)
	if err != nil {
		fmt.Printf("%s\n", err)
	}
}

func apiCreateMedia(w http.ResponseWriter, r *http.Request) {
	db, ctx, err := getDb(Database{})
	if err != nil {
		logError("failed to get database: ", err)
	}
	err = r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
	}
	newMedia := mdb.CreateMediaParams{}
	jsonUser := formMapJson(r)
	json.Unmarshal(jsonUser, newMedia)

	_ = dbCreateMedia(db, ctx, newMedia)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(newMedia)
	if err != nil {
		fmt.Printf("%s\n", err)
	}
}

func apiCreateMediaDimension(w http.ResponseWriter, r *http.Request) {
	db, ctx, err := getDb(Database{})
	if err != nil {
		logError("failed to get database: ", err)
	}
	err = r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
	}
	newMediaDimension := mdb.CreateMediaDimensionParams{}
	jsonUser := formMapJson(r)
	json.Unmarshal(jsonUser, newMediaDimension)

	_ = dbCreateMediaDimension(db, ctx, newMediaDimension)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(newMediaDimension)
	if err != nil {
		fmt.Printf("%s\n", err)
	}
}

func apiCreateRoute(w http.ResponseWriter, r *http.Request) {
	db, ctx, err := getDb(Database{})
	if err != nil {
		logError("failed to get database: ", err)
	}
	err = r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
	}
	newRoute := mdb.CreateRouteParams{}
	jsonUser := formMapJson(r)
	json.Unmarshal(jsonUser, newRoute)

	_ = dbCreateRoute(db, ctx, newRoute)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(newRoute)
	if err != nil {
		fmt.Printf("%s\n", err)
	}
}

func apiCreateTable(w http.ResponseWriter, r *http.Request) {
	db, ctx, err := getDb(Database{})
	if err != nil {
		logError("failed to get database: ", err)
	}
	err = r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
	}
	l := r.FormValue("label")
	newTable := mdb.Tables{
		Label: ns(l),
	}

	_ = dbCreateTable(db, ctx, newTable)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(newTable)
	if err != nil {
		fmt.Printf("%s\n", err)
	}
}

func apiCreateToken(w http.ResponseWriter, r *http.Request) {
	db, ctx, err := getDb(Database{})
	if err != nil {
		logError("failed to get database: ", err)
	}
	err = r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
	}
	newToken := mdb.CreateTokenParams{}
	jsonUser := formMapJson(r)
	json.Unmarshal(jsonUser, newToken)

	_ = dbCreateToken(db, ctx, newToken)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(newToken)
	if err != nil {
		fmt.Printf("%s\n", err)
	}
}

func apiCreateUser(w http.ResponseWriter, r *http.Request) {
	db, ctx, err := getDb(Database{})
	if err != nil {
		logError("failed to get database: ", err)
	}
	err = r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
	}
	newUser := mdb.CreateUserParams{}
	jsonUser := formMapJson(r)
	json.Unmarshal(jsonUser, newUser)

	_ = dbCreateUser(db, ctx, newUser)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(newUser)
	if err != nil {
		fmt.Printf("%s\n", err)
	}
}
