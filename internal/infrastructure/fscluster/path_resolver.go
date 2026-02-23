package fscluster

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

const appDirName = "ceph-doctor"

var errHomeNotSet = errors.New("HOME is not set")

func defaultRootDir() (string, error) {
	if xdgStateHome, ok := os.LookupEnv("XDG_STATE_HOME"); ok && strings.TrimSpace(xdgStateHome) != "" {
		return filepath.Join(xdgStateHome, appDirName), nil
	}

	home, ok := os.LookupEnv("HOME")
	if !ok || strings.TrimSpace(home) == "" {
		return "", errHomeNotSet
	}

	return filepath.Join(home, ".local", "state", appDirName), nil
}
