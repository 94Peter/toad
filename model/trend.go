package model

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type TrendConf struct {
	DailyItem []string `yaml:"dailyItem"`
}

var (
	tm *trendModel

	dailyQueryType string
)

func GetTrendModel(imr interModelRes) *trendModel {
	if tm != nil {
		return tm
	}

	tm = &trendModel{
		imr: imr,
	}
	return tm
}

type trendModel struct {
	imr interModelRes
}

type dailyTrend struct {
	Category string  `json:"category"`
	Name     string  `json:"name"`
	CType    string  `json:"type"`
	Value    float64 `json:"value"`
	Unit     string  `json:"unit"`
	Msg      string  `json:"msg"`
}

func getQueryType(items []string) string {
	var queryType []string
	for _, t := range items {
		queryType = append(queryType, fmt.Sprintf(`"type"='%s'`, t))
	}
	return strings.Join(queryType, " OR ")
}

func (tr *trendModel) GetDailyData(t time.Time) []*dailyTrend {
	if dailyQueryType == "" {
		ri := GetReportItmeModel(tr.imr)
		allowType := ri.GetDailyDataItem()
		dailyQueryType = getQueryType(allowType)
	}
	if dailyQueryType == "" {
		return nil
	}
	const qtpl = `SELECT "category","type", "name","unit", "value", "msg" FROM "pica"."autogen"."dailyReport" WHERE "time" = '%s' AND "value" > 0 AND (%s)`
	db := tr.imr.GetTSDB()
	result, err := db.Query(fmt.Sprintf(qtpl, t.Format(time.RFC3339), dailyQueryType))
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

	var dailyTrendList []*dailyTrend
	var myvalue float64
	var jvalue json.Number
	for _, value := range data.Values {
		jvalue = value[5].(json.Number)
		myvalue, _ = jvalue.Float64()
		dailyTrendList = append(dailyTrendList, &dailyTrend{
			Category: value[1].(string),
			CType:    value[2].(string),
			Name:     value[3].(string),
			Unit:     value[4].(string),
			Value:    myvalue,
			Msg:      value[6].(string),
		})
	}
	return dailyTrendList
}

type row []interface{}

func (r row) addColumn(c interface{}) row {
	return append(r, c)
}

type MonthTrend struct {
	Columns []string `json:"columns"`
	Data    []row
	Total   row
}

func (mt *MonthTrend) ToJSON() ([]byte, error) {
	result := make(map[string]interface{})
	result["columns"] = mt.Columns

	// data
	dataLen := len(mt.Data)
	if dataLen == 0 {
		result["data"] = make([][]string, 0, 0)
	} else {
		dataResult := make([][]string, dataLen)
		var curRow row
		var rowLen int
		for i := 0; i < dataLen; i++ {
			curRow = mt.Data[i]
			rowLen = len(curRow)
			dataResult[i] = make([]string, rowLen)
			for j := 0; j < rowLen; j++ {
				if s, ok := curRow[j].(string); ok {
					dataResult[i][j] = s
				} else if vf, ok := curRow[j].(valueField); ok {
					dataResult[i][j] = vf.ToString()
				} else {
					dataResult[i][j] = "unknown"
				}
			}
		}
		result["data"] = dataResult
	}

	// total
	totalLen := len(mt.Total)
	if totalLen == 0 {
		result["total"] = make([]string, 0)
	} else {
		totalRow := make([]string, totalLen)
		i := 0
		for _, r := range mt.Total {
			if s, ok := r.(string); ok {
				totalRow[i] = s
			} else if vf, ok := r.(valueField); ok {
				totalRow[i] = vf.ToString()
			} else {
				totalRow[i] = "unknown"
			}
			i++
		}
		result["total"] = totalRow
	}
	return json.Marshal(result)
}

func (mt *MonthTrend) addColumn(c string) {
	mt.Columns = append(mt.Columns, c)
}

