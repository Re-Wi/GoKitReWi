package helpers

var (
	// {"Path", "ID", "Time", "X", "Y", "Pressure", "Slot"}
	FigureList = [][]interface{}{
		{nil, nil, nil, nil, nil, nil, nil},
		{nil, nil, nil, nil, nil, nil, nil},
		{nil, nil, nil, nil, nil, nil, nil},
		{nil, nil, nil, nil, nil, nil, nil},
		{nil, nil, nil, nil, nil, nil, nil},
		{nil, nil, nil, nil, nil, nil, nil},
		{nil, nil, nil, nil, nil, nil, nil},
		{nil, nil, nil, nil, nil, nil, nil},
		{nil, nil, nil, nil, nil, nil, nil},
		{nil, nil, nil, nil, nil, nil, nil},
	}
)

func UpdateFigureList(list_index int64, col_index int, newData interface{}) {
	FigureList[list_index][col_index] = newData
}
