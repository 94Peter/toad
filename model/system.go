package model

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/94peter/toad/resource/db"
)

type SystemAccount struct {
	Account string `json:"id"`
	Name    string `json:"name"`
	//Branch  string `json:"branch"`
	Email string `json:"email"`
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

func (systemM *SystemModel) GetAccountData(branch string) ([]byte, error) {
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

func (systemM *SystemModel) GetBranchData() ([]byte, error) {
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
