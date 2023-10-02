package util

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/cnoe-io/idpbuilder/globals"
	"github.com/google/go-cmp/cmp"
)

func TestWriteFS(t *testing.T) {
	cases := []struct {
		name        string
		srcFS       fs.FS
		expectErr   error
		expectFiles map[string][]byte
	}{{
		name: "single file",
		srcFS: fstest.MapFS{
			"testfile": &fstest.MapFile{
				Data: []byte("testdata"),
				Mode: 0666,
			},
		},
		expectFiles: map[string][]byte{
			"testfile": []byte("testdata"),
		},
	}, {
		name: "file in subdir",
		srcFS: fstest.MapFS{
			"somedir": &fstest.MapFile{
				Mode: fs.ModeDir,
			},
			"somedir/testfile": &fstest.MapFile{
				Data: []byte("testdata"),
				Mode: 0666,
			},
		},
		expectFiles: map[string][]byte{
			"somedir/testfile": []byte("testdata"),
		},
	}}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			workDir, err := os.MkdirTemp("", fmt.Sprintf("%s-fs_test.go-%s", globals.ProjectName, tc.name))
			if err != nil {
				t.Fatalf("creating tempdir: %v", err)
			}
			defer os.RemoveAll(workDir)

			err = WriteFS(tc.srcFS, workDir)
			if err != tc.expectErr {
				t.Errorf("unexpected error writing fs: %v", err)
			}

			for expectPath, expectData := range tc.expectFiles {
				fullExpectPath := filepath.Join(workDir, expectPath)
				gotData, err := os.ReadFile(fullExpectPath)
				if err != nil {
					t.Errorf("Opening expected file: %v", err)
				}

				if diff := cmp.Diff(string(expectData), string(gotData)); diff != "" {
					t.Errorf("Expected data in %q mismatch (-want +got):\n%s", expectPath, diff)
				}
			}
		})
	}
}
