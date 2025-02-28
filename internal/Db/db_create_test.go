package db

import (
	"reflect"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	mdb "github.com/hegner123/modulacms/db-sqlite"
)

var CreateTestTable string

func TestCreateDBCopy(t *testing.T) {
	testTable, err := CopyDb("create_tests.db", false)

	if err != nil {
		return
	}
	CreateTestTable = testTable
}

func TestCreateAdminDatatype(t *testing.T) {
	times := TimestampS()
	db := GetDb(Database{Src: CreateTestTable})

	insertedAdminDatatype := CreateAdminDatatype(db.Connection, db.Context, mdb.CreateAdminDatatypeParams{
		AdminRouteID: ni(1),
		ParentID:     ni(1),
		Label:        "label",
		Type:         "text",
		Author:       "systemtest",
		AuthorID:     int64(1),
		DateCreated:  ns(times),
		DateModified: ns(times),
		History:      ns(""),
	})

	expected := mdb.AdminDatatypes{
		AdminRouteID: ni(1),
		ParentID:     ni(1),
		Label:        "label",
		Type:         "text",
		Author:       "systemtest",
		AuthorID:     int64(1),
		DateCreated:  ns(times),
		DateModified: ns(times),
		History:      ns(""),
	}

	if reflect.DeepEqual(insertedAdminDatatype, expected) {
		t.FailNow()
	}
}

func TestCreateAdminField(t *testing.T) {
	times := TimestampS()
	db := GetDb(Database{Src: CreateTestTable})

	insertedAdminField := CreateAdminField(db.Connection, db.Context, mdb.CreateAdminFieldParams{
		AdminRouteID: ni(1),
		ParentID:     ni(1),
		Label:        "label",
		Data:         "Poop on floor and watch human clean up pee in human's bed until he cleans the litter box damn that dog so poop on the floor, break a planter, sprint, eat own hair, vomit hair, hiss, chirp at birds, eat a squirrel, hide from fireworks, lick toe beans, attack christmas tree. ",
		Type:         "text",
		Author:       "systemtest",
		AuthorID:     int64(1),
		DateCreated:  ns(times),
		DateModified: ns(times),
		History:      ns(""),
	})

	expected := mdb.AdminFields{
		AdminRouteID: ni(1),
		ParentID:     ni(1),
		Label:        "label",
		Data:         "Poop on floor and watch human clean up pee in human's bed until he cleans the litter box damn that dog so poop on the floor, break a planter, sprint, eat own hair, vomit hair, hiss, chirp at birds, eat a squirrel, hide from fireworks, lick toe beans, attack christmas tree. ",
		Type:         "text",
		Author:       "systemtest",
		AuthorID:     int64(1),
		DateCreated:  ns(times),
		DateModified: ns(times),
		History:      ns(""),
	}

	if reflect.DeepEqual(insertedAdminField, expected) {
		t.FailNow()
	}
}

func TestCreateAdminRoute(t *testing.T) {
	times := TimestampS()
	db := GetDb(Database{Src: CreateTestTable})

	insertedAdminRoute := CreateAdminRoute(db.Connection, db.Context, mdb.CreateAdminRouteParams{
		Author:       "systemtest",
		AuthorID:     int64(1),
		Slug:         "/test",
		Title:        "Test",
		Status:       0,
		DateCreated:  ns(times),
		DateModified: ns(times),
	})

	expected := mdb.AdminRoutes{
		Author:       "systemtest",
		AuthorID:     1,
		Slug:         "/test",
		Title:        "Test",
		Status:       0,
		DateCreated:  ns(times),
		DateModified: ns(times),
	}

	if reflect.DeepEqual(insertedAdminRoute, expected) {
		t.FailNow()
	}
}
func TestCreateDatatype(t *testing.T) {
	times := TimestampS()
	db := GetDb(Database{Src: CreateTestTable})

	_, err := CreateDataType(db.Connection, db.Context, mdb.CreateDatatypeParams{
		RouteID:      ni(1),
		Label:        "Parent",
		Type:         "text",
		Author:       "systemtest",
		AuthorID:     int64(1),
		DateCreated:  ns(times),
		DateModified: ns(times),
	})
	if err != nil {
		return
	}

	insertedDatatypes, err := CreateDataType(db.Connection, db.Context, mdb.CreateDatatypeParams{
		RouteID:      ni(1),
		ParentID:     ni(1),
		Label:        "title",
		Type:         "text",
		Author:       "systemtest",
		AuthorID:     int64(1),
		DateCreated:  ns(times),
		DateModified: ns(times),
	})
	if err != nil {
		return
	}

	expected := mdb.Datatypes{
		RouteID:      ni(1),
		ParentID:     ni(1),
		Label:        "title",
		Type:         "text",
		Author:       "systemtest",
		AuthorID:     1,
		DateCreated:  ns(times),
		DateModified: ns(times),
	}

	if reflect.DeepEqual(insertedDatatypes, expected) {
		t.FailNow()
	}
}

