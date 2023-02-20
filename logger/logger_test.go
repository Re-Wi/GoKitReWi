package logger

import (
	"github.com/XM-GO/PandaKit/logger"
	"reflect"
	"testing"
)

func TestExportExcel(t *testing.T) {
	got := []string{"a", "d"}
	want := []string{"a", "d"}
	log := logger.InitLog("./default.log", "info")
	log.Info("OKOKOKOK~~~~~~~~")
	if !reflect.DeepEqual(want, got) {
		t.Errorf("expected:%v, got:%v", want, got)
	}
}
