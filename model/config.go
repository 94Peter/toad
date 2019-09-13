package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/94peter/toad/resource/db"
)

type ConfigParameter struct {
	Date   time.Time `json:"date"`
	IT     float64   `json:"IT"`
	NHI    float64   `json:"NHI"`
	LI     float64   `json:"LI"`
	NHI2nd float64   `json:"NHI2nd"`
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
}

type ConfigSaler struct {
	Sid            string    `json:"id"`
	SName          string    `json:"name"`
	ZeroDate       time.Time `json:"zeroDate"`
	ValidDate      time.Time `json:"validDate"`
	Title          string    `json:"title"`
	Salary         int       `json:"salary"`
	Pay            int       `json:"pay"`
	Percent        float64   `json:"percent"`
	FPercent       float64   `json:"fPercent"`
	Branch         string    `json:"branch"`
	PayrollBracket int       `json:"payrollBracket"` //投保金額
	Enrollment     int       `json:"enrollment"`     //加保(眷屬人數)
	Association    int       `json:"association"`    //公會
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
	AccountItemList     []*AccountItem
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

	const qspl = `SELECT branch, rent, AgentSign, CommercialFee FROM public.ConfigBranch;`
	//const qspl = `SELECT arid,sales	FROM public.ar;`
	db := configM.imr.GetSQLDB()
	rows, err := db.SQLCommand(fmt.Sprintf(qspl))
	if err != nil {
		return nil
	}
	var cbDataList []*ConfigBranch

	for rows.Next() {
		var cb ConfigBranch

		// if err := rows.Scan(&r.ARid, &s); err != nil {
		// 	fmt.Println("err Scan " + err.Error())
		// }
		if err := rows.Scan(&cb.Branch, &cb.Rent, &cb.AgentSign, &cb.CommercialFee); err != nil {
			fmt.Println("err Scan " + err.Error())
		}

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
	case "AccountItem":
		return json.Marshal(configM.AccountItemList)
	default:
		fmt.Println("unknown config type")
		break
	}
	return json.Marshal(amorM.amortizationList)
}

func (configM *ConfigModel) CreateConfigBranch(cb *ConfigBranch) (err error) {

	const sql = `INSERT INTO public.ConfigBranch
	( branch, rent, agentsign, CommercialFee)
	VALUES ($1, $2, $3, $4)
	;`

	interdb := configM.imr.GetSQLDB()
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}

	res, err := sqldb.Exec(sql, cb.Branch, cb.Rent, cb.AgentSign, cb.CommercialFee)
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

	const sql = `UPDATE public.configbranch
				SET rent=$2, agentsign=$3 ,CommercialFee=$4
				WHERE branch=$1;`

	interdb := configM.imr.GetSQLDB()
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}

	res, err := sqldb.Exec(sql, Branch, cb.Rent, cb.AgentSign, cb.CommercialFee)
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

func (configM *ConfigModel) GetConfigParameterData(today, end time.Time) []*ConfigParameter {

	const qspl = `SELECT date, nhi, LI, nhi2nd, it FROM public.ConfigParameter;`
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
		if err := rows.Scan(&cp.Date, &cp.NHI, &cp.LI, &cp.NHI2nd, &cp.IT); err != nil {
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

func (configM *ConfigModel) UpdateConfigParameter(cp *ConfigParameter) (err error) {

	const sql = `UPDATE public.configparameter
				SET nhi=$2, LI=$3, nhi2nd=$4, it=$5
				WHERE date=$1;`

	interdb := configM.imr.GetSQLDB()
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}

	res, err := sqldb.Exec(sql, cp.Date, cp.NHI, cp.LI, cp.NHI2nd, cp.IT)
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

func (configM *ConfigModel) DeleteConfigParameter(Date time.Time) (err error) {

	const sql = `DELETE FROM public.configparameter	WHERE Date=$1;`

	interdb := configM.imr.GetSQLDB()
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}

	res, err := sqldb.Exec(sql, Date)
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
	( date, nhi, LI, nhi2nd, it)
	VALUES ($1, $2, $3, $4, $5);`

	interdb := configM.imr.GetSQLDB()
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}

	res, err := sqldb.Exec(sql, cp.Date, cp.NHI, cp.LI, cp.NHI2nd, cp.IT)
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

	const qspl = `SELECT sid, sname, branch, zerodate, validdate, title, percent, fpercent,
				  salary, pay, payrollbracket, enrollment, association
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
		if err := rows.Scan(&cs.Sid, &cs.SName, &cs.Branch, &cs.ZeroDate, &cs.ValidDate, &cs.Title, &cs.Percent, &cs.FPercent,
			&cs.Salary, &cs.Pay, &cs.PayrollBracket, &cs.Enrollment, &cs.Association); err != nil {
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

func (configM *ConfigModel) CreateConfigSaler(cs *ConfigSaler) (err error) {

	const sql = `INSERT INTO public.configsaler(
		sid, sname, branch, zerodate, validdate, title, percent, fpercent, salary,
		pay, payrollbracket, enrollment, association)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13);`

	interdb := configM.imr.GetSQLDB()
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}

	res, err := sqldb.Exec(sql, cs.Sid, cs.SName, cs.Branch, cs.ZeroDate, cs.ValidDate, cs.Title, cs.Percent, cs.FPercent, cs.Salary,
		cs.Pay, cs.PayrollBracket, cs.Enrollment, cs.Association)
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

	return nil
}

func (configM *ConfigModel) UpdateConfigSaler(cs *ConfigSaler, Sid string) (err error) {

	const sql = `UPDATE public.configsaler
	SET zerodate=$2, validdate=$3, title=$4, percent=$5, fpercent=$6, salary=$7,
	pay=$8, payrollbracket=$9, enrollment=$10, association=$11	
	WHERE sid=$1 and zerodate=$2;`

	interdb := configM.imr.GetSQLDB()
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}

	res, err := sqldb.Exec(sql, Sid, cs.ZeroDate, cs.ValidDate, cs.Title, cs.Percent, cs.FPercent, cs.Salary,
		cs.Pay, cs.PayrollBracket, cs.Enrollment, cs.Association)
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
		return errors.New("Invalid operation, UpdateSaler")
	}

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
