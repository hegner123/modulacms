package mEmbed

import (
	"fmt"

	db "github.com/hegner123/modulacms/internal/Db"
	utility "github.com/hegner123/modulacms/internal/Utility"
)

func ReadEmbedFS(directory string) {

	dir, err := db.SqlFiles.ReadDir(directory)
	if err != nil {
		utility.LogError("error in ReadEmbedFS ", err)
	}
	for key, value := range dir {
		fmt.Printf("%d:%s\n", key, value)
	}
}
