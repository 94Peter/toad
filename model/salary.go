package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"toad/excel"
	"toad/pdf"
	"toad/permission"
	"toad/resource/db"
	"toad/txt"
	"toad/util"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type BranchSalary struct {
	BSid     string    `json:"BSid"`
	Branch   string    `json:"branch"`
	StrDate  string    `json:"-"` //建立string Date
	Date     time.Time `json:"date"`
	LastDate time.Time `json:"date"` //上次建立薪資的日期隔天
	Name     string    `json:"name"`
	Total    string    `json:"total"`
	Lock     string    `json:"lock"`
	//SalerSalaryList []*SalerSalary `json:"commissionList"`
}

type SalerSalary struct {
	Sid           string `json:"id"`
	BSid          string `json:"bsid"`
	Branch        string `json:"branch"`
	Date          string `json:"date"`
	SName         string `json:"name"`
	Salary        int    `json:"salary"`
	Pbonus        int    `json:"pbonus"`
	Lbonus        int    `json:"lbonus"`
	Abonus        int    `json:"abonus"`
	Total         int    `json:"total"`
	SP            int    `json:"sp"`
	Tax           int    `json:"tax"`
	LaborFee      int    `json:"laborFee"`
	HealthFee     int    `json:"healthFee"`
	Welfare       int    `json:"welfare"`
	CommercialFee int    `json:"commercialFee"`
	Other         int    `json:"other"`
	Description   string `json:"description"`
	TAmount       int    `json:"transferAmount"`
	Workday       int    `json:"workday"`
	ManagerBonus  int    `json:"managerBonus"`
	Lock          string `json:"lock"`
	Code          string `json:"code"`
	ManagerID     string `json:"-"`
}

type NHISalary struct {
	Sid            string `json:"id"`
	BSid           string `json:"bsid"`
	SName          string `json:"name"`
	PayrollBracket int    `json:"payrollBracket"`
	Salary         int    `json:"salary"`
	Pbonus         int    `json:"pbonus"`
	Bonus          int    `json:"bonus"`
	Total          int    `json:"total"`
	SalaryBalance  int    `json:"salaryBalance"`
	PD             int    `json:"PD"`
	FourBouns      int    `json:"fourBouns"`
	SP             int    `json:"SP"`
	FourSP         int    `json:"fourSP"`
	PTSP           int    `json:"PTSP"`
	Code           string `json:"code"`
	/////////pdf
	Title       string `json:"-"`
	Description string `json:"-"`
	Branch      string `json:"-"`
}

type IncomeExpense struct {
	BSid string `json:"bsid"`

	// Income struct {
	// 	SR           int `json:"sr"`
	// 	Salesamounts int `json:"salesamounts"`
	// 	Businesstax  int `json:"businesstax"`
	// } `json:"income"`

	Income Income `json:"income"`

	Expense Expense `json:"expense"`

	BusinessIncomeTax int `json:"businessIncomeTax"`
	Aftertax          int `json:"afterTax"`
	Pretax            int `json:"pretax"`
	Lastloss          int `json:"lastLoss"`
	ManagerBonus      int `json:"managerBonus"`
	EarnAdjust        int `json:"earnAdjust"`
}

type Income struct {
	SR           int `json:"sr"`
	Salesamounts int `json:"salesamounts"`
	Businesstax  int `json:"businesstax"`
}

type Expense struct {
	Pbonus        int     `json:"pbonus"`
	LBonus        int     `json:"lBonus"`
	Salary        int     `json:"salary"`
	Prepay        int     `json:"prepay"`
	Pocket        int     `json:"pocket"`
	Amorcost      int     `json:"amorcost"`
	Agentsign     int     `json:"agentsign"`
	Rent          int     `json:"rent"`
	Commercialfee int     `json:"commercialFee"`
	Annualbonus   int     `json:"annualBonus"`
	AnnualRatio   float64 `json:"annualRatio"`
	SalerFee      int     `json:"salerFee"`
}

type TransferSalary struct {
	Branch  string
	Date    string
	Account string
	Amount  int
	BankNo  string
	IDNo    string
}

type Cid struct {
	Sid string `json:"sid"`
	Rid string `json:"rid"`
}

type CloseAccount struct {
	id        string    `json:"-"`
	Date      time.Time `json:"date"`
	CloseDate time.Time `json:"closeDate"`
	Uid       string    `json:"uid"`
	Status    string    `json:"-"`
}

var (
	salaryM *SalaryModel
	pr      = message.NewPrinter(language.English)
)

type SalaryModel struct {
	imr               interModelRes
	db                db.InterSQLDB
	salerSalaryList   []*SalerSalary
	branchSalaryList  []*BranchSalary
	NHISalaryList     []*NHISalary
	IncomeExpenseList []*IncomeExpense
	MailList          []*ConfigSaler

	TransferSalaryList []*TransferSalary

	SystemAccountList []*SystemAccount
	CommissionList    []*Commission
	FnamePdf          string
	SMTPConf          util.SendMail

	CloseAccount *CloseAccount
}

func GetSalaryModel(imr interModelRes) *SalaryModel {
	if salaryM != nil {
		return salaryM
	}

	salaryM = &SalaryModel{
		imr: imr,
	}
	return salaryM
}

func (salaryM *SalaryModel) SetSMTPConf(conf util.SendMail) {
	if salaryM == nil {
		return
	}
	salaryM.SMTPConf = conf
}

func (salaryM *SalaryModel) GetBranchSalaryData(date, dbname string) []*BranchSalary {

	sql := "SELECT bsid, to_timestamp(bsid::int), branch, name, total, lock FROM public.branchsalary" +
		" where date Like '%" + date + "%' order by bsid;"

	db := salaryM.imr.GetSQLDBwithDbname(dbname)
	rows, err := db.SQLCommand(sql)

	if err != nil {
		return nil
	}
	var bsDataList []*BranchSalary

	for rows.Next() {
		var bs BranchSalary

		if err := rows.Scan(&bs.BSid, &bs.Date, &bs.Branch, &bs.Name, &bs.Total, &bs.Lock); err != nil {
			fmt.Println("err Scan " + err.Error())
			return nil
		}

		bsDataList = append(bsDataList, &bs)
	}

	// out, err := json.Marshal(arList)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(string(out))
	salaryM.branchSalaryList = bsDataList
	return salaryM.branchSalaryList
}

func (salaryM *SalaryModel) Json(config string) ([]byte, error) {
	switch config {
	case "AccountSettlement":
		return json.Marshal(salaryM.CloseAccount)
	case "BranchSalary":
		return json.Marshal(salaryM.branchSalaryList)
	case "SalerSalary":
		return json.Marshal(salaryM.salerSalaryList)
	case "NHISalary":
		return json.Marshal(salaryM.NHISalaryList)
	case "ManagerBonus":
		return json.Marshal(salaryM.IncomeExpenseList)
	default:
		fmt.Println("unknown config type")
		break
	}
	return nil, nil
}
func (salaryM *SalaryModel) GetPDFByte() []byte {
	//取巧作法
	p := pdf.GetOriPDF()
	data := p.GetBytesPdf()
	p = pdf.GetNewPDF()
	return data
}

func (salaryM *SalaryModel) PDF(dbname string, mtype int, isNew bool, things ...string) {
	var p *pdf.Pdf
	if isNew {
		p = pdf.GetNewPDF()
	} else {
		p = pdf.GetOriPDF()
	}
	send := ""
	for _, it := range things {
		send = it
	}

	table := pdf.GetDataTable(mtype)

	//Header長度重製
	for index, e := range table.RawData {
		//fmt.Println(e)
		pdf.ResizeWidth(table, p.GetTextWidth(e.Text), index)
	}

	fmt.Println("SalaryModel mtype:", mtype)
	switch mtype {
	case pdf.BranchSalary: //1
		if len(salaryM.salerSalaryList) <= 0 {
			return
		}
		data, T_Salary, T_Pbonus, T_Abonus, T_Lbonus, T_Total, T_SP, T_Tax, T_LaborFee, T_HealthFee, T_Welfare, T_CommercialFee, T_TAmount, T_Other := salaryM.addBranchSalaryInfoTable(table, p)
		fmt.Println("DrawTablePDF:")
		p.DrawTablePDF(data)
		fmt.Println("CustomizedBranchSalary:")
		p.CustomizedBranchSalary(data, T_Salary, T_Pbonus, T_Abonus, T_Lbonus, T_Total, T_SP, T_Tax, T_LaborFee, T_HealthFee, T_Welfare, T_CommercialFee, T_TAmount, T_Other)
		date, _ := util.ADtoROC(salaryM.salerSalaryList[0].Date, "file")
		p.WriteFile(salaryM.salerSalaryList[0].Branch + "薪資表" + date)
		break
	case pdf.AgentSign: //5
		data, T_SR, T_Bonus := salaryM.addAgentSignInfoTable(table, p)
		//data.RawData = data.RawData[data.ColumnLen:]
		fmt.Println("DrawTablePDF")
		p.DrawTablePDF(data)
		fmt.Println("CustomizedAgentSign:T_SR ", T_SR, " T_Bonus", T_Bonus)
		//SR, Bonus := p.CustomizedAgentSign(data, "saler.Name", T_Bonus, T_SR)
		//total_SR += SR
		//total_Bonus += Bonus
		p.CustomizedAgentSign(table, T_Bonus, T_SR)
		break
	case pdf.SalarCommission: //8
		mailList, err := salaryM.getSalerEmail(dbname)
		if err != nil {
			//getSalerEmail 失敗
			fmt.Println(err)
			return
		}
		//用mailList做檔案名稱
		for _, saler := range mailList {
			//fmt.Println("saler:", saler.SName)
			p := pdf.GetNewPDF()
			table = pdf.GetDataTable(mtype)
			//Header長度重製
			for index, e := range table.RawData {
				pdf.ResizeWidth(table, p.GetTextWidth(e.Text), index)
			}
			//根據saler.Sid比對所有資料，相同的寫入pdf
			data, T_SR, T_Bonus, date := salaryM.addSalerCommissionInfoTable(table, p, salaryM.CommissionList, saler.Sid)
			p.DrawTablePDF(data)
			p.CustomizedSalerCommission(data, saler.SName, int(T_Bonus), int(T_SR))
			//fmt.Println("pdf.SalarCommission date:", date, "#", saler.SName)
			//mailList 有其他店的人
			if date == "error" {
				continue
			}
			date, _ = util.ADtoROC(date, "file")
			fname := saler.Branch + "-" + saler.SName + "-" + saler.Code + "-傭金表" + date
			p.WriteFile(fname)

			// if send == "true" {
			// 	util.RunSendMail(salaryM.SMTPConf.Host, salaryM.SMTPConf.Port, salaryM.SMTPConf.Password, salaryM.SMTPConf.User, saler.Email, pdf.ReportToString(mtype), "薪資測試\r\n開啟若有密碼，則為000000或者您的身分證號碼", fname+".pdf")
			// }
		}
		break
	case pdf.SalerSalary: //7
		if len(salaryM.salerSalaryList) > 0 {
			//systemM.GetAccountData()
			//mailList, err := salaryM.getSalerEmail()

			for index, element := range salaryM.salerSalaryList {
				p := pdf.GetNewPDF()
				table := pdf.GetDataTable(mtype)
				//Header長度重製
				for index, e := range table.RawData {
					//fmt.Println(e)
					pdf.ResizeWidth(table, p.GetTextWidth(e.Text), index)
				}

				data := salaryM.addSalerSalaryInfoTable(table, p, index, element)
				//fmt.Println("DrawTablePDF:type:7")
				p.DrawTablePDF(data)
				date, _ := util.ADtoROC(element.Date, "file")
				fname := element.Branch + "-" + element.SName + "-" + element.Code + "-" + "薪資表" + date
				p.WriteFile(fname)

				if send == "true" && element.Sid == element.ManagerID {
					fmt.Println("Build PDFIncomeStatement")
					salaryM.PDFIncomeStatement(element.Branch, element.SName, element.Code, date, dbname)
				}
				//p = nil
				// if send == "true" {
				// 	if err != nil {
				// 		//getSalerEmail 失敗
				// 		fmt.Println(err)
				// 		return
				// 	}

				// 	for _, myAccount := range mailList {
				// 		if myAccount.Sid == element.Sid {
				// 			//fmt.Println(fname, " ", myAccount, element)
				// 			util.RunSendMail(salaryM.SMTPConf.Host, salaryM.SMTPConf.Port, salaryM.SMTPConf.Password, salaryM.SMTPConf.User, myAccount.Email, pdf.ReportToString(mtype), "薪資測試\r\n開啟若有密碼，則為000000或者您的身分證號碼", fname+".pdf")
				// 			//util.RunSendMail(salaryM.SMTPConf.Host, salaryM.SMTPConf.Port, salaryM.SMTPConf.Password, salaryM.SMTPConf.User, "geassyayaoo3@gmail.com", pdf.ReportToString(mtype), "薪資測試\r\n開啟若有密碼，則為123456", fname+".pdf")

				// 		}
				// 	}
				// }
			}
		}

		break
	case pdf.SR: // 6
		//table = pdf.GetDataTable(mtype)
		if len(salaryM.CommissionList) > 0 {
			fmt.Println("SR")
			data, _, _ := salaryM.addSRInfoTable(table, p)
			p.DrawTablePDF(data)
			p.NewLine(25)
		}
		break
	case pdf.NHI: //3
		//table = pdf.GetDataTable(mtype)
		if len(salaryM.NHISalaryList) > 0 {
			fmt.Println("NHI")
			data, T_PayrollBracket, T_Salary, T_Pbonus, T_Bonus, T_Total, T_Balance, T_PTSP,
				T_PD, T_FourBouns, T_SP, T_FourSP, T_Tax, T_SPB := salaryM.addNHIInfoTable(table, p)
			// data, _, _, _, _, _, _, _,
			// 	_, _, _, _, _ := salaryM.addNHIInfoTable(table, p)
			fmt.Println("DrawTablePDF")
			p.DrawTablePDF(data)
			p.NewLine(25)
			p.CustomizedNHI(data, T_PayrollBracket, T_Salary, T_Pbonus, T_Bonus, T_Total, T_Balance, T_PTSP,
				T_PD, T_FourBouns, T_SP, T_FourSP, T_Tax, T_SPB)
		}
		break
	case 0:
		//salaryM.PDFIncomeStatement("明湖店")

		break
	}

	return //p.GetBytesPdf()
}

func (salaryM *SalaryModel) PDFIncomeStatement(branch, SName, Code, date, dbname string) {
	p := pdf.GetNewPDF(pdf.PageSizeA4_)
	indexM := GetIndexModel(salaryM.imr)
	mdate, _ := time.Parse(time.RFC3339, "2019-12-31T16:00:00Z")
	incomeStatement, err := indexM.GetIncomeStatement(branch, dbname, mdate)
	if err != nil {
		fmt.Println("ERROR PDFIncomeStatement:", err)
		return
	}
	if incomeStatement == nil {
		fmt.Println("incomeStatement null")
		return
	}

	p.CustomizedIncomeStatement(incomeStatement.Branch, incomeStatement.Income.SR, incomeStatement.Income.Salesamounts, incomeStatement.Income.Businesstax,
		incomeStatement.Expense.Pbonus, incomeStatement.Expense.LBonus, incomeStatement.Expense.Salary, incomeStatement.Expense.Prepay, incomeStatement.Expense.Pocket,
		incomeStatement.Expense.Amorcost, incomeStatement.Expense.Agentsign, incomeStatement.Expense.Rent, incomeStatement.Expense.Commercialfee, incomeStatement.Expense.Annualbonus, incomeStatement.Expense.SalerFee,
		incomeStatement.BusinessIncomeTax, incomeStatement.Aftertax, incomeStatement.Pretax, incomeStatement.Lastloss, incomeStatement.ManagerBonus, incomeStatement.Expense.Agentsign)

	fname := branch + "-" + SName + "-" + Code + "-" + "損益表" + date
	p.WriteFile(fname)

	// (branch string, SR, Salesamounts, Businesstax,
	// 	Pbonus, LBonus, Salary, Prepay, Pocket, Amorcost, Agentsign, Rent, Commercialfee, Annualbonus, AnnualRatio, SalerFee,
	// 	BusinessIncomeTax, Aftertax, Pretax, Lastloss, ManagerBonus int)

}

