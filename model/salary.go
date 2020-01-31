package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/94peter/toad/excel"

	"dforcepro.com/report"
	"github.com/94peter/toad/pdf"
	"github.com/94peter/toad/resource/db"
	"github.com/94peter/toad/util"
)

type BranchSalary struct {
	BSid    string `json:"BSid"`
	Branch  string `json:"branch"`
	strDate string `json:"-"` //建立string Date
	Date    string `json:"date"`
	Name    string `json:"name"`
	Total   string `json:"total"`
	Lock    string `json:"lock"`
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

type Cid struct {
	Sid string `json:"sid"`
	Rid string `json:"rid"`
}

var (
	salaryM *SalaryModel
)

type SalaryModel struct {
	imr               interModelRes
	db                db.InterSQLDB
	salerSalaryList   []*SalerSalary
	branchSalaryList  []*BranchSalary
	NHISalaryList     []*NHISalary
	IncomeExpenseList []*IncomeExpense

	SystemAccountList []*SystemAccount
	CommissionList    []*Commission
	FnamePdf          string
	SMTPConf          util.SendMail
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

func (salaryM *SalaryModel) GetBranchSalaryData(date string) []*BranchSalary {

	const spl = `SELECT bsid, to_timestamp(bsid::int), branch, name, total, lock FROM public.branchsalary
				where date Like '%s';`

	db := salaryM.imr.GetSQLDB()
	rows, err := db.SQLCommand(fmt.Sprintf(spl, date))
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

func (salaryM *SalaryModel) PDF(mtype int, isNew bool, things ...string) {
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
		fmt.Println(e)
		pdf.ResizeWidth(table, p.GetTextWidth(e.Text), index)
	}

	fmt.Println("mtype:", mtype)
	switch mtype {
	case pdf.BranchSalary:
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
		//var total_SR, total_Bonus = 0.0, 0.0
		// for _, saler := range salaryM.SystemAccountList {
		// 	table.RawData = table.RawData[:table.ColumnLen]
		// 	fmt.Println("saler:", saler.Name)
		// 	data, T_SR, T_Bonus := salaryM.addAgentSignInfoTable(table, p, salaryM.CommissionList, saler.Account)
		// 	//data.RawData = data.RawData[data.ColumnLen:]
		// 	fmt.Println("DrawTablePDF")
		// 	if len(table.RawData) > 0 && len(table.RawData) != table.ColumnLen {
		// 		p.DrawTablePDF(data)
		// 	}
		// 	fmt.Println("CustomizedAgentSign")
		// 	SR, Bonus := p.CustomizedAgentSign(data, saler.Name, T_Bonus, T_SR)
		// 	total_SR += SR
		// 	total_Bonus += Bonus

		// 	fmt.Println("clear Header")
		// }
		data, T_SR, T_Bonus := salaryM.addAgentSignInfoTable(table, p)
		//data.RawData = data.RawData[data.ColumnLen:]
		fmt.Println("DrawTablePDF")
		p.DrawTablePDF(data)
		fmt.Println("CustomizedAgentSign")
		//SR, Bonus := p.CustomizedAgentSign(data, "saler.Name", T_Bonus, T_SR)
		//total_SR += SR
		//total_Bonus += Bonus
		p.CustomizedAgentSign(table, T_SR, T_Bonus)
		break
	case pdf.SalarCommission: //8
		mailList, err := salaryM.getSalerEmail()
		if err != nil {
			//getSalerEmail 失敗
			fmt.Println(err)
			return
		}
		for _, saler := range mailList {
			fmt.Println("saler:", saler.SName)
			table = pdf.GetDataTable(mtype)
			data, T_SR, T_Bonus := salaryM.addSalerCommissionInfoTable(table, p, salaryM.CommissionList, saler.Sid)
			p.DrawTablePDF(data)
			p.CustomizedSalerCommission(data, saler.SName, int(T_Bonus), int(T_SR))
		}
		break
	case pdf.SalerSalary: //7
		if len(salaryM.salerSalaryList) > 0 {
			//systemM.GetAccountData()
			mailList, err := salaryM.getSalerEmail()

			for index, element := range salaryM.salerSalaryList {
				p := pdf.GetNewPDF()
				table := pdf.GetDataTable(mtype)
				//Header長度重製
				for index, e := range table.RawData {
					fmt.Println(e)
					pdf.ResizeWidth(table, p.GetTextWidth(e.Text), index)
				}

				data := salaryM.addSalerSalaryInfoTable(table, p, index, element)
				fmt.Println("DrawTablePDF:")
				p.DrawTablePDF(data)
				date, _ := util.ADtoROC(element.Date, "file")
				fname := element.Branch + "-" + element.SName + "薪資表" + date
				p.WriteFile(fname)
				//p = nil
				if send == "true" {
					if err != nil {
						//getSalerEmail 失敗
						fmt.Println(err)
						return
					}

					for _, myAccount := range mailList {
						if myAccount.Sid == element.Sid {
							fmt.Println(myAccount, element)
							fmt.Println(fname)
							util.RunSendMail(salaryM.SMTPConf.Host, salaryM.SMTPConf.Port, salaryM.SMTPConf.Password, salaryM.SMTPConf.User, "geassyayaoo3@gmail.com", pdf.ReportToString(mtype), "開啟若有密碼，則為123456", fname+".pdf")
						}
					}
				}
			}

		}
		//

		// for _, salerSalary := range salaryM.salerSalaryList {
		// 	for _, element := range systemM.systemAccountList {
		// 		if element.Account == salerSalary.Sid {
		// 			fmt.Println(element, salerSalary)
		// 			util.RunSendMail(smtpConf.Host, smtpConf.Port, smtpConf.Password, smtpConf.User, "geassyayaoo3@gmail.com", "subject", body, fname)
		// 		}
		// 	}
		// }
		// body := "testbody"
		// fname := "hello.pdf"
		// //conf := di.GetSMTPConf()
		// fmt.Println(smtpConf)

		//date, _ := util.ADtoROC(salaryM.salerSalaryList[0].Date, "file")
		//p.WriteFile(salaryM.salerSalaryList[0].Branch + "薪資表" + date)

		break
	case pdf.SR:
		//table = pdf.GetDataTable(mtype)
		if len(salaryM.CommissionList) > 0 {
			fmt.Println("SR")
			data, _, _ := salaryM.addSRInfoTable(table, p)
			p.DrawTablePDF(data)
			p.NewLine(25)
		}
		break
	case pdf.NHI:
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
	}

	return //p.GetBytesPdf()
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

func (salaryM *SalaryModel) DeleteSalary(ID string) (err error) {
	const sql = `
				delete from public.SalerSalary where bsid = '%s';
				delete from public.NHISalary where bsid = '%s';
				delete from public.BranchSalary where bsid = '%s';
				delete from public.incomeexpense where bsid = '%s';
				UPDATE public.commission SET bsid = null , status = 'normal' WHERE bsid = '%s';
				`

	interdb := salaryM.imr.GetSQLDB()
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

	// id, err := res.RowsAffected()
	// if err != nil {
	// 	fmt.Println("PG Affecte Wrong: ", err)
	// 	return err
	// }
	// if id <= 0 {
	// 	return errors.New("not found anything")
	// }
	return nil
}

func (salaryM *SalaryModel) CreateSalary(bs *BranchSalary, cid []*Cid) (err error) {

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
	interdb := salaryM.imr.GetSQLDB()
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}

	fakeId := time.Now().Unix()
	bs.BSid = strconv.Itoa(int(fakeId))

	res, err := sqldb.Exec(sql, fakeId, bs.Date, bs.Name, bs.Date+"-01")
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

	scsErr := salaryM.SetCommissionBSid(bs, cid)
	if scsErr != nil {
		return nil
		//return css_err
	}

	cssErr := salaryM.CreateSalerSalary(bs, cid)
	if cssErr != nil {
		return nil
		//return css_err
	}

	_err := salaryM.UpdateBranchSalaryTotal()
	if _err != nil {
		return nil
		//return css_err
	}

	// const bppsql = `INSERT INTO public.branchprepay
	// (ppid, branch, cost)
	// VALUES ($1, $2, $3)
	// ;`

	// i := 0
	// for range salerlist {
	// 	if salerlist[i].Bid == "testBid" {
	// 		break
	// 	}
	// 	i++
	// }
	// i := 0
	// for range prepay.PrePay {

	// 	bppres, err := sqldb.Exec(bppsql, fakeId, prepay.PrePay[i].Branch, prepay.PrePay[i].Cost)
	// 	//res, err := sqldb.Exec(sql, unix_time, receivable.Date, receivable.CNo, receivable.Sales)
	// 	if err != nil {
	// 		fmt.Println(err)
	// 		return err
	// 	}

	// 	bppid, err := bppres.RowsAffected()
	// 	if err != nil {
	// 		fmt.Println("PG Affecte Wrong: ", err)
	// 		return err
	// 	}
	// 	if bppid == 0 {
	// 		return errors.New("Invalid operation, CreateBranchPrepay")
	// 	}
	// 	i++
	// }

	return nil
}

func (salaryM *SalaryModel) CreateSalerSalary(bs *BranchSalary, cid []*Cid) (err error) {

	const sql = `INSERT INTO public.salersalary
	(bsid, sid, date,  branch, sname, salary, pbonus, total, laborfee, healthfee, welfare, commercialfee, tamount, year, sp)
	SELECT BS.bsid, A.sid, COALESCE(C.dateID, $1) dateID, A.branch, A.sname,  A.Salary, COALESCE(C.Pbonus,0)+ COALESCE(extra.bonus,0) Pbonus, 
	COALESCE(A.Salary+  C.Pbonus,A.Salary) total, A.PayrollBracket*CP.LI*0.2/100 LaborFee,A.PayrollBracket*CP.nhi*0.3/100 HealthFee,
	COALESCE(A.Salary+  C.Pbonus,A.Salary)*0.01 Welfare, COALESCE(A.Salary+  C.Pbonus,A.Salary)*cb.commercialFee/100 commercialFee,
	(COALESCE(A.Salary+  C.Pbonus,A.Salary)*0.98 -  A.PayrollBracket*(CP.LI*0.2+CP.nhi*0.3)/100 ) Tamount,
	$3 ,
	(CASE WHEN A.salary = 0 and A.association = 1 then 0 
		WHEN A.salary = 0 and A.association = 0 then COALESCE(A.Salary+  C.Pbonus,A.Salary) * cp.nhi2nd / 100 
		else
			( CASE WHEN ((COALESCE(A.Salary+  C.Pbonus,A.Salary)) - 4 * A.PayrollBracket) > 0 then ((COALESCE(A.Salary+  C.Pbonus,A.Salary)) - 4 * A.PayrollBracket) * cp.nhi2nd / 100 else 0 end)
		end
	   ) sp
	FROM public.ConfigSaler A
	Inner Join ( 
		select sid, max(zerodate) zerodate from public.configsaler cs 
		where now() > zerodate
		group by sid 
	) B on A.sid=B.sid and A.zeroDate = B.zeroDate
	left join (
	SELECT c.sid , to_char( to_timestamp(date_part('epoch',r.date)::int),'YYYY-MM')::varchar(50) dateID, sum(c.bonus) Pbonus
	FROM public.receipt r, public.commission c
	where c.rid = r.rid and to_timestamp(date_part('epoch',r.date)::int) >= $2::date and to_timestamp(date_part('epoch',r.date)::int) < ( $2::date + '1 month'::interval) and c.bsid is null
	group by dateID , c.sid 
	) C on C.sid = A.Sid 
	cross join (
		select  c.date, c.nhi, c.li, c.nhi2nd, c.mmw from public.ConfigParameter C
		inner join(
			select  max(date) date from public.ConfigParameter 
		) A on A.date = C.date limit 1
	) CP
	left join public.branchsalary BS on BS.branch = A.Branch and BS.date = $1
	left join (
		select sum(bonus) bonus , bsid , sid from public.commission c
		group by bsid , sid
	) extra on extra.bsid = BS.bsid and A.sid = extra.sid
	left join(
		select branch , commercialFee from public.configbranch 
	) CB on CB.branch = A.branch
	where BS.bsid is not null
	ON CONFLICT (bsid,sid,date,branch) DO Nothing;	
	`

	year := bs.Date[0:4]
	fmt.Println(year)

	interdb := salaryM.imr.GetSQLDB()
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}
	//fmt.Println("BSID:" + bs.BSid)
	//fmt.Println(bs.Date)
	res, err := sqldb.Exec(sql, bs.Date, bs.Date+"-01", year)
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

	cieErr := salaryM.CreateIncomeExpense(bs)
	if cieErr != nil {
		return nil
		//return css_err
	}

	ucias_err := salaryM.CreateNHISalary(year)
	if ucias_err != nil {
		return nil
		//return ucias_err
	}

	ucias_err = salaryM.UpdateCommissionBSidAndStatus(bs, cid)
	if ucias_err != nil {
		return nil
		//return ucias_err
	}

	return nil
}

func (salaryM *SalaryModel) SetCommissionBSid(bs *BranchSalary, cid []*Cid) (err error) {

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

		year := bs.Date[0:4]
		fmt.Println(year)

		interdb := salaryM.imr.GetSQLDB()
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
	}
	return nil
}

