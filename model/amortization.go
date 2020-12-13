package model

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"toad/pdf"
	"toad/resource/db"
	"toad/util"
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

func (amorM *AmortizationModel) GetAmortizationData(beginDate, endDate time.Time, branch, dbname string) []*Amortization {

	const qspl = `SELECT amorid, branch, Date, itemname, gaincost, amortizationyearlimit, monthlyamortizationamount, firstamortizationamount, hasamortizationamount, notamortizationamount, isover
				FROM public.amortization
				where branch like '%s' and 
				extract(epoch from Date) >= '%d' and extract(epoch from Date) <= '%d'
				order by Date;`
	//(Date >= '%s' and Date < ('%s'::date + '1 month'::interval))
	//const qspl = `SELECT arid,sales	FROM public.ar;`
	//db := amorM.imr.GetSQLDB()
	db := amorM.imr.GetSQLDBwithDbname(dbname)

	rows, err := db.SQLCommand(fmt.Sprintf(qspl, branch, beginDate.Unix(), endDate.Unix()))
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

func (amorM *AmortizationModel) getAmortizationDataByAmorID(id, dbname string) *Amortization {

	const qspl = `SELECT amorid, Date FROM public.amortization where amorid = '%s';`
	//(Date >= '%s' and Date < ('%s'::date + '1 month'::interval))
	//const qspl = `SELECT arid,sales	FROM public.ar;`
	//db := amorM.imr.GetSQLDB()
	db := amorM.imr.GetSQLDBwithDbname(dbname)
	sql, err := db.ConnectSQLDB()
	rows, err := sql.Query(fmt.Sprintf(qspl, id))
	if err != nil {
		return nil
	}
	amor := &Amortization{}

	for rows.Next() {
		// if err := rows.Scan(&r.ARid, &s); err != nil {
		// 	fmt.Println("err Scan " + err.Error())
		// }
		if err := rows.Scan(&amor.AmorId, &amor.Date); err != nil {
			fmt.Println("err Scan " + err.Error())
		}
	}
	defer db.Close()
	return amor
}

func (amorM *AmortizationModel) DeleteAmortization(ID, dbname string) (err error) {
	fmt.Println("DeleteAmortization")
	interdb := amorM.imr.GetSQLDBwithDbname(dbname)
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}

	delamor := amorM.getAmortizationDataByAmorID(ID, dbname)
	if delamor.AmorId == "" {
		err = errors.New("not found amortization")
		return
	}
	_, err = salaryM.CheckValidCloseDate(delamor.Date, dbname, sqldb)
	if err != nil {
		return
	}

	const sql = `
				 delete from public.Amortization where Amorid = '%s';
				 delete from public.amormap where Amorid = '%s';				 
				 `

	del := fmt.Sprintf(sql, ID, ID)
	//interdb := amorM.imr.GetSQLDB()

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
	defer sqldb.Close()
	return nil
}

func (amorM *AmortizationModel) CreateAmortization(amor *Amortization, dbname string) (err error) {

	const sql = `INSERT INTO public.amortization
	(AmorId , branch, date, itemname, gaincost, amortizationyearlimit, monthlyamortizationamount, firstamortizationamount, notAmortizationAmount)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $5)
	;`

	interdb := amorM.imr.GetSQLDBwithDbname(dbname)
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}

	_, err = salaryM.CheckValidCloseDate(amor.Date, dbname, sqldb)
	if err != nil {
		return
	}

	fakeId := time.Now().Unix()
	amor.AmorId = fmt.Sprintf("%d", fakeId)
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
	amorM.InsertFirstAmortizationData(amor, sqldb)
	amorM.CreateAmorMap(sqldb)
	amorM.UpdateAmortizationData(sqldb)
	defer sqldb.Close()
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

