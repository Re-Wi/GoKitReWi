package helpers

import (
	"time"
)

//type DateTime struct {
//	time.Time
//}

//// 格式常量
//const dtLayout = "2006-01-02 15:04:05"

type PlanFilename struct {
	Project    string
	StartTime  time.Time
	FinishTime time.Time
	Status     string
	Suffix     string
}

type PlanContent struct {
	ProjectName string        `yaml:"projectName" json:"projectName"`
	StartTime   time.Time     `yaml:"startTime"   json:"startTime"  `
	FinishTime  time.Time     `yaml:"finishTime"  json:"finishTime" `
	Status      string        `yaml:"status"      json:"status"     `
	PlanList    []PlanListEle `yaml:"planList"    json:"planList"   `
}
type PlanListEle struct {
	Plan       string    `yaml:"plan"       json:"plan"      `
	Progress   uint8     `yaml:"progress"   json:"progress"  `
	PlanTime   time.Time `yaml:"planTime"   json:"planTime"  `
	FinishTime time.Time `yaml:"finishTime" json:"finishTime"`
}

//// 自定义反序列化方法，实现UnmarshalJSON()接口
//func (dt *DateTime) UnmarshalJSON(b []byte) (err error) {
//	s := strings.Trim(string(b), "\"") //去掉首尾的"
//	local, _ := time.LoadLocation("Asia/Shanghai")
//	dt.Time, err = time.ParseInLocation(dtLayout, s, local) //格式化时间格式
//	if err != nil {
//		return err
//	}
//	fmt.Printf("[dt.Time:]=%v\n", dt.Time)
//	return
//}
//
//// 自定义序列化,实现MarshalJSON()接口
//func (dt *DateTime) MarshalJSON() ([]byte, error) {
//	return []byte(fmt.Sprintf("\"%s\"", dt.Time.Format(dtLayout))), nil
//}
//
//// TODO
//// 未完成YAML自定义
//// 自定义反序列化方法，UnmarshalYAML()接口
//// Implements the Unmarshaler interface of the yaml pkg.
//func (dt *DateTime) UnmarshalYAML(b []byte) (err error) {
//	s := strings.Trim(string(b), "\"") //去掉首尾的"
//	local, _ := time.LoadLocation("Asia/Shanghai")
//	dt.Time, err = time.ParseInLocation(dtLayout, s, local) //格式化时间格式
//	if err != nil {
//		return err
//	}
//	fmt.Printf("[dt.Time:]=%v\n", dt.Time)
//	return
//}
//
//// 自定义序列化,MarshalYAML()接口
//func (dt *DateTime) MarshalYAML() ([]byte, error) {
//	return []byte(fmt.Sprintf("\"%s\"", dt.Time.Format(dtLayout))), nil
//}
