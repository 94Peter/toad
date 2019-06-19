package db

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type TestAccount struct {
	Phone string
	Name  string
}

func (ta *TestAccount) GetID() string {
	return ta.Phone
}
func Test_Save(t *testing.T) {
	dbc := DBConf{
		FirebaseConf: &firebaseConf{
			CredentialsFile: "firebaseServiceKey.json",
			DatabaseURL:     "https://pica957.firebaseio.com/",
		},
	}
	var err error
	fireDB := dbc.GetDB()
	err = fireDB.C("account").Save(&TestAccount{
		Phone: "0919966667",
		Name:  "Peter2",
	})
	if err != nil {
		fmt.Println(err.Error())
	}

	nta := TestAccount{}
	err = fireDB.GetByID("0919966667", &nta)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(nta.Name)

	assert.True(t, false)
}

func Test_CreateUser(t *testing.T) {
	dbc := DBConf{
		FirebaseConf: &firebaseConf{
			CredentialsFile: "firebaseServiceKey.json",
			DatabaseURL:     "https://pica957.firebaseio.com/",
		},
	}
	var err error
	fireAuth := dbc.GetAuth()
	err = fireAuth.CreateUser("0919966667", "peter", "ch.focke@gmail.com", "password", "admin")
	fmt.Println(err)
	assert.True(t, false)
}

func Test_SetUserDisable(t *testing.T) {
	dbc := DBConf{
		FirebaseConf: &firebaseConf{
			CredentialsFile: "firebaseServiceKey.json",
			DatabaseURL:     "https://pica957.firebaseio.com/",
		},
	}
	var err error
	fireAuth := dbc.GetAuth()
	err = fireAuth.SetUserDisable("0919966667", false)
	fmt.Println(err)
	assert.True(t, false)
}

func Test_VerifyToken(t *testing.T) {
	dbc := DBConf{
		FirebaseConf: &firebaseConf{
			CredentialsFile: "firebaseServiceKey.json",
			DatabaseURL:     "https://pica957.firebaseio.com/",
		},
	}
	fireAuth := dbc.GetAuth()
	token := "eyJhbGciOiJSUzI1NiIsImtpZCI6IjY1NmMzZGQyMWQwZmVmODgyZTA5ZTBkODY5MWNhNWM3ZjJiMGQ2MjEiLCJ0eXAiOiJKV1QifQ.eyJuYW1lIjoiUGV0ZXIiLCJwZXJtaXNzaW9uIjoic2FsZXMiLCJzdGF0ZSI6ImluaXQiLCJpc3MiOiJodHRwczovL3NlY3VyZXRva2VuLmdvb2dsZS5jb20vcGljYTk1NyIsImF1ZCI6InBpY2E5NTciLCJhdXRoX3RpbWUiOjE1NTU5OTI4NTcsInVzZXJfaWQiOiIwOTE5OTY2NjY3Iiwic3ViIjoiMDkxOTk2NjY2NyIsImlhdCI6MTU1NTk5Mjg1OCwiZXhwIjoxNTU1OTk2NDU4LCJlbWFpbCI6ImNoLmZvY2tlQGdtYWlsLmNvbSIsImVtYWlsX3ZlcmlmaWVkIjpmYWxzZSwicGhvbmVfbnVtYmVyIjoiKzg4NjkxOTk2NjY2NyIsImZpcmViYXNlIjp7ImlkZW50aXRpZXMiOnsicGhvbmUiOlsiKzg4NjkxOTk2NjY2NyJdLCJlbWFpbCI6WyJjaC5mb2NrZUBnbWFpbC5jb20iXX0sInNpZ25faW5fcHJvdmlkZXIiOiJwYXNzd29yZCJ9fQ.FHldRVi92AxgNiqB8nKsdFvIQhryGxJzIU68pBqepROcRGACFqkFc4UyshY7UcZ02S-ZIjTWn7SObRbXJ23lgs0wTUAnopO6cBtQqOhZrGGsxLmTRIBXssC97adOKDMzDEkM1QFPOKmHxZQXHUusUo_JYVNvN2HXxbuSOZMwnW7cEzQm647RKDF-Zxa1u23M668aQY6atUJR5ZdRRKl76SmHf7_-bW2rjQ1n1uOx7TsAcTsyI1onqaOLZ6Fxwl0w8OUNGHIg1kK7NJiFtYksocqwL9z3mbkVsNkQy8bdJ5VopIX7qwJ75RLwak7AnJqSRiEaR6o4IqUs3A3_WE6GIQ"
	uid, err := fireAuth.VerifyToken(token)
	fmt.Println(uid, err)
	assert.True(t, false)
}

func Test_TimeSub(t *testing.T) {
	day1 := "2019-04-20T00:00:00+08:00"
	day2 := "2019-04-23T00:00:00+08:00"

	time1, _ := time.Parse(time.RFC3339, day1)
	time2, _ := time.Parse(time.RFC3339, day2)
	d := time2.Sub(time1)
	fmt.Println(d.Hours())
	assert.True(t, false)
}