func (mt *MonthTrend) addRow(r row) {
	mt.Data = append(mt.Data, r)
	rowLen := len(r)
	if mt.Total == nil {
		mt.Total = make(row, rowLen)
		mt.Total[0] = "total"
		for i := 1; i < rowLen; i++ {
			mt.Total[i] = r[i]
		}
		return
	}
	totalLen := len(mt.Total)
	if totalLen != rowLen {
		return
	}
	for i := 1; i < rowLen; i++ {
		totalField, ok1 := mt.Total[i].(valueField)
		rowField, ok2 := r[i].(valueField)
		if ok1 && ok2 {
			totalField.subtotal(rowField)
			mt.Total[i] = totalField
		}
	}
}

type valueField struct {
	Value     float64
	Unit      string
	TotalFunc string
}

const (
	TotalFuncAdd = "add"
	TotalFuncMax = "max"

	AC_Stock        = "庫存"
	AC_Permformance = "最高紀錄"
)

func (vf *valueField) subtotal(new valueField) {
	if new.TotalFunc == TotalFuncMax {
		if new.Value > vf.Value {
			vf.Value = new.Value
		}
	} else {
		vf.Value = vf.Value + new.Value
	}

}

func (vf *valueField) ToString() string {
	if vf.Unit == "萬" {
		return fmt.Sprintf("%.2f", vf.Value)
	}
	return fmt.Sprintf("%.1f", vf.Value)
}

// 取得全區動態
func (tr *trendModel) GetAreaData(start, end time.Time) *MonthTrend {
	ri := GetReportItmeModel(tr.imr)
	allowType := ri.GetMonthDataItem()
	queryType := getQueryType(allowType)
	const qtpl = `SELECT "category","type","unit", "value" FROM "pica"."autogen"."dailyReport" WHERE "time" >= '%s' AND "time" <= '%s' AND (%s) GROUP BY "category"`
	q := fmt.Sprintf(qtpl, start.Format(time.RFC3339), end.Format(time.RFC3339), queryType)
	dc := GetDictionaryCategory(tr.imr.GetDB())
	var rawNames []string
	for _, o := range dc.Order {
		if o == nil {
			continue
		}
		rawNames = append(rawNames, o.Name)
	}

	db := tr.imr.GetTSDB()
	result, err := db.Query(q)
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
	mm := GetMemberModel(tr.imr)
	dataMap := make(map[string]map[string]valueField)
	var dataValue json.Number
	var floatValue float64
	var categoryName string
	var entrustValue valueField
	var addEntrustValue float64
	var ok bool
	for _, s := range influxR.Series {
		itemMap := make(map[string]valueField)
		for _, v := range s.Values {
			dataValue = v[4].(json.Number)
			floatValue, _ = dataValue.Float64()
			itemMap[v[2].(string)] = valueField{
				Value:     floatValue,
				Unit:      v[3].(string),
				TotalFunc: TotalFuncAdd,
			}
		}
		categoryName = s.Tags["category"]

		if entrustValue, ok = itemMap["委託"]; ok {
			addEntrustValue = entrustValue.Value
		} else {
			addEntrustValue = 0
		}
		itemMap[AC_Stock] = valueField{
			Value:     float64(mm.GetCategoryStock(categoryName)) + addEntrustValue,
			Unit:      "件",
			TotalFunc: TotalFuncAdd,
		}
		itemMap[AC_Permformance] = valueField{
			Value:     float64(mm.GetCategoryMaxPerformace(categoryName)),
			Unit:      "",
			TotalFunc: TotalFuncMax,
		}
		dataMap[categoryName] = itemMap
	}

	mt := &MonthTrend{}
	firstColumnName := "分店"
	mt.addColumn(firstColumnName)
	allowType = append(allowType, AC_Stock, AC_Permformance)
	reportTitles := append(ri.GetReportTitle(), "庫存(件)", "最高紀錄(萬)")
	for _, item := range reportTitles {
		mt.addColumn(item)
	}
	var dm map[string]valueField
	var vf valueField

	var u *User
	for _, rawName := range rawNames {
		r := row{}
		r = r.addColumn(rawName)
		if u = mm.GetMember(rawName); u != nil {
			r = r.addColumn(u.GetMedalToString())
		} else {
			r = r.addColumn("0,0,0,0")
		}
		dm, ok = dataMap[rawName]
		if !ok {
			for _, _ = range allowType {
				r = r.addColumn("-")
			}
		} else {
			for _, item := range allowType {
				vf, ok = dm[item]
				if !ok {
					r = r.addColumn("-")
				} else {
					r = r.addColumn(vf)
				}
			}
		}
		mt.addRow(r)
	}
	return mt
}

