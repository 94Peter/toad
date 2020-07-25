package model

import (
	"encoding/json"
	"fmt"
	"time"

	"toad/resource/db"
)

//`json:"id"` 回傳重新命名
type Info struct {
	Receivable  int `json:"receivable"`  //累積應收
	Performance int `json:"performance"` //本月業績
}

type IncomeStatement struct {
	Branch string `json:"branch"`
	Income Income `json:"income"`

	Expense Expense `json:"expense"`

	BusinessIncomeTax int `json:"businessIncomeTax"`
	Aftertax          int `json:"afterTax"`
	Pretax            int `json:"pretax"`
	Lastloss          int `json:"lastLoss"`
	ManagerBonus      int `json:"managerBonus"`
	EarnAdjust        int `json:"earnAdjust"`
}

type IndexModel struct {
	imr             interModelRes
	db              db.InterSQLDB
	info            *Info
	incomeStatement *IncomeStatement
}

var (
	indexM *IndexModel
)

func GetIndexModel(imr interModelRes) *IndexModel {
	if indexM != nil {
		return indexM
	}

	indexM = &IndexModel{
		imr: imr,
	}
	return indexM
}

func (indexM *IndexModel) Json(mtype string) ([]byte, error) {
	switch mtype {
	case "info":
		return json.Marshal(indexM.info)
	case "incomeStatement":
		return json.Marshal(indexM.incomeStatement)
	default:
		fmt.Println("unknown config type")
		break
	}
	return nil, nil
}

func (indexM *IndexModel) GetInfoData(date time.Time) *Info {

	// const sql = `SELECT
	// 			ar.arid, ar.date, ar.cno, ar.casename, ar.type, ar.name, ar.amount,
	// 				COALESCE((SELECT SUM(d.fee) FROM public.deduct d WHERE ar.arid = d.arid),0) AS SUM_Fee,
	// 				COALESCE((SELECT SUM(r.amount) FROM public.receipt r WHERE ar.arid = r.arid),0) AS SUM_RA,
	// 				ar.sales
	// 			where ar.arid like '%%s%'  OR ar.cno like '%%s%' OR ar.casename like '%%s%' OR ar.type like '%%s%' OR ar.name like '%%s%'
	// 			FROM public.ar ar
	// 			group by ar.arid;`
	const sql = `SELECT  SUM(ar.amount) amount,      
	SUM(COALESCE((SELECT SUM(r.amount) FROM public.receipt r WHERE ar.arid = r.arid),0)) AS SUM_RA ,
	COALESCE((SELECT  sum(amount) FROM public.receipt where extract(epoch from date) >= '%d' and extract(epoch from date - '1 month'::interval)  < '%d'  ),0) AS Performance
	FROM public.ar ar `
	/*
	*balance equal ar.amount - COALESCE((SELECT SUM(r.amount) FROM public.receipt r WHERE ar.arid = r.arid),0) AS SUM_RA
	*but I do with r.Balance = r.Amount - r.RA
	 */

	//const qspl = `SELECT arid,sales	FROM public.ar;`
	db := indexM.imr.GetSQLDB()
	//fmt.Println(sql)
	//rows, err := db.SQLCommand(fmt.Sprintf(sql))
	//t := time.Now()
	//curDate := fmt.Sprintf("%d-%02d-01", t.Year(), t.Month())
	// formatted := fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02d",
	// 	t.Year(), t.Month(), t.Day(),
	// 	t.Hour(), t.Minute(), t.Second())
	//curDate := fmt.Sprintf("%d-%02d-01", t.Year(), t.Month())

	b, _ := time.Parse(time.RFC3339, "2019-12-31T16:00:00Z")

	//curDate := fmt.Sprintf("2020-01-01")
	//fmt.Println(curDate)
	rows, err := db.SQLCommand(fmt.Sprintf(sql, b.Unix(), b.Unix()))
	//rows, err := db.SQLCommand(fmt.Sprintf(sql, curDate, curDate))
	if err != nil {
		fmt.Println(err)
		return nil
	}
	var data *Info

	for rows.Next() {
		var info Info
		var Amount, RA = 0, 0
		//var col_sales string
		// if err := rows.Scan(&r.ARid, &s); err != nil {
		// 	fmt.Println("err Scan " + err.Error())
		// }
		if err := rows.Scan(&Amount, &RA, &info.Performance); err != nil {
			fmt.Println("err Scan " + err.Error())
		}

		info.Receivable = Amount - RA

		//先顛倒，前端沒弄好
		//info.Receivable = info.Performance
		//info.Performance = Amount - RA

		data = &info
	}

	// out, err := json.Marshal(data)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(string(out))

	indexM.info = data
	return indexM.info

}

