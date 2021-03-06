package model

import (
	"errors"
	"fmt"
	"time"

	"toad/resource/db"
	"toad/util"

	"github.com/mitchellh/mapstructure"
)

const (
	categoryC       = "category"
	userC           = "user"
	UserPerSales    = "sales"
	UserStateInit   = "init"
	UserStateNormal = "normal"
	UserStateReset  = "reset"
)

type interModelRes interface {
	GetSQLDB() db.InterSQLDB
	GetDB() db.InterDB //firebase
	GetSQLDBwithDbname(string) db.InterSQLDB
}

var (
	memM *memberModel
)

type memberModel struct {
	di interModelRes
	cu *categoryUser
	Cu *categoryUser
}

type categoryUser struct {
	db             db.InterDB
	sqldb          db.InterSQLDB
	DictionaryUser map[string]*User   `json:"-"`
	CategoryUsers  map[string][]*User `json:"c"`
}

type User struct {
	Account    string    `json:"account"`
	Name       string    `json:"name"`
	Permission string    `json:"permission"`
	Dbname     string    `json:"-"`
	Password   string    `json:"-"`
	CreateDate time.Time `json:"createDate"`
	Lasttime   time.Time `json:"lasttime"`
	State      string    `json:"state"`
	Disable    bool      `json:"disable"`
	Category   string    `json:"-"`
	Branch     string    `json:"branch"`
}

func GetMemberModel(mr interModelRes) *memberModel {
	cu := &categoryUser{
		db: mr.GetDB(),
		//sqldb: mr.GetSQLDB(),
	}
	cu.load()

	memM = &memberModel{
		cu: cu,
		di: mr,
	}
	return memM
}

func (dc *categoryUser) GetID() string {
	const id = "1"
	return id
}

func (dc *categoryUser) load() error {
	if dc.db == nil {
		fmt.Println("db not set")
		return errors.New("db not set")
	}
	// err := dc.db.C(userC).GetByID(dc.GetID(), dc)
	// if err != nil {
	// 	return err
	// }
	// dc.DictionaryUser = make(map[string]*User)
	// for _, s := range dc.CategoryUsers {
	// 	for _, u := range s {
	// 		dc.DictionaryUser[u.Account] = u
	// 	}
	// }
	// fmt.Println(dc)
	return nil
}

func (dc *categoryUser) test(phone, displayName, email, pwd, permission string) {

	//err = aa.Cu.Db.CreateUser("0919966667", "peter", "ch.focke@gmail.com", "password", "admin")
}

//phone, displayName, email, pwd, permission string
func (memM *memberModel) CreateUser(user *User) error {
	//user.Dbname = "test"
	//phone=>Account也用帶入。
	err := memM.cu.db.CreateUser(user.Account, user.Name, user.Account, user.Password, user.Permission, user.Dbname)
	if err != nil {
		fmt.Println("CreateUser:", err)
		return err
	}

	err = memM.UpdateState(user.Account, UserStateInit)
	if err != nil {
		fmt.Println("CreateUser UpdateState:", err)
		return err
	}
	/*
	* Local DB 資訊儲存
	 */
	const sql = `INSERT INTO public.account
	(account , name, permission, state , branch)
	VALUES ($1, $2, $3, $4, $5)	;`

	sqldb, err := memM.di.GetSQLDBwithDbname(user.Dbname).ConnectSQLDB()
	if err != nil {
		return err
	}

	res, err := sqldb.Exec(sql, user.Account, user.Name, user.Permission, UserStateInit, user.Branch)
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
		return errors.New("[CreateUser]: save Local DB Error")
	}
	defer sqldb.Close()
	return nil
}

func (memM *memberModel) DeleteUser(uid, dbname string) error {
	err := memM.cu.db.DeleteUser(uid)
	if err != nil {
		fmt.Println(err)
		return err
	}
	/*
	* Local DB 資訊儲存
	 */
	const sql = `DELETE FROM public.account WHERE account = $1`

	sqldb, err := memM.di.GetSQLDBwithDbname(dbname).ConnectSQLDB()
	if err != nil {
		return err
	}

	res, err := sqldb.Exec(sql, uid)
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
		return errors.New("[DeleteUser]: save Local DB Error")
	}
	defer sqldb.Close()
	return nil
}

func (memM *memberModel) SetUserDisable(uid, dbname string, disable bool) error {
	err := memM.cu.db.SetUserDisable(uid, disable)
	if err != nil {
		fmt.Println(err)
		return err
	}
	/*
	* Local DB 資訊儲存
	 */
	sqldb, err := memM.di.GetSQLDBwithDbname(dbname).ConnectSQLDB()
	if err != nil {
		return err
	}
	const sql = `UPDATE public.account SET disable = $2 WHERE account = $1`
	// sqldb, err := memM.cu.sqldb.ConnectSQLDB()
	// if err != nil {
	// 	return err
	// }
	setAble := 0
	if disable {
		setAble = 1
	}

	res, err := sqldb.Exec(sql, uid, setAble)
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
		return errors.New("[SetUserDisable]: save Local DB Error")
	}
	defer sqldb.Close()
	return nil
}

