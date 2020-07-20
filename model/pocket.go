package model

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"toad/pdf"
	"toad/resource/db"
	"toad/util"
)

type Pocket struct {
	Pid         string    `json:"id"`
	Date        time.Time `json:"date"`
	CircleID    string    `json:"-"` // for sql used
	ItemName    string    `json:"itemName"`
	Branch      string    `json:"branch"`
	Description string    `json:"description"`
	Income      int       `json:"income"`
	Fee         int       `json:"fee"`
	Balance     int       `json:"balance"`
}

var (
	pocketM *PocketModel
)

type PocketModel struct {
	imr        interModelRes
	db         db.InterSQLDB
	pocketList []*Pocket
}

func GetPocketModel(imr interModelRes) *PocketModel {
	if pocketM != nil {
		return pocketM
	}

	pocketM = &PocketModel{
		imr: imr,
	}
	return pocketM
}

func (pocketM *PocketModel) GetPocketData(beginDate, endDate, branch string) []*Pocket {

	const qspl = `SELECT Pid, Date, branch, itemname, description, income, fee, balance FROM public.pocket 
				where branch like '%s' and (Date >= '%s' and Date < ('%s'::date + '1 month'::interval)) 
				ORDER BY branch, date, pid asc;`

	db := pocketM.imr.GetSQLDB()
	rows, err := db.SQLCommand(fmt.Sprintf(qspl, branch, beginDate+"-01", endDate+"-01"))
	if err != nil {
		fmt.Println(err)
		return nil
	}
	var pocketDataList []*Pocket

	for rows.Next() {
		var pocket Pocket

		// if err := rows.Scan(&r.ARid, &s); err != nil {
		// 	fmt.Println("err Scan " + err.Error())
		// }
		if err := rows.Scan(&pocket.Pid, &pocket.Date, &pocket.Branch, &pocket.ItemName, &pocket.Description, &pocket.Income, &pocket.Fee, &pocket.Balance); err != nil {
			fmt.Println("err Scan " + err.Error())
		}

		pocketDataList = append(pocketDataList, &pocket)
	}

	// out, err := json.Marshal(arList)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(string(out))
	pocketM.pocketList = pocketDataList
	return pocketM.pocketList
}

func (pocketM *PocketModel) Json() ([]byte, error) {
	return json.Marshal(pocketM.pocketList)
}

func (pocketM *PocketModel) PDF() []byte {
	p := pdf.GetNewPDF(pdf.PageSizeA4)

	table := pdf.GetDataTable(pdf.Pocket)

	data, T_Income, T_Fee, T_Balance := pocketM.addInfoTable(table, p)
	//data, _, _, _ := amorM.addAmorInfoTable(table, p)

	p.CustomizedPocketTitle(data, "零用金")
	p.DrawTablePDF(data)
	p.CustomizedPocket(data, T_Income, T_Fee, T_Balance)
	return p.GetBytesPdf()

}

