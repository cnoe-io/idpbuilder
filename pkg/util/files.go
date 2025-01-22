package util

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

func CopyDirectory(scrDir, dest string) error {
	entries, err := os.ReadDir(scrDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		sourcePath := filepath.Join(scrDir, entry.Name())
		destPath := filepath.Join(dest, entry.Name())

		fileInfo, err := os.Stat(sourcePath)
		if err != nil {
			return err
		}

		switch fileInfo.Mode() & os.ModeType {
		case os.ModeDir:
			if err := CreateIfNotExists(destPath, 0755); err != nil {
				return err
			}
			if err := CopyDirectory(sourcePath, destPath); err != nil {
				return err
			}
		case os.ModeSymlink:
			continue
		default:
			if err := Copy(sourcePath, destPath); err != nil {
				return err
			}
		}

		fInfo, err := entry.Info()
		if err != nil {
			return err
		}
		if err := os.Chmod(destPath, fInfo.Mode()); err != nil {
			return err
		}
	}
	return nil
}

func Copy(srcFile, dstFile string) error {
	out, err := os.Create(dstFile)
	if err != nil {
		return err
	}

	defer out.Close()

	in, err := os.Open(srcFile)
	if err != nil {
		return err
	}

	defer in.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	return nil
}

func Exists(filePath string) bool {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return false
	}

	return true
}

func CreateIfNotExists(dir string, perm os.FileMode) error {
	if Exists(dir) {
		return nil
	}

	if err := os.MkdirAll(dir, perm); err != nil {
		return fmt.Errorf("failed to create directory: '%s', error: '%s'", dir, err.Error())
	}

	return nil
}

func ApplyTemplate(in []byte, templateData any) ([]byte, error) {
	funcMap := template.FuncMap{
		"indentNewLines": templateIndentNewlines,
	}
	t, err := template.New("template").Funcs(funcMap).Parse(string(in))
	if err != nil {
		return nil, err
	}

	// Execute the template with the file content and write the output to the destination file
	ret := bytes.Buffer{}
	err = t.Execute(&ret, templateData)
	if err != nil {
		return nil, err
	}

	return ret.Bytes(), nil
}

// indent given string with given number of spaces whenever a newline symbol is found.
func templateIndentNewlines(n int, val string) string {
	return strings.Replace(val, "\n", "\n"+strings.Repeat(" ", n), -1)
}
