package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"dforcepro.com/report"
	"github.com/94peter/toad/pdf"
	"github.com/94peter/toad/resource/db"
	"github.com/94peter/toad/util"
)

type PrePay struct {
	PPid        string          `json:"id"`
	Date        time.Time       `json:"date"`
	ItemName    string          `json:"itemName"`
	Description string          `json:"description"`
	Fee         int             `json:"fee"`
	PrePay      []*BranchPrePay `json:"prepay"`
}

type BranchPrePay struct {
	PPid   string `json:"-"`
	Branch string `json:"branch"`
	Cost   int    `json:"cost"`
}

var (
	prepayM *PrePayModel
)

type PrePayModel struct {
	imr        interModelRes
	db         db.InterSQLDB
	prepayList []*PrePay
}

func GetPrePayModel(imr interModelRes) *PrePayModel {
	if prepayM != nil {
		return prepayM
	}

	prepayM = &PrePayModel{
		imr: imr,
	}
	return prepayM
}

func (prepayM *PrePayModel) GetPrePayData(startDate, endDate string) []*PrePay {

	const PrePayspl = `SELECT PPid, Date, itemname, description, fee FROM public.PrePay
					   where (Date >= '%s' and Date < ('%s'::date + '1 month'::interval)) ;`

	db := prepayM.imr.GetSQLDB()
	rows, err := db.SQLCommand(fmt.Sprintf(PrePayspl, startDate+"-01", endDate+"-01"))
	if err != nil {
		return nil
	}
	var prepayDataList []*PrePay

	for rows.Next() {
		var prepay PrePay

		if err := rows.Scan(&prepay.PPid, &prepay.Date, &prepay.ItemName, &prepay.Description, &prepay.Fee); err != nil {
			fmt.Println("err Scan " + err.Error())
			return nil
		}

		BranchPrePayspl := `SELECT branch, cost FROM public.BranchPrePay where PPid ='` + prepay.PPid + `';`
		//fmt.Println(BranchPrePayspl)
		bpprows, err := db.SQLCommand(fmt.Sprintf(BranchPrePayspl))
		if err != nil {
			return nil
		}
		var BranchPrePayDataList []*BranchPrePay
		for bpprows.Next() {
			var bpp BranchPrePay

			if err := bpprows.Scan(&bpp.Branch, &bpp.Cost); err != nil {
				fmt.Println("err Scan " + err.Error())
				return nil
			}
			BranchPrePayDataList = append(BranchPrePayDataList, &bpp)
			prepay.PrePay = BranchPrePayDataList
		}

		prepayDataList = append(prepayDataList, &prepay)
	}

	// out, err := json.Marshal(arList)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(string(out))
	prepayM.prepayList = prepayDataList
	return prepayM.prepayList
}

func (prepayM *PrePayModel) Json() ([]byte, error) {
	return json.Marshal(prepayM.prepayList)
}

func (prepayM *PrePayModel) PDF() []byte {
	p := pdf.GetNewPDF()

	table := pdf.GetDataTable(pdf.Prepay)

	//取得現有店家
	branchbyte, err := systemM.GetBranchData()
	if err != nil {
		fmt.Println(err)
		return nil
	}
	//將店家資料變成string array , 順便新增最大欄位數
	branchList := []string{}
	s := strings.Split(string(branchbyte), "\"")
	for _, each := range s {
		fmt.Println(each)
		if each != "," && each != "[" && each != "]" {
			branchList = append(branchList, each)
			table.ColumnWidth = append(table.ColumnWidth, pdf.TextWidth)
		}
	}
	table.ColumnLen = len(table.ColumnWidth)
	fmt.Println(" T len", len(table.ColumnWidth))

	data, Total := prepayM.addInfoTable(table, p, branchList)
	fmt.Println(" data len", len(data.ColumnWidth))
	p.CustomizedPrepayTitle(data, "代支費用", branchList)
	data.RawData = data.RawData[4:]
	p.DrawTablePDF(data)
	p.CustomizedPrepay(data, Total)
	return p.GetBytesPdf()

}

func (prepayM *PrePayModel) DeletePrePay(ID string) (err error) {
	const sql = `
				delete from public.PrePay where PPid = '%s';
				delete from public.BranchPrePay where PPid = '%s';				
				`

	interdb := prepayM.imr.GetSQLDB()
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}

	res, err := sqldb.Exec(fmt.Sprintf(sql, ID, ID))
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
		return errors.New("not found anything")
	}
	return nil
}