func (salaryM *SalaryModel) EXCEL(mtype int) {
	// var ex *excel.Excel
	ex := excel.GetOriExcel()

	fmt.Println("mtype:", mtype)
	switch mtype {
	case excel.PayrollTransfer:

		//data := excel.GetDataTable(mtype)
		DataList := salaryM.addPayrollTransferInfoTable(mtype)
		ex.FillText(DataList)
		break
	case excel.IncomeTaxReturn:
		fmt.Println("EXCEL IncomeTaxReturn")
		DataList := salaryM.addIncomeTaxReturnInfoTable(mtype)
		ex.FillText(DataList)

		break
	}

	return //p.GetBytesPdf()
}

func (salaryM *SalaryModel) TXT(mtype int) {
	fmt.Println("TXT:", mtype)
	switch mtype {
	case txt.SalaryTransfer:
		txt.Write(salaryM.makeTxtTransferSalary())
		break
	}

	return //p.GetBytesPdf()
}

func (salaryM *SalaryModel) DeleteSalary(ID, dbname string) (err error) {
	const sql = `
				delete from public.SalerSalary where bsid = '%s';
				delete from public.NHISalary where bsid = '%s';
				delete from public.BranchSalary where bsid = '%s';
				delete from public.incomeexpense where bsid = '%s';
				UPDATE public.commission SET bsid = null , status = 'normal' WHERE bsid = '%s';
				`

	interdb := salaryM.imr.GetSQLDBwithDbname(dbname)
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}

	_, err = sqldb.Exec(fmt.Sprintf(sql, ID, ID, ID, ID, ID))
	//res, err := sqldb.Exec(sql, unix_time, receivable.Date, receivable.CNo, receivable.Sales)
	if err != nil {
		fmt.Println(err)
		return err
	}

	defer sqldb.Close()
	return nil
}

func (salaryM *SalaryModel) isOK_CreateSalary(dbname string) (err error) {
	// //( (case when cb.sid is not null then ie.managerbonus else 0 end) + ss.tamount )
	sql := `select  c.date, c.nhi, c.li, c.nhi2nd, c.mmw from public.ConfigParameter C `
	// //where ss.bsid = '%s' and ss.sid like '%s'`
	var Tag = true
	db := salaryM.imr.GetSQLDBwithDbname(dbname)
	rows, err := db.SQLCommand(sql)
	if err != nil {
		return err
	}
	for rows.Next() {
		Tag = false
		break
	}
	if Tag {
		fmt.Println("基礎參數未填")
		return errors.New("基礎參數未填")
	}
	//######
	Tag = false
	sql = `select branch , sid  from public.configbranch`
	rows, err = db.SQLCommand(sql)
	if err != nil {
		return err
	}
	var s = ""
	for rows.Next() {
		var sid NullString
		var branch string
		if err := rows.Scan(&branch, &sid); err != nil {
			fmt.Println("err Scan " + err.Error())
			return err
		}
		if !sid.Valid {
			s += branch + "店長為空。"
			Tag = true
		}
	}
	if Tag {
		fmt.Println(s)
		return errors.New(s)
	}

	return nil
}
func (salaryM *SalaryModel) getNextDayFromLastTimeSalary(dbname string) (mtime time.Time, err error) {
	mtime = time.Now()
	interdb := salaryM.imr.GetSQLDBwithDbname(dbname)

	const sql = `SELECT max(date) FROM public.BranchSalary;`

	rows, err := interdb.SQLCommand(fmt.Sprintf(sql))
	if err != nil {
		return
	}

	loc, _ := time.LoadLocation("Asia/Taipei")
	t := time.Now().In(loc)
	y, m, _ := t.Date()
	t = time.Date(y, m, 1, 0, 0, 0, 0, loc)
	strTime := fmt.Sprintf("%d-%02d-%02d", t.Year(), t.Month(), 1) //default 月初
	//Scan失敗使用default值,now + 1 day
	for rows.Next() {
		if err := rows.Scan(&strTime); err != nil {
			fmt.Println("getNextDayFromLastTimeSalary err Scan " + err.Error())
		}
	}

	mtime, err = time.Parse(time.RFC3339, strTime+"T00:00:00+08:00")

	year, month, day := mtime.Date()

	mtime = time.Date(year, month, day+1, 0, 0, 0, 0, loc)

	return
}

/**
*確認是否可以建立薪資
*建立薪資表時，建立關帳日。
*取得上一次建立薪資的隔天(預設當月月初)，後續抓取攤提費用、代支使用
*建立分店總表[根據幾家分店，建立基本bsid]
*針對傭金綁定bsid
*計算並建立業務薪資
*更新分店總表數值
**/
func (salaryM *SalaryModel) CreateSalary(bs *BranchSalary, cid []*Cid, dbname, permission string) (err error) {

	err = salaryM.isOK_CreateSalary(dbname)
	if err != nil {
		fmt.Println("CreateSalary err:" + err.Error())
		return err
	}

	// mCaD := &CloseAccount{
	// 	CloseDate: bs.Date,
	// 	Uid:       "salary",
	// }
	//salaryM.CloseAccountSettlement(mCaD, permission, dbname) // 2020-12-09 說不做關帳

	// ca, err := salaryM.CheckValidCloseDate(bs.Date, dbname)
	// if err != nil {
	// 	return
	// }
	// fmt.Println("ca:", ca.CloseDate)

	bs.Date = setDayEndDate(bs.Date)
	fmt.Println(bs.Date)
	fmt.Println(bs.StrDate)

	t, err := salaryM.getNextDayFromLastTimeSalary(dbname)
	bs.LastDate = t
	fmt.Println("bs.LastDate:", bs.LastDate)

	// const sql = `INSERT INTO public.branchsalary
	// 			(BSid, date, branch, name)
	// 			SELECT sum(1) over (order by branch) + $1, $2, branch, $3 FROM public.configbranch;`
	const sql = `INSERT INTO public.branchsalary
				(BSid, date, branch, name)	
				SELECT sum(1) over (order by cb.branch) + $1 , $2, cb.branch, $3 
				FROM public.configbranch cb 
				left join (
						SELECT  tmp.branch , 
					(CASE 
					WHEN bsid is NULL THEN 0
					ELSE 1
					END
					) hasBind
					FROM public.receipt r, public.commission c
					inner join(
						SELECT A.sid, A.Branch FROM public.ConfigSaler A 
						Inner Join ( 
							select sid, max(zerodate) zerodate from public.configsaler cs 
							where now() > zerodate
							group by sid 
						) B on A.sid=B.sid and A.zeroDate = B.zeroDate
					) tmp on tmp.sid = c.sid
					where c.rid = r.rid and r.date >= $4 and Date < ( $4::date + '1 month'::interval ) 
					group by tmp.branch , hasBind
				) tmp on tmp.branch = cb.branch
				where tmp.hasbind is null or tmp.hasbind = 0
				;`
	//使得每個BSid + 1
	interdb := salaryM.imr.GetSQLDBwithDbname(dbname)
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}
	defer sqldb.Close()

	fakeId := time.Now().Unix()
	bs.BSid = strconv.Itoa(int(fakeId))

	res, err := sqldb.Exec(sql, fakeId, bs.StrDate, bs.Name, bs.Date)
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
	fmt.Println("CreateSalary:", id)

	if id == 0 {
		return errors.New("Invalid operation, CreateBranchSalary")
	}

	////將曾經排除的傭金加入
	// scsErr := salaryM.SetCommissionBSid(bs, cid, dbname)
	// if scsErr != nil {
	// 	return nil
	// 	//return css_err
	// }

	cssErr := salaryM.CreateSalerSalary(bs, cid, dbname)
	if cssErr != nil {
		return nil
		//return css_err
	}

	_err := salaryM.UpdateBranchSalaryTotal(dbname)
	if _err != nil {
		return nil
		//return css_err
	}

	return nil
}

/*
	福利金: 總薪資*0.01
	商耕費: 總薪資*比例/100
	*0.01 vs *1/100 答案是不一樣的。
	所以使用CAST轉型後用ROUND去修正答案。
*/
func (salaryM *SalaryModel) CreateSalerSalary(bs *BranchSalary, cid []*Cid, dbname string) (err error) {

	const sql = `INSERT INTO public.salersalary
	(bsid, sid, date,  branch, sname, salary, pbonus, total, laborfee, healthfee, welfare, commercialfee, year, sp, tamount)
	SELECT BS.bsid, A.sid, COALESCE(C.dateID, $1) dateID, A.branch, A.sname,  A.Salary, COALESCE(C.Pbonus,0) Pbonus, 
	COALESCE(A.Salary+  COALESCE(C.Pbonus,0), A.Salary) total, ROUND(A.InsuredAmount*CP.LI*0.2/100) LaborFee,ROUND(A.PayrollBracket*CP.nhi*0.3/100) HealthFee,
	ROUND(COALESCE(A.Salary+  COALESCE(C.Pbonus,0), A.Salary)*0.01) Welfare,  ROUND( CAST(COALESCE(A.Salary+  COALESCE(C.Pbonus,0) ,A.Salary) *(cb.commercialFee/100) as numeric) ) commercialFee,
	$2 ,
	(CASE WHEN A.salary = 0 and A.association = 1 then 0 
		WHEN (COALESCE(A.Salary + COALESCE(C.Pbonus,0) ,A.Salary)) <= CP.mmw then 0	 	
	   WHEN A.salary = 0 and A.association = 0 then COALESCE(A.Salary+  COALESCE(C.Pbonus,0) ,A.Salary) * cp.nhi2nd / 100 	 	
	   else
		   ( CASE WHEN ((COALESCE(A.Salary+  COALESCE(C.Pbonus,0) ,A.Salary)) - 4 * A.PayrollBracket) > 0 then ((COALESCE(A.Salary+  COALESCE(C.Pbonus,0) ,A.Salary)) - 4 * A.PayrollBracket) * cp.nhi2nd / 100 else 0 end)
	   end
	  ) sp ,
	  (COALESCE(A.Salary+  COALESCE(C.Pbonus,0),A.Salary) - ROUND(COALESCE(A.Salary+  COALESCE(C.Pbonus,0), A.Salary)*0.01) - ROUND( CAST(COALESCE(A.Salary+  COALESCE(C.Pbonus,0) ,A.Salary) *(cb.commercialFee/100) as numeric) ) 
	  -  ROUND(A.InsuredAmount*CP.LI*0.2/100) - ROUND(A.PayrollBracket*CP.nhi*0.3/100) ) - 
	 (CASE WHEN A.salary = 0 and A.association = 1 then 0 
		 WHEN (COALESCE(A.Salary + COALESCE(C.Pbonus,0) ,A.Salary)) <= CP.mmw then 0	 	
		WHEN A.salary = 0 and A.association = 0 then COALESCE(A.Salary+  COALESCE(C.Pbonus,0),A.Salary) * cp.nhi2nd / 100 	 	
		else
			( CASE WHEN ((COALESCE(A.Salary+  COALESCE(C.Pbonus,0),A.Salary)) - 4 * A.PayrollBracket) > 0 then ((COALESCE(A.Salary+  COALESCE(C.Pbonus,0),A.Salary)) - 4 * A.PayrollBracket) * cp.nhi2nd / 100 else 0 end)
		end
	   ) Tamount
	FROM public.ConfigSaler A
	Inner Join ( 
		select sid, max(zerodate) zerodate from public.configsaler cs 
		where now() > zerodate
		group by sid 
	) B on A.sid=B.sid and A.zeroDate = B.zeroDate
	left join (
	SELECT c.sid , to_char(r.date at time zone 'UTC' at time zone 'Asia/Taipei','YYYY-MM')::varchar(50) dateID, sum(c.bonus) Pbonus
	FROM public.receipt r, public.commission c
	where c.rid = r.rid and c.bsid is null and c.status = 'normal' and extract(epoch from Date)  <= $3
	group by dateID , c.sid 
	) C on C.sid = A.Sid 
	cross join (
		select  c.date, c.nhi, c.li, c.nhi2nd, c.mmw from public.ConfigParameter C
		inner join(
			select  max(date) date from public.ConfigParameter 
		) A on A.date = C.date limit 1
	) CP
	left join public.branchsalary BS on BS.branch = A.Branch and BS.date = $1
	
	left join(
		select branch , commercialFee from public.configbranch 
	) CB on CB.branch = A.branch
	where BS.bsid is not null
	ON CONFLICT (bsid,sid,date,branch) DO Nothing;	
	`

	year := bs.StrDate[0:4]
	fmt.Println(year)

	interdb := salaryM.imr.GetSQLDBwithDbname(dbname)
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}
	defer sqldb.Close()
	//fmt.Println("BSID:" + bs.BSid)
	//fmt.Println(bs.Date)
	//GCP local time zone是+0時區，預設前端丟進來的是+8時區

	// b, _ := time.Parse(time.RFC3339, bs.Date+"-01T00:00:00+08:00")
	// fmt.Println("CreateSalerSalary:", bs.Date+"-01 =>", b.Unix())

	//res, err := sqldb.Exec(sql, bs.StrDate, year, salaryM.CloseAccount.CloseDate.Unix())
	res, err := sqldb.Exec(sql, bs.StrDate, year, bs.Date.Unix())
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
	fmt.Println("CreateSalerSalary:", id)

	if id == 0 {
		fmt.Println("CreateSalerSalary, no create anyone ")
		return errors.New("CreateSalerSalary, not found any commission")
	}

	//綁定更改BSid (一筆都沒有也無所謂(表示只有底薪))
	_ = salaryM.UpdateCommissionBSidAndStatus(bs, cid, dbname)

	//綁定更改BSid後才可建立紅利表，預設使用5%成本(年終提撥)
	cieErr := salaryM.CreateIncomeExpense(bs, dbname)
	if cieErr != nil {
		return nil
		//return css_err
	}

	ucnhi_err := salaryM.CreateNHISalary(year, dbname)
	if ucnhi_err != nil {
		return nil
		//return ucias_err
	}

	return nil
}

func (salaryM *SalaryModel) SetCommissionBSid(bs *BranchSalary, cid []*Cid, dbname string) (err error) {

	for _, cid := range cid {
		fmt.Println("*cid:", cid.Rid)
		fmt.Println("*sid:", cid.Sid)

		const sql = `Update public.commission
					Set bsid = subQuery.bsid , status = 'join'
					FROM (
					SELECT bs.bsid
					FROM public.commission c
					inner join(
						SELECT A.sid, A.sname, A.branch
						FROM public.ConfigSaler A 
						Inner Join ( 
							select sid , max(zerodate) zerodate from public.configsaler cs 
							where now() > zerodate
							group by sid 
						) B on A.sid=B.sid and A.zeroDate = B.zeroDate
					) saler on saler.sid = c.sid
					inner join(
						SELECT bsid, date, branch
						FROM public.branchsalary
						where date = $1
					) bs on bs.branch = saler.branch
					where c.sid = $2 and c.rid = $3
					) subQuery
					where sid = $2 and rid = $3;`

		year := bs.StrDate[0:4]
		fmt.Println(year)

		interdb := salaryM.imr.GetSQLDBwithDbname(dbname)
		sqldb, err := interdb.ConnectSQLDB()
		if err != nil {
			return err
		}
		//fmt.Println("BSID:" + bs.BSid)
		//fmt.Println(bs.Date)
		res, err := sqldb.Exec(sql, bs.Date, cid.Sid, cid.Rid)
		//res, err := sqldb.Exec(sql, unix_time, receivable.Date, receivable.CNo, receivable.Sales)
		if err != nil {
			fmt.Println("[Update err] ", err)
			return err
		}
		id, err := res.RowsAffected()
		if err != nil {
			fmt.Println("PG Affecte Wrong: ", err)
			return err
		}
		fmt.Println("SetCommissionBSid:", id)

		if id == 0 {
			fmt.Println("SetCommissionBSid, not found any commission ")

		}
		defer sqldb.Close()
	}
	return nil
}