func (salaryM *SalaryModel) CreateIncomeExpense(bs *BranchSalary) (err error) {

	//(subtable.pretaxTotal + subtable.PreTax )  lastloss ,   應該不包含這期虧損
	const sql = `INSERT INTO public.incomeexpense
	(bsid, Pbonus ,LBonus, salary, prepay, pocket, amorcost, sr, annualbonus, salesamounts,  businesstax, agentsign, rent, commercialfee, pretax, businessincometax, aftertax,  lastloss, managerbonus, annualratio )	
	WITH  vals  AS (VALUES ( 'none' ) )
	SELECT subtable.bsid , subtable.Pbonus, subtable.LBonus , subtable.salary, subtable.prepay, subtable.pocket , subtable.thisMonthAmor , subtable.sr, subtable.annualbonus, subtable.salesamounts , subtable.businesstax , subtable.agentsign , subtable.rent,
	subtable.commercialfee, subtable.PreTax , ( CASE WHEN subtable.PreTax > 0 then subtable.PreTax * 0.2 else 0 end ) BusinessIncomeTax, 
	subtable.PreTax - ( CASE WHEN subtable.PreTax > 0 then subtable.PreTax * 0.2 else 0 end ) AfterTax , 
	(subtable.pretaxTotal)  lastloss ,  
	( CASE WHEN (subtable.PreTax - ( CASE WHEN subtable.PreTax > 0 then subtable.PreTax * 0.2 else 0 end )) + (subtable.pretaxTotal + subtable.PreTax ) + 0 > 0 then 
	            (subtable.PreTax - ( CASE WHEN subtable.PreTax > 0 then subtable.PreTax * 0.2 else 0 end )) + (subtable.pretaxTotal + subtable.PreTax ) + 0
	  else 0 end
	) managerbonus  , subtable.annualratio
	FROM vals as v
	cross join (
	SELECT incomeexpense.branch , COALESCE(incomeexpense.pretaxTotal ,0) pretaxTotal , BS.Bsid,BonusTable.PBonus , BonusTable.LBonus , BonusTable.Salary , COALESCE(prepayTable.prepay,0) prepay , COALESCE(pocketTable.pocket,0) pocket , COALESCE(amorTable.thisMonthAmor,0) thisMonthAmor,
	COALESCE(commissionTable.SR,0) SR, COALESCE(commissionTable.SR / 1.05 ,0) salesamounts , COALESCE(commissionTable.SR - commissionTable.SR / 1.05 ,0) businesstax, configTable.agentsign, configTable.rent, configTable.commercialfee, 
	( COALESCE(commissionTable.SR,0)/1.05  - COALESCE(amorTable.thisMonthAmor,0) - configTable.agentsign - configTable.rent - COALESCE(pocketTable.pocket,0) - COALESCE(prepayTable.prepay,0) - BonusTable.PBonus - 
	BonusTable.Salary - BonusTable.LBonus - COALESCE(commissionTable.SR,0) * 0.05 - configTable.commercialfee - 0  ) PreTax ,
	COALESCE(commissionTable.SR * configTable.annualratio ,0) Annualbonus , configTable.annualratio
	FROM public.branchsalary  BS
	inner join (
	  SELECT sum(BonusTable.pbonus) PBonus , sum(BonusTable.lbonus) LBonus, sum(BonusTable.Salary) Salary, bsid  FROM public.SalerSalary BonusTable group by bsid
	) BonusTable on BonusTable.bsid = BS.bsid
	left join (
		SELECT sum(cost) prepay , branch FROM public.prepay PP 
		inner join public.BranchPrePay BPP on PP.ppid = BPP.ppid 	
		where to_char(date ,'YYYY-MM') = $1
		group by branch
	) prepayTable on prepayTable.branch = BS.branch
	left join(
		SELECT sum(fee) pocket , branch FROM public.Pocket 		
		where circleid = $1
		group by branch
	) pocketTable on pocketTable.branch = BS.branch
	left join(
	    SELECT to_char(amor.date,'yyyy-MM') , branch , sum(cost) thismonthamor FROM public.amortization amor
		inner join (
			SELECT amorid, date, cost FROM public.amormap
		) amormap on amormap.amorid = amor.amorid
		where isover = 0 and to_char(amor.date,'yyyy-MM') = $1
		group by to_char(amor.date,'yyyy-MM') , amor.branch		
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
	) incomeexpense on incomeexpense.bsid = BS.bsid 	
	where date = $1
	) subtable
	ON CONFLICT (bsid) DO Nothing;
	`

	interdb := salaryM.imr.GetSQLDB()
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}
	//fmt.Println("BSID:" + bs.BSid)
	//fmt.Println(bs.Date)
	res, err := sqldb.Exec(sql, bs.Date)
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
	fmt.Println("CreateIncomEexpense:", id)

	if id == 0 {
		fmt.Println("CreateIncomEexpense, no create anyone ")
		return errors.New("CreateIncomeExpense, no create anyone ")
	}

	return nil
}

