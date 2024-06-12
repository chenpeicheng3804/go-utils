package util

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErgodicPathFile(t *testing.T) {
	// Test case 1: empty path
	_, err := ErgodicPathFile("", "condition")
	assert.Equal(t, errors.New("path cannot be empty"), err)

	// Test case 2: empty condition
	_, err = ErgodicPathFile("/tmp/path/to/dir", "")
	assert.Equal(t, errors.New("condition cannot be empty"), err)

	// Test case 3: path does not exist
	_, err = ErgodicPathFile("/tmp/path/to/dir/path", "condition")
	assert.Equal(t, errors.New("path does not exist"), err)

	// Test case 4: valid path and condition
	expectedFiles := []string{"/tmp/path/to/dir/file1.txt", "/tmp/path/to/dir/subdir/file2.txt"}
	f, err := ErgodicPathFile("/tmp/path/to/dir", ".txt")
	assert.NoError(t, err)
	assert.Equal(t, expectedFiles, f.Files)
}

func TestReadFile1(t *testing.T) {
	Read, err := ReadFile("/home/demo/git/xdz/ttk/db/upgrade/gl/data_gl_upgrade.sql")
	if err != nil {
		t.Errorf("ReadFile() error = %v", err)
	}
	for _, v := range Read {
		fmt.Println("1")
		fmt.Println(v)
		fmt.Println("2")
	}

}
