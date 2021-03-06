package model

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"toad/resource/db"
)

var ACTION_BUY = "buy"
var ACTION_SELL = "sell"

//`json:"id"` 回傳重新命名
type AR struct {
	ARid         string      `json:"id"`
	Date         time.Time   `json:"completionDate"`
	CNo          string      `json:"contractNo"`
	Customer     Customer    `json:"customer"`
	CaseName     string      `json:"caseName"`
	Amount       int         `json:"amount"`
	Fee          int         `json:"fee"`            //應扣費用
	Cost         int         `json:"cost"`           //已扣費用
	Balance      int         `json:"balance"`        //未收金額
	RA           int         `json:"receivedAmount"` //已收金額
	ReturnAmount int         `json:"returnAmount"`   //折讓金額
	Sales        []*MAPSaler `json:"sales"`
	//DeductList []*Deduct   `json:"deductList"`
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

type Customer struct {
	Action string `json:"type"`
	Name   string `json:"name"`
}

type Saler struct {
	SName   string  `json:"name"`
	Percent float64 `json:"proportion"` //{"{\"BName\":\"123\",\"Bid\":\"13\",\"Persent\":12}","{\"BName\":\"123\",\"Bid\":\"13\",\"Persent\":12}"}
	Sid     string  `json:"account"`
	Branch  string  `json:"branch"`
	Title   string  `json:"title"`
}

type MAPSaler struct {
	SName        string  `json:"name"`
	Percent      float64 `json:"proportion"` //{"{\"BName\":\"123\",\"Bid\":\"13\",\"Persent\":12}","{\"BName\":\"123\",\"Bid\":\"13\",\"Persent\":12}"}
	Sid          string  `json:"account"`
	Branch       string  `json:"branch"`
	BonusPercent float64 `json:"percent"`
}

type HouseGoMAPSaler struct {
	SName   string  `json:"name"`
	Percent float64 `json:"proportion"` //{"{\"BName\":\"123\",\"Bid\":\"13\",\"Persent\":12}","{\"BName\":\"123\",\"Bid\":\"13\",\"Persent\":12}"}
	Sid     string  `json:"id"`
}

type HouseGo struct {
	ARid string `json:"arid"`
	ID   string `json:"id"`
	Data string `json:"data"`
}

type AccountReceivable struct {
	db db.InterSQLDB
	//res interModelRes
	ar []*AR
}

type mapRidBsid struct {
	rid        string
	bsid       string
	branch     string
	salaryName string
}

type inputhouseGoAR struct {
	ARid     int       `json:"id"`
	Date     time.Time `json:"completionDate"` //成交日期
	CNo      string    `json:"contractNo"`
	CaseName string    `json:"caseName"`

	Sell struct {
		Amount int    `json:"amount"`
		Name   string `json:"name"`
	} `json:"sell"`

	Buyer struct {
		Amount int    `json:"amount"`
		Name   string `json:"name"`
	} `json:"buyer"`
	Sales []*HouseGoMAPSaler `json:"sales"`
}

var (
	am *ARModel
)

func GetARModel(imr interModelRes) *ARModel {
	if am != nil {
		return am
	}

	am = &ARModel{
		imr: imr,
	}
	return am
}

type ARModel struct {
	imr interModelRes
	db  db.InterSQLDB
	//arList    []*AR
	salerList []*Saler
	hgList    []*HouseGo
}

func (am *ARModel) GetSalerData(branch, dbname string) []*Saler {

	const qspl = `SELECT A.sid, A.sname, A.branch, A.percent, A.title
					FROM public.ConfigSaler A 
					Inner Join ( 
						select sid, max(zerodate) zerodate from public.configsaler cs 
						where now() > zerodate
						group by sid 
					) B on A.sid=B.sid and A.zeroDate = B.zeroDate
					where A.branch like '%s';`
	//const qspl = `SELECT arid,sales	FROM public.ar;`
	db := am.imr.GetSQLDBwithDbname(dbname)
	//fmt.Println(fmt.Sprintf(qspl, branch))
	rows, err := db.SQLCommand(fmt.Sprintf(qspl, branch))
	if err != nil {
		fmt.Println(err)
		return nil
	}
	var salerList []*Saler

	for rows.Next() {
		var saler Saler

		// if err := rows.Scan(&r.ARid, &s); err != nil {
		// 	fmt.Println("err Scan " + err.Error())
		// }
		if err := rows.Scan(&saler.Sid, &saler.SName, &saler.Branch, &saler.Percent, &saler.Title); err != nil {
			fmt.Println("err Scan " + err.Error())
		}

		salerList = append(salerList, &saler)
	}

	// out, err := json.Marshal(arList)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(string(out))
	am.salerList = salerList
	return am.salerList
}

func (am *ARModel) GetARData(key, status, branch, dbname string, date time.Time) []*AR {
	begin_str := strconv.Itoa(int(date.Unix()))
	index := "%" + key + "%"
	sql := "SELECT ar.arid, ar.date, ar.cno, ar.casename, ar.type, ar.name, ar.amount, " +
		"	COALESCE((SELECT SUM(d.fee) FROM public.deduct d WHERE ar.arid = d.arid),0) AS SUM_Fee," +
		"	COALESCE((SELECT SUM(r.amount) FROM public.receipt r WHERE ar.arid = r.arid),0) AS SUM_RA," +
		"   COALESCE((SELECT SUM(re.amount) FROM public.returns re WHERE ar.arid = re.arid),0) AS SUM_Return," +
		"   COALESCE((SELECT SUM(r.fee) FROM public.receipt r WHERE ar.arid = r.arid),0) AS SUM_RFee " +
		"FROM public.ar ar	" +
		"where (ar.arid like '" + index + "' OR ar.cno like '" + index + "' OR ar.casename like '" + index + "' OR ar.type like '" + index + "' OR ar.name like '" + index + "')" +
		" and extract(epoch from ar.date) >= '" + begin_str + "' " +
		"group by ar.arid order by ar.date desc , ar.cno;"
	/*
	*balance equal ar.amount - COALESCE((SELECT SUM(r.amount) FROM public.receipt r WHERE ar.arid = r.arid),0) AS SUM_RA
	*but I do with r.Balance = r.Amount - r.RA
	 */

	db := am.imr.GetSQLDBwithDbname(dbname)
	//fmt.Println(sql)
	//rows, err := db.SQLCommand(fmt.Sprintf(sql))

	rows, err := db.SQLCommand(sql)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	var arDataList []*AR
	var Final_arDataList []*AR
	for rows.Next() {
		var r AR

		var ctm Customer
		//var col_sales string
		// if err := rows.Scan(&r.ARid, &s); err != nil {
		// 	fmt.Println("err Scan " + err.Error())
		// }
		if err := rows.Scan(&r.ARid, &r.Date, &r.CNo, &r.CaseName, &ctm.Action, &ctm.Name, &r.Amount, &r.Fee, &r.RA, &r.ReturnAmount, &r.Cost); err != nil {
			fmt.Println("err Scan " + err.Error())
		}
		r.Customer = ctm
		r.Balance = r.Amount - r.RA //未收 = 應收 - 已收
		r.Sales = []*MAPSaler{}
		//r.DeductList = []*Deduct{}

		if status == "0" { //全部
			arDataList = append(arDataList, &r)
		} else if status == "1" && r.Balance == 0 { //已完款
			fmt.Println("status=1,", r)
			arDataList = append(arDataList, &r)
		} else if status == "2" && r.Balance > 0 { //未完款
			fmt.Println("status=2,", r)
			arDataList = append(arDataList, &r)
		}

	}

	Mapsql := "SELECT arid, sid, proportion, sname, branch, percent " +
		"		FROM public.armap " +
		" where arid in ( " +
		" SELECT arid from public.armap where branch like '%" + branch + "%' " +
		") ;"
	rows, err = db.SQLCommand(Mapsql)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	for rows.Next() {
		var arid string
		var saler MAPSaler

		if err := rows.Scan(&arid, &saler.Sid, &saler.Percent, &saler.SName, &saler.Branch, &saler.BonusPercent); err != nil {
			fmt.Println("err Scan " + err.Error())
		}

		for _, ar := range arDataList {
			if ar.ARid == arid {
				ar.Sales = append(ar.Sales, &saler)
				//fmt.Println(arid)
				break
			}
		}
	}

	if branch == "" {
		return arDataList
	} else {
		for _, ar := range arDataList {
			for _, saler := range ar.Sales {
				if saler.Branch == branch {
					Final_arDataList = append(Final_arDataList, ar)
					break
				}
			}
		}
	}

	// by_m := "1980-01-01T00:00:00.000Z"
	// ey_m := "2200-12-31T00:00:00.000Z"
	// b, _ := time.Parse(time.RFC3339, by_m)
	// e, _ := time.Parse(time.RFC3339, ey_m)
	// deductList := decuctModel.GetDeductData(b, e, "%", "", dbname)
	// for _, deduct := range deductList {
	// 	for _, ar := range arDataList {
	// 		if ar.ARid == deduct.ARid {
	// 			ar.DeductList = append(ar.DeductList, deduct)
	// 			break
	// 		}
	// 	}
	// }

	// out, err := json.Marshal(arList)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(string(out))
	//am.arList = arDataList

	return Final_arDataList

}

func (am *ARModel) GetHouseGoData(today, end time.Time, key, dbname string) []*HouseGo {

	//index := "%" + key + "%"
	sql := "SELECT arid, id, data FROM public.housego"

	/*
	*balance equal ar.amount - COALESCE((SELECT SUM(r.amount) FROM public.receipt r WHERE ar.arid = r.arid),0) AS SUM_RA
	*but I do with r.Balance = r.Amount - r.RA
	 */

	//const qspl = `SELECT arid,sales	FROM public.ar;`
	db := am.imr.GetSQLDBwithDbname(dbname)
	//fmt.Println(sql)
	//rows, err := db.SQLCommand(fmt.Sprintf(sql))

	rows, err := db.SQLCommand(sql)
	if err != nil {
		fmt.Println("GetHouseGoData:", err)
		return nil
	}
	var hgList []*HouseGo

	for rows.Next() {
		var hg HouseGo

		//var col_sales string
		// if err := rows.Scan(&r.ARid, &s); err != nil {
		// 	fmt.Println("err Scan " + err.Error())
		// }
		if err := rows.Scan(&hg.ARid, &hg.ID, &hg.Data); err != nil {
			fmt.Println("err Scan " + err.Error())
		}

		hgList = append(hgList, &hg)
	}

	am.hgList = hgList
	return am.hgList

}

func (am *ARModel) GetSalerDataByID(sqldb *sql.DB, id string) *Saler {

	sql := "SELECT branch, percent, sid FROM public.configsaler where sid = $1"

	rows, err := sqldb.Query(sql, id)
	if err != nil {
		fmt.Println("am GetSalerDataByID:", err)
		return nil
	}
	var data *Saler
	data = nil
	for rows.Next() {
		var s Saler
		//var col_sales string
		// if err := rows.Scan(&r.ARid, &s); err != nil {
		// 	fmt.Println("err Scan " + err.Error())
		// }
		if err := rows.Scan(&s.Branch, &s.Percent, &s.Sid); err != nil {
			fmt.Println("err Scan " + err.Error())
			return nil
		}
		data = &s
	}

	return data

}

func (am *ARModel) Json(mtype string) ([]byte, error) {
	switch mtype {
	// case "ar":
	// 	return json.Marshal(am.arList)
	case "saler":
		return json.Marshal(am.salerList)
	case "housego":
		return json.Marshal(am.hgList)
	default:
		fmt.Println("unknown config type")
		break
	}
	return nil, nil
}

func (am *ARModel) UpdateAccountReceivable(amount int, ID, dbname string, salerList []*MAPSaler) (err error) {

	fmt.Println("UpdateAccountReceivable")
	const sql = `Update public.ar t1
					set	amount = $1
				FROM (
					SELECT ar.arid, ar.amount, coalesce(sum(r.amount),0) sunreceipt
					FROM public.ar ar
					LEFT JOIN public.receipt r ON ar.arid = r.arid
					where ar.arid = $2
					group by ar.arid				
				)as t2 
				where t1.arid = $2  and  sunreceipt <= $1 ;`

	interdb := am.imr.GetSQLDBwithDbname(dbname)
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}

	dataF := am.checkEditable(ID, sqldb)
	errmsg := ""
	old_bsid := ""
	for i := 0; i < len(dataF); i++ {
		element := dataF[i]

		if element.bsid != "" {
			if i == 0 {
				errmsg += element.branch + element.salaryName + ". "
				old_bsid = element.bsid
			}
			//去除重複店家判斷
			if i+1 < len(dataF) {
				if old_bsid == dataF[i+1].bsid {
					//skip
					i++
				} else {
					errmsg += element.branch + element.salaryName + ". "
					old_bsid = element.bsid
				}
			}

		}

	}
	fmt.Println(dataF)
	if errmsg != "" {
		return errors.New("[ERROR]:" + errmsg + " 已產生此相關薪資")
	}

	//fmt.Println("sqldb Exec " + sql)
	res, err := sqldb.Exec(sql, amount, ID)
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
	if id <= 0 {
		return errors.New("[ERROR]: Maybe id is not found or amount is not allowed. (amount should be greater then sum of receive amount)")
	}
	//刪除ARMAP 重建
	am.DeleteARandDeductMAP(ID, sqldb)
	am.SaveARMAP(salerList, ID, sqldb)
	am.SaveDeductMAP(ID, sqldb)
	//delete old receipt commission && rebuild receipt!
	doneMsg := ""
	for i := 0; i < len(dataF); i++ {
		element := dataF[i]
		if element.bsid == "" {
			receipt := rm.GetReceiptDataByID(sqldb, element.rid)
			if receipt.Rid != "" {
				msg, err := rm.DeleteReceiptData(element.rid, dbname, sqldb) //bsid無綁定狀況，刪除commission。 // 後續爭議款可能還要改成自動又開發票
				doneMsg += msg
				if err != nil { //not found receipt是正常的，會取出相同的rid(但可能不同人) 但都可以進行刪除
					// if doneMsg != "" {
					// 	return errors.New("[ERROR]:" + doneMsg + " 已刪除收款," + err.Error())
					// } else {
					// 	return err
					// }
				}
				err = rm.CreateReceipt(receipt, dbname, sqldb, nil)
				if err != nil {
					fmt.Println("CreateReceipt on update ar error:", receipt.Rid)
					return errors.New("[ERROR]: CreateReceipt on update ar, " + err.Error())
				}
			}
		}
	}

	defer sqldb.Close()
	return nil
}

