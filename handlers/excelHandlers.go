package handlers

import (
	"fmt"
	"github.com/xuri/excelize/v2"
)

var (
	excel          = excelize.NewFile()
	sheetName      = "Sheet1"
	rowIndex       = 1
	headerTextData = []interface{}{"Path", "ID", "Time", "X", "Y", "Pressure", "Slot"}
	pathTextLine   = []interface{}{nil, nil, nil, nil, nil, nil, nil}
	headerJsonData = []interface{}{"ID", "X", "Y", "Count", "Signal", "CreateTime"}
	pathJsonLine   = []interface{}{nil, nil, nil, nil, nil, nil}
)

// 创建 Excel 文档
func SaveExcel(filename string) {
	// 创建一个工作表
	index, _ := excel.NewSheet(sheetName)
	// 设置单元格的值
	//err := f.SetCellValue("Sheet2", "A2", "Hello world.")
	//if err != nil {
	//	return
	//}
	//err := Excel.SetCellValue("Sheet1", "B2", 100)
	//if err != nil {
	//	return
	//}
	// 设置工作簿的默认工作表
	//excel.SetActiveSheet(excel.GetActiveSheetIndex())
	excel.SetActiveSheet(index)
	excel.SetSheetName("Sheet1", sheetName)
	// 根据指定路径保存文件
	if err := excel.SaveAs(filename); err != nil {
		fmt.Println(err)
	}
}

func InsertTextHeaderLine() {
	err := excel.SetSheetRow(sheetName, fmt.Sprintf("A%d", rowIndex), &headerTextData)
	if err != nil {
		return
	}
	rowIndex += 1
}

func InsertJsonHeaderLine() {
	err := excel.SetSheetRow(sheetName, fmt.Sprintf("A%d", rowIndex), &headerJsonData)
	if err != nil {
		return
	}
	rowIndex += 1
}

func InsertTextPathLine(filepath string) {
	pathTextLine[0] = filepath
	err := excel.SetSheetRow(sheetName, fmt.Sprintf("A%d", rowIndex), &pathTextLine)
	if err != nil {
		return
	}
	rowIndex += 1
}
func InsertJsonPathLine(filepath string) {
	pathJsonLine[0] = filepath
	err := excel.SetSheetRow(sheetName, fmt.Sprintf("A%d", rowIndex), &pathJsonLine)
	if err != nil {
		return
	}
	rowIndex += 1
}

func InsertDataLine(rowData []interface{}) {
	//fmt.Println(RowData[list_index])
	//fmt.Println(&RowData[list_index])
	err := excel.SetSheetRow(sheetName, fmt.Sprintf("A%d", rowIndex), &rowData)
	if err != nil {
		return
	}
	rowIndex += 1
}

// 读取 Excel 文档
func ReadExcel() {
	f, err := excelize.OpenFile("Book1.xlsx")
	if err != nil {
		fmt.Println(err)
		return
	}
	// 获取工作表中指定单元格的值
	cell, err := f.GetCellValue("Sheet1", "B2")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(cell)
	// 获取 Sheet1 上所有单元格
	rows, err := f.GetRows("Sheet1")
	for _, row := range rows {
		for _, colCell := range row {
			fmt.Print(colCell, "\t")
		}
		fmt.Println()
	}
}
