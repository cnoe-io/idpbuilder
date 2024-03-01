package util

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
)

type FS interface {
	ReadDir(name string) ([]fs.DirEntry, error)
	ReadFile(name string) ([]byte, error)
}

func ConvertFSToBytes(inFS FS, name string, templateData any) ([][]byte, error) {
	d, err := inFS.ReadDir(name)
	if err != nil {
		return nil, err
	}

	var rawResources [][]byte

	for _, f := range d {
		rawResource, err := inFS.ReadFile(path.Join(name, f.Name()))
		if err != nil {
			return nil, err
		}

		if returnedRawResource, err := ApplyTemplate(rawResource, templateData); err == nil {
			rawResources = append(rawResources, returnedRawResource)
		} else {
			return nil, err
		}
	}

	return rawResources, nil
}

func CopyFile(src fs.File, dest string) error {
	srcStat, srcStatErr := src.Stat()
	if srcStatErr != nil {
		return srcStatErr
	}

	destFn := filepath.Join(dest, srcStat.Name())

	destf, destErr := os.OpenFile(
		destFn,
		os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
		srcStat.Mode(),
	)
	if destErr != nil {
		return fmt.Errorf("opening a file for writing: %w", destErr)
	}

	_, err := io.Copy(destf, src)
	if err != nil {
		return fmt.Errorf("copying %s to %s", srcStat.Name(), destFn)
	}

	return destf.Close()
}

func CopyDir(src fs.FS, dest string) error {
	ents, err := fs.ReadDir(src, ".")
	if err != nil {
		return fmt.Errorf("reading src: %w", err)
	}

	for _, sdent := range ents {
		info, err := sdent.Info()
		if err != nil {
			return fmt.Errorf("reading file info: %v", err)
		}
		switch {
		case info.IsDir():
			subDest := filepath.Join(dest, sdent.Name())

			if err := os.Mkdir(subDest, 0700); err != nil {
				return fmt.Errorf("mkdir on %s: %w", subDest, err)
			}

			subFS, err := fs.Sub(src, sdent.Name())
			if err != nil {
				return fmt.Errorf("reading the sub directory: %w", err)
			}

			if err := CopyDir(subFS, subDest); err != nil {
				return err
			}
		case info.Mode().IsRegular():
			srcf, err := src.Open(info.Name())
			if err != nil {
				return err
			}
			if err := CopyFile(srcf, dest); err != nil {
				return err
			}
		}
	}

	return nil
}

func WriteFS(src fs.FS, dest string) error {
	destInfo, destErr := os.Lstat(dest)
	if destErr != nil {
		return destErr
	}

	if !destInfo.IsDir() {
		return errors.New("the destination must be a directory")
	}

	return CopyDir(src, dest)
}