func TestCreateField(t *testing.T) {
	times := TimestampS()
	db := GetDb(Database{Src: CreateTestTable})

	insertedFields, _ := CreateField(db.Connection, db.Context, mdb.CreateFieldParams{
		RouteID:      ni(1),
		ParentID:     ni(1),
		Label:        "Parent",
		Data:         "Test Field",
		Type:         "text",
		Author:       "systemtest",
		AuthorID:     int64(1),
		DateCreated:  ns(times),
		DateModified: ns(times),
	})
	expected := mdb.Fields{
		RouteID:      ni(1),
		ParentID:     ni(1),
		Label:        "Parent",
		Data:         "Test Field",
		Type:         "text",
		Author:       "systemtest",
		AuthorID:     int64(1),
		DateCreated:  ns(times),
		DateModified: ns(times),
	}

	if reflect.DeepEqual(insertedFields, expected) {
		t.FailNow()
	}
}

func TestCreateContentData(t *testing.T) {
	times := TimestampS()
	db := GetDb(Database{Src: CreateTestTable})

	insertedContentData := CreateContentData(db.Connection, db.Context, mdb.CreateContentDataParams{
		AdminDtID:    int64(1),
		DateCreated:  ns(times),
		DateModified: ns(times),
		History:      ns(""),
	})

	expected := mdb.ContentData{
		AdminDtID:    int64(1),
		DateCreated:  ns(times),
		DateModified: ns(times),
		History:      ns(""),
	}

	if reflect.DeepEqual(insertedContentData, expected) {
		t.FailNow()
	}
}
func TestCreateContentField(t *testing.T) {
	times := TimestampS()
	db := GetDb(Database{Src: CreateTestTable})

	insertedContentField := CreateContentField(db.Connection, db.Context, mdb.CreateContentFieldParams{
		ContentFieldID: int64(1),
		ContentDataID:  int64(1),
		AdminFieldID:   int64(1),
		FieldValue:     "Poop on floor and watch human clean up pee in human's bed until he cleans the litter box damn that dog so poop on the floor, break a planter, sprint, eat own hair, vomit hair, hiss, chirp at birds, eat a squirrel, hide from fireworks, lick toe beans, attack christmas tree. ",
		History:        ns(""),
		DateCreated:    ns(times),
		DateModified:   ns(times),
	})

	expected := mdb.ContentFields{
		ContentFieldID: int64(1),
		ContentDataID:  int64(1),
		AdminFieldID:   int64(1),
		FieldValue:     "Poop on floor and watch human clean up pee in human's bed until he cleans the litter box damn that dog so poop on the floor, break a planter, sprint, eat own hair, vomit hair, hiss, chirp at birds, eat a squirrel, hide from fireworks, lick toe beans, attack christmas tree. ",
		History:        ns(""),
		DateCreated:    ns(times),
		DateModified:   ns(times),
	}

	if reflect.DeepEqual(insertedContentField, expected) {
		t.FailNow()
	}
}

func TestCreateMedia(t *testing.T) {
	times := TimestampS()
	db := GetDb(Database{Src: CreateTestTable})

	insertedMedia := CreateMedia(db.Connection, db.Context, mdb.CreateMediaParams{
		Name:               ns("test.png"),
		DisplayName:        ns("Test"),
		Alt:                ns("test"),
		Caption:            ns("test"),
		Description:        ns("test"),
		Author:             "systemtest",
		AuthorID:           1,
		DateCreated:        ns(times),
		DateModified:       ns(times),
		Url:                ns("public/2024/11/test.png"),
		Mimetype:           ns("image/png"),
		Dimensions:         ns("1000x1000"),
		OptimizedMobile:    ns("public/2024/11/test-mobile.png"),
		OptimizedTablet:    ns("public/2024/11/test-tablet.png"),
		OptimizedDesktop:   ns("public/2024/11/test-desktop.png"),
		OptimizedUltraWide: ns("public/2024/11/test-ultra.png"),
	})

	expected := mdb.Media{
		Name:               ns("test.png"),
		DisplayName:        ns("Test"),
		Alt:                ns("test"),
		Caption:            ns("test"),
		Description:        ns("test"),
		Author:             "systemtest",
		AuthorID:           1,
		DateCreated:        ns(times),
		DateModified:       ns(times),
		Url:                ns("public/2024/11/test.png"),
		Mimetype:           ns("image/png"),
		Dimensions:         ns("1000x1000"),
		OptimizedMobile:    ns("public/2024/11/test-mobile.png"),
		OptimizedTablet:    ns("public/2024/11/test-tablet.png"),
		OptimizedDesktop:   ns("public/2024/11/test-desktop.png"),
		OptimizedUltraWide: ns("public/2024/11/test-ultra.png"),
	}

	if reflect.DeepEqual(insertedMedia, expected) {
		t.FailNow()
	}
}

