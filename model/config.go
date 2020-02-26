package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/94peter/toad/resource/db"
)

type ConfigParameter struct {
	ID   string    `json:"id"`
	Date time.Time `json:"date"`
	//IT     float64   `json:"IT"`
	MMW    int     `json:"MMW"` //最低基本薪資
	NHI    float64 `json:"NHI"`
	LI     float64 `json:"LI"`
	NHI2nd float64 `json:"NHI2nd"`
	//AnnualRatio float64 `json:"annualRatio"`
}

type AccountItem struct {
	ItemName string `json:"itemName"`
	Valid    bool   `json:"-"`
}

type ConfigBranch struct {
	Branch        string  `json:"branch"`
	Rent          int     `json:"rent"`
	AgentSign     int     `json:"agentSign"`
	CommercialFee float64 `json:"commercialFee"`
	Manager       string  `json:"manager"`
	AnnualRatio   float64 `json:"annualRatio"`
	Sid           string  `json:"sid"`
}

type NullString struct {
	Value string
	Valid bool // Valid is true if Time is not NULL
}
type NullInt struct {
	Value int64
	Valid bool // Valid is true if Time is not NULL
}
type NullFloat struct {
	Value float64
	Valid bool // Valid is true if Time is not NULL
}
type ConfigSaler struct {
	Sid string `json:"sid"`
	//Csid     string    `json:"csid"`
	SName    string    `json:"name"`
	ZeroDate time.Time `json:"zeroDate"`
	//ValidDate time.Time `json:"validDate"`
	Title  string `json:"title"`
	Salary int    `json:"salary"`
	//Pay       int       `json:"pay"`
	Percent float64 `json:"percent"`
	//FPercent       float64   `json:"fPercent"`
	Branch         string `json:"branch"`
	PayrollBracket int    `json:"payrollBracket"` //投保金額
	Enrollment     int    `json:"enrollment"`     //加保(眷屬人數)
	Association    int    `json:"association"`    //公會
	Address        string `json:"address"`
	Birth          string `json:"birth"`
	IdentityNum    string `json:"identityNum"`
	BankAccount    string `json:"bankAccount"`
	Email          string `json:"email"`
	Phone          string `json:"phone"`
	Remark         string `json:"remark"`
	// excel used
	Tamount int    `json:"-"`
	CurDate string `json:"-"`
}

type ConfigSalary struct {
	Sid            string  `json:"sid"`
	SName          string  `json:"name"`
	ZeroDate       string  `json:"zeroDate"`
	Title          string  `json:"title"`
	Salary         int     `json:"salary"`
	Percent        float64 `json:"percent"`
	Branch         string  `json:"branch"`
	PayrollBracket int     `json:"payrollBracket"` //投保金額
	Enrollment     int     `json:"enrollment"`     //加保(眷屬人數)
	Association    int     `json:"association"`    //公會
	Remark         string  `json:"remark"`
}

var (
	configM *ConfigModel
)

type ConfigModel struct {
	imr                 interModelRes
	db                  db.InterSQLDB
	ConfigBranchList    []*ConfigBranch
	ConfigParameterList []*ConfigParameter
	ConfigSalerList     []*ConfigSaler
	ConfigSalaryList    []*ConfigSalary
	AccountItemList     []*AccountItem
}

//refer https://stackoverflow.com/questions/24564619/nullable-time-time-in-golang
func (ns *NullString) Scan(value interface{}) error {
	ns.Value, ns.Valid = value.(string)
	return nil
	//just keep the example
	// if nt.Valid {
	// 	// use nt.Time

	// } else {
	// 	// NULL value
	// }
}
func (ns *NullInt) Scan(value interface{}) error {
	ns.Value, ns.Valid = value.(int64)
	return nil
	//just keep the example
	// if nt.Valid {
	// 	// use nt.Time

	// } else {
	// 	// NULL value
	// }
}
func (ns *NullFloat) Scan(value interface{}) error {
	ns.Value, ns.Valid = value.(float64)
	return nil
	//just keep the example
	// if nt.Valid {
	// 	// use nt.Time

	// } else {
	// 	// NULL value
	// }
}

