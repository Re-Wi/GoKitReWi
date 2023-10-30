package helpers

import "strconv"

func ConvertScientificToInteger(scientific string) (int64, error) {
	// 将科学计数法字符串解析为浮点数
	f, err := strconv.ParseFloat(scientific, 64)
	if err != nil {
		return 0, err
	}

	// 将浮点数转换为整数
	i := int64(f)

	// 将整数乘以1000，得到最终的结果
	result := i * 1000

	return result, nil
}
