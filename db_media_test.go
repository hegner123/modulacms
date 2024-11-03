package main

import (
	"fmt"
	"testing"
)


func TestAddMedia(t *testing.T){
    db,err := getDb(Database{})
    if err!=nil {
        fmt.Printf("%s\n",err)
    }
    times := timestamp()
    media := Media{Name: "example.png",DisplayName: "Example", Alt: "example",
        Caption: "example", Description: "exmple image for the purposes of testing",Class: "flex-image",
        CreatedBy: 0, DateCreated: times, DateModified: times, Url: "https:localhost/example.png",MimeType: "png",
        Dimensions:"1000x1000" ,OptimizedMobile: "", OptimizedTablet: "", OptimizedDesktop: "", OptimizedUltrawide: ""}
    
    rowsChanged,err :=createMedia(db,media)
    if err!=nil {
        return
    }
    expected := int64(1)
    if rowsChanged!= expected{
        t.Errorf("rows changed does not equal insert statements")
    }

}


func TestDeleteMedia(t *testing.T) {
    db,err := getDb(Database{})
    if err!=nil {
        fmt.Printf("%s\n",err)
    }
    rowsChanged,err :=deleteMediaByName(db,"name","example.png")
    if err!=nil {
        fmt.Printf("%s\n",err)
    }
    expected := int64(1)

    if rowsChanged!= expected{
        t.Errorf("rows changed does not equal insert statements")
    }

}
