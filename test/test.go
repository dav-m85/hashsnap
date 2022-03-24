package test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

type P struct{}

// Setup copies files from
func Setup(t *testing.T) (tmp string, teardown func()) {
	teardown = func() {}

	var err error

	wd, err := os.Getwd()
	if err != nil {
		t.Errorf("cannot create temp dir for fixtures: %s", err)
		return
	}

	tmp, err = os.MkdirTemp(os.TempDir(), "hashsnap_test")
	if err != nil {
		t.Errorf("cannot create temp dir for fixtures: %s", err)
		return
	}

	cmd := exec.Command("cp", "-R", filepath.Join(wd, "../test/fixtures"), tmp)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err = cmd.Run(); err != nil {
		t.Errorf("cannot copy fixtures over: %s", err)
		return
	}

	fmt.Println(os.Getwd())

	teardown = func() {
		// err = exec.Command("rm", "-rf", tmp).Run() // Smells like danger
		// if err != nil {
		// 	t.Error(err)
		// 	return
		// }
	}
	return
}
