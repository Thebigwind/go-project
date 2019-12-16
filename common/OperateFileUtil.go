package common

import (
	"errors"
	"os"
)

func CheckReadFile(pathStr string) error {
	info, err := os.Stat(pathStr)
	if err != nil {
		if os.IsNotExist(err) {
			return errors.New("No such file.")
		} else if os.IsPermission(err) {
			return errors.New("Permission denied !")
		} else {
			return err
		}
	}
	if info.IsDir() {
		return errors.New("Not a file.")
	}
	if info.Mode().Perm()&os.FileMode(128) == 0 {
		return errors.New("Cannot read the file. Permission denied !")
	}
	return nil
}

func FileExists(pathStr string) error {
	_, err := os.Stat(pathStr)
	if err != nil {
		if os.IsNotExist(err) {
			return errors.New("No such file.")
		} else if os.IsPermission(err) {
			return errors.New("Permission denied !")
		} else {
			return err
		}
	}
	return nil
}
