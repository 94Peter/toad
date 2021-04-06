package model

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"toad/resource/db"
)

//`json:"id"` 回傳重新命名
type Returns struct {
	Return_id   string             `json:"return_id"`
	Date        time.Time          `json:"date"`
	Description string             `json:"description"`
	Amount      int                `json:"amount"`
	Status      string             `json:"status"`
	Arid        string             `json:"arid"`
	Sales       []*ReturnMAPSaler  `json:"sales"`
	BranchList  []*ReturnMAPBranch `json:"branchList"`
}

var (
	returnsM *ReturnsModel
)

type ReturnMAPSaler struct {
	SName string `json:"name"`
	//Percent      float64 `json:"proportion"`
	//BonusPercent float64 `json:"percent"`
	Sid    string `json:"account"`
	Branch string `json:"branch"`
	SR     int    `json:"sr"`

	return_id string `json:"-"`
}

type ReturnMAPBranch struct {
	Branch    string `json:"branch"`
	Return_id string `json:"-"`
	Arid      string `json:"-"`
	SR        int    `json:"sr"`
}

func GetReturnsModel(imr interModelRes) *ReturnsModel {
	if returnsM != nil {
		return returnsM
	}

	returnsM = &ReturnsModel{
		imr: imr,
	}
	return returnsM
}

type ReturnsModel struct {
	imr interModelRes
	db  db.InterSQLDB
}

func (returnsM *ReturnsModel) GetReturnsData(branch, dbname string) []*Returns {

	const sql = `SELECT return_id, arid, date, status, description, amount FROM public.Returns ;`
	//const qspl = `SELECT arid,sales	FROM public.ar;`
	db := returnsM.imr.GetSQLDBwithDbname(dbname)
	//fmt.Println(fmt.Sprintf(qspl, branch))
	sqldb, err := db.ConnectSQLDB()
	rows, err := sqldb.Query(fmt.Sprintf(sql))
	if err != nil {
		fmt.Println(err)
		return nil
	}
	var rList []*Returns

	for rows.Next() {
		var r Returns

		// if err := rows.Scan(&r.ARid, &s); err != nil {
		// 	fmt.Println("err Scan " + err.Error())
		// }
		if err := rows.Scan(&r.Return_id, &r.Arid, &r.Date, &r.Status, &r.Description, &r.Amount); err != nil {
			fmt.Println("err Scan " + err.Error())
		}
		salesList := []*ReturnMAPSaler{}
		branchList := []*ReturnMAPBranch{}
		r.Sales = salesList
		r.BranchList = branchList
		rList = append(rList, &r)
	}
	returnsM.GetReturnsCommissionMap(rList, sqldb)
	returnsM.GetReturnsBranchMap(rList, sqldb)

	return rList
}
func (returnsM *ReturnsModel) GetReturnsDataByID(ID string, sqldb *sql.DB) *Returns {

	const sql = `SELECT return_id, arid, date, status, description, amount FROM public.Returns where return_id = $1 ;`
	//const qspl = `SELECT arid,sales	FROM public.ar;`

	//fmt.Println(fmt.Sprintf(qspl, branch))
	rows, err := sqldb.Query(fmt.Sprintf(sql), ID)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	for rows.Next() {
		var r Returns

		// if err := rows.Scan(&r.ARid, &s); err != nil {
		// 	fmt.Println("err Scan " + err.Error())
		// }
		if err := rows.Scan(&r.Return_id, &r.Arid, &r.Date, &r.Status, &r.Description, &r.Amount); err != nil {
			fmt.Println("err Scan " + err.Error())
		}

		return &r
	}

	return nil
}

