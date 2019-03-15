package watcher

import "os"

// 判断文件是否相同
func IsSameFile(fileInfo1, fileInfo2 os.FileInfo) bool {
	return os.SameFile(fileInfo1, fileInfo2)
}
