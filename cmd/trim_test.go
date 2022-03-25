package cmd

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/matryer/is"
)

// Setup copies files from
func setup(t *testing.T) (tmp string) {
	var err error

	tmp = t.TempDir()
	// tmp, err = os.MkdirTemp(os.TempDir(), "hashsnap_test")
	// if err != nil {
	// 	t.Errorf("cannot create temp dir for fixtures: %s", err)
	// 	return
	// }

	_, filename, _, _ := runtime.Caller(0)
	// t.Logf("Current test filename: %s", filename)

	cmd := exec.Command("cp", "-R", filepath.Join(filepath.Dir(filename), "testdata")+"/.", tmp)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err = cmd.Run(); err != nil {
		t.Errorf("cannot copy fixtures over: %s", err)
		return
	}

	return
}

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

func TestInfo(t *testing.T) {
	tmp := setup(t)
	is := is.New(t)

	Output = &bytes.Buffer{}

	is.NoErr(Create(NewOptions(tmp+"/dir1"), []string{}))

	is.NoErr(Info(NewOptions(tmp+"/dir1"), []string{}))

	t.Log(Output)
}

func TestTrim(t *testing.T) {
	tmp := setup(t)
	is := is.New(t)

	Output = &bytes.Buffer{}

	is.NoErr(Create(NewOptions(tmp+"/dir1"), []string{}))

	is.NoErr(Create(NewOptions(tmp+"/dir1_copy"), []string{}))
	is.NoErr(Info(NewOptions(tmp+"/dir1"), []string{}))

	is.NoErr(Trim(NewOptions(tmp+"/dir1_copy"), []string{tmp + "/dir1/.hsnap"}))

	t.Log(Output)
}
