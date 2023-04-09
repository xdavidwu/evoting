package store

import (
	"os"
	"path"
)

const (
	projectName	= "evoting"
	serverName	= projectName + "-server"
	clientName	= projectName + "-client"
)

func dataDir(name string) string {
	dataHome := os.Getenv("XDG_DATA_HOME")
	if dataHome == "" {
		dataHome = path.Join(os.Getenv("HOME"), ".local/share")
	}
	return path.Join(dataHome, name)
}

func ServerDataDir() string {
	return dataDir(serverName)
}

func ClientDataDir() string {
	return dataDir(clientName)
}
