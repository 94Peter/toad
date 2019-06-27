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
	GetByID(id string, doc interface{}) error
}

type InterSQLDB interface {
	C(c string) InterSQLDB
	Close() error
	Query(cmd string) (res *sql.Rows, err error)

	IsDBExist() bool
	CreateDB() error
	CreateTable() error
}

type DBConf struct {
	FirebaseConf *firebaseConf `yaml:"firebase"`
	SqlDBConf    *sqldbConf    `yaml:"sqldatabase"`

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
	dbc.SqlDBConf = &sqldbConf{
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

type sqldbConf struct {
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
	if dbc.sqldb != nil {
		return dbc.sqldb
	}
	fmt.Println("GetSQLDB")

	dbc.sqldb = &sqlDB{
		ctx:      context.Background(),
		dburl:    dbc.SqlDBConf.DatabaseURL,
		user:     dbc.SqlDBConf.User,
		password: dbc.SqlDBConf.Password,
		db:       dbc.SqlDBConf.DB,
		port:     dbc.SqlDBConf.Port,
	}

	if !dbc.sqldb.IsDBExist() {
		dbc.sqldb.CreateDB()
		dbc.sqldb.CreateTable()
	}
	return dbc.sqldb
}