//介紹累加SQL
//http://www.blogjava.net/jxhkwhy/articles/200482.html  介紹重複問題
//https://codeday.me/bug/20180207/129828.html 介紹OVER (OVER也會遇到重複問題)
//SELECT pid,circleID , date, branch, itemname, description, income, fee, balance
// , sum(income-fee) OVER ( Order by pid asc) AS cum_amt
// FROM   public.pocket
// ORDER BY date asc ;
//
func (pocketM *PocketModel) DeletePocket(ID string) (err error) {

	const sql = `Delete from public.pocket	where Pid = $1 ;`

	interdb := pocketM.imr.GetSQLDB()
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
	fmt.Println(id)

	if id == 0 {
		return errors.New("Invalid operation, maybe not found the pocket")
	}
	pocketM.UpdatePocketBalance(sqldb)
	return nil
}
func (pocketM *PocketModel) AddorUpdatePocketMonthBalance(sqldb *sql.DB) (err error) {
	const sql = `INSERT INTO public.pocket(pid, circleid, date, branch, itemname, description, balance)
				select to_timestamp(p1.CircleID,'YYYY-MM')+ '1 month'::interval pid,
				to_char(to_timestamp(p1.CircleID,'YYYY-MM')+ '1 month'::interval,'YYYY-MM') CircleID,
				to_timestamp(p1.CircleID,'YYYY-MM')+ '1 month'::interval date,
				p1.branch, '' , '上期結餘', p2.balance
				from public.pocket p2
				inner join(
				select p1.circleid, MAX(p1.pid) as pid, p1.branch  from public.pocket p1 where description != '上期結餘' group by p1.circleid, p1.branch
				) p1 on p1.pid = p2.pid and p1.branch = p2.branch				
				ON CONFLICT (pid,branch) DO UPDATE SET balance = excluded.balance;`
	res, err := sqldb.Exec(sql)

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
		fmt.Println("AddPocketMonthBalance nothing to do ")
	}
	/*
	 *DELETE FROM table WHERE
	  table.id = (SELECT id FROM another_table);
	*/
	const delSQL = `DELETE FROM public.pocket
					WHERE pid = (
					select  p1.pid
					from public.pocket p1
					left join(
						select count(pid) count, to_char(to_timestamp(tmp.CircleID,'YYYY-MM')+ '1 month'::interval,'YYYY-MM') CircleID,
							branch from public.pocket tmp
						where tmp.description != '上期結餘'
						group by CircleID, branch
						) p2 on p1.CircleID = p2.CircleID and p1.branch = p2.branch
					where p1.description = '上期結餘' and p2.count is NULL);`
	res, err = sqldb.Exec(delSQL)

	if err != nil {
		fmt.Println(err)
		return err
	}
	id, err = res.RowsAffected()
	if err != nil {
		fmt.Println("PG Affecte Wrong(delSQL): ", err)
		return err
	}
	fmt.Println(id)

	if id == 0 {
		fmt.Println("delSQL nothing to do ")
	} else {
		fmt.Println("remove empty last month balance")
	}
	return nil
}

func (pocketM *PocketModel) UpdatePocketBalance(sqldb *sql.DB) (err error) {
	const sql = ` UPDATE public.pocket 
				SET balance = subquery.balance   
				FROM (
				SELECT pid as id,sum(income-fee) OVER (partition by branch Order by pid asc) AS balance
				FROM   public.pocket 
				where description != '上期結餘'
				ORDER BY pid asc
				) AS subquery
				WHERE pid=subquery.id;`

	res, err := sqldb.Exec(sql)

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
		fmt.Println("UpdatePocketBalance nothing to do ")
	}
	pocketM.AddorUpdatePocketMonthBalance(sqldb)
	return nil
}

func (pocketM *PocketModel) CreatePocket(pocket *Pocket) (err error) {

	const sql = `INSERT INTO public.pocket
	(Pid , date, CircleID, branch, itemname, description, income, fee)
	VALUES ($1, to_timestamp($2,'YYYY-MM-DD'), $3, $4, $5, $6, $7, $8)
	;`

	interdb := pocketM.imr.GetSQLDB()
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}
	t := time.Now()
	h := t.Hour()
	m := t.Minute()
	s := t.Second()
	n := t.Nanosecond()
	timein := pocket.Date.Add(time.Hour*time.Duration(h) +
		time.Minute*time.Duration(m) + time.Second*time.Duration(s) + time.Nanosecond*time.Duration(n))
	fakeId := timein.Unix()
	res, err := sqldb.Exec(sql, fakeId, pocket.Date, pocket.CircleID, pocket.Branch, pocket.ItemName, pocket.Description, pocket.Income, pocket.Fee)
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
		return errors.New("Invalid operation, CreatePocket")
	}
	pocketM.UpdatePocketBalance(sqldb)
	return nil
}

func (pocketM *PocketModel) UpdatePocket(ID string, pocket *Pocket) (err error) {
	a, e := json.Marshal(pocket)
	if e != nil {
		fmt.Println(e)
	}
	fmt.Println(string(a))
	const sql = `UPDATE public.pocket
				SET pid = $1 , date = to_timestamp($2,'YYYY-MM-DD'), branch=$3, itemname=$4, description=$5, circleid=$6, income=$7, fee=$8
				WHERE pid= $9` // and Branch = $3;`

	interdb := pocketM.imr.GetSQLDB()
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}
	t := time.Now()
	h := t.Hour()
	m := t.Minute()
	s := t.Second()
	n := t.Nanosecond()
	timein := pocket.Date.Add(time.Hour*time.Duration(h) +
		time.Minute*time.Duration(m) + time.Second*time.Duration(s) + time.Nanosecond*time.Duration(n))
	fakeId := timein.Unix()
	res, err := sqldb.Exec(sql, fakeId, pocket.Date, pocket.Branch, pocket.ItemName, pocket.Description, pocket.CircleID, pocket.Income, pocket.Fee, ID)
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
		return errors.New("Not found Pocket")
	}
	pocketM.UpdatePocketBalance(sqldb)
	return nil
}

