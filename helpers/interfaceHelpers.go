package helpers

import (
	"fmt"
	"reflect"
)

// 你提供的 StructToMap 函数用于将结构体的字段映射为 map[string]string。
func StructToMap(input interface{}) map[string]string {
	result := make(map[string]string)
	val := reflect.ValueOf(input)

	if val.Kind() != reflect.Struct {
		fmt.Println("Input is not a struct")
		return result
	}

	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		key := typ.Field(i).Name
		value := fmt.Sprintf("%v", field.Interface())
		result[key] = value
	}

	return result
}

// 这个函数接收一个接口类型的参数，返回一个map[string]interface{}类型的值和一个bool类型的值，表示转换是否成功。在函数内部，使用类型断言将接口类型转换为map[string]interface{}类型，如果转换成功，则返回转换后的值和true；否则返回空map和false。
func InterfaceToMap(i interface{}) (map[string]interface{}, bool) {
	m, ok := i.(map[string]interface{})
	if !ok {
		// 如果断言失败，返回一个空map和false表示失败
		return make(map[string]interface{}), false
	}
	return m, true
}

// 该函数将接收一个接口类型的参数，该参数应该是一个切片，每个元素都应该是一个map[string]interface{}类型。该函数将返回一个map数组类型，每个元素都是一个map[string]interface{}类型。如果输入数据不符合要求，该函数将返回一个错误。
func InterfaceToMapArray(data interface{}) ([]map[string]interface{}, error) {
	// 使用类型断言将接口类型转为切片类型
	arr, ok := data.([]interface{})
	if !ok {
		return nil, fmt.Errorf("input data is not a slice")
	}

	// 遍历切片，将每个元素转为map[string]interface{}类型
	result := make([]map[string]interface{}, len(arr))
	for i, v := range arr {
		m, ok := v.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("element %d is not a map", i)
		}
		result[i] = m
	}

	return result, nil
}

// 通过 value.([]interface{}) 进行类型断言，将接口类型转换为 []interface{} 类型。然后遍历切片，并通过类型断言将每个元素转换为字符串类型。最终得到的字符串切片就是我们需要的 []string 类型
func InterfaceToStrSlice(value interface{}) ([]string, error) {
	slice, ok := value.([]interface{})
	if !ok {
		return nil, fmt.Errorf("Type assertion to []interface{} failed")
	}

	strSlice := make([]string, len(slice))
	for i, v := range slice {
		str, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("Type assertion to string failed")
		}
		strSlice[i] = str
	}

	return strSlice, nil
}
