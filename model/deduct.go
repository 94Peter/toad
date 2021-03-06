package model

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"toad/resource/db"
)

//`json:"id"` 回傳重新命名
type Deduct struct {
	ARid   string    `json:"-"`
	Did    string    `json:"id"`
	Status string    `json:"status"`
	Date   time.Time `json:"date"` // 支付日
	//CostDate    time.Time   `json:"costDate"` // 扣款日
	Fee         int         `json:"fee"`
	Description string      `json:"description"`
	Item        string      `json:"item"`
	ReceiveDate time.Time   `json:"receiveDate"` //成交日
	CNo         string      `json:"contractNo"`
	CaseName    string      `json:"caseName"`
	Type        string      `json:"type"`
	CheckNumber string      `json:"checkNumber"`
	Sales       []*MAPSaler `json:"sales"`
	//ReceiptList []*Receipt  `json:"receiptList"`
}

type DeductCost struct {
	Rid         string    `json:"rid"`
	Date        time.Time `json:"date"`     // 成交日
	CostDate    time.Time `json:"costDate"` // 扣款日
	Fee         int       `json:"fee"`      //扣款價格
	Amount      int       `json:"amount"`   //收款價格
	Description string    `json:"description"`
	Item        string    `json:"item"`
	CNo         string    `json:"contractNo"`
	CaseName    string    `json:"caseName"`
	Type        string    `json:"type"`
}

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
	imr interModelRes
	db  db.InterSQLDB
	//deductList []*Deduct
}

func (decuctModel *DeductModel) GetReceiptFeeOnDeductData(begin, end time.Time, dbname string) []*DeductCost {

	const sql = `SELECT R.rid, AR.date, R.date, AR.cno, AR.casename, (Case When AR.type = 'buy' then '買' When AR.type = 'sell' then '賣' else 'unknown' End ) as type , R.item , R.fee, R.description, R.Amount
					FROM public.receipt R
					inner join public.ar AR on AR.arid = R.arid				
					 where extract(epoch from r.date) >= '%d' and extract(epoch from r.date - '1 month'::interval) <= '%d' and R.Fee > 0
					order by r.date desc , AR.cno `
	db := decuctModel.imr.GetSQLDBwithDbname(dbname)
	rows, err := db.SQLCommand(fmt.Sprintf(sql, begin.Unix(), end.Unix()))
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}

	var DataList []*DeductCost
	for rows.Next() {

		var d DeductCost

		if err := rows.Scan(&d.Rid, &d.Date, &d.CostDate, &d.CNo, &d.CaseName, &d.Type, &d.Item, &d.Fee, &d.Description, &d.Amount); err != nil {
			fmt.Println("err Scan " + err.Error())
		}

		DataList = append(DataList, &d)
	}

	return DataList
}

func (decuctModel *DeductModel) UpdateDeductCostData(dc *DeductCost, dbname string) error {

	db := decuctModel.imr.GetSQLDBwithDbname(dbname)

	mdb, err := db.ConnectSQLDB()
	if err != nil {
		fmt.Println(err)
		return err
	}

	r := rm.GetReceiptDataByID(mdb, dc.Rid)
	if r.Rid == "" {
		return errors.New("not found receipt")
	}

	c := cm.GetCommissionDataByRID(mdb, dc.Rid)
	if c.Bsid != "" {
		return errors.New("此收款傭金已納入薪資")
	}

	_, err = salaryM.CheckValidCloseDate(r.Date, dbname, mdb)
	if err != nil {
		return err
	}

	fmt.Println("UpdateReceiptData")

	const sql = `Update public.receipt set amount = $1 ,date = $2 , fee = $3 , item = $4 , description = $5 where Rid = $6;`
	fmt.Println("UpdateReceiptData:", dc.CostDate)
	res, err := mdb.Exec(sql, dc.Amount, dc.CostDate, dc.Fee, dc.Item, dc.Description, dc.Rid)
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

	cm.DeleteCommissionData(dc.Rid, dbname, mdb)
	r.Amount = dc.Amount
	r.Fee = dc.Fee
	r.Date = dc.Date
	r.Item = dc.Item
	r.Description = dc.Description
	err = cm.CreateCommission(r, mdb)
	if err != nil {
		return err
	}

	defer mdb.Close()
	return nil

}

