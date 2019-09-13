package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/94peter/toad/resource/db"
)

type CModel struct {
	imr   interModelRes
	db    db.InterSQLDB
	cList []*Commission
}

var (
	cm *CModel
)

type Commission struct {
	Sid      string    `json:"sid"` //no return this key 業務人員ID
	Rid      string    `json:"rid"` //收據ID
	Date     time.Time `json:"date"`
	Item     string    `json:"item"`    //合約+案名+買賣方
	Amount   int       `json:"amount"`  //金額
	Fee      int       `json:"fee"`     //口款金額收款
	SR       float64   `json:"sr"`      //實績 sales report or sales records
	Bonus    float64   `json:"bonus"`   //獎金
	CPercent float64   `json:"percent"` //比例
	SName    string    `json:"name"`    //業務姓名
	ARid     string    `json:"-"`       //程式內部比對用
}

func GetCModel(imr interModelRes) *CModel {
	if cm != nil {
		return cm
	}

	cm = &CModel{
		imr: imr,
	}
	return cm
}

func (cm *CModel) GetCommissiontData(today, end time.Time) []*Commission {
	fmt.Println("GetCommissiontData")
	//if invoiceno is null in Database return ""
	const qsql = `SELECT c.sid, c.rid, r.date, c.item, r.amount, 0 , c.sname, c.cpercent, c.sr, c.bonus, r.arid
				FROM public.commission c 
				inner JOIN public.receipt r on r.rid = c.rid;`
	//left JOIN (select sum(fee) fee, count(rid) ,arid from public.deduct group by arid) as tmp on tmp.arid = r.arid
	db := cm.imr.GetSQLDB()
	rows, err := db.SQLCommand(fmt.Sprintf(qsql))
	if err != nil {
		fmt.Println(err)
		return nil
	}
	var cDataList []*Commission
	fmt.Println("SQLCommand Done")
	for rows.Next() {
		var c Commission

		// if err := rows.Scan(&r.ARid, &s); err != nil {
		// 	fmt.Println("err Scan " + err.Error())
		// }
		if err := rows.Scan(&c.Sid, &c.Rid, &c.Date, &c.Item, &c.Amount, &c.Fee, &c.SName, &c.CPercent, &c.SR, &c.Bonus, &c.ARid); err != nil {
			fmt.Println("err Scan " + err.Error())
		}

		cDataList = append(cDataList, &c)
	}

	const feesql = `select sum(fee) fee, arid from public.deduct group by arid;`
	rows, err = db.SQLCommand(fmt.Sprintf(feesql))
	if err != nil {
		fmt.Println(err)
		return nil
	}
	for rows.Next() {
		var arid string
		var fee int
		Rid := ""
		// if err := rows.Scan(&r.ARid, &s); err != nil {
		// 	fmt.Println("err Scan " + err.Error())
		// }
		if err := rows.Scan(&fee, &arid); err != nil {
			fmt.Println("err Scan " + err.Error())
		}
		for _, commission := range cDataList {
			if Rid == commission.Rid || Rid == "" {
				//fmt.Println("arid "+arid+","+commission.ARid+" fee ", fee)
				if commission.ARid == arid {
					commission.Fee = fee
					Rid = commission.Rid
					commission.SR = float64(commission.Amount-fee) * commission.CPercent / 100
					commission.Bonus = commission.Bonus * float64(commission.Amount-fee) / float64(commission.Amount-0)
				}
			} else {
				break
			}
		}
	}

	cm.cList = cDataList
	return cm.cList
}

func (cm *CModel) Json() ([]byte, error) {
	return json.Marshal(cm.cList)
}

func (cm *CModel) CreateCommission(rt *Receipt) (err error) {

	const sql = `INSERT INTO public.commission
	(Sid, Rid, Item, SName, CPercent, sr, bonus , arid)
	select armap.sid, $1, ar.cno ||' '|| ar.casename ||' '|| ar.type, armap.sname, armap.proportion, $2 * armap.proportion / 100 ,  $2 * armap.proportion / 100 * cs.percent /100 , $3::VARCHAR
	from public.ar ar
	inner join 	public.armap armap on armap.arid = ar.arid 
	inner join 	public.configsaler cs on cs.sid = armap.sid
	where ar.arid = $3;`

	// select armap.sid, $1, ar.cno ||' '|| ar.casename ||' '|| ar.type, armap.sname, armap.proportion
	// from public.ar ar
	// inner join 	public.armap armap on armap.arid = ar.arid
	// where ar.arid = $2 ;`

	interdb := cm.imr.GetSQLDB()
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}
	fmt.Println(rt.Rid)
	fmt.Println(rt.Amount)
	fmt.Println(rt.ARid)
	res, err := sqldb.Exec(sql, rt.Rid, rt.Amount, rt.ARid)
	//res, err := sqldb.Exec(sql, unix_time, receivable.Date, receivable.CNo, receivable.Sales)
	if err != nil {
		fmt.Println("CreateCommission:", err)
		return err
	}
	id, err := res.RowsAffected()
	if err != nil {
		fmt.Println("PG Affecte Wrong: ", err)
		return err
	}
	fmt.Println(id)

	if id == 0 {
		return errors.New("Invalid operation, CreateCommission")
	}

	return nil
}

func (cm *CModel) UpdateCommission(com *Commission, rid, sid string) (err error) {
	/*
	 * 更新原則: 比例換算，新值= 舊值 / 舊的比例 * 新的比例
	 * 獎金比例使用舊的。
	 */
	const sql = `UPDATE public.commission
		SET cpercent= $1::double precision, sr= sr / cpercent * $1::double precision , bonus = bonus * $1::double precision / cpercent
		WHERE sid= $3 and rid= $2 ;`

	// const sql = `Update public.commission cm
	// 	set CPercent = $1::double precision , sr = $1 / 100.0 * tmp.amount, bonus = sr * tmp.percent / 100.0
	// 	FROM (
	// 		select cs.percent, r.amount ,
	// 		(
	// 		SELECT sum(d.fee) as fee FROM public.deduct d, public.receipt r, public.commission cm
	// 		where d.arid = r.arid and r.rid = cm.rid and cm.rid = $2
	// 		group by d.arid
	// 		)
	// 		from public.commission as cm
	// 		inner JOIN public.configsaler cs on cs.sid = cm.sid
	// 		inner JOIN public.receipt r on r.rid = cm.rid
	// 		where cm.rid =  $2 and cm.sid = $3
	// 	)as tmp
	// 	where cm.sid = $3 and  cm.rid =  $2;`

	// const sql = `Update public.commission cm
	// 			set CPercent = $1::double precision , sr = $1 / 100.0 * tmp.amount, bouns = sr * tmp.percent / 100.0
	// 			FROM (
	// 				select cs.percent , cm.sname, r.amount , cm.rid , cm.sid
	// 				from public.commission as cm
	// 				inner JOIN public.configsaler cs on cm.sid = $2
	// 				inner JOIN public.receipt r on r.rid = $3
	// 				where cm.sid = cs.bid  and cm.rid = r.rid
	// 			)as tmp
	// 			where cm.sid = $2 and  cm.rid = $3
	// 			`

	interdb := cm.imr.GetSQLDB()
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}
	//res, err := sqldb.Exec(sql)
	res, err := sqldb.Exec(sql, com.CPercent, rid, sid)
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
		return errors.New("Invalid operation, UpdateCommission")
	}

	return nil
}
