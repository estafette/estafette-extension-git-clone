package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetTargetDir(t *testing.T) {
	t.Run("ReturnsEstafetteWorkIfSubdirIsDot", func(t *testing.T) {

		// act
		path := getTargetDir(".")

		assert.Equal(t, "/estafette-work", path)
	})

	t.Run("ReturnsEstafetteWorkSubdirIfSubdirIsSingleWord", func(t *testing.T) {

		// act
		path := getTargetDir("scripts")

		assert.Equal(t, "/estafette-work/scripts", path)
	})

	t.Run("ReturnsEstafetteWorkSubdirIfSubdirIsDotSlashSingleWord", func(t *testing.T) {

		// act
		path := getTargetDir("./scripts")

		assert.Equal(t, "/estafette-work/scripts", path)
	})

	t.Run("ReturnsEstafetteWorkSubdirIfSubdirIsMultipleWordsSeparatedBySlash", func(t *testing.T) {

		// act
		path := getTargetDir("scripts/sub")

		assert.Equal(t, "/estafette-work/scripts/sub", path)
	})

	t.Run("ReturnsEstafetteWorkSubdirIfSubdirIsDotSlashMultipleWordsSeparatedBySlash", func(t *testing.T) {

		// act
		path := getTargetDir("./scripts/sub")

		assert.Equal(t, "/estafette-work/scripts/sub", path)
	})

	t.Run("ReturnsEstafetteWorkSubdirIfSubdirIsAbsolutePath", func(t *testing.T) {

		// act
		path := getTargetDir("/scripts")

		assert.Equal(t, "/estafette-work/scripts", path)
	})
}
