package util

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

const (
	LeftGoTemplateDelim  = "#{"
	RightGoTemplateDelim = "}#"
)

var (
	templateFuncMap = template.FuncMap{
		"indentNewLines": templateIndentNewlines,
	}
	templateParser = template.New("template").Funcs(templateFuncMap).
			Delims(LeftGoTemplateDelim, RightGoTemplateDelim).
			Option("missingkey=error")
)

func CopyDirectory(scrDir, dest string, templateData any) error {
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
			if err := CopyDirectory(sourcePath, destPath, templateData); err != nil {
				return err
			}
		case os.ModeSymlink:
			continue
		default:
			if err := Copy(sourcePath, destPath, templateData); err != nil {
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

func Copy(srcFile, dstFile string, templateData any) error {
	inB, err := os.ReadFile(srcFile)
	if err != nil {
		return err
	}

	rendered, err := ApplyTemplateWithCustomDelim(inB, templateData)
	if err != nil {
		return fmt.Errorf("applying template: %w", err)
	}

	return os.WriteFile(dstFile, rendered, 0644)
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

	// Execute the template with the file content and write the output
	ret := bytes.Buffer{}
	err = t.Execute(&ret, templateData)
	if err != nil {
		return nil, err
	}

	return ret.Bytes(), nil
}

func ApplyTemplateWithCustomDelim(in []byte, templateData any) ([]byte, error) {
	t, err := templateParser.Parse(string(in))
	if err != nil {
		return nil, err
	}

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