// 取得分區動態
func (tr *trendModel) GetCategoryData(c string, start, end time.Time) *MonthTrend {
	ri := GetReportItmeModel(tr.imr)
	allowType := ri.GetMonthDataItem()
	queryType := getQueryType(allowType)
	const qtpl = `SELECT "name","type","unit", "value" FROM "pica"."autogen"."dailyReport" WHERE "time" >= '%s' AND "time" <= '%s' AND "category" = '%s' AND (%s) GROUP BY "account"`
	q := fmt.Sprintf(qtpl, start.Format(time.RFC3339), end.Format(time.RFC3339), c, queryType)

	mm := GetMemberModel(tr.imr)
	userList := mm.GetUserByCategory(c)
	var rawNames []string
	for _, u := range userList {
		rawNames = append(rawNames, u["name"].(string))
	}

	db := tr.imr.GetTSDB()
	result, err := db.Query(q)
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
	dataMap := make(map[string]map[string]valueField)
	var dataValue json.Number
	var floatValue float64
	var categoryName string
	var entrustValue valueField
	var addEntrustValue float64
	var ok bool
	for _, s := range influxR.Series {
		itemMap := make(map[string]valueField)
		for _, v := range s.Values {
			dataValue = v[4].(json.Number)
			floatValue, _ = dataValue.Float64()
			itemMap[v[2].(string)] = valueField{
				Value:     floatValue,
				Unit:      v[3].(string),
				TotalFunc: TotalFuncAdd,
			}
		}
		categoryName = s.Tags["account"]
		m := mm.GetMember(categoryName)
		categoryName = m.Name

		if entrustValue, ok = itemMap["委託"]; ok {
			addEntrustValue = entrustValue.Value
		} else {
			addEntrustValue = 0
		}
		itemMap[AC_Stock] = valueField{
			Value:     float64(m.Stock) + addEntrustValue,
			Unit:      "件",
			TotalFunc: TotalFuncAdd,
		}
		itemMap[AC_Permformance] = valueField{
			Value:     float64(m.Performance),
			Unit:      "",
			TotalFunc: TotalFuncMax,
		}
		dataMap[categoryName] = itemMap
	}
	mt := &MonthTrend{}
	firstColumnName := "姓名"

	mt.addColumn(firstColumnName)
	allowType = append(allowType, AC_Stock, AC_Permformance)

	reportTitles := append(ri.GetReportTitle(), "庫存(件)", "最高紀錄(萬)")
	for _, item := range reportTitles {
		mt.addColumn(item)
	}

	var dm map[string]valueField
	var vf valueField
	var u *User
	for _, rawName := range rawNames {
		r := row{}
		r = r.addColumn(rawName)
		if u = mm.GetMember(rawName); u != nil {
			r = r.addColumn(u.GetMedalToString())
		} else {
			r = r.addColumn("0,0,0,0")
		}
		dm, ok = dataMap[rawName]
		if !ok {
			for _, _ = range allowType {
				r = r.addColumn("-")
			}
		} else {
			for _, item := range allowType {
				vf, ok = dm[item]
				if !ok {
					r = r.addColumn("-")
				} else {
					r = r.addColumn(vf)
				}
			}
		}
		mt.addRow(r)
	}
	return mt
}
