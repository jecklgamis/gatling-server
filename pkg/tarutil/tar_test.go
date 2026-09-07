package tarutil

import (
	"os"

	"github.com/jecklgamis/gatling-server/pkg/fileioutil"
	test "github.com/jecklgamis/gatling-server/pkg/testing"
	"path/filepath"
	"testing"
)

func TestExtractTarGz(t *testing.T) {
	dir, _ := os.MkdirTemp("", "")
	err := Extract("testdata/gatling-scala-example-user-files.tar.gz", dir)
	test.Assert(t, err == nil, "failed to extract archive")
	test.Assertf(t, fileioutil.DirExists(filepath.Join(dir, "bodies")), "expecting bodies directory")
	test.Assertf(t, fileioutil.DirExists(filepath.Join(dir, "resources")), "expecting resources directory")
	test.Assertf(t, fileioutil.DirExists(filepath.Join(dir, "binaries")), "expecting binaries directory")
}

func TestCompressDir(t *testing.T) {
	srcDir, _ := os.MkdirTemp("", "src")
	err := fileioutil.CopyFile("testdata/some.txt", filepath.Join(srcDir, "some.txt"))
	test.Assertf(t, err == nil, "unable to copy file %v", err)

	dstDir, _ := os.MkdirTemp("", "dst")

	err = CompressDir(srcDir, dstDir, "some.tar.gz")
	test.Assertf(t, err == nil, "unable to archive")
	test.Assert(t, fileioutil.FileExist(filepath.Join(dstDir, "some.tar.gz")), "expecting file to exist")
}