func (decuctModel *DeductModel) GetDeductData(by_m, ey_m time.Time, mtype, arid, dbname string) []*Deduct {

	// const qspl = `SELECT D.arid, D.Did, D.date , D.status, D.item, D.fee, D.Description, D.checkNumber , AR.date, AR.CNo, AR.CaseName, AR.type FROM public.deduct as D
	// inner join public.ar as AR on AR.arid = D.arid
	// where ( extract(epoch from ar.date) >= '%d' and extract(epoch from ar.date - '1 month'::interval) < '%d' )
	// and (D.item like '%s' OR  D.status like '%s');`
	const sql = `select D.* from (
		SELECT D.arid, D.Did, D.date , D.status, D.item, D.fee, D.Description, D.checkNumber , AR.date, AR.CNo, AR.CaseName, AR.type FROM public.deduct as D 
			inner join public.ar as AR on AR.arid = D.arid
			where ( extract(epoch from ar.date) >= '%d' and extract(epoch from ar.date - '1 month'::interval) < '%d' ) 
			and (D.item like '%s' OR  D.status like '%s')
			and D.arid like '%s'
		) D
		Left JOIN (
		SELECT max(date) date, arid  FROM public.receipt where fee > 0 group by arid
		) r on D.arid = r.arid
		 order by  case when r.arid is null then 1 else 0 end asc, r.date desc ;`

	db := decuctModel.imr.GetSQLDBwithDbname(dbname)
	sqldb, err := db.ConnectSQLDB()
	rows, err := sqldb.Query(fmt.Sprintf(sql, by_m.Unix(), ey_m.Unix(), mtype, mtype, "%"+arid+"%"))
	if err != nil {
		fmt.Println(err)
		return nil
	}
	var deductDataList []*Deduct

	for rows.Next() {
		var d Deduct
		/*null time cannot scan into time.Time */
		var Ddate, RDate NullTime

		if err := rows.Scan(&d.ARid, &d.Did, &Ddate, &d.Status, &d.Item, &d.Fee, &d.Description, &d.CheckNumber, &RDate, &d.CNo, &d.CaseName, &d.Type); err != nil {
			fmt.Println("err Scan " + err.Error())
		}

		d.Date = Ddate.Time
		d.ReceiveDate = RDate.Time
		//d.ReceiptList = []*Receipt{}
		deductDataList = append(deductDataList, &d)

	}

	//找出sales
	const Mapsql = `SELECT did, sid, proportion, sname	FROM public.deductmap; `
	rows, err = sqldb.Query(Mapsql)
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

	//找出receipt
	// const receiptMapsql = `SELECT rid, date, amount, fee, arid FROM public.receipt where fee > 0 order by date desc; `
	// rows, err = sqldb.Query(receiptMapsql)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return nil
	// }

	// for rows.Next() {

	// 	var r Receipt

	// 	if err := rows.Scan(&r.Rid, &r.Date, &r.Amount, &r.Fee, &r.ARid); err != nil {
	// 		fmt.Println("err Scan " + err.Error())
	// 	}

	// 	for _, deduct := range deductDataList {
	// 		if deduct.ARid == r.ARid {
	// 			//無中斷，需重複使用。
	// 			deduct.ReceiptList = append(deduct.ReceiptList, &r)
	// 		}
	// 	}

	// }

	//decuctModel.deductList = deductDataList
	defer sqldb.Close()
	return deductDataList

}

// func (decuctModel *DeductModel) Json() ([]byte, error) {
// 	return json.Marshal(decuctModel.deductList)
// }

func (decuctModel *DeductModel) CreateDeduct(deduct *Deduct, dbname string) (err error) {
	interdb := decuctModel.imr.GetSQLDBwithDbname(dbname)
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}

	_, err = salaryM.CheckValidCloseDate(time.Now(), dbname, sqldb)
	if err != nil {
		return
	}

	/*為了在Deduct Table中 找到對應的收款明細，以便取得收款時間。
	//若先建立應扣款項(所以會找不到應收款項Rid)，Rid就會是null
	*/
	const sql = `INSERT INTO public.deduct (did, arid, item, description, fee, rid)
				WITH  vals  AS (VALUES ($1, $2, $3, $4, $5::integer)) 
				SELECT v.* , r.rid  FROM vals as v
				Left join public.receipt AS r on r.date = (select MIN(Date) FROM public.receipt where arid = $2) and v.column2 = r.arid limit 1;`

	//interdb := decuctModel.imr.GetSQLDB()

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
	//decuctModel.setFeeToCommission(sqldb, deduct.Did, deduct.ARid, dbname)
	defer sqldb.Close()
	return nil
}