func (returnsM *ReturnsModel) GetReturnsBranchMap(rList []*Returns, sqldb *sql.DB) {

	const sql = `SELECT return_id, arid, sr, branch FROM public.returnsbmap ;`

	rows, err := sqldb.Query(fmt.Sprintf(sql))
	if err != nil {
		fmt.Println(err)
		return
	}

	for rows.Next() {
		var bms ReturnMAPBranch

		// if err := rows.Scan(&r.ARid, &s); err != nil {
		// 	fmt.Println("err Scan " + err.Error())
		// }
		if err := rows.Scan(&bms.Return_id, &bms.Arid, &bms.SR, &bms.Branch); err != nil {
			fmt.Println("err Scan " + err.Error())
		}

		for _, r := range rList {
			if r.Return_id == bms.Return_id {
				r.BranchList = append(r.BranchList, &bms)
				//fmt.Println(arid)
				break
			}
		}

	}

}
func (returnsM *ReturnsModel) GetReturnsCommissionMap(rList []*Returns, sqldb *sql.DB) {

	const sql = `SELECT branch, sid, sname, fee, rid FROM public.commission where rid like 'r%';`

	rows, err := sqldb.Query(sql)
	if err != nil {
		fmt.Println("GetReturnsCommissionMap:", err)
		return
	}

	for rows.Next() {
		var bms ReturnMAPSaler

		// if err := rows.Scan(&r.ARid, &s); err != nil {
		// 	fmt.Println("err Scan " + err.Error())
		// }
		if err := rows.Scan(&bms.Branch, &bms.Sid, &bms.SName, &bms.SR, &bms.return_id); err != nil {
			fmt.Println("err Scan " + err.Error())
		}

		for _, r := range rList {
			if r.Return_id == bms.return_id {
				r.Sales = append(r.Sales, &bms)
				//fmt.Println(arid)
				break
			}
		}

	}

}
func (returnsM *ReturnsModel) GetInvoiceDataMap(rList []*Returns, sqldb *sql.DB) {

	const sql = `SELECT branch, sid, sname, fee, rid FROM public.commission where rid like 'r%';`

	rows, err := sqldb.Query(sql)
	if err != nil {
		fmt.Println("GetReturnsCommissionMap:", err)
		return
	}

	for rows.Next() {
		var bms ReturnMAPSaler

		// if err := rows.Scan(&r.ARid, &s); err != nil {
		// 	fmt.Println("err Scan " + err.Error())
		// }
		if err := rows.Scan(&bms.Branch, &bms.Sid, &bms.SName, &bms.SR, &bms.return_id); err != nil {
			fmt.Println("err Scan " + err.Error())
		}

		for _, r := range rList {
			if r.Return_id == bms.return_id {
				r.Sales = append(r.Sales, &bms)
				//fmt.Println(arid)
				break
			}
		}

	}

}

func (returnsM *ReturnsModel) UpdateReturns(returns *Returns, dbname string) (err error) {

	const sql = `Update public.returns 
					set	amount = $1 , description = $2 , status = $4			
				where return_id = $3 ;`

	interdb := returnsM.imr.GetSQLDBwithDbname(dbname)
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}
	tmp := returnsM.GetReturnsDataByID(returns.Return_id, sqldb)
	if tmp == nil {
		return errors.New("[ERROR]: 無該選擇之折讓單")
	} else {
		returns.Arid = tmp.Arid
		fmt.Println("returns arid: ", returns.Arid)
	}
	cList := returnsM.GetCommissionDataByRerutnsId(sqldb, returns.Return_id)
	errmsg := ""
	for _, element := range cList {
		if element.Bsid != "" {
			errmsg += element.SName + "(" + element.Branch + ")"
		}
	}
	if errmsg != "" {
		defer sqldb.Close()
		return errors.New("[ERROR]:" + errmsg + "折讓單已綁定薪資表")
	}
	//確認可以更新的話，每次都重建資料。
	returnsM.resetReturns(returns.Return_id, sqldb)

	status := "0"
	if len(returns.Sales) > 0 {
		status = "1"
	}
	if len(returns.BranchList) > 0 {
		status = "2"
	}
	// if len(returns.Sales) > 0 {
	// 	status = "1"
	// }
	res, err := sqldb.Exec(sql, returns.Amount, returns.Description, returns.Return_id, status)
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
		return errors.New("[ERROR]: 更新折讓單錯誤")
	}
	// 	//刪除ARMAP 重建
	// 	am.DeleteARandDeductMAP(ID, sqldb)
	// 	am.SaveARMAP(salerList, ID, sqldb)
	// 	am.SaveDeductMAP(ID, sqldb)
	// 	//delete old receipt commission && rebuild receipt!
	// 	doneMsg := ""
	for i := 0; i < len(returns.Sales); i++ {
		saler := returns.Sales[i]
		err = returnsM.CreateCommissionByReturns(returns.Return_id, returns.Arid, saler, sqldb)
		if err != nil {
			fmt.Println("[ERROR] 建立傭金錯誤 returns.Return_id:", returns.Return_id)
			return err
		}
	}

	for i := 0; i < len(returns.BranchList); i++ {
		bms := returns.BranchList[i]
		err = returnsM.SaveReturnMAP(returns.Return_id, returns.Arid, bms, sqldb)
		if err != nil {
			fmt.Println("[ERROR] 建立分店扣除業績錯誤 returns.Return_id:", returns.Return_id)
			return err
		}
	}

	defer sqldb.Close()
	return nil
}

