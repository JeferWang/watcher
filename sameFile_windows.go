package watcher

import "os"

// 判断文件是否相同
func sameFile(f1, f2 os.FileInfo) {
	return f1.ModTime() == f2.ModTime() &&
		f1.Mode() == f2.Mode() &&
		f1.Size() == f2.Size() &&
		f1.IsDir() == f2.IsDir()
}