func (salaryM *SalaryModel) CreateIncomeExpense(bs *BranchSalary, dbname string) (err error) {
	//(subtable.pretaxTotal + subtable.PreTax )  lastloss ,   應該不包含這期虧損
	const sql = `INSERT INTO public.incomeexpense
	(bsid, Pbonus ,LBonus, salary, prepay, pocket, amorcost, sr, annualbonus, salesamounts,  businesstax, agentsign, rent, commercialfee, pretax, businessincometax, aftertax,  lastloss, managerbonus, annualratio )	
	WITH  vals  AS (VALUES ( 'none' ) )
	SELECT subtable.bsid , subtable.Pbonus, subtable.LBonus , subtable.salary, subtable.prepay, subtable.pocket , subtable.thisMonthAmor , subtable.sr, subtable.annualbonus, subtable.salesamounts , subtable.businesstax , subtable.agentsign , subtable.rent,
	subtable.tCFee, subtable.PreTax , ( CASE WHEN subtable.PreTax > 0 then subtable.PreTax * 0.2 else 0 end ) BusinessIncomeTax, 
	subtable.PreTax - ( CASE WHEN subtable.PreTax > 0 then subtable.PreTax * 0.2 else 0 end ) AfterTax , 
	(subtable.pretaxTotal)  lastloss ,  
	( CASE WHEN (subtable.PreTax - ( CASE WHEN subtable.PreTax > 0 then subtable.PreTax * 0.2 else 0 end )) + (subtable.pretaxTotal) + 0 > 0 then 
				(subtable.PreTax - ( CASE WHEN subtable.PreTax > 0 then subtable.PreTax * 0.2 else 0 end ) + (subtable.pretaxTotal) + 0) * 0.2
	  else 0 end
	) managerbonus  , subtable.annualratio
	FROM vals as v
	cross join (
	SELECT incomeexpense.branch , COALESCE(incomeexpense.pretaxTotal ,0) pretaxTotal , BS.Bsid,BonusTable.PBonus , BonusTable.LBonus , BonusTable.Salary , COALESCE(prepayTable.prepay,0) prepay , COALESCE(pocketTable.pocket,0) pocket , COALESCE(amorTable.thisMonthAmor,0) thisMonthAmor,
	COALESCE(commissionTable.SR,0) SR, COALESCE(commissionTable.SR / 1.05 ,0) salesamounts , COALESCE(commissionTable.SR - commissionTable.SR / 1.05 ,0) businesstax, configTable.agentsign, configTable.rent, configTable.commercialfee, 
	( COALESCE(commissionTable.SR,0)/1.05  - COALESCE(amorTable.thisMonthAmor,0) - configTable.agentsign - configTable.rent - COALESCE(pocketTable.pocket,0) - COALESCE(prepayTable.prepay,0) - BonusTable.PBonus - 
	BonusTable.Salary - BonusTable.LBonus - COALESCE(commissionTable.SR,0) * 0.05 - BonusTable.tCFee - 0  ) PreTax ,
	COALESCE(commissionTable.SR * configTable.annualratio / 100 ,0) Annualbonus , configTable.annualratio, BonusTable.tCFee
	FROM public.branchsalary  BS
	inner join (
	  SELECT sum(BonusTable.pbonus) PBonus , sum(BonusTable.lbonus) LBonus, sum(BonusTable.Salary) Salary, sum(commercialfee) tCFee, bsid  FROM public.SalerSalary BonusTable group by bsid
	) BonusTable on BonusTable.bsid = BS.bsid
	left join (
		SELECT sum(cost) prepay , branch FROM public.prepay PP 
		inner join public.BranchPrePay BPP on PP.ppid = BPP.ppid 	
		where  extract(epoch from date) >= $3 and extract(epoch from date) <= $4
		group by branch
	) prepayTable on prepayTable.branch = BS.branch
	left join(
		SELECT sum(fee) pocket , branch FROM public.Pocket 		
		where extract(epoch from date) >= $3 and extract(epoch from date) <= $4
		group by branch
	) pocketTable on pocketTable.branch = BS.branch
	left join(
	    SELECT  branch , sum(cost) thismonthamor FROM public.amortization amor
		inner join (
			SELECT amorid, date, cost FROM public.amormap
			where date >= $2 and date <= $1
		) amormap on amormap.amorid = amor.amorid
		where isover = 0 
		group by  amor.branch		
	) amorTable on amorTable.branch = BS.branch
	left join(
		Select sum(SR) SR , bsid FROM public.commission 
		where bsid is not null
		group by bsid
	) commissionTable on commissionTable.bsid = BS.bsid 
	inner join(
		Select branch, rent, agentsign, commercialfee , annualratio FROM public.configbranch	
	) configTable on configTable.branch = BS.branch 
	left join(
		Select sum(pretax) OVER (partition by branch Order by Date asc) pretaxTotal , branch , Date qq , IE.bsid FROM public.incomeexpense IE
		inner join public.BranchSalary BS on  IE.bsid = BS.bsid
		where date >  $1
	) incomeexpense on incomeexpense.bsid = BS.bsid 	
	where date = $1
	) subtable
	ON CONFLICT (bsid) DO Nothing;
	`

	interdb := salaryM.imr.GetSQLDBwithDbname(dbname)
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}
	defer sqldb.Close()
	////////////
	loc, _ := time.LoadLocation("Asia/Taipei")
	t := bs.LastDate.In(loc)
	lastDate := fmt.Sprintf("%d-%02d-%02d", t.Year(), t.Month(), t.Day())
	///////////
	//fmt.Println("BSID:" + bs.BSid)
	//fmt.Println(bs.Date)
	res, err := sqldb.Exec(sql, bs.StrDate, lastDate, bs.LastDate.Unix(), bs.Date.Unix())
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
	fmt.Println("CreateIncomEexpense:", id, " bs.Date:", bs.StrDate)

	if id == 0 {
		fmt.Println("CreateIncomEexpense, no create anyone ")
		return errors.New("CreateIncomeExpense, no create anyone ")
	}

	return nil
}

func (salaryM *SalaryModel) UpdateCommissionBSidAndStatus(bs *BranchSalary, cid []*Cid, dbname string) (err error) {

	const sql = `Update public.commission as com
				set bsid = subquery.bsid, status = 'join'
				from (
				SELECT c.sid, c.rid, SS.bsid
				FROM public.receipt r
				inner join public.commission c on c.rid = r.rid and c.status = 'normal' and
				extract(epoch from r.date) <= $1 and c.bsid is null
				inner join public.SalerSalary SS on SS.date = to_char(r.date at time zone 'UTC' at time zone 'Asia/Taipei','yyyy-MM') and SS.Sid = C.sid
				) AS subquery
				where com.sid = subquery.sid and com.rid = subquery.rid	;	
				`

	interdb := salaryM.imr.GetSQLDBwithDbname(dbname)
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}
	//fmt.Println("BSID:" + bs.BSid)
	//fmt.Println(bs.Date)
	// b, _ := time.Parse(time.RFC3339, bs.Date+"-01T00:00:00+08:00")
	//fmt.Println("CreateSalerSalary:", bs.Date+"-01 =>", b.Unix())
	//b, _ := time.ParseInLocation("2006-01-02", bs.Date+"-01", time.Local)
	//fmt.Println("UpdateCommissionBSidAndStatus:", bs.Date+"-01 =>", b.Unix())
	res, err := sqldb.Exec(sql, bs.Date.Unix())
	//res, err := sqldb.Exec(sql, salaryM.CloseAccount.CloseDate.Unix())
	if err != nil {
		fmt.Println("[UpdateCommissionBSidAndStatus err] ", err)
		return err
	}
	id, err := res.RowsAffected()
	if err != nil {
		fmt.Println("PG Affecte Wrong: ", err)
		return err
	}
	fmt.Println(id)

	if id == 0 {
		fmt.Println("UpdateCommissionBSidAndStatus, not found any commission")
		//return errors.New("UpdateCommissionBSidAndStatus, not found any commission")
	}
	defer sqldb.Close()
	return nil
}

func (salaryM *SalaryModel) UpdateBranchSalaryTotal(dbname string) (err error) {

	const sql = `UPDATE public.branchsalary BS
				set total = tmp.total
				FROM (
					SELECT sum(total) total, bsid  From public.salersalary					
					group by bsid 
				)as tmp where tmp.bsid = bs.bsid;	
				`
	//where date = $1
	interdb := salaryM.imr.GetSQLDBwithDbname(dbname)
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}
	//fmt.Println("BSID:" + bs.BSid)
	//fmt.Println(bs.Date)
	res, err := sqldb.Exec(sql)
	//res, err := sqldb.Exec(sql, unix_time, receivable.Date, receivable.CNo, receivable.Sales)
	if err != nil {
		fmt.Println("[Update err] ", err)
		return err
	}
	id, err := res.RowsAffected()
	if err != nil {
		fmt.Println("PG Affecte Wrong: ", err)
		return err
	}
	fmt.Println(id)

	if id == 0 {
		fmt.Println("UpdateBranchSalaryTotal, not found any salary")
		return errors.New("UpdateBranchSalaryTotal, not found any salary")
	}
	defer sqldb.Close()
	return nil
}

func (salaryM *SalaryModel) GetSalerSalaryData(bsID, sid, dbname string) []*SalerSalary {
	//( (case when cb.sid is not null then ie.managerbonus else 0 end) + ss.tamount )
	sql := `SELECT ss.sid, ss.bsid, ss.sname, ss.date, ss.branch, ss.salary, ss.pbonus, ss.lbonus, ss.abonus, 
				ss.total, 
				ss.sp, ss.tax, ss.laborfee, ss.healthfee, ss.welfare, ss.CommercialFee, ss.other, 
				 tamount,
				COALESCE(ss.description,''), ss.workday , bs.lock,
				(case when cb.sid is not null then ie.managerbonus else 0 end) managerbonus,
				cs.code, cb.sid
				FROM public.salersalary ss 
				inner join public.branchsalary bs on bs.bsid = ss.bsid
				left join public.incomeexpense ie on ie.bsid = ss.bsid
				left join  public.configbranch cb on cb.branch = ss.branch and ss.sid = cb.sid
				inner join  public.configsaler cs on cs.sid = ss.sid
				where ss.bsid = '%s' order by cs.code`
	//where ss.bsid = '%s' and ss.sid like '%s'`

	db := salaryM.imr.GetSQLDBwithDbname(dbname)
	rows, err := db.SQLCommand(fmt.Sprintf(sql, bsID))
	if err != nil {
		return nil
	}
	var ssDataList []*SalerSalary

	for rows.Next() {
		var ss SalerSalary
		var ManagerID NullString

		if err := rows.Scan(&ss.Sid, &ss.BSid, &ss.SName, &ss.Date, &ss.Branch, &ss.Salary, &ss.Pbonus, &ss.Lbonus, &ss.Abonus, &ss.Total,
			&ss.SP, &ss.Tax, &ss.LaborFee, &ss.HealthFee, &ss.Welfare, &ss.CommercialFee, &ss.Other, &ss.TAmount, &ss.Description, &ss.Workday, &ss.Lock, &ss.ManagerBonus,
			&ss.Code, &ManagerID); err != nil {
			fmt.Println("salaryM err Scan " + err.Error())
			return nil
		}

		ss.ManagerID = ManagerID.Value

		ssDataList = append(ssDataList, &ss)
	}

	// out, err := json.Marshal(ssDataList)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(string(out))
	salaryM.salerSalaryList = ssDataList

	if len(salaryM.salerSalaryList) > 0 {
		ss := salaryM.salerSalaryList[0]
		strtime, _ := util.ADtoROC(ss.Date, "file")
		salaryM.FnamePdf = ss.Branch + "薪資表" + strtime

	}

	return salaryM.salerSalaryList
}

func (salaryM *SalaryModel) GetIncomeExpenseData(bsID, dbname string) []*IncomeExpense {

	const spl = `SELECT bsid, sr, businesstax, salesamounts, pbonus, lbonus, amorcost, agentsign, rent, commercialfee, salary, prepay, pocket, annualbonus, salerfee, pretax, aftertax, earnadjust, lastloss, businessincometax, managerbonus, annualratio
	FROM public.incomeexpense where bsid = '%s';`
	db := salaryM.imr.GetSQLDBwithDbname(dbname)
	rows, err := db.SQLCommand(fmt.Sprintf(spl, bsID))
	if err != nil {
		return nil
	}
	var IEDataList []*IncomeExpense

	for rows.Next() {
		var ie IncomeExpense

		if err := rows.Scan(&ie.BSid, &ie.Income.SR, &ie.Income.Businesstax, &ie.Income.Salesamounts, &ie.Expense.Pbonus, &ie.Expense.LBonus, &ie.Expense.Amorcost, &ie.Expense.Agentsign,
			&ie.Expense.Rent, &ie.Expense.Commercialfee, &ie.Expense.Salary, &ie.Expense.Prepay, &ie.Expense.Pocket, &ie.Expense.Annualbonus, &ie.Expense.SalerFee, &ie.Pretax,
			&ie.Aftertax, &ie.EarnAdjust, &ie.Lastloss, &ie.BusinessIncomeTax, &ie.ManagerBonus, &ie.Expense.AnnualRatio); err != nil {
			fmt.Println("err Scan " + err.Error())
			return nil
		}

		IEDataList = append(IEDataList, &ie)
	}

	salaryM.IncomeExpenseList = IEDataList
	return salaryM.IncomeExpenseList
}

func (salaryM *SalaryModel) GetNHISalaryData(bsID, dbname string) []*NHISalary {

	fmt.Println("GetNHISalaryData")
	const spl = `SELECT nhi.sid, nhi.bsid, nhi.sname, nhi.payrollbracket, nhi.salary, nhi.pbonus, nhi.bonus, nhi.total, nhi.pd, nhi.salarybalance, nhi.fourbouns, ss.sp, nhi.foursp, nhi.ptsp, cs.code
				FROM public.nhisalary nhi
				inner join (
					select bsid, sid, sp from public.salersalary 
				) ss on ss.bsid = nhi.bsid and ss.sid = nhi.sid
				inner join  public.configsaler cs on cs.sid = ss.sid				
				where nhi.bsid = '%s' order by cs.code`

	db := salaryM.imr.GetSQLDBwithDbname(dbname)
	rows, err := db.SQLCommand(fmt.Sprintf(spl, bsID))
	if err != nil {
		fmt.Println(err)
		return nil
	}
	var nhiDataList []*NHISalary

	for rows.Next() {
		var nhi NHISalary

		if err := rows.Scan(&nhi.Sid, &nhi.BSid, &nhi.SName, &nhi.PayrollBracket, &nhi.Salary, &nhi.Pbonus, &nhi.Bonus, &nhi.Total,
			&nhi.PD, &nhi.SalaryBalance, &nhi.FourBouns, &nhi.SP, &nhi.FourSP, &nhi.PTSP, &nhi.Code); err != nil {
			fmt.Println("err Scan " + err.Error())
			return nil
		}

		nhiDataList = append(nhiDataList, &nhi)
	}

	// out, err := json.Marshal(arList)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(string(out))
	salaryM.NHISalaryList = nhiDataList
	return salaryM.NHISalaryList
}

func (salaryM *SalaryModel) ExportNHISalaryData(bsID, dbname string) []*NHISalary {

	fmt.Println("ExportNHISalaryData")
	const spl = `SELECT nhi.sid, nhi.bsid, nhi.sname, nhi.payrollbracket, nhi.salary, nhi.pbonus, nhi.bonus, nhi.total, nhi.pd, nhi.salarybalance, nhi.fourbouns, ss.sp, nhi.foursp, nhi.ptsp, cs.title , coalesce(ss.description,''), cs.branch
				FROM public.nhisalary nhi
				inner join (
					select sid, bsid, sp ,description from public.salersalary 
				) ss on ss.bsid = nhi.bsid and ss.sid = nhi.sid
				inner join(
					SELECT A.sid, A.sname, A.branch, A.percent, A.title, A.code
					FROM public.ConfigSaler A 					
				) cs on cs.sid = nhi.sid
				where nhi.bsid = '%s' order by cs.code`

	db := salaryM.imr.GetSQLDBwithDbname(dbname)
	rows, err := db.SQLCommand(fmt.Sprintf(spl, bsID))
	if err != nil {
		fmt.Println(err)
		return nil
	}

	for rows.Next() {
		var nhi NHISalary

		if err := rows.Scan(&nhi.Sid, &nhi.BSid, &nhi.SName, &nhi.PayrollBracket, &nhi.Salary, &nhi.Pbonus, &nhi.Bonus, &nhi.Total,
			&nhi.PD, &nhi.SalaryBalance, &nhi.FourBouns, &nhi.SP, &nhi.FourSP, &nhi.PTSP, &nhi.Title, &nhi.Description, &nhi.Branch); err != nil {
			fmt.Println("err Scan " + err.Error())
			return nil
		}

		salaryM.NHISalaryList = append(salaryM.NHISalaryList, &nhi)
	}

	return salaryM.NHISalaryList
}

