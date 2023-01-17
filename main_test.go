package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileDiff(t *testing.T) {

	assert := assert.New(t)

	// TODO: probably would need to test table those tests
	t.Run("should be able to detect chunk changes", func(t *testing.T) {
		// given
		original := []byte("Hello everyone, this will be a very short text about nothing. It's only purpose if for testing")
		// change everyone -> Everyone, and if -> IF
		updated := []byte("Hello Everyone, this will be a very short text about nothing. It's only purpose IF for testing")

		// original file
		originalFile, err := createTempTestFile(original)
		defer os.Remove(originalFile.Name())
		defer originalFile.Close()
		assert.NoError(err)

		// updated file
		updatedFile, err := createTempTestFile(updated)
		defer os.Remove(updatedFile.Name())
		defer updatedFile.Close()
		assert.NoError(err)

		// when
		delta, err := FileDiff(originalFile, updatedFile, 8, 4)
		assert.NoError(err)
		assert.NotNil(delta)

		// figure out how to test it
		assert.True(len(delta.Changed) > 0)
		assert.True(len(delta.Reused) > 0)

	})

	t.Run("should be able to detect chunk removals at the beginning of the file", func(t *testing.T) {

	})

	t.Run("should be able to detect chunk removals at the end of the file", func(t *testing.T) {

	})

	t.Run("should be able to detect chunk removals in the middle of the file", func(t *testing.T) {

	})

	t.Run("should be able to detect chunk additions at the end of the file", func(t *testing.T) {

	})

	t.Run("should be able to detect chunk additions at the beginning of the file", func(t *testing.T) {

	})

	t.Run("should be able to detect chunk additions in the middle of the file", func(t *testing.T) {

	})
}

func createTempTestFile(fileContent []byte) (file *os.File, err error) {
	// Write the original and updated bytes to temporary files
	file, err = os.CreateTemp("", "original")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer func() {
		if err != nil {
			removalErr := os.Remove(file.Name())
			if removalErr != nil {
				err = fmt.Errorf("%w, error removing file on error", err)
			}
		}
	}()

	if _, err = file.Write(fileContent); err != nil {
		return nil, fmt.Errorf("failed to write file on disk: %w", err)
	}
	if err = file.Close(); err != nil {
		return nil, fmt.Errorf("failed to close file: %w", err)
	}

	// Open the file
	file, err = os.Open(file.Name())
	if err != nil {
		return nil, fmt.Errorf("failed to open a file: %w", err)
	}
	defer func() {
		if err != nil {
			closeErr := file.Close()
			if closeErr != nil {
				err = fmt.Errorf("%w, error closing file on error", err)
			}
		}
	}()
	return file, nil
}
