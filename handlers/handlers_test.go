package handlers

import (
	"GinGoReWi/utils/helpers"
	"fmt"
	"reflect"
	"testing"
	"time"
)

func TestExportExcel(t *testing.T) {
	got := []string{"a", "d"}
	want := []string{"a", "d"}
	if !reflect.DeepEqual(want, got) {
		t.Errorf("expected:%v, got:%v", want, got)
	}
}
func TestPlanFilenameParsing(t *testing.T) {
	local, _ := time.LoadLocation("Asia/Shanghai")
	got := PlanFilenameParsing("abcd__2023-02-07__2023-04-09.F.yaml")
	startTime, _ := time.ParseInLocation("2006-01-02", "2023-02-07", local)
	endTime, _ := time.ParseInLocation("2006-01-02", "2023-04-09", local)
	want := helpers.PlanFilename{"abcd", startTime, endTime, "F", "yaml"}
	fmt.Println(startTime)
	fmt.Println(endTime)
	fmt.Printf("%s,%s,%s,%s,%s \r\n", got.Project, got.StartTime.Format("2006-01-02 15:04:05"), got.FinishTime, got.Status, got.Suffix)
	if !reflect.DeepEqual(&want, got) {
		t.Errorf("expected:%v, got:%v", want, got)
	}
}

func TestPlanContentParsing(t *testing.T) {
	fmt.Println("~~~~~~~~~~~~~~~~~~~~~~~~~~~")
	plan, err := PlanContentParsing("../../resource/template/yaml/abcd__2023-02-07__2023-04-09.F.yaml")
	fmt.Println(plan, err)
	got := uint8(60)
	want := plan.PlanList[2].Progress
	if !reflect.DeepEqual(want, got) {
		t.Errorf("expected:%v, got:%v", want, got)
	}
}

func TestPlanContentSave(t *testing.T) {
	planList := make([]helpers.PlanListEle, 0)
	var planEle helpers.PlanListEle
	local, _ := time.LoadLocation("Asia/Shanghai")
	planEle.Plan = "abcd"
	planEle.PlanTime, _ = time.ParseInLocation("2006-01-02 15:04:05", "2023-02-13 12:00:00", local)   //格式化时间格式
	planEle.FinishTime, _ = time.ParseInLocation("2006-01-02 15:04:05", "2023-02-14 12:00:00", local) //格式化时间格式
	planEle.Progress = 10
	planList = append(planList, planEle)
	planList = append(planList, planEle)
	planList = append(planList, planEle)
	planList = append(planList, planEle)
	plan := helpers.PlanContent{ProjectName: "1234", PlanList: planList}
	err := PlanContentSave(helpers.CheckAndCreateDir("../../resource/plan")+"/test001.yaml", &plan)
	fmt.Println(err)
	//got := uint32(2)
	//want := plan.PlanList[2].Id
	//if !reflect.DeepEqual(want, got) {
	//	t.Errorf("expected:%v, got:%v", want, got)
}