func (memM *memberModel) ChangePwd(uid string, pwd string) error {
	err := memM.cu.db.ChangePwd(uid, pwd)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func (memM *memberModel) UpdateState(uid string, state string) error {
	err := memM.cu.db.UpdateState(uid, state)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func (memM *memberModel) UpdateDbname(uid string, dbname string) error {
	err := memM.cu.db.UpdateDbname(uid, dbname)
	if err != nil {
		fmt.Println(err)
		return err
	}
	db := memM.di.GetSQLDBwithDbname(dbname)
	db.InitDB()
	return nil
}

func (memM *memberModel) UpdateUser(user *User, dbname string) error {
	err := memM.cu.db.UpdateUser(user.Account, user.Name, user.Permission, dbname)
	if err != nil {
		fmt.Println(err)
		return err
	}
	/*
	* Local DB 資訊儲存
	 */
	const sql = `UPDATE public.account SET name = $2 , permission = $3 , branch = $4 WHERE account = $1`

	//sqldb, err := memM.cu.sqldb.ConnectSQLDB()
	sqldb, err := memM.di.GetSQLDBwithDbname(dbname).ConnectSQLDB()
	if err != nil {
		return err
	}

	res, err := sqldb.Exec(sql, user.Account, user.Name, user.Permission, user.Branch)
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
		return errors.New("[UpdateUser]: save Local DB Error")
	}
	defer sqldb.Close()
	return nil
}

// func (memM *memberModel) VerifyToken(idToken string) string {
// 	res, err := memM.cu.db.VerifyToken(idToken)
// 	if err != nil {
// 		fmt.Println(err)
// 		return ""
// 	}
// 	return res
// }

func (memM *memberModel) VerifyToken(ftoken string) *User {
	uid, err := memM.cu.db.VerifyToken(ftoken)
	if err != nil {
		return nil
	}
	claim, err := memM.cu.db.GetUser(uid)
	if err != nil {
		fmt.Println(err.Error())
	}

	//convert map[string] to struct
	user := User{}
	mapstructure.Decode(claim, &user)
	//user.Permission = permission.Office
	user.State = "OK"
	ubranch, err := memM.GetAccountUserDataByID(uid, user.Dbname)
	if err == nil {
		user.Branch = ubranch.Branch
	}

	return &user
}

func (memM *memberModel) GetAccountUserData(dbname string) ([]*User, error) {

	const qspl = `SELECT account, name, permission, createdate, lasttime, state, disable, branch FROM public.account;`
	//(Date >= '%s' and Date < ('%s'::date + '1 month'::interval))
	//const qspl = `SELECT arid,sales	FROM public.ar;`
	//db := memM.cu.sqldb
	db := memM.di.GetSQLDBwithDbname(dbname)

	rows, err := db.SQLCommand(fmt.Sprintf(qspl))
	if err != nil {
		return nil, err
	}
	var userDataList []*User

	for rows.Next() {
		var user User
		var lasttime NullTime
		var disable = 0
		// if err := rows.Scan(&r.ARid, &s); err != nil {
		// 	fmt.Println("err Scan " + err.Error())
		// }
		if err := rows.Scan(&user.Account, &user.Name, &user.Permission, &user.CreateDate, &lasttime, &user.State, &disable, &user.Branch); err != nil {
			fmt.Println("err Scan " + err.Error())
		}
		user.Lasttime = lasttime.Time
		userDataList = append(userDataList, &user)
	}

	// out, err := json.Marshal(arList)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(string(out))

	return userDataList, nil
}

func (u *User) GetToken(jwtConf *util.JwtConf) (string, error) {

	token, err := jwtConf.GetToken(map[string]interface{}{
		"sub":    u.Account,
		"nam":    u.Name,
		"per":    u.Permission,
		"cat":    u.Category,
		"dbname": u.Dbname,
	})
	if err != nil {
		return "", err
	}
	return *token, nil
}

func (memM *memberModel) GetAccountUserDataByID(account, dbname string) (*User, error) {

	const qspl = `SELECT account, name, permission, createdate, lasttime, state, disable, branch FROM public.account where account = '%s';`
	//(Date >= '%s' and Date < ('%s'::date + '1 month'::interval))
	//const qspl = `SELECT arid,sales	FROM public.ar;`

	//db := memM.cu.sqldb
	db := memM.di.GetSQLDBwithDbname(dbname)

	rows, err := db.SQLCommand(fmt.Sprintf(qspl, account))
	if err != nil {
		return nil, err
	}
	user := &User{}

	for rows.Next() {

		var lasttime NullTime
		var disable = 0
		// if err := rows.Scan(&r.ARid, &s); err != nil {
		// 	fmt.Println("err Scan " + err.Error())
		// }
		if err := rows.Scan(&user.Account, &user.Name, &user.Permission, &user.CreateDate, &lasttime, &user.State, &disable, &user.Branch); err != nil {
			fmt.Println("err Scan " + err.Error())
		}
		user.Lasttime = lasttime.Time
	}

	// out, err := json.Marshal(arList)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(string(out))

	return user, nil
}
