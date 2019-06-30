package model

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/94peter/toad/resource/db"
)

type Receipt struct {
	Rid           string    `json:"-"` //no return this key
	ARid          string    `json:"id"`
	Date          time.Time `json:"date"`
	CNo           string    `json:"contractNo"`
	CaseName      string    `json:"caseName"`
	CustomertType string    `json:"customertType"`
	Name          string    `json:"customerName"`
	Amount        int       `json:"amount"`    //收款
	InvoiceNo     string    `json:"invoiceNo"` //發票號碼
}

type RTModel struct {
	imr    interModelRes
	db     db.InterSQLDB
	rtList []*Receipt
}

var (
	rm *RTModel
)

func GetRTModel(imr interModelRes) *RTModel {
	if rm != nil {
		return rm
	}

	rm = &RTModel{
		imr: imr,
	}
	return rm
}

func (rm *RTModel) GetReceiptData(today, end time.Time) []*Receipt {
	fmt.Println("GetReceiptData")
	//if invoiceno is null in Database return ""
	const qspl = `SELECT arid, date, cno, casename, type, name, amount, COALESCE(NULLIF(invoiceno, null),'') FROM public.receipt;`
	db := rm.imr.GetSQLDB()
	rows, err := db.SQLCommand(fmt.Sprintf(qspl))
	if err != nil {
		return nil
	}
	var rtDataList []*Receipt
	fmt.Println("SQLCommand Done")
	for rows.Next() {
		var rt Receipt

		// if err := rows.Scan(&r.ARid, &s); err != nil {
		// 	fmt.Println("err Scan " + err.Error())
		// }
		if err := rows.Scan(&rt.ARid, &rt.Date, &rt.CNo, &rt.CaseName, &rt.CustomertType, &rt.Name, &rt.Amount, &rt.InvoiceNo); err != nil {
			fmt.Println("err Scan " + err.Error())
		}

		rtDataList = append(rtDataList, &rt)
	}
	fmt.Println("rtDataList Done")
	out, err := json.Marshal(rtDataList)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(out))

	rm.rtList = rtDataList
	return rm.rtList
}

func (rm *RTModel) Json() ([]byte, error) {
	return json.Marshal(rm.rtList)
}