func (am *ARModel) SaveDeductMAP(ID string, sqldb *sql.DB) {
	fmt.Println("SaveDeductMAP")
	const mapSql = `INSERT INTO public.DEDUCTMAP (Did, Sid, proportion , SName )
					(
						select * from  
						(
							SELECT d.did, armap.* FROM public.ar ar
							inner join public.deduct d on ar.arid = d.arid
							cross JOIN (
							select Sid, proportion , SName from public.armap where arid = $1
							) armap 
							where ar.arid = $1 
						) tmp						
					) ;`

	_, err := sqldb.Exec(mapSql, ID)
	//res, err := sqldb.Exec(sql, unix_time, receivable.Date, receivable.CNo, receivable.Sales)
	if err != nil {
		fmt.Println("SaveDeductMAP:", err)
	}

}

func (am *ARModel) SaveARMAP(salerList []*MAPSaler, ID string, sqldb *sql.DB) error {
	const mapSql = `INSERT INTO public.ARMAP(
		ARid, Sid, proportion , SName, branch, percent)
		VALUES ($1, $2, $3, $4, $5, $6);`
	count := 0
	for _, element := range salerList {
		// element is the element from someSlice for where we are
		//res, err := sqldb.Exec(mapSql, ID, element.Sid, element.Percent, element.SName, element.Branch)

		// s := am.GetSalerDataByID(sqldb, element.Sid)
		// if s == nil {
		// 	fmt.Println("SaveARMAP unknown error")
		// 	return errors.New("SaveARMAP unknown error")
		// }

		res, err := sqldb.Exec(mapSql, ID, element.Sid, element.Percent, element.SName, element.Branch, element.BonusPercent)
		if err != nil {
			fmt.Println("SaveARMAP ", err)
			return err
		}
		id, err := res.RowsAffected()
		if err != nil {
			fmt.Println("PG Affecte Wrong: ", err)
		}
		count += int(id)
	}
	fmt.Println("SaveARMAP:", count)
	return nil
}

