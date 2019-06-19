package model

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/94peter/pica/resource/db"
)

const (
	reportItemC = "reportItem"
)

var (
	riModel *reportItemModel
)

func GetReportItmeModel(mr interModelRes) *reportItemModel {
	if riModel != nil {
		return riModel
	}
	riModel = &reportItemModel{
		db:  mr.GetDB(),
		res: mr,
	}
	riModel.Load()
	return riModel
}

type ReportItem struct {
	Display string `json:"name"`    // 項目名稱
	Unit    string `json:"unit"`    // 單位
	HasText bool   `json:"hasText"` // 是否含動態說明
	IsShow  bool   `json:"isShow"`  // 是否顯示在動態報表
}

func (ri *ReportItem) IsValid() bool {
	if ri.Display == "" {
		return false
	}
	if ri.Unit == "" {
		return false
	}
	return true
}

type reportItemModel struct {
	db  db.InterDB
	res interModelRes

	Item []*ReportItem
}

func (rim *reportItemModel) Save() error {
	if rim.Item == nil {
		return errors.New("no item")
	}

	return rim.db.C(reportItemC).Save(rim)
}

func (rim *reportItemModel) Load() error {
	if rim.db == nil {
		return errors.New("db not set")
	}
	err := rim.db.C(reportItemC).GetByID(rim.GetID(), rim)
	if err != nil {
		return err
	}
	return nil
}

func (rim *reportItemModel) Json() ([]byte, error) {
	return json.Marshal(rim.Item)
}

func (rim *reportItemModel) GetID() string {
	return "1"
}

func (rim *reportItemModel) GetDailyDataItem() []string {
	return rim.res.GetTrendItems()
}

func (rim *reportItemModel) GetMonthDataItem() []string {
	var result []string
	for _, item := range rim.Item {
		if item.IsShow {
			result = append(result, item.Display)
		}
	}
	return result
}

func (rim *reportItemModel) GetReportTitle() []string {
	var result []string
	for _, item := range rim.Item {
		if item.IsShow {
			result = append(result, fmt.Sprintf("%s(%s)", item.Display, item.Unit))
		}
	}
	return result
}
