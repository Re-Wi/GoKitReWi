package databases

import (
	"fmt"
	"gorm.io/gorm"
	"log"
	"reflect"
	"testing"
)

type Test struct {
	ID   uint
	Name string
	Age  uint8
}

func TestExportExcel(t *testing.T) {
	var (
		Db *gorm.DB // gorm
	)

	dbGorm := DbGorm{Type: "sqlite3"}
	dbGorm.Dsn = "./sqlite3.db"
	dbGorm.MaxIdleConns = 10
	dbGorm.MaxOpenConns = 10
	Db = dbGorm.GormInit()
	test := Test{Name: "Jinzhu", Age: 18}
	fmt.Print(test.Name)

	test = Test{Name: "ReWi", Age: 22}
	fmt.Print(test.Name)
	got := test.Name

	result := Db.Create(&test) // 通过数据的指针来创建
	log.Print(result)
	want := "ReWi"
	if !reflect.DeepEqual(want, got) {
		t.Errorf("expected:%v, got:%v", want, got)
	}
}