func GetConfigModel(imr interModelRes) *ConfigModel {
	if configM != nil {
		return configM
	}

	configM = &ConfigModel{
		imr: imr,
	}
	return configM
}

func (configM *ConfigModel) GetConfigBranchData(today, end time.Time) []*ConfigBranch {

	const qspl = `SELECT branch, rent, AgentSign, CommercialFee , Manager , Sid , annualratio FROM public.ConfigBranch;`
	//const qspl = `SELECT arid,sales	FROM public.ar;`
	db := configM.imr.GetSQLDB()
	rows, err := db.SQLCommand(fmt.Sprintf(qspl))
	if err != nil {
		return nil
	}
	var cbDataList []*ConfigBranch

	for rows.Next() {
		var cb ConfigBranch
		var manager, sid NullString
		// if err := rows.Scan(&r.ARid, &s); err != nil {
		// 	fmt.Println("err Scan " + err.Error())
		// }
		if err := rows.Scan(&cb.Branch, &cb.Rent, &cb.AgentSign, &cb.CommercialFee, &manager, &sid, &cb.AnnualRatio); err != nil {
			fmt.Println("err Scan " + err.Error())
		}
		cb.Manager = manager.Value
		cb.Sid = sid.Value
		cbDataList = append(cbDataList, &cb)
	}

	// out, err := json.Marshal(arList)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(string(out))
	configM.ConfigBranchList = cbDataList
	return configM.ConfigBranchList
}

func (configM *ConfigModel) Json(config string) ([]byte, error) {
	switch config {
	case "ConfigBranch":
		return json.Marshal(configM.ConfigBranchList)
	case "ConfigParameter":
		// f := map[string]interface{}{}
		// for _, param := range configM.ConfigParameterList {
		// 	fmt.Println(param.Param)
		// 	f[param.Param] = param.Value
		// }
		// return json.Marshal(f)
		return json.Marshal(configM.ConfigParameterList)
	case "ConfigSaler":
		return json.Marshal(configM.ConfigSalerList)
	case "ConfigSalary":
		return json.Marshal(configM.ConfigSalaryList)
	case "AccountItem":
		return json.Marshal(configM.AccountItemList)
	default:
		fmt.Println("unknown config type")
		break
	}
	return json.Marshal(amorM.amortizationList)
}

func (configM *ConfigModel) CreateConfigBranch(data []string) (err error) {

	sqlStr := "INSERT INTO public.ConfigBranch ( branch ) VALUES "
	for _, row := range data {
		sqlStr += "('"
		sqlStr += row
		sqlStr += "'), "
	}
	//trim the last ,
	sqlStr = sqlStr[0 : len(sqlStr)-2]
	sqlStr += "ON CONFLICT (branch) DO Nothing ;"
	fmt.Println(sqlStr)
	interdb := configM.imr.GetSQLDB()
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}

	res, err := sqldb.Exec(sqlStr)
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
		return errors.New("Invalid operation, CreateConfigBranch")
	}

	return nil
}

func (configM *ConfigModel) CreateConfigBranchWithManager(cb *ConfigBranch) (err error) {

	const sql = `INSERT INTO public.ConfigBranch
	( branch, rent, agentsign, CommercialFee, manager, sid)
	WITH  vals  AS (VALUES ( $1, $2::integer, $3::integer, $4::integer, $5, $6 ))
	SELECT v.* FROM vals as v
	WHERE EXISTS (SELECT sid FROM public.configsaler where sid = $6 and branch = $1);
	;`

	interdb := configM.imr.GetSQLDB()
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}

	res, err := sqldb.Exec(sql, cb.Branch, cb.Rent, cb.AgentSign, cb.CommercialFee, cb.Manager, cb.Sid)
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
		return errors.New("Invalid operation, CreateConfigBranch")
	}

	return nil
}

