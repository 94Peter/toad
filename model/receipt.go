package model

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"toad/resource/db"
)

type Receipt struct {
	Rid            string    `json:"id"`
	ARid           string    `json:"-"` //no return this key
	Date           time.Time `json:"date"`
	CompletionDate time.Time `json:"completionDate"` // 成交日
	CNo            string    `json:"contractNo"`
	CaseName       string    `json:"caseName"`
	CustomerType   string    `json:"customertType"`
	Name           string    `json:"customerName"`
	Amount         int       `json:"amount"`      //收款
	Fee            int       `json:"fee"`         //收款
	Item           string    `json:"item"`        //項目
	Description    string    `json:"description"` //備註
	//InvoiceNo    string    `json:"invoiceNo"` //發票號碼
	// ---
	InvoiceData []*Invoice `json:"invoiceData"`
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

func (rm *RTModel) UpdateReceiptData(input *Receipt, dbname string) error {

	db := rm.imr.GetSQLDBwithDbname(dbname)

	mdb, err := db.ConnectSQLDB()
	if err != nil {
		fmt.Println(err)
		return err
	}

	r := rm.GetReceiptDataByID(mdb, input.Rid)
	if r.Rid == "" {
		return errors.New("not found receipt")
	}

	c := cm.GetCommissionDataByRID(mdb, input.Rid)
	if c.Bsid != "" {
		return errors.New("此收款傭金已納入薪資")
	}

	_, err = salaryM.CheckValidCloseDate(r.Date, dbname, mdb)
	if err != nil {
		return err
	}

	fmt.Println("UpdateReceiptData:", input)

	//const sql = `Update public.receipt set amount = $1 ,date = $2 where Rid = $3;`
	const sql = `Update public.receipt set amount = $1 ,date = $2 , fee = $3 , item = $4 , description = $5 where Rid = $6;`
	//res, err := mdb.Exec(sql, amount, Date, Rid)
	res, err := mdb.Exec(sql, input.Amount, input.Date, input.Fee, input.Item, input.Description, input.Rid)
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

	cm.DeleteCommissionData(input.Rid, dbname, mdb)
	input.ARid = r.ARid
	err = cm.CreateCommission(input, mdb)
	if err != nil {
		return err
	}

	defer mdb.Close()
	return nil
}

//Delete Commission data at same time
func (rm *RTModel) DeleteReceiptData(Rid, dbname string, mdb *sql.DB) (string, error) {
	if mdb == nil {
		db := rm.imr.GetSQLDBwithDbname(dbname)
		mdb, _ = db.ConnectSQLDB()
		defer mdb.Close()
	}

	r := rm.GetReceiptDataByID(mdb, Rid)
	if r == nil || r.Rid == "" {
		return "", errors.New("not found receipt")
	}

	c := cm.GetCommissionDataByRID(mdb, Rid)
	if c.Bsid != "" {
		return "", errors.New("此收款傭金已納入薪資")
	}

	ivData := rm.GetInvoiceDataByRid(mdb, Rid)
	if len(ivData) > 0 {
		msg := ""
		for _, element := range ivData {
			msg += element.InvoiceNo + " "
		}
		return "", errors.New("[ERROR]" + msg + "發票號碼已開立，需先處理")
	}

	_, err := salaryM.CheckValidCloseDate(r.Date, dbname, mdb)
	if err != nil {
		return "", err
	}

	fmt.Println("DeleteReceiptData:")

	const sql = `Delete FROM public.receipt where Rid = '%s';
				 Delete From public.commission where Rid = '%s';
				 `

	res, err := mdb.Exec(fmt.Sprintf(sql, Rid, Rid))
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	id, err := res.RowsAffected()
	if err != nil {
		fmt.Println("PG Affecte Wrong: ", err)
		return "", err
	}

	fmt.Println("RowsAffected: ", id)

	if id <= 0 {
		return "", errors.New("not found receipt")
	}

	return r.CaseName + r.CustomerType, nil
}

func (rm *RTModel) GetReceiptDataByID(sqldb *sql.DB, rid string) *Receipt {

	fmt.Println("GetReceiptDataByID:", rid)
	const qspl = `SELECT R.arid, R.rid, R.date, AR.cno, AR.casename, (Case When AR.type = 'buy' then '買' When AR.type = 'sell' then '賣' else 'unknown' End ) as type , AR.name , R.amount
					FROM public.receipt R
					inner join public.ar AR on AR.arid = R.arid
					left join public.invoice iv on iv.rid = r.rid				
					where r.rid = '%s'`

	//db := rm.imr.GetSQLDBwithDbname(dbname)
	//sqldb, err := db.ConnectSQLDB()
	rows, err := sqldb.Query(fmt.Sprintf(qspl, rid))
	if err != nil {
		fmt.Println("[rows err]:", err)
		return nil
	}

	var rt Receipt
	for rows.Next() {

		fmt.Println("scan start")
		if err := rows.Scan(&rt.ARid, &rt.Rid, &rt.Date, &rt.CNo, &rt.CaseName, &rt.CustomerType, &rt.Name, &rt.Amount); err != nil {
			fmt.Println("err Scan " + err.Error())
		}
		fmt.Println("scan end")
	}
	rt.InvoiceData = rm.GetInvoiceDataByRid(sqldb, rid)
	fmt.Println("GetReceiptDataByID Done")

	return &rt
}

func (rm *RTModel) GetReceiptData(begin, end time.Time, key, dbname string) []*Receipt {
	index := "%" + key + "%"
	begin_str := strconv.Itoa(int(begin.Unix()))
	end_str := strconv.Itoa(int(end.Unix()))
	sql := "SELECT R.rid, R.date, AR.cno, AR.casename, (Case When AR.type = 'buy' then '買' When AR.type = 'sell' then '賣' else 'unknown' End ) as type , AR.name , R.amount , R.Fee, AR.Date,  R.item , R.fee, R.description  " +
		" FROM public.receipt R " +
		" inner join public.ar AR on AR.arid = R.arid	" +
		" where extract(epoch from r.date) >= '" + begin_str + "' and extract(epoch from r.date - '1 month'::interval) <= '" + end_str + "' " +
		" and ar.arid like '" + index + "' OR ar.cno like '" + index + "' OR ar.casename like '" + index + "' OR ar.type like '" + index + "' OR ar.name like '" + index + "' " +
		" order by R.date asc , AR.cno "
	db := rm.imr.GetSQLDBwithDbname(dbname)
	rows, err := db.SQLCommand(sql)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}

	var rtDataList []*Receipt
	for rows.Next() {
		var rt Receipt
		var Item NullString
		var Description NullString

		if err := rows.Scan(&rt.Rid, &rt.Date, &rt.CNo, &rt.CaseName, &rt.CustomerType, &rt.Name, &rt.Amount, &rt.Fee, &rt.CompletionDate, &Item, &rt.Fee, &Description); err != nil {
			fmt.Println("err Scan " + err.Error())
		}
		rt.Item = Item.Value
		rt.Description = Description.Value
		rt.InvoiceData = []*Invoice{}
		rtDataList = append(rtDataList, &rt)
	}
	fmt.Println("rtDataList Done")
	// out, err := json.Marshal(rtDataList)
	// if err != nil {
	// 	fmt.Println("err rtDataList")
	// 	return nil
	// }
	//fmt.Println(string(out))

	const Mapsql = `SELECT rid, branch, invoiceno, buyerid, sellerid, title, date, amount	FROM public.invoice; `
	rows, err = db.SQLCommand(Mapsql)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	for rows.Next() {

		var iv Invoice

		if err := rows.Scan(&iv.Rid, &iv.Branch, &iv.InvoiceNo, &iv.BuyerID, &iv.SellerID, &iv.Title, &iv.Date, &iv.TotalAmount); err != nil {
			fmt.Println("err Scan " + err.Error())
		}

		for _, rt := range rtDataList {
			if rt.Rid == iv.Rid {
				rt.InvoiceData = append(rt.InvoiceData, &iv)
				break
			}
		}

	}

	rm.rtList = rtDataList
	return rm.rtList
}

