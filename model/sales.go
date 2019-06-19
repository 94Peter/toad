package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/94peter/pica/resource/db"
)

var (
	sm *salesModel
)

func GetSalesModel(imr interModelRes) *salesModel {
	if sm != nil {
		return sm
	}
	rs := &reportState{}
	rs.Load(imr.GetDB())

	sm = &salesModel{
		imr:  imr,
		tsdb: imr.GetTSDB(),
		rs:   rs,
	}
	return sm
}

type Answer struct {
	ReportItem
	Value float64   `json:"value"`
	Msg   string    `json:"msg"`
	Date  time.Time `json:"-"`
	U     *User     `json:"-"`
}

const (
	answerM = "dailyReport"

	reportStateC = "reportState"

	AllowReportHour = 18
)

func (a *Answer) valid() bool {
	if a.U == nil {
		return false
	}
	return true
}

func (a *Answer) GetMeasurement() string {
	return answerM
}

func (a *Answer) GetTags() map[string]string {
	return map[string]string{
		"type":     a.Display,
		"category": a.U.Category,
		"account":  a.U.Account,
	}
}

func (a *Answer) GetFields() map[string]interface{} {
	return map[string]interface{}{
		"name":  a.U.Name,
		"value": a.Value,
		"unit":  a.Unit,
		"msg":   a.Msg,
	}
}
func (a *Answer) GetTime() time.Time {
	return time.Date(a.Date.Year(), a.Date.Month(), a.Date.Day(), 0, 0, 0, 0, a.Date.Location())
}

type reportState struct {
	ReportDate time.Time
	RS         map[string]time.Time
}

// 是否為今日的狀態 (每日的18以後才要刷新)
func (rs *reportState) isRefreshState(loc *time.Location) bool {
	now := time.Now().In(loc)
	yearMonth := (now.Year() == rs.ReportDate.Year() && now.Month() == rs.ReportDate.Month())
	day := (now.Day()-rs.ReportDate.Day() == 1)
	return yearMonth && day && now.Hour() >= AllowReportHour
}

func (rs *reportState) refresh(loc *time.Location) {
	now := time.Now().In(loc)
	rs.ReportDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	rs.RS = make(map[string]time.Time)
}

func (rs *reportState) addState(account string, reportTime time.Time) bool {
	d := reportTime.Sub(rs.ReportDate)
	if d.Nanoseconds() < 0 {
		return false
	}
	if d.Nanoseconds() > 0 {
		rs.ReportDate = reportTime
		rs.RS = make(map[string]time.Time)
	}
	if rs.RS == nil {
		rs.RS = make(map[string]time.Time)
	}
	rs.RS[account] = time.Now()
	return true
}

func (rs *reportState) Save(mydb db.InterDB) error {
	if rs.RS == nil {
		return errors.New("no data")
	}

	return mydb.C(reportStateC).Save(rs)
}

func (rs *reportState) Load(mydb db.InterDB) error {
	if mydb == nil {
		return errors.New("db not set")
	}
	err := mydb.C(reportStateC).GetByID(rs.GetID(), rs)
	if err != nil {
		return err
	}
	return nil
}

func (rs *reportState) GetID() string {
	return "1"
}

type salesModel struct {
	imr  interModelRes
	tsdb db.InterTSDB

	rs *reportState
}

// 業績回報
func (sm *salesModel) Report(answer []*Answer) error {
	l := len(answer)
	if l == 0 {
		return errors.New("no answer")
	}
	points := make([]db.InterPoint, l)
	for i := 0; i < l; i++ {
		if !answer[i].valid() {
			return errors.New("invalid answer")
		}
		points[i] = answer[i]
	}
	fmt.Println("report")
	err := sm.tsdb.Save(points...)
	if err != nil {
		return err
	}
	acc := answer[0].U.Account
	reportDate := answer[0].Date
	if isAdd := sm.rs.addState(acc, reportDate); isAdd {
		return sm.rs.Save(sm.imr.GetDB())
	}
	return nil
}

func (sm *salesModel) GetAnswer(t time.Time, acc string) []*Answer {
	const qtpl = `SELECT "type", "unit", "value", "msg" FROM "pica"."autogen"."dailyReport" WHERE "time" = '%s' AND "account" = '%s'`
	result, err := sm.tsdb.Query(fmt.Sprintf(qtpl, t.Format(time.RFC3339), acc))
	if err != nil {
		return nil
	}

	if len(result) == 0 {
		return nil
	}
	influxR := result[0]
	if len(influxR.Series) == 0 {
		return nil
	}
	data := influxR.Series[0]

	type queryAnwser struct {
		SType string
		Unit  string
		Value float64
		Msg   string
	}
	var answerNumber json.Number
	answerMap := make(map[string]queryAnwser)
	for _, value := range data.Values {
		answerNumber = value[3].(json.Number)
		valueFloat64, _ := answerNumber.Float64()
		qa := queryAnwser{
			SType: value[1].(string),
			Unit:  value[2].(string),
			Value: valueFloat64,
		}
		if len(value) == 5 {
			qa.Msg = value[4].(string)
		}
		answerMap[qa.SType] = qa
	}

	riModel := GetReportItmeModel(sm.imr)
	items := riModel.Item
	length := len(items)
	if length == 0 {
		return nil
	}
	answers := make([]*Answer, length)
	var myAnswer queryAnwser
	for i, item := range items {
		myAnswer = answerMap[item.Display]
		answers[i] = &Answer{
			ReportItem: *item,
			Msg:        myAnswer.Msg,
			Value:      myAnswer.Value,
		}
	}
	return answers
}

//  取得未回報的清單
func (sm *salesModel) GetNotReportor(category string) []map[string]interface{} {
	if sm.rs.isRefreshState(sm.imr.GetLocation()) {
		sm.rs.refresh(sm.imr.GetLocation())
	}
	mm := GetMemberModel(sm.imr)
	users := mm.GetUserByCategory(category)
	var acc string
	var result []map[string]interface{}
	for _, u := range users {
		acc = u["account"].(string)
		if _, ok := sm.rs.RS[acc]; !ok {
			result = append(result, u)
		}
	}
	return result
}