func (configM *ConfigModel) UpdateConfigBranch(Branch string, cb *ConfigBranch) (err error) {

	const sql = `UPDATE public.configbranch CB
					SET rent=$2, agentsign=$3 ,CommercialFee=$4, manager = $5 , Sid = $6 , annualratio=$7
					FROM (
					SELECT sid FROM public.configsaler where sid = $6 and branch = $1
					) CS Where CB.branch = $1;`

	interdb := configM.imr.GetSQLDB()
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}
	fmt.Println(cb.Sid)
	fmt.Println(Branch)
	res, err := sqldb.Exec(sql, Branch, cb.Rent, cb.AgentSign, cb.CommercialFee, cb.Manager, cb.Sid, cb.AnnualRatio)
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
		return errors.New("Invalid operation,UpdateConfigBranch")
	}

	return nil
}

func (configM *ConfigModel) DeleteConfigBranch(Branch string) (err error) {

	const sql = `DELETE FROM public.configbranch WHERE branch = $1;`

	interdb := configM.imr.GetSQLDB()
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}

	res, err := sqldb.Exec(sql, Branch)
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
		return errors.New("Invalid operation. ")
	}

	return nil
}

func (configM *ConfigModel) GetConfigParameterData(today, end time.Time) []*ConfigParameter {

	const qspl = `SELECT id, date, nhi, LI, nhi2nd, MMW  FROM public.ConfigParameter;`
	//const qspl = `SELECT arid,sales	FROM public.ar;`
	db := configM.imr.GetSQLDB()
	rows, err := db.SQLCommand(fmt.Sprintf(qspl))
	if err != nil {
		return nil
	}
	var cpDataList []*ConfigParameter

	for rows.Next() {
		var cp ConfigParameter

		// if err := rows.Scan(&r.ARid, &s); err != nil {
		// 	fmt.Println("err Scan " + err.Error())
		// }
		if err := rows.Scan(&cp.ID, &cp.Date, &cp.NHI, &cp.LI, &cp.NHI2nd, &cp.MMW); err != nil {
			fmt.Println("err Scan " + err.Error())
		}

		cpDataList = append(cpDataList, &cp)
	}

	// out, err := json.Marshal(arList)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(string(out))
	configM.ConfigParameterList = cpDataList
	return configM.ConfigParameterList
}

func (configM *ConfigModel) UpdateConfigParameter(cp *ConfigParameter, ID string) (err error) {

	const sql = `UPDATE public.configparameter
				SET date=$1, nhi=$2, LI=$3, nhi2nd=$4, MMW=$5 
				WHERE id=$6;`

	interdb := configM.imr.GetSQLDB()
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}

	res, err := sqldb.Exec(sql, cp.Date, cp.NHI, cp.LI, cp.NHI2nd, cp.MMW, ID)
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
		return errors.New("Invalid operation, update ConfigParameter")

	}

	return nil
}

func (configM *ConfigModel) DeleteConfigParameter(ID string) (err error) {

	const sql = `DELETE FROM public.configparameter	WHERE id=$1;`

	interdb := configM.imr.GetSQLDB()
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
		return errors.New("Invalid operation, DeleteConfigParameter")
	}

	return nil
}

func (configM *ConfigModel) CreateConfigParameter(cp *ConfigParameter) (err error) {

	const sql = `INSERT INTO public.ConfigParameter
	(id, date, nhi, LI, nhi2nd, MMW )
	VALUES ($1, $2, $3, $4, $5, $6);`

	interdb := configM.imr.GetSQLDB()
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}
	fakeId := time.Now().Unix()
	res, err := sqldb.Exec(sql, fakeId, cp.Date, cp.NHI, cp.LI, cp.NHI2nd, cp.MMW)
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
		return errors.New("Invalid operation, CreateConfigParameter")
	}

	return nil
}

