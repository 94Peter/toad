package db

import (
	"context"
	"database/sql"
	"fmt"
)

type interDoc interface {
	GetID() string
}

type InterDB interface {
	C(c string) InterDB
	Save(doc interDoc) error
	CreateUser(phone, displayName, email, pwd, permission, dbname string) error
	DeleteUser(uid string) error
	SetUserDisable(uid string, disable bool) error
	ChangePwd(uid string, pwd string) error
	UpdateState(uid string, state string) error
	UpdateDbname(uid string, dbname string) error
	UpdateUser(uid, display, permission, dbname string) error
	VerifyToken(idToken string) (string, error)
	GetByID(id string, doc interface{}) error
	GetUser(uid string) (map[string]interface{}, error)
}

type InterSQLDB interface {
	C(c string) InterSQLDB
	Close() error

	SQLCommand(cmd string) (res *sql.Rows, err error)
	ConnectSQLDB() (*sql.DB, error)

	InitDB() bool
	//CreateDB() error
	//CreateARTable() error
	//CreateReceiptTable() error
	//InitTable() error
}

type DBConf struct {
	FirebaseConf *firebaseConf `yaml:"firebase"`
	SqlDBConf    *SqldbConf    `yaml:"sqldatabase"`

	db    InterDB
	sqldb InterSQLDB
}

func (dbc *DBConf) SetFirebase(file, url string) {
	dbc.FirebaseConf = &firebaseConf{
		CredentialsFile: file,
		DatabaseURL:     url,
	}
}

func (dbc *DBConf) SetSqldatabase(host, user, password, db string, port int) {
	dbc.SqlDBConf = &SqldbConf{
		DatabaseURL: host,
		Port:        port,
		User:        user,
		Password:    password,
		DB:          db,
	}
}

type firebaseConf struct {
	CredentialsFile string `yaml:"credentialsFile"`
	DatabaseURL     string `yaml:"databaseURL"`
}

type SqldbConf struct {
	DatabaseURL string `yaml:"databaseURL"`
	Port        int    `yaml:"port"`
	User        string `yaml:"user"`
	Password    string `yaml:"password"`
	DB          string `yaml:"db"`
}

func (dbc *DBConf) GetDB() InterDB {
	if dbc.db != nil {
		return dbc.db
	}

	dbc.db = &firebaseDB{
		credentialsFile: dbc.FirebaseConf.CredentialsFile,
		ctx:             context.Background(),
		dburl:           dbc.FirebaseConf.DatabaseURL,
	}

	return dbc.db
}

func (dbc *DBConf) GetSQLDB() InterSQLDB {
	fmt.Println("GetSQLDB")
	if dbc.sqldb != nil {
		return dbc.sqldb
	}
	//fmt.Println("GetSQLDB")

	dbc.sqldb = &SqlDB{
		Ctx:      context.Background(),
		Dburl:    dbc.SqlDBConf.DatabaseURL,
		User:     dbc.SqlDBConf.User,
		Password: dbc.SqlDBConf.Password,
		Db:       dbc.SqlDBConf.DB,
		Port:     dbc.SqlDBConf.Port,
	}

	dbc.sqldb.InitDB()

	return dbc.sqldb
}
