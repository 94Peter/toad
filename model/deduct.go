package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/94peter/toad/resource/db"
)

//`json:"id"` 回傳重新命名
type Deduct struct {
	ARid        string    `json:"-"`
	Did         string    `json:"id"`
	Status      string    `json:"status"`
	Date        time.Time `json:"date"`
	Fee         int       `json:"fee"`
	Description string    `json:"description"`
	Item        string    `json:"item"`
	ReceiveDate time.Time `json:"receiveDate"`
	CNo         string    `json:"contractNo"`
	CaseName    string    `json:"caseName"`
	Type        string    `json:"type"`
}

type NullTime struct {
	Time  time.Time
	Valid bool // Valid is true if Time is not NULL
}

// type Receipt struct {
// 	Rid       string
// 	Date      time.Time `json:"date"`
// 	CNo       string
// 	Customer  customer
// 	CaseName  string
// 	ARid      string `json:"id"`
// 	Amount    int    `json:"amount"` //收款
// 	InvoiceNo string //發票號碼
// }

var (
	decuctModel *DeductModel
)

func GetDecuctModel(imr interModelRes) *DeductModel {
	if decuctModel != nil {
		return decuctModel
	}

	decuctModel = &DeductModel{
		imr: imr,
	}
	return decuctModel
}

type DeductModel struct {
	imr        interModelRes
	db         db.InterSQLDB
	deductList []*Deduct
}

//refer https://stackoverflow.com/questions/24564619/nullable-time-time-in-golang
func (nt *NullTime) Scan(value interface{}) error {
	nt.Time, nt.Valid = value.(time.Time)
	return nil
	//just keep the example
	// if nt.Valid {
	// 	// use nt.Time

	// } else {
	// 	// NULL value
	// }
}

func (decuctModel *DeductModel) GetDeductData(by_m, ey_m, mtype string) []*Deduct {

	//
	const qspl = `SELECT D.Did, D.date , D.status, D.item, D.fee, D.Description, R.date, AR.CNo, AR.CaseName, AR.type FROM public.deduct as D 
				inner join public.ar as AR on AR.arid = D.arid
				Left join public.receipt as R on R.rid = D.rid
				where (R.date >= '%s' and R.date < ('%s'::date + '1 month'::interval) or R.date is null ) 
				and (D.item like '%s' OR  D.status like '%s');`
	//where D.rid = R.rid;`
	//const qspl = `SELECT arid,sales	FROM public.ar;`
	fmt.Println(fmt.Sprintf(qspl, by_m+"-01", ey_m+"-01", mtype, mtype))
	db := decuctModel.imr.GetSQLDB()
	rows, err := db.SQLCommand(fmt.Sprintf(qspl, by_m+"-01", ey_m+"-01", mtype, mtype))
	if err != nil {
		fmt.Println(err)
		return nil
	}
	var deductDataList []*Deduct

	for rows.Next() {
		var d Deduct
		/*null time cannot scan into time.Time */
		var Ddate, RDate NullTime

		if err := rows.Scan(&d.Did, &Ddate, &d.Status, &d.Item, &d.Fee, &d.Description, &RDate, &d.CNo, &d.CaseName, &d.Type); err != nil {
			fmt.Println("err Scan " + err.Error())
		}

		d.Date = Ddate.Time
		d.ReceiveDate = RDate.Time

		deductDataList = append(deductDataList, &d)
	}

	decuctModel.deductList = deductDataList
	return decuctModel.deductList

}

func (decuctModel *DeductModel) Json() ([]byte, error) {
	return json.Marshal(decuctModel.deductList)
}

func (decuctModel *DeductModel) CreateDeduct(deduct *Deduct) (err error) {
	fmt.Println("arid:", deduct.ARid)
	/*為了在Deduct Table中 找到對應的收款明細，以便取得收款時間。
	//若先建立應扣款項(所以會找不到應收款項Rid)，Rid就會是null
	*/
	const sql = `INSERT INTO public.deduct (did, arid, item, description, fee, rid)
				WITH  vals  AS (VALUES ($1, $2, $3, $4, $5::integer)) 
				SELECT v.* , r.rid  FROM vals as v
				Left join public.receipt AS r on r.date = (select MIN(Date) FROM public.receipt where arid = $2) and v.column2 = r.arid limit 1;`

	interdb := decuctModel.imr.GetSQLDB()
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}

	fakeid := time.Now().Unix()
	fmt.Println("fakeid did:", fakeid)
	res, err := sqldb.Exec(sql, fakeid, deduct.ARid, deduct.Item, deduct.Description, deduct.Fee)
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
		return errors.New("Invalid operation, CreateDeduct")
	}
	return nil
}
func (decuctModel *DeductModel) DeleteDeduct(ID string) (err error) {

	const sql = `DELETE FROM public.deduct WHERE Did=$1 and  status != '已支付';`

	interdb := decuctModel.imr.GetSQLDB()
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}

	res, err := sqldb.Exec(sql, ID)
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
		return errors.New("Invalid operation, Delete Deduct")
	}

	return nil
}

func (decuctModel *DeductModel) UpdateDeduct(Did, status, date string) (err error) {

	// const sql = `UPDATE public.deduct
	// 			SET status=$1
	// 			WHERE did = $2;`

	sql := fmt.Sprintf("UPDATE public.deduct Set status = $1, date = $2 Where did = $3")
	// if mtype == "date" {
	// 	sql = fmt.Sprintf("UPDATE public.deduct Set %s = to_timestamp($1 ,'YYYY-MM-DD hh24:mi:ss') Where did = $2", mtype)
	// }

	// if mtype == "status" {
	// 	sql = fmt.Sprintf("UPDATE public.deduct Set %s = $1 Where did = $2", mtype)
	// }
	interdb := decuctModel.imr.GetSQLDB()
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}

	res, err := sqldb.Exec(sql, status, getNil(date), Did)
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
		return errors.New("Invalid operation, UpdateDeduct")
	}

	return nil
}
func (decuctModel *DeductModel) UpdateDeductRid(ARid string) (err error) {
	const sql = `Update public.deduct 
				set rid = tmp.ReceiptID
				FROM (
					select D.arid as temID , R.rid as ReceiptID FROM public.deduct D , public.receipt R
					where D.arid = R.arid and R.date = (select MIN(Date) FROM public.receipt where D.arid = $1)
				) as tmp where arid = tmp.temID`

	// 	/*
	// Update public.deduct D
	// set D.rid = ?
	// FROM (*/

	// 	select MIN(Date) FROM public.receipt  where arid = '1566575681'
	// -- )as tmp
	// --where D.arid = '1566575681'
	// and ( SELECT rid from public.deduct WHERE arid = '1566575681') is null

	interdb := decuctModel.imr.GetSQLDB()
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}

	// fakeid := time.Now().Unix()

	res, err := sqldb.Exec(sql, ARid)
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
		return errors.New("Invalid operation, UpdateDeductRid")
	}
	return nil
}

func getNil(msg string) (response *string) {
	if msg == "" {
		return nil
	} else {
		return &msg
	}
}
