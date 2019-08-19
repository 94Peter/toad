package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/94peter/toad/resource/db"
)

type Pocket struct {
	Pid      string    `json:"id"`
	Date     time.Time `json:"date"`
	ItemName string    `json:"itemName"`
	Branch   string    `json:"branch"`
	Describe string    `json:"describe"`
	Income   int       `json:"income"`
	Fee      int       `json:"fee"`
	Balance  int       `json:"balance"`
}

var (
	pocketM *PocketModel
)

type PocketModel struct {
	imr        interModelRes
	db         db.InterSQLDB
	pocketList []*Pocket
}

func GetPocketModel(imr interModelRes) *PocketModel {
	if pocketM != nil {
		return pocketM
	}

	pocketM = &PocketModel{
		imr: imr,
	}
	return pocketM
}

func (pocketM *PocketModel) GetPocketData(today, end time.Time) []*Pocket {

	const qspl = `SELECT Pid, Date, branch, itemname, describe, income, fee, balance FROM public.pocket;`
	db := pocketM.imr.GetSQLDB()
	rows, err := db.SQLCommand(fmt.Sprintf(qspl))
	if err != nil {
		return nil
	}
	var pocketDataList []*Pocket

	for rows.Next() {
		var pocket Pocket

		// if err := rows.Scan(&r.ARid, &s); err != nil {
		// 	fmt.Println("err Scan " + err.Error())
		// }
		if err := rows.Scan(&pocket.Pid, &pocket.Date, &pocket.Branch, &pocket.ItemName, &pocket.Describe, &pocket.Income, &pocket.Fee, &pocket.Balance); err != nil {
			fmt.Println("err Scan " + err.Error())
		}

		pocketDataList = append(pocketDataList, &pocket)
	}

	// out, err := json.Marshal(arList)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(string(out))
	pocketM.pocketList = pocketDataList
	return pocketM.pocketList
}

func (pocketM *PocketModel) Json() ([]byte, error) {
	return json.Marshal(pocketM.pocketList)
}

func (pocketM *PocketModel) CreatePocket(pocket *Pocket) (err error) {

	const sql = `INSERT INTO public.pocket
	(Pid , date, branch, itemname, describe, income, fee, balance)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	;`

	interdb := pocketM.imr.GetSQLDB()
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}

	fakeId := time.Now().Unix()

	res, err := sqldb.Exec(sql, fakeId, time.Now(), pocket.Branch, pocket.ItemName, pocket.Describe, pocket.Income, pocket.Fee, pocket.Balance)
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
		return errors.New("Invalid operation, CreatePocket")
	}

	return nil
}
