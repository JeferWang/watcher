package watcher

import "strings"

func IsHiddenFile(path string) (bool, error) {
	return strings.HasPrefix(path, "."), nil
}
