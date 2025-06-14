package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateOutput(t *testing.T) {
	// Create temporary directory for test files
	tempDir := t.TempDir()
	
	t.Run("ValidateValidOutput", func(t *testing.T) {
		// Create test Go output
		goOutput := filepath.Join(tempDir, "go_output.json")
		goData := `[
			{
				"instance_type": "t2.micro",
				"memory": "1.0",
				"vCPU": 1,
				"family": "General purpose"
			},
			{
				"instance_type": "t2.small", 
				"memory": "2.0",
				"vCPU": 1,
				"family": "General purpose"
			}
		]`
		err := os.WriteFile(goOutput, []byte(goData), 0644)
		require.NoError(t, err)
		
		// Create test Python output
		pythonOutput := filepath.Join(tempDir, "python_output.json")
		pythonData := `[
			{
				"instance_type": "t2.micro",
				"memory": "1.0",
				"vCPU": 1,
				"family": "General purpose"
			},
			{
				"instance_type": "t2.medium",
				"memory": "4.0", 
				"vCPU": 2,
				"family": "General purpose"
			}
		]`
		err = os.WriteFile(pythonOutput, []byte(pythonData), 0644)
		require.NoError(t, err)
		
		// Validate - should not return error even with differences
		err = validateOutput(goOutput, pythonOutput, "test")
		assert.NoError(t, err)
	})
	
	t.Run("ValidateInvalidJSONFiles", func(t *testing.T) {
		// Create invalid JSON file
		invalidJSON := filepath.Join(tempDir, "invalid.json")
		err := os.WriteFile(invalidJSON, []byte("invalid json"), 0644)
		require.NoError(t, err)
		
		validJSON := filepath.Join(tempDir, "valid.json")
		err = os.WriteFile(validJSON, []byte("[]"), 0644)
		require.NoError(t, err)
		
		// Should return error for invalid JSON
		err = validateOutput(invalidJSON, validJSON, "test")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse Go output")
		
		err = validateOutput(validJSON, invalidJSON, "test")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse Python output")
	})
	
	t.Run("ValidateNonexistentFiles", func(t *testing.T) {
		nonexistent := filepath.Join(tempDir, "nonexistent.json")
		validJSON := filepath.Join(tempDir, "valid.json")
		err := os.WriteFile(validJSON, []byte("[]"), 0644)
		require.NoError(t, err)
		
		// Should return error for nonexistent files
		err = validateOutput(nonexistent, validJSON, "test")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read Go output")
		
		err = validateOutput(validJSON, nonexistent, "test")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read Python output")
	})
}

func TestCLIFlags(t *testing.T) {
	// Test that CLI flags are properly initialized
	assert.NotNil(t, rootCmd)
	assert.Equal(t, "scraper", rootCmd.Use)
	assert.Contains(t, rootCmd.Short, "Cloud instance scraper")
	
	// Check that required flags exist
	flags := rootCmd.Flags()
	
	outputFlag := flags.Lookup("output")
	assert.NotNil(t, outputFlag)
	assert.Equal(t, "www", outputFlag.DefValue)
	
	concurrencyFlag := flags.Lookup("concurrency")
	assert.NotNil(t, concurrencyFlag)
	assert.Equal(t, "10", concurrencyFlag.DefValue)
	
	servicesFlag := flags.Lookup("services")
	assert.NotNil(t, servicesFlag)
	
	regionsFlag := flags.Lookup("regions")
	assert.NotNil(t, regionsFlag)
	
	sampleFlag := flags.Lookup("sample")
	assert.NotNil(t, sampleFlag)
	assert.Equal(t, "0", sampleFlag.DefValue)
	
	cacheFlag := flags.Lookup("use-cache")
	assert.NotNil(t, cacheFlag)
	assert.Equal(t, "false", cacheFlag.DefValue)
}