//;

func (am *ARModel) DeleteARandDeductMAP(ID string, sqldb *sql.DB) (err error) {
	fmt.Println("DeleteARandDeductMAP")
	sql := `delete from public.armap where arid = $1`

	_, err = sqldb.Exec(sql, ID)

	sql = `DELETE FROM public.deductmap WHERE did IN (SELECT did FROM public.deduct WHERE arid = $1) ;`

	_, err = sqldb.Exec(sql, ID)

	return nil
}

func (am *ARModel) DeleteAccountReceivable(ID, dbname string) (err error) {
	fmt.Println("DeleteAccountReceivable")
	const sql = `				
				delete from public.receipt where arid = '%s';
				delete from public.commission where arid = '%s';
				DELETE FROM public.deductmap WHERE did IN (SELECT did FROM public.deduct WHERE arid = '%s') ;
				delete from public.deduct where arid = '%s';					 			
				delete from public.armap where arid = '%s';			
				delete from public.returns where arid = '%s';			
				delete from public.returnsbmap where arid = '%s';			
				delete from public.ar where arid = '%s';
				`

	interdb := am.imr.GetSQLDBwithDbname(dbname)
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}
	fmt.Println("sqldb Exec")
	res, err := sqldb.Exec(fmt.Sprintf(sql, ID, ID, ID, ID, ID, ID, ID, ID))
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
	if id <= 0 {
		return errors.New("not found anything")
	}
	defer sqldb.Close()
	return nil
}

