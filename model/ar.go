package model

import (
	"encoding/json"
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
	//CustomerType string    `json:"id"`
	CaseName string   `json:"caseName"`
	Amount   int      `json:"amount"`
	Fee      int      `json:"fee"`
	Balance  int      `json:"balance"`
	RA       int      `json:"receivedAmount"`
	Sales    []*Saler `json:"sales"`
}

type customer struct {
	Action string `json:"type"`
	Name   string `json:"name"`
}

type Saler struct {
	BName   string `json:"name"`
	Percent int    `json:"proportion"` //{"{\"BName\":\"123\",\"Bid\":\"13\",\"Persent\":12}","{\"BName\":\"123\",\"Bid\":\"13\",\"Persent\":12}"}
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

func (am *ARModel) GetData(today, end time.Time) []*AR {

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
	var arList []*AR

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

		arList = append(arList, &r)
	}

	// out, err := json.Marshal(arList)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(string(out))

	return arList
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

// func (a *Answer) GetTime() time.Time {
// 	return time.Date(a.Date.Year(), a.Date.Month(), a.Date.Day(), 0, 0, 0, 0, a.Date.Location())
// }

// func fetchSales(sale *sale) error {

// }
