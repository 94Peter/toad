package model

import (
	"encoding/json"
	"errors"
	"fmt"

	"toad/resource/db"
)

type PModel struct {
	imr interModelRes
	db  db.InterSQLDB
}

var (
	pm  *PModel
	ADD = true
	DEL = false
)

func GetPModel(imr interModelRes) *PModel {
	if pm != nil {
		return pm
	}

	pm = &PModel{
		imr: imr,
	}
	return pm
}

//取得應收帳款內的sales清單
func GetSalesList(imr interModelRes, arid string) (saler []*Saler, err error) {
	const Ssql = `Select sales from public.ar where ar.arid = '%s'`
	interdb := imr.GetSQLDB()
	rows, err := interdb.SQLCommand(fmt.Sprintf(Ssql, arid))
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	var salerlist []*Saler
	for rows.Next() {
		var str string
		if err := rows.Scan(&str); err != nil {
			fmt.Println("err Scan " + err.Error())
		}
		//The origin saler data on database
		err := json.Unmarshal([]byte(str), &salerlist)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
	}

	return salerlist, nil
}

func UpdateSalerList(salerlist []*Saler, Type bool) (salerList []*Saler) {

	if Type {
		saler := &Saler{
			SName:   "testBname",
			Percent: 3,
			Sid:     "testBid",
		}
		salerlist = append(salerlist, saler)

	} else {

		i := 0
		for range salerlist {
			if salerlist[i].Sid == "testBid" {
				break
			}
			i++
		}
		// Remove the element at index i from a.
		copy(salerlist[i:], salerlist[i+1:])     // Shift a[i+1:] left one index.
		salerlist[len(salerlist)-1] = nil        // Erase last element (write zero value).
		salerlist = salerlist[:len(salerlist)-1] // Truncate slice.

	}

	return salerlist
}

//更新應收帳款的saler
func UpdateARSales(imr interModelRes, arid string, Type bool) (err error) {

	const Usql = `Update public.ar set sales = $1 where ar.arid = $2`
	interdb := imr.GetSQLDB()

	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}

	salerlist, err := GetSalesList(imr, arid)
	if err != nil {
		return err
	}

	salerlist = UpdateSalerList(salerlist, Type)

	out, err := json.Marshal(salerlist)
	if err != nil {
		fmt.Println(err)
		return errors.New("saler data failed")
	}

	fmt.Println(string(out))

	res, err := sqldb.Exec(Usql, string(out), arid)
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
		return errors.New("Operation error in UpdateARSales")
	}
	return nil
}

//建立 修改 刪除 收款單時，需要更改應收款項計算項目
func UpdateARInfo(imr interModelRes, arid string) (err error) {
	//https://stackoverflow.com/questions/2334712/how-do-i-update-from-a-select-in-sql-server
	const sql = `Update public.ar
				 set
					ra = t2.sum , balance = amount - fee -t2.sum
				FROM (
					 select sum(amount) from public.receipt where arid = $1 group by arid  
				)as t2 where ar.arid = $1`

	interdb := imr.GetSQLDB()
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
		const reset = `Update public.ar	set ra = 0 , balance = amount - fee , sales = '[]' where arid = $1`
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