func (returnsM *ReturnsModel) SaveReturnMAP(return_id, arid string, bms *ReturnMAPBranch, sqldb *sql.DB) (err error) {

	const mapSql = `INSERT INTO public.returnsbmap(
						branch, return_id, arid, sr)
						VALUES ($1, $2, $3, $4) ;`

	_, err = sqldb.Exec(mapSql, bms.Branch, return_id, arid, bms.SR)
	//res, err := sqldb.Exec(sql, unix_time, receivable.Date, receivable.CNo, receivable.Sales)
	if err != nil {
		fmt.Println("SaveReturnMAP:", err)
	}

	return nil
}

func (returnsM *ReturnsModel) DeleteReturns(ID, dbname string) (err error) {

	const sql = `
				delete from public.returns where return_id = '%s';			
				delete from public.returnsbmap where return_id = '%s';
				delete from public.commission where rid = '%s';
				`

	interdb := returnsM.imr.GetSQLDBwithDbname(dbname)
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}

	cList := returnsM.GetCommissionDataByRerutnsId(sqldb, ID)
	errmsg := ""
	for _, element := range cList {
		if element.Bsid != "" {
			errmsg += element.SName + "(" + element.Branch + ")"
		}
	}
	if errmsg != "" {
		defer sqldb.Close()
		return errors.New("[ERROR]:" + errmsg + "折讓單已綁定薪資表")
	}

	res, err := sqldb.Exec(fmt.Sprintf(sql, ID, ID, ID))
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

func (returnsM *ReturnsModel) resetReturns(ID string, sqldb *sql.DB) (err error) {

	const sql = `				
				delete from public.returnsbmap where return_id = '%s';
				delete from public.commission where rid = '%s';
				`

	res, err := sqldb.Exec(fmt.Sprintf(sql, ID, ID))
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
		return errors.New("resetReturns not found anything")
	}
	return nil
}

// func (am *ARModel) DeleteHouseGo(ID, dbname string) (err error) {
// 	const sql = `DELETE FROM public.housego where id = '%s';`

// 	interdb := am.imr.GetSQLDBwithDbname(dbname)
// 	sqldb, err := interdb.ConnectSQLDB()
// 	if err != nil {
// 		return err
// 	}
// 	fmt.Println("sqldb Exec")
// 	res, err := sqldb.Exec(fmt.Sprintf(sql, ID))
// 	//res, err := sqldb.Exec(sql, unix_time, receivable.Date, receivable.CNo, receivable.Sales)
// 	if err != nil {
// 		fmt.Println(err)
// 		return err
// 	}

// 	id, err := res.RowsAffected()
// 	if err != nil {
// 		fmt.Println("PG Affecte Wrong: ", err)
// 		return err
// 	}
// 	if id <= 0 {
// 		return errors.New("DeleteHouseGo not found anything")
// 	}
// 	defer sqldb.Close()
// 	return nil
// }

