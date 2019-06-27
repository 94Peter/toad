package model

import (
	"errors"

	"github.com/94peter/toad/resource/db"
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
}

type memberModel struct {
	cu *categoryUser
}

type categoryUser struct {
	db db.InterSQLDB
}

func GetMemberModel(mr interModelRes) *memberModel {
	cu := &categoryUser{
		db: mr.GetSQLDB(),
	}
	cu.load()

	return &memberModel{

		cu: cu,
	}
}

func (dc *categoryUser) load() error {
	if dc.db == nil {
		return errors.New("db not set")
	}
	return nil
}
