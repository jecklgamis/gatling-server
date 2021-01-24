package workspace

import (
	util "github.com/jecklgamis/gatling-server/pkg/fileioutil"
	test "github.com/jecklgamis/gatling-server/pkg/testing"
	"io/ioutil"
	"path/filepath"
	"testing"
)

func TestUserFilesDirMustBeAbsolute(t *testing.T) {
	_, err := NewUserFilesDir(".")
	test.Assertf(t, err != nil, "unable to create user files dir %v", err)
}

func TestNewUserFilesDir(t *testing.T) {
	userFilesDir, err := NewUserFilesDir(someNonExistingDir(t))
	test.Assert(t, err == nil, "unable to user files dir")
	test.Assert(t, util.DirExists(userFilesDir.BaseDir), "base dir does not exist")
	test.Assert(t, util.DirExists(userFilesDir.Simulations), "simulations dir does not exist")
	test.Assert(t, util.DirExists(userFilesDir.Binaries), "binaries dir does not exist")
	test.Assert(t, util.DirExists(userFilesDir.Resources), "resources dir does not exist")
	test.Assert(t, util.DirExists(userFilesDir.Results), "results dir does not exist")
}

func TestNewUserFilesDirFailsIfExist(t *testing.T) {
	dir, err := ioutil.TempDir("", "some-dir")
	test.Assert(t, err == nil, "unable create dir")
	test.Assert(t, util.DirExists(dir), "dir must exist")
	_, err = NewUserFilesDir(dir)
	test.Assert(t, err != nil, "expecting to fail")
}

func TestNewUserFilesDirFailsIfNotAbsolute(t *testing.T) {
	_, err := NewUserFilesDir(".")
	test.Assert(t, err != nil, "expecting to fail")
}

func someNonExistingDir(t *testing.T) string {
	dir, err := ioutil.TempDir("", "")
	test.Assert(t, err == nil, "unable create dir")
	return filepath.Join(dir, "some-dir")
}
