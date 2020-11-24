package utils

import (
	"os"

	"yunion.io/x/pkg/errors"
)

func EnsureDir(dirName string) error {
	if _, err := os.Stat(dirName); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(dirName, os.ModePerm); err != nil {
				return errors.Wrapf(err, "mkdir %q", dirName)
			}
		} else {
			return err
		}
	}
	return nil
}

func OpenOrCreateFile(fileName string) (*os.File, error) {
	if _, err := os.Stat(fileName); err != nil {
		if os.IsNotExist(err) {
			f, err := os.Create(fileName)
			if err != nil {
				return nil, errors.Wrapf(err, "create file %q", fileName)
			}
			return f, nil
		}
		return nil, err
	} else {
		return os.OpenFile(fileName, os.O_RDWR, 0)
	}
}
