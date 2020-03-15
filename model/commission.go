package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"dforcepro.com/report"
	"github.com/94peter/toad/pdf"
	"github.com/94peter/toad/resource/db"
)

type CModel struct {
	imr   interModelRes
	db    db.InterSQLDB
	cList []*Commission
}

var (
	cm *CModel
)

type Commission struct {
	Sid      string    `json:"sid"` //no return this key 業務人員ID
	Rid      string    `json:"rid"` //收據ID
	Date     time.Time `json:"date"`
	Item     string    `json:"item"`    //合約+案名+買賣方
	Amount   int       `json:"amount"`  //金額
	Fee      int       `json:"fee"`     //扣款金額收款
	SR       float64   `json:"sr"`      //實績 sales report or sales records
	Bonus    float64   `json:"bonus"`   //獎金
	CPercent float64   `json:"percent"` //傭金比例(Now)
	SName    string    `json:"name"`    //業務姓名
	ARid     string    `json:"-"`       //程式內部比對用
	Status   string    `json:"status"`  // normal join remove
	//only PDF used
	Branch      string `json:"-"` // normal join remove
	Percent     string `json:"-"` //獎金比例
	InvoiceNo   string `json:"-"` //發票號碼
	ReceiveDate string `json:"-"` //收據 入帳日期
	Checknumber string `json:"-"` //票號
	DedectItem  string `json:"-"` //pdf 備註 >> dedeuct的Item

}

func GetCModel(imr interModelRes) *CModel {
	if cm != nil {
		return cm
	}

	cm = &CModel{
		imr: imr,
	}
	return cm
}

func (cm *CModel) ExportCommissiontData(data []*Commission) []*Commission {
	fmt.Println("exportCommissiontData")
	//if invoiceno is null in Database return ""

	const qsql = `SELECT c.sid, c.rid, r.date, c.item, r.amount, 0 , c.sname, c.cpercent, ( (r.amount - coalesce(d.fee,0) )* c.cpercent/100) sr, ( (r.amount - coalesce(d.fee,0) )* c.cpercent/100 * cs.percent/100) bonus,
	r.arid, c.status , cs.branch, cs.percent, to_char(r.date,'yyyy-MM-dd') , COALESCE(NULLIF(r.invoiceno, null),'') , coalesce(d.checknumber,'') , coalesce(d.fee,0) , coalesce(d.item,'')
	FROM public.commission c
	inner JOIN public.receipt r on r.rid = c.rid
	Inner Join (
			SELECT A.sid, A.branch, A.percent, A.title
				FROM public.ConfigSaler A
				Inner Join (
					select sid, max(zerodate) zerodate from public.configsaler cs
					where now() > zerodate
					group by sid
				) B on A.sid=B.sid and A.zeroDate = B.zeroDate
		) cs on c.sid=cs.sid
	left join(
		select rid, checknumber , fee, item from public.deduct
	) d on d.rid = r.rid
	where c.rid = '%s' and c.sid ='%s';`

	//left JOIN (select sum(fee) fee, count(rid) ,arid from public.deduct group by arid) as tmp on tmp.arid = r.arid
	db := cm.imr.GetSQLDB()
	var cDataList []*Commission
	for _, element := range data {
		fmt.Println(fmt.Sprintf(qsql, element.Rid, element.Sid))
		rows, err := db.SQLCommand(fmt.Sprintf(qsql, element.Rid, element.Sid))
		if err != nil {
			fmt.Println(err)
			return nil
		}

		fmt.Println("SQLCommand Done")
		for rows.Next() {
			fmt.Println("rows.Next ")
			var c Commission

			if err := rows.Scan(&c.Sid, &c.Rid, &c.Date, &c.Item, &c.Amount, &c.Fee, &c.SName, &c.CPercent, &c.SR, &c.Bonus, &c.ARid, &c.Status, &c.Branch, &c.Percent, &c.ReceiveDate, &c.InvoiceNo, &c.Checknumber, &c.Fee, &c.DedectItem); err != nil {
				fmt.Println("err Scan " + err.Error())
			}

			out2, _ := json.Marshal(c)
			fmt.Println("c :", string(out2))

			cDataList = append(cDataList, &c)
		}

	}
	cm.cList = cDataList
	out, _ := json.Marshal(cm.cList)
	fmt.Println("cm.cList :", string(out))

	return cm.cList
}

