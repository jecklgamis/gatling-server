package workspace

import (
	"fmt"
	util "github.com/jecklgamis/gatling-server/pkg/fileioutil"
	"log"
	"os"
	"path/filepath"
)

type UserFilesDir struct {
	BaseDir     string
	Simulations string
	Binaries    string
	Resources   string
	Results     string
}

func NewUserFilesDir(baseDir string) (*UserFilesDir, error) {
	if util.DirExists(baseDir) {
		err := fmt.Errorf("base dir %s exists already", baseDir)
		log.Println(err)
		return nil, err
	}
	if !filepath.IsAbs(baseDir) {
		err := fmt.Errorf("base dir %s is not absolute", baseDir)
		log.Println(err)
		return nil, err
	}
	userFilesDir := &UserFilesDir{
		BaseDir:     baseDir,
		Simulations: filepath.Join(baseDir, "simulations"),
		Binaries:    filepath.Join(baseDir, "binaries"),
		Resources:   filepath.Join(baseDir, "resources"),
		Results:     filepath.Join(baseDir, "results")}
	if err := userFilesDir.create(0744); err != nil {
		return nil, err
	}
	return userFilesDir, nil
}

func (r *UserFilesDir) create(perm os.FileMode) error {
	util.CreateDirIfNotExist(r.BaseDir, perm)
	util.CreateDirIfNotExist(r.Simulations, perm)
	util.CreateDirIfNotExist(r.Binaries, perm)
	util.CreateDirIfNotExist(r.Resources, perm)
	util.CreateDirIfNotExist(r.Results, perm)
	return nil
}
