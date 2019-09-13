package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/94peter/toad/resource/db"
)

type Receipt struct {
	Rid           string    `json:"id"`
	ARid          string    `json:"-"` //no return this key
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

	const sql = `Update public.receipt set amount = $1 ,date = $2 where Rid = $3;`
	db := rm.imr.GetSQLDB()

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
	}

	return nil
}

func (rm *RTModel) DeleteReceiptData(Rid string) error {

	fmt.Println("DeleteReceiptData:")

	const sql = `Delete FROM public.receipt where Rid = $1;`
	db := rm.imr.GetSQLDB()

	mdb, err := db.ConnectSQLDB()
	if err != nil {
		fmt.Println(err)
		return err
	}
	res, err := mdb.Exec(sql, Rid)
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
	}

	return nil
}

func (rm *RTModel) GetReceiptData(begin, end string) []*Receipt {

	//if invoiceno is null in Database return ""
	//const qspl = `SELECT rid, date, cno, casename, type, name, amount, COALESCE(NULLIF(invoiceno, null),'') FROM public.receipt;`
	const qspl = `SELECT R.rid, R.date, AR.cno, AR.casename, AR.type, AR.name , R.amount, COALESCE(NULLIF(I.invoice, null),'')
				  FROM public.receipt R
				  inner join public.ar AR on AR.arid = R.arid
				  left join public.invoice I on  I.Rid = R.rid
				  where R.date >= '%s' and R.date <= '%s'`
	db := rm.imr.GetSQLDB()
	rows, err := db.SQLCommand(fmt.Sprintf(qspl, begin, end))
	if err != nil {
		return nil
	}

	var rtDataList []*Receipt
	for rows.Next() {
		var rt Receipt

		// if err := rows.Scan(&r.ARid, &s); err != nil {
		// 	fmt.Println("err Scan " + err.Error())
		// }
		if err := rows.Scan(&rt.Rid, &rt.Date, &rt.CNo, &rt.CaseName, &rt.CustomertType, &rt.Name, &rt.Amount, &rt.InvoiceNo); err != nil {
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

func (rm *RTModel) CreateReceipt(rt *Receipt) (err error) {
	fmt.Println("CreateReceipt : arid is ", rt.ARid)

	//arid exist
	//(加總歷史收款明細 + 此筆單子) <= 應收款項的收款
	const sql = `INSERT INTO public.receipt (Rid, Date, Amount, ARid)
				SELECT * FROM (SELECT $1::varchar(50), to_timestamp($2,'YYYY-MM-DD hh24:mi:ss'), $3::INTEGER , $4::varchar(50)) AS tmp
				WHERE  
					EXISTS ( SELECT arid from public.ar ar WHERE arid = $4 ) 
				and ( select $3 + COALESCE(SUM(amount),0) FROM public.receipt  where arid = $4 ) <=  (SELECT amount from public.ar ar WHERE arid = $4)				
				;`
	//and ( select sum(amount)+$3 FROM public.receipt  where arid = $4 group by arid ) <=  (SELECT amount from public.ar ar WHERE arid = $4);`

	interdb := rm.imr.GetSQLDB()
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}
	fmt.Println("sqldb Exec")

	out, err := json.Marshal(rt)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(out))
	fmt.Println(string(sql))
	t := time.Now().Unix()
	res, err := sqldb.Exec(sql, t, rt.Date, rt.Amount, rt.ARid)
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
	fmt.Println(id)

	if id == 0 {
		return errors.New("Invalid operation, may be the ID does not exist or amount is not vaild")
	}
	// Rid := fmt.Sprintf("%v", t)
	// rt.setRid(Rid)
	// err = UpdateARInfo(am.imr, rt.ARid)
	// if err != nil {
	// 	return err
	// }
	// fmt.Println("UpdateARSales [GO]")
	// err = UpdateARSales(am.imr, rt.ARid, ADD)
	// if err != nil {
	// 	return err
	// }

	fmt.Println("CreateCommission [GO]")
	//init cm on createReceiptEndpoint  at ar.go(api)
	Rid := fmt.Sprintf("%v", t)
	rt.setRid(Rid)
	err = cm.CreateCommission(rt)
	if err != nil {
		return err
	}

	fmt.Println("UpdateDeductRid [GO]")
	err = decuctModel.UpdateDeductRid(rt.ARid)
	if err != nil {
		fmt.Println(err.Error())
	}

	return nil
}
