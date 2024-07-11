package logger

import (
	"reflect"
	"testing"
)

func TestExportExcel(t *testing.T) {
	got := []string{"a", "d"}
	want := []string{"a", "d"}
	log := InitLog("./default", "trace")
	log.Trace("Trace~~~")
	log.Debug("Debug~~~")
	log.Info("Info~~~")
	log.Warn("Warn~~~")
	log.Error("Error~~~")
	// log.Panic("Panic~~~")
	if !reflect.DeepEqual(want, got) {
		t.Errorf("expected:%v, got:%v", want, got)
	}
}
