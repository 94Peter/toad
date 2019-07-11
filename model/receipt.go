package model

import (
	"encoding/json"
	"errors"
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

func (rm *RTModel) UpdateReceiptData(amount int, Date, Rid string) error {
	fmt.Println("UpdateReceiptData")
	arid := ""
	selectSQL := fmt.Sprintf("Select arid FROM public.receipt where Rid = '%s'", Rid)
	const sql = `Update public.receipt set amount = $1 ,date = $2 where Rid = $3;`
	db := rm.imr.GetSQLDB()

	rows, err := db.SQLCommand(selectSQL)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	for rows.Next() {
		if err := rows.Scan(&arid); err != nil {
			fmt.Println("err Scan " + err.Error())
		}
	}

	fmt.Println("arid ", arid)
	if arid == "" {
		return errors.New("not found receipt")
	}

	mdb, err := db.ConnectSQLDB()
	if err != nil {
		fmt.Println(err)
		return err
	}
	res, err := mdb.Exec(sql, amount, Date, Rid)
	//res, err := sqldb.Exec(sql, unix_time, receivable.Date, receivable.CNo, receivable.Sales)
	if err != nil {
		fmt.Println(err)
		return err
	}

	id, err := res.RowsAffected()
	if err != nil {
		fmt.Println("PG Affecte Wrong: ", err)
		return err
	}
	fmt.Println("RowsAffected: ", id)
	if id <= 0 {
		return errors.New("not found receipt")
	} else if arid != "" {
		//am.UpdateARInfo(arid)
		err := UpdateARInfo(rm.imr, arid)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}
	return nil
}

func (rm *RTModel) DeleteReceiptData(Rid string) error {

	fmt.Println("DeleteReceiptData")

	arid := ""
	selectSQL := fmt.Sprintf("Select arid FROM public.receipt where Rid = '%s'", Rid)
	const sql = `Delete FROM public.receipt where Rid = $1;`
	db := rm.imr.GetSQLDB()

	fmt.Println(selectSQL)
	rows, err := db.SQLCommand(selectSQL)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	for rows.Next() {
		if err := rows.Scan(&arid); err != nil {
			fmt.Println("err Scan " + err.Error())
		}
	}

	fmt.Println("arid ", arid)

	if arid == "" {
		return errors.New("not found receipt")
	}

	mdb, err := db.ConnectSQLDB()
	if err != nil {
		fmt.Println(err)
		return err
	}
	res, err := mdb.Exec(sql, Rid)
	//res, err := sqldb.Exec(sql, unix_time, receivable.Date, receivable.CNo, receivable.Sales)
	if err != nil {
		fmt.Println(err)
		return err
	}

	id, err := res.RowsAffected()
	if err != nil {
		fmt.Println("PG Affecte Wrong: ", err)
		return err
	}

	fmt.Println("RowsAffected: ", id)

	if id <= 0 {
		return errors.New("not found receipt")
	} else if arid != "" {
		//am.UpdateARInfo(arid)
		err := UpdateARInfo(rm.imr, arid)
		if err != nil {
			fmt.Println(err)
			return err
		}
		err = UpdateARSales(rm.imr, arid, DEL)
		if err != nil {
			return err
		}
	}

	return nil
}

func (rm *RTModel) GetReceiptData(today, end time.Time) []*Receipt {
	fmt.Println("GetReceiptData")
	//if invoiceno is null in Database return ""
	const qspl = `SELECT rid, date, cno, casename, type, name, amount, COALESCE(NULLIF(invoiceno, null),'') FROM public.receipt;`
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
	// out, err := json.Marshal(rtDataList)
	// if err != nil {
	// 	fmt.Println("err rtDataList")
	// 	return nil
	// }
	//fmt.Println(string(out))

	rm.rtList = rtDataList
	return rm.rtList
}

func (rm *RTModel) Json() ([]byte, error) {
	return json.Marshal(rm.rtList)
}