func (salaryM *SalaryModel) UpdateCommissionBSidAndStatus(bs *BranchSalary, cid []*Cid) (err error) {

	const sql = `Update public.commission as com
				set bsid = subquery.bsid, status = 'join'
				from (
				SELECT c.sid, c.rid, SS.bsid, to_char(r.date,'YYYY-MM')::varchar(50) date
				FROM public.receipt r
				inner join public.commission c on c.rid = r.rid and 
				to_timestamp(date_part('epoch',r.date)::int) >= $1 and to_timestamp(date_part('epoch',r.date)::int) < ( $1::date + '1 month'::interval)
				inner join public.SalerSalary SS on SS.date = to_char(r.date,'YYYY-MM') and SS.Sid = C.sid
				) AS subquery
				where com.sid = subquery.sid and com.rid = subquery.rid	;	
				`

	interdb := salaryM.imr.GetSQLDB()
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}
	//fmt.Println("BSID:" + bs.BSid)
	//fmt.Println(bs.Date)
	res, err := sqldb.Exec(sql, bs.Date+"-01")
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
		return errors.New("UpdateCommissionBSidAndStatus, not found any commission")
	}

	return nil
}

func (salaryM *SalaryModel) UpdateBranchSalaryTotal() (err error) {

	const sql = `UPDATE public.branchsalary BS
				set total = tmp.total
				FROM (
					SELECT sum(total) total, bsid  From public.salersalary					
					group by bsid 
				)as tmp where tmp.bsid = bs.bsid;	
				`
	//where date = $1
	interdb := salaryM.imr.GetSQLDB()
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

	return nil
}

func (salaryM *SalaryModel) GetSalerSalaryData(bsID, sid string) []*SalerSalary {
	//( (case when cb.sid is not null then ie.managerbonus else 0 end) + ss.tamount )
	sql := `SELECT ss.sid, ss.bsid, ss.sname, ss.date, ss.branch, ss.salary, ss.pbonus, ss.lbonus, ss.abonus, 
				ss.total, 
				ss.sp, ss.tax, ss.laborfee, ss.healthfee, ss.welfare, ss.CommercialFee, ss.other, 
				 tamount,
				COALESCE(ss.description,''), ss.workday , bs.lock,
				(case when cb.sid is not null then ie.managerbonus else 0 end) managerbonus
				FROM public.salersalary ss 
				inner join public.branchsalary bs on bs.bsid = ss.bsid
				left join public.incomeexpense ie on ie.bsid = ss.bsid
				left join  public.configbranch cb on cb.branch = ss.branch and ss.sid = cb.sid
				where ss.bsid = '%s' `
	//where ss.bsid = '%s' and ss.sid like '%s'`

	db := salaryM.imr.GetSQLDB()
	rows, err := db.SQLCommand(fmt.Sprintf(sql, bsID))
	if err != nil {
		return nil
	}
	var ssDataList []*SalerSalary

	for rows.Next() {
		var ss SalerSalary

		if err := rows.Scan(&ss.Sid, &ss.BSid, &ss.SName, &ss.Date, &ss.Branch, &ss.Salary, &ss.Pbonus, &ss.Lbonus, &ss.Abonus, &ss.Total,
			&ss.SP, &ss.Tax, &ss.LaborFee, &ss.HealthFee, &ss.Welfare, &ss.CommercialFee, &ss.Other, &ss.TAmount, &ss.Description, &ss.Workday, &ss.Lock, &ss.ManagerBonus); err != nil {
			fmt.Println("err Scan " + err.Error())
			return nil
		}

		ssDataList = append(ssDataList, &ss)
	}

	out, err := json.Marshal(ssDataList)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(out))
	salaryM.salerSalaryList = ssDataList

	if len(salaryM.salerSalaryList) > 0 {
		ss := salaryM.salerSalaryList[0]
		strtime, _ := util.ADtoROC(ss.Date, "file")
		salaryM.FnamePdf = ss.Branch + "薪資表" + strtime
		// systemM.GetAccountData(salaryM.salerSalaryList[0].Branch)

		// for _, salerSalary := range salaryM.salerSalaryList {
		// 	for _, element := range systemM.systemAccountList {
		// 		if element.Account == salerSalary.Sid {
		// 			fmt.Println(element, salerSalary)
		// 			util.RunSendMail(smtpConf.Host, smtpConf.Port, smtpConf.Password, smtpConf.User, "geassyayaoo3@gmail.com", "subject", body, fname)
		// 		}
		// 	}
		// }
		// body := "testbody"
		// fname := "hello.pdf"
		// //conf := di.GetSMTPConf()
		// fmt.Println(smtpConf)
	}

	return salaryM.salerSalaryList
}

