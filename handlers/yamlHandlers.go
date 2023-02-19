package handlers

import (
	"bufio"
	"fmt"
	"github.com/Re-Wi/GoKitReWi/helpers"
	"github.com/XM-GO/PandaKit/utils"
	"gopkg.in/yaml.v3"
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

// 计划文件内容解析
func PlanContentParsing(filepath string) (helpers.PlanContent, error) {
	//方法一：
	//var plan helpers.PlanContent
	//// 存储解析数据
	//result := make(map[string]interface{})
	//err := utils.LoadYml(filepath, &result)
	////map转json
	//bytes, _ := json.Marshal(result)
	//stringData := string(bytes)
	//// 将 JSON 格式的数据解析到结构体中
	//err = json.Unmarshal([]byte(stringData), &plan)
	//if err != nil {
	//	fmt.Println("Error decoding JSON:", err)
	//}
	//方法二：
	var plan helpers.PlanContent
	err := utils.LoadYml(filepath, &plan)
	if err != nil {
		fmt.Printf("Unmarsh file %v to %T fail: %v", filepath, plan, err)
	}
	return plan, err
}

func PlanContentSave[T helpers.PlanContent](fileName string, conf *T) error {
	data, err := yaml.Marshal(conf)
	if err != nil {
		fmt.Printf("Marshal conf %v fail: %v", conf, err)
		return err
	}

	err = ioutil.WriteFile(fileName, data, 0666)
	if err != nil {
		fmt.Printf("Write yaml fail: %v", err)
		return err
	}

	return nil
}