func (indexM *IndexModel) GetIncomeStatement(branch string, date time.Time) (*IncomeStatement, error) {

	//本來使用此sql，但有可能branchsalary為空
	const sql = `WITH  vals  AS (VALUES ( 'none' ) )
	SELECT subtable.branch , sum(subtable.Pbonus), sum(subtable.LBonus) , sum(subtable.salary), sum(subtable.prepay), sum(subtable.pocket) , 
		sum(subtable.thisMonthAmor) Amor, sum(subtable.sr) SR, sum(subtable.annualbonus)::int annualbonus, sum(subtable.salesamounts)::int , sum(subtable.businesstax)::int , sum(subtable.agentsign) , sum(subtable.rent),
		sum(subtable.commercialfee) commercialfee, sum(subtable.PreTax)::int PreTax, sum( CASE WHEN subtable.PreTax > 0 then subtable.PreTax * 0.2 else 0 end )::int BusinessIncomeTax, 
		sum( subtable.PreTax - ( CASE WHEN subtable.PreTax > 0 then subtable.PreTax * 0.2 else 0 end ) )::int AfterTax , 
		sum(subtable.pretaxTotal)  lastloss ,  
		sum( CASE WHEN (subtable.PreTax - ( CASE WHEN subtable.PreTax > 0 then subtable.PreTax * 0.2 else 0 end )) + (subtable.pretaxTotal + subtable.PreTax ) + 0 > 0 then 
					(subtable.PreTax - ( CASE WHEN subtable.PreTax > 0 then subtable.PreTax * 0.2 else 0 end )) + (subtable.pretaxTotal + subtable.PreTax ) + 0
		  else 0 end
		)::int managerbonus 
		FROM vals as v
		cross join (
		SELECT incomeexpense.branch , COALESCE(incomeexpense.pretaxTotal ,0) pretaxTotal , BS.Bsid,BonusTable.PBonus , BonusTable.LBonus , BonusTable.Salary , COALESCE(prepayTable.prepay,0) prepay , COALESCE(pocketTable.pocket,0) pocket , COALESCE(amorTable.thisMonthAmor,0) thisMonthAmor,
		COALESCE(commissionTable.SR,0) SR, COALESCE(commissionTable.SR / 1.05 ,0) salesamounts , COALESCE(commissionTable.SR - commissionTable.SR / 1.05 ,0) businesstax, configTable.agentsign, configTable.rent, configTable.commercialfee, 
		( COALESCE(commissionTable.SR,0)/1.05  - COALESCE(amorTable.thisMonthAmor,0) - configTable.agentsign - configTable.rent - COALESCE(pocketTable.pocket,0) - COALESCE(prepayTable.prepay,0) - BonusTable.PBonus - 
		BonusTable.Salary - BonusTable.LBonus - COALESCE(commissionTable.SR,0) * 0.05 - configTable.commercialfee - 0  ) PreTax ,
		COALESCE(commissionTable.SR * configTable.annualratio / 100 ,0) Annualbonus , configTable.annualratio
		FROM public.branchsalary  BS
		inner join (
		  SELECT sum(BonusTable.pbonus) PBonus , sum(BonusTable.lbonus) LBonus, sum(BonusTable.Salary) Salary, bsid  FROM public.SalerSalary BonusTable group by bsid
		) BonusTable on BonusTable.bsid = BS.bsid
		left join (
			SELECT sum(cost) prepay , branch FROM public.prepay PP 
			inner join public.BranchPrePay BPP on PP.ppid = BPP.ppid 	
			where to_char(date ,'YYYY-MM') = '%s'
			group by branch
		) prepayTable on prepayTable.branch = BS.branch
		left join(
			SELECT sum(fee) pocket , branch FROM public.Pocket 		
			where circleid = '%s'
			group by branch
		) pocketTable on pocketTable.branch = BS.branch
		left join(
			SELECT to_char(amor.date,'yyyy-MM') , branch , sum(cost) thismonthamor FROM public.amortization amor
			inner join (
				SELECT amorid, date, cost FROM public.amormap
			) amormap on amormap.amorid = amor.amorid
			where isover = 0 and to_char(amor.date,'yyyy-MM') = '%s'
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
		where date = '%s'
		) subtable
	where branch = '%s'
	group by subtable.branch
	`
	const incomeSql = ` WITH  vals  AS (VALUES ( 'none' ) )
	SELECT SUM(SR) SR, SUM(bonus) bonus,  SUM(salary) salary
	FROM vals as v
	cross join (
	select cb.annualratio, cs.sid, cs.branch,cs.salary, c.* from configsaler cs
		   left join (
		    select * from commission c inner join receipt r on r.rid = c.rid
		    where c.status != 'remove' and extract(epoch from r.date) >= '%d' and extract(epoch from r.date - '1 month'::interval) < '%d'
		   ) c on c.sid = cs.sid
		   inner join public.configbranch cb on cb.branch = cs.branch
		   where cs.branch='%s'
	) subtable;`

	const amorSql = `select COALESCE(sum(cost),0) from public.amormap amp
	inner join public.amortization am on am.amorid = amp.amorid
	where amp.date = '%s' and am.branch = '%s'`

	const configBranchSql = `select rent, agentsign, commercialfee , annualratio from public.configbranch where branch='%s';`

	const pocketSql = `SELECT COALESCE(sum(fee),0) from public.pocket where circleid = '%s' and branch = '%s';`

	const prepaySql = `select sum(cost) from prepay pp
			inner join BranchPrePay bpp on bpp.ppid = pp.ppid
			where pp.date < ('%s'::date + '1 month'::interval) and pp.date >= ('%s'::date) and  bpp.branch = '%s' `

	const lastlossSql = `Select lastloss
			FROM public.incomeexpense IE
			inner join public.BranchSalary BS on  IE.bsid = BS.bsid
			where date = '%s' and branch = '%s';`

	lt := time.Now().AddDate(0, -1, 0)
	lastMonthDate := fmt.Sprintf("%d-%02d-01", lt.Year(), lt.Month())
	fmt.Println("lastMonthDate:", lastMonthDate)
	//curDate := fmt.Sprintf("%d-%02d-01", t.Year(), t.Month())
	//layout := "2006-01-02"
	//curDate := fmt.Sprintf("2020-01")

	mdate, _ := time.Parse(time.RFC3339, "2020-01-01T00:00:00+08:00")

	curDate := fmt.Sprintf("%d-%02d", mdate.Year(), mdate.Month())
	fmt.Println(curDate)
	//t, _ := time.Parse(layout, curDate+"-01")

	mdb := indexM.imr.GetSQLDB()
	db, err := mdb.ConnectSQLDB()
	defer db.Close()

	rows, err := db.Query(fmt.Sprintf(incomeSql, mdate.Unix(), mdate.Unix(), branch))
	if err != nil {
		fmt.Println(err)
		return nil, nil
	}
	//fmt.Println(fmt.Sprintf(incomeSql, t.Unix(), t.Unix(), branch))
	// 收入/薪資支出 年終提播
	var SR, Salary, Bonus, Salesamounts, Businesstax int
	for rows.Next() {
		if err := rows.Scan(&SR, &Bonus, &Salary); err != nil {
			fmt.Println("err Scan " + err.Error())
			return nil, err
		}
	}
	Salesamounts = int(round(float64(SR)/1.05, 1))
	Businesstax = SR - Salesamounts

	rows, err = db.Query(fmt.Sprintf(amorSql, curDate, branch))
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	var Amor int
	for rows.Next() {
		if err := rows.Scan(&Amor); err != nil {
			fmt.Println("err Scan " + err.Error())
			return nil, err
		}
	}

	rows, err = db.Query(fmt.Sprintf(pocketSql, curDate, branch))
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	var Pocket int
	for rows.Next() {
		if err := rows.Scan(&Pocket); err != nil {
			fmt.Println("err Scan " + err.Error())
			return nil, err
		}
	}

	rows, err = db.Query(fmt.Sprintf(prepaySql, curDate+"-01", curDate+"-01", branch))
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	var Prepay int
	for rows.Next() {
		if err := rows.Scan(&Prepay); err != nil {
			fmt.Println("err Scan " + err.Error())
			return nil, err
		}
	}

	rows, err = db.Query(fmt.Sprintf(configBranchSql, branch))
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	var Rent, Agentsign int
	var Commmercialfee, Annualratio float64
	for rows.Next() {
		if err := rows.Scan(&Rent, &Agentsign, &Commmercialfee, &Annualratio); err != nil {
			fmt.Println("err Scan " + err.Error())
			return nil, err
		}
	}

	rows, err = db.Query(fmt.Sprintf(lastlossSql, branch))
	if err != nil {
		return nil, err
	}
	var lastloss int
	for rows.Next() {
		if err := rows.Scan(&lastloss); err != nil {
			fmt.Println("err Scan " + err.Error())
			return nil, err
		}
	}

	var Pretax, Aftertax, BusinessIncomeTax, ManagerBonus int
	Pretax = Salesamounts - (Amor + Agentsign + Rent + Pocket + Salary + Prepay + Bonus + int(round(Commmercialfee*float64(Salary+Bonus)/100, 1)) + int(round(Annualratio*float64(SR)/100, 1)))
	if Pretax > 0 {
		BusinessIncomeTax = int(round(float64(Pretax)*0.8, 1))
	} else {
		BusinessIncomeTax = 0
	}
	Aftertax = Pretax - BusinessIncomeTax
	ManagerBonus = int(round(float64(Aftertax+lastloss+0)*0.2, 1))
	if ManagerBonus < 0 {
		ManagerBonus = 0
	}
	income := Income{
		SR:           SR,
		Salesamounts: Salesamounts,
		Businesstax:  Businesstax,
	}
	expense := Expense{
		Amorcost:      Amor,
		Agentsign:     Agentsign,
		Rent:          Rent,
		Pocket:        Pocket,
		Salary:        Salary,
		Prepay:        Prepay,
		Pbonus:        Bonus,
		Annualbonus:   int(round(Annualratio*float64(SR)/100, 1)),
		Commercialfee: int(round(Commmercialfee*float64(Salary+Bonus)/100, 1)),
	}
	data := &IncomeStatement{
		Aftertax:          Aftertax,
		Lastloss:          lastloss,
		Pretax:            Pretax,
		ManagerBonus:      ManagerBonus,
		BusinessIncomeTax: BusinessIncomeTax,
		Income:            income,
		Expense:           expense,
	}
	data.Aftertax = Aftertax

	// fmt.Println("SR:", SR)
	// fmt.Println("Salesamounts:", Salesamounts)
	// fmt.Println("Businesstax:", Businesstax)
	// fmt.Println("Salary:", Salary)
	// fmt.Println("Bonus:", Bonus)
	// fmt.Println("Amor:", Amor)
	// fmt.Println("Pocket:", Pocket)
	// fmt.Println("Prepay:", Prepay)
	// fmt.Println("Rent:", Rent)
	// fmt.Println("Agentsign:", Agentsign)
	// fmt.Println("Commmercialfee:", int(round(Commmercialfee*float64(Salary+Bonus)/100, 1)))
	// fmt.Println("Annualratio:", int(round(Annualratio*float64(SR)/100, 1)))
	// fmt.Println("lastloss:", lastloss)
	// fmt.Println("Pretax:", Pretax)
	// fmt.Println("Aftertax:", Aftertax)
	// fmt.Println("ManagerBonus:", ManagerBonus)

	//indexM.incomeStatement = data
	return data, err
}