func (cm *CModel) ExportCommissiontDataByBSid(bsid string) []*Commission {
	fmt.Println("exportCommissiontData")
	//if invoiceno is null in Database return ""

	const qsql = `SELECT c.sid, c.rid, r.date, c.item, r.amount, 0 , c.sname, c.cpercent, ( (r.amount - coalesce(d.fee,0) )* c.cpercent/100) sr, ( (r.amount - coalesce(d.fee,0) )* c.cpercent/100 * cs.percent/100) bonus,
	r.arid, c.status , cs.branch, cs.percent, to_char(r.date,'yyyy-MM-dd') , COALESCE(NULLIF(iv.invoiceno, null),'') , coalesce(d.checknumber,'') , coalesce(d.fee,0) , coalesce(d.item,'')
	FROM public.commission c
	inner JOIN public.receipt r on r.rid = c.rid
	Inner Join (
			SELECT A.sid, A.branch, A.percent, A.title
				FROM public.ConfigSaler A
				Inner Join (
					select sid, max(zerodate) zerodate from public.configsaler cs
					where now() > zerodate
					group by sid
				) B on A.sid=B.sid and A.zeroDate = B.zeroDate
		) cs on c.sid=cs.sid
	left join(
		select rid, checknumber , fee, item from public.deduct
	) d on d.rid = r.rid
	left join(
		select rid,  invoiceno from public.invoice 
	) iv on r.rid = iv.rid
	where c.bsid = '%s' order by c.arid asc;` //根據案子分類

	//left JOIN (select sum(fee) fee, count(rid) ,arid from public.deduct group by arid) as tmp on tmp.arid = r.arid
	db := cm.imr.GetSQLDB()
	var cDataList []*Commission

	rows, err := db.SQLCommand(fmt.Sprintf(qsql, bsid))
	if err != nil {
		fmt.Println(err)
		return nil
	}

	for rows.Next() {

		var c Commission

		if err := rows.Scan(&c.Sid, &c.Rid, &c.Date, &c.Item, &c.Amount, &c.Fee, &c.SName, &c.CPercent, &c.SR, &c.Bonus, &c.ARid, &c.Status, &c.Branch, &c.Percent, &c.ReceiveDate, &c.InvoiceNo, &c.Checknumber, &c.Fee, &c.DedectItem); err != nil {
			fmt.Println("err Scan " + err.Error())
		}

		out2, _ := json.Marshal(c)
		fmt.Println("exportCommissiontData c :", string(out2))

		cDataList = append(cDataList, &c)
	}

	cm.cList = cDataList
	out, _ := json.Marshal(cm.cList)
	fmt.Println("exportCommissiontData cm.cList :", string(out))

	return cm.cList
}

