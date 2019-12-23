package api

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/94peter/pica/permission"
)

type AdminAPI bool

func (api AdminAPI) Enable() bool {
	return bool(api)
}

func (api AdminAPI) GetAPIs() *[]*APIHandler {
	return &[]*APIHandler{
		&APIHandler{Path: "/v1/category", Next: api.getCategoryEndpoint, Method: "GET", Auth: false, Group: permission.All},
		&APIHandler{Path: "/v1/category", Next: api.t, Method: "POST", Auth: false, Group: permission.All},
	}
}

var auth_token = "eyJhbGciOiJSUzI1NiIsImtpZCI6ImRldiIsInR5cCI6IkpXVCJ9.eyJleHAiOjEwNzkzOTAyMzIxLCJpYXQiOjE1NzA1MzAyODQsImlzcyI6InBpY2Fpc3MiLCJzeXMiOiJ0b2FkIn0.dCeCH2cYCm5MewP2lCpLGJV4ka4C8j4joHL23YlphRQJpOemKBRLReCXKFQh1GhdnFKXh6xh9ULox_BUBZxckdRDoJo5-R7fXM7eOy5hIRFyOwO8FOuKJ50QddR0qoLbuLbzIklJncxDRftBcujuOFFAFEBIkR5Nq9TyBEgIkSI"

type test_struct struct {
	Test string
}

func (api *AdminAPI) getCategoryEndpoint(w http.ResponseWriter, req *http.Request) {
	//db := di.GetSQLDB()
	//db.Query("select * from public.ab")
	//isDB := db.InitDB()

	client := &http.Client{}

	req, err := http.NewRequest("GET", "https://pica957.appspot.com/v1/toad/category", nil)
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
		log.Fatal(err)
		return
	}

	fmt.Println(string(sitemap))

	w.Write(sitemap)
}

func (api *AdminAPI) t(w http.ResponseWriter, req *http.Request) {
	//db := di.GetSQLDB()
	//db.Query("select * from public.ab")
	//isDB := db.InitDB()
	var t test_struct

	bodyBytes, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Fatal(err)
	}
	bodyString := string(bodyBytes)
	fmt.Println("body:" + bodyString)

	// err := json.NewDecoder(req.Body).Decode(&t)
	// if err != nil {
	// 	w.WriteHeader(http.StatusBadRequest)
	// 	w.Write([]byte("Invalid JSON format"))
	// 	return
	// }
	// fmt.Println("data:", t.Test)

	w.Write([]byte(fmt.Sprintf("hi..%s", t.Test)))
}

/*
DELETE FROM public.ar;
DELETE FROM public.armap;
DELETE FROM public.branchsalary;
DELETE FROM public.salersalary;
DELETE FROM public.incomeexpense;
DELETE FROM public.receipt;
DELETE FROM public.commission;
*/
