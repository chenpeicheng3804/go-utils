package util

import (
	"errors"
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
	assert.Equal(t, expectedFiles, f.files)
}