func (cm *CModel) GetCommissiontData(start, end, status string) []*Commission {
	fmt.Println("GetCommissiontData")
	//if invoiceno is null in Database return ""

	const qsql = `SELECT c.sid, c.rid, r.date, c.item, r.amount, c.fee , c.sname, c.cpercent, c.sr, c.bonus, r.arid, c.status
				FROM public.commission c
				inner JOIN public.receipt r on r.rid = c.rid
				where to_timestamp(date_part('epoch',r.date)::int) >= '%s' and to_timestamp(date_part('epoch',r.date)::int) < '%s'::date + '1 month'::interval
				and c.status like '%s';`

	// const qsql = `SELECT c.sid, c.rid, r.date, c.item, r.amount, 0 , c.sname, c.cpercent, ( (r.amount - d.fee)* c.cpercent/100) sr, ( (r.amount - d.fee)* c.cpercent/100 * cs.percent/100) bonus,
	// r.arid, c.status , cs.branch, cs.percent, to_char(r.date,'yyyy-MM-dd') , COALESCE(NULLIF(r.invoiceno, null),'') , d.checknumber , d.fee , d.item
	// FROM public.commission c
	// inner JOIN public.receipt r on r.rid = c.rid
	// Inner Join (
	// 		SELECT A.sid, A.branch, A.percent, A.title
	// 			FROM public.ConfigSaler A
	// 			Inner Join (
	// 				select sid, max(zerodate) zerodate from public.configsaler cs
	// 				where now() > zerodate
	// 				group by sid
	// 			) B on A.sid=B.sid and A.zeroDate = B.zeroDate
	// 	) cs on c.sid=cs.sid
	// inner join(
	// 	select rid, checknumber , fee, item from public.deduct
	// ) d on d.rid = r.rid
	// where to_timestamp(date_part('epoch',r.date)::int) >= '%s' and to_timestamp(date_part('epoch',r.date)::int) < '%s'::date + '1 month'::interval
	// and c.status like '%s';`

	//left JOIN (select sum(fee) fee, count(rid) ,arid from public.deduct group by arid) as tmp on tmp.arid = r.arid
	db := cm.imr.GetSQLDB()
	rows, err := db.SQLCommand(fmt.Sprintf(qsql, start, end, status))
	if err != nil {
		fmt.Println(err)
		return nil
	}
	var cDataList []*Commission
	fmt.Println("SQLCommand Done")
	for rows.Next() {
		var c Commission

		// if err := rows.Scan(&r.ARid, &s); err != nil {
		// 	fmt.Println("err Scan " + err.Error())
		// }
		if err := rows.Scan(&c.Sid, &c.Rid, &c.Date, &c.Item, &c.Amount, &c.Fee, &c.SName, &c.CPercent, &c.SR, &c.Bonus, &c.ARid, &c.Status); err != nil {
			fmt.Println("err Scan " + err.Error())
		}
		// if err := rows.Scan(&c.Sid, &c.Rid, &c.Date, &c.Item, &c.Amount, &c.Fee, &c.SName, &c.CPercent, &c.SR, &c.Bonus, &c.ARid, &c.Status, &c.Branch, &c.Percent, &c.ReceiveDate, &c.InvoiceNo, &c.Checknumber, &c.Fee, &c.DedectItem); err != nil {
		// 	fmt.Println("err Scan " + err.Error())
		// }

		cDataList = append(cDataList, &c)
	}

	// const feesql = `select sum(fee) fee, arid from public.deduct group by arid;`
	// rows, err = db.SQLCommand(fmt.Sprintf(feesql))
	// if err != nil {
	// 	fmt.Println(err)
	// 	return nil
	// }
	// for rows.Next() {
	// 	var arid string
	// 	var fee int
	// 	Rid := ""
	// 	// if err := rows.Scan(&r.ARid, &s); err != nil {
	// 	// 	fmt.Println("err Scan " + err.Error())
	// 	// }
	// 	if err := rows.Scan(&fee, &arid); err != nil {
	// 		fmt.Println("err Scan " + err.Error())
	// 	}
	// 	for _, commission := range cDataList {
	// 		if Rid == commission.Rid || Rid == "" {
	// 			//fmt.Println("arid "+arid+","+commission.ARid+" fee ", fee)
	// 			if commission.ARid == arid {
	// 				commission.Fee = fee
	// 				Rid = commission.Rid
	// 				commission.SR = float64(commission.Amount-fee) * commission.CPercent / 100
	// 				commission.Bonus = commission.Bonus * float64(commission.Amount-fee) / float64(commission.Amount-0)
	// 			}
	// 		} else {
	// 			break
	// 		}
	// 	}
	// }

	cm.cList = cDataList
	return cm.cList
}

//to_timestamp(SS.date, 'YYYY-MM')
func (cm *CModel) Json() ([]byte, error) {
	return json.Marshal(cm.cList)
}

func (cm *CModel) GetBytePDF() []byte {
	p := pdf.GetOriPDF()
	data := p.GetBytesPdf()
	p = pdf.GetNewPDF()
	return data
}

