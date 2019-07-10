package model

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/94peter/toad/resource/db"
)

type CModel struct {
	imr   interModelRes
	db    db.InterSQLDB
	cList []*Commission
}

var (
	cm *CModel
)

type Commission struct {
	Bid     string    `json:"-"`  //no return this key 業務人員ID
	Rid     string    `json:"id"` //收據ID
	Date    time.Time `json:"date"`
	Item    string    `json:"item"`    //合約+案名+買賣方
	Amount  int       `json:"amount"`  //金額
	Fee     int       `json:"fee"`     //口款金額收款
	SR      float64   `json:"sr"`      //實績 sales report or sales records
	Bouns   float64   `json:"bouns"`   //獎金
	Percent float64   `json:"percent"` //比例
	Bname   string    `json:"bname"`   //業務姓名
}

func GetCModel(imr interModelRes) *CModel {
	if cm != nil {
		return cm
	}

	cm = &CModel{
		imr: imr,
	}
	return cm
}

func (cm *CModel) GetCommissiontData(today, end time.Time) []*Commission {
	fmt.Println("GetCommissiontData")
	//if invoiceno is null in Database return ""
	const qspl = `SELECT rid, date, item, amount, fee, sr, bouns, percent, bname FROM public.commission;`
	db := cm.imr.GetSQLDB()
	rows, err := db.SQLCommand(fmt.Sprintf(qspl))
	if err != nil {
		fmt.Println(err)
		return nil
	}
	var cDataList []*Commission
	fmt.Println("SQLCommand Done")
	for rows.Next() {
		var c Commission

		// if err := rows.Scan(&r.ARid, &s); err != nil {
		// 	fmt.Println("err Scan " + err.Error())
		// }
		if err := rows.Scan(&c.Rid, &c.Date, &c.Item, &c.Amount, &c.Fee, &c.SR, &c.Bouns, &c.Percent, &c.Bname); err != nil {
			fmt.Println("err Scan " + err.Error())
		}

		cDataList = append(cDataList, &c)
	}
	fmt.Println("cDataList Done")
	out, err := json.Marshal(cDataList)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(out))

	cm.cList = cDataList
	return cm.cList
}

func (cm *CModel) Json() ([]byte, error) {
	return json.Marshal(cm.cList)
}
