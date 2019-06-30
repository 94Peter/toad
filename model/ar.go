package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/94peter/toad/resource/db"
)

//`json:"id"` 回傳重新命名
type AR struct {
	ARid     string    `json:"id"`
	Date     time.Time `json:"completionDate"`
	CNo      string    `json:"contractNo"`
	Customer customer  `json:"customer"`
	CaseName string    `json:"caseName"`
	Amount   int       `json:"amount"`
	Fee      int       `json:"fee"`
	Balance  int       `json:"balance"`        //未收金額
	RA       int       `json:"receivedAmount"` //已收金額
	Sales    []*Saler  `json:"sales"`
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

type customer struct {
	Action string `json:"type"`
	Name   string `json:"name"`
}

type Saler struct {
	BName   string  `json:"name"`
	Percent float64 `json:"proportion"` //{"{\"BName\":\"123\",\"Bid\":\"13\",\"Persent\":12}","{\"BName\":\"123\",\"Bid\":\"13\",\"Persent\":12}"}
	//Bid     string  `json:"Bid"`
}

type AccountReceivable struct {
	db db.InterSQLDB
	//res interModelRes
	ar []*AR
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
	imr    interModelRes
	db     db.InterSQLDB
	arList []*AR
}

func (am *ARModel) GetARData(today, end time.Time) []*AR {

	//ri := GetARModel(ar.imr)
	//const qtpl = `SELECT arid, date, cno, "caseName", type, name, amount, fee, ra, balance, sales	FROM public.ar;`
	//const qtpl = `SELECT arid	FROM public.ar;`
	const qspl = `SELECT arid, date, cno, casename, type, name, amount, fee, ra, balance, sales	FROM public.ar;`
	//const qspl = `SELECT arid,sales	FROM public.ar;`
	db := am.imr.GetSQLDB()
	rows, err := db.SQLCommand(fmt.Sprintf(qspl))
	if err != nil {
		return nil
	}
	var arDataList []*AR

	for rows.Next() {
		var r AR

		var ctm customer
		var col_sales string
		// if err := rows.Scan(&r.ARid, &s); err != nil {
		// 	fmt.Println("err Scan " + err.Error())
		// }
		if err := rows.Scan(&r.ARid, &r.Date, &r.CNo, &r.CaseName, &ctm.Action, &ctm.Name, &r.Amount, &r.Fee, &r.RA, &r.Balance, &col_sales); err != nil {
			fmt.Println("err Scan " + err.Error())
		}
		r.Customer = ctm

		err := json.Unmarshal([]byte(col_sales), &r.Sales)
		if err != nil {
			fmt.Println(err)
		}

		arDataList = append(arDataList, &r)
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

func (am *ARModel) Json() ([]byte, error) {
	return json.Marshal(am.arList)
}

func (am *ARModel) CreateAccountReceivable(receivable *AR) (err error) {
	fmt.Println("CreateAccountReceivable")

	const sql = `INSERT INTO public.ar(
		arid, date, cno, casename, type, name, amount, fee ,sales)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);`

	interdb := am.imr.GetSQLDB()
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}
	fmt.Println("sqldb Exec")

	out, err := json.Marshal(receivable)
	if err != nil {
		panic(err)
	}
	fmt.Println("receivable data:", string(out))
	salers, err := json.Marshal(receivable.Sales)
	if err != nil {
		panic(err)
	}
	fmt.Println("salers data:", string(salers))

	t := time.Now().Unix()
	//unix_time := time.Time(receivable.CompletionDate).Unix()

	out, err2 := json.Marshal(receivable.Sales)
	if err2 != nil {
		panic(err)
	}
	fmt.Println(string(out))
	fmt.Println(receivable.Fee)

	res, err := sqldb.Exec(sql, t, receivable.Date, receivable.CNo, receivable.CaseName, receivable.Customer.Action, receivable.Customer.Name, receivable.Amount, receivable.Fee, string(salers))
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

	return nil
}
func (am *ARModel) CreateReceipt(rt *Receipt) (err error) {
	fmt.Println("CreateReceipt")

	//balance <= Amount 未收金額<=繳款金額
	const sql = `INSERT INTO public.receipt
	(Rid, date, cno, casename, type, name, amount, ARid)
	select $1, $2, cno, casename, type, name, $3, arid
	from public.ar where arid = $4 and balance >= $3;`

	interdb := am.imr.GetSQLDB()
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

	err = am.updateARInfo(rt.ARid)
	if err != nil {
		return err
	}

	return nil
}

//建立 修改 刪除 收款單時，需要更改應收款項計算項目
func (am *ARModel) updateARInfo(arid string) (err error) {
	//https://stackoverflow.com/questions/2334712/how-do-i-update-from-a-select-in-sql-server
	const sql = `Update public.ar
				 set
					ra = t2.sum , balance = amount - fee -t2.sum
				FROM (
					 select sum(amount) from public.receipt where arid = $1 group by arid  
				)as t2 where
				ar.arid = $1`

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

	fmt.Println(id)
	return nil
}

// func (a *Answer) GetTime() time.Time {
// 	return time.Date(a.Date.Year(), a.Date.Month(), a.Date.Day(), 0, 0, 0, 0, a.Date.Location())
// }

// func fetchSales(sale *sale) error {

// }
