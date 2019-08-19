package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/94peter/toad/resource/db"
)

type PrePay struct {
	PPid     string          `json:"id"`
	Date     time.Time       `json:"date"`
	ItemName string          `json:"itemName"`
	Describe string          `json:"describe"`
	Fee      int             `json:"fee"`
	PrePay   []*BranchPrePay `json:"prepay"`
}

type BranchPrePay struct {
	PPid   string `json:"id"`
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

func (prepayM *PrePayModel) GetPrePayData(today, end time.Time) []*PrePay {

	const PrePayspl = `SELECT PPid, Date, itemname, describe, fee FROM public.PrePay;`
	db := prepayM.imr.GetSQLDB()
	rows, err := db.SQLCommand(fmt.Sprintf(PrePayspl))
	if err != nil {
		return nil
	}
	var prepayDataList []*PrePay

	for rows.Next() {
		var prepay PrePay

		if err := rows.Scan(&prepay.PPid, &prepay.Date, &prepay.ItemName, &prepay.Describe, &prepay.Fee); err != nil {
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

func (prepayM *PrePayModel) CreatePrePay(prepay *PrePay) (err error) {

	const sql = `INSERT INTO public.prepay
	(ppid, date, itemname, describe, fee)
	VALUES ($1, $2, $3, $4, $5)
	;`

	interdb := prepayM.imr.GetSQLDB()
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}

	fakeId := time.Now().Unix()

	res, err := sqldb.Exec(sql, fakeId, prepay.Date, prepay.ItemName, prepay.Describe, prepay.Fee)
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
