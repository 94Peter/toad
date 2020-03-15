package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"dforcepro.com/report"
	"github.com/94peter/toad/pdf"
	"github.com/94peter/toad/resource/db"
	"github.com/94peter/toad/util"
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
	const sql = `
				 delete from public.Amortization where Amorid = '%s';
				 delete from public.amormap where Amorid = '%s';				 
				 `

	del := fmt.Sprintf(sql, ID, ID)
	interdb := amorM.imr.GetSQLDB()
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}
	fmt.Println("sqldb Exec")
	res, err := sqldb.Exec(del)
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
	(AmorId , branch, date, itemname, gaincost, amortizationyearlimit, monthlyamortizationamount, firstamortizationamount, notAmortizationAmount)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $5)
	;`

	interdb := amorM.imr.GetSQLDB()
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}

	fakeId := time.Now().Unix()

	res, err := sqldb.Exec(sql, fakeId, amor.Branch, amor.Date, amor.Itemname, amor.Gaincost, amor.AmortizationYearLimit, amor.MonthlyAmortizationAmount, amor.FirstAmortizationAmount)
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

	amorM.CreateAmorMap()
	amorM.UpdateAmortizationData()

	return nil
}

func (amorM *AmortizationModel) PDF() []byte {
	p := pdf.GetNewPDF()
	table := pdf.GetDataTable(pdf.Amortization)

	data, T_Month, T_Has, T_not := amorM.addAmorInfoTable(table, p)
	//data, _, _, _ := amorM.addAmorInfoTable(table, p)
	p.CustomizedAmortizationTitle(data, "成本攤提表")
	p.DrawTablePDF(data)
	p.CustomizedAmortization(data, T_Month, T_Has, T_not)
	return p.GetBytesPdf()
}

func (amorM *AmortizationModel) CreateAmorMap() (err error) {

	const sql = `INSERT INTO public.amormap
	(amorid, date, cost)
	SELECT amorid, t.CircleID, (Case When hasamortizationamount = 0 then firstamortizationamount else monthlyamortizationamount end) as cost
		FROM public.amortization a
	right join (
		select to_char(dates,'YYYY-MM') CircleID, dates from generate_series('2017-01-01'::timestamp, now(), '1 months') as gs(dates)	
	) t on t.dates >= to_char(a.date,'YYYY-MM-01')::timestamp
	where a.date is not null and notamortizationamount > 0
	ON CONFLICT (amorid,date) DO UPDATE SET cost = excluded.cost
	;`

	interdb := amorM.imr.GetSQLDB()
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}

	res, err := sqldb.Exec(sql)
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
		println("Invalid operation,CreateAmorMap")
		return errors.New("Invalid operation, CreateAmorMap")
	}

	return nil
}

func (amorM *AmortizationModel) UpdateAmortizationData() (err error) {

	const sql = `UPDATE public.amortization t1
				SET  hasamortizationamount = t2.cost , notamortizationamount = t1.gaincost - t2.cost
				FROM (
					Select SUM(cost) as cost, amorid FROM public.amormap group by amorid
				)as t2 where t2.amorid = t1.amorid;`

	interdb := amorM.imr.GetSQLDB()
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}
	_, err = sqldb.Exec(sql)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func (amorM *AmortizationModel) addAmorInfoTable(tabel *pdf.DataTable, p *pdf.Pdf) (tabel_final *pdf.DataTable,
	T_Month, T_Has, T_Not int) {

	//text := "fd"
	//width := mypdf.MeasureTextWidth(text)
	//tabel.ColumnLen
	T_Month, T_Has, T_Not = 0, 0, 0

	for _, element := range amorM.amortizationList {
		//
		text := element.Itemname
		pdf.ResizeWidth(tabel, p.GetTextWidth(text), 0)
		vs := &pdf.TableStyle{
			Text:  text,
			Bg:    report.ColorWhite,
			Front: report.ColorTableLine,
		}
		tabel.RawData = append(tabel.RawData, vs)
		//
		//text = element.Date.Format("2006-01-02 15:04:05")
		text = element.Date.Format("2006-01-02")
		text, _ = util.ADtoROC(text, "ch")
		pdf.ResizeWidth(tabel, p.GetTextWidth(text), 1)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    report.ColorWhite,
			Front: report.ColorTableLine,
		}
		tabel.RawData = append(tabel.RawData, vs)
		//
		text = strconv.Itoa(element.Gaincost)
		pdf.ResizeWidth(tabel, p.GetTextWidth(text), 2)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    report.ColorWhite,
			Front: report.ColorTableLine,
		}
		tabel.RawData = append(tabel.RawData, vs)
		//
		text = strconv.Itoa(element.AmortizationYearLimit)
		pdf.ResizeWidth(tabel, p.GetTextWidth(text), 3)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    report.ColorWhite,
			Front: report.ColorTableLine,
		}
		tabel.RawData = append(tabel.RawData, vs)
		//
		T_Month += element.MonthlyAmortizationAmount
		text = strconv.Itoa(element.MonthlyAmortizationAmount)
		pdf.ResizeWidth(tabel, p.GetTextWidth(text), 4)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    report.ColorWhite,
			Front: report.ColorTableLine,
		}
		tabel.RawData = append(tabel.RawData, vs)
		//
		T_Has += element.Hasamortizationamount
		text = strconv.Itoa(element.Hasamortizationamount)
		pdf.ResizeWidth(tabel, p.GetTextWidth(text), 5)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    report.ColorWhite,
			Front: report.ColorTableLine,
		}
		tabel.RawData = append(tabel.RawData, vs)
		//
		T_Not += element.Notamortizationamount
		text = strconv.Itoa(element.Notamortizationamount)
		pdf.ResizeWidth(tabel, p.GetTextWidth(text), 6)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    report.ColorWhite,
			Front: report.ColorTableLine,
		}
		tabel.RawData = append(tabel.RawData, vs)
	}

	tabel_final = tabel
	return
}
