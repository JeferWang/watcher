package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func main() {
	fileList := make(map[string]os.FileInfo)
	err := filepath.Walk("/Users/jeferwang/projects/pos_web/app/Common/BudgetCommon.php", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			fileList[path] = info
		}
		fmt.Println(fileList)
		return nil
	})
	if err != nil {
		log.Fatalln(err)
	}
}
