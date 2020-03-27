package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/94peter/toad/resource/db"
)

var ACTION_BUY = "buy"
var ACTION_SELL = "sell"

//`json:"id"` 回傳重新命名
type AR struct {
	ARid     string      `json:"id"`
	Date     time.Time   `json:"completionDate"`
	CNo      string      `json:"contractNo"`
	Customer Customer    `json:"customer"`
	CaseName string      `json:"caseName"`
	Amount   int         `json:"amount"`
	Fee      int         `json:"fee"`
	Balance  int         `json:"balance"`        //未收金額
	RA       int         `json:"receivedAmount"` //已收金額
	Sales    []*MAPSaler `json:"sales"`
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
	SName   string  `json:"name"`
	Percent float64 `json:"proportion"` //{"{\"BName\":\"123\",\"Bid\":\"13\",\"Persent\":12}","{\"BName\":\"123\",\"Bid\":\"13\",\"Persent\":12}"}
	Sid     string  `json:"account"`
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
	imr       interModelRes
	db        db.InterSQLDB
	arList    []*AR
	salerList []*Saler
	hgList    []*HouseGo
}

func (am *ARModel) GetSalerData(branch string) []*Saler {

	const qspl = `SELECT A.sid, A.sname, A.branch, A.percent, A.title
					FROM public.ConfigSaler A 
					Inner Join ( 
						select sid, max(zerodate) zerodate from public.configsaler cs 
						where now() > zerodate
						group by sid 
					) B on A.sid=B.sid and A.zeroDate = B.zeroDate
					where A.branch like '%s';`
	//const qspl = `SELECT arid,sales	FROM public.ar;`
	db := am.imr.GetSQLDB()
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

func (am *ARModel) GetARData(today, end time.Time, key string) []*AR {

	// const sql = `SELECT
	// 			ar.arid, ar.date, ar.cno, ar.casename, ar.type, ar.name, ar.amount,
	// 				COALESCE((SELECT SUM(d.fee) FROM public.deduct d WHERE ar.arid = d.arid),0) AS SUM_Fee,
	// 				COALESCE((SELECT SUM(r.amount) FROM public.receipt r WHERE ar.arid = r.arid),0) AS SUM_RA,
	// 				ar.sales
	// 			where ar.arid like '%%s%'  OR ar.cno like '%%s%' OR ar.casename like '%%s%' OR ar.type like '%%s%' OR ar.name like '%%s%'
	// 			FROM public.ar ar
	// 			group by ar.arid;`
	index := "%" + key + "%"
	sql := "SELECT ar.arid, ar.date, ar.cno, ar.casename, ar.type, ar.name, ar.amount, " +
		"	COALESCE((SELECT SUM(d.fee) FROM public.deduct d WHERE ar.arid = d.arid),0) AS SUM_Fee," +
		"	COALESCE((SELECT SUM(r.amount) FROM public.receipt r WHERE ar.arid = r.arid),0) AS SUM_RA" +
		"   " +
		"FROM public.ar ar	" +
		"where ar.arid like '" + index + "' OR ar.cno like '" + index + "' OR ar.casename like '" + index + "' OR ar.type like '" + index + "' OR ar.name like '" + index + "' " +
		"group by ar.arid;"
	/*
	*balance equal ar.amount - COALESCE((SELECT SUM(r.amount) FROM public.receipt r WHERE ar.arid = r.arid),0) AS SUM_RA
	*but I do with r.Balance = r.Amount - r.RA
	 */

	db := am.imr.GetSQLDB()
	//fmt.Println(sql)
	//rows, err := db.SQLCommand(fmt.Sprintf(sql))

	rows, err := db.SQLCommand(sql)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	var arDataList []*AR

	for rows.Next() {
		var r AR

		var ctm Customer
		//var col_sales string
		// if err := rows.Scan(&r.ARid, &s); err != nil {
		// 	fmt.Println("err Scan " + err.Error())
		// }
		if err := rows.Scan(&r.ARid, &r.Date, &r.CNo, &r.CaseName, &ctm.Action, &ctm.Name, &r.Amount, &r.Fee, &r.RA); err != nil {
			fmt.Println("err Scan " + err.Error())
		}
		r.Customer = ctm
		r.Balance = r.Amount - r.RA
		// err := json.Unmarshal([]byte(col_sales), &r.Sales)
		// if err != nil {
		// 	fmt.Println(err)
		// }

		arDataList = append(arDataList, &r)
	}

	const Mapsql = `SELECT arid, sid, proportion, sname	FROM public.armap; `
	rows, err = db.SQLCommand(Mapsql)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	for rows.Next() {
		var arid string
		var saler MAPSaler

		if err := rows.Scan(&arid, &saler.Sid, &saler.Percent, &saler.SName); err != nil {
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
	// out, err := json.Marshal(arList)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(string(out))
	am.arList = arDataList
	return am.arList
	// influxR := result[0]
	// if len(influxR.Series) == 0 {
	// 	return nil
	// }
	// data := influxR.Series[0]

	// var arList []*dailyTrend
	// var myvalue float64
	// var jvalue json.Number
	// for _, value := range data.Values {
	// 	jvalue = value[5].(json.Number)
	// 	myvalue, _ = jvalue.Float64()
	// 	dailyTrendList = append(dailyTrendList, &dailyTrend{
	// 		Category: value[1].(string),
	// 		CType:    value[2].(string),
	// 		Name:     value[3].(string),
	// 		Unit:     value[4].(string),
	// 		Value:    myvalue,
	// 		Msg:      value[6].(string),
	// 	})
	// }
	// return dailyTrendList

}

func (am *ARModel) GetHouseGoData(today, end time.Time, key string) []*HouseGo {

	//index := "%" + key + "%"
	sql := "SELECT arid, id, data FROM public.housego"

	/*
	*balance equal ar.amount - COALESCE((SELECT SUM(r.amount) FROM public.receipt r WHERE ar.arid = r.arid),0) AS SUM_RA
	*but I do with r.Balance = r.Amount - r.RA
	 */

	//const qspl = `SELECT arid,sales	FROM public.ar;`
	db := am.imr.GetSQLDB()
	//fmt.Println(sql)
	//rows, err := db.SQLCommand(fmt.Sprintf(sql))

	rows, err := db.SQLCommand(sql)
	if err != nil {
		fmt.Println(err)
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

func (am *ARModel) Json(mtype string) ([]byte, error) {
	switch mtype {
	case "ar":
		return json.Marshal(am.arList)
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

func (am *ARModel) UpdateAccountReceivable(amount int, ID string, salerList []*MAPSaler) (err error) {
	fmt.Println("UpdateAccountReceivable")
	const sql = `Update public.ar t1
					set	amount = $1
				FROM (
					SELECT ar.arid, ar.amount, sum(r.amount)
					FROM public.ar ar
					LEFT JOIN public.receipt r ON ar.arid = r.arid
					where ar.arid = $2
					group by ar.arid
					HAVING sum(r.amount)<= $1
				)as t2 
				where t1.arid = $2;`

	interdb := am.imr.GetSQLDB()
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
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
	//連動更改ARMAP TABLE的數值
	am.UpdateAccountReceivableSalerProportion(salerList, ID)
	//連動更改傭金明細TABLE的數值
	am.RefreshCommissionBonus(ID)

	return nil
}

func (am *ARModel) UpdateAccountReceivableSalerProportion(salerList []*MAPSaler, ID string) (err error) {
	fmt.Println("UpdateAccountReceivable")
	const sql = `Update public.armap set proportion = $1				
				where arid = $2 and sid = $3`

	interdb := am.imr.GetSQLDB()
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}
	for _, element := range salerList {
		// element is the element from someSlice for where we are
		res, err := sqldb.Exec(sql, element.Percent, ID, element.Sid)
		//res, err := sqldb.Exec(sql, unix_time, receivable.Date, receivable.CNo, receivable.Sales)
		if err != nil {
			fmt.Println("ARMAP ", err)
			return err
		}
		id, err := res.RowsAffected()
		if err != nil {
			fmt.Println("PG Affecte Wrong: ", err)
			return err
		}
		fmt.Println("UpdateAccountReceivableSalerProportion:", id)
	}

	return nil
}

func (am *ARModel) RefreshCommissionBonus(ID string) (err error) {

	const sql = `Update public.commission t1
					set cpercent = t2.proportion, sr= (t2.amount - t2.fee)*t2.proportion/100, bonus= (t2.amount - t2.fee)*t2.proportion/100*t2.percent /100
				FROM(	
					SELECT  map.proportion , c.bsid, c.sid, c.rid, r.amount, c.fee , c.sr, c.bonus,  cs.percent
						FROM public.commission c
						inner JOIN public.receipt r on r.rid = c.rid 			
						inner join public.configsaler cs on cs.sid = c.sid
						inner join public.armap map on map.sid = c.sid  and map.arid = r.arid and map.arid = $1
						WHERE c.bsid is null
				) as t2 where t1.sid = t2.sid and t1.rid = t2.rid `
	interdb := am.imr.GetSQLDB()
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

	if id >= 1 {
		fmt.Println("RefreshCommissionBonus success")
	} else {
		fmt.Println("RefreshCommissionBonus error : something error")
	}

	return nil
}

func (am *ARModel) DeleteAccountReceivable(ID string) (err error) {
	fmt.Println("DeleteAccountReceivable")
	const sql = `
				delete from public.ar where arid = '%s';
				delete from public.receipt where arid = '%s';
				delete from public.commission where arid = '%s';
				delete from public.deduct where arid = '%s';
				delete from public.armap where arid = '%s';			
				`

	interdb := am.imr.GetSQLDB()
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}
	fmt.Println("sqldb Exec")
	res, err := sqldb.Exec(fmt.Sprintf(sql, ID, ID, ID, ID, ID))
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
	return nil
}

func (am *ARModel) DeleteHouseGo(ID string) (err error) {
	const sql = `DELETE FROM public.housego where id = '%s';`

	interdb := am.imr.GetSQLDB()
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
	return nil
}

func (am *ARModel) CreateAccountReceivable(receivable *AR, json string) (err error) {
	fmt.Println("CreateAccountReceivable")
	//to_timestamp
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

	interdb := am.imr.GetSQLDB()
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
	fmt.Println(receivable.Date.In(_UTC))
	fmt.Println(receivable.Date.In(_UTC).Unix())

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
		const mapSql = `INSERT INTO public.ARMAP(
			ARid, Sid, proportion , SName )
			VALUES ($1, $2, $3, $4);`

		for _, element := range receivable.Sales {
			// element is the element from someSlice for where we are
			res, err := sqldb.Exec(mapSql, fakeId, element.Sid, element.Percent, element.SName)
			//res, err := sqldb.Exec(sql, unix_time, receivable.Date, receivable.CNo, receivable.Sales)
			if err != nil {
				fmt.Println("ARMAP ", err)
				return err
			}
			id, err := res.RowsAffected()
			if err != nil {
				fmt.Println("PG Affecte Wrong: ", err)
				return err
			}
			fmt.Println(id)
		}
	}
	if id == 0 {
		am.CreateHouseGoDuplicate(fakeId, json)
		return errors.New("duplicate data")
	}
	return nil
}

func (am *ARModel) CreateHouseGoDuplicate(ID, data string) (err error) {

	//不知道為什麼用$字號 放入數字會報錯。
	const sql = `INSERT INTO public.housego
				(arid, id, data)
				VALUES ('%d', '%s', '%s');
				`

	interdb := am.imr.GetSQLDB()
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
	return nil
}

func (rt *Receipt) setRid(id string) {
	rt.Rid = id
}

//not used now (move to public model)
//建立 修改 刪除 收款單時，需要更改應收款項計算項目
func (am *ARModel) UpdateARInfo(arid string) (err error) {
	//https://stackoverflow.com/questions/2334712/how-do-i-update-from-a-select-in-sql-server
	const sql = `Update public.ar
				 set
					ra = t2.sum , balance = amount - fee -t2.sum
				FROM (
					 select sum(amount) from public.receipt where arid = $1 group by arid  
				)as t2 where ar.arid = $1`

	interdb := am.imr.GetSQLDB()
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

	return nil
}

// func (a *Answer) GetTime() time.Time {
// 	return time.Date(a.Date.Year(), a.Date.Month(), a.Date.Day(), 0, 0, 0, 0, a.Date.Location())
// }

// func fetchSales(sale *sale) error {

// }
func (am *ARModel) UpgradeARInfo(arid string) (err error) {

	select_sql := "SELECT arid, id, data FROM public.housego where arid = '" + arid + "';"

	/*
	*balance equal ar.amount - COALESCE((SELECT SUM(r.amount) FROM public.receipt r WHERE ar.arid = r.arid),0) AS SUM_RA
	*but I do with r.Balance = r.Amount - r.RA
	 */

	//const qspl = `SELECT arid,sales	FROM public.ar;`
	db := am.imr.GetSQLDB()
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
	am.DeleteAccountReceivable(Oldarid + "_b")
	am.DeleteAccountReceivable(Oldarid + "_s")
	am.DeleteHouseGo(Oldarid)
	ar := iGoAR.GetAR(ACTION_BUY)
	err = am.CreateAccountReceivable(ar, data)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	ar = iGoAR.GetAR(ACTION_SELL)
	err = am.CreateAccountReceivable(ar, data)
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
