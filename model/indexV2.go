package model

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"toad/resource/db"
)

type InfoV2 struct {
	Receivable  int            `json:"receivable"`      //累積應收
	Performance []*Performance `json:"performanceList"` //業績清單
}

type Performance struct {
	Date        string `json:"date"`
	Performance int    `json:"performance"` //本月業績
}

type IndexModelV2 struct {
	imr             interModelRes
	db              db.InterSQLDB
	info            *InfoV2
	incomeStatement *IncomeStatement
}

var (
	indexMV2 *IndexModelV2
)

func GetIndexModelV2(imr interModelRes) *IndexModelV2 {
	if indexMV2 != nil {
		return indexMV2
	}

	indexMV2 = &IndexModelV2{
		imr: imr,
	}
	return indexMV2
}

func (indexMV2 *IndexModelV2) Json(mtype string) ([]byte, error) {
	switch mtype {
	case "info":
		return json.Marshal(indexMV2.info)
	default:
		fmt.Println("unknown config type")
		break
	}
	return nil, nil
}

func (indexMV2 *IndexModelV2) GetInfoData(date time.Time, dbname string) {

	const sql_performance = `
		select COALESCE(sum(amount),0) Performance, to_char(mydates,'YYYY-MM') date from (
			select * from (
				select d + interval '1 month' - interval '1 second' - interval '8 hour' as mydates from generate_series('%s'::date, '%s'::date, '1 month') as d
			) dayT LEFT join(
				select * from public.receipt
			) r on r.date < dayT.mydates
		) data 
		group by mydates order by mydates asc`

	const sql = `SELECT  COALESCE(SUM(ar.amount),0) amount,      
	COALESCE(SUM(COALESCE((SELECT SUM(r.amount) FROM public.receipt r WHERE ar.arid = r.arid),0)),0) AS RA ,
	COALESCE( (SELECT sum(fee) from public.deduct),0) deduct 
	FROM public.ar ar `
	db := indexMV2.imr.GetSQLDBwithDbname(dbname)
	sqldb, err := db.ConnectSQLDB()
	defer sqldb.Close()
	//rows, err := db.SQLCommand(fmt.Sprintf(sql, date.Unix(), date.Unix()))
	rows, err := sqldb.Query(fmt.Sprintf(sql))
	if err != nil {
		fmt.Println("GetInfoData stepV1:", err)
		return
	}
	var data *InfoV2

	for rows.Next() {
		var md InfoV2
		var Amount, SUM_RA, SUM_Deduct int

		if err := rows.Scan(&Amount, &SUM_RA, &SUM_Deduct); err != nil {
			fmt.Println("err Scan " + err.Error())
		}

		md.Receivable = (Amount - SUM_RA - SUM_Deduct)
		data = &md
	}

	loc, _ := time.LoadLocation("Asia/Taipei")
	t := date.In(loc)
	//t := date
	y, _, _ := t.Date()
	year := strconv.Itoa(y)
	rows, err = sqldb.Query(fmt.Sprintf(sql_performance, year+"-01-01", year+"-12-01"))
	if err != nil {
		fmt.Println(err)
		return
	}
	var pList []*Performance
	for rows.Next() {
		var p Performance

		if err := rows.Scan(&p.Performance, &p.Date); err != nil {
			fmt.Println("GetInfoData stepV2 err Scan " + err.Error())
		}

		pList = append(pList, &p)

	}
	data.Performance = pList
	// out, err := json.Marshal(data)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(string(out))

	indexMV2.info = data

}