func (cm *CModel) PDF(isNew bool) {
	// var p *pdf.Pdf
	// if isNew {
	// 	p = pdf.GetNewPDF()
	// } else {
	// 	p = pdf.GetOriPDF()
	// }

	p := pdf.GetOriPDF()

	tabel := pdf.GetDataTable(pdf.Commission)
	data, SR, Bonus := cm.addDataIntoTable(tabel, p)
	//p.DrawPDF(pdf.GetDataTable(""))
	p.DrawTablePDF(data)
	//init PDFX is 10
	pdfx := 10.0
	textw := 0.0
	//5 is 姓名欄位
	for i := 0; i < 5; i++ {
		textw += tabel.ColumnWidth[i]
	}
	pdfx += textw
	BranchName := "此薪資表無傭金"
	if len(cm.cList) > 0 {
		BranchName = cm.cList[0].Branch
	}
	p.DrawRectangle(textw, pdf.TextHeight, report.ColorWhite, "FD")
	p.FillText(BranchName, 12, report.ColorTableLine, alignCenter, valignMiddle, textw, pdf.TextHeight)

	p.SetPdf_XY(pdfx, -1)
	p.DrawRectangle(tabel.ColumnWidth[5]+tabel.ColumnWidth[6], pdf.TextHeight, report.ColorWhite, "FD")
	p.FillText("合計金額", 12, report.ColorTableLine, alignCenter, valignMiddle, tabel.ColumnWidth[5]+tabel.ColumnWidth[6], pdf.TextHeight)
	p.SetPdf_XY(pdfx+tabel.ColumnWidth[5]+tabel.ColumnWidth[6], -1)
	p.DrawRectangle(tabel.ColumnWidth[7], pdf.TextHeight, report.ColorWhite, "FD")
	p.FillText(fmt.Sprintf("%.f", SR), 12, report.ColorTableLine, alignCenter, valignMiddle, tabel.ColumnWidth[7], pdf.TextHeight)
	p.SetPdf_XY(pdfx+tabel.ColumnWidth[5]+tabel.ColumnWidth[6]+tabel.ColumnWidth[7], -1)
	p.DrawRectangle(tabel.ColumnWidth[8], pdf.TextHeight, report.ColorWhite, "FD")
	p.FillText(fmt.Sprintf("%.f", Bonus), 12, report.ColorTableLine, alignCenter, valignMiddle, tabel.ColumnWidth[8], pdf.TextHeight)

	p.NewLine(25)
	p.NewLine(25) //空一行

	fmt.Println(SR, ":", Bonus)
	return //p.GetBytesPdf()  這邊使用GetBytesPdf 會莫名其妙多一頁面
}

func (cm *CModel) CreateCommission(rt *Receipt) (err error) {
	/**
		預防薪資錯誤，若收款日期當月已建立薪資表，則自動將此傭金編入remove。#但不行，會造成傭金錯誤。
		or armap.sid = cs.identityNum 條件查詢新增，因為住通串接，他們帶入的可能是身分證，本來使用的sid是電話號碼。

	**/
	const sql = `INSERT INTO public.commission
	(Sid, Rid, Item, SName, CPercent, sr, bonus , arid)
	select armap.sid, $1, ar.cno ||' '|| ar.casename ||' '|| (Case When AR.type = 'buy' then '買' When AR.type = 'sell' then '賣' else 'unknown' End ), armap.sname,
	armap.proportion, $2 * armap.proportion / 100 ,  $2 * armap.proportion / 100 * cs.percent /100 , $3::VARCHAR
	from public.ar ar
	inner join 	public.armap armap on armap.arid = ar.arid 
	inner join 	(			
			select cs.branch, cs.sid, cs.percent, cs.identitynum from public.configsaler cs 
			inner join (
				select sid, max(zerodate) zerodate from public.configsaler cs 
				where now() > zerodate
				group by sid
			) tmp on tmp.sid = cs.sid and tmp.zerodate = cs.zerodate		
		)	cs  on cs.sid = armap.sid or armap.sid = cs.identityNum
	where ar.arid = $3;`
	// const sql = `INSERT INTO public.commission
	// (Sid, Rid, Item, SName, CPercent, sr, bonus , arid, status)
	// select armap.sid, $1, ar.cno ||' '|| ar.casename ||' '|| (Case When AR.type = 'buy' then '買' When AR.type = 'sell' then '賣' else 'unknown' End ), armap.sname, armap.proportion, $2 * armap.proportion / 100 ,  $2 * armap.proportion / 100 * cs.percent /100 , $3::VARCHAR,
	// ( case when tmp.branch is NULL then 'normal' else 'remove' end) status
	// from public.ar ar
	// inner join 	public.armap armap on armap.arid = ar.arid
	// inner join 	(
	// 		select cs.branch, cs.sid, cs.percent from public.configsaler cs
	// 		inner join (
	// 			select sid, max(zerodate) zerodate from public.configsaler cs
	// 			where now() > zerodate
	// 			group by sid
	// 		) tmp on tmp.sid = cs.sid and tmp.zerodate = cs.zerodate
	// 	)	cs  on cs.sid = armap.sid
	// left join (
	// 	select branch from public.branchsalary BS
	// 	where BS.date = to_char($4::date ,'YYYY-MM')::varchar(7)
	// ) tmp on tmp.branch = cs.branch
	// where ar.arid = $3;`
	// select armap.sid, $1, ar.cno ||' '|| ar.casename ||' '|| ar.type, armap.sname, armap.proportion
	// from public.ar ar
	// inner join 	public.armap armap on armap.arid = ar.arid
	// where ar.arid = $2 ;`

	interdb := cm.imr.GetSQLDB()
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}
	fmt.Println("CreateCommission Rid:", rt.Rid)
	//fmt.Println(rt.Amount)
	fmt.Println("CreateCommission ARid:", rt.ARid)
	res, err := sqldb.Exec(sql, rt.Rid, rt.Amount, rt.ARid)
	//res, err := sqldb.Exec(sql, unix_time, receivable.Date, receivable.CNo, receivable.Sales)
	if err != nil {
		fmt.Println("CreateCommission:", err)
		return err
	}
	id, err := res.RowsAffected()
	if err != nil {
		fmt.Println("PG Affecte Wrong: ", err)
		return err
	}
	fmt.Println(id)

	if id == 0 {
		return errors.New("Invalid operation, CreateCommission")
	}

	return nil
}

