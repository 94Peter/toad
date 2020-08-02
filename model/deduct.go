package model

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"toad/resource/db"
)

//`json:"id"` 回傳重新命名
type Deduct struct {
	ARid        string      `json:"-"`
	Did         string      `json:"id"`
	Status      string      `json:"status"`
	Date        time.Time   `json:"date"`
	Fee         int         `json:"fee"`
	Description string      `json:"description"`
	Item        string      `json:"item"`
	ReceiveDate time.Time   `json:"receiveDate"`
	CNo         string      `json:"contractNo"`
	CaseName    string      `json:"caseName"`
	Type        string      `json:"type"`
	CheckNumber string      `json:"checkNumber"`
	Sales       []*MAPSaler `json:"sales"`
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

func (decuctModel *DeductModel) GetDeductData(by_m, ey_m time.Time, mtype string) []*Deduct {

	//
	// const qspl = `SELECT D.Did, D.date , D.status, D.item, D.fee, D.Description, D.checkNumber , R.date, AR.CNo, AR.CaseName, AR.type FROM public.deduct as D
	// 			inner join public.ar as AR on AR.arid = D.arid
	// 			Left join public.receipt as R on R.rid = D.rid
	// 			where ( to_timestamp(date_part('epoch',R.date)::int) >= '%s' and to_timestamp(date_part('epoch',R.date)::int) < ('%s'::date + '1 month'::interval) or R.date is null )
	// 			and (D.item like '%s' OR  D.status like '%s');`

	const qspl = `SELECT D.Did, D.date , D.status, D.item, D.fee, D.Description, D.checkNumber , R.date, AR.CNo, AR.CaseName, AR.type FROM public.deduct as D 
	inner join public.ar as AR on AR.arid = D.arid
	Left join public.receipt as R on R.rid = D.rid
	where ( extract(epoch from r.date) >= '%d' and extract(epoch from r.date - '1 month'::interval) < '%d' or R.date is null ) 
	and (D.item like '%s' OR  D.status like '%s');`

	//where D.rid = R.rid;`
	//const qspl = `SELECT arid,sales	FROM public.ar;`
	//fmt.Println(fmt.Sprintf(qspl, by_m+"-01", ey_m+"-01", mtype, mtype))
	db := decuctModel.imr.GetSQLDB()
	rows, err := db.SQLCommand(fmt.Sprintf(qspl, by_m.Unix(), ey_m.Unix(), mtype, mtype))
	if err != nil {
		fmt.Println(err)
		return nil
	}
	var deductDataList []*Deduct

	for rows.Next() {
		var d Deduct
		/*null time cannot scan into time.Time */
		var Ddate, RDate NullTime

		if err := rows.Scan(&d.Did, &Ddate, &d.Status, &d.Item, &d.Fee, &d.Description, &d.CheckNumber, &RDate, &d.CNo, &d.CaseName, &d.Type); err != nil {
			fmt.Println("err Scan " + err.Error())
		}

		d.Date = Ddate.Time
		d.ReceiveDate = RDate.Time

		deductDataList = append(deductDataList, &d)
	}

	//找出sales
	const Mapsql = `SELECT did, sid, proportion, sname	FROM public.deductmap; `
	rows, err = db.SQLCommand(Mapsql)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	for rows.Next() {
		var did string
		var saler MAPSaler

		if err := rows.Scan(&did, &saler.Sid, &saler.Percent, &saler.SName); err != nil {
			fmt.Println("err Scan " + err.Error())
		}

		for _, deduct := range deductDataList {
			if deduct.Did == did {
				deduct.Sales = append(deduct.Sales, &saler)
				//fmt.Println(arid)
				break
			}
		}

	}

	decuctModel.deductList = deductDataList
	return decuctModel.deductList

}

func (decuctModel *DeductModel) Json() ([]byte, error) {
	return json.Marshal(decuctModel.deductList)
}

func (decuctModel *DeductModel) CreateDeduct(deduct *Deduct) (err error) {
	fmt.Println("arid:", deduct.ARid)

	out, err := json.Marshal(deduct)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(out))

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

	//應收扣款帳款成功才建立業務對應表，對應表要從ARMAP自動找出(預設配對正常)
	if id > 0 {
		const mapSql = `INSERT INTO public.DEDUCTMAP(Did, Sid, proportion , SName )
		(select $1, Sid, proportion , SName from public.armap where arid = $2);`

		res, err := sqldb.Exec(mapSql, fakeid, deduct.ARid)
		if err != nil {
			fmt.Println("DEDUCT MAP ", err)
			return err
		}
		id, err := res.RowsAffected()
		if err != nil {
			fmt.Println("PG Affecte Wrong: ", err)
			return err
		}
		fmt.Println(id)
	}

	deduct.Did = fmt.Sprintf("%d", fakeid)
	decuctModel.setFeeToCommission(sqldb, deduct.Did, deduct.ARid)

	return nil
}