func (am *ARModel) DeleteHouseGo(ID, dbname string) (err error) {
	const sql = `DELETE FROM public.housego where id = '%s';`

	interdb := am.imr.GetSQLDBwithDbname(dbname)
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}
	fmt.Println("sqldb Exec")
	res, err := sqldb.Exec(fmt.Sprintf(sql, ID))
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
	if id <= 0 {
		return errors.New("DeleteHouseGo not found anything")
	}
	defer sqldb.Close()
	return nil
}

func (am *ARModel) CreateAccountReceivable(receivable *AR, json, dbname string) (err error) {
	fmt.Println("CreateAccountReceivable")

	const sql = `INSERT INTO public.ar(
		ARid, Date, CNo, CaseName, Type, Name, Amount)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (ARid) DO nothing
		;`
	// ON CONFLICT (ARid) DO UPDATE
	// SET Date = excluded.Date,
	// 	CNo = excluded.CNo,
	// 	Type = excluded.Type,
	// 	CaseName = excluded.CaseName,
	// 	Name = excluded.Name,
	// 	Amount = excluded.Amount

	interdb := am.imr.GetSQLDBwithDbname(dbname)
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}

	//fakeId := time.Now().Unix()
	fakeId := strconv.FormatInt(time.Now().Unix(), 10)
	if receivable.ARid != "" {
		fakeId = receivable.ARid
	}
	//unix_time := time.Time(receivable.CompletionDate).Unix()

	fmt.Println("fakeId:", fakeId)
	fmt.Println("receivable.Date:", receivable.Date)
	fmt.Println("Unix:", receivable.Date.Unix())

	_UTC, err1 := time.LoadLocation("") //等同于"UTC"
	if err1 != nil {
		fmt.Println(err1)
	}

	res, err := sqldb.Exec(sql, fakeId, receivable.Date.In(_UTC), receivable.CNo, receivable.CaseName, receivable.Customer.Action, receivable.Customer.Name, receivable.Amount)
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

	// //住通重複ID的話，需要刪除本來的ID對應，重新建立。
	// if receivable.ARid != "" {
	// 	const mapClearSql = `DELETE FROM public.ARMAP WHERE ARid = $1`
	// 	_, err := sqldb.Exec(mapClearSql, receivable.ARid)
	// 	if err != nil {
	// 		fmt.Println("DELETE ARMAP ", err)
	// 	}
	// }

	//應收帳款成功才建立應收帳款的業務對應表
	if id > 0 {
		err = am.SaveARMAP(receivable.Sales, fakeId, sqldb)
		if err != nil {
			return err
		}
		// const mapSql = `INSERT INTO public.ARMAP(
		// 	ARid, Sid, proportion , SName, branch, percent)
		// 	VALUES ($1, $2, $3, $4, $5, $6);`

		// for _, element := range receivable.Sales {
		// 	s := am.GetSalerDataByID(sqldb, element.Sid)
		// 	if s == nil {
		// 		fmt.Println("ARMAP unknown error")
		// 		return errors.New("ARMAP unknown error")
		// 	}
		// 	// element is the element from someSlice for where we are
		// 	res, err := sqldb.Exec(mapSql, fakeId, element.Sid, element.Percent, element.SName, s.Branch, s.Percent)
		// 	//res, err := sqldb.Exec(sql, unix_time, receivable.Date, receivable.CNo, receivable.Sales)
		// 	if err != nil {
		// 		fmt.Println("ARMAP ", err)
		// 		return err
		// 	}
		// 	id, err := res.RowsAffected()
		// 	if err != nil {
		// 		fmt.Println("PG Affecte Wrong: ", err)
		// 		return err
		// 	}
		// 	fmt.Println(id)
		// }
	}
	if id == 0 {
		am.CreateHouseGoDuplicate(fakeId, json, dbname)
		return errors.New("duplicate data")
	}
	defer sqldb.Close()
	return nil
}
func (am *ARModel) GetSqlDB(dbname string) *sql.DB {
	interdb := am.imr.GetSQLDBwithDbname(dbname)
	sqldb, _ := interdb.ConnectSQLDB()
	return sqldb
}
func (am *ARModel) CheckARExist(mtype, name, branch string, sqldb *sql.DB) (data string) {

	if mtype == "買方" {
		mtype = "buy"
	}
	if mtype == "賣方" {
		mtype = "sell"
	}
	//不知道為什麼用$字號 放入數字會報錯。
	const sql = `SELECT arid FROM public.ar
				 where type = $1 and name = $2;
				`

	// interdb := am.imr.GetSQLDBwithDbname(dbname)
	// sqldb, err := interdb.ConnectSQLDB()
	// if err != nil {
	// 	return err
	// }

	rows, err := sqldb.Query(sql, mtype, name)
	//res, err := sqldb.Exec(sql, unix_time, receivable.Date, receivable.CNo, receivable.Sales)
	if err != nil {
		fmt.Println("checkARExist:", err)
	}
	var ar AR
	var index = 0
	for rows.Next() {
		if err := rows.Scan(&ar.ARid); err != nil {
			fmt.Println("err Scan " + err.Error())
		}
		index++
	}
	fmt.Println("ar.ARid:", ar.ARid, " index:", index)
	if index >= 2 {
		//fmt.Println("超過兩個應收款...請自行判斷")
		return ""
	}
	if ar.ARid == "" {
		return ""
	}
	//defer sqldb.Close()
	return ar.ARid
}

