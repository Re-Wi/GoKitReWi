package handlers

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"path"
	"strings"
)

var (
	JsonPathList []string
	jsonReader   *bufio.Reader
)

// var file_locker sync.Mutex //config file locker
// file_locker.Lock()
// data, err := io.ReadFile(filename) //read config file
// file_locker.Unlock()

// func main() {
// 	// jsonData := `[
// 	// 	{
// 	// 		"count": 8,
// 	// 		"create_time": "2023-01-11 16:00:58.811",
// 	// 		"id": 1,
// 	// 		"points": "['(1164,526)', '(731,180)', '(816,359)', '(475,354)', '(150,505)', '(968,499)', '(353,478)', '(523,175)']",
// 	// 		"signal": -1
// 	// 	},
// 	// 	{
// 	// 		"count": 8,
// 	// 		"create_time": "2023-01-11 16:01:08.816",
// 	// 		"id": 2,
// 	// 		"points": "['(1171,528)', '(728,180)', '(816,359)', '(473,358)', '(149,500)', '(970,498)', '(350,478)', '(522,179)']",
// 	// 		"signal": -1
// 	// 	}]`

// 	// keys := make([]PublicKey, 0)
// 	// err := json.Unmarshal([]byte(s), &keys)
// 	// if err == nil {
// 	// 	fmt.Printf("%+v\n", keys)
// 	// 	fmt.Printf("~~~~~~:%+v\n", keys[0].Id)
// 	// } else {
// 	// 	fmt.Println(err)
// 	// 	fmt.Printf("%+v\n", keys)
// 	// }

//		jsonFile, err := os.Open("tests/touch_log.json")
//		if err != nil {
//			fmt.Println("error opening json file")
//			return
//		}
//		defer jsonFile.Close()
//		jsonData, err := ioutil.ReadAll(jsonFile)
//		if err != nil {
//			fmt.Println("error reading json file")
//			return
//		}
//		keys := make([]PublicKey, 0)
//		err = json.Unmarshal([]byte(jsonData), &keys)
//		if err == nil {
//			fmt.Printf("%+v\n", keys)
//			fmt.Printf("~~~~~~:%+v\n", keys[0].Id)
//		} else {
//			fmt.Println(err)
//			fmt.Printf("%+v\n", keys)
//		}
//	}
//
// 获取根目录下直属所有文件（不包括文件夹及其中的文件）
func GetAllJsonFile(pathname string, s []string) ([]string, error) {
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
func GetJsonFiles(folder string) {
	files, _ := ioutil.ReadDir(folder)
	for _, file := range files {
		if strings.Contains(file.Name(), ".idea") || strings.Contains(file.Name(), "rewi") || strings.Contains(file.Name(), ".vs") || strings.Contains(file.Name(), ".git") {
			fmt.Println("Skip :", file.Name())
			continue
		}
		if file.IsDir() {
			// GetJsonFiles(folder + "/" + file.Name())
		} else {
			var filename = file.Name()
			if strings.Contains(strings.ToLower(path.Ext(filename)), ".json") {
				JsonPathList = append(JsonPathList, folder+"/"+filename)
			}
			// fmt.Println(folder + "/" + file.Name())
		}
	}
}