//因為傭金明細只需要一筆有應扣費用，hard code更新。
func (decuctModel *DeductModel) setFeeToCommission(sqldb *sql.DB, Did, ARid string, things ...interface{}) (string, error) {

	if ARid == "" {
		const qspl = `SELECT D.Did, D.arid FROM public.deduct as D Where D.Did = '%s';`
		db := decuctModel.imr.GetSQLDB()
		rows, err := db.SQLCommand(fmt.Sprintf(qspl, Did))
		if err != nil {
			fmt.Println(err)
			return "", nil
		}
		d := &Deduct{}
		for rows.Next() {
			/*null time cannot scan into time.Time */
			if err := rows.Scan(&d.Did, &d.ARid); err != nil {
				fmt.Println("err Scan " + err.Error())
			}
		}
		ARid = d.ARid

		for _, params := range things {
			fmt.Println(params, ":", ARid)
			return ARid, nil
		}
	}

	sql := `Update public.commission set fee = COALESCE((select sum(fee) from public.deduct where arid = $1),0)
			where rid = (SELECT min(rid) FROM public.commission c where arid = $1) and bsid is null`
	// if mtype == "date" {
	// 	sql = fmt.Sprintf("UPDATE public.deduct Set %s = to_timestamp($1 ,'YYYY-MM-DD hh24:mi:ss') Where did = $2", mtype)
	// }

	// if mtype == "status" {
	// 	sql = fmt.Sprintf("UPDATE public.deduct Set %s = $1 Where did = $2", mtype)
	// }

	// interdb := decuctModel.imr.GetSQLDB()
	// sqldb, err := interdb.ConnectSQLDB()
	// if err != nil {
	// 	return "", err
	// }

	res, err := sqldb.Exec(sql, ARid)
	//res, err := sqldb.Exec(sql, unix_time, receivable.Date, receivable.CNo, receivable.Sales)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	id, err := res.RowsAffected()
	if err != nil {
		fmt.Println("PG Affecte Wrong: ", err)
		return "", err
	}
	fmt.Println(id)
	if id >= 1 {
		fmt.Println("setFeeToCommission success:arid", ARid)
		decuctModel.updateSRBonusToCommission(ARid, sqldb)
	} else if id == 0 {
		fmt.Println("setFeeToCommission fail : maybe rid not exist")
	}
	return "", nil
}

func (decuctModel *DeductModel) updateSRBonusToCommission(ARid string, sqldb *sql.DB) (string, error) {

	sql := `Update public.commission t1
	set sr = (t2.amount - t2.fee) * t2.cpercent / 100 , bonus = (t2.amount - t2.fee) * t2.cpercent / 100 * t2.percent /100
	FROM(
	SELECT c.sid, c.rid, r.date, c.item, r.amount, c.fee , c.sname, c.cpercent, c.sr, c.bonus, r.arid, c.status , cs.percent
					FROM public.commission c
					inner JOIN public.receipt r on r.rid = c.rid				
					inner join 	(			
						select cs.branch, cs.sid, cs.percent, cs.identitynum from public.configsaler cs 
						inner join (
							select sid, max(zerodate) zerodate from public.configsaler cs 
							where now() > zerodate
							group by sid
						) tmp on tmp.sid = cs.sid and tmp.zerodate = cs.zerodate		
					)	cs  on cs.sid = c.sid 
					where c.arid = $1 
	) as t2 where t1.sid = t2.sid and t1.rid = t2.rid and t1.bsid is null
	;`
	//fee不知道能不能為0 呵呵
	//t1.bsid is null 已經綁定薪資表的不給改
	//where c.arid = $1 and fee != 0

	// if mtype == "date" {
	// 	sql = fmt.Sprintf("UPDATE public.deduct Set %s = to_timestamp($1 ,'YYYY-MM-DD hh24:mi:ss') Where did = $2", mtype)
	// }

	// if mtype == "status" {
	// 	sql = fmt.Sprintf("UPDATE public.deduct Set %s = $1 Where did = $2", mtype)
	// }

	// interdb := decuctModel.imr.GetSQLDB()
	// sqldb, err := interdb.ConnectSQLDB()
	// if err != nil {
	// 	return "", err
	// }

	res, err := sqldb.Exec(sql, ARid)
	//res, err := sqldb.Exec(sql, unix_time, receivable.Date, receivable.CNo, receivable.Sales)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	id, err := res.RowsAffected()
	if err != nil {
		fmt.Println("PG Affecte Wrong: ", err)
		return "", err
	}
	fmt.Println(id)
	if id >= 1 {
		fmt.Println("updateSRBonusToCommission success")
	} else if id == 0 {
		fmt.Println("updateSRBonusToCommission fail : maybe rid not exist")
	}
	return "", nil

}