func (am *ARModel) CreateHouseGoDuplicate(ID, data, dbname string) (err error) {

	//不知道為什麼用$字號 放入數字會報錯。
	const sql = `INSERT INTO public.housego
				(arid, id, data)
				VALUES ('%d', '%s', '%s');
				`

	interdb := am.imr.GetSQLDBwithDbname(dbname)
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}
	fmt.Println("CreateHouseGoDuplicate Exec")

	//fakeId := time.Now().Unix()
	fakeId := time.Now().Unix()

	///fmt.Println("fakeId ", fakeId)
	//fmt.Println("ID ", ID)
	//fmt.Println("data :", data)
	data = strings.Replace(data, " ", "", -1)
	data = strings.Replace(data, "\n", "", -1)
	// ID 不取 "_b" && "_s"
	sss := fmt.Sprintf(sql, fakeId, ID[0:len(ID)-2], data)
	//fmt.Println("sss :", sss)
	res, err := sqldb.Exec(sss)
	if err != nil {
		fmt.Println("[error]CreateHouseGoDuplicate:", err)
		return err
	}

	id, err := res.RowsAffected()
	if err != nil {
		fmt.Println("PG Affecte Wrong: ", err)
		return err
	}
	if id <= 0 {
		return errors.New("CreateHouseGoDuplicate not found anything")
	}
	defer sqldb.Close()
	return nil
}

