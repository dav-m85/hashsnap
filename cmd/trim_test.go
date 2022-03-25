package cmd

import (
	"bytes"
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

func TestCreate(t *testing.T) {
	tmp, teardown := test.Setup(t)
	defer teardown()

	is := is.New(t)

	Output = &bytes.Buffer{}

	var err error
	err = Create(NewOptions(tmp), []string{})
	is.NoErr(err)

	err = Info(NewOptions(tmp), []string{})
	is.NoErr(err)

	t.Log(Output)
}