func (returnsM *ReturnsModel) GetReceiptDataByArid(sqldb *sql.DB, arid string) *Receipt {

	const qspl = `SELECT r.arid
					FROM public.receipt r							
					where r.arid = '%s'`

	//db := rm.imr.GetSQLDBwithDbname(dbname)
	//sqldb, err := db.ConnectSQLDB()
	rows, err := sqldb.Query(fmt.Sprintf(qspl, arid))
	if err != nil {
		fmt.Println("[rows err]:", err)
		return nil
	}

	var rt Receipt
	for rows.Next() {

		fmt.Println("scan start")
		if err := rows.Scan(&rt.ARid); err != nil {
			fmt.Println("err Scan " + err.Error())
		}
		fmt.Println("scan end")
		return &rt
	}
	return nil

}

func (returnsM *ReturnsModel) CreateReturns(returns *Returns, dbname string) (string, error) {

	const sql = `INSERT INTO public.Returns(
		ARid, Description, Return_id, Status, Amount, Date)
		VALUES ($1, $2, $3, '0', $4, now())
		ON CONFLICT (Return_id) DO nothing
		;`
	// ON CONFLICT (ARid) DO UPDATE
	// SET Date = excluded.Date,
	// 	CNo = excluded.CNo,
	// 	Type = excluded.Type,
	// 	CaseName = excluded.CaseName,
	// 	Name = excluded.Name,
	// 	Amount = excluded.Amount

	interdb := returnsM.imr.GetSQLDBwithDbname(dbname)
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return "", err
	}

	receipt := returnsM.GetReceiptDataByArid(sqldb, returns.Arid)
	if receipt == nil {
		fmt.Println("receipt is not exist, create deduct")
		d := &Deduct{
			ARid:        returns.Arid,
			Item:        "其他",
			Description: "折讓退回." + returns.Description,
			Fee:         returns.Amount,
		}
		return "", decuctModel.CreateDeduct(d, dbname)
	}

	fakeId := time.Now().Unix()
	strInt64 := "r" + strconv.FormatInt(fakeId, 10)

	res, err := sqldb.Exec(sql, returns.Arid, returns.Description, strInt64, returns.Amount)
	//res, err := sqldb.Exec(sql, unix_time, receivable.Date, receivable.CNo, receivable.Sales)
	if err != nil {
		fmt.Println("CreateReturns:", err)
		return "", err
	}

	id, err := res.RowsAffected()
	if err != nil {
		fmt.Println("PG Affecte Wrong: ", err)
		return "", err
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

	// if id > 0 {
	// 	err = returnsM.SaveReturnMAP(fakeId, sqldb)
	// 	if err != nil {
	// 		return err
	// 	}
	// }
	// 	// const mapSql = `INSERT INTO public.ARMAP(
	// 	// 	ARid, Sid, proportion , SName, branch, percent)
	// 	// 	VALUES ($1, $2, $3, $4, $5, $6);`

	// 	// for _, element := range receivable.Sales {
	// 	// 	s := am.GetSalerDataByID(sqldb, element.Sid)
	// 	// 	if s == nil {
	// 	// 		fmt.Println("ARMAP unknown error")
	// 	// 		return errors.New("ARMAP unknown error")
	// 	// 	}
	// 	// 	// element is the element from someSlice for where we are
	// 	// 	res, err := sqldb.Exec(mapSql, fakeId, element.Sid, element.Percent, element.SName, s.Branch, s.Percent)
	// 	// 	//res, err := sqldb.Exec(sql, unix_time, receivable.Date, receivable.CNo, receivable.Sales)
	// 	// 	if err != nil {
	// 	// 		fmt.Println("ARMAP ", err)
	// 	// 		return err
	// 	// 	}
	// 	// 	id, err := res.RowsAffected()
	// 	// 	if err != nil {
	// 	// 		fmt.Println("PG Affecte Wrong: ", err)
	// 	// 		return err
	// 	// 	}
	// 	// 	fmt.Println(id)
	// 	// }
	// }

	defer sqldb.Close()
	return strInt64, nil
}
func (returnsM *ReturnsModel) CreateCommissionByReturns(return_id, arid string, returnsSaler *ReturnMAPSaler, sqldb *sql.DB) (err error) {

	const sql = `INSERT INTO public.commission
	(Sid, Rid, Item, SName, CPercent, sr, bonus , arid, fee, branch)
	select $1, $2, ar.cno ||' '|| ar.casename ||' '|| (Case When AR.type = 'buy' then '買' When AR.type = 'sell' then '賣' else 'unknown' End ), $3,
	100 , $4 * -1 ,   $4 * -1 * armap.percent / 100 , $5::VARCHAR , $4, $6
	from public.ar ar
	inner join 	public.armap armap on armap.arid = ar.arid 	and armap.sid = $1
	where ar.arid = $5
	ON CONFLICT (Sid,Rid) DO UPDATE SET sr = excluded.sr, bonus = excluded.bonus, fee = excluded.fee, branch = excluded.branch, SName = excluded.SName`

	res, err := sqldb.Exec(sql, returnsSaler.Sid, return_id, returnsSaler.SName, returnsSaler.SR, arid, returnsSaler.Branch)
	//res, err := sqldb.Exec(sql, unix_time, receivable.Date, receivable.CNo, receivable.Sales)
	if err != nil {
		fmt.Println("CreateCommissionByReturns:", err)
		return err
	}
	id, err := res.RowsAffected()
	if err != nil {
		fmt.Println("PG Affecte Wrong: ", err)
		return err
	}

	if id == 0 {
		return errors.New("Invalid operation, CreateCommissionByReturns")
	}

	return nil
}