func (salaryM *SalaryModel) CreateNHISalary(year, dbname string) (err error) {

	const sql = `INSERT INTO public.nhisalary
	(sid, bsid, sname, payrollbracket, salary, pbonus, bonus, total, salarybalance, pd, fourbouns, sp, foursp, ptsp)
	SELECT  SS.sid, SS.BSid, SS.Sname, CS.PayrollBracket, SS.Salary, SS.Pbonus, (SS.Lbonus + ie.managerbonus - SS.abonus) bonus, 
	(SS.Salary + SS.Pbonus + (SS.Lbonus + ie.managerbonus  - SS.abonus) ) Total ,
	( (SS.Salary + SS.Pbonus + (SS.Lbonus + ie.managerbonus  - SS.abonus) ) - CS.PayrollBracket) SalaryBalance,
	sum( (SS.Salary + SS.Pbonus + (SS.Lbonus + ie.managerbonus  - SS.abonus) ) - CS.PayrollBracket) over (partition by SS.year,SS.sid order by SS.date) PD ,
	(CS.PayrollBracket * 4) FourBouns, 0 SP,
	(CASE WHEN sum(SS.Total - CS.PayrollBracket) over (partition by SS.year,SS.sid order by SS.date) - (CS.PayrollBracket * 4) > 0 then (SS.Total - CS.PayrollBracket) *CP.nhi2nd /100  ELSE 0 END ) FourSP, 
	(CASE WHEN CS.association = 0 and CS.PayrollBracket <=0 then SS.Total*CP.nhi2nd ELSE 0 END ) PTSP  
	FROM public.salersalary SS
		Inner Join (
			select A.Sid, A.PayrollBracket, A.association  FROM public.ConfigSaler A 
			Inner Join ( 
				select sid, max(zerodate) zerodate from public.configsaler cs 
				where now() > zerodate
				group by sid 
			) B on A.sid=B.sid and A.zeroDate = B.zeroDate
		) CS on SS.sid = CS.sid
		cross join (
			select  c.date, c.nhi, c.li, c.nhi2nd, c.mmw from public.ConfigParameter C
			inner join(
				select  max(date) date from public.ConfigParameter 
			) A on A.date = C.date limit 1
		) CP
		inner join (
			Select bsid , managerbonus from public.incomeexpense 
		) ie on ie.bsid = ss.bsid
		where year = $1
	ON CONFLICT (bsid,sid) DO Nothing ;` //UPDATE SET balance = excluded.balance;`

	interdb := salaryM.imr.GetSQLDBwithDbname(dbname)
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}
	//fmt.Println("BSID:" + bs.BSid)
	//fmt.Println(bs.Date)
	res, err := sqldb.Exec(sql, year)
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
	fmt.Println("CreateNHISalary:", id)

	if id == 0 {
		fmt.Println("CreateNHISalary, not found any salary ")
	}
	defer sqldb.Close()
	return nil
}

func (salaryM *SalaryModel) UpdateSalerSalaryData(ss *SalerSalary, bsid, dbname string) (err error) {
	const sql = `UPDATE public.salersalary
	SET lbonus= $1, abonus= $2, total= subquery.msalary + pbonus + $1 - $2, tax = $3, other = $4,  description= $5, workday= $6,
	laborfee =  $12,
	healthfee = $13,
	commercialFee =  ROUND( (salary + pbonus + $1 - $2) * subquery.commercialRatio/100 ) ,
	salary = $11 ,
	sp = $9 , welfare = $10 ,
	tamount = subquery.msalary + pbonus + $1 - $2 - $3 - $4 - $9 - $10 - (subquery.msalary + pbonus + $1 - $2)* subquery.commercialRatio /100 - $12::integer - $13::integer
	FROM(
		Select ROUND( Case When $6 >= 30 then $11 else $11::float * $6 / 30 END)::integer msalary, commercialFee as commercialRatio FROM public.ConfigSaler A
		left join(
			select branch , commercialFee from public.configbranch
		) CB on CB.branch = A.branch
		WHERE sid= $7
	) as subquery
	WHERE sid= $7 and bsid = $8;`

	/*	const sql = `UPDATE public.salersalary
		SET lbonus= $1, abonus= $2, total= subquery.msalary + pbonus + $1 - $2, tax = $3, other = $4,  description= $5, workday= $6,
		laborfee = ( Case When $6 >= 30 then subquery.laborfee else subquery.laborfee * $6 / 30 END),
		healthfee = ( Case When $6 >= 30 then subquery.healthfee else 0 END),
		commercialFee =  ROUND( (salary + pbonus + $1 - $2) * subquery.commercialRatio/100 ) ,
		salary = subquery.msalary ,
		sp = $9 , welfare = $10 ,
		tamount = subquery.msalary + pbonus + $1 - $2 - $3 - $4 - $9 - $10 - (subquery.msalary + pbonus + $1 - $2)* subquery.commercialRatio /100 - ( Case When $6 >= 30 then subquery.laborfee else subquery.laborfee * $6 / 30 END) - ( Case When $6 >= 30 then subquery.healthfee else 0 END)
		FROM(
			Select ROUND( Case When $6 >= 30 then salary else salary::float * $6 / 30 END)::integer msalary, commercialFee as commercialRatio, ROUND(A.payrollbracket * CP.li * 0.2 / 100) laborfee, ROUND(A.payrollbracket * CP.nhi * 0.3 / 100) healthfee FROM public.ConfigSaler A
			cross join (
				select  c.date, c.nhi, c.li, c.nhi2nd, c.mmw from public.ConfigParameter C
				inner join(
					select  max(date) date from public.ConfigParameter
				) A on A.date = C.date limit 1
			) CP
			left join(
				select branch , commercialFee from public.configbranch
			) CB on CB.branch = A.branch
			WHERE sid= $7
		) as subquery
		WHERE sid= $7 and bsid = $8;`
	*/
	interdb := salaryM.imr.GetSQLDBwithDbname(dbname)
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}
	//fmt.Println("BSID:" + bs.BSid)
	//fmt.Println(bs.Date)
	res, err := sqldb.Exec(sql, ss.Lbonus, ss.Abonus, ss.Tax, ss.Other, ss.Description, ss.Workday, ss.Sid, bsid, ss.SP, ss.Welfare, ss.Salary, ss.LaborFee, ss.HealthFee)
	//res, err := sqldb.Exec(sql, ss.Lbonus, ss.Abonus, ss.Tax, ss.Other, ss.Description, ss.Workday, ss.Sid, bsid, ss.SP, ss.Welfare)
	//res, err := sqldb.Exec(sql, unix_time, receivable.Date, receivable.CNo, receivable.Sales)
	if err != nil {
		fmt.Println("[Update err] ", err)
		return err
	}
	id, err := res.RowsAffected()
	if err != nil {
		fmt.Println("PG Affecte Wrong: ", err)
		return err
	}
	fmt.Println(id)

	if id == 0 {
		fmt.Println("UpdateSalerSalaryData, not found any salary")
		return errors.New("UpdateSalerSalaryData, not found any salary")
	}
	defer sqldb.Close()
	return nil
}

func (salaryM *SalaryModel) UpdateIncomeExpenseData(ie *IncomeExpense, bsid, dbname string) (err error) {

	//annualBonus 是上筆算好的，sr * $4 / 100是用新的AnnualRatio算出新的annualBonus
	const sql = `UPDATE public.incomeExpense
	SET salerfee = $2 , earnadjust = $3::integer , pretax = (pretax + salerFee - $2 + annualBonus - sr * $4 / 100) , annualratio = $4 , annualBonus = sr * $4 / 100 ,
	businessincometax = (CASE WHEN (pretax + salerFee - $2 + annualBonus - sr * $4 / 100) * 0.2 > 0 then (pretax + salerFee - $2 + annualBonus - sr * $4 / 100) * 0.2 else 0 end ),
	aftertax = (pretax + salerFee - $2 + annualBonus - sr * $4 / 100) - (CASE WHEN (pretax + salerFee - $2 + annualBonus - sr * $4 / 100) * 0.2 > 0 then (pretax + salerFee - $2 + annualBonus - sr * $4 / 100) * 0.2 else 0 end ) , 
	managerbonus = (CASE WHEN (
		((pretax + salerFee - $2 + annualBonus - sr * $4 / 100) - (CASE WHEN (pretax + salerFee - $2 + annualBonus - sr * $4 / 100) * 0.2 > 0 then (pretax + salerFee - $2 + annualBonus - sr * $4 / 100) * 0.2 else 0 end ) + lastLoss + $3 ) > 0
	) then (
		((pretax + salerFee - $2 + annualBonus - sr * $4 / 100 - $2) - (CASE WHEN (pretax + salerFee - $2 + annualBonus - sr * $4 / 100 - $2) * 0.2 > 0 then (pretax + salerFee - $2 + annualBonus - sr * $4 / 100 - $2) * 0.2 else 0 end ) + lastLoss + $3 ) * 0.2
	)
	ELSE 0 END)
	WHERE bsid = $1;`

	interdb := salaryM.imr.GetSQLDBwithDbname(dbname)
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}
	fmt.Println(ie.EarnAdjust)
	fmt.Println(ie.Expense.SalerFee)
	res, err := sqldb.Exec(sql, bsid, ie.Expense.SalerFee, ie.EarnAdjust, ie.Expense.AnnualRatio)
	//res, err := sqldb.Exec(sql, unix_time, receivable.Date, receivable.CNo, receivable.Sales)
	if err != nil {
		fmt.Println("[Update err] ", err)
		return err
	}
	id, err := res.RowsAffected()
	if err != nil {
		fmt.Println("PG Affecte Wrong: ", err)
		return err
	}
	fmt.Println("[UpdateIncomeExpenseData] ", id)

	if id == 0 {
		fmt.Println("UpdateIncomeExpenseData, not found any salary")
		return errors.New("UpdateIncomeExpenseData, not found any salary")
	}

	ummb_err := salaryM.UpdateManagerByManagerBonus(bsid, dbname)
	if ummb_err != nil {
		return nil
		//return ucias_err
	}
	defer sqldb.Close()
	return nil
}

//[專門]針對因為更改紅利表，所以更新店長的薪資
func (salaryM *SalaryModel) UpdateManagerByManagerBonus(bsid, dbname string) (err error) {
	//welfare = subquery.Total * 0.01,
	const sql = `UPDATE public.salersalary salersalary
	SET total= subquery.Total , 
		sp = subquery.sp ,		
		commercialfee = subquery.Total * subquery.CommercialFee / 100 ,
		tamount = subquery.Total - subquery.sp - subquery.tax - subquery.laborfee - subquery.healthfee - subquery.welfare - subquery.Total * subquery.CommercialFee / 100 - subquery.other		
	FROM(
	SELECT ss.sid, ss.bsid, ss.sname, ss.date, ss.branch, ss.salary, ss.pbonus, ss.lbonus, ss.abonus, 
	 ss.salary + ss.pbonus + ss.lbonus - ss.abonus + (case when cb.sid is not null then ie.managerbonus else 0 end)	Total, 
	(CASE WHEN ss.salary = 0 and cs.association = 1 then 0 
		WHEN ss.salary = 0 and cs.association = 0 then COALESCE(ss.Salary+  ss.Pbonus,ss.Salary) * cp.nhi2nd / 100 
		else
			( CASE WHEN ((COALESCE(ss.Salary+  ss.Pbonus, ss.Salary)) - 4 * cs.PayrollBracket) > 0 then ((COALESCE(ss.Salary+  ss.Pbonus,ss.Salary)) - 4 * cs.PayrollBracket) * cp.nhi2nd / 100 else 0 end)
		end
	) sp,
	ss.tax, ss.laborfee, ss.healthfee, ss.welfare, cb.CommercialFee, ss.other, 
	COALESCE(ss.description,''), ss.workday , bs.lock,
	(case when cb.sid is not null then ie.managerbonus else 0 end) managerbonus
	FROM public.salersalary ss 
	inner join public.branchsalary bs on bs.bsid = ss.bsid
	left join public.incomeexpense ie on ie.bsid = ss.bsid
	inner join public.configbranch cb on cb.branch = ss.branch and ss.sid = cb.sid
	inner join (
		select A.sid , A.PayrollBracket, A.association from public.ConfigSaler A
		Inner Join ( 
			select sid, max(zerodate) zerodate from public.configsaler 
			where now() > zerodate
			group by sid 
		) B on A.sid=B.sid and A.zeroDate = B.zeroDate
	) cs on ss.sid = cs.sid 
	cross join (
		select  c.date, c.nhi, c.li, c.nhi2nd, c.mmw from public.ConfigParameter C
		inner join(
			select  max(date) date from public.ConfigParameter 
		) A on A.date = C.date limit 1
	) CP
	where ss.bsid = $1
	) as subquery
	WHERE salersalary.bsid = $1 and salersalary.sid = subquery.sid`

	interdb := salaryM.imr.GetSQLDBwithDbname(dbname)
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}

	res, err := sqldb.Exec(sql, bsid)
	//res, err := sqldb.Exec(sql, unix_time, receivable.Date, receivable.CNo, receivable.Sales)
	if err != nil {
		fmt.Println("[Update err] ", err)
		return err
	}
	id, err := res.RowsAffected()
	if err != nil {
		fmt.Println("PG Affecte Wrong: ", err)
		return err
	}
	fmt.Println("[UpdateManagerByManagerBonus] ", id)

	if id == 0 {
		fmt.Println("UpdateManagerByManagerBonus, not found any salary")
		return errors.New("UpdateManagerByManagerBonus, not found any salary")
	}
	defer sqldb.Close()
	return nil
}

func (salaryM *SalaryModel) LockBranchSalary(bsid, lock, dbname string) (err error) {

	const sql = `UPDATE public.branchsalary	SET lock = $2	WHERE bsid = $1;`

	interdb := salaryM.imr.GetSQLDBwithDbname(dbname)
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}

	res, err := sqldb.Exec(sql, bsid, lock)
	//res, err := sqldb.Exec(sql, unix_time, receivable.Date, receivable.CNo, receivable.Sales)
	if err != nil {
		fmt.Println("[Update err] ", err)
		return err
	}
	id, err := res.RowsAffected()
	if err != nil {
		fmt.Println("PG Affecte Wrong: ", err)
		return err
	}
	fmt.Println(id)

	if id == 0 {
		fmt.Println("LockBranchSalary, not found any salary")
		return errors.New("LockBranchSalary, not found any salary")
	}

	_err := salaryM.UpdateBranchSalaryTotal(dbname)
	if _err != nil {
		return nil
		//return css_err
	}
	defer sqldb.Close()
	return nil
}