func (decuctModel *DeductModel) DeleteDeduct(ID string) (err error) {

	interdb := decuctModel.imr.GetSQLDB()
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}

	//nothing 回傳arid用
	arid, _ := decuctModel.setFeeToCommission(sqldb, ID, "", "nothing")

	const sql = `DELETE FROM public.deduct WHERE Did=$1 and  status != '已支付';`

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

	decuctModel.setFeeToCommission(sqldb, "", arid)
	return nil
}

func (decuctModel *DeductModel) UpdateDeduct(Did, status, date, checkNumber string) (err error) {

	sql := fmt.Sprintf("UPDATE public.deduct Set status = $1, date = $2 , checkNumber = $3 Where did = $4")

	interdb := decuctModel.imr.GetSQLDB()
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}

	res, err := sqldb.Exec(sql, status, getNil(date), checkNumber, Did)
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
	decuctModel.setFeeToCommission(sqldb, Did, "")
	return nil
}

func (decuctModel *DeductModel) UpdateDeductFee(Did string, fee int) (err error) {

	sql := fmt.Sprintf("UPDATE public.deduct Set fee = $1 Where did = $2 and status = '未支付'")

	interdb := decuctModel.imr.GetSQLDB()
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}

	res, err := sqldb.Exec(sql, fee, Did)
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
		return errors.New("Invalid operation, UpdateDeductFee")
	}
	decuctModel.setFeeToCommission(sqldb, Did, "")
	return nil
}

func (decuctModel *DeductModel) UpdateDeductItem(Did, item string) (err error) {

	sql := fmt.Sprintf("UPDATE public.deduct Set item = $1 Where did = $2")

	interdb := decuctModel.imr.GetSQLDB()
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}

	res, err := sqldb.Exec(sql, item, Did)
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

	if id == 0 {
		return errors.New("Invalid operation, UpdateDeductItem")
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

func (decuctModel *DeductModel) UpdateDeductSales(Did string, salerList []*MAPSaler) (err error) {

	fmt.Println("UpdateDeductSales")
	interdb := decuctModel.imr.GetSQLDB()
	sqldb, err := interdb.ConnectSQLDB()
	decuctModel.DeleteDeductMAP(Did)
	decuctModel.SaveDeductMAP(salerList, Did, sqldb)

	return nil
}

func (decuctModel *DeductModel) SaveDeductMAP(salerList []*MAPSaler, ID string, sqldb *sql.DB) {
	const mapSql = `INSERT INTO public.DeductMAP(
		Did, Sid, proportion , SName )
		VALUES ($1, $2, $3, $4);`
	count := 0
	for _, element := range salerList {
		// element is the element from someSlice for where we are
		res, err := sqldb.Exec(mapSql, ID, element.Sid, element.Percent, element.SName)
		//res, err := sqldb.Exec(sql, unix_time, receivable.Date, receivable.CNo, receivable.Sales)
		if err != nil {
			fmt.Println("SaveDeductMAP ", err)
		}
		id, err := res.RowsAffected()
		if err != nil {
			fmt.Println("PG Affecte Wrong: ", err)
		}
		count += int(id)
	}
	fmt.Println("SaveDeductMAP:", count)
}

func (decuctModel *DeductModel) DeleteDeductMAP(Did string) {

	interdb := decuctModel.imr.GetSQLDB()
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		fmt.Println(err)
	}
	const sql = `delete from public.deductmap where did = $1`
	_, err = sqldb.Exec(sql, Did)
}

func getNil(msg string) (response *string) {
	if msg == "" {
		return nil
	} else {
		return &msg
	}
}