func (cm *CModel) UpdateCommission(com *Commission, rid, sid string) (err error) {
	/*
	 * 更新原則: 比例換算，新值= 舊值 / 舊的比例 * 新的比例
	 * 獎金比例使用舊的。
	 */
	const sql = `UPDATE public.commission
		SET cpercent= $1::double precision, sr= sr / cpercent * $1::double precision , bonus = bonus * $1::double precision / cpercent
		WHERE sid= $3 and rid= $2 ;`

	interdb := cm.imr.GetSQLDB()
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}
	//res, err := sqldb.Exec(sql)
	res, err := sqldb.Exec(sql, com.CPercent, rid, sid)
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
		return errors.New("[UpdateCommission] Invalid operation, maybe not found commission")
	}

	return nil
}

func (cm *CModel) RefreshCommissionBonus(Sid, Rid, mtype string) (err error) {
	if strings.ToLower(mtype) == "all" {
		Rid = "%"
	}

	const sql = `Update public.commission t1
					set sr = (t2.amount - t2.fee) * t2.cpercent / 100 , bonus = (t2.amount - t2.fee) * t2.cpercent / 100 * t2.percent /100
				FROM(
				SELECT c.bsid, c.sid, c.rid, r.amount, c.fee , c.cpercent, c.sr, c.bonus,  cs.percent
								FROM public.commission c
								inner JOIN public.receipt r on r.rid = c.rid				
								inner join 	(			
									select cs.sid, cs.percent from public.configsaler cs 
									inner join (
										select sid, max(zerodate) zerodate from public.configsaler cs 
										where now() > zerodate
										group by sid
									) tmp on tmp.sid = cs.sid and tmp.zerodate = cs.zerodate		
								)	cs  on cs.sid = c.sid
								WHERE c.bsid is null
				) as t2 where t1.sid = t2.sid and t1.rid = t2.rid and t1.sid = $1 and t1.rid like $2`
	interdb := cm.imr.GetSQLDB()
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}

	res, err := sqldb.Exec(sql, Sid, Rid)
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

func (cm *CModel) UpdateCommissionStatus(rid, sid string) (err error) {
	/*
	 * 更新原則: 比例換算，新值= 舊值 / 舊的比例 * 新的比例
	 * 獎金比例使用舊的。
	 */
	const sql = `UPDATE public.commission
		SET status = 'remove'
		WHERE sid= $2 and rid= $1 and bsid is null;`

	interdb := cm.imr.GetSQLDB()
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}
	//res, err := sqldb.Exec(sql)
	res, err := sqldb.Exec(sql, rid, sid)
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
		return errors.New("[UpdateCommission] Invalid operation, maybe not found commission or already bind salary")
	}

	return nil
}