func (rm *RTModel) GetInvoiceDataByRid(sqldb *sql.DB, rid string) []*Invoice {
	const Mapsql = `SELECT rid, branch, invoiceno, buyerid, sellerid, title, date, amount	FROM public.invoice where rid = $1; `
	rows, err := sqldb.Query(Mapsql, rid)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	InvoiceData := []*Invoice{}
	for rows.Next() {

		var iv Invoice

		if err := rows.Scan(&iv.Rid, &iv.Branch, &iv.InvoiceNo, &iv.BuyerID, &iv.SellerID, &iv.Title, &iv.Date, &iv.TotalAmount); err != nil {
			fmt.Println("err Scan " + err.Error())
		}

		InvoiceData = append(InvoiceData, &iv)

	}
	return InvoiceData
}

/*
func (rm *RTModel) GetReceiptDataByRid(rid, dbname string) *Receipt {

	//if invoiceno is null in Database return ""
	//const qspl = `SELECT rid, date, cno, casename, type, name, amount, COALESCE(NULLIF(invoiceno, null),'') FROM public.receipt;`
	//left join public.invoice I on  I.Rid = R.rid
	//
	fmt.Println("GetReceiptDataByRid:", rid)
	const qspl = `SELECT R.rid, R.date, AR.cno, AR.casename, (Case When AR.type = 'buy' then '買' When AR.type = 'sell' then '賣' else 'unknown' End ) as type , AR.name , R.amount, COALESCE(NULLIF(iv.invoiceno, null),'')
					FROM public.receipt R
					inner join public.ar AR on AR.arid = R.arid
					LEFT join public.invoice iv on iv.rid = R.rid
					where R.rid = '%s' `
	db := rm.imr.GetSQLDBwithDbname(dbname)
	rows, err := db.SQLCommand(fmt.Sprintf(qspl, rid))
	if err != nil {
		return nil
	}

	for rows.Next() {
		rt := &Receipt{}

		// if err := rows.Scan(&r.ARid, &s); err != nil {
		// 	fmt.Println("err Scan " + err.Error())
		// }
		if err := rows.Scan(&rt.Rid, &rt.Date, &rt.CNo, &rt.CaseName, &rt.CustomerType, &rt.Name, &rt.Amount); err != nil {
			fmt.Println("err Scan " + err.Error())
		}

		return rt
	}

	return nil
}
*/
func (rm *RTModel) Json() ([]byte, error) {
	return json.Marshal(rm.rtList)
}

