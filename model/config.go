package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/94peter/toad/resource/db"
)

type ConfigParameter struct {
	Param string  `json:"param"`
	Value float64 `json:"value"`
}

type AccountItem struct {
	ItemName string `json:"itemName"`
	Valid    bool   `json:"-"`
}

type ConfigBranch struct {
	Branch    string `json:"branch"`
	Rent      int    `json:"rent"`
	AgentSign int    `json:"agentSign"`
}

type ConfigBusiness struct {
	Bid       string    `json:"id"`
	BName     string    `json:"name"`
	ZeroDate  time.Time `json:"zeroDate"`
	ValidDate time.Time `json:"validDate"`
	Title     string    `json:"title"`
	Salary    int       `json:"salary"`
	Pay       int       `json:"pay"`
	Percent   float64   `json:"percent"`
}

var (
	configM *ConfigModel
)

type ConfigModel struct {
	imr                 interModelRes
	db                  db.InterSQLDB
	ConfigBranchList    []*ConfigBranch
	ConfigParameterList []*ConfigParameter
	ConfigBusinessList  []*ConfigBusiness
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

	const qspl = `SELECT branch, rent, AgentSign FROM public.ConfigBranch;`
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
		if err := rows.Scan(&cb.Branch, &cb.Rent, &cb.AgentSign); err != nil {
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
		f := map[string]interface{}{}

		for _, param := range configM.ConfigParameterList {
			fmt.Println(param.Param)
			f[param.Param] = param.Value
		}

		return json.Marshal(f)
		//return json.Marshal(configM.ConfigParameterList)
	case "ConfigBusiness":
		return json.Marshal(configM.ConfigBusinessList)
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
	( branch, rent, agentsign)
	VALUES ($1, $2, $3)
	;`

	interdb := configM.imr.GetSQLDB()
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}

	res, err := sqldb.Exec(sql, cb.Branch, cb.Rent, cb.AgentSign)
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

func (configM *ConfigModel) UpdateConfigBranch(cb *ConfigBranch) (err error) {

	const sql = `UPDATE public.configbranch
				SET rent=$2, agentsign=$3
				WHERE branch=$1;`

	interdb := configM.imr.GetSQLDB()
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}

	res, err := sqldb.Exec(sql, cb.Branch, cb.Rent, cb.AgentSign)
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

	const qspl = `SELECT param, value FROM public.ConfigParameter;`
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
		if err := rows.Scan(&cp.Param, &cp.Value); err != nil {
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

func (configM *ConfigModel) UpdateConfigParameter(cp []*ConfigParameter) (err error) {

	const sql = `UPDATE public.configparameter
				SET value=$2
				WHERE param=$1;`

	interdb := configM.imr.GetSQLDB()
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}
	for _, param := range cp {

		res, err := sqldb.Exec(sql, param.Param, param.Value)
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

	}

	return nil
}

func (configM *ConfigModel) CreateConfigParameter(cp ConfigParameter) (err error) {

	const sql = `INSERT INTO public.ConfigParameter
	( param, value)
	VALUES ($1, $2)
	;`

	interdb := configM.imr.GetSQLDB()
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}

	res, err := sqldb.Exec(sql, cp.Param, cp.Value)
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

func (configM *ConfigModel) GetConfigBusinessData(today, end time.Time) []*ConfigBusiness {

	const qspl = `SELECT Bid, BName, ZeroDate, Title, Salary, Percent, Pay ,ValidDate FROM public.ConfigBusiness;`
	//const qspl = `SELECT arid,sales	FROM public.ar;`
	db := configM.imr.GetSQLDB()
	rows, err := db.SQLCommand(fmt.Sprintf(qspl))
	if err != nil {
		return nil
	}
	var cbDataList []*ConfigBusiness

	for rows.Next() {
		var cb ConfigBusiness

		// if err := rows.Scan(&r.ARid, &s); err != nil {
		// 	fmt.Println("err Scan " + err.Error())
		// }
		if err := rows.Scan(&cb.Bid, &cb.BName, &cb.ZeroDate, &cb.Title, &cb.Salary, &cb.Percent, &cb.Pay, &cb.ValidDate); err != nil {
			fmt.Println("err Scan " + err.Error())
		}

		cbDataList = append(cbDataList, &cb)
	}

	// out, err := json.Marshal(arList)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(string(out))
	configM.ConfigBusinessList = cbDataList
	return configM.ConfigBusinessList
}

func (configM *ConfigModel) CreateConfigBusiness(cb *ConfigBusiness) (err error) {

	const sql = `INSERT INTO public.configbusiness(
		bid, bname, zerodate, validdate, title, percent, salary, pay)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8);
	;`

	interdb := configM.imr.GetSQLDB()
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}

	res, err := sqldb.Exec(sql, cb.Bid, cb.BName, cb.ZeroDate, cb.ValidDate, cb.Title, cb.Percent, cb.Salary, cb.Pay)
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
		return errors.New("Invalid operation, ConfigBusiness")
	}

	return nil
}

func (configM *ConfigModel) UpdateConfigBusiness(cb *ConfigBusiness) (err error) {

	const sql = `UPDATE public.configbusiness
	SET  bname=$2, zerodate=$3, validdate=$4, title=$5, percent=$6, salary=$7, pay=$8
	WHERE bid=$1;`

	interdb := configM.imr.GetSQLDB()
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}

	res, err := sqldb.Exec(sql, cb.Bid, cb.BName, cb.ZeroDate, cb.ValidDate, cb.Title, cb.Percent, cb.Salary, cb.Pay)
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
		return errors.New("Invalid operation, UpdateBusiness")
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
