package media

import (
	"testing"

	config "github.com/hegner123/modulacms/internal/Config"
)

func TestOptimize(t *testing.T) {
	c := config.Config{
		Db_Driver: "sqlite",
		Db_Name:   "modula.db",
		Db_URL:    "./modula.db",
	}
	 OptimizeUpload("./test.png", "test.png", c)

}