//加入pdf資料 type=1
func (salaryM *SalaryModel) addBranchSalaryInfoTable(table *pdf.DataTable, p *pdf.Pdf) (table_final *pdf.DataTable,
	T_Salary, T_Pbonus, T_Abonus, T_Lbonus, T_Total, T_SP, T_Tax, T_LaborFee, T_HealthFee, T_Welfare, T_CommercialFee, T_TAmount, T_Other int) {

	//text := "fd"
	//width := mypdf.MeasureTextWidth(text)
	//table.ColumnLen
	T_Salary, T_Pbonus, T_Abonus, T_Lbonus, T_Total = 0, 0, 0, 0, 0
	T_SP, T_Tax, T_LaborFee, T_HealthFee, T_Welfare, T_CommercialFee = 0, 0, 0, 0, 0, 0
	T_TAmount, T_Other = 0, 0

	for index, element := range salaryM.salerSalaryList {
		//fmt.Println("addBranchSalaryInfoTable:", table.ColumnWidth[index])
		//建立千分位
		pr := message.NewPrinter(language.English)

		text := strconv.Itoa(index + 1)
		pdf.ResizeWidth(table, p.GetTextWidth(text), 0)
		var vs = &pdf.TableStyle{
			Text:  text,
			Bg:    pdf.ColorWhite,
			Front: pdf.ColorTableLine,
		}
		table.RawData = append(table.RawData, vs)
		//
		text = element.SName
		pdf.ResizeWidth(table, p.GetTextWidth(text), 1)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    pdf.ColorWhite,
			Front: pdf.ColorTableLine,
		}
		table.RawData = append(table.RawData, vs)
		//底薪
		T_Salary += element.Salary
		//text = strconv.Itoa(element.Salary)
		text = pr.Sprintf("%d", element.Salary)
		pdf.ResizeWidth(table, p.GetTextWidth(text), 2)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    pdf.ColorWhite,
			Front: pdf.ColorTableLine,
			Align: pdf.AlignRight,
		}
		table.RawData = append(table.RawData, vs)
		//績效
		T_Pbonus += element.Pbonus
		//text = strconv.Itoa(element.Pbonus)
		text = pr.Sprintf("%d", element.Pbonus)
		pdf.ResizeWidth(table, p.GetTextWidth(text), 3)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    pdf.ColorWhite,
			Front: pdf.ColorTableLine,
			Align: pdf.AlignRight,
		}
		table.RawData = append(table.RawData, vs)
		//領導
		T_Lbonus += element.Lbonus
		//text = strconv.Itoa(element.Lbonus)
		text = pr.Sprintf("%d", element.Lbonus)
		pdf.ResizeWidth(table, p.GetTextWidth(text), 4)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    pdf.ColorWhite,
			Front: pdf.ColorTableLine,
			Align: pdf.AlignRight,
		}
		table.RawData = append(table.RawData, vs)
		//
		T_Abonus += element.Abonus
		//text = strconv.Itoa(element.Abonus)
		text = pr.Sprintf("%d", element.Abonus)
		pdf.ResizeWidth(table, p.GetTextWidth(text), 5)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    pdf.ColorWhite,
			Front: pdf.ColorTableLine,
			Align: pdf.AlignRight,
		}
		table.RawData = append(table.RawData, vs)
		//
		T_Total += element.Total
		//text = strconv.Itoa(element.Total)
		text = pr.Sprintf("%d", element.Total)
		pdf.ResizeWidth(table, p.GetTextWidth(text), 6)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    If(true, pdf.ColorBlack, pdf.ColorWhite).(pdf.Color),
			Front: pdf.ColorTableLine,
			Align: pdf.AlignRight,
		}
		table.RawData = append(table.RawData, vs)
		//
		T_SP += element.SP
		//text = strconv.Itoa(element.SP)
		text = pr.Sprintf("%d", element.SP)
		pdf.ResizeWidth(table, p.GetTextWidth(text), 7)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    If(true, pdf.ColorWhite, pdf.ColorWhite).(pdf.Color),
			Front: pdf.ColorTableLine,
			Align: pdf.AlignRight,
		}
		table.RawData = append(table.RawData, vs)
		//
		T_Tax += element.Tax
		//text = strconv.Itoa(element.Tax)
		text = pr.Sprintf("%d", element.Tax)
		pdf.ResizeWidth(table, p.GetTextWidth(text), 8)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    If(true, pdf.ColorWhite, pdf.ColorWhite).(pdf.Color),
			Front: pdf.ColorTableLine,
			Align: pdf.AlignRight,
		}
		table.RawData = append(table.RawData, vs)
		//
		T_LaborFee += element.LaborFee
		//text = strconv.Itoa(element.LaborFee)
		text = pr.Sprintf("%d", element.LaborFee)
		pdf.ResizeWidth(table, p.GetTextWidth(text), 9)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    If(true, pdf.ColorWhite, pdf.ColorWhite).(pdf.Color),
			Front: pdf.ColorTableLine,
			Align: pdf.AlignRight,
		}
		table.RawData = append(table.RawData, vs)
		//
		T_HealthFee += element.HealthFee
		//text = strconv.Itoa(element.HealthFee)
		text = pr.Sprintf("%d", element.HealthFee)
		pdf.ResizeWidth(table, p.GetTextWidth(text), 10)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    If(true, pdf.ColorWhite, pdf.ColorWhite).(pdf.Color),
			Front: pdf.ColorTableLine,
			Align: pdf.AlignRight,
		}
		table.RawData = append(table.RawData, vs)
		//
		T_Welfare += element.Welfare
		text = pr.Sprintf("%d", element.LaborFee)
		text = strconv.Itoa(element.Welfare)
		pdf.ResizeWidth(table, p.GetTextWidth(text), 11)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    If(true, pdf.ColorWhite, pdf.ColorWhite).(pdf.Color),
			Front: pdf.ColorTableLine,
			Align: pdf.AlignRight,
		}
		table.RawData = append(table.RawData, vs)
		//
		T_CommercialFee += +element.CommercialFee
		text = pr.Sprintf("%d", element.CommercialFee)
		pdf.ResizeWidth(table, p.GetTextWidth(text), 12)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    If(true, pdf.ColorWhite, pdf.ColorWhite).(pdf.Color),
			Front: pdf.ColorTableLine,
			Align: pdf.AlignRight,
		}

		table.RawData = append(table.RawData, vs)
		//
		T_Other += element.Other
		text = pr.Sprintf("%d", element.Other)
		pdf.ResizeWidth(table, p.GetTextWidth(text), 13)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    If(true, pdf.ColorWhite, pdf.ColorWhite).(pdf.Color),
			Front: pdf.ColorTableLine,
			Align: pdf.AlignRight,
		}
		table.RawData = append(table.RawData, vs)
		//
		T_TAmount += element.TAmount
		text = pr.Sprintf("%d", element.TAmount)
		pdf.ResizeWidth(table, p.GetTextWidth(text), 14)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    If(true, pdf.ColorWhite, pdf.ColorWhite).(pdf.Color),
			Front: pdf.ColorTableLine,
			Align: pdf.AlignRight,
		}
		table.RawData = append(table.RawData, vs)
		//
		text = element.Description
		pdf.ResizeWidth(table, p.GetTextWidth(text), 15)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    If(true, pdf.ColorWhite, pdf.ColorWhite).(pdf.Color),
			Front: pdf.ColorTableLine,
			Align: pdf.AlignRight,
		}
		table.RawData = append(table.RawData, vs)
	}
	table_final = table
	return
}

func (salaryM *SalaryModel) addSalerSalaryInfoTable(table *pdf.DataTable, p *pdf.Pdf, index int, element *SalerSalary) (table_final *pdf.DataTable) {

	fmt.Println(index)

	///
	text := strconv.Itoa(index + 1)
	pdf.ResizeWidth(table, p.GetTextWidth(text), 0)
	var vs = &pdf.TableStyle{
		Text:  text,
		Bg:    pdf.ColorWhite,
		Front: pdf.ColorTableLine,
	}
	table.RawData = append(table.RawData, vs)
	//
	text = element.SName
	pdf.ResizeWidth(table, p.GetTextWidth(text), 1)
	vs = &pdf.TableStyle{
		Text:  text,
		Bg:    pdf.ColorWhite,
		Front: pdf.ColorTableLine,
	}
	table.RawData = append(table.RawData, vs)
	//底薪

	text = pr.Sprintf("%d", element.Salary)
	pdf.ResizeWidth(table, p.GetTextWidth(text), 2)
	vs = &pdf.TableStyle{
		Text:  text,
		Bg:    pdf.ColorWhite,
		Front: pdf.ColorTableLine,
		Align: pdf.AlignRight,
	}
	table.RawData = append(table.RawData, vs)
	//績效
	text = pr.Sprintf("%d", element.Pbonus)
	pdf.ResizeWidth(table, p.GetTextWidth(text), 3)
	vs = &pdf.TableStyle{
		Text:  text,
		Bg:    pdf.ColorWhite,
		Front: pdf.ColorTableLine,
		Align: pdf.AlignRight,
	}
	table.RawData = append(table.RawData, vs)
	//領導
	text = pr.Sprintf("%d", element.Lbonus)
	pdf.ResizeWidth(table, p.GetTextWidth(text), 4)
	vs = &pdf.TableStyle{
		Text:  text,
		Bg:    pdf.ColorWhite,
		Front: pdf.ColorTableLine,
		Align: pdf.AlignRight,
	}
	table.RawData = append(table.RawData, vs)
	//

	text = pr.Sprintf("%d", element.Abonus)
	pdf.ResizeWidth(table, p.GetTextWidth(text), 5)
	vs = &pdf.TableStyle{
		Text:  text,
		Bg:    pdf.ColorWhite,
		Front: pdf.ColorTableLine,
		Align: pdf.AlignRight,
	}
	table.RawData = append(table.RawData, vs)
	//

	text = pr.Sprintf("%d", element.Total)
	pdf.ResizeWidth(table, p.GetTextWidth(text), 6)
	vs = &pdf.TableStyle{
		Text:  text,
		Bg:    If(true, pdf.ColorWhite, pdf.ColorWhite).(pdf.Color),
		Front: pdf.ColorTableLine,
		Align: pdf.AlignRight,
	}
	table.RawData = append(table.RawData, vs)
	//

	text = pr.Sprintf("%d", element.SP)
	pdf.ResizeWidth(table, p.GetTextWidth(text), 7)
	vs = &pdf.TableStyle{
		Text:  text,
		Bg:    If(true, pdf.ColorWhite, pdf.ColorWhite).(pdf.Color),
		Front: pdf.ColorTableLine,
		Align: pdf.AlignRight,
	}
	table.RawData = append(table.RawData, vs)
	//

	text = pr.Sprintf("%d", element.Tax)
	pdf.ResizeWidth(table, p.GetTextWidth(text), 8)
	vs = &pdf.TableStyle{
		Text:  text,
		Bg:    If(true, pdf.ColorWhite, pdf.ColorWhite).(pdf.Color),
		Front: pdf.ColorTableLine,
		Align: pdf.AlignRight,
	}
	table.RawData = append(table.RawData, vs)
	//

	text = pr.Sprintf("%d", element.LaborFee)
	pdf.ResizeWidth(table, p.GetTextWidth(text), 9)
	vs = &pdf.TableStyle{
		Text:  text,
		Bg:    If(true, pdf.ColorWhite, pdf.ColorWhite).(pdf.Color),
		Front: pdf.ColorTableLine,
		Align: pdf.AlignRight,
	}
	table.RawData = append(table.RawData, vs)
	//

	text = pr.Sprintf("%d", element.HealthFee)
	pdf.ResizeWidth(table, p.GetTextWidth(text), 10)
	vs = &pdf.TableStyle{
		Text:  text,
		Bg:    If(true, pdf.ColorWhite, pdf.ColorWhite).(pdf.Color),
		Front: pdf.ColorTableLine,
		Align: pdf.AlignRight,
	}
	table.RawData = append(table.RawData, vs)
	//

	text = pr.Sprintf("%d", element.Welfare)
	pdf.ResizeWidth(table, p.GetTextWidth(text), 11)
	vs = &pdf.TableStyle{
		Text:  text,
		Bg:    If(true, pdf.ColorWhite, pdf.ColorWhite).(pdf.Color),
		Front: pdf.ColorTableLine,
		Align: pdf.AlignRight,
	}
	table.RawData = append(table.RawData, vs)
	//

	text = pr.Sprintf("%d", element.CommercialFee)
	pdf.ResizeWidth(table, p.GetTextWidth(text), 12)
	vs = &pdf.TableStyle{
		Text:  text,
		Bg:    If(true, pdf.ColorWhite, pdf.ColorWhite).(pdf.Color),
		Front: pdf.ColorTableLine,
		Align: pdf.AlignRight,
	}
	table.RawData = append(table.RawData, vs)
	//

	text = pr.Sprintf("%d", element.Other)
	pdf.ResizeWidth(table, p.GetTextWidth(text), 13)
	vs = &pdf.TableStyle{
		Text:  text,
		Bg:    If(true, pdf.ColorWhite, pdf.ColorWhite).(pdf.Color),
		Front: pdf.ColorTableLine,
		Align: pdf.AlignRight,
	}
	table.RawData = append(table.RawData, vs)
	//

	text = pr.Sprintf("%d", element.TAmount)
	pdf.ResizeWidth(table, p.GetTextWidth(text), 14)
	vs = &pdf.TableStyle{
		Text:  text,
		Bg:    If(true, pdf.ColorWhite, pdf.ColorWhite).(pdf.Color),
		Front: pdf.ColorTableLine,
		Align: pdf.AlignRight,
	}

	table.RawData = append(table.RawData, vs)
	//
	text = element.Description
	pdf.ResizeWidth(table, p.GetTextWidth(text), 15)
	vs = &pdf.TableStyle{
		Text:  text,
		Bg:    If(true, pdf.ColorWhite, pdf.ColorWhite).(pdf.Color),
		Front: pdf.ColorTableLine,
	}
	table.RawData = append(table.RawData, vs)
	//}
	table_final = table
	return
}

func (salaryM *SalaryModel) ExportSR(bsID, dbname string) {
	const qsql = `SELECT ss.sid, ss.sname ,  coalesce(sum(tmp.SR),0)  ,coalesce( sum( tmp.SR * cs.percent/100)  , 0 ) bonus , cs.branch , ss.date
	from salersalary ss
		left join(
			SELECT c.bsid, c.sid, c.rid,  (r.amount * c.cpercent/100 - coalesce(c.fee,0)) sr					
			FROM public.commission c
			inner JOIN public.receipt r on r.rid = c.rid		
			left join(
				select rid, fee from public.deduct
			) d on d.rid = r.rid		
		) tmp on ss.bsid = tmp.bsid and tmp.sid = ss.sid
	Inner Join (
		SELECT A.sid, A.branch, A.percent, A.code
			FROM public.ConfigSaler A			
	) cs on cs.sid=ss.sid 
	where ss.bsid = '%s'
	group by cs.branch , ss.date, ss.sid, ss.sname , cs.code
	order by cs.code `

	db := cm.imr.GetSQLDBwithDbname(dbname)

	cDataList := []*Commission{}
	//salerList := []*SystemAccount{}

	rows, err := db.SQLCommand(fmt.Sprintf(qsql, bsID))
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("SQLCommand Done")
	var date = ""
	for rows.Next() {
		var c Commission
		// var Item, Branch, DedectItem NullString
		// var Amount, Fee NullInt
		var SR, Bonus NullFloat
		if err := rows.Scan(&c.Sid, &c.SName, &SR, &Bonus, &c.Branch, &date); err != nil {
			fmt.Println("err Scan " + err.Error())
		}
		c.SR = SR.Value
		c.Bonus = Bonus.Value
		cDataList = append(cDataList, &c)
	}
	if len(cDataList) > 0 {
		//systemM.GetAccountData()
		//salerList = systemM.systemAccountList
		date, _ = util.ADtoROC(date, "file")
		salaryM.FnamePdf = cDataList[0].Branch + "實績分配表" + date

	}
	// else {
	// 	salerList = nil
	// }

	// out, _ := json.Marshal(salerList)
	// fmt.Println("salerList :", string(out))
	// out, _ = json.Marshal(cDataList)
	// fmt.Println("cDataList :", string(out))
	//salaryM.SystemAccountList = salerList
	salaryM.CommissionList = cDataList
	return

}

