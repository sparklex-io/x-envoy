package config

import (
	"github.com/stretchr/testify/assert"
	"os"
	"path"
	"testing"
)

func TestLoadConfig_FileDoesNotExist(t *testing.T) {
	_, err := LoadConfig("nonexistent.yaml")
	assert.Error(t, err)
}

func TestLoadConfig_HappyPath(t *testing.T) {
	// Create a temporary config file
	tmpfile, err := os.CreateTemp("", "config")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name()) // clean up

	// Write a valid config to the file
	content := []byte(`
PrivateKey: "myPrivateKey"
SparkleX:
  URL: "http://sparklex.com"
  ReducerAddress: "0x123"
Ethereum:
  URL: "http://ethereum.com"
  GERAddress: "0x456"
BSC:
  URL: "http://bsc.com"
  GERAddress: "0x789"
`)
	if _, err := tmpfile.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	// Load the config
	absPath := tmpfile.Name()
	config, err := LoadConfig(path.Base(absPath), path.Dir(absPath)+"/")
	if err != nil {
		t.Fatal(err)
	}

	// Assert the config values
	assert.Equal(t, "myPrivateKey", config.PrivateKey)
	assert.Equal(t, "http://sparklex.com", config.SparkleX.URL)
	assert.Equal(t, "0x123", config.SparkleX.ReducerAddress)
	assert.Equal(t, "http://ethereum.com", config.Ethereum.URL)
	assert.Equal(t, "0x456", config.Ethereum.GERAddress)
	assert.Equal(t, "http://bsc.com", config.BSC.URL)
	assert.Equal(t, "0x789", config.BSC.GERAddress)
}
