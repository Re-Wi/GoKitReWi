package helpers

import (
	"fmt"
	"testing"
)

func TestInterfaceToMap(t *testing.T) {
	// 创建一个示例的 map[string]interface{}
	data := map[string]interface{}{
		"ID":      1,
		"Name":    "John Doe",
		"Age":     30,
		"Country": "USA",
	}

	// 使用 InterfaceToMap 函数将 interface{} 转换为 map[string]interface{}
	result, ok := InterfaceToMap(data)

	if ok {
		fmt.Println("Conversion successful:")
		for key, value := range result {
			fmt.Printf("%s: %v\n", key, value)
		}
	} else {
		fmt.Println("Conversion failed.")
	}
}

func TestStructToMap(t *testing.T) {

	type User struct {
		Account    string
		Username   string
		Password   string
		Email      string
		Gender     string
		Admin      int
		Signature  string
		UpdateUser int
		CreateTime string
		UpdateTime string
		Status     string
	}
	req := User{
		Account:    "user123",
		Username:   "John Doe",
		Password:   "securepass",
		Email:      "john@example.com",
		Gender:     "male",
		Admin:      1,
		Signature:  "Hello, World!",
		UpdateUser: 42,
		CreateTime: "2023-01-01",
		UpdateTime: "2023-01-02",
		Status:     "active",
	}
	result := StructToMap(req)
	// result, ok := InterfaceToMap(req)
	// if ok != true {
	// 	fmt.Println("InterfaceToMapArray failed")
	// }

	for key, value := range result {
		fmt.Printf("%s: %s\n", key, value)
	}
}
