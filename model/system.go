package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"toad/resource/db"
)

//取名有點誤會，這是給業務員用的struct
type SystemAccount struct {
	Account string `json:"account"`
	Name    string `json:"name"`
	//Branch  string `json:"branch"`
	Email      string `json:"email"`
	Auth       string `json:"auth"`
	Password   string `json:"-"`
	Phone      string `json:"-"` //保留欄位 目前沒用到
	CreateDate string `json:"createDate"`
}

type SystemBranch struct {
	Branch string `json:"branch"`
}

var (
	systemM *SystemModel
)
var (
	auth_token = "eyJhbGciOiJSUzI1NiIsImtpZCI6ImRldiIsInR5cCI6IkpXVCJ9.eyJleHAiOjEwNzkzOTAyMzIxLCJpYXQiOjE1NzA1MzAyODQsImlzcyI6InBpY2Fpc3MiLCJzeXMiOiJ0b2FkIn0.dCeCH2cYCm5MewP2lCpLGJV4ka4C8j4joHL23YlphRQJpOemKBRLReCXKFQh1GhdnFKXh6xh9ULox_BUBZxckdRDoJo5-R7fXM7eOy5hIRFyOwO8FOuKJ50QddR0qoLbuLbzIklJncxDRftBcujuOFFAFEBIkR5Nq9TyBEgIkSI"
	picaURL    = "https://pica957.appspot.com/"
)

type SystemModel struct {
	imr               interModelRes
	db                db.InterSQLDB
	systemAccountList []*SystemAccount
	systemBranchList  []*SystemBranch
}

func GetSystemModel(imr interModelRes) *SystemModel {
	if systemM != nil {
		return systemM
	}

	systemM = &SystemModel{
		imr: imr,
	}
	return systemM
}

func (systemM *SystemModel) GetAccountData(dbname string) error {
	const qspl = `SELECT account, name, permission, createdate, lasttime, state, disable FROM public.account;`
	//const qspl = `SELECT arid,sales	FROM public.ar;`
	db := systemM.imr.GetSQLDBwithDbname(dbname)
	rows, err := db.SQLCommand(fmt.Sprintf(qspl))
	if err != nil {
		fmt.Println(err)
		return err
	}
	var saDataList []*SystemAccount

	for rows.Next() {
		var sa SystemAccount

		//var lasttime NullTime

		// if err := rows.Scan(&r.ARid, &s); err != nil {
		// 	fmt.Println("err Scan " + err.Error())
		// }
		// if err := rows.Scan(&user.Account, &user.Name, &user.Permission, &user.CreateDate, &lasttime, &user.State, &disable); err != nil {
		// 	fmt.Println("err Scan " + err.Error())
		// }

		if err := rows.Scan(&sa.Account, &sa.Name, &sa.Auth, &sa.CreateDate, &sa.Email, &sa.Phone); err != nil {
			fmt.Println("err Scan " + err.Error())
		}

		saDataList = append(saDataList, &sa)
	}

	// out, err := json.Marshal(arList)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(string(out))
	systemM.systemAccountList = saDataList
	return nil
}

func (systemM *SystemModel) GetAccountDataFromPICA(branch string) ([]byte, error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", picaURL+"v1/toad/user", nil)
	if err != nil {
		// handle error
		fmt.Println(err)
	}

	req.Header.Set("auth-token", auth_token)
	q := req.URL.Query()
	q.Add("c", branch)
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		// handle error
		fmt.Println(err)
	}
	defer resp.Body.Close()

	sitemap, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil
	}

	if len(sitemap) <= 0 {
		fmt.Println("nil")
		return nil, nil
	}

	fmt.Println("sitemap\n" + string(sitemap))

	var systemAccountList []*SystemAccount
	err = json.Unmarshal(sitemap, &systemAccountList)
	if err != nil {
		return nil, err
	} else {
		out, err := json.Marshal(systemAccountList)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		fmt.Println("Out\n" + string(out))
	}
	systemM.systemAccountList = systemAccountList
	return sitemap, err
}

func (systemM *SystemModel) GetBranchDataFromPICA() ([]byte, error) {
	// var systemBranchDataList []*SystemBranch
	// var s1, s2, s3, s4 SystemBranch
	// s1.Branch = "北京店"
	// s2.Branch = "東京店"
	// s3.Branch = "西京店"
	// s4.Branch = "南京店"
	// systemBranchDataList = append(systemBranchDataList, &s1)
	// systemBranchDataList = append(systemBranchDataList, &s2)
	// systemBranchDataList = append(systemBranchDataList, &s3)
	// systemBranchDataList = append(systemBranchDataList, &s4)

	// out, err := json.Marshal(arList)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(string(out))
	// systemM.systemBranchList = systemBranchDataList

	client := &http.Client{}

	req, err := http.NewRequest("GET", picaURL+"v1/toad/category", nil)
	if err != nil {
		// handle error
		fmt.Println(err)
	}

	req.Header.Set("auth-token", auth_token)

	resp, err := client.Do(req)
	if err != nil {
		// handle error
		fmt.Println(err)
	}
	defer resp.Body.Close()

	sitemap, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil
	}

	fmt.Println(string(sitemap))

	return sitemap, err
}

