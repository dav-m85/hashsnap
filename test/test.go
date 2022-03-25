package test

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

type P struct{}

// Setup copies files from
func Setup(t *testing.T) (tmp string, teardown func()) {
	teardown = func() {}

	var err error

	tmp = t.TempDir()
	// tmp, err = os.MkdirTemp(os.TempDir(), "hashsnap_test")
	// if err != nil {
	// 	t.Errorf("cannot create temp dir for fixtures: %s", err)
	// 	return
	// }

	_, filename, _, _ := runtime.Caller(0)
	// t.Logf("Current test filename: %s", filename)

	cmd := exec.Command("cp", "-R", filepath.Join(filepath.Dir(filename), "fixtures")+"/.", tmp)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err = cmd.Run(); err != nil {
		t.Errorf("cannot copy fixtures over: %s", err)
		return
	}

	teardown = func() {
		// err = exec.Command("rm", "-rf", tmp).Run() // Smells like danger
		// if err != nil {
		// 	t.Error(err)
		// 	return
		// }
	}
	return
}