func (cm *CModel) addDataIntoTable(tabel *pdf.DataTable, p *pdf.Pdf) (*pdf.DataTable, float64, float64) {

	var TotalSR = 0.0
	var TotalBouns = 0.0
	//text := "fd"
	//width := mypdf.MeasureTextWidth(text)
	//tabel.ColumnLen
	var transactionID = ""
	var sameRow = true
	for index, element := range cm.cList {
		if transactionID == element.Rid {
			sameRow = true
		} else {
			transactionID = element.Rid
			sameRow = false
		}

		fmt.Println(index)
		fmt.Println(tabel.ColumnWidth[index])
		text := ""
		//放空白
		if sameRow {
			vs := &pdf.TableStyle{
				Text:  text,
				Bg:    report.ColorWhite,
				Front: report.ColorTableLine,
			}
			tabel.RawData = append(tabel.RawData, vs)
			tabel.RawData = append(tabel.RawData, vs)
			tabel.RawData = append(tabel.RawData, vs)
			tabel.RawData = append(tabel.RawData, vs)
			tabel.RawData = append(tabel.RawData, vs)
		} else {
			/// 西元轉民國
			text := element.ReceiveDate
			TWyear, _ := strconv.Atoi(text[0:4])
			fmt.Println(TWyear)
			TWyear = TWyear - 1911
			TW_Date := fmt.Sprintf("%d/%s/%s", TWyear, text[5:7], text[8:10])
			pdf.ResizeWidth(tabel, p.GetTextWidth(text), 0)
			vs := &pdf.TableStyle{
				Text:  TW_Date,
				Bg:    report.ColorWhite,
				Front: report.ColorTableLine,
			}
			tabel.RawData = append(tabel.RawData, vs)

			//
			text = element.InvoiceNo
			pdf.ResizeWidth(tabel, p.GetTextWidth(text), 1)
			vs = &pdf.TableStyle{
				Text:  element.InvoiceNo,
				Bg:    report.ColorWhite,
				Front: report.ColorTableLine,
			}
			tabel.RawData = append(tabel.RawData, vs)

			//
			text = element.Item
			pdf.ResizeWidth(tabel, p.GetTextWidth(text), 2)
			vs = &pdf.TableStyle{
				Text:  element.Item,
				Bg:    report.ColorWhite,
				Front: report.ColorTableLine,
			}
			tabel.RawData = append(tabel.RawData, vs)
			//
			text = strconv.Itoa(element.Amount)
			pdf.ResizeWidth(tabel, p.GetTextWidth(text), 3)
			vs = &pdf.TableStyle{
				Text:  strconv.Itoa(element.Amount),
				Bg:    report.ColorWhite,
				Front: report.ColorTableLine,
			}
			tabel.RawData = append(tabel.RawData, vs)
			//
			text = strconv.Itoa(element.Fee)
			pdf.ResizeWidth(tabel, p.GetTextWidth(text), 4)
			vs = &pdf.TableStyle{
				Text:  strconv.Itoa(element.Fee),
				Bg:    report.ColorWhite,
				Front: report.ColorTableLine,
			}
			tabel.RawData = append(tabel.RawData, vs)
		}
		//
		text = element.SName
		pdf.ResizeWidth(tabel, p.GetTextWidth(text), 5)
		vs := &pdf.TableStyle{
			Text:  element.SName,
			Bg:    report.ColorWhite,
			Front: report.ColorTableLine,
		}
		tabel.RawData = append(tabel.RawData, vs)
		//
		text = fmt.Sprintf("%.f%s", element.CPercent, "%")
		pdf.ResizeWidth(tabel, p.GetTextWidth(text), 6)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    If(true, report.ColorWhite, report.ColorWhite).(report.Color),
			Front: report.ColorTableLine,
		}
		tabel.RawData = append(tabel.RawData, vs)
		//
		TotalSR += element.SR
		text = fmt.Sprintf("%.f", element.SR)
		pdf.ResizeWidth(tabel, p.GetTextWidth(text), 7)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    If(true, report.ColorWhite, report.ColorWhite).(report.Color),
			Front: report.ColorTableLine,
		}
		tabel.RawData = append(tabel.RawData, vs)
		//
		TotalBouns += element.Bonus
		text = fmt.Sprintf("%.f", element.Bonus)
		pdf.ResizeWidth(tabel, p.GetTextWidth(text), 8)
		vs = &pdf.TableStyle{
			Text:  fmt.Sprintf("%.f", element.Bonus),
			Bg:    If(true, report.ColorWhite, report.ColorWhite).(report.Color),
			Front: report.ColorTableLine,
		}
		tabel.RawData = append(tabel.RawData, vs)
		//備註 => dedeuct的Item
		text = element.DedectItem
		pdf.ResizeWidth(tabel, p.GetTextWidth(text), 9)
		vs = &pdf.TableStyle{
			Text:  element.DedectItem,
			Bg:    If(true, report.ColorWhite, report.ColorWhite).(report.Color),
			Front: report.ColorTableLine,
		}
		tabel.RawData = append(tabel.RawData, vs)
		//
		text = element.Branch
		pdf.ResizeWidth(tabel, p.GetTextWidth(text), 10)
		vs = &pdf.TableStyle{
			Text:  element.Branch,
			Bg:    If(true, report.ColorWhite, report.ColorWhite).(report.Color),
			Front: report.ColorTableLine,
		}
		tabel.RawData = append(tabel.RawData, vs)
		//
		text = element.Percent
		pdf.ResizeWidth(tabel, p.GetTextWidth(text), 11)
		vs = &pdf.TableStyle{
			Text:  element.Percent,
			Bg:    If(true, report.ColorWhite, report.ColorWhite).(report.Color),
			Front: report.ColorTableLine,
		}
		tabel.RawData = append(tabel.RawData, vs)
		//
		// text = element.Percent
		// pdf.ResizeWidth(tabel, p.GetTextWidth(text), 12)
		vs = &pdf.TableStyle{
			Text:  "",
			Bg:    If(true, report.ColorWhite, report.ColorWhite).(report.Color),
			Front: report.ColorTableLine,
		}
		tabel.RawData = append(tabel.RawData, vs)
		//
		text = element.Checknumber
		pdf.ResizeWidth(tabel, p.GetTextWidth(text), 13)
		vs = &pdf.TableStyle{
			Text:  element.Checknumber,
			Bg:    If(true, report.ColorWhite, report.ColorWhite).(report.Color),
			Front: report.ColorTableLine,
		}
		tabel.RawData = append(tabel.RawData, vs)
	}
	return tabel, TotalSR, TotalBouns
}

