package model

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/94peter/toad/resource/db"
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

func (indexM *IndexModel) GetInfoData() *Info {

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
	COALESCE((SELECT  sum(amount) FROM public.receipt where date between '%s' and ('%s'::date + '1 month'::interval)),0) AS Performance
	FROM public.ar ar `
	/*
	*balance equal ar.amount - COALESCE((SELECT SUM(r.amount) FROM public.receipt r WHERE ar.arid = r.arid),0) AS SUM_RA
	*but I do with r.Balance = r.Amount - r.RA
	 */

	//const qspl = `SELECT arid,sales	FROM public.ar;`
	db := indexM.imr.GetSQLDB()
	//fmt.Println(sql)
	//rows, err := db.SQLCommand(fmt.Sprintf(sql))
	t := time.Now()
	// formatted := fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02d",
	// 	t.Year(), t.Month(), t.Day(),
	// 	t.Hour(), t.Minute(), t.Second())
	curDate := fmt.Sprintf("%d-%02d-01", t.Year(), t.Month())
	fmt.Println(curDate)

	rows, err := db.SQLCommand(fmt.Sprintf(sql, curDate, curDate))
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
		// err := json.Unmarshal([]byte(col_sales), &r.Sales)
		// if err != nil {
		// 	fmt.Println(err)
		// }

		data = &info
	}

	// out, err := json.Marshal(arList)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(string(out))
	indexM.info = data
	return indexM.info

}

func (indexM *IndexModel) GetIncomeStatement(branch string) *IncomeStatement {

	//(subtable.pretaxTotal + subtable.PreTax )  lastloss ,   應該不包含這期虧損
	const sql = `WITH  vals  AS (VALUES ( 'none' ) )
	SELECT subtable.branch , sum(subtable.Pbonus), sum(subtable.LBonus) , sum(subtable.salary), sum(subtable.prepay), sum(subtable.pocket) , 
		sum(subtable.thisMonthAmor) , sum(subtable.sr), sum(subtable.annualbonus), sum(subtable.salesamounts) , sum(subtable.businesstax) , sum(subtable.agentsign) , sum(subtable.rent),
		sum(subtable.commercialfee), sum(subtable.PreTax) , sum( CASE WHEN subtable.PreTax > 0 then subtable.PreTax * 0.2 else 0 end ) BusinessIncomeTax, 
		sum( subtable.PreTax - ( CASE WHEN subtable.PreTax > 0 then subtable.PreTax * 0.2 else 0 end ) ) AfterTax , 
		sum(subtable.pretaxTotal)  lastloss ,  
		sum( CASE WHEN (subtable.PreTax - ( CASE WHEN subtable.PreTax > 0 then subtable.PreTax * 0.2 else 0 end )) + (subtable.pretaxTotal + subtable.PreTax ) + 0 > 0 then 
					(subtable.PreTax - ( CASE WHEN subtable.PreTax > 0 then subtable.PreTax * 0.2 else 0 end )) + (subtable.pretaxTotal + subtable.PreTax ) + 0
		  else 0 end
		) managerbonus 
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
			where to_char(date ,'YYYY-MM') = '2019-10'
			group by branch
		) prepayTable on prepayTable.branch = BS.branch
		left join(
			SELECT sum(fee) pocket , branch FROM public.Pocket 		
			where circleid = '2019-10'
			group by branch
		) pocketTable on pocketTable.branch = BS.branch
		left join(
			SELECT to_char(amor.date,'yyyy-MM') , branch , sum(cost) thismonthamor FROM public.amortization amor
			inner join (
				SELECT amorid, date, cost FROM public.amormap
			) amormap on amormap.amorid = amor.amorid
			where isover = 0 and to_char(amor.date,'yyyy-MM') = '2019-10'
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
		where date = '2019-10'
		) subtable
	where branch = '%s'
	group by subtable.branch
	`

	db := indexM.imr.GetSQLDB()
	rows, err := db.SQLCommand(fmt.Sprintf(sql, branch))
	if err != nil {
		return nil
	}

	var data *IncomeStatement

	for rows.Next() {
		var ie IncomeStatement

		if err := rows.Scan(&ie.Branch, &ie.Expense.Pbonus, &ie.Expense.LBonus, &ie.Expense.Salary, &ie.Expense.Prepay, &ie.Expense.Pocket,
			&ie.Expense.Amorcost, &ie.Income.SR, &ie.Expense.Annualbonus, &ie.Income.Salesamounts, &ie.Income.Businesstax, &ie.Expense.Agentsign, &ie.Expense.Rent,
			&ie.Expense.Commercialfee, &ie.Pretax, &ie.BusinessIncomeTax,
			&ie.Aftertax, &ie.Lastloss, &ie.ManagerBonus); err != nil {
			fmt.Println("err Scan " + err.Error())
			return nil
		}

		data = &ie
	}
	indexM.incomeStatement = data
	return indexM.incomeStatement
}