func (salaryM *SalaryModel) GetSalerCommission(bsID, dbname string) {
	const qsql = `SELECT ss.sid, ss.sname , tmp.item, tmp.amount, tmp.fee, tmp.cpercent, tmp.sr, (tmp.sr * cs.percent/100) bonus , tmp.remark , cs.branch, tmp.mdate  from salersalary ss
				left join(
					SELECT c.bsid, c.sid, c.rid, r.date, (c.item || ' ' || ar.name) item, r.amount, 0 , c.sname, c.cpercent, ( r.amount * c.cpercent/100 - coalesce(c.fee,0)) sr, 
					r.arid, c.status ,  to_char(r.date at time zone 'UTC' at time zone 'Asia/Taipei','yyyy-MM-dd') mdate, COALESCE(NULLIF(iv.invoiceno, null),'') , coalesce(d.checknumber,'') , coalesce(c.fee,0) fee , coalesce(d.item,'') remark
					FROM public.commission c
					inner JOIN public.receipt r on r.rid = c.rid		
					inner JOIN public.ar ar on ar.arid = c.arid	
					left join(
						select rid, checknumber , fee, item from public.deduct
					) d on d.rid = r.rid		
					left join(
						select rid,  invoiceno from public.invoice 
					) iv on r.rid = iv.rid
				) tmp on ss.bsid = tmp.bsid and tmp.sid = ss.sid
			Inner Join (
				SELECT A.sid, A.branch, A.percent, A.title,A.code
					FROM public.ConfigSaler A					
			) cs on cs.sid=ss.sid 			
			where ss.bsid = '%s' order by code,mdate, sid asc;`

	db := cm.imr.GetSQLDBwithDbname(dbname)

	cDataList := []*Commission{}

	rows, err := db.SQLCommand(fmt.Sprintf(qsql, bsID))
	if err != nil {
		fmt.Println(err)
		return
	}

	for rows.Next() {
		var c Commission
		var mdate, Item, Branch, DedectItem NullString
		var Amount, Fee NullInt
		var CPercent, SR, Bonus NullFloat
		if err := rows.Scan(&c.Sid, &c.SName, &Item, &Amount, &Fee, &CPercent, &SR, &Bonus, &DedectItem, &Branch, &mdate); err != nil {
			fmt.Println("err Scan " + err.Error())
		}
		fmt.Println("mdate:", mdate.Value)

		time, err := time.ParseInLocation("2006-01-02", mdate.Value, time.Local)
		if err == nil {
			c.Date = time
		}

		c.Item = Item.Value
		c.Amount = int(Amount.Value)
		c.Fee = int(Fee.Value)
		c.CPercent = CPercent.Value
		c.SR = SR.Value
		c.Bonus = Bonus.Value
		c.Branch = Branch.Value
		c.DedectItem = DedectItem.Value
		cDataList = append(cDataList, &c)
	}

	salaryM.CommissionList = cDataList
	return
}

func (salaryM *SalaryModel) GetAgentSign(bsID, dbname string) {
	const qsql = `SELECT ss.sid, ss.sname , tmp.item, tmp.amount, tmp.fee, tmp.cpercent, tmp.sr, ( (tmp.amount - coalesce(tmp.fee,0) )* tmp.cpercent/100 * cs.percent/100) bonus , tmp.remark , cs.branch, cs.percent   from salersalary ss
				inner join(
					SELECT c.bsid, c.sid, c.rid, r.date, (c.item || ' ' || ar.name) item, r.amount, 0 , c.sname, c.cpercent, ( r.amount * c.cpercent/100- coalesce(c.fee,0)) sr, 
					r.arid, c.status ,  to_char(r.date,'yyyy-MM-dd') , COALESCE(NULLIF(iv.invoiceno, null),'') , coalesce(d.checknumber,'') , coalesce(c.fee,0) fee , coalesce(d.item,'') remark
					FROM public.commission c
					inner JOIN public.receipt r on r.rid = c.rid	
					inner JOIN public.ar ar on ar.arid = c.arid			
					left join(
						select rid, checknumber , fee, item from public.deduct
					) d on d.rid = r.rid		
					left join(
						select rid,  invoiceno from public.invoice 
					) iv on r.rid = iv.rid
				) tmp on ss.bsid = tmp.bsid and tmp.sid = ss.sid
			Inner Join (
				SELECT A.sid, A.branch, A.percent, A.title, A.code
					FROM public.ConfigSaler A					
			) cs on cs.sid=ss.sid 
			where ss.bsid = '%s' order by cs.code, ss.sid asc;`

	db := cm.imr.GetSQLDBwithDbname(dbname)

	rows, err := db.SQLCommand(fmt.Sprintf(qsql, bsID))
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("SQLCommand Done")
	for rows.Next() {
		var c Commission
		var Item, Branch, DedectItem NullString
		var Amount, Fee NullInt
		var CPercent, SR, Bonus NullFloat
		if err := rows.Scan(&c.Sid, &c.SName, &Item, &Amount, &Fee, &CPercent, &SR, &Bonus, &DedectItem, &Branch, &c.Percent); err != nil {
			fmt.Println("err Scan " + err.Error())
		}
		c.Item = Item.Value
		c.Amount = int(Amount.Value)
		c.Fee = int(Fee.Value)
		c.CPercent = CPercent.Value
		c.SR = SR.Value
		c.Bonus = Bonus.Value
		c.Branch = Branch.Value
		c.DedectItem = DedectItem.Value
		salaryM.CommissionList = append(salaryM.CommissionList, &c)
	}
	//這邊在做啥小? 忘了
	if len(salaryM.CommissionList) > 0 {
		systemM.GetAccountData(dbname)
		salerList := systemM.systemAccountList
		for _, element := range salerList {
			salaryM.SystemAccountList = append(salaryM.SystemAccountList, element)
		}
	}

	return
}

func (salaryM *SalaryModel) ExportIncomeTaxReturn(bsID, dbname string) {

	const dsql = `SELECT ss.date , ss.branch from salersalary ss where ss.bsid = '%s'`

	const qsql = `SELECT ss.sid, ss.sname , cs.identitynum, cs.address, ss.total,  cs.branch, ss.date from salersalary ss			
				Inner Join (
					SELECT A.sid, A.branch, A.identitynum, A.address , A.bankaccount, A.code
						FROM public.ConfigSaler A						
				) cs on cs.sid=ss.sid 
				where ss.date||'-01' >= '%s' and ss.date||'-01' <= '%s' and ss.branch = '%s'	
				order by cs.code, ss.sid asc;`

	db := cm.imr.GetSQLDBwithDbname(dbname)

	// cDataList := []*Commission{}
	// salerList := []*SystemAccount{}

	//查詢日期區間
	rows, err := db.SQLCommand(fmt.Sprintf(dsql, bsID))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("SQLCommand Done")
	date := ""
	branch := ""
	for rows.Next() {
		if err := rows.Scan(&date, &branch); err != nil {
			fmt.Println("err Scan " + err.Error())
		}
	}

	if date == "" {
		fmt.Println("no data here")
		return
	}
	//藉由日期區間查詢
	year := date[0:4]
	rows, err = db.SQLCommand(fmt.Sprintf(qsql, year+"-01-01", date+"-01", branch))
	fmt.Println(fmt.Sprintf(qsql, year+"-01-01", date+"-01", branch))
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("SQLCommand Done")
	for rows.Next() {
		var cs ConfigSaler
		if err := rows.Scan(&cs.Sid, &cs.SName, &cs.IdentityNum, &cs.Address, &cs.Salary, &cs.Branch, &cs.CurDate); err != nil {
			fmt.Println("err Scan " + err.Error())
		}

		configM.ConfigSalerList = append(configM.ConfigSalerList, &cs)
	}

	// out, _ := json.Marshal(configM.ConfigSalerList)
	// fmt.Println("cDataList :", string(out))

	return
}

func (salaryM *SalaryModel) ExportPayrollTransfer(bsID, dbname string) {
	const qsql = `SELECT ss.sid, ss.sname , cs.identitynum, cs.bankaccount, ss.tamount,  cs.branch   from salersalary ss			
			Inner Join (
				SELECT A.sid, A.branch, A.identitynum, A.title , A.bankaccount, A.code
					FROM public.ConfigSaler A
					
			) cs on cs.sid=ss.sid 
			where ss.bsid = '%s'
			order by cs.code, ss.sid asc;`

	db := cm.imr.GetSQLDBwithDbname(dbname)

	// cDataList := []*Commission{}
	// salerList := []*SystemAccount{}

	rows, err := db.SQLCommand(fmt.Sprintf(qsql, bsID))
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("SQLCommand Done")
	for rows.Next() {
		var cs ConfigSaler
		if err := rows.Scan(&cs.Sid, &cs.SName, &cs.IdentityNum, &cs.BankAccount, &cs.Tamount, &cs.Branch); err != nil {
			fmt.Println("err Scan " + err.Error())
		}

		configM.ConfigSalerList = append(configM.ConfigSalerList, &cs)
	}

	// out, _ := json.Marshal(salerList)
	// fmt.Println("salerList :", string(out))
	// out, _ = json.Marshal(cDataList)
	// fmt.Println("cDataList :", string(out))

	return
}

func (salaryM *SalaryModel) addAgentSignInfoTable(table *pdf.DataTable, p *pdf.Pdf) (table_final *pdf.DataTable,
	T_SR, T_Bonus float64) {

	//text := "fd"
	//width := mypdf.MeasureTextWidth(text)
	//table.ColumnLen
	T_SR, T_Bonus = 0.0, 0.0
	tmp_SR, tmp_Bonus := 0.0, 0.0
	sid := ""
	sname := ""
	index := 0
	for i, element := range salaryM.CommissionList {

		if sid != element.Sid {
			sid = element.Sid
			if index != 0 {
				///
				text := ""
				var vs = &pdf.TableStyle{
					Text:  text,
					Bg:    pdf.ColorWhite,
					Front: pdf.ColorTableLine,
				}
				table.RawData = append(table.RawData, vs)
				table.RawData = append(table.RawData, vs)
				table.RawData = append(table.RawData, vs)

				text = sname
				pdf.ResizeWidth(table, p.GetTextWidth(text), 3)
				vs = &pdf.TableStyle{
					Text:  text,
					Bg:    pdf.ColorWhite,
					Front: pdf.ColorTableLine,
				}
				table.RawData = append(table.RawData, vs)
				//
				text = "合計"
				pdf.ResizeWidth(table, p.GetTextWidth(text), 4)
				vs = &pdf.TableStyle{
					Text:  text,
					Bg:    pdf.ColorWhite,
					Front: pdf.ColorTableLine,
				}
				table.RawData = append(table.RawData, vs)
				//
				text = pr.Sprintf("%d", int(tmp_SR))
				pdf.ResizeWidth(table, p.GetTextWidth(text), 5)
				vs = &pdf.TableStyle{
					Text:  text,
					Bg:    pdf.ColorWhite,
					Front: pdf.ColorTableLine,
					Align: pdf.AlignRight,
				}
				table.RawData = append(table.RawData, vs)
				//
				text = pr.Sprintf("%d", int(tmp_Bonus))
				pdf.ResizeWidth(table, p.GetTextWidth(text), 6)
				vs = &pdf.TableStyle{
					Text:  text,
					Bg:    If(true, pdf.ColorWhite, pdf.ColorWhite).(pdf.Color),
					Front: pdf.ColorTableLine,
					Align: pdf.AlignRight,
				}
				table.RawData = append(table.RawData, vs)
				//
				text = ""
				vs = &pdf.TableStyle{
					Text:  text,
					Bg:    pdf.ColorWhite,
					Front: pdf.ColorTableLine,
				}
				table.RawData = append(table.RawData, vs)
				table.RawData = append(table.RawData, vs)
				table.RawData = append(table.RawData, vs)
			}
			tmp_SR, tmp_Bonus = 0.0, 0.0
			index = 0
		}
		if element.Sid == sid {
			///
			index++
			text := element.Item
			pdf.ResizeWidth(table, p.GetTextWidth(text), 0)
			var vs = &pdf.TableStyle{
				Text:  text,
				Bg:    pdf.ColorWhite,
				Front: pdf.ColorTableLine,
				Align: pdf.AlignLeft,
			}
			table.RawData = append(table.RawData, vs)
			//
			text = pr.Sprintf("%d", element.Amount)
			pdf.ResizeWidth(table, p.GetTextWidth(text), 1)
			vs = &pdf.TableStyle{
				Text:  text,
				Bg:    pdf.ColorWhite,
				Front: pdf.ColorTableLine,
				Align: pdf.AlignRight,
			}
			table.RawData = append(table.RawData, vs)
			//應扣
			text = pr.Sprintf("%d", element.Fee)
			pdf.ResizeWidth(table, p.GetTextWidth(text), 2)
			vs = &pdf.TableStyle{
				Text:  text,
				Bg:    pdf.ColorWhite,
				Front: pdf.ColorTableLine,
				Align: pdf.AlignRight,
			}
			table.RawData = append(table.RawData, vs)
			//
			//fmt.Println("element.SName:", element.SName, "element.Sid:", element.SName, "   sid:", sid)
			text = element.SName
			sname = element.SName
			pdf.ResizeWidth(table, p.GetTextWidth(text), 3)
			vs = &pdf.TableStyle{
				Text:  text,
				Bg:    pdf.ColorWhite,
				Front: pdf.ColorTableLine,
			}
			table.RawData = append(table.RawData, vs)
			//
			text = pr.Sprintf("%.1f%s", element.CPercent, "%")
			pdf.ResizeWidth(table, p.GetTextWidth(text), 4)
			vs = &pdf.TableStyle{
				Text:  text,
				Bg:    pdf.ColorWhite,
				Front: pdf.ColorTableLine,
			}
			table.RawData = append(table.RawData, vs)
			//
			T_SR += element.SR
			tmp_SR += element.SR
			text = pr.Sprintf("%d", int(element.SR))
			pdf.ResizeWidth(table, p.GetTextWidth(text), 5)
			vs = &pdf.TableStyle{
				Text:  text,
				Bg:    pdf.ColorWhite,
				Front: pdf.ColorTableLine,
				Align: pdf.AlignRight,
			}
			table.RawData = append(table.RawData, vs)
			//
			element.Bonus = round(float64(element.Bonus), 0) //對第一位小數 四捨五入
			T_Bonus += float64(int(element.Bonus))
			tmp_Bonus += float64(int(element.Bonus))
			text = pr.Sprintf("%d", int(element.Bonus))
			pdf.ResizeWidth(table, p.GetTextWidth(text), 6)
			vs = &pdf.TableStyle{
				Text:  text,
				Bg:    If(true, pdf.ColorWhite, pdf.ColorWhite).(pdf.Color),
				Front: pdf.ColorTableLine,
				Align: pdf.AlignRight,
			}
			table.RawData = append(table.RawData, vs)
			//
			text = element.DedectItem
			pdf.ResizeWidth(table, p.GetTextWidth(text), 7)
			vs = &pdf.TableStyle{
				Text:  text,
				Bg:    If(true, pdf.ColorWhite, pdf.ColorWhite).(pdf.Color),
				Front: pdf.ColorTableLine,
			}
			table.RawData = append(table.RawData, vs)
			//
			text = element.Branch
			pdf.ResizeWidth(table, p.GetTextWidth(text), 8)
			vs = &pdf.TableStyle{
				Text:  text,
				Bg:    If(true, pdf.ColorWhite, pdf.ColorWhite).(pdf.Color),
				Front: pdf.ColorTableLine,
			}
			table.RawData = append(table.RawData, vs)
			//
			text = element.Percent
			pdf.ResizeWidth(table, p.GetTextWidth(text), 9)
			vs = &pdf.TableStyle{
				Text:  text,
				Bg:    If(true, pdf.ColorWhite, pdf.ColorWhite).(pdf.Color),
				Front: pdf.ColorTableLine,
			}
			table.RawData = append(table.RawData, vs)
		}
		//最後一筆合計 hard code
		if i == len(salaryM.CommissionList)-1 {
			///
			text := ""
			var vs = &pdf.TableStyle{
				Text:  text,
				Bg:    pdf.ColorWhite,
				Front: pdf.ColorTableLine,
			}
			table.RawData = append(table.RawData, vs)
			table.RawData = append(table.RawData, vs)
			table.RawData = append(table.RawData, vs)

			text = sname
			pdf.ResizeWidth(table, p.GetTextWidth(text), 3)
			vs = &pdf.TableStyle{
				Text:  text,
				Bg:    pdf.ColorWhite,
				Front: pdf.ColorTableLine,
			}
			table.RawData = append(table.RawData, vs)
			//
			text = "合計"
			pdf.ResizeWidth(table, p.GetTextWidth(text), 4)
			vs = &pdf.TableStyle{
				Text:  text,
				Bg:    pdf.ColorWhite,
				Front: pdf.ColorTableLine,
			}
			table.RawData = append(table.RawData, vs)
			//
			text = pr.Sprintf("%d", int(tmp_SR))
			pdf.ResizeWidth(table, p.GetTextWidth(text), 5)
			vs = &pdf.TableStyle{
				Text:  text,
				Bg:    pdf.ColorWhite,
				Front: pdf.ColorTableLine,
				Align: pdf.AlignRight,
			}
			table.RawData = append(table.RawData, vs)
			//
			text = pr.Sprintf("%d", int(tmp_Bonus))
			pdf.ResizeWidth(table, p.GetTextWidth(text), 6)
			vs = &pdf.TableStyle{
				Text:  text,
				Bg:    If(true, pdf.ColorWhite, pdf.ColorWhite).(pdf.Color),
				Front: pdf.ColorTableLine,
				Align: pdf.AlignRight,
			}
			table.RawData = append(table.RawData, vs)
			//
			text = ""
			vs = &pdf.TableStyle{
				Text:  text,
				Bg:    pdf.ColorWhite,
				Front: pdf.ColorTableLine,
			}
			table.RawData = append(table.RawData, vs)
			table.RawData = append(table.RawData, vs)
			table.RawData = append(table.RawData, vs)
		}
	}
	table_final = table
	return
}