func If(condition bool, colorA, colorB interface{}) interface{} {
	if condition {
		return colorA
	}
	return colorB
}

func (cm *CModel) agentSignTable(tabel *pdf.DataTable) *pdf.DataTable {
	for _, element := range cm.cList {
		var vs = &pdf.TableStyle{
			Text:  element.Item,
			Bg:    report.ColorWhite,
			Front: report.ColorTableLine,
		}
		tabel.RawData = append(tabel.RawData, vs)
		vs = &pdf.TableStyle{
			Text:  strconv.Itoa(element.Amount),
			Bg:    report.ColorWhite,
			Front: report.ColorTableLine,
		}
		tabel.RawData = append(tabel.RawData, vs)
		vs = &pdf.TableStyle{
			Text:  strconv.Itoa(element.Fee),
			Bg:    report.ColorWhite,
			Front: report.ColorTableLine,
		}
		tabel.RawData = append(tabel.RawData, vs)
		vs = &pdf.TableStyle{
			Text:  element.SName,
			Bg:    report.ColorWhite,
			Front: report.ColorTableLine,
		}
		tabel.RawData = append(tabel.RawData, vs)
		vs = &pdf.TableStyle{
			Text:  strconv.FormatFloat(element.CPercent, 'E', -1, 32),
			Bg:    report.ColorWhite,
			Front: report.ColorTableLine,
		}
		tabel.RawData = append(tabel.RawData, vs)
		vs = &pdf.TableStyle{
			Text:  strconv.FormatFloat(element.SR, 'E', -1, 32),
			Bg:    report.ColorWhite,
			Front: report.ColorTableLine,
		}
		tabel.RawData = append(tabel.RawData, vs)
		vs = &pdf.TableStyle{
			Text:  strconv.FormatFloat(element.Bonus, 'E', -1, 32),
			Bg:    If(true, report.ColorWhite, report.ColorWhite).(report.Color),
			Front: report.ColorTableLine,
		}
		tabel.RawData = append(tabel.RawData, vs)
		vs = &pdf.TableStyle{
			Text:  "",
			Bg:    If(true, report.ColorWhite, report.ColorWhite).(report.Color),
			Front: report.ColorTableLine,
		}
		tabel.RawData = append(tabel.RawData, vs)
		vs = &pdf.TableStyle{
			Text:  "店家",
			Bg:    If(true, report.ColorWhite, report.ColorWhite).(report.Color),
			Front: report.ColorTableLine,
		}
		tabel.RawData = append(tabel.RawData, vs)
	}

	// rm := GetRTModel(cm.imr)
	// rm.GetReceiptData("", "")

	return tabel
}