func TestCreateMediaDimension(t *testing.T) {
	db := GetDb(Database{Src: CreateTestTable})

	insertedMediaDimension := CreateMediaDimension(db.Connection, db.Context, mdb.CreateMediaDimensionParams{
		Label:       ns("Desktop"),
		Width:       ni(1920),
		Height:      ni(1080),
		AspectRatio: ns("16:9"),
	})

	expected := mdb.MediaDimensions{
		Label:       ns("Desktop"),
		Width:       ni(1920),
		Height:      ni(1080),
		AspectRatio: ns("16:9"),
	}

	if reflect.DeepEqual(insertedMediaDimension, expected) {
		t.FailNow()
	}
}
func TestCreateRole(t *testing.T) {
	db := GetDb(Database{Src: CreateTestTable})

	insertedRole := db.CreateRole(db.Connection, db.Context, mdb.CreateRoleParams{
		Label: "admin",
	})

	expected := mdb.Roles{
		Label: "admin",
	}

	if reflect.DeepEqual(insertedRole, expected) {
		t.FailNow()
	}
}

func TestCreateRoute(t *testing.T) {
	times := TimestampS()
	db := GetDb(Database{Src: CreateTestTable})

	insertedRoute := CreateRoute(db.Connection, db.Context, mdb.CreateRouteParams{
		Author:       "systemtest",
		AuthorID:     1,
		Slug:         "/test",
		Title:        "Test",
		Status:       0,
		DateCreated:  ns(times),
		DateModified: ns(times),
	})

	expected := mdb.Routes{
		Author:       "systemtest",
		AuthorID:     1,
		Slug:         "/test",
		Title:        "Test",
		Status:       0,
		DateCreated:  ns(times),
		DateModified: ns(times),
	}

	if reflect.DeepEqual(insertedRoute, expected) {
		t.FailNow()
	}
}

func TestCreateTables(t *testing.T) {
	db := GetDb(Database{Src: CreateTestTable})

	insertedTables := CreateTable(db.Connection, db.Context, mdb.Tables{Label: ns("Test")})

	expected := mdb.Tables{
		Label: ns("Test"),
	}

	if reflect.DeepEqual(insertedTables, expected) {
		t.FailNow()
	}
}

func TestCreateToken(t *testing.T) {
	db := GetDb(Database{Src: CreateTestTable})
	var (
		key         []byte
		tk          *jwt.Token
		signedToken string
	)

	key = GenerateKey()
	tk = jwt.New(jwt.SigningMethodHS256)
	signedToken, err := tk.SignedString(key)
	if err != nil {
		return
	}

	_, err = time.ParseDuration("24h")
	if err != nil {
		return
	}

	times := TimestampS()

	insertedToken := CreateToken(db.Connection, db.Context, mdb.CreateTokenParams{
		UserID:    1,
		IssuedAt:  times,
		ExpiresAt: TimestampS(),
		TokenType: "refresh",
		Token:     signedToken,
		Revoked:   nb(false),
	})

	expected := mdb.Tokens{
		UserID:    1,
		IssuedAt:  times,
		ExpiresAt: TimestampS(),
		TokenType: "refresh",
		Token:     signedToken,
		Revoked:   nb(false),
	}

	if reflect.DeepEqual(insertedToken, expected) {
		t.FailNow()
	}
}

func TestCreateUser(t *testing.T) {
	times := TimestampS()
	db := GetDb(Database{Src: CreateTestTable})

	insertedUser := CreateUser(db.Connection, db.Context, mdb.CreateUserParams{
		DateCreated:  ns(times),
		DateModified: ns(times),
		Username:     "systemtest",
		Name:         "systemtest",
		Email:        "test2@modulacmstest.com",
		Hash:         "has",
		Role:         int64(1),
	})

	expected := mdb.Users{
		DateCreated:  ns(times),
		DateModified: ns(times),
		Username:     "systemtest",
		Name:         "systemtest",
		Email:        "test2@modulacmstest.com",
		Hash:         "has",
		Role:         int64(1),
	}

	if reflect.DeepEqual(insertedUser, expected) {
		t.FailNow()
	}
}
