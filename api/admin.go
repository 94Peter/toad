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

type test_struct struct {
	Test string
}

func (api *AdminAPI) getCategoryEndpoint(w http.ResponseWriter, req *http.Request) {
	//db := di.GetSQLDB()
	//db.Query("select * from public.ab")
	//isDB := db.InitDB()

	w.Write([]byte(fmt.Sprintf("hi..")))
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