// func (am *ARModel) CreateHouseGoDuplicate(ID, data, dbname string) (err error) {

// 	//不知道為什麼用$字號 放入數字會報錯。
// 	const sql = `INSERT INTO public.housego
// 				(arid, id, data)
// 				VALUES ('%d', '%s', '%s');
// 				`

// 	interdb := am.imr.GetSQLDBwithDbname(dbname)
// 	sqldb, err := interdb.ConnectSQLDB()
// 	if err != nil {
// 		return err
// 	}
// 	fmt.Println("CreateHouseGoDuplicate Exec")

// 	//fakeId := time.Now().Unix()
// 	fakeId := time.Now().Unix()

// 	///fmt.Println("fakeId ", fakeId)
// 	//fmt.Println("ID ", ID)
// 	//fmt.Println("data :", data)
// 	data = strings.Replace(data, " ", "", -1)
// 	data = strings.Replace(data, "\n", "", -1)
// 	// ID 不取 "_b" && "_s"
// 	sss := fmt.Sprintf(sql, fakeId, ID[0:len(ID)-2], data)
// 	//fmt.Println("sss :", sss)
// 	res, err := sqldb.Exec(sss)
// 	if err != nil {
// 		fmt.Println("[error]CreateHouseGoDuplicate:", err)
// 		return err
// 	}

// 	id, err := res.RowsAffected()
// 	if err != nil {
// 		fmt.Println("PG Affecte Wrong: ", err)
// 		return err
// 	}
// 	if id <= 0 {
// 		return errors.New("CreateHouseGoDuplicate not found anything")
// 	}
// 	defer sqldb.Close()
// 	return nil
// }
func (returnsM *ReturnsModel) GetCommissionDataByRerutnsId(sqldb *sql.DB, return_id string) []*Commission {

	const sql = `SELECT bsid, rid, sid, sname, branch
					FROM public.commission 						
					where rid  = $1 `

	//db := rm.imr.GetSQLDBwithDbname(dbname)
	//sqldb, err := db.ConnectSQLDB()
	rows, err := sqldb.Query(sql, return_id)
	if err != nil {
		fmt.Println("[GetCommissionDataByRerutnsId err]:", err)
		return nil
	}
	var cList []*Commission
	for rows.Next() {
		var c Commission
		var Bsid NullString
		if err := rows.Scan(&Bsid, &c.Rid, &c.Sid, &c.SName, &c.Branch); err != nil {
			fmt.Println("GetCommissionDataByRerutnsId err Scan " + err.Error())
		}
		c.Bsid = Bsid.Value
		cList = append(cList, &c)
		fmt.Println(c)
	}

	return cList

}