//因為傭金明細只需要一筆有應扣費用，hard code更新。
//如果things有帶入，功能回傳arid用。
func (decuctModel *DeductModel) setFeeToCommission_backup(sqldb *sql.DB, Did, ARid string, things ...interface{}) (string, error) {
	fmt.Println("setFeeToCommission")
	if ARid == "" {
		const qspl = `SELECT D.Did, D.arid FROM public.deduct as D Where D.Did = '%s';`
		//db := decuctModel.imr.GetSQLDBwithDbname(dbname)

		rows, err := sqldb.Query(fmt.Sprintf(qspl, Did))
		//rows, err := db.SQLCommand(fmt.Sprintf(qspl, Did))
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

	// sql := `Update public.commission set fee = COALESCE((select sum(fee) from public.deduct where arid = $1),0)
	// 		where rid = (SELECT min(rid) FROM public.commission c where arid = $1) and bsid is null`
	sql := `UPDATE public.commission c 
			SET fee = subquery.fee
			FROM (
				select map.sid, map.sname, COALESCE(sum(d.fee * map.proportion / 100 ),0) fee from public.deductmap map
			LEFT JOIN public.deduct d on d.did = map.did
			where d.arid  = $1
			group by map.sid, map.sname
			) AS subquery
			where c.rid = (SELECT min(rid) FROM public.commission  where arid = $1) and c.bsid is null and subquery.sid  = c.sid`

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
	set sr = t2.amount * t2.cpercent / 100 - t2.fee, bonus = (t2.amount * t2.cpercent / 100 - t2.fee) * t2.percent /100
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

func (decuctModel *DeductModel) DeleteDeduct(ID, dbname string) (err error) {

	interdb := decuctModel.imr.GetSQLDBwithDbname(dbname)
	sqldb, err := interdb.ConnectSQLDB()

	d := decuctModel.getDeductByID(ID, dbname, sqldb)
	if d.Did == "" {
		return errors.New("not found decuct")
	}
	_, err = salaryM.CheckValidCloseDate(d.Date, dbname, sqldb)
	if err != nil {
		return
	}

	/*
	* 邏輯是先取得ARID後刪除，然後從算傭金應扣。
	 */

	//nothing 回傳arid用
	//arid, _ := decuctModel.setFeeToCommission(sqldb, ID, "", "nothing")

	const sql = `DELETE FROM public.deduct WHERE Did=$1 and  status != '已支付'
	and ( (select COALESCE(SUM(fee),0) FROM public.receipt where arid = $2) <=  (select COALESCE(SUM(fee),0) - $3 FROM public.deduct where arid = $2) ) 
	;`

	res, err := sqldb.Exec(sql, ID, d.ARid, d.Fee)
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
		return errors.New("Invalid operation, 請確認未支付狀態 或 收款金額的已扣款項")
	}

	//decuctModel.setFeeToCommission(sqldb, "", arid)
	defer sqldb.Close()
	return nil
}

func (decuctModel *DeductModel) UpdateDeduct(Did, status, date, checkNumber, dbname string) (err error) {

	interdb := decuctModel.imr.GetSQLDBwithDbname(dbname)
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}

	d := decuctModel.getDeductByID(Did, dbname, sqldb)
	if d.Did == "" {
		return errors.New("not found decuct")
	}
	_, err = salaryM.CheckValidCloseDate(d.Date, dbname, sqldb)
	if err != nil {
		return
	}

	sql := fmt.Sprintf("UPDATE public.deduct Set status = $1, date = $2 , checkNumber = $3 Where did = $4")

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
	//decuctModel.setFeeToCommission(sqldb, Did, "")
	defer sqldb.Close()
	return nil
}

func (decuctModel *DeductModel) UpdateDeductFee(Did, dbname string, fee int) (err error) {

	interdb := decuctModel.imr.GetSQLDBwithDbname(dbname)
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}

	d := decuctModel.getDeductByID(Did, dbname, sqldb)
	if d.Did == "" {
		return errors.New("not found decuct")
	}
	_, err = salaryM.CheckValidCloseDate(d.Date, dbname, sqldb)
	if err != nil {
		return
	}

	const sql = `UPDATE public.deduct Set fee = $1 Where did = $2 and status = '未支付'  
	and ( (select COALESCE(SUM(fee),0) FROM public.receipt where arid = $3) <=  (select COALESCE(SUM(fee),0) - $4 + $1 FROM public.deduct where arid = $3) ) ;`

	res, err := sqldb.Exec(sql, fee, Did, d.ARid, d.Fee)
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
		return errors.New("Invalid operation, 請確認未支付狀態 或 收款金額的已扣款項")
	}
	//decuctModel.setFeeToCommission(sqldb, Did, "")
	defer sqldb.Close()
	return nil
}

