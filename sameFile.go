package watcher

import "os"

func IsSameFile(fileInfo1, fileInfo2 os.FileInfo) bool {
	return os.SameFile(fileInfo1, fileInfo2)
}