// DELETE FROM table
// WHERE table.id = (SELECT id FROM another_table);

// DELETE FROM public.pocket
// WHERE pid = (
// select  p1.pid
// from public.pocket p1
// left join(
// select count(pid) count, to_char(to_timestamp(tmp.CircleID,'YYYY-MM')+ '1 month'::interval,'YYYY-MM') CircleID,
// 	branch from public.pocket tmp
// where tmp.description != '上期結餘'
// group by CircleID, branch
// ) p2 on p1.CircleID = p2.CircleID and p1.branch = p2.branch
// where p1.description = '上期結餘' and p2.count is NULL
// );

/*
select  p1.pid , p1.branch, p1.circleID, p1.description, p2.count
from public.pocket p1
left join(
select count(pid) count, to_char(to_timestamp(tmp.CircleID,'YYYY-MM')+ '1 month'::interval,'YYYY-MM') CircleID,
	branch from public.pocket tmp
where tmp.description != '上期結餘'
group by CircleID, branch
) p2 on p1.CircleID = p2.CircleID and p1.branch = p2.branch
where p1.description = '上期結餘' and p2.count is NULL
;*/

func (pocketM *PocketModel) addInfoTable(tabel *pdf.DataTable, p *pdf.Pdf) (tabel_final *pdf.DataTable,
	T_Income, T_Fee, T_Balance int) {

	//text := "fd"
	//width := mypdf.MeasureTextWidth(text)
	//tabel.ColumnLen
	T_Income, T_Fee, T_Balance = 0, 0, 0

	for _, element := range pocketM.pocketList {
		//
		text := element.Date.Format("2006-01-02")
		text, _ = util.ADtoROC(text, "ch")
		pdf.ResizeWidth(tabel, p.GetTextWidth(text), 0)
		vs := &pdf.TableStyle{
			Text:  text,
			Bg:    pdf.ColorWhite,
			Front: pdf.ColorTableLine,
		}
		tabel.RawData = append(tabel.RawData, vs)
		//
		text = element.Branch
		pdf.ResizeWidth(tabel, p.GetTextWidth(text), 1)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    pdf.ColorWhite,
			Front: pdf.ColorTableLine,
		}
		tabel.RawData = append(tabel.RawData, vs)
		//
		text = element.ItemName
		pdf.ResizeWidth(tabel, p.GetTextWidth(text), 2)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    pdf.ColorWhite,
			Front: pdf.ColorTableLine,
		}
		tabel.RawData = append(tabel.RawData, vs)
		//
		text = element.Description
		pdf.ResizeWidth(tabel, p.GetTextWidth(text), 3)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    pdf.ColorWhite,
			Front: pdf.ColorTableLine,
		}
		tabel.RawData = append(tabel.RawData, vs)
		//
		T_Income += element.Income
		text = pr.Sprintf("%d", element.Income)
		pdf.ResizeWidth(tabel, p.GetTextWidth(text), 4)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    pdf.ColorWhite,
			Front: pdf.ColorTableLine,
			Align: pdf.AlignRight,
		}
		tabel.RawData = append(tabel.RawData, vs)
		//
		T_Fee += element.Fee
		text = pr.Sprintf("%d", element.Fee)
		pdf.ResizeWidth(tabel, p.GetTextWidth(text), 5)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    pdf.ColorWhite,
			Front: pdf.ColorTableLine,
			Align: pdf.AlignRight,
		}
		tabel.RawData = append(tabel.RawData, vs)
		//
		T_Balance = element.Balance
		text = pr.Sprintf("%d", element.Balance)
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