func (decuctModel *DeductModel) UpdateDeductItem(Did, item, dbname string) (err error) {

	interdb := decuctModel.imr.GetSQLDBwithDbname(dbname)
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}

	d := decuctModel.getDeductByID(Did, dbname, sqldb)
	if d.Did == "" {
		return errors.New(" not found decuct")
	}
	_, err = salaryM.CheckValidCloseDate(d.Date, dbname, sqldb)
	if err != nil {
		return
	}

	sql := fmt.Sprintf("UPDATE public.deduct Set item = $1 Where did = $2")

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
	defer sqldb.Close()
	return nil
}

func (decuctModel *DeductModel) UpdateDeductRid(ARid string, sqldb *sql.DB) (err error) {
	const sql = `Update public.deduct 
				set rid = tmp.ReceiptID
				FROM (
					select D.arid as temID , R.rid as ReceiptID FROM public.deduct D , public.receipt R
					where D.arid = R.arid and R.date = (select MIN(Date) FROM public.receipt where D.arid = $1)
				) as tmp where arid = tmp.temID`

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

func (decuctModel *DeductModel) UpdateDeductSales(Did, dbname string, salerList []*MAPSaler) (err error) {

	interdb := decuctModel.imr.GetSQLDBwithDbname(dbname)
	sqldb, err := interdb.ConnectSQLDB()

	d := decuctModel.getDeductByID(Did, dbname, sqldb)
	if d.Did == "" {
		return errors.New("not found decuct")
	}
	_, err = salaryM.CheckValidCloseDate(d.Date, dbname, sqldb)
	if err != nil {
		return
	}

	fmt.Println("UpdateDeductSales")
	// interdb := decuctModel.imr.GetSQLDBwithDbname(dbname)
	// sqldb, err := interdb.ConnectSQLDB()

	decuctModel.DeleteDeductMAP(Did, sqldb)
	//decuctModel.SaveDeductMAP(salerList, Did, sqldb)

	defer sqldb.Close()
	return nil
}

// func (decuctModel *DeductModel) SaveDeductMAP(salerList []*MAPSaler, ID string, sqldb *sql.DB) {
// 	const mapSql = `INSERT INTO public.DeductMAP(
// 		Did, Sid, proportion , SName )
// 		VALUES ($1, $2, $3, $4);`
// 	count := 0
// 	for _, element := range salerList {
// 		// element is the element from someSlice for where we are
// 		res, err := sqldb.Exec(mapSql, ID, element.Sid, element.Percent, element.SName)
// 		//res, err := sqldb.Exec(sql, unix_time, receivable.Date, receivable.CNo, receivable.Sales)
// 		if err != nil {
// 			fmt.Println("SaveDeductMAP ", err)
// 		}
// 		id, err := res.RowsAffected()
// 		if err != nil {
// 			fmt.Println("PG Affecte Wrong: ", err)
// 		}
// 		count += int(id)
// 	}
// 	fmt.Println("SaveDeductMAP:", count)
// 	//連動更新傭金
// 	if count > 0 {
// 		arid, _ := decuctModel.setFeeToCommission(sqldb, ID, "", "nothing")
// 		decuctModel.setFeeToCommission(sqldb, "", arid)
// 	}
// }

func (decuctModel *DeductModel) DeleteDeductMAP(Did string, sqldb *sql.DB) {

	const sql = `delete from public.deductmap where did = $1`
	sqldb.Exec(sql, Did)
}

func getNil(msg string) (response *string) {
	if msg == "" {
		return nil
	} else {
		return &msg
	}
}

func (decuctModel *DeductModel) getDeductByID(ID, dbname string, sqldb *sql.DB) *Deduct {

	const sql = `SELECT did, Date, arid ,fee  FROM public.deduct where Did = '%s';`
	//where (Date >= '%s' and Date < ('%s'::date + '1 month'::interval))
	// db := prepayM.imr.GetSQLDB()
	// sqldb, err := db.ConnectSQLDB()

	if sqldb == nil {
		fmt.Println("getDeductByID")
		sqldb, _ = decuctModel.imr.GetSQLDBwithDbname(dbname).ConnectSQLDB()
		defer sqldb.Close()

	}

	rows, err := sqldb.Query(fmt.Sprintf(sql, ID))
	if err != nil {
		fmt.Println(err)
		return nil
	}

	deduct := &Deduct{}

	for rows.Next() {
		var lasttime NullTime
		if err := rows.Scan(&deduct.Did, &lasttime, &deduct.ARid, &deduct.Fee); err != nil {
			fmt.Println("err Scan " + err.Error())
			return nil
		}
		deduct.Date = lasttime.Time
	}
	if deduct.Date.IsZero() {
		fmt.Println("deduct", deduct)
	}

	return deduct
}
