package handlers

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"path"
	"strings"
)

var (
	YamlPathList []string
	YamlReader   *bufio.Reader
)

// 获取根目录下直属所有文件（不包括文件夹及其中的文件）
func GetAllYamlFile(pathname string, s []string) ([]string, error) {
	rd, err := ioutil.ReadDir(pathname)
	if err != nil {
		fmt.Println("read dir fail:", err)
		return s, err
	}

	for _, fi := range rd {
		if !fi.IsDir() {
			fullName := pathname + "/" + fi.Name()
			s = append(s, fullName)
		}
	}
	return s, nil
}

// 获取当前项目根目录下所有文件（包括文件夹中的文件）
func GetYamlFiles(folder string) {
	files, _ := ioutil.ReadDir(folder)
	for _, file := range files {
		if strings.Contains(file.Name(), ".idea") || strings.Contains(file.Name(), "rewi") || strings.Contains(file.Name(), ".vs") || strings.Contains(file.Name(), ".git") {
			fmt.Println("Skip :", file.Name())
			continue
		}
		if file.IsDir() {
			// GetYamlFiles(folder + "/" + file.Name())
		} else {
			var filename = file.Name()
			if strings.Contains(strings.ToLower(path.Ext(filename)), ".yaml") {
				YamlPathList = append(YamlPathList, folder+"/"+filename)
			}
			// fmt.Println(folder + "/" + file.Name())
		}
	}
}