func (rt *Receipt) setRid(id string) {
	rt.Rid = id
}

//not used now (move to public model)
//建立 修改 刪除 收款單時，需要更改應收款項計算項目
func (am *ARModel) UpdateARInfo(arid, dbname string) (err error) {
	//https://stackoverflow.com/questions/2334712/how-do-i-update-from-a-select-in-sql-server
	const sql = `Update public.ar
				 set
					ra = t2.sum , balance = amount - fee -t2.sum
				FROM (
					 select sum(amount) from public.receipt where arid = $1 group by arid  
				)as t2 where ar.arid = $1`

	interdb := am.imr.GetSQLDBwithDbname(dbname)
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}
	fmt.Println("sqldb Exec")
	res, err := sqldb.Exec(sql, arid)
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

	if id <= 0 {
		fmt.Println("No any receipt, so reset infomation of account receivable")
		const reset = `Update public.ar	set ra = 0 , balance = amount - fee  where arid = $1`
		res, err := sqldb.Exec(reset, arid)
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
		fmt.Println("reset update:", id)
		return err
	}
	defer sqldb.Close()
	return nil
}

// func (a *Answer) GetTime() time.Time {
// 	return time.Date(a.Date.Year(), a.Date.Month(), a.Date.Day(), 0, 0, 0, 0, a.Date.Location())
// }