func (salaryM *SalaryModel) addSalerCommissionInfoTable(table *pdf.DataTable, p *pdf.Pdf, cList []*Commission, sid string) (table_final *pdf.DataTable,
	T_SR, T_Bonus float64, date string) {

	//text := "fd"
	//width := mypdf.MeasureTextWidth(text)
	//table.ColumnLen
	T_SR, T_Bonus = 0.0, 0.0
	date = "error"
	for _, element := range cList {
		if element.Sid == sid {
			date = element.Date.Format("2006-01-02 15:04:05")
			///
			text := element.Item
			pdf.ResizeWidth(table, p.GetTextWidth(text), 0)
			var vs = &pdf.TableStyle{
				Text:  text,
				Bg:    pdf.ColorWhite,
				Front: pdf.ColorTableLine,
				Align: pdf.AlignLeft,
			}
			table.RawData = append(table.RawData, vs)
			//
			text = pr.Sprintf("%d", element.Amount)
			pdf.ResizeWidth(table, p.GetTextWidth(text), 1)
			vs = &pdf.TableStyle{
				Text:  text,
				Bg:    pdf.ColorWhite,
				Front: pdf.ColorTableLine,
				Align: pdf.AlignRight,
			}
			table.RawData = append(table.RawData, vs)
			//應扣
			text = pr.Sprintf("%d", element.Fee)
			pdf.ResizeWidth(table, p.GetTextWidth(text), 2)
			vs = &pdf.TableStyle{
				Text:  text,
				Bg:    pdf.ColorWhite,
				Front: pdf.ColorTableLine,
				Align: pdf.AlignRight,
			}
			table.RawData = append(table.RawData, vs)
			//
			//fmt.Println("element.SName:", element.SName, "element.Sid:", element.SName, "   sid:", sid)
			text = element.SName
			pdf.ResizeWidth(table, p.GetTextWidth(text), 3)
			vs = &pdf.TableStyle{
				Text:  text,
				Bg:    pdf.ColorWhite,
				Front: pdf.ColorTableLine,
			}
			table.RawData = append(table.RawData, vs)
			//
			text = fmt.Sprintf("%.1f", element.CPercent)
			pdf.ResizeWidth(table, p.GetTextWidth(text), 4)
			vs = &pdf.TableStyle{
				Text:  text,
				Bg:    pdf.ColorWhite,
				Front: pdf.ColorTableLine,
			}
			table.RawData = append(table.RawData, vs)
			//
			T_SR += element.SR
			text = pr.Sprintf("%d", int(element.SR))
			pdf.ResizeWidth(table, p.GetTextWidth(text), 5)
			vs = &pdf.TableStyle{
				Text:  text,
				Bg:    pdf.ColorWhite,
				Front: pdf.ColorTableLine,
				Align: pdf.AlignRight,
			}
			table.RawData = append(table.RawData, vs)
			//
			T_Bonus += element.Bonus
			text = pr.Sprintf("%d", int(element.Bonus))
			pdf.ResizeWidth(table, p.GetTextWidth(text), 6)
			vs = &pdf.TableStyle{
				Text:  text,
				Bg:    If(true, pdf.ColorWhite, pdf.ColorWhite).(pdf.Color),
				Front: pdf.ColorTableLine,
				Align: pdf.AlignRight,
			}
			table.RawData = append(table.RawData, vs)
			//
			text = element.DedectItem
			pdf.ResizeWidth(table, p.GetTextWidth(text), 7)
			vs = &pdf.TableStyle{
				Text:  text,
				Bg:    If(true, pdf.ColorWhite, pdf.ColorWhite).(pdf.Color),
				Front: pdf.ColorTableLine,
			}
			table.RawData = append(table.RawData, vs)
		}
	}
	table_final = table
	return
}

func (salaryM *SalaryModel) addSRInfoTable(table *pdf.DataTable, p *pdf.Pdf) (table_final *pdf.DataTable,
	T_SR, T_Bonus int) {

	//text := "fd"
	//width := mypdf.MeasureTextWidth(text)
	//table.ColumnLen
	T_SR, T_Bonus = 0, 0

	for _, element := range salaryM.CommissionList {
		//
		text := element.SName
		pdf.ResizeWidth(table, p.GetTextWidth(text), 0)
		vs := &pdf.TableStyle{
			Text:  text,
			Bg:    pdf.ColorWhite,
			Front: pdf.ColorTableLine,
		}
		table.RawData = append(table.RawData, vs)
		//SR
		T_SR += int(element.SR)
		text = pr.Sprintf("%d", int(element.SR))
		pdf.ResizeWidth(table, p.GetTextWidth(text), 1)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    pdf.ColorWhite,
			Front: pdf.ColorTableLine,
			Align: pdf.AlignRight,
		}
		table.RawData = append(table.RawData, vs)
		//績效
		T_Bonus += int(element.Bonus)
		text = pr.Sprintf("%d", int(element.Bonus))
		pdf.ResizeWidth(table, p.GetTextWidth(text), 2)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    pdf.ColorWhite,
			Front: pdf.ColorTableLine,
			Align: pdf.AlignRight,
		}
		table.RawData = append(table.RawData, vs)
	}

	text := "合計"
	pdf.ResizeWidth(table, p.GetTextWidth(text), 0)
	vs := &pdf.TableStyle{
		Text:  text,
		Bg:    pdf.ColorWhite,
		Front: pdf.ColorTableLine,
	}
	table.RawData = append(table.RawData, vs)
	text = pr.Sprintf("%d", T_SR)
	pdf.ResizeWidth(table, p.GetTextWidth(text), 1)
	vs = &pdf.TableStyle{
		Text:  text,
		Bg:    pdf.ColorWhite,
		Front: pdf.ColorTableLine,
		Align: pdf.AlignRight,
	}
	table.RawData = append(table.RawData, vs)
	text = pr.Sprintf("%d", T_Bonus)
	pdf.ResizeWidth(table, p.GetTextWidth(text), 1)
	vs = &pdf.TableStyle{
		Text:  text,
		Bg:    pdf.ColorWhite,
		Front: pdf.ColorTableLine,
		Align: pdf.AlignRight,
	}

	table.RawData = append(table.RawData, vs)

	table_final = table
	return
}

//3
func (salaryM *SalaryModel) addNHIInfoTable(table *pdf.DataTable, p *pdf.Pdf) (table_final *pdf.DataTable,
	T_PayrollBracket, T_Salary, T_Pbonus, T_Bonus, T_Total, T_Balance, T_PTSP,
	T_PD, T_FourBouns, T_SP, T_FourSP, T_Tax, T_SPB int) {

	T_PayrollBracket, T_Salary, T_Pbonus, T_Bonus, T_Total, T_Balance = 0, 0, 0, 0, 0, 0
	T_PTSP, T_PD, T_FourBouns, T_SP, T_FourSP, T_Tax, T_SPB = 0, 0, 0, 0, 0, 0, 0

	Branch := ""

	for _, element := range salaryM.NHISalaryList {

		if element.Branch != Branch {

			Branch = element.Branch
			pdf.ResizeWidth(table, p.GetTextWidth(Branch), 0)
			vs := &pdf.TableStyle{
				Text:  Branch,
				Bg:    pdf.ColorWhite,
				Front: pdf.ColorTableLine,
			}
			table.RawData = append(table.RawData, vs)
			for i := 1; i < 17; i++ {
				vs := &pdf.TableStyle{
					Text:  "",
					Bg:    pdf.ColorWhite,
					Front: pdf.ColorTableLine,
				}
				table.RawData = append(table.RawData, vs)
			}
		}
		//
		text := element.SName
		pdf.ResizeWidth(table, p.GetTextWidth(text), 0)
		vs := &pdf.TableStyle{
			Text:  text,
			Bg:    pdf.ColorWhite,
			Front: pdf.ColorTableLine,
		}
		table.RawData = append(table.RawData, vs)
		//
		T_PayrollBracket += element.PayrollBracket
		text = pr.Sprintf("%d", element.PayrollBracket)
		pdf.ResizeWidth(table, p.GetTextWidth(text), 1)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    pdf.ColorWhite,
			Front: pdf.ColorTableLine,
			Align: pdf.AlignRight,
		}
		table.RawData = append(table.RawData, vs)
		//
		T_Salary += element.Salary
		text = pr.Sprintf("%d", element.Salary)
		pdf.ResizeWidth(table, p.GetTextWidth(text), 2)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    pdf.ColorWhite,
			Front: pdf.ColorTableLine,
			Align: pdf.AlignRight,
		}
		table.RawData = append(table.RawData, vs)
		//
		T_Pbonus += element.Pbonus
		text = pr.Sprintf("%d", element.Pbonus)
		pdf.ResizeWidth(table, p.GetTextWidth(text), 3)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    pdf.ColorWhite,
			Front: pdf.ColorTableLine,
			Align: pdf.AlignRight,
		}
		table.RawData = append(table.RawData, vs)
		//
		T_Bonus += element.Bonus
		text = pr.Sprintf("%d", element.Bonus)
		pdf.ResizeWidth(table, p.GetTextWidth(text), 4)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    pdf.ColorWhite,
			Front: pdf.ColorTableLine,
			Align: pdf.AlignRight,
		}
		table.RawData = append(table.RawData, vs)
		//兼職
		text = "0"
		pdf.ResizeWidth(table, p.GetTextWidth(text), 5)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    pdf.ColorWhite,
			Front: pdf.ColorTableLine,
			Align: pdf.AlignRight,
		}
		table.RawData = append(table.RawData, vs)
		//
		T_Total += element.Total
		text = pr.Sprintf("%d", element.Total)
		pdf.ResizeWidth(table, p.GetTextWidth(text), 6)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    pdf.ColorWhite,
			Front: pdf.ColorTableLine,
			Align: pdf.AlignRight,
		}
		table.RawData = append(table.RawData, vs)
		//
		T_Balance += element.SalaryBalance
		text = pr.Sprintf("%d", element.SalaryBalance)
		pdf.ResizeWidth(table, p.GetTextWidth(text), 7)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    pdf.ColorWhite,
			Front: pdf.ColorTableLine,
			Align: pdf.AlignRight,
		}
		table.RawData = append(table.RawData, vs)
		//
		T_PD += element.PD
		text = pr.Sprintf("%d", element.PD)
		pdf.ResizeWidth(table, p.GetTextWidth(text), 8)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    pdf.ColorWhite,
			Front: pdf.ColorTableLine,
			Align: pdf.AlignRight,
		}
		table.RawData = append(table.RawData, vs)
		//
		T_FourBouns += element.FourBouns
		text = pr.Sprintf("%d", element.FourBouns)
		pdf.ResizeWidth(table, p.GetTextWidth(text), 9)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    pdf.ColorWhite,
			Front: pdf.ColorTableLine,
			Align: pdf.AlignRight,
		}
		table.RawData = append(table.RawData, vs)
		//補充保費薪資差額 4倍-薪資差額
		T_SPB += element.FourBouns - element.SalaryBalance
		text = pr.Sprintf("%d", element.FourBouns-element.SalaryBalance)
		pdf.ResizeWidth(table, p.GetTextWidth(text), 10)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    pdf.ColorWhite,
			Front: pdf.ColorTableLine,
			Align: pdf.AlignRight,
		}
		table.RawData = append(table.RawData, vs)
		//
		T_FourSP += element.FourSP
		text = pr.Sprintf("%d", element.FourSP)
		pdf.ResizeWidth(table, p.GetTextWidth(text), 11)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    pdf.ColorWhite,
			Front: pdf.ColorTableLine,
			Align: pdf.AlignRight,
		}
		table.RawData = append(table.RawData, vs)
		//
		T_FourSP += element.FourSP
		text = pr.Sprintf("%d", element.FourSP)
		pdf.ResizeWidth(table, p.GetTextWidth(text), 12)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    pdf.ColorWhite,
			Front: pdf.ColorTableLine,
			Align: pdf.AlignRight,
		}
		table.RawData = append(table.RawData, vs)
		//
		T_PTSP += element.PTSP
		text = pr.Sprintf("%d", element.PTSP)
		pdf.ResizeWidth(table, p.GetTextWidth(text), 13)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    pdf.ColorWhite,
			Front: pdf.ColorTableLine,
			Align: pdf.AlignRight,
		}
		table.RawData = append(table.RawData, vs)
		//
		T_Tax += int(float64(element.Total) * 0.05)
		text = pr.Sprintf("%d", (int(float64(element.Total) * 0.05)))
		pdf.ResizeWidth(table, p.GetTextWidth(text), 14)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    pdf.ColorWhite,
			Front: pdf.ColorTableLine,
			Align: pdf.AlignRight,
		}
		table.RawData = append(table.RawData, vs)
		//
		text = element.Title
		pdf.ResizeWidth(table, p.GetTextWidth(text), 15)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    pdf.ColorWhite,
			Front: pdf.ColorTableLine,
		}
		table.RawData = append(table.RawData, vs)
		//
		text = element.Description
		pdf.ResizeWidth(table, p.GetTextWidth(text), 16)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    pdf.ColorWhite,
			Front: pdf.ColorTableLine,
		}
		table.RawData = append(table.RawData, vs)
	}

	table_final = table
	return
}

func (salaryM *SalaryModel) addPayrollTransferInfoTable(mtype int) (DataList []*excel.DataTable) {
	DataList = []*excel.DataTable{}
	table := excel.GetDataTable(mtype)
	var offset = 0
	for index, element := range configM.ConfigSalerList {

		if table.SheetName != element.Branch {
			//new branch (need to create sheet)
			if index != 0 {
				DataList = append(DataList, table)
			}

			table = excel.GetDataTable(mtype)
			offset = index
		}
		table.SheetName = element.Branch
		table.RawData["A"+strconv.Itoa(index+2-offset)] = element.SName
		table.RawData["B"+strconv.Itoa(index+2-offset)] = element.IdentityNum
		table.RawData["C"+strconv.Itoa(index+2-offset)] = element.BankAccount
		table.RawData["D"+strconv.Itoa(index+2-offset)] = pr.Sprintf("%d", element.Tamount)
		//table.RawData["D"+strconv.Itoa(index+2-offset)] = strconv.Itoa(element.Tamount)
	}
	DataList = append(DataList, table)

	return
}

func (salaryM *SalaryModel) addIncomeTaxReturnInfoTable(mtype int) (DataList []*excel.DataTable) {
	DataList = []*excel.DataTable{}
	table := excel.GetDataTable(mtype)
	salarID_pos := map[string]string{}
	var offset = 0

	//var offset_id_pos = 0
	fmt.Println("addIncomeTaxReturnInfoTable")
	for index, element := range configM.ConfigSalerList {

		if table.SheetName != element.Branch {
			//new branch (need to create sheet)
			if index != 0 {
				DataList = append(DataList, table)
			}

			table = excel.GetDataTable(mtype)
			offset = 0
		}
		table.SheetName = element.Branch

		pos := salarID_pos[element.Sid]
		fmt.Println("pos:", pos)
		if pos == "" {
			fmt.Println("pos is nil")

			salarID_pos[element.Sid] = strconv.Itoa(offset + 2)
			pos = salarID_pos[element.Sid]
			table.RawData["A"+strconv.Itoa(offset+2)] = element.SName
			table.RawData["B"+strconv.Itoa(offset+2)] = element.IdentityNum
			table.RawData["C"+strconv.Itoa(offset+2)] = element.Address
			offset++
		}

		mm, _ := strconv.Atoi(element.CurDate[5:])
		fmt.Println(excel.AZTable[2+mm]+pos, " set ", strconv.Itoa(element.Salary))
		table.RawData[excel.AZTable[2+mm]+pos] = strconv.Itoa(element.Salary)
	}
	DataList = append(DataList, table)

	return
}