func (configM *ConfigModel) GetConfigSalerData(branch string) []*ConfigSaler {

	const qspl = `SELECT  sid, sname, branch, zerodate,  title, percent, 
				  salary,  payrollbracket, enrollment, association, address, birth, identityNum , bankAccount , email, phone , remark
				  FROM public.ConfigSaler where branch like '%s';`
	//const qspl = `SELECT arid,sales	FROM public.ar;`
	db := configM.imr.GetSQLDB()
	fmt.Println(fmt.Sprintf(qspl, branch))
	rows, err := db.SQLCommand(fmt.Sprintf(qspl, branch))
	if err != nil {
		fmt.Println(err)
		return nil
	}
	var cbDataList []*ConfigSaler

	for rows.Next() {
		var cs ConfigSaler

		// if err := rows.Scan(&r.ARid, &s); err != nil {
		// 	fmt.Println("err Scan " + err.Error())
		// }
		if err := rows.Scan(&cs.Sid, &cs.SName, &cs.Branch, &cs.ZeroDate, &cs.Title, &cs.Percent,
			&cs.Salary, &cs.PayrollBracket, &cs.Enrollment, &cs.Association, &cs.Address, &cs.Birth, &cs.IdentityNum, &cs.BankAccount, &cs.Email, &cs.Phone, &cs.Remark); err != nil {
			fmt.Println("err Scan " + err.Error())
		}

		cbDataList = append(cbDataList, &cs)
	}

	// out, err := json.Marshal(arList)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(string(out))
	configM.ConfigSalerList = cbDataList
	return configM.ConfigSalerList
}

func (configM *ConfigModel) CheckConfigSaler(identitynum, zeroDate string) (r string, err error) {

	const sql = `SELECT zerodate, branch FROM public.configsaler where identitynum = '%s' group by branch, zerodate;;`
	//and zerodate = '%s

	interdb := configM.imr.GetSQLDB()

	rows, err := interdb.SQLCommand(fmt.Sprintf(sql, identitynum))
	if err != nil {
		return "", err
	}
	var sList []*ConfigSaler
	r = ""
	for rows.Next() {
		var saler ConfigSaler

		// if err := rows.Scan(&r.ARid, &s); err != nil {
		// 	fmt.Println("err Scan " + err.Error())
		// }
		if err := rows.Scan(&saler.ZeroDate, &saler.Branch); err != nil {
			fmt.Println("err Scan " + err.Error())
		}

		if saler.ZeroDate.String()[:10] == zeroDate {
			//fmt.Println(saler.ZeroDate.String())
			return "[Danger]:重複資料存在: 起始日、身份證字號", nil
		}
		r += saler.Branch + " "
		sList = append(sList, &saler)
	}

	out, err := json.Marshal(sList)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(out))

	if r == "" {
		return "OK", nil
	}

	return "[Info]:" + r + "等店 存在重複業務", nil
}

func (configM *ConfigModel) CreateConfigSaler(cs *ConfigSaler) (err error) {

	const sql = `INSERT INTO public.configsaler(
		sid, sname, branch, zerodate,  title, percent,  salary,
		 payrollbracket, enrollment, association, address, birth, identityNum, bankAccount, phone , email, remark)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17);`

	interdb := configM.imr.GetSQLDB()
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}

	res, err := sqldb.Exec(sql, cs.Sid, cs.SName, cs.Branch, cs.ZeroDate, cs.Title, cs.Percent, cs.Salary,
		cs.PayrollBracket, cs.Enrollment, cs.Association, cs.Address, cs.Birth, cs.IdentityNum, cs.BankAccount, cs.Email, cs.Phone, cs.Remark)
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
		return errors.New("Invalid operation, ConfigSaler")
	}
	//新增預設紀錄

	configM.CreateConfigSalary(cs.GetConfigSalary())
	return nil
}

func (configM *ConfigModel) DeleteConfigSaler(sid string) (err error) {

	const sql = `DELETE FROM public.configsaler	WHERE sid=$1 ;`

	interdb := configM.imr.GetSQLDB()
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}

	res, err := sqldb.Exec(sql, sid)
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
		return errors.New("Invalid operation, maybe not found the saler")
	}
	//成功的話，同時刪除薪資資訊
	const qsql = `DELETE FROM public.configsalary WHERE sid=$1;`
	res, err = sqldb.Exec(qsql, sid)
	id, err = res.RowsAffected()
	fmt.Println(id)
	return nil
}

