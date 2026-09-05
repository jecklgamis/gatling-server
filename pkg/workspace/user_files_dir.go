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
	Libraries   string
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
		Results:     filepath.Join(baseDir, "results"),
		Libraries:   filepath.Join(baseDir, "lib"),
	}
	if err := userFilesDir.create(0744); err != nil {
		return nil, err
	}
	return userFilesDir, nil
}

func (r *UserFilesDir) create(perm os.FileMode) error {
	if err := util.CreateDirIfNotExist(r.BaseDir, perm); err != nil {
		return err
	}
	if err := util.CreateDirIfNotExist(r.Simulations, perm); err != nil {
		return err
	}
	if err := util.CreateDirIfNotExist(r.Results, perm); err != nil {
		return err
	}
	return nil
}