func (salaryM *SalaryModel) getSalerEmail(dbname string, things ...string) ([]*ConfigSaler, error) {

	branch := "%"
	for _, it := range things {
		branch = it
	}

	const qspl = `SELECT A.sid, A.sname, A.branch, A.Email, A.code	FROM public.ConfigSaler A 				
					where A.branch like '%s';`
	//const qspl = `SELECT arid,sales	FROM public.ar;`
	db := salaryM.imr.GetSQLDBwithDbname(dbname)
	//fmt.Println(fmt.Sprintf(qspl, branch))
	rows, err := db.SQLCommand(fmt.Sprintf(qspl, branch))
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	var saList []*ConfigSaler

	for rows.Next() {
		var saler ConfigSaler

		// if err := rows.Scan(&r.ARid, &s); err != nil {
		// 	fmt.Println("err Scan " + err.Error())
		// }
		if err := rows.Scan(&saler.Sid, &saler.SName, &saler.Branch, &saler.Email, &saler.Code); err != nil {
			fmt.Println("err Scan " + err.Error())
		}

		saList = append(saList, &saler)
	}
	salaryM.MailList = saList
	return saList, nil

}

//更新薪資表
func (salaryM *SalaryModel) ReFreshSalerSalary(Bsid, dbname string) error {
	salaryM.RefreshCommissionBonusbyBsid(Bsid, dbname)
	const sql = `UPDATE public.salersalary t
				SET total= subquery.salary + subquery.pbonus + lbonus - abonus ,
				laborfee = ( Case When workday >= 30 then subquery.laborfee else subquery.laborfee * workday / 30 END),
				healthfee = ( Case When workday >= 30 then subquery.healthfee else 0 END) ,
				sp = subquery.sp,
				tamount = subquery.salary + subquery.pbonus + lbonus - abonus - tax - other - subquery.sp - welfare - commercialFee - ( Case When workday >= 30 then subquery.laborfee else subquery.laborfee * workday / 30 END) - ( Case When workday >= 30 then subquery.healthfee else 0 END)
				FROM (
					Select A.Sid, A.salary, A.association, COALESCE(extra.bonus,0) pbonus, A.payrollbracket , (A.insuredamount * CP.li * 0.2 / 100) laborfee, (A.payrollbracket * CP.nhi * 0.3 / 100) healthfee ,	 CP.* , 
					(CASE WHEN A.salary = 0 and A.association = 1 then 0 
					WHEN (COALESCE(A.Salary + COALESCE(extra.bonus,0) ,A.Salary)) <= CP.mmw then 0	 	
					WHEN A.salary = 0 and A.association = 0 then COALESCE(A.Salary+  COALESCE(extra.bonus,0) ,A.Salary) * cp.nhi2nd / 100 	 	
					else
						( CASE WHEN ((COALESCE(A.Salary+  COALESCE(extra.bonus,0) ,A.Salary)) - 4 * A.PayrollBracket) > 0 then ((COALESCE(A.Salary+  COALESCE(extra.bonus,0) ,A.Salary)) - 4 * A.PayrollBracket) * cp.nhi2nd / 100 else 0 end)
					end
					) sp
					FROM public.ConfigSaler A 
					Inner Join ( 
						select sid, max(zerodate) zerodate from public.configsaler cs 
						where now() > zerodate -- and Sid = $7
						group by sid 
					) B on A.sid=B.sid and A.zeroDate = B.zeroDate
					cross join ( 
						select  c.date, c.nhi, c.li, c.nhi2nd, c.mmw from public.ConfigParameter C
						inner join(
							select  max(date) date from public.ConfigParameter 
						) A on A.date = C.date limit 1
					) CP
					left join (
						select sum(bonus) bonus , bsid , sid from public.commission c
						group by bsid , sid
					) extra on extra.bsid = $1 and A.sid = extra.sid
					left join(
						select branch , commercialFee from public.configbranch 
					) CB on CB.branch = A.branch
				) as subquery
				WHERE t.bsid = $1 and t.sid = subquery.sid;`
	interdb := salaryM.imr.GetSQLDBwithDbname(dbname)
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		fmt.Printf(err.Error())
	}
	defer sqldb.Close()
	res, err := sqldb.Exec(sql, Bsid)
	//res, err := sqldb.Exec(sql, unix_time, receivable.Date, receivable.CNo, receivable.Sales)
	if err != nil {
		fmt.Println(err)
	}

	id, err := res.RowsAffected()
	if err != nil {
		fmt.Println("PG Affecte Wrong: ", err)
	}
	fmt.Println(fmt.Sprintf("更新bsid[%s] %d資料", Bsid, id))
	if id <= 0 {
		return errors.New("not found bsid:" + Bsid)
	}
	//TODO:: 紅利店長表更新
	salaryM.UpdateManagerByManagerBonus(Bsid, dbname)
	//TODO:: 二代健保表更新
	salaryM.RefreshNHISalary(Bsid, dbname)
	//TODO:: 總表更新
	_err := salaryM.UpdateBranchSalaryTotal(dbname)
	if _err != nil {
		fmt.Println(_err)
		return nil
		//return css_err
	}

	return nil
}

//更新傭金byBsid
func (salaryM *SalaryModel) RefreshCommissionBonusbyBsid(Bsid, dbname string) (err error) {

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
								WHERE c.bsid = $1
				) as t2 where t1.sid = t2.sid and t1.rid = t2.rid`
	interdb := salaryM.imr.GetSQLDBwithDbname(dbname)
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}

	_, err = sqldb.Exec(sql, Bsid)
	//res, err := sqldb.Exec(sql, unix_time, receivable.Date, receivable.CNo, receivable.Sales)
	if err != nil {
		fmt.Println("RefreshCommissionBonusbyBsid:", err)
		return err
	}
	defer sqldb.Close()
	return nil
}

//更新二代健保表
func (salaryM *SalaryModel) RefreshNHISalary(bsid, dbname string) (err error) {

	const sql = `INSERT INTO public.nhisalary
	(sid, bsid, sname, payrollbracket, salary, pbonus, bonus, total, salarybalance, pd, fourbouns, sp, foursp, ptsp)
	SELECT  SS.sid, SS.BSid, SS.Sname, CS.PayrollBracket, SS.Salary, SS.Pbonus, (SS.Lbonus + ie.managerbonus - SS.abonus) bonus, 
	(SS.Salary + SS.Pbonus + (SS.Lbonus + ie.managerbonus  - SS.abonus) ) Total ,
	( (SS.Salary + SS.Pbonus + (SS.Lbonus + ie.managerbonus  - SS.abonus) ) - CS.PayrollBracket) SalaryBalance,
	sum( (SS.Salary + SS.Pbonus + (SS.Lbonus + ie.managerbonus  - SS.abonus) ) - CS.PayrollBracket) over (partition by SS.year,SS.sid order by SS.date) PD ,
	(CS.PayrollBracket * 4) FourBouns, 0 SP,
	(CASE WHEN sum(SS.Total - CS.PayrollBracket) over (partition by SS.year,SS.sid order by SS.date) - (CS.PayrollBracket * 4) > 0 then (SS.Total - CS.PayrollBracket) *CP.nhi2nd /100  ELSE 0 END ) FourSP, 
	(CASE WHEN CS.association = 0 and CS.PayrollBracket <=0 then SS.Total*CP.nhi2nd ELSE 0 END ) PTSP  
	FROM public.salersalary SS
		Inner Join (
			select A.Sid, A.PayrollBracket, A.association  FROM public.ConfigSaler A 
			Inner Join ( 
				select sid, max(zerodate) zerodate from public.configsaler cs 
				where now() > zerodate
				group by sid 
			) B on A.sid=B.sid and A.zeroDate = B.zeroDate
		) CS on SS.sid = CS.sid
		cross join (
			select  c.date, c.nhi, c.li, c.nhi2nd, c.mmw from public.ConfigParameter C
			inner join(
				select  max(date) date from public.ConfigParameter 
			) A on A.date = C.date limit 1
		) CP
		inner join (
			Select bsid , managerbonus from public.incomeexpense 
		) ie on ie.bsid = ss.bsid
		inner join (
			Select bsid , lock from public.BranchSalary 
		) bs on bs.bsid = ss.bsid		
	WHERE SS.bsid =  $1  and bs.lock = '未完成'
    ON CONFLICT (bsid,sid) DO UPDATE SET sname = excluded.sname, payrollbracket= excluded.payrollbracket,
		salary= excluded.salary, pbonus= excluded.pbonus, bonus= excluded.bonus, total= excluded.total , salarybalance= excluded.salarybalance,
		pd= excluded.pd, fourbouns= excluded.fourbouns, sp= excluded.sp, foursp= excluded.foursp, ptsp=excluded.ptsp ;`

	interdb := salaryM.imr.GetSQLDBwithDbname(dbname)
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}

	res, err := sqldb.Exec(sql, bsid)
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
	fmt.Println("RefreshNHISalary:", id)

	if id == 0 {
		fmt.Println("RefreshNHISalary, not found any salary ")
	}
	defer sqldb.Close()
	return nil
}

//13 txt 薪資簡易版
func (salaryM *SalaryModel) MakeTxtTransferSalary(bsid, dbname string) error {

	const qspl = `SELECT s.branch, s.date, s.sid, s.tamount, c.bankaccount, '822' FROM public.salersalary s
	INNER JOIN public.configsaler c on c.sid = s.sid 
	where s.bsid = $1;`
	//const qspl = `SELECT arid,sales	FROM public.ar;`
	db := salaryM.imr.GetSQLDBwithDbname(dbname)
	sqldb, err := db.ConnectSQLDB()
	//fmt.Println(fmt.Sprintf(qspl, branch))
	rows, err := sqldb.Query(qspl, bsid)
	if err != nil {
		fmt.Println(err)
		return err
	}
	var tList []*TransferSalary

	for rows.Next() {
		var t TransferSalary

		if err := rows.Scan(&t.Branch, &t.Date, &t.IDNo, &t.Amount, &t.Account, &t.BankNo); err != nil {
			fmt.Println("makeTxtTransferSalary err Scan " + err.Error())
			return err
		}

		tList = append(tList, &t)
	}

	out, err := json.Marshal(tList)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(out))
	defer sqldb.Close()
	salaryM.TransferSalaryList = tList

	return nil
}

//轉文字
func (salaryM *SalaryModel) makeTxtTransferSalary() (string, string) {
	data := ""
	branch_date := ""
	for _, element := range salaryM.TransferSalaryList {
		if data != "" {
			data += "\n"
		}
		data += fmt.Sprintf("%016s%016d%s%11s", element.Account, element.Amount, element.BankNo, element.IDNo)
	}
	if len(salaryM.TransferSalaryList) > 0 {
		text, _ := util.ADtoROC(salaryM.TransferSalaryList[0].Date, "file")
		branch_date = salaryM.TransferSalaryList[0].Branch + text
	}
	return data, branch_date
}

//Action:取得目前關帳日期
func (salaryM *SalaryModel) GetAccountSettlement(dbname string) (ca CloseAccount, err error) {
	const sql = `SELECT id, uid, closedate, status, date
					FROM public.accountsettlement Where status = '1';`

	ca = CloseAccount{}

	interdb := salaryM.imr.GetSQLDBwithDbname(dbname)
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return
	}

	rows, err := sqldb.Query(sql)
	if err != nil {
		fmt.Println(err)
		return
	}

	for rows.Next() {
		if err := rows.Scan(&ca.id, &ca.Uid, &ca.CloseDate, &ca.Status, &ca.Date); err != nil {
			fmt.Println("err Scan " + err.Error())
		}

	}
	fmt.Println(ca)
	fmt.Println(ca.CloseDate.Unix())
	salaryM.CloseAccount = &ca
	defer sqldb.Close()
	return
}

//Action:測試刪除用
func (salaryM *SalaryModel) DeleteAccountSettlement(dbname string) (ca CloseAccount, err error) {
	const sql = `delete FROM public.accountsettlement;`

	ca = CloseAccount{}
	interdb := salaryM.imr.GetSQLDBwithDbname(dbname)
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return
	}

	_, err = sqldb.Query(sql)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer sqldb.Close()
	return
}

//Action:會計關帳
func (salaryM *SalaryModel) CloseAccountSettlement(ca *CloseAccount, per, dbname string) (err error) {

	oriCa, err := salaryM.CheckValidCloseDate(ca.CloseDate, dbname)
	if err != nil && per != permission.Admin {
		return
	}

	// oriCa, err := salaryM.GetAccountSettlement()
	// if err != nil {
	// 	return err
	// }

	ca.CloseDate = setDayEndDate(ca.CloseDate)

	interdb := salaryM.imr.GetSQLDBwithDbname(dbname)
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}
	defer sqldb.Close()

	fakeId := time.Now().Unix()
	/**條件敘述:
	1.關帳基本上只能往以後的日期關。
	2.Admin可往前關
	(所以關帳日可能回朔，取最大closedate不行)=>多用status判斷
	**/

	id := int64(-1)

	if oriCa.id == "" || per == permission.Admin {
		salaryM.updateAllAccountSettlementStatus(dbname)
		//資料庫預設空的，直接設定
		fmt.Println("case2 ca:", ca.CloseDate.Unix())
		sql := `INSERT INTO public.accountsettlement(id, uid, closedate )
					select $1, $2, to_timestamp($3);`
		res, err := sqldb.Exec(sql, fakeId, ca.Uid, ca.CloseDate.Unix())
		if err != nil {
			fmt.Println(err)
			return err
		}
		id, err = res.RowsAffected()
	} else {
		//資料庫有數據。
		fmt.Println("case1 ca:", ca)
		sql := `INSERT INTO public.accountsettlement(id, uid, closedate)
					select $1, $2, $3 
					where exists (
						select * from accountsettlement where $4 > $5
					   );  `
		res, err := sqldb.Exec(sql, fakeId, ca.Uid, ca.CloseDate, ca.CloseDate, oriCa.CloseDate)
		if err != nil {
			fmt.Println(err)
			return err
		}
		id, err = res.RowsAffected()
	}

	if err != nil {
		fmt.Println("CloseAccountSettlement:", err)
		return err
	}
	//更動status，紀錄目前關帳的日期
	if id > 0 {
		err = salaryM.updateAccountSettlementStatus(oriCa, dbname)
		if err != nil {
			return err
		}
	} else if id == -1 {
		return errors.New("CloseAccountSettlement unknown error")
	} else {
		fmt.Println("[ERROR] CloseAccountSettlement id:", id)
		return errors.New("[ERROR] CloseAccountSettlement failed")
	}

	return nil
}

//Action:會計關帳
func (salaryM *SalaryModel) updateAccountSettlementStatus(oriCa *CloseAccount, dbname string) error {

	const sql = `update public.accountsettlement set status = '0' where id = $1	;`

	interdb := salaryM.imr.GetSQLDBwithDbname(dbname)
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}

	res, err := sqldb.Exec(sql, oriCa.id)
	defer sqldb.Close()
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
		fmt.Println("[ERROR] updateAccountSettlementStatus id:", id)
		return errors.New("[ERROR] updateAccountSettlementStatus id:0")
	}
	return nil
}

func (salaryM *SalaryModel) updateAllAccountSettlementStatus(dbname string) error {

	const sql = `update public.accountsettlement set status = '0';`

	interdb := salaryM.imr.GetSQLDBwithDbname(dbname)
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}

	sqldb.Exec(sql)
	defer sqldb.Close()

	return nil
}

func (salaryM *SalaryModel) CheckValidCloseDate(t time.Time, dbname string) (*CloseAccount, error) {

	//關帳日在建資料的時間點之後，不給建立
	ca, _ := salaryM.GetAccountSettlement(dbname)
	if ca.CloseDate.After(t) {
		errtime := ca.CloseDate.Format("2006-01-02")
		return &ca, errors.New("關帳日期錯誤:" + errtime)
	}
	salaryM.CloseAccount = &ca
	return &ca, nil
}

func setDayEndDate(t time.Time) time.Time {
	loc, _ := time.LoadLocation("Asia/Taipei")
	taipei := t.In(loc)
	y, m, d := taipei.In(loc).Date()
	fmt.Println("setEnd:", time.Date(y, m, d, 23, 59, 59, 99, loc).Unix())
	return time.Date(y, m, d, 23, 59, 59, 99, loc)
	//return t
}
