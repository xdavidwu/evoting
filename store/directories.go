package store

import "os"

const (
	projectName	= "evoting"
	serverName	= projectName + "-server"
	clientName	= projectName + "-client"
)

func dataDir(name string) string {
	dataHome := os.Getenv("XDG_DATA_HOME")
	if dataHome == "" {
		dataHome = os.Getenv("HOME") + "/.local/share"
	}
	return dataHome + "/" + name
}

func ServerDataDir() string {
	return dataDir(serverName)
}

func ClientDataDir() string {
	return dataDir(clientName)
}