func (salaryM *SalaryModel) GetIncomeExpenseData(bsID string) []*IncomeExpense {

	const spl = `SELECT bsid, sr, businesstax, salesamounts, pbonus, lbonus, amorcost, agentsign, rent, commercialfee, salary, prepay, pocket, annualbonus, salerfee, pretax, aftertax, earnadjust, lastloss, businessincometax, managerbonus, annualratio
	FROM public.incomeexpense where bsid = '%s';`
	db := salaryM.imr.GetSQLDB()
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

func (salaryM *SalaryModel) GetNHISalaryData(bsID string) []*NHISalary {

	fmt.Println("GetNHISalaryData")
	const spl = `SELECT nhi.sid, nhi.bsid, nhi.sname, nhi.payrollbracket, nhi.salary, nhi.pbonus, nhi.bonus, nhi.total, nhi.pd, nhi.salarybalance, nhi.fourbouns, ss.sp, nhi.foursp, nhi.ptsp
				FROM public.nhisalary nhi
				inner join (
					select bsid, sp from public.salersalary 
				) ss on ss.bsid = nhi.bsid
				where nhi.bsid = '%s'`

	db := salaryM.imr.GetSQLDB()
	rows, err := db.SQLCommand(fmt.Sprintf(spl, bsID))
	if err != nil {
		fmt.Println(err)
		return nil
	}
	var nhiDataList []*NHISalary

	for rows.Next() {
		var nhi NHISalary

		if err := rows.Scan(&nhi.Sid, &nhi.BSid, &nhi.SName, &nhi.PayrollBracket, &nhi.Salary, &nhi.Pbonus, &nhi.Bonus, &nhi.Total,
			&nhi.PD, &nhi.SalaryBalance, &nhi.FourBouns, &nhi.SP, &nhi.FourSP, &nhi.PTSP); err != nil {
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

func (salaryM *SalaryModel) ExportNHISalaryData(bsID string) []*NHISalary {

	fmt.Println("ExportNHISalaryData")
	const spl = `SELECT nhi.sid, nhi.bsid, nhi.sname, nhi.payrollbracket, nhi.salary, nhi.pbonus, nhi.bonus, nhi.total, nhi.pd, nhi.salarybalance, nhi.fourbouns, ss.sp, nhi.foursp, nhi.ptsp, cs.title , coalesce(ss.description,''), cs.branch
				FROM public.nhisalary nhi
				inner join (
					select sid, bsid, sp ,description from public.salersalary 
				) ss on ss.bsid = nhi.bsid and ss.sid = nhi.sid
				inner join(
					SELECT A.sid, A.sname, A.branch, A.percent, A.title
					FROM public.ConfigSaler A 
					Inner Join ( 
						select sid, max(zerodate) zerodate from public.configsaler cs 
						where now() > zerodate
						group by sid 
					) B on A.sid=B.sid and A.zeroDate = B.zeroDate
				) cs on cs.sid = nhi.sid
				where nhi.bsid = '%s'`

	db := salaryM.imr.GetSQLDB()
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

	// out, err := json.Marshal(arList)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(string(out))

	return salaryM.NHISalaryList
}

func (salaryM *SalaryModel) CreateNHISalary(year string) (err error) {

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

	interdb := salaryM.imr.GetSQLDB()
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

	return nil
}

func (salaryM *SalaryModel) UpdateSalerSalaryData(ss *SalerSalary, bsid string) (err error) {

	const sql = `UPDATE public.salersalary
				SET lbonus= $1, abonus= $2, total= salary + pbonus + $1 - $2, tax = $3, other = $4,  description= $5, workday= $6,
				laborfee = ( Case When $6 >= 30 then subquery.laborfee else subquery.laborfee * $6 / 30 END),
				healthfee = ( Case When $6 >= 30 then subquery.healthfee else 0 END),
				tamount = salary + pbonus + $1 - $2 - $3 - $4 - welfare - commercialFee - ( Case When $6 >= 30 then subquery.laborfee else subquery.laborfee * $6 / 30 END) - ( Case When $6 >= 30 then subquery.healthfee else 0 END)
				FROM(
					Select (A.payrollbracket * CP.li * 0.2 / 100) laborfee, (A.payrollbracket * CP.nhi * 0.2 / 100) healthfee FROM public.ConfigSaler A 
					Inner Join ( 
						select sid, max(zerodate) zerodate from public.configsaler cs 
						where now() > zerodate and Sid = $7
						group by sid 
					) B on A.sid=B.sid and A.zeroDate = B.zeroDate
					cross join ( 
						select  c.date, c.nhi, c.li, c.nhi2nd, c.mmw from public.ConfigParameter C
						inner join(
							select  max(date) date from public.ConfigParameter 
						) A on A.date = C.date limit 1
					) CP
				) as subquery
				WHERE sid= $7 and bsid = $8;`

	interdb := salaryM.imr.GetSQLDB()
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}
	//fmt.Println("BSID:" + bs.BSid)
	//fmt.Println(bs.Date)
	res, err := sqldb.Exec(sql, ss.Lbonus, ss.Abonus, ss.Tax, ss.Other, ss.Description, ss.Workday, ss.Sid, bsid)
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

	return nil
}

func (salaryM *SalaryModel) UpdateIncomeExpenseData(ie *IncomeExpense, bsid string) (err error) {

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

	interdb := salaryM.imr.GetSQLDB()
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

	ummb_err := salaryM.UpdateManagerByManagerBonus(bsid)
	if ummb_err != nil {
		return nil
		//return ucias_err
	}

	return nil
}

func (salaryM *SalaryModel) UpdateManagerByManagerBonus(bsid string) (err error) {

	const sql = `UPDATE public.salersalary salersalary
	SET total= subquery.Total , 
		sp = subquery.sp ,
		welfare = subquery.Total * 0.01,
		commercialfee = subquery.Total * subquery.CommercialFee / 100 ,
		tamount = subquery.Total - subquery.sp - subquery.tax - subquery.laborfee - subquery.healthfee - subquery.Total * 0.01 - subquery.Total * subquery.CommercialFee / 100 - subquery.other		
	FROM(
	SELECT ss.sid, ss.bsid, ss.sname, ss.date, ss.branch, ss.salary, ss.pbonus, ss.lbonus, ss.abonus, 
	 ss.salary + ss.pbonus + ss.lbonus - ss.abonus + (case when cb.sid is not null then ie.managerbonus else 0 end)	Total, 
	(CASE WHEN ss.salary = 0 and cs.association = 1 then 0 
		WHEN ss.salary = 0 and cs.association = 0 then COALESCE(ss.Salary+  ss.Pbonus,ss.Salary) * cp.nhi2nd / 100 
		else
			( CASE WHEN ((COALESCE(ss.Salary+  ss.Pbonus, ss.Salary)) - 4 * cs.PayrollBracket) > 0 then ((COALESCE(ss.Salary+  ss.Pbonus,ss.Salary)) - 4 * cs.PayrollBracket) * cp.nhi2nd / 100 else 0 end)
		end
	) sp,
	ss.tax, ss.laborfee, ss.healthfee, ss.welfare, cb.CommercialFee, ss.other, tamount,
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

	interdb := salaryM.imr.GetSQLDB()
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

	return nil
}

func (salaryM *SalaryModel) LockBranchSalary(bsid, lock string) (err error) {

	const sql = `UPDATE public.branchsalary	SET lock = $2	WHERE bsid = $1;`

	interdb := salaryM.imr.GetSQLDB()
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

	_err := salaryM.UpdateBranchSalaryTotal()
	if _err != nil {
		return nil
		//return css_err
	}

	return nil
}

func (salaryM *SalaryModel) addBranchSalaryInfoTable(table *pdf.DataTable, p *pdf.Pdf) (table_final *pdf.DataTable,
	T_Salary, T_Pbonus, T_Abonus, T_Lbonus, T_Total, T_SP, T_Tax, T_LaborFee, T_HealthFee, T_Welfare, T_CommercialFee, T_TAmount, T_Other int) {

	//text := "fd"
	//width := mypdf.MeasureTextWidth(text)
	//table.ColumnLen
	T_Salary, T_Pbonus, T_Abonus, T_Lbonus, T_Total = 0, 0, 0, 0, 0
	T_SP, T_Tax, T_LaborFee, T_HealthFee, T_Welfare, T_CommercialFee = 0, 0, 0, 0, 0, 0
	T_TAmount, T_Other = 0, 0

	for index, element := range salaryM.salerSalaryList {
		fmt.Println(index)
		fmt.Println(table.ColumnWidth[index])
		///
		text := strconv.Itoa(index + 1)
		pdf.ResizeWidth(table, p.GetTextWidth(text), 0)
		var vs = &pdf.TableStyle{
			Text:  text,
			Bg:    report.ColorWhite,
			Front: report.ColorTableLine,
		}
		table.RawData = append(table.RawData, vs)
		//
		text = element.SName
		pdf.ResizeWidth(table, p.GetTextWidth(text), 1)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    report.ColorWhite,
			Front: report.ColorTableLine,
		}
		table.RawData = append(table.RawData, vs)
		//底薪
		T_Salary += element.Salary
		text = strconv.Itoa(element.Salary)
		pdf.ResizeWidth(table, p.GetTextWidth(text), 2)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    report.ColorWhite,
			Front: report.ColorTableLine,
		}
		table.RawData = append(table.RawData, vs)
		//績效
		T_Pbonus += element.Pbonus
		text = strconv.Itoa(element.Pbonus)
		pdf.ResizeWidth(table, p.GetTextWidth(text), 3)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    report.ColorWhite,
			Front: report.ColorTableLine,
		}
		table.RawData = append(table.RawData, vs)
		//領導
		T_Lbonus += element.Lbonus
		text = strconv.Itoa(element.Lbonus)
		pdf.ResizeWidth(table, p.GetTextWidth(text), 4)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    report.ColorWhite,
			Front: report.ColorTableLine,
		}
		table.RawData = append(table.RawData, vs)
		//
		T_Abonus += element.Abonus
		text = strconv.Itoa(element.Abonus)
		pdf.ResizeWidth(table, p.GetTextWidth(text), 5)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    report.ColorWhite,
			Front: report.ColorTableLine,
		}
		table.RawData = append(table.RawData, vs)
		//
		T_Total += element.Total
		text = strconv.Itoa(element.Total)
		pdf.ResizeWidth(table, p.GetTextWidth(text), 6)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    If(true, report.ColorWhite, report.ColorWhite).(report.Color),
			Front: report.ColorTableLine,
		}
		table.RawData = append(table.RawData, vs)
		//
		T_SP += element.SP
		text = strconv.Itoa(element.SP)
		pdf.ResizeWidth(table, p.GetTextWidth(text), 7)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    If(true, report.ColorWhite, report.ColorWhite).(report.Color),
			Front: report.ColorTableLine,
		}
		table.RawData = append(table.RawData, vs)
		//
		T_Tax += element.Tax
		text = strconv.Itoa(element.Tax)
		pdf.ResizeWidth(table, p.GetTextWidth(text), 8)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    If(true, report.ColorWhite, report.ColorWhite).(report.Color),
			Front: report.ColorTableLine,
		}
		table.RawData = append(table.RawData, vs)
		//
		T_LaborFee += element.LaborFee
		text = strconv.Itoa(element.LaborFee)
		pdf.ResizeWidth(table, p.GetTextWidth(text), 9)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    If(true, report.ColorWhite, report.ColorWhite).(report.Color),
			Front: report.ColorTableLine,
		}
		table.RawData = append(table.RawData, vs)
		//
		T_HealthFee += element.HealthFee
		text = strconv.Itoa(element.HealthFee)
		pdf.ResizeWidth(table, p.GetTextWidth(text), 10)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    If(true, report.ColorWhite, report.ColorWhite).(report.Color),
			Front: report.ColorTableLine,
		}
		table.RawData = append(table.RawData, vs)
		//
		T_Welfare += element.Welfare
		text = strconv.Itoa(element.Welfare)
		pdf.ResizeWidth(table, p.GetTextWidth(text), 11)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    If(true, report.ColorWhite, report.ColorWhite).(report.Color),
			Front: report.ColorTableLine,
		}
		table.RawData = append(table.RawData, vs)
		//
		T_CommercialFee += +element.CommercialFee
		text = strconv.Itoa(element.CommercialFee)
		pdf.ResizeWidth(table, p.GetTextWidth(text), 12)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    If(true, report.ColorWhite, report.ColorWhite).(report.Color),
			Front: report.ColorTableLine,
		}
		table.RawData = append(table.RawData, vs)
		//
		T_Other += element.Other
		text = strconv.Itoa(element.Other)
		pdf.ResizeWidth(table, p.GetTextWidth(text), 13)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    If(true, report.ColorWhite, report.ColorWhite).(report.Color),
			Front: report.ColorTableLine,
		}
		table.RawData = append(table.RawData, vs)
		//
		T_TAmount += element.TAmount
		text = strconv.Itoa(element.TAmount)
		pdf.ResizeWidth(table, p.GetTextWidth(text), 14)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    If(true, report.ColorWhite, report.ColorWhite).(report.Color),
			Front: report.ColorTableLine,
		}
		table.RawData = append(table.RawData, vs)
		//
		text = element.Description
		pdf.ResizeWidth(table, p.GetTextWidth(text), 15)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    If(true, report.ColorWhite, report.ColorWhite).(report.Color),
			Front: report.ColorTableLine,
		}
		table.RawData = append(table.RawData, vs)
	}
	table_final = table
	return
}

func (salaryM *SalaryModel) addSalerSalaryInfoTable(table *pdf.DataTable, p *pdf.Pdf, index int, element *SalerSalary) (table_final *pdf.DataTable) {

	//text := "fd"
	//width := mypdf.MeasureTextWidth(text)
	//table.ColumnLen

	//for index, element := range salaryM.salerSalaryList {
	fmt.Println(index)

	///
	text := strconv.Itoa(index + 1)
	pdf.ResizeWidth(table, p.GetTextWidth(text), 0)
	var vs = &pdf.TableStyle{
		Text:  text,
		Bg:    report.ColorWhite,
		Front: report.ColorTableLine,
	}
	table.RawData = append(table.RawData, vs)
	//
	text = element.SName
	pdf.ResizeWidth(table, p.GetTextWidth(text), 1)
	vs = &pdf.TableStyle{
		Text:  text,
		Bg:    report.ColorWhite,
		Front: report.ColorTableLine,
	}
	table.RawData = append(table.RawData, vs)
	//底薪

	text = strconv.Itoa(element.Salary)
	pdf.ResizeWidth(table, p.GetTextWidth(text), 2)
	vs = &pdf.TableStyle{
		Text:  text,
		Bg:    report.ColorWhite,
		Front: report.ColorTableLine,
	}
	table.RawData = append(table.RawData, vs)
	//績效
	text = strconv.Itoa(element.Pbonus)
	pdf.ResizeWidth(table, p.GetTextWidth(text), 3)
	vs = &pdf.TableStyle{
		Text:  text,
		Bg:    report.ColorWhite,
		Front: report.ColorTableLine,
	}
	table.RawData = append(table.RawData, vs)
	//領導

	text = strconv.Itoa(element.Lbonus)
	pdf.ResizeWidth(table, p.GetTextWidth(text), 4)
	vs = &pdf.TableStyle{
		Text:  text,
		Bg:    report.ColorWhite,
		Front: report.ColorTableLine,
	}
	table.RawData = append(table.RawData, vs)
	//

	text = strconv.Itoa(element.Abonus)
	pdf.ResizeWidth(table, p.GetTextWidth(text), 5)
	vs = &pdf.TableStyle{
		Text:  text,
		Bg:    report.ColorWhite,
		Front: report.ColorTableLine,
	}
	table.RawData = append(table.RawData, vs)
	//

	text = strconv.Itoa(element.Total)
	pdf.ResizeWidth(table, p.GetTextWidth(text), 6)
	vs = &pdf.TableStyle{
		Text:  text,
		Bg:    If(true, report.ColorWhite, report.ColorWhite).(report.Color),
		Front: report.ColorTableLine,
	}
	table.RawData = append(table.RawData, vs)
	//

	text = strconv.Itoa(element.SP)
	pdf.ResizeWidth(table, p.GetTextWidth(text), 7)
	vs = &pdf.TableStyle{
		Text:  text,
		Bg:    If(true, report.ColorWhite, report.ColorWhite).(report.Color),
		Front: report.ColorTableLine,
	}
	table.RawData = append(table.RawData, vs)
	//

	text = strconv.Itoa(element.Tax)
	pdf.ResizeWidth(table, p.GetTextWidth(text), 8)
	vs = &pdf.TableStyle{
		Text:  text,
		Bg:    If(true, report.ColorWhite, report.ColorWhite).(report.Color),
		Front: report.ColorTableLine,
	}
	table.RawData = append(table.RawData, vs)
	//

	text = strconv.Itoa(element.LaborFee)
	pdf.ResizeWidth(table, p.GetTextWidth(text), 9)
	vs = &pdf.TableStyle{
		Text:  text,
		Bg:    If(true, report.ColorWhite, report.ColorWhite).(report.Color),
		Front: report.ColorTableLine,
	}
	table.RawData = append(table.RawData, vs)
	//

	text = strconv.Itoa(element.HealthFee)
	pdf.ResizeWidth(table, p.GetTextWidth(text), 10)
	vs = &pdf.TableStyle{
		Text:  text,
		Bg:    If(true, report.ColorWhite, report.ColorWhite).(report.Color),
		Front: report.ColorTableLine,
	}
	table.RawData = append(table.RawData, vs)
	//

	text = strconv.Itoa(element.Welfare)
	pdf.ResizeWidth(table, p.GetTextWidth(text), 11)
	vs = &pdf.TableStyle{
		Text:  text,
		Bg:    If(true, report.ColorWhite, report.ColorWhite).(report.Color),
		Front: report.ColorTableLine,
	}
	table.RawData = append(table.RawData, vs)
	//

	text = strconv.Itoa(element.CommercialFee)
	pdf.ResizeWidth(table, p.GetTextWidth(text), 12)
	vs = &pdf.TableStyle{
		Text:  text,
		Bg:    If(true, report.ColorWhite, report.ColorWhite).(report.Color),
		Front: report.ColorTableLine,
	}
	table.RawData = append(table.RawData, vs)
	//

	text = strconv.Itoa(element.Other)
	pdf.ResizeWidth(table, p.GetTextWidth(text), 13)
	vs = &pdf.TableStyle{
		Text:  text,
		Bg:    If(true, report.ColorWhite, report.ColorWhite).(report.Color),
		Front: report.ColorTableLine,
	}
	table.RawData = append(table.RawData, vs)
	//

	text = strconv.Itoa(element.TAmount)
	pdf.ResizeWidth(table, p.GetTextWidth(text), 14)
	vs = &pdf.TableStyle{
		Text:  text,
		Bg:    If(true, report.ColorWhite, report.ColorWhite).(report.Color),
		Front: report.ColorTableLine,
	}
	table.RawData = append(table.RawData, vs)
	//
	text = element.Description
	pdf.ResizeWidth(table, p.GetTextWidth(text), 15)
	vs = &pdf.TableStyle{
		Text:  text,
		Bg:    If(true, report.ColorWhite, report.ColorWhite).(report.Color),
		Front: report.ColorTableLine,
	}
	table.RawData = append(table.RawData, vs)
	//}
	table_final = table
	return
}

func (salaryM *SalaryModel) ExportSR(bsID string) {
	const qsql = `SELECT ss.sid, ss.sname ,  coalesce(sum(tmp.SR),0)  ,coalesce( sum( tmp.SR * cs.percent/100)  , 0 ) bonus , cs.branch , ss.date
	from salersalary ss
		left join(
			SELECT c.bsid, c.sid, c.rid,  ((r.amount - coalesce(d.fee,0) )* c.cpercent/100) sr					
			FROM public.commission c
			inner JOIN public.receipt r on r.rid = c.rid		
			left join(
				select rid, fee from public.deduct
			) d on d.rid = r.rid		
		) tmp on ss.bsid = tmp.bsid and tmp.sid = ss.sid
	Inner Join (
		SELECT A.sid, A.branch, A.percent
			FROM public.ConfigSaler A
			Inner Join (
				select sid, max(zerodate) zerodate from public.configsaler cs
				where now() > zerodate
				group by sid
			) B on A.sid=B.sid and A.zeroDate = B.zeroDate
	) cs on cs.sid=ss.sid 
	where ss.bsid = '%s'
	group by ss.sid , ss.sname , cs.branch , ss.date`

	db := cm.imr.GetSQLDB()

	cDataList := []*Commission{}
	salerList := []*SystemAccount{}

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

	} else {
		salerList = nil
	}

	out, _ := json.Marshal(salerList)
	fmt.Println("salerList :", string(out))
	out, _ = json.Marshal(cDataList)
	fmt.Println("cDataList :", string(out))
	//salaryM.SystemAccountList = salerList
	salaryM.CommissionList = cDataList
	return

}

func (salaryM *SalaryModel) GetSalerCommission(bsID string) {
	const qsql = `SELECT ss.sid, ss.sname , tmp.item, tmp.amount, tmp.fee, tmp.cpercent, tmp.sr, ( (tmp.amount - coalesce(tmp.fee,0) )* tmp.cpercent/100 * cs.percent/100) bonus , tmp.remark , cs.branch   from salersalary ss
				left join(
					SELECT c.bsid, c.sid, c.rid, r.date, c.item, r.amount, 0 , c.sname, c.cpercent, ( (r.amount - coalesce(d.fee,0) )* c.cpercent/100) sr, 
					r.arid, c.status ,  to_char(r.date,'yyyy-MM-dd') , COALESCE(NULLIF(r.invoiceno, null),'') , coalesce(d.checknumber,'') , coalesce(d.fee,0) fee , coalesce(d.item,'') remark
					FROM public.commission c
					inner JOIN public.receipt r on r.rid = c.rid		
					left join(
						select rid, checknumber , fee, item from public.deduct
					) d on d.rid = r.rid		
				) tmp on ss.bsid = tmp.bsid and tmp.sid = ss.sid
			Inner Join (
				SELECT A.sid, A.branch, A.percent, A.title
					FROM public.ConfigSaler A
					Inner Join (
						select sid, max(zerodate) zerodate from public.configsaler cs
						where now() > zerodate
						group by sid
					) B on A.sid=B.sid and A.zeroDate = B.zeroDate
			) cs on cs.sid=ss.sid 
			where ss.bsid = '%s' order by ss.sid asc;`

	// const qsql = `SELECT ss.sid, ss.sname , tmp.item, tmp.amount, tmp.fee, tmp.cpercent, tmp.sr, tmp.bonus, tmp.remark from salersalary ss
	// left join(
	// 	SELECT c.bsid, c.sid, c.rid, r.date, c.item, r.amount, 0 , c.sname, c.cpercent, ( (r.amount - coalesce(d.fee,0) )* c.cpercent/100) sr, ( (r.amount - coalesce(d.fee,0) )* c.cpercent/100 * cs.percent/100) bonus,
	// 	r.arid, c.status , cs.branch, cs.percent, to_char(r.date,'yyyy-MM-dd') , COALESCE(NULLIF(r.invoiceno, null),'') , coalesce(d.checknumber,'') , coalesce(d.fee,0) fee , coalesce(d.item,'') remark
	// 	FROM public.commission c
	// 	inner JOIN public.receipt r on r.rid = c.rid
	// 	Inner Join (
	// 			SELECT A.sid, A.branch, A.percent, A.title
	// 				FROM public.ConfigSaler A
	// 				Inner Join (
	// 					select sid, max(zerodate) zerodate from public.configsaler cs
	// 					where now() > zerodate
	// 					group by sid
	// 				) B on A.sid=B.sid and A.zeroDate = B.zeroDate
	// 		) cs on c.sid=cs.sid
	// 	left join(
	// 		select rid, checknumber , fee, item from public.deduct
	// 	) d on d.rid = r.rid
	// ) tmp on ss.bsid = tmp.bsid and tmp.sid = ss.sid
	// where ss.bsid = '%s' order by ss.sid asc;`

	//left JOIN (select sum(fee) fee, count(rid) ,arid from public.deduct group by arid) as tmp on tmp.arid = r.arid
	db := cm.imr.GetSQLDB()

	cDataList := []*Commission{}
	salerList := []*SystemAccount{}

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
		if err := rows.Scan(&c.Sid, &c.SName, &Item, &Amount, &Fee, &CPercent, &SR, &Bonus, &DedectItem, &Branch); err != nil {
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
		cDataList = append(cDataList, &c)
	}
	if len(cDataList) > 0 {
		//systemM.GetAccountData()
		//salerList = systemM.systemAccountList
	} else {
		//salerList = nil
	}

	out, _ := json.Marshal(salerList)
	fmt.Println("salerList :", string(out))
	out, _ = json.Marshal(cDataList)
	fmt.Println("cDataList :", string(out))
	//salaryM.SystemAccountList = salerList
	salaryM.CommissionList = cDataList
	return
}

func (salaryM *SalaryModel) GetAgentSign(bsID string) {
	const qsql = `SELECT ss.sid, ss.sname , tmp.item, tmp.amount, tmp.fee, tmp.cpercent, tmp.sr, ( (tmp.amount - coalesce(tmp.fee,0) )* tmp.cpercent/100 * cs.percent/100) bonus , tmp.remark , cs.branch, cs.percent   from salersalary ss
				inner join(
					SELECT c.bsid, c.sid, c.rid, r.date, c.item, r.amount, 0 , c.sname, c.cpercent, ( (r.amount - coalesce(d.fee,0) )* c.cpercent/100) sr, 
					r.arid, c.status ,  to_char(r.date,'yyyy-MM-dd') , COALESCE(NULLIF(r.invoiceno, null),'') , coalesce(d.checknumber,'') , coalesce(d.fee,0) fee , coalesce(d.item,'') remark
					FROM public.commission c
					inner JOIN public.receipt r on r.rid = c.rid		
					left join(
						select rid, checknumber , fee, item from public.deduct
					) d on d.rid = r.rid		
				) tmp on ss.bsid = tmp.bsid and tmp.sid = ss.sid
			Inner Join (
				SELECT A.sid, A.branch, A.percent, A.title
					FROM public.ConfigSaler A
					Inner Join (
						select sid, max(zerodate) zerodate from public.configsaler cs
						where now() > zerodate
						group by sid
					) B on A.sid=B.sid and A.zeroDate = B.zeroDate
			) cs on cs.sid=ss.sid 
			where ss.bsid = '%s' order by ss.sid asc;`

	db := cm.imr.GetSQLDB()

	// cDataList := []*Commission{}
	// salerList := []*SystemAccount{}

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
	//這邊在做啥小?
	if len(salaryM.CommissionList) > 0 {
		systemM.GetAccountData()
		salerList := systemM.systemAccountList
		for _, element := range salerList {
			salaryM.SystemAccountList = append(salaryM.SystemAccountList, element)
		}
	}

	return
}

func (salaryM *SalaryModel) ExportIncomeTaxReturn(bsID string) {

	const dsql = `SELECT ss.date , ss.branch from salersalary ss where ss.bsid = '%s'`

	const qsql = `SELECT ss.sid, ss.sname , cs.identitynum, cs.address, ss.total,  cs.branch, ss.date from salersalary ss			
				Inner Join (
					SELECT A.sid, A.branch, A.identitynum, A.address , A.bankaccount
						FROM public.ConfigSaler A
						Inner Join (
							select sid, max(zerodate) zerodate from public.configsaler cs
							where now() > zerodate
							group by sid
						) B on A.sid=B.sid and A.zeroDate = B.zeroDate
				) cs on cs.sid=ss.sid 
				where ss.date||'-01' >= '%s' and ss.date||'-01' <= '%s' and ss.branch = '%s'	
				order by ss.sid asc;`

	db := cm.imr.GetSQLDB()

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
		return
	}
	//藉由日期區間查詢
	year := date[0:3]
	rows, err = db.SQLCommand(fmt.Sprintf(qsql, year+"-01-01", date+"-01", branch))
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

	out, _ := json.Marshal(configM.ConfigSalerList)
	fmt.Println("cDataList :", string(out))

	return
}

func (salaryM *SalaryModel) ExportPayrollTransfer(bsID string) {
	const qsql = `SELECT ss.sid, ss.sname , cs.identitynum, cs.bankaccount, ss.tamount,  cs.branch   from salersalary ss			
			Inner Join (
				SELECT A.sid, A.branch, A.identitynum, A.title , A.bankaccount
					FROM public.ConfigSaler A
					Inner Join (
						select sid, max(zerodate) zerodate from public.configsaler cs
						where now() > zerodate
						group by sid
					) B on A.sid=B.sid and A.zeroDate = B.zeroDate
			) cs on cs.sid=ss.sid 
			where ss.bsid = '%s'
			order by ss.sid asc;`

	db := cm.imr.GetSQLDB()

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
					Bg:    report.ColorWhite,
					Front: report.ColorTableLine,
				}
				table.RawData = append(table.RawData, vs)
				table.RawData = append(table.RawData, vs)
				table.RawData = append(table.RawData, vs)

				//
				fmt.Println("element.SName:", element.SName, "element.Sid:", element.SName, "   sid:", sid)
				text = sname
				pdf.ResizeWidth(table, p.GetTextWidth(text), 3)
				vs = &pdf.TableStyle{
					Text:  text,
					Bg:    report.ColorWhite,
					Front: report.ColorTableLine,
				}
				table.RawData = append(table.RawData, vs)
				//
				text = "合計"
				pdf.ResizeWidth(table, p.GetTextWidth(text), 4)
				vs = &pdf.TableStyle{
					Text:  text,
					Bg:    report.ColorWhite,
					Front: report.ColorTableLine,
				}
				table.RawData = append(table.RawData, vs)
				//
				text = fmt.Sprintf("%.1f", tmp_SR)
				pdf.ResizeWidth(table, p.GetTextWidth(text), 5)
				vs = &pdf.TableStyle{
					Text:  text,
					Bg:    report.ColorWhite,
					Front: report.ColorTableLine,
				}
				table.RawData = append(table.RawData, vs)
				//
				text = fmt.Sprintf("%.1f", tmp_Bonus)
				pdf.ResizeWidth(table, p.GetTextWidth(text), 6)
				vs = &pdf.TableStyle{
					Text:  text,
					Bg:    If(true, report.ColorWhite, report.ColorWhite).(report.Color),
					Front: report.ColorTableLine,
				}
				table.RawData = append(table.RawData, vs)
				//
				text = ""
				vs = &pdf.TableStyle{
					Text:  text,
					Bg:    report.ColorWhite,
					Front: report.ColorTableLine,
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
				Bg:    report.ColorWhite,
				Front: report.ColorTableLine,
			}
			table.RawData = append(table.RawData, vs)
			//
			text = strconv.Itoa(element.Amount)
			pdf.ResizeWidth(table, p.GetTextWidth(text), 1)
			vs = &pdf.TableStyle{
				Text:  text,
				Bg:    report.ColorWhite,
				Front: report.ColorTableLine,
			}
			table.RawData = append(table.RawData, vs)
			//應扣
			text = strconv.Itoa(element.Fee)
			pdf.ResizeWidth(table, p.GetTextWidth(text), 2)
			vs = &pdf.TableStyle{
				Text:  text,
				Bg:    report.ColorWhite,
				Front: report.ColorTableLine,
			}
			table.RawData = append(table.RawData, vs)
			//
			fmt.Println("element.SName:", element.SName, "element.Sid:", element.SName, "   sid:", sid)
			text = element.SName
			sname = element.SName
			pdf.ResizeWidth(table, p.GetTextWidth(text), 3)
			vs = &pdf.TableStyle{
				Text:  text,
				Bg:    report.ColorWhite,
				Front: report.ColorTableLine,
			}
			table.RawData = append(table.RawData, vs)
			//
			text = fmt.Sprintf("%.f%s", element.CPercent, "%")
			pdf.ResizeWidth(table, p.GetTextWidth(text), 4)
			vs = &pdf.TableStyle{
				Text:  text,
				Bg:    report.ColorWhite,
				Front: report.ColorTableLine,
			}
			table.RawData = append(table.RawData, vs)
			//
			T_SR += element.SR
			tmp_SR += element.SR
			text = fmt.Sprintf("%.1f", element.SR)
			pdf.ResizeWidth(table, p.GetTextWidth(text), 5)
			vs = &pdf.TableStyle{
				Text:  text,
				Bg:    report.ColorWhite,
				Front: report.ColorTableLine,
			}
			table.RawData = append(table.RawData, vs)
			//
			T_Bonus += element.Bonus
			tmp_Bonus += element.Bonus
			text = fmt.Sprintf("%.1f", element.Bonus)
			pdf.ResizeWidth(table, p.GetTextWidth(text), 6)
			vs = &pdf.TableStyle{
				Text:  text,
				Bg:    If(true, report.ColorWhite, report.ColorWhite).(report.Color),
				Front: report.ColorTableLine,
			}
			table.RawData = append(table.RawData, vs)
			//
			text = element.DedectItem
			pdf.ResizeWidth(table, p.GetTextWidth(text), 7)
			vs = &pdf.TableStyle{
				Text:  text,
				Bg:    If(true, report.ColorWhite, report.ColorWhite).(report.Color),
				Front: report.ColorTableLine,
			}
			table.RawData = append(table.RawData, vs)
			//
			text = element.Branch
			pdf.ResizeWidth(table, p.GetTextWidth(text), 8)
			vs = &pdf.TableStyle{
				Text:  text,
				Bg:    If(true, report.ColorWhite, report.ColorWhite).(report.Color),
				Front: report.ColorTableLine,
			}
			table.RawData = append(table.RawData, vs)
			//
			text = element.Percent
			pdf.ResizeWidth(table, p.GetTextWidth(text), 9)
			vs = &pdf.TableStyle{
				Text:  text,
				Bg:    If(true, report.ColorWhite, report.ColorWhite).(report.Color),
				Front: report.ColorTableLine,
			}
			table.RawData = append(table.RawData, vs)
		}
		//最後一筆合計 hard code
		if i == len(salaryM.CommissionList)-1 {
			///
			text := ""
			var vs = &pdf.TableStyle{
				Text:  text,
				Bg:    report.ColorWhite,
				Front: report.ColorTableLine,
			}
			table.RawData = append(table.RawData, vs)
			table.RawData = append(table.RawData, vs)
			table.RawData = append(table.RawData, vs)

			//
			fmt.Println("element.SName:", element.SName, "element.Sid:", element.SName, "   sid:", sid)
			text = sname
			pdf.ResizeWidth(table, p.GetTextWidth(text), 3)
			vs = &pdf.TableStyle{
				Text:  text,
				Bg:    report.ColorWhite,
				Front: report.ColorTableLine,
			}
			table.RawData = append(table.RawData, vs)
			//
			text = "合計"
			pdf.ResizeWidth(table, p.GetTextWidth(text), 4)
			vs = &pdf.TableStyle{
				Text:  text,
				Bg:    report.ColorWhite,
				Front: report.ColorTableLine,
			}
			table.RawData = append(table.RawData, vs)
			//
			text = fmt.Sprintf("%.1f", tmp_SR)
			pdf.ResizeWidth(table, p.GetTextWidth(text), 5)
			vs = &pdf.TableStyle{
				Text:  text,
				Bg:    report.ColorWhite,
				Front: report.ColorTableLine,
			}
			table.RawData = append(table.RawData, vs)
			//
			text = fmt.Sprintf("%.1f", tmp_Bonus)
			pdf.ResizeWidth(table, p.GetTextWidth(text), 6)
			vs = &pdf.TableStyle{
				Text:  text,
				Bg:    If(true, report.ColorWhite, report.ColorWhite).(report.Color),
				Front: report.ColorTableLine,
			}
			table.RawData = append(table.RawData, vs)
			//
			text = ""
			vs = &pdf.TableStyle{
				Text:  text,
				Bg:    report.ColorWhite,
				Front: report.ColorTableLine,
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
	T_SR, T_Bonus float64) {

	//text := "fd"
	//width := mypdf.MeasureTextWidth(text)
	//table.ColumnLen
	T_SR, T_Bonus = 0.0, 0.0

	for _, element := range cList {
		if element.Sid == sid {
			///
			text := element.Item
			pdf.ResizeWidth(table, p.GetTextWidth(text), 0)
			var vs = &pdf.TableStyle{
				Text:  text,
				Bg:    report.ColorWhite,
				Front: report.ColorTableLine,
			}
			table.RawData = append(table.RawData, vs)
			//
			text = strconv.Itoa(element.Amount)
			pdf.ResizeWidth(table, p.GetTextWidth(text), 1)
			vs = &pdf.TableStyle{
				Text:  text,
				Bg:    report.ColorWhite,
				Front: report.ColorTableLine,
			}
			table.RawData = append(table.RawData, vs)
			//應扣
			text = strconv.Itoa(element.Fee)
			pdf.ResizeWidth(table, p.GetTextWidth(text), 2)
			vs = &pdf.TableStyle{
				Text:  text,
				Bg:    report.ColorWhite,
				Front: report.ColorTableLine,
			}
			table.RawData = append(table.RawData, vs)
			//
			fmt.Println("element.SName:", element.SName, "element.Sid:", element.SName, "   sid:", sid)
			text = element.SName
			pdf.ResizeWidth(table, p.GetTextWidth(text), 3)
			vs = &pdf.TableStyle{
				Text:  text,
				Bg:    report.ColorWhite,
				Front: report.ColorTableLine,
			}
			table.RawData = append(table.RawData, vs)
			//
			text = fmt.Sprintf("%.f", element.CPercent)
			pdf.ResizeWidth(table, p.GetTextWidth(text), 4)
			vs = &pdf.TableStyle{
				Text:  text,
				Bg:    report.ColorWhite,
				Front: report.ColorTableLine,
			}
			table.RawData = append(table.RawData, vs)
			//
			T_SR += element.SR
			text = fmt.Sprintf("%.f", element.SR)
			pdf.ResizeWidth(table, p.GetTextWidth(text), 5)
			vs = &pdf.TableStyle{
				Text:  text,
				Bg:    report.ColorWhite,
				Front: report.ColorTableLine,
			}
			table.RawData = append(table.RawData, vs)
			//
			T_Bonus += element.Bonus
			text = fmt.Sprintf("%.f", element.Bonus)
			pdf.ResizeWidth(table, p.GetTextWidth(text), 6)
			vs = &pdf.TableStyle{
				Text:  text,
				Bg:    If(true, report.ColorWhite, report.ColorWhite).(report.Color),
				Front: report.ColorTableLine,
			}
			table.RawData = append(table.RawData, vs)
			//
			text = element.DedectItem
			pdf.ResizeWidth(table, p.GetTextWidth(text), 7)
			vs = &pdf.TableStyle{
				Text:  text,
				Bg:    If(true, report.ColorWhite, report.ColorWhite).(report.Color),
				Front: report.ColorTableLine,
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
			Bg:    report.ColorWhite,
			Front: report.ColorTableLine,
		}
		table.RawData = append(table.RawData, vs)
		//SR
		T_SR += int(element.SR)
		text = fmt.Sprintf("%.f", element.SR)
		pdf.ResizeWidth(table, p.GetTextWidth(text), 1)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    report.ColorWhite,
			Front: report.ColorTableLine,
		}
		table.RawData = append(table.RawData, vs)
		//績效
		T_Bonus += int(element.Bonus)
		text = fmt.Sprintf("%.f", element.Bonus)
		pdf.ResizeWidth(table, p.GetTextWidth(text), 2)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    report.ColorWhite,
			Front: report.ColorTableLine,
		}
		table.RawData = append(table.RawData, vs)
	}

	text := "合計"
	pdf.ResizeWidth(table, p.GetTextWidth(text), 0)
	vs := &pdf.TableStyle{
		Text:  text,
		Bg:    report.ColorWhite,
		Front: report.ColorTableLine,
	}
	table.RawData = append(table.RawData, vs)
	text = strconv.Itoa(T_SR)
	pdf.ResizeWidth(table, p.GetTextWidth(text), 1)
	vs = &pdf.TableStyle{
		Text:  text,
		Bg:    report.ColorWhite,
		Front: report.ColorTableLine,
	}
	table.RawData = append(table.RawData, vs)
	text = strconv.Itoa(T_Bonus)
	pdf.ResizeWidth(table, p.GetTextWidth(text), 1)
	vs = &pdf.TableStyle{
		Text:  text,
		Bg:    report.ColorWhite,
		Front: report.ColorTableLine,
	}
	table.RawData = append(table.RawData, vs)

	table_final = table
	return
}
func (salaryM *SalaryModel) addNHIInfoTable(table *pdf.DataTable, p *pdf.Pdf) (table_final *pdf.DataTable,
	T_PayrollBracket, T_Salary, T_Pbonus, T_Bonus, T_Total, T_Balance, T_PTSP,
	T_PD, T_FourBouns, T_SP, T_FourSP, T_Tax, T_SPB int) {

	fmt.Println("table.ColumnWidth[index]", len(table.ColumnWidth))

	T_PayrollBracket, T_Salary, T_Pbonus, T_Bonus, T_Total, T_Balance = 0, 0, 0, 0, 0, 0
	T_PTSP, T_PD, T_FourBouns, T_SP, T_FourSP, T_Tax, T_SPB = 0, 0, 0, 0, 0, 0, 0

	Branch := ""

	for _, element := range salaryM.NHISalaryList {

		if element.Branch != Branch {
			Branch = element.Branch
			pdf.ResizeWidth(table, p.GetTextWidth(Branch), 0)
			vs := &pdf.TableStyle{
				Text:  Branch,
				Bg:    report.ColorWhite,
				Front: report.ColorTableLine,
			}
			table.RawData = append(table.RawData, vs)
			for i := 1; i < 17; i++ {
				vs := &pdf.TableStyle{
					Text:  "",
					Bg:    report.ColorWhite,
					Front: report.ColorTableLine,
				}
				table.RawData = append(table.RawData, vs)
			}
		}
		//
		text := element.SName
		pdf.ResizeWidth(table, p.GetTextWidth(text), 0)
		vs := &pdf.TableStyle{
			Text:  text,
			Bg:    report.ColorWhite,
			Front: report.ColorTableLine,
		}
		table.RawData = append(table.RawData, vs)
		//
		T_PayrollBracket += element.PayrollBracket
		text = strconv.Itoa(element.PayrollBracket)
		pdf.ResizeWidth(table, p.GetTextWidth(text), 1)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    report.ColorWhite,
			Front: report.ColorTableLine,
		}
		table.RawData = append(table.RawData, vs)
		//
		T_Salary += element.Salary
		text = strconv.Itoa(element.Salary)
		pdf.ResizeWidth(table, p.GetTextWidth(text), 2)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    report.ColorWhite,
			Front: report.ColorTableLine,
		}
		table.RawData = append(table.RawData, vs)
		//
		T_Pbonus += element.Pbonus
		text = strconv.Itoa(element.Pbonus)
		pdf.ResizeWidth(table, p.GetTextWidth(text), 3)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    report.ColorWhite,
			Front: report.ColorTableLine,
		}
		table.RawData = append(table.RawData, vs)
		//
		T_Bonus += element.Bonus
		text = strconv.Itoa(element.Bonus)
		pdf.ResizeWidth(table, p.GetTextWidth(text), 4)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    report.ColorWhite,
			Front: report.ColorTableLine,
		}
		table.RawData = append(table.RawData, vs)
		//兼職
		text = "0"
		pdf.ResizeWidth(table, p.GetTextWidth(text), 5)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    report.ColorWhite,
			Front: report.ColorTableLine,
		}
		table.RawData = append(table.RawData, vs)
		//
		T_Total += element.Total
		text = strconv.Itoa(element.Total)
		pdf.ResizeWidth(table, p.GetTextWidth(text), 6)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    report.ColorWhite,
			Front: report.ColorTableLine,
		}
		table.RawData = append(table.RawData, vs)
		//
		T_Balance += element.SalaryBalance
		text = strconv.Itoa(element.SalaryBalance)
		pdf.ResizeWidth(table, p.GetTextWidth(text), 7)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    report.ColorWhite,
			Front: report.ColorTableLine,
		}
		table.RawData = append(table.RawData, vs)
		//
		T_PD += element.PD
		text = strconv.Itoa(element.PD)
		pdf.ResizeWidth(table, p.GetTextWidth(text), 8)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    report.ColorWhite,
			Front: report.ColorTableLine,
		}
		table.RawData = append(table.RawData, vs)
		//
		T_FourBouns += element.FourBouns
		text = strconv.Itoa(element.FourBouns)
		pdf.ResizeWidth(table, p.GetTextWidth(text), 9)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    report.ColorWhite,
			Front: report.ColorTableLine,
		}
		table.RawData = append(table.RawData, vs)
		//補充保費薪資差額 4倍-薪資差額
		T_SPB += element.FourBouns - element.SalaryBalance
		text = strconv.Itoa(element.FourBouns - element.SalaryBalance)
		pdf.ResizeWidth(table, p.GetTextWidth(text), 10)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    report.ColorWhite,
			Front: report.ColorTableLine,
		}
		table.RawData = append(table.RawData, vs)
		//
		T_FourSP += element.FourSP
		text = strconv.Itoa(element.FourSP)
		pdf.ResizeWidth(table, p.GetTextWidth(text), 11)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    report.ColorWhite,
			Front: report.ColorTableLine,
		}
		table.RawData = append(table.RawData, vs)
		//
		T_FourSP += element.FourSP
		text = strconv.Itoa(element.FourSP)
		pdf.ResizeWidth(table, p.GetTextWidth(text), 12)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    report.ColorWhite,
			Front: report.ColorTableLine,
		}
		table.RawData = append(table.RawData, vs)
		//
		T_PTSP += element.PTSP
		text = strconv.Itoa(element.PTSP)
		pdf.ResizeWidth(table, p.GetTextWidth(text), 13)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    report.ColorWhite,
			Front: report.ColorTableLine,
		}
		table.RawData = append(table.RawData, vs)
		//
		T_Tax += int(float64(element.Total) * 0.05)
		text = strconv.Itoa(int(float64(element.Total) * 0.05))
		pdf.ResizeWidth(table, p.GetTextWidth(text), 14)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    report.ColorWhite,
			Front: report.ColorTableLine,
		}
		table.RawData = append(table.RawData, vs)
		//
		text = element.Title
		pdf.ResizeWidth(table, p.GetTextWidth(text), 15)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    report.ColorWhite,
			Front: report.ColorTableLine,
		}
		table.RawData = append(table.RawData, vs)
		//
		text = element.Description
		pdf.ResizeWidth(table, p.GetTextWidth(text), 16)
		vs = &pdf.TableStyle{
			Text:  text,
			Bg:    report.ColorWhite,
			Front: report.ColorTableLine,
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
		fmt.Println(element)
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
		table.RawData["D"+strconv.Itoa(index+2-offset)] = strconv.Itoa(element.Tamount)
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
		fmt.Println(element)
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

func (salaryM *SalaryModel) getSalerEmail(things ...string) ([]*ConfigSaler, error) {

	branch := "%"
	for _, it := range things {
		branch = it
	}

	const qspl = `SELECT A.sid, A.sname, A.branch, A.Email
					FROM public.ConfigSaler A 
					Inner Join ( 
						select sid, max(zerodate) zerodate from public.configsaler cs 
						where now() > zerodate
						group by sid 
					) B on A.sid=B.sid and A.zeroDate = B.zeroDate
					where A.branch like '%s';`
	//const qspl = `SELECT arid,sales	FROM public.ar;`
	db := salaryM.imr.GetSQLDB()
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
		if err := rows.Scan(&saler.Sid, &saler.SName, &saler.Branch, &saler.Email); err != nil {
			fmt.Println("err Scan " + err.Error())
		}

		saList = append(saList, &saler)
	}

	out, err := json.Marshal(saList)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(out))

	return saList, nil

}
