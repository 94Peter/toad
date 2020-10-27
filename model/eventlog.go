package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"toad/resource/db"
)

type EventLog struct {
	Account string    `json:"account"`
	Name    string    `json:"name"`
	Msg     string    `json:"msg"`
	Auth    string    `json:"auth"`
	Type    string    `json:"type"`
	Date    time.Time `json:"date"`
}

var (
	logM *EventLogModel
)

type EventLogModel struct {
	imr          interModelRes
	db           db.InterSQLDB
	eventLogList []*EventLog
}

func GetEventLogModel(imr interModelRes) *EventLogModel {
	if logM != nil {
		return logM
	}

	logM = &EventLogModel{
		imr: imr,
	}
	return logM
}

func (logM *EventLogModel) GetEventLogData(dbname string) error {
	const qspl = `SELECT account, name, auth, date, type, msg FROM public.eventlog;`
	//const qspl = `SELECT arid,sales	FROM public.ar;`
	db := logM.imr.GetSQLDBwithDbname(dbname)
	rows, err := db.SQLCommand(fmt.Sprintf(qspl))
	if err != nil {
		return nil
	}
	var logList []*EventLog

	for rows.Next() {
		var log EventLog

		// if err := rows.Scan(&r.ARid, &s); err != nil {
		// 	fmt.Println("err Scan " + err.Error())
		// }
		if err := rows.Scan(&log.Account, &log.Name, &log.Auth, &log.Date, &log.Type, &log.Msg); err != nil {
			fmt.Println("err Scan " + err.Error())
		}

		logList = append(logList, &log)
	}

	// out, err := json.Marshal(arList)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(string(out))
	logM.eventLogList = logList
	return nil
}

func (logM *EventLogModel) Json(mtype string) ([]byte, error) {
	switch mtype {
	case "branch":
		return json.Marshal(logM.eventLogList)
	case "account":
		return json.Marshal(logM.eventLogList)
	default:
		fmt.Println("unknown config type")
		break
	}
	return json.Marshal(logM.eventLogList)
}

func (logM *EventLogModel) CreateEventLog(eventLog *EventLog, dbname string) (err error) {

	const sql = `INSERT INTO public.EventLog(id, account, name, auth, msg, type, date)	VALUES ($1, $2, $3, $4, $5, $6, $7) ON CONFLICT (id) DO nothing;`
	//and ( select sum(amount)+$3 FROM public.receipt  where arid = $4 group by arid ) <=  (SELECT amount from public.ar ar WHERE arid = $4);`

	interdb := logM.imr.GetSQLDBwithDbname(dbname)
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}
	fakeid := time.Now().Unix()

	res, err := sqldb.Exec(sql, fakeid, eventLog.Account, eventLog.Name, eventLog.Auth, eventLog.Msg, eventLog.Type, time.Now())
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
		return errors.New("Invalid operation, CreateEventLog")
	}

	return nil
}