// func fetchSales(sale *sale) error {

// }
func (am *ARModel) UpgradeARInfo(arid, dbname string) (err error) {

	select_sql := "SELECT arid, id, data FROM public.housego where arid = '" + arid + "';"

	/*
	*balance equal ar.amount - COALESCE((SELECT SUM(r.amount) FROM public.receipt r WHERE ar.arid = r.arid),0) AS SUM_RA
	*but I do with r.Balance = r.Amount - r.RA
	 */

	//const qspl = `SELECT arid,sales	FROM public.ar;`
	db := am.imr.GetSQLDBwithDbname(dbname)
	//fmt.Println(sql)
	//rows, err := db.SQLCommand(fmt.Sprintf(sql))

	rows, err := db.SQLCommand(select_sql)

	if err != nil {
		fmt.Println(err)
		return nil
	}
	var hgList []*HouseGo

	var data string    //原本的json字串
	var Oldarid string //儲存於AR TABLE的arid (不含 _s or _b)
	for rows.Next() {
		var hg HouseGo

		if err := rows.Scan(&hg.ARid, &Oldarid, &data); err != nil {
			fmt.Println("err Scan " + err.Error())
		}

		hgList = append(hgList, &hg)
	}

	iGoAR := inputhouseGoAR{}
	err = json.Unmarshal([]byte(data), &iGoAR)
	if err != nil {
		return err
	}
	fmt.Println(iGoAR)
	am.DeleteAccountReceivable(Oldarid+"_b", dbname)
	am.DeleteAccountReceivable(Oldarid+"_s", dbname)
	am.DeleteHouseGo(Oldarid, dbname)
	ar := iGoAR.GetAR(ACTION_BUY)
	err = am.CreateAccountReceivable(ar, data, dbname)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	ar = iGoAR.GetAR(ACTION_SELL)
	err = am.CreateAccountReceivable(ar, data, dbname)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	return nil
}

