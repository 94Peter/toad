package model

import (
	"fmt"
	"testing"
	"time"

	"github.com/94peter/pica/resource/db"
	"github.com/94peter/pica/resource/sms"
	"github.com/stretchr/testify/assert"
)

type TestAccount struct {
	Phone string
	Name  string
}

func (ta *TestAccount) GetID() string {
	return ta.Phone
}

func Test_AddCategory(t *testing.T) {
	dic := dictionaryCategory{}
	dic.Add("test1")
	dic.Add("test2")
	dic.Add("test3")
	j, _ := dic.Json()
	fmt.Println(string(j))
	dic.Remove("test2")
	j, _ = dic.Json()
	fmt.Println(string(j))
	assert.True(t, false)
}

func Test_SaveDictionary(t *testing.T) {
	dbc := db.DBConf{}
	dbc.SetFirebase("firebaseServiceKey.json", "https://pica957.firebaseio.com/")
	var err error
	fireDB := dbc.GetDB()
	dic := dictionaryCategory{db: fireDB}
	dic.Add("test1")
	dic.Add("test2")
	dic.Add("test3")
	err = dic.Save()
	fmt.Println(err)
	assert.True(t, false)
}

func Test_LoadDictionary(t *testing.T) {
	dbc := db.DBConf{}
	dbc.SetFirebase("firebaseServiceKey.json", "https://pica957.firebaseio.com/")
	var err error
	fireDB := dbc.GetDB()
	dicNew := dictionaryCategory{db: fireDB}
	err = dicNew.load()
	fmt.Println(err)
	j, _ := dicNew.Json()
	fmt.Println(string(j))
	dicNew.Add("test5")
	dicNew.Remove("test1")
	j, _ = dicNew.Json()
	fmt.Println(string(j))
	assert.True(t, false)
}

func Test_SaveCategoryUser(t *testing.T) {
	dbc := db.DBConf{}
	dbc.SetFirebase("firebaseServiceKey.json", "https://pica957.firebaseio.com/")

	fireDB := dbc.GetDB()
	ud := categoryUser{
		db: fireDB,
	}
	ud.add(&User{
		Account:    "peter",
		Category:   "A",
		Permission: "sales",
		Pwd:        "123",
	})
	j, _ := ud.json()
	fmt.Println(string(j))
	ud.save()

	assert.True(t, false)
}

func Test_LoadCategoryUser(t *testing.T) {
	dbc := &db.DBConf{}
	dbc.SetFirebase("firebaseServiceKey.json", "https://pica957.firebaseio.com/")

	fireDB := dbc.GetDB()
	ud := categoryUser{
		db: fireDB,
	}
	ud.load()
	j, _ := ud.json()
	fmt.Println(string(j))
	u := ud.get("peter")
	fmt.Println(u)
	ud.remove(u)
	j, _ = ud.json()
	fmt.Println(string(j))
	assert.True(t, false)
}

type testDI struct{}

func (di *testDI) GetDB() db.InterDB {
	dbc := &db.DBConf{}
	dbc.SetFirebase("firebaseServiceKey.json", "https://pica957.firebaseio.com/")
	return dbc.GetDB()
}
func (di *testDI) GetAuth() db.InterAuth {
	dbc := &db.DBConf{}
	dbc.SetFirebase("firebaseServiceKey.json", "https://pica957.firebaseio.com/")
	return dbc.GetAuth()
}

func (di *testDI) GetSMS() sms.InterSMS {
	return nil
}
func (di *testDI) GetLoginURL() string {
	return ""
}

func (di *testDI) GetTSDB() db.InterTSDB {
	// localhost:8086
	// host: influxdb.beta.hvac-cloud.org:8086
	// db: pica
	dbc := &db.DBConf{}
	dbc.SetInfluxDB("influxdb.beta.hvac-cloud.org:8086", "pica")
	return dbc.GetTSDB()
}

func (di *testDI) GetTrendItems() []string {
	return []string{"aa", "bb", "cc"}
}

func (di *testDI) GetLocation() *time.Location {
	return nil
}

func Test_CreateUser(t *testing.T) {
	di := testDI{}
	mm := GetMemberModel(&di)

	err := mm.CreateUser(&User{
		Account:    "0933919389",
		Name:       "domo admin",
		Email:      "jiuns0103@gmail.com",
		Phone:      "0933919389",
		Category:   "admin",
		Permission: "admin",
		Pwd:        "123456",
		State:      UserStateInit,
	})
	fmt.Println(err)
	assert.True(t, false)
}

func Test_TestVerifyToken(t *testing.T) {
	di := testDI{}
	mm := GetMemberModel(&di)

	u := mm.VerifyToken("eyJhbGciOiJSUzI1NiIsImtpZCI6IjY1NmMzZGQyMWQwZmVmODgyZTA5ZTBkODY5MWNhNWM3ZjJiMGQ2MjEiLCJ0eXAiOiJKV1QifQ.eyJuYW1lIjoiUGV0ZXIiLCJpc3MiOiJodHRwczovL3NlY3VyZXRva2VuLmdvb2dsZS5jb20vcGljYTk1NyIsImF1ZCI6InBpY2E5NTciLCJhdXRoX3RpbWUiOjE1NTU5OTQ5MzcsInVzZXJfaWQiOiIwOTE5OTY2NjY3Iiwic3ViIjoiMDkxOTk2NjY2NyIsImlhdCI6MTU1NTk5NDkzOSwiZXhwIjoxNTU1OTk4NTM5LCJlbWFpbCI6ImNoLmZvY2tlQGdtYWlsLmNvbSIsImVtYWlsX3ZlcmlmaWVkIjpmYWxzZSwicGhvbmVfbnVtYmVyIjoiKzg4NjkxOTk2NjY2NyIsImZpcmViYXNlIjp7ImlkZW50aXRpZXMiOnsicGhvbmUiOlsiKzg4NjkxOTk2NjY2NyJdLCJlbWFpbCI6WyJjaC5mb2NrZUBnbWFpbC5jb20iXX0sInNpZ25faW5fcHJvdmlkZXIiOiJwYXNzd29yZCJ9fQ.aW8jskWhitE0phHZ1w584N7AxNVnEDaAYlmVEmT4kuk5c2MxG5ZBBLAUTgpn8WifnPZ4PLdvwdvryFpU9iKksycQuHBePkL4FCBKpTghnsATlEvyBOznWkllJPNrPXpV3rsliYbXGL3vdpUYZUpveshhnuF9LEpNagR7RbkrVTzSlrGDkJ6dLpaK__9sEXbWbDMEvfYgIm1VM3EY7SKPoJHtjnmYPHl3Vkukv4E46WQ562fOf1TYGrYug-2TiMY4lbpW6HCSvBeLEzHWYV4DZpeQ0pWVBnIdRcnm8-zViXqEIrhku920dtibtTp2E1SlkHvsPQoJRyTZmLskWZBLJg")
	fmt.Println(*u)
	assert.True(t, false)
}

func Test_GetRandPwd(t *testing.T) {
	pwd := GetRandPwd(8)
	fmt.Println(pwd)
	assert.True(t, false)
}