func (systemM *SystemModel) GetBranchData(dbname string) ([]byte, error) {
	//if invoiceno is null in Database return ""
	//const qspl = `SELECT rid, date, cno, casename, type, name, amount, COALESCE(NULLIF(invoiceno, null),'') FROM public.receipt;`
	//left join public.invoice I on  I.Rid = R.rid
	//
	fmt.Println("GetBranchData")
	const qspl = `SELECT branch	FROM public.configbranch;`
	db := systemM.imr.GetSQLDBwithDbname(dbname)
	rows, err := db.SQLCommand(qspl)
	if err != nil {
		return nil, nil
	}
	branchList := []string{}
	for rows.Next() {
		var branch NullString

		// if err := rows.Scan(&r.ARid, &s); err != nil {
		// 	fmt.Println("err Scan " + err.Error())
		// }
		if err := rows.Scan(&branch); err != nil {
			fmt.Println("err Scan " + err.Error())
		}

		branchList = append(branchList, branch.Value)
	}

	//stringSlice := []string{"hello", "bye"}
	// stringByte := "\x00" + strings.Join(branchList, "\x20\x00") // x20 = space and x00 = null
	// fmt.Println([]byte(stringByte))
	// fmt.Println(string([]byte(stringByte)))
	fmt.Println(branchList)
	data := "["
	for i, str := range branchList {
		if i != 0 {
			data += ","
		}
		data += "\"" + str + "\""
	}
	data += "]"
	fmt.Println(data)
	return []byte(data), nil
}

func (systemM *SystemModel) Json(mtype string) ([]byte, error) {
	switch mtype {
	case "branch":
		return json.Marshal(systemM.systemBranchList)
	case "account":
		return json.Marshal(systemM.systemAccountList)
	default:
		fmt.Println("unknown config type")
		break
	}
	return nil, nil
}

func (systemM *SystemModel) CreateSystemAccount(systemAccount *SystemAccount, dbname string) (err error) {

	const sql = `INSERT INTO public.account(account, passoword, name, auth, createdate, email, phone)	VALUES ($1, $2, $3, $4, $5, $6, $7) ON CONFLICT (account) DO nothing;`
	//and ( select sum(amount)+$3 FROM public.receipt  where arid = $4 group by arid ) <=  (SELECT amount from public.ar ar WHERE arid = $4);`

	interdb := systemM.imr.GetSQLDBwithDbname(dbname)
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}

	res, err := sqldb.Exec(sql, systemAccount.Account, encryptPassword(systemAccount.Password), systemAccount.Name, systemAccount.Auth, time.Now(), systemAccount.Email, systemAccount.Phone)
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
		return errors.New("Invalid operation, CreateSystemAccount")
	}

	return nil
}

func (systemM *SystemModel) UpdateSystemAccount(systemAccount *SystemAccount, dbname string) (err error) {

	const sql = `UPDATE public.account 
	SET name=$3 ,auth=$4 , email = $5 , phone = $6
	Where account = $1 and passoword=$2;`

	interdb := systemM.imr.GetSQLDBwithDbname(dbname)
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}

	res, err := sqldb.Exec(sql, systemAccount.Account, encryptPassword(systemAccount.Password), systemAccount.Name, systemAccount.Auth, systemAccount.Email, systemAccount.Phone)
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
		return errors.New("Invalid operation, UpdateSystemAccount")
	}

	return nil
}

func encryptPassword(pwd string) string {
	return pwd
	//return util.MD5(pwd)
}

func (systemM *SystemModel) UpdateSystemAccountPassword(newpassword, dbname string, systemAccount *SystemAccount) (err error) {

	const sql = `UPDATE public.account 
	SET passoword=$3
	Where account = $1 and passoword=$2;`

	interdb := systemM.imr.GetSQLDBwithDbname(dbname)
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}

	res, err := sqldb.Exec(sql, systemAccount.Account, encryptPassword(systemAccount.Password), encryptPassword(newpassword))
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
		return errors.New("Invalid operation, UpdateSystemAccount")
	}

	return nil
}

func (systemM *SystemModel) DeleteSystemAccount(account, dbname string) (err error) {

	const sql = `DELETE FROM public.account WHERE account = $1;`

	interdb := systemM.imr.GetSQLDBwithDbname(dbname)
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return err
	}

	res, err := sqldb.Exec(sql, account)
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
		return errors.New("Invalid operation, DeleteSystemAccount")
	}

	return nil
}