func (configM *ConfigModel) UpdateConfigSaler(cs *ConfigSaler, Sid string) (err error) {

	const sql = `UPDATE public.configsaler
	SET zerodate=$2,  title=$3, percent=$4, salary=$5,
	payrollbracket=$6, enrollment=$7, association=$8, address=$9, birth=$10, identitynum=$11, bankaccount= $12 , email = $13  ,branch = $14, remark = $15
	WHERE sid=$1`

	interdb := configM.imr.GetSQLDB()
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}

	res, err := sqldb.Exec(sql, Sid, cs.ZeroDate, cs.Title, cs.Percent, cs.Salary,
		cs.PayrollBracket, cs.Enrollment, cs.Association, cs.Address, cs.Birth, cs.IdentityNum, cs.BankAccount, cs.Email, cs.Branch, cs.Remark)
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
		return errors.New("Invalid operation, maybe not found the saler")
	}

	return nil
}

func (configM *ConfigModel) GetConfigSalaryData(sid string) (err error) {

	// const sql = `SELECT A.*
	// 				FROM public.configsalary A
	// 				Inner Join (
	// 					select sid, max(zerodate) zerodate from public.configsalary cs
	// 					where now() > to_timestamp(zerodate,'YYYY-MM')
	// 					group by sid
	// 				) B on A.sid=B.sid and A.zeroDate = B.zeroDate
	// 			  where branch sid = '%s';`
	const sql = `SELECT sid, sname, branch, zerodate, title, percent, salary,payrollbracket, enrollment, association,remark
	  FROM public.configsalary where  sid like '%s' order by zeroDate desc;`
	//const qspl = `SELECT arid,sales	FROM public.ar;`
	db := configM.imr.GetSQLDB()
	fmt.Println(fmt.Sprintf(sql, sid))
	rows, err := db.SQLCommand(fmt.Sprintf(sql, sid))
	if err != nil {
		fmt.Println(err)
		return err
	}
	var cbDataList []*ConfigSalary

	for rows.Next() {
		var cs ConfigSalary

		// if err := rows.Scan(&r.ARid, &s); err != nil {
		// 	fmt.Println("err Scan " + err.Error())
		// }
		if err := rows.Scan(&cs.Sid, &cs.SName, &cs.Branch, &cs.ZeroDate, &cs.Title, &cs.Percent,
			&cs.Salary, &cs.PayrollBracket, &cs.Enrollment, &cs.Association, &cs.Remark); err != nil {
			fmt.Println("err Scan " + err.Error())
		}

		cbDataList = append(cbDataList, &cs)
	}

	// out, err := json.Marshal(arList)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(string(out))
	configM.ConfigSalaryList = cbDataList
	return nil
}

func (configM *ConfigModel) CreateConfigSalary(cs *ConfigSalary) (err error) {

	const sql = `INSERT INTO public.configsalary(
		sid, sname, branch, zerodate, title, percent, salary, payrollbracket, enrollment, association, remark)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (sid,zerodate) DO UPDATE SET sname = excluded.sname, title = excluded.title, branch = excluded.branch,
		percent = excluded.percent, salary = excluded.salary, payrollbracket = excluded.payrollbracket,
		enrollment = excluded.enrollment , association = excluded.association , remark = excluded.remark ;`

	interdb := configM.imr.GetSQLDB()
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}

	res, err := sqldb.Exec(sql, cs.Sid, cs.SName, cs.Branch, cs.ZeroDate, cs.Title, cs.Percent, cs.Salary,
		cs.PayrollBracket, cs.Enrollment, cs.Association, cs.Remark)
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
		return errors.New("Invalid operation, CreateConfigSalary")
	}
	configM.WorkValidDate()
	return nil
}
func (configM *ConfigModel) DeleteConfigSalary(sid, zerodate string) (err error) {

	const sql = `DELETE FROM public.configsalary WHERE sid=$1 and zerodate = $2;`

	interdb := configM.imr.GetSQLDB()
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}

	res, err := sqldb.Exec(sql, sid, zerodate)
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
		return errors.New("Invalid operation, maybe not found the salary of saler ")
	}
	configM.WorkValidDate()
	return nil
}

