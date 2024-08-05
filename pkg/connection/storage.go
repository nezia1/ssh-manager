package connection

import (
	"io"
	"os"
	"path/filepath"

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
	storageFile, err := openStorageFile(os.O_APPEND | os.O_CREATE | os.O_WRONLY)

	if err != nil {
		return err
	}

	b, err := toml.Marshal(cm)

	if err != nil {
		return err
	}

	_, err = storageFile.Write(b)

	if err != nil {
		return err
	}

	return nil
}

func (cm *ConnectionManager) LoadFromDisk() error {
	storageFile, err := openStorageFile(os.O_APPEND | os.O_CREATE | os.O_RDONLY)

	if err != nil {
		return err
	}

	defer storageFile.Close()

	b, err := io.ReadAll(storageFile)

	if err != nil {
		return err
	}

	err = toml.Unmarshal(b, cm)

	if err != nil {
		return err
	}

	return nil
}

func configurationPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	storagePath := filepath.Join(configDir, StorageDirPrefix, StorageFileName)
	return storagePath, nil
}

func openStorageFile(flags int) (*os.File, error) {
	storagePath, err := configurationPath()
	if err != nil {
		return nil, err
	}

	// check if directory containing the file exists
	_, err = os.Stat(filepath.Dir(storagePath))
	if os.IsNotExist(err) {
		err = os.MkdirAll(filepath.Dir(storagePath), StorageDirPerm)
		if err != nil {
			return nil, err
		}
	}

	storageFile, err := os.OpenFile(storagePath, flags, StorageFilePerm)

	if err != nil {
		return nil, err
	}

	return storageFile, nil
}
