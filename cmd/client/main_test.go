package main

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(t *testing.T) {
	t.Run("get-wow", func(t *testing.T) {
		t.Run("should initialize get-wow app", func(t *testing.T) {
			rootCmd := setupCommands()
			rootCmd.SetArgs([]string{"get-wow", "--noop", "--logs-file", "../../test.log"})
			require.NoError(t, rootCmd.Execute())
		})
		t.Run("should fail if bad log level", func(t *testing.T) {
			rootCmd := setupCommands()
			rootCmd.SilenceErrors = true
			rootCmd.SilenceUsage = true
			rootCmd.SetArgs([]string{"get-wow", "--noop", "-l", faker.Word(), "--logs-file", "../../test.log"})
			assert.Error(t, rootCmd.Execute())
		})
		t.Run("should fail if unexpected env", func(t *testing.T) {
			rootCmd := setupCommands()
			rootCmd.SilenceErrors = true
			rootCmd.SilenceUsage = true
			rootCmd.SetArgs([]string{"get-wow", "--noop", "-e", faker.Word(), "--logs-file", "../../test.log"})
			gotErr := rootCmd.Execute()
			assert.ErrorContains(t, gotErr, "failed to read config")
		})
	})
	t.Run("solve-challenge", func(t *testing.T) {
		t.Run("should solve the challenge", func(t *testing.T) {
			rootCmd := setupCommands()
			challenge := faker.Word()
			rootCmd.SetArgs([]string{"solve-challenge", "--silent", "--challenge", challenge, "--logs-file", "../../test.log"})
			require.NoError(t, rootCmd.Execute())
		})
	})
}
