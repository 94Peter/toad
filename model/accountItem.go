package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/94peter/toad/resource/db"
)

type AccountItem struct {
	ItemName string `json:"itemName"`
	Valid    bool   `json:"-"`
}

var (
	AIM *AccountItemModel
)

type AccountItemModel struct {
	imr             interModelRes
	db              db.InterSQLDB
	accountItemList []*AccountItem
}

func GetAccountItemModelModel(imr interModelRes) *AccountItemModel {
	if AIM != nil {
		return AIM
	}

	AIM = &AccountItemModel{
		imr: imr,
	}
	return AIM
}

func (AIM *AccountItemModel) GetAccountItemData(today, end time.Time) []*AccountItem {

	const qspl = `SELECT AccountItemName, Valid FROM public.AccountItem;`
	//const qspl = `SELECT arid,sales	FROM public.ar;`
	db := AIM.imr.GetSQLDB()
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
	AIM.accountItemList = AccountItemList
	return AIM.accountItemList
}

func (AIM *AccountItemModel) Json() ([]byte, error) {
	return json.Marshal(AIM.accountItemList)
}

func (AIM *AccountItemModel) CreateAccountItem(aitem *AccountItem) (err error) {

	const sql = `INSERT INTO public.AccountItem
	(AccountItemName)
	VALUES ($1)
	;`

	interdb := AIM.imr.GetSQLDB()
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
		return errors.New("Invalid operation, CreateAmortization")
	}

	return nil
}
