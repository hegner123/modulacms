package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"

	modula_db "github.com/hegner123/modulacms/internal/Db"
)

func LogGetVersion() string {
	file, err := os.Open("version.json")
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	defer file.Close()
	bytes, err := io.ReadAll(file)
	if err != nil {
		return "Error reading file:"
	}
	return string(bytes)
}

func PopError(err error) string {
	unwrappedErr := strings.Split(err.Error(), " ")
	msg := fmt.Sprint(unwrappedErr[len(unwrappedErr)-1])
	return msg
}

func LogError(message string, err error, args ...any) {
	var messageParts []string
	messageParts = append(messageParts, message)
	for _, arg := range args {
		switch v := arg.(type) {
		case fmt.Stringer: // If the type implements fmt.Stringer, use String()
			messageParts = append(messageParts, v.String())
		default:
			messageParts = append(messageParts, fmt.Sprintf("%+v", arg)) // Format structs nicely
		}
	}
	fullMessage := strings.Join(messageParts, " ")

	// Format the final error message
	er := fmt.Errorf("%sErr: %s\n%v\n%s", RED, fullMessage, err, RESET)
	if er != nil {
		fmt.Printf("%s\n", er)
	}
}

func pLog(args ...any) {
	fmt.Printf("%s", BLUE)
	for _, arg := range args {
		fmt.Print(arg)
	}
	fmt.Printf("%s\n", RESET)
}

func logDb(dbName string) {
    
	db, ctx, err := modula_db.getDb(modula_db.Database{src: dbName})
	if err != nil {
		logError("failed to : ", err)
	}
	adminroutes := dbListAdminRoute(db, ctx)
	datatypes := dbListDatatype(db, ctx)
	users := dbListUser(db, ctx)
	fields := dbListField(db, ctx)
	routes := dbListRoute(db, ctx)
    pLog(users)
	pLog(adminroutes)
    pLog(datatypes)
    pLog(fields)
    pLog(routes)
}


func logStruct(struc any){
jsonStr, err := json.Marshal(struc)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println(string(jsonStr))
}
func PrintStringFields(v interface{}) {
	val := reflect.ValueOf(v)

	// Ensure we're working with a struct or pointer to a struct
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		fmt.Println("Error: input must be a struct or a pointer to a struct")
		return
	}

	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := val.Field(i)

        if field.Name == "AdminDtID"{
            fmt.Printf("\n%s : %d\n",field.Name, fieldValue.Int())
        }

	}
}
