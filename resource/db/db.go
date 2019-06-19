package db

import (
	"context"
	"time"

	influx "github.com/influxdata/influxdb/client/v2"
)

type interDoc interface {
	GetID() string
}

type InterDB interface {
	C(c string) InterDB
	Save(doc interDoc) error
	GetByID(id string, doc interface{}) error
}

type InterAuth interface {
	CreateUser(phone, displayName, email, pwd, permission string) error
	SetUserDisable(uid string, disable bool) error
	ChangePwd(uid, newPwd string) error
	UpdateState(uid, state string) error
	UpdateUser(uid, displayName, permission string) error
	VerifyToken(idToken string) (string, error)
	DeleteUser(uid string) error
}

type InterPoint interface {
	GetMeasurement() string
	GetTags() map[string]string
	GetFields() map[string]interface{}
	GetTime() time.Time
}

type InterTSDB interface {
	Save(points ...InterPoint) error
	Query(cmd string) (res []influx.Result, err error)
	Close() error

	IsDBExist() bool
	CreateDB() error
}

type DBConf struct {
	FirebaseConf *firebaseConf `yaml:"firebase"`
	InfluxDBConf *tsdbConf     `yaml:"influx"`

	db   InterDB
	auth InterAuth
	tsdb InterTSDB
}

func (dbc *DBConf) SetFirebase(file, url string) {
	dbc.FirebaseConf = &firebaseConf{
		CredentialsFile: file,
		DatabaseURL:     url,
	}
}

func (dbc *DBConf) SetInfluxDB(host, db string) {
	dbc.InfluxDBConf = &tsdbConf{
		Host: host,
		DB:   db,
	}
}

type firebaseConf struct {
	CredentialsFile string `yaml:"credentialsFile"`
	DatabaseURL     string `yaml:"databaseURL"`
}

type tsdbConf struct {
	Host string
	DB   string
}

func (dbc *DBConf) GetTSDB() InterTSDB {
	if dbc.tsdb != nil {
		return dbc.tsdb
	}

	dbc.tsdb = &influxDB{
		host: dbc.InfluxDBConf.Host,
		db:   dbc.InfluxDBConf.DB,
	}

	if !dbc.tsdb.IsDBExist() {
		dbc.tsdb.CreateDB()
	}

	return dbc.tsdb
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

func (dbc *DBConf) GetAuth() InterAuth {
	if dbc.auth != nil {
		return dbc.auth
	}
	dbc.auth = &firebaseDB{
		credentialsFile: dbc.FirebaseConf.CredentialsFile,
		ctx:             context.Background(),
	}
	return dbc.auth
}
