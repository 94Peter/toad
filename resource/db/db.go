package db

import (
	"context"
)

type interDoc interface {
	GetID() string
}

type InterDB interface {
	C(c string) InterDB
	Save(doc interDoc) error
	GetByID(id string, doc interface{}) error
}

type DBConf struct {
	FirebaseConf *firebaseConf `yaml:"firebase"`

	db InterDB
}

func (dbc *DBConf) SetFirebase(file, url string) {
	dbc.FirebaseConf = &firebaseConf{
		CredentialsFile: file,
		DatabaseURL:     url,
	}
}

type firebaseConf struct {
	CredentialsFile string `yaml:"credentialsFile"`
	DatabaseURL     string `yaml:"databaseURL"`
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
