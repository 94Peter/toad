package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/94peter/toad/resource/db"
)

type PrePay struct {
	PPid        string          `json:"id"`
	Date        time.Time       `json:"date"`
	ItemName    string          `json:"itemName"`
	Description string          `json:"description"`
	Fee         int             `json:"fee"`
	PrePay      []*BranchPrePay `json:"prepay"`
}

type BranchPrePay struct {
	PPid   string `json:"-"`
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

	const PrePayspl = `SELECT PPid, Date, itemname, description, fee FROM public.PrePay;`
	db := prepayM.imr.GetSQLDB()
	rows, err := db.SQLCommand(fmt.Sprintf(PrePayspl))
	if err != nil {
		return nil
	}
	var prepayDataList []*PrePay

	for rows.Next() {
		var prepay PrePay

		if err := rows.Scan(&prepay.PPid, &prepay.Date, &prepay.ItemName, &prepay.Description, &prepay.Fee); err != nil {
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

func (prepayM *PrePayModel) DeletePrePay(ID string) (err error) {
	const sql = `
				delete from public.PrePay where PPid = '%s';
				delete from public.BranchPrePay where PPid = '%s';				
				`

	interdb := prepayM.imr.GetSQLDB()
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}

	res, err := sqldb.Exec(fmt.Sprintf(sql, ID, ID))
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
	if id <= 0 {
		return errors.New("not found anything")
	}
	return nil
}

func (prepayM *PrePayModel) CreatePrePay(prepay *PrePay) (err error) {

	const sql = `INSERT INTO public.prepay
	(ppid, date, itemname, description, fee)
	VALUES ($1, $2, $3, $4, $5)
	;`

	interdb := prepayM.imr.GetSQLDB()
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}

	fakeId := time.Now().Unix()

	res, err := sqldb.Exec(sql, fakeId, prepay.Date, prepay.ItemName, prepay.Description, prepay.Fee)
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

func (prepayM *PrePayModel) UpdatePrePay(ID string, prepay *PrePay) (err error) {

	const sql = `UPDATE public.prepay
	SET date= $2, itemname= $3, description= $4, fee=$5
	WHERE ppid = $1;
	;`

	interdb := prepayM.imr.GetSQLDB()
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}

	res, err := sqldb.Exec(sql, ID, prepay.Date, prepay.ItemName, prepay.Description, prepay.Fee)
	//res, err := sqldb.Exec(sql, unix_time, receivable.Date, receivable.CNo, receivable.Sales)
	if err != nil {
		fmt.Println("[UPDATE err] ", err)
		return err
	}
	id, err := res.RowsAffected()
	if err != nil {
		fmt.Println("PG Affecte Wrong: ", err)
		return err
	}
	fmt.Println(id)

	if id == 0 {
		return errors.New("Not found prepay")
	}

	const DELbpp = `DELETE FROM public.branchprepay	WHERE ppid= $1 ;`
	_, err = sqldb.Exec(DELbpp, ID)
	if err != nil {
		fmt.Println("DELETE branchprepay Wrong: ", err)
		return err
	}
	const bppsql = `INSERT INTO public.branchprepay
					(ppid, branch, cost)
					VALUES ($1, $2, $3)				
					;`

	i := 0
	for range prepay.PrePay {
		bppres, err := sqldb.Exec(bppsql, ID, prepay.PrePay[i].Branch, prepay.PrePay[i].Cost)
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
			return errors.New("Invalid operation , UpdateBranchPrepay")
		}
		i++
	}
	return nil
}