func (prepayM *PrePayModel) CreatePrePay(prepay *PrePay) (err error) {

	const sql = `INSERT INTO public.prepay
	(ppid, date, itemname, description, fee)
	VALUES ($1, $2, $3, $4, $5)
	;`

	interdb := prepayM.imr.GetSQLDB()
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}

	fakeId := time.Now().Unix()

	res, err := sqldb.Exec(sql, fakeId, prepay.Date, prepay.ItemName, prepay.Description, prepay.Fee)
	//res, err := sqldb.Exec(sql, unix_time, receivable.Date, receivable.CNo, receivable.Sales)
	if err != nil {
		fmt.Println("[Insert err] ", err)
		return err
	}
	id, err := res.RowsAffected()
	if err != nil {
		fmt.Println("PG Affecte Wrong: ", err)
		return err
	}
	fmt.Println(id)

	if id == 0 {
		return errors.New("Invalid operation, CreatePrepay")
	}

	const bppsql = `INSERT INTO public.branchprepay
	(ppid, branch, cost)
	VALUES ($1, $2, $3)
	;`

	// i := 0
	// for range salerlist {
	// 	if salerlist[i].Bid == "testBid" {
	// 		break
	// 	}
	// 	i++
	// }
	i := 0
	for range prepay.PrePay {

		bppres, err := sqldb.Exec(bppsql, fakeId, prepay.PrePay[i].Branch, prepay.PrePay[i].Cost)
		//res, err := sqldb.Exec(sql, unix_time, receivable.Date, receivable.CNo, receivable.Sales)
		if err != nil {
			fmt.Println(err)
			return err
		}

		bppid, err := bppres.RowsAffected()
		if err != nil {
			fmt.Println("PG Affecte Wrong: ", err)
			return err
		}
		if bppid == 0 {
			return errors.New("Invalid operation, CreateBranchPrepay")
		}
		i++
	}

	return nil
}

func (prepayM *PrePayModel) UpdatePrePay(ID string, prepay *PrePay) (err error) {

	const sql = `UPDATE public.prepay
	SET date= $2, itemname= $3, description= $4, fee=$5
	WHERE ppid = $1;
	;`

	interdb := prepayM.imr.GetSQLDB()
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}

	res, err := sqldb.Exec(sql, ID, prepay.Date, prepay.ItemName, prepay.Description, prepay.Fee)
	//res, err := sqldb.Exec(sql, unix_time, receivable.Date, receivable.CNo, receivable.Sales)
	if err != nil {
		fmt.Println("[UPDATE err] ", err)
		return err
	}
	id, err := res.RowsAffected()
	if err != nil {
		fmt.Println("PG Affecte Wrong: ", err)
		return err
	}
	fmt.Println(id)

	if id == 0 {
		return errors.New("Not found prepay")
	}

	const DELbpp = `DELETE FROM public.branchprepay	WHERE ppid= $1 ;`
	_, err = sqldb.Exec(DELbpp, ID)
	if err != nil {
		fmt.Println("DELETE branchprepay Wrong: ", err)
		return err
	}
	const bppsql = `INSERT INTO public.branchprepay
					(ppid, branch, cost)
					VALUES ($1, $2, $3)				
					;`

	i := 0
	for range prepay.PrePay {
		bppres, err := sqldb.Exec(bppsql, ID, prepay.PrePay[i].Branch, prepay.PrePay[i].Cost)
		//res, err := sqldb.Exec(sql, unix_time, receivable.Date, receivable.CNo, receivable.Sales)
		if err != nil {
			fmt.Println(err)
			return err
		}

		bppid, err := bppres.RowsAffected()
		if err != nil {
			fmt.Println("PG Affecte Wrong: ", err)
			return err
		}
		if bppid == 0 {
			return errors.New("Invalid operation , UpdateBranchPrepay")
		}
		i++
	}
	return nil
}

func (prepayM *PrePayModel) addInfoTable(tabel *pdf.DataTable, p *pdf.Pdf, branch []string) (tabel_final *pdf.DataTable,
	Total []int) {
	//Total[0] For 支出金額
	Total = []int{}
	for i := 0; i < len(branch)+1; i++ {
		Total = append(Total, 0)
	}

	for _, element := range prepayM.prepayList {
		//
		text := element.Date.Format("2006-01-02")
		text, _ = util.ADtoROC(text, "ch")
		pdf.ResizeWidth(tabel, p.GetTextWidth(text), 0)
		vs := &pdf.TableStyle{
			Text:  text,
			Bg:    report.ColorWhite,
			Front: report.ColorTableLine,
		}
		tabel.RawData = append(tabel.RawData, vs)
		//
		text = element.ItemName
		pdf.ResizeWidth(tabel, p.GetTextWidth(text), 1)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    report.ColorWhite,
			Front: report.ColorTableLine,
		}
		tabel.RawData = append(tabel.RawData, vs)
		//
		text = element.Description
		pdf.ResizeWidth(tabel, p.GetTextWidth(text), 2)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    report.ColorWhite,
			Front: report.ColorTableLine,
		}
		tabel.RawData = append(tabel.RawData, vs)
		//
		Total[0] += element.Fee
		text = pr.Sprintf("%d", element.Fee)
		pdf.ResizeWidth(tabel, p.GetTextWidth(text), 3)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    report.ColorWhite,
			Front: report.ColorTableLine,
			Align: pdf.AlignRight,
		}
		tabel.RawData = append(tabel.RawData, vs)
		//
		index := 4
		for k := 0; k < len(branch); k++ {
			f := true
			text = "-"
			for _, Prepay := range element.PrePay {
				if Prepay.Branch == branch[k] {
					text = pr.Sprintf("%d", Prepay.Cost)
					Total[k+1] += Prepay.Cost
					f = false
				}
			}

			pdf.ResizeWidth(tabel, p.GetTextWidth(text), index+k)
			vs = &pdf.TableStyle{
				Text:  text,
				Bg:    report.ColorWhite,
				Front: report.ColorTableLine,
				Align: If(f, pdf.AlignCenter, pdf.AlignRight).(int),
			}
			tabel.RawData = append(tabel.RawData, vs)
		}

	}

	tabel_final = tabel
	return
}
