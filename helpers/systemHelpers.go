package helpers

import (
	"fmt"
	"os"
	"path/filepath"
)

func IsExist(path string) bool {

	_, err := os.Stat(path)

	return err == nil || os.IsExist(err)

	// 或者

	//return err == nil || !os.IsNotExist(err)

	// 或者

	//return !os.IsNotExist(err)

}

// basePath是固定目录路径,不包含具体的文件名，如果你传成了 /home/xx.txt, xx.txt也会被当成目录
func CheckAndCreateDir(basePath string) (dirPath string) {
	folderPath := filepath.Join(basePath)

	err := os.MkdirAll(folderPath, os.ModePerm)

	if err != nil {
		fmt.Println("创建目录报错")
		fmt.Println(err)
	}
	return folderPath
}
