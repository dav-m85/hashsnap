package cmd

import (
	"fmt"
	"os"
	"testing"

	"github.com/dav-m85/hashsnap/test"
	"github.com/matryer/is"
)

// func TestMain(m *testing.M) {
// 	var err error
// 	var cleanup func()
// 	DB, cleanup, err = test.DockerDB()
// 	if err != nil {
// 		cleanup()
// 		fmt.Println(err.Error())
// 		os.Exit(1)
// 	}

// 	DB.AutoMigrate(&db.User{}, &db.Content{}, &db.Follow{}, &db.Favorite{})
// 	db.NDB = DB // HORRIBLE
// 	code := m.Run()

// 	cleanup()
// 	os.Exit(code)
// }

func TestTrim(t *testing.T) {
	tmp, teardown := test.Setup(t)
	defer teardown()

	fmt.Println(tmp)
	wd, err := os.Getwd()
	t.Log(wd, err)
	is := is.New(t)
	is.NoErr(nil)
	fmt.Print("Yo")
}