/**
TODO::
*月初啟動，所以時間用now()即可。
*validdate != '0001-01-01' and validdate < now() 條件
*因要寫入log，故不採用update from select 語句，採用使用for loop完成
**/
func (configM *ConfigModel) WorkValidDate() (err error) {

	const sql = `UPDATE public.configsaler cs
	SET sname=subquery.sname, branch=subquery.branch, zerodate=to_timestamp(subquery.zerodate,'YYYY-MM-DD'), title=subquery.title, percent=subquery.percent, 
		salary=subquery.salary, payrollbracket=subquery.payrollbracket, enrollment=subquery.enrollment, association=subquery.association, 
		remark=subquery.remark 	
	FROM(
		SELECT A.sid, A.zerodate, A.sname, A.branch, A.title, A.percent, A.salary, A.payrollbracket, A.enrollment, A.association, A.remark
		FROM public.configsalary A 
		Inner Join ( 
			select sid, max(zerodate) zerodate from public.configsalary cs 
			where now() >= to_timestamp(zerodate,'YYYY-MM') + '1 month'::interval
			group by sid 
		) B on A.sid=B.sid and A.zeroDate = B.zeroDate
	) AS subquery
	WHERE subquery.sid = cs.sid;`
	//const qspl = `SELECT arid,sales	FROM public.ar;`
	db := configM.imr.GetSQLDB()
	sqldb, err := db.ConnectSQLDB()
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
	fmt.Println("WorkValidDate:", id)

	return nil
}

func (configM *ConfigModel) GetAccountItemData(today, end time.Time) []*AccountItem {

	const qspl = `SELECT AccountItemName, Valid FROM public.AccountItem;`
	//const qspl = `SELECT arid,sales	FROM public.ar;`
	db := configM.imr.GetSQLDB()
	rows, err := db.SQLCommand(fmt.Sprintf(qspl))
	if err != nil {
		return nil
	}
	var AccountItemList []*AccountItem

	for rows.Next() {
		var aitem AccountItem

		// if err := rows.Scan(&r.ARid, &s); err != nil {
		// 	fmt.Println("err Scan " + err.Error())
		// }
		if err := rows.Scan(&aitem.ItemName, &aitem.Valid); err != nil {
			fmt.Println("err Scan " + err.Error())
		}

		AccountItemList = append(AccountItemList, &aitem)
	}

	// out, err := json.Marshal(arList)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(string(out))
	configM.AccountItemList = AccountItemList
	return configM.AccountItemList
}

func (configM *ConfigModel) CreateAccountItem(aitem *AccountItem) (err error) {

	const sql = `INSERT INTO public.AccountItem
	(AccountItemName)
	VALUES ($1)
	;`

	interdb := configM.imr.GetSQLDB()
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}

	res, err := sqldb.Exec(sql, aitem.ItemName)
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
		return errors.New("Invalid operation, Create AccountItem")
	}

	return nil
}

func (configM *ConfigModel) UpdateAccountItem(oldItemName string, aitem *AccountItem) (err error) {

	const sql = `UPDATE public.AccountItem
				SET AccountItemName=$2
				WHERE AccountItemName=$1;
				;`

	interdb := configM.imr.GetSQLDB()
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}

	res, err := sqldb.Exec(sql, oldItemName, aitem.ItemName)
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
		return errors.New("Invalid operation, Update AccountItem")
	}

	return nil
}
func (configM *ConfigModel) DeleteAccountItem(ItemName string) (err error) {

	const sql = `DELETE FROM public.accountitem WHERE AccountItemName=$1;`

	interdb := configM.imr.GetSQLDB()
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}

	res, err := sqldb.Exec(sql, ItemName)
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
		return errors.New("Invalid operation, Delete AccountItem")
	}

	return nil
}

func (cs *ConfigSaler) GetConfigSalary() *ConfigSalary {

	return &ConfigSalary{
		Sid:            cs.Sid,
		SName:          cs.SName,
		Salary:         cs.Salary,
		Percent:        cs.Percent,
		Title:          cs.Title,
		ZeroDate:       cs.ZeroDate.Format("2006-01-02"),
		Branch:         cs.Branch,
		PayrollBracket: cs.PayrollBracket,
		Enrollment:     cs.Enrollment,
		Association:    cs.Association,
		Remark:         cs.Remark,
	}
}