func (iGoAR *inputhouseGoAR) GetAR(action string) *AR {
	var customer = Customer{}
	var amount = 0
	var arid string
	customer.Action = "none"
	if action == ACTION_BUY {
		customer.Action = ACTION_BUY
		amount = iGoAR.Buyer.Amount
		arid = strconv.Itoa(iGoAR.ARid) + "_b"
	} else if action == ACTION_SELL {
		arid = strconv.Itoa(iGoAR.ARid) + "_s"
		customer.Action = ACTION_SELL
		amount = iGoAR.Sell.Amount
	}

	customer.Name = iGoAR.Buyer.Name

	var sales []*MAPSaler
	for _, element := range iGoAR.Sales {
		var data = MAPSaler{}
		data.Percent = element.Percent
		data.SName = element.SName
		data.Sid = element.Sid
		sales = append(sales, &data)
	}

	return &AR{
		ARid:     arid,
		Amount:   amount,
		Date:     iGoAR.Date,
		CNo:      iGoAR.CNo,
		CaseName: iGoAR.CaseName,
		Customer: customer,
		//Fee:      iAR.Fee,
		Sales: sales,
	}
}

func (am *ARModel) checkEditable(ID string, sqldb *sql.DB) []mapRidBsid {

	const mapSql = `select * from (
		select c.rid , COALESCE(bs.bsid,'') bsid , COALESCE(bs.branch,''), COALESCE(bs.name,'') from (
			select arid, rid from receipt where arid = $1
		) r inner join (
		  select * from 
			(select rid, bsid from commission) tmpc
		) c on c.rid = r.rid 
		LEFT JOIN branchsalary bs on bs.bsid = c.bsid		
	   ) tmp order by bsid desc;`

	rows, err := sqldb.Query(mapSql, ID)
	//res, err := sqldb.Exec(sql, unix_time, receivable.Date, receivable.CNo, receivable.Sales)
	if err != nil {
		fmt.Println("checkEditable:", err)
	}

	mapList := []mapRidBsid{}
	for rows.Next() {
		var mapId mapRidBsid
		if err := rows.Scan(&mapId.rid, &mapId.bsid, &mapId.branch, &mapId.salaryName); err != nil {
			fmt.Println("err Scan " + err.Error())
		}
		mapList = append(mapList, mapId)

	}

	return mapList
}

func (am *ARModel) ReCount() {

	date := "1980-12-31T00:00:00.000Z"

	t, _ := time.Parse(time.RFC3339, date)

	arList := am.GetARData("", "0", "", "toad", t)
	for _, element := range arList {
		am.UpdateAccountReceivable(element.Amount, element.ARid, "toad", element.Sales)
	}
}
