package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/94peter/toad/resource/db"
)

type Amortization struct {
	AmorId                    string    `json:"id"`
	Branch                    string    `json:"branch"`
	Date                      time.Time `json:"date"`
	Itemname                  string    `json:"itemName"`
	Gaincost                  int       `json:"gainCost"`
	AmortizationYearLimit     int       `json:"amortizationYearLimit"`
	MonthlyAmortizationAmount int       `json:"monthlyAmortizationAmount"`
	FirstAmortizationAmount   int       `json:"firstAmortizationAmount"`
	Hasamortizationamount     int       `json:"hasAmortizationAmount"`
	Notamortizationamount     int       `json:"notAmortizationAmount"`
	IsOver                    bool      `json:"-"`
}

var (
	amorM *AmortizationModel
)

type AmortizationModel struct {
	imr              interModelRes
	db               db.InterSQLDB
	amortizationList []*Amortization
}

func GetAmortizationModel(imr interModelRes) *AmortizationModel {
	if amorM != nil {
		return amorM
	}

	amorM = &AmortizationModel{
		imr: imr,
	}
	return amorM
}

func (amorM *AmortizationModel) GetAmortizationData(beginDate, endDate, branch string) []*Amortization {

	const qspl = `SELECT amorid, branch, Date, itemname, gaincost, amortizationyearlimit, monthlyamortizationamount, firstamortizationamount, hasamortizationamount, notamortizationamount, isover
				FROM public.amortization
				where branch like '%s' and (Date >= '%s' and Date < ('%s'::date + '1 month'::interval))  ;`
	//const qspl = `SELECT arid,sales	FROM public.ar;`
	db := amorM.imr.GetSQLDB()
	rows, err := db.SQLCommand(fmt.Sprintf(qspl, branch, beginDate+"-01", endDate+"-01"))
	if err != nil {
		return nil
	}
	var amorDataList []*Amortization

	for rows.Next() {
		var amor Amortization

		// if err := rows.Scan(&r.ARid, &s); err != nil {
		// 	fmt.Println("err Scan " + err.Error())
		// }
		if err := rows.Scan(&amor.AmorId, &amor.Branch, &amor.Date, &amor.Itemname, &amor.Gaincost, &amor.AmortizationYearLimit, &amor.MonthlyAmortizationAmount, &amor.FirstAmortizationAmount, &amor.Hasamortizationamount, &amor.Notamortizationamount, &amor.IsOver); err != nil {
			fmt.Println("err Scan " + err.Error())
		}

		amorDataList = append(amorDataList, &amor)
	}

	// out, err := json.Marshal(arList)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(string(out))
	amorM.amortizationList = amorDataList
	return amorM.amortizationList
}

func (amorM *AmortizationModel) Json() ([]byte, error) {
	return json.Marshal(amorM.amortizationList)
}

func (amorM *AmortizationModel) DeleteAmortization(ID string) (err error) {
	fmt.Println("DeleteAmortization")
	const sql = `delete from public.Amortization where Amorid = $1;`

	interdb := amorM.imr.GetSQLDB()
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}
	fmt.Println("sqldb Exec")
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
	if id <= 0 {
		return errors.New("not found amortization")
	}
	return nil
}

func (amorM *AmortizationModel) CreateAmortization(amor *Amortization) (err error) {

	const sql = `INSERT INTO public.amortization
	(AmorId , branch, date, itemname, gaincost, amortizationyearlimit, monthlyamortizationamount, firstamortizationamount)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	;`

	interdb := amorM.imr.GetSQLDB()
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}

	fakeId := time.Now().Unix()

	res, err := sqldb.Exec(sql, fakeId, amor.Branch, time.Now(), amor.Itemname, amor.Gaincost, amor.AmortizationYearLimit, amor.MonthlyAmortizationAmount, amor.FirstAmortizationAmount)
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
		return errors.New("Invalid operation, CreateAmortization")
	}

	return nil
}