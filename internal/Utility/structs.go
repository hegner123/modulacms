package main

import mdb "github.com/hegner123/modulacms/db-sqlite"

type DbEndpoints struct{
    Content Content
}

type Content struct{
    AdminDts []mdb.AdminDatatypes
    AdminFields []mdb.AdminFields
    AdminRoutes []mdb.AdminRoutes
}


type Backup struct {
	Hash    string
	DbFile  string
	Archive string
}

type S3Credintials struct {
	AccessKey string
	SecretKey string
	URL       string
}