func (rm *RTModel) CreateReceipt(rt *Receipt, dbname string, sqldb *sql.DB, idTime *time.Time) (err error) {
	if sqldb == nil {
		interdb := rm.imr.GetSQLDBwithDbname(dbname)
		sqldb, _ = interdb.ConnectSQLDB()
		defer sqldb.Close()
	}

	_, err = salaryM.CheckValidCloseDate(rt.Date, dbname, sqldb)
	if err != nil {
		fmt.Println("CreateReceipt:", err)
		return err
	}
	fmt.Println("CreateReceipt : arid is ", rt.ARid)
	/*
	*前端時間 會送 yyyy-mm-dd 16:00:00 的UTC時間，方便計算，此地直接 加8小。
	*arid exist
	*(加總歷史收款明細 + 此筆單子) <= 應收款項的收款  to_timestamp($2,'YYYY-MM-DD hh24:mi:ss')
	**/
	// const sql = `INSERT INTO public.receipt (Rid, Date, Amount, ARid, Fee, item , description)
	// 			SELECT * FROM (SELECT $1::varchar(50), $2 , $3::INTEGER , $4::varchar(50) , $5::INTEGER, $6::varchar(50) , $7::varchar(50)  ) AS tmp
	// 			WHERE
	// 				EXISTS ( SELECT arid from public.ar ar WHERE arid = $4 )
	// 			and ( select $3 + COALESCE(SUM(amount),0) FROM public.receipt  where arid = $4 ) <=  (SELECT amount from public.ar ar WHERE arid = $4)
	// 			and ( select $5 + COALESCE(SUM(fee),0) FROM public.receipt  where arid = $4 ) <=  (SELECT COALESCE(SUM(fee),0) from public.deduct WHERE arid = $4	)
	// 			;`
	const sql = `INSERT INTO public.receipt (Rid, Date, Amount, ARid, Fee, item , description)
				SELECT $1::varchar(50), $2 , $3::INTEGER , $4::varchar(50) , $5::INTEGER, $6::varchar(50) , $7::varchar(50)  AS tmp 
				WHERE  
					EXISTS ( SELECT arid from public.ar ar WHERE arid = $4 ) 
				and ( select $3 + COALESCE(SUM(amount),0) FROM public.receipt  where arid = $4 ) <=  (SELECT amount from public.ar ar WHERE arid = $4)
				and ( select $5 + COALESCE(SUM(fee),0) FROM public.receipt  where arid = $4 ) <=  (SELECT COALESCE(SUM(fee),0) from public.deduct WHERE arid = $4	)			
				;`
	//and ( select sum(amount)+$3 FROM public.receipt  where arid = $4 group by arid ) <=  (SELECT amount from public.ar ar WHERE arid = $4);`

	// out, err := json.Marshal(rt)
	// if err != nil {
	// 	panic(err)
	// }
	//fmt.Println(string(out))
	//fmt.Println(string(sql))
	var t int64
	if idTime == nil {
		t = time.Now().Unix()
	} else {
		t = idTime.Unix()
	}

	res, err := sqldb.Exec(sql, t, rt.Date.UTC(), rt.Amount, rt.ARid, rt.Fee, rt.Item, rt.Description)
	//res, err := sqldb.Exec(sql, unix_time, receivable.Date, receivable.CNo, receivable.Sales)
	if err != nil {
		fmt.Println("CreateReceipt:", err)
		return err
	}

	id, err := res.RowsAffected()
	if err != nil {
		fmt.Println("PG Affecte Wrong: ", err)
		return err
	}
	fmt.Println(id)

	if id == 0 {
		return errors.New("Invalid operation, 請確認 [扣款金額、已收金額] 有無超過應收款設定範圍")
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
	err = cm.CreateCommission(rt, sqldb)
	if err != nil {
		return err
	}

	// fmt.Println("UpdateDeductRid [GO]")
	// err = decuctModel.UpdateDeductRid(rt.ARid, sqldb)
	// if err != nil {
	// 	fmt.Println("UpdateDeductRid:" + err.Error())
	// }

	return nil
}
