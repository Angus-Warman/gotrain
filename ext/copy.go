package ext

import (
	"io"
	"os"
	"path/filepath"
)

func Copy(src, dst string) error {
	dstFolder := filepath.Dir(dst)

	err := os.MkdirAll(dstFolder, 0755)

	if err != nil {
		return err
	}

	srcFile, err := os.Open(src)

	if err != nil {
		return err
	}

	defer srcFile.Close()

	dstFile, err := os.Create(dst)

	if err != nil {
		return err
	}

	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)

	if err != nil {
		return err
	}

	return nil
}
