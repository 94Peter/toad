package model

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/94peter/toad/resource/db"
)

type AR struct {
	ARid         string
	CNo          string
	CaseName     string
	CustomerType string
	Name         string
	Amount       int
	Fee          int
	Balance      int
	RA           int
	Sales        []*sale
	Date         time.Time
}

type sale struct {
	bName   string
	percent float64
	Bid     string
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

func (ar *ARModel) GetData(today, end time.Time) []*AR {

	//ri := GetARModel(ar.imr)
	//const qtpl = `SELECT arid, date, cno, "caseName", type, name, amount, fee, ra, balance, sales	FROM public.ar;`
	const qtpl = `SELECT arid, date, cno, casename, type, name, amount, fee, ra, balance, sales	FROM public.ar;`
	//const qtpl = `SELECT arid	FROM public.ar;`
	db := ar.imr.GetSQLDB()
	rows, err := db.Query(fmt.Sprintf(qtpl))
	if err != nil {
		return nil
	}
	var arList []*AR

	for rows.Next() {
		var r AR
		if err := rows.Scan(&r.ARid, &r.Date, &r.CNo, &r.CaseName, &r.CustomerType, &r.Name, &r.Amount, &r.Fee, &r.RA, &r.Balance, &r.Sales); err != nil {
			fmt.Println("err Scan " + err.Error())
		}
		fmt.Println(r.ARid)

		out, err := json.Marshal(r)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(out))
		arList = append(arList, &r)
	}

	out, err := json.Marshal(arList)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(out))

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
