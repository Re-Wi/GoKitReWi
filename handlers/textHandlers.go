package handlers

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strings"
)

var (
	TextPathList []string
	textReader   *bufio.Reader
	LineArray    [4]string
)

// 获取根目录下直属所有文件（不包括文件夹及其中的文件）
func GetAllTextFile(pathname string, s []string) ([]string, error) {
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
func GetTextFiles(folder string) {
	files, _ := ioutil.ReadDir(folder)
	for _, file := range files {
		if strings.Contains(file.Name(), ".idea") || strings.Contains(file.Name(), "rewi") {
			fmt.Println("Skip :", file.Name())
			continue
		}
		if file.IsDir() {
			GetTextFiles(folder + "/" + file.Name())
		} else {
			var filename = file.Name()
			if strings.Contains(strings.ToLower(path.Ext(filename)), ".txt") {
				TextPathList = append(TextPathList, folder+"/"+filename)
			}
			//fmt.Println(folder + "/" + file.Name())
		}
	}
}
func OpenTextFile(filepath string) error {
	f, err := os.OpenFile(filepath, os.O_RDONLY, 0666)
	if err != nil {
		return err
	}
	//defer fileHanle.Close()
	textReader = bufio.NewReader(f)
	return nil
}
func ParseTextLine() error {
	line, _, err := textReader.ReadLine()
	if err == io.EOF {
		return err
	}
	lineData := string(line)
	if len(lineData) < 5 {
		return errors.New("too short")
	}
	//fmt.Printf("read result:%v\n", lineData)
	re := regexp.MustCompile(`[0-9]+\.[0-9]+`)
	matchArr := re.FindAllString(lineData, -1)
	if len(matchArr) <= 0 {
		return errors.New("no content")
	}
	LineArray[0] = matchArr[0]
	re = regexp.MustCompile(`[A-Z_]+`)
	matchArr = re.FindAllString(lineData, -1)
	LineArray[1] = matchArr[0]
	LineArray[2] = matchArr[1]
	re = regexp.MustCompile(`[a-z0-9]{8}`)
	matchArr = re.FindAllString(lineData, -1)
	LineArray[3] = matchArr[0]
	//fmt.Println(LineArray)
	return nil
}