func (amorM *AmortizationModel) CreateAmorMap(sqldb *sql.DB) (err error) {

	const sql = `INSERT INTO public.amormap	
	(amorid, date, cost)	
	select amorid,  to_char(mydates at time zone 'UTC' at time zone 'Asia/Taipei','yyyy-MM-dd') CircleID, monthlyamortizationamount from amortization a
	CROSS JOIN generate_series(a.date,  a.date+(a.amortizationyearlimit * 12  -1 || ' months')::interval, '1 months') AS mydates
	where notamortizationamount > 0  and mydates <= (date_trunc('month', now()) + interval '1 month' - interval '1 day')::date
	ON CONFLICT (amorid,date) DO NOTHING
	;`
	// const sql = `INSERT INTO public.amormap
	// (amorid, date, cost)
	// select amorid,  to_char(mydates,'YYYY-MM') CircleID, monthlyamortizationamount from amortization a
	// CROSS JOIN generate_series(a.date,  a.date+(a.amortizationyearlimit * 12  -1 || ' months')::interval, '1 months') AS mydates
	// where notamortizationamount > 0  and mydates <= (date_trunc('month', now()) + interval '1 month' - interval '1 day')::date
	// ON CONFLICT (amorid,date) DO NOTHING
	// ;`

	//ON CONFLICT (amorid,date) DO UPDATE SET cost = excluded.cost
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

func (amorM *AmortizationModel) InsertFirstAmortizationData(amor *Amortization, sqldb *sql.DB) (err error) {

	const sql = `INSERT INTO public.amormap
				(amorid, date, cost)
				values( $1, $2, $3)		
				ON CONFLICT (amorid,date) DO NOTHING		
				`

	loc, _ := time.LoadLocation("Asia/Taipei")
	t := amor.Date.In(loc).Format("2006-01-02")[0:10]

	res, err := sqldb.Exec(sql, amor.AmorId, t, amor.FirstAmortizationAmount)
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

	if id == 0 {
		fmt.Println("Invalid operation, InsertFirstAmortizationData")
		return errors.New("Invalid operation, InsertFirstAmortizationData")
	}

	return nil
}

func (amorM *AmortizationModel) UpdateAmortizationData(sqldb *sql.DB) (err error) {

	const sql = `UPDATE public.amortization t1
				SET  hasamortizationamount = t2.cost , notamortizationamount = t1.gaincost - t2.cost
				FROM (
					Select SUM(cost) as cost, amorid FROM public.amormap group by amorid
				)as t2 where t2.amorid = t1.amorid;`

	_, err = sqldb.Exec(sql)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func (amorM *AmortizationModel) RefreshAmortizationData(dbname string) {

	interdb := amorM.imr.GetSQLDBwithDbname(dbname)
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		fmt.Println(err)
		return
	}
	amorM.CreateAmorMap(sqldb)
	amorM.UpdateAmortizationData(sqldb)
	defer sqldb.Close()
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
			Bg:    pdf.ColorWhite,
			Front: pdf.ColorTableLine,
		}
		tabel.RawData = append(tabel.RawData, vs)
		//
		//text = element.Date.Format("2006-01-02 15:04:05")
		text = element.Date.Format("2006-01-02")
		text, _ = util.ADtoROC(text, "ch")
		pdf.ResizeWidth(tabel, p.GetTextWidth(text), 1)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    pdf.ColorWhite,
			Front: pdf.ColorTableLine,
		}
		tabel.RawData = append(tabel.RawData, vs)
		//
		text = pr.Sprintf("%d", element.Gaincost)
		pdf.ResizeWidth(tabel, p.GetTextWidth(text), 2)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    pdf.ColorWhite,
			Front: pdf.ColorTableLine,
			Align: pdf.AlignRight,
		}
		tabel.RawData = append(tabel.RawData, vs)
		//
		text = strconv.Itoa(element.AmortizationYearLimit)
		pdf.ResizeWidth(tabel, p.GetTextWidth(text), 3)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    pdf.ColorWhite,
			Front: pdf.ColorTableLine,
		}
		tabel.RawData = append(tabel.RawData, vs)
		//
		T_Month += element.MonthlyAmortizationAmount
		text = pr.Sprintf("%d", element.MonthlyAmortizationAmount)
		pdf.ResizeWidth(tabel, p.GetTextWidth(text), 4)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    pdf.ColorWhite,
			Front: pdf.ColorTableLine,
			Align: pdf.AlignRight,
		}
		tabel.RawData = append(tabel.RawData, vs)
		//
		T_Has += element.Hasamortizationamount
		text = pr.Sprintf("%d", element.Hasamortizationamount)
		pdf.ResizeWidth(tabel, p.GetTextWidth(text), 5)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    pdf.ColorWhite,
			Front: pdf.ColorTableLine,
			Align: pdf.AlignRight,
		}
		tabel.RawData = append(tabel.RawData, vs)
		//
		T_Not += element.Notamortizationamount
		text = pr.Sprintf("%d", element.Notamortizationamount)
		pdf.ResizeWidth(tabel, p.GetTextWidth(text), 6)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    pdf.ColorWhite,
			Front: pdf.ColorTableLine,
			Align: pdf.AlignRight,
		}
		tabel.RawData = append(tabel.RawData, vs)
	}

	tabel_final = tabel
	return
}
