package handlers

import (
	"fmt"
	"github.com/Re-Wi/GoKitReWi/helpers"
	"io/ioutil"
	"path"
	"strings"
	"time"
)

var filePathList []string

// 获取根目录下直属所有文件（不包括文件夹及其中的文件）
func GetAllFiles(pathname string, s []string) ([]string, error) {
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
func GetPathFiles(folder string, suffix string) []string {
	files, _ := ioutil.ReadDir(folder)
	for _, file := range files {
		if strings.Contains(file.Name(), ".idea") || strings.Contains(file.Name(), "rewi") || strings.Contains(file.Name(), ".vs") || strings.Contains(file.Name(), ".git") {
			fmt.Println("Skip :", file.Name())
			continue
		}
		if file.IsDir() {
			GetPathFiles(folder+"/"+file.Name(), suffix)
		} else {
			var filename = file.Name()
			if strings.Contains(strings.ToLower(path.Ext(filename)), suffix) {
				filePathList = append(filePathList, folder+"/"+filename)
			}
			//fmt.Println(folder + "/" + file.Name())
		}
	}
	return filePathList
}

// 计划文件名解析
func PlanFilenameParsing(filename string) *helpers.PlanFilename {
	p := new(helpers.PlanFilename)
	filenameList := strings.Split(filename, ".")

	p.Status = filenameList[1]
	p.Suffix = filenameList[2]
	filenameList = strings.Split(filenameList[0], "__")

	//fmt.Println(filenameList)
	p.Project = filenameList[0]
	local, _ := time.LoadLocation("Asia/Shanghai")
	p.StartTime, _ = time.ParseInLocation("2006-01-02", filenameList[1], local)
	p.FinishTime, _ = time.ParseInLocation("2006-01-02", filenameList[2], local)

	return p
}
