package connection

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pelletier/go-toml"
)

const (
	StorageFileName  = "connections.toml"
	StorageDirPrefix = "ssh-manager"
	StorageDirPerm   = 0755
	StorageFilePerm  = 0600
)

// Saves the connections as a TOML file in the user config directory
func (cm ConnectionManager) SaveToDisk() error {

	err := ensureStorageFile()
	if err != nil {
		return err
	}

	storagePath, err := storageFilePath()
	if err != nil {
		return err
	}

	b, err := toml.Marshal(cm)

	if err != nil {
		return err
	}

	err = os.WriteFile(storagePath, b, StorageFilePerm)

	if err != nil {
		return err
	}

	return nil
}

// loadFromDisk loads the connections from the TOML file in the user config directory. If the file does not exist, it will be created, and an empty ConnectionManager will be initialized.
func (cm *ConnectionManager) loadFromDisk() error {
	storagePath, err := storageFilePath()

	if err != nil {
		return err
	}

	err = ensureStorageFile()

	if err != nil {
		return err
	}

	b, err := os.ReadFile(storagePath)

	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	if err := toml.Unmarshal(b, cm); err != nil {
		return fmt.Errorf("failed to unmarshal connections: %w", err)
	}

	return nil
}

// storageFilePath returns the path to the storage file in the user config directory.
func storageFilePath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user config directory: %w", err)
	}
	storagePath := filepath.Join(configDir, StorageDirPrefix, StorageFileName)
	return storagePath, nil
}

// ensureStorageFile ensures that the storage file exists. If the file does not exist, it will be created.
func ensureStorageFile() error {
	storageFilePath, err := storageFilePath()
	if err != nil {
		return err
	}

	storageDir := filepath.Dir(storageFilePath)

	// ensure directory exists
	if _, err := os.Stat(storageDir); os.IsNotExist(err) {
		if err := os.MkdirAll(storageDir, StorageDirPerm); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	// ensure file exists
	if _, err := os.Stat(storageFilePath); os.IsNotExist(err) {
		if _, err := os.Create(storageFilePath); err != nil {
			return fmt.Errorf("failed to create file: %w", err)
		}
	}

	return nil
}

// FetchConnections is a helper function for bubbletea. It fetches the connections from disk, and returns a message. This is useful for loading the model from the Init function.
//
// It returns a ConnectionsFetchedMsg message.
func (cm ConnectionManager) FetchConnections() tea.Msg {
	cm = ConnectionManager{}
	err := cm.loadFromDisk()

	if err != nil {
		return err
	}

	return ConnectionsFetchedMsg{FetchedManager: cm}
}

// ConnectionsFetchedMsg is a bubbletea message that is sent when the connections have been fetched from disk.
type ConnectionsFetchedMsg struct {
	FetchedManager ConnectionManager
}
