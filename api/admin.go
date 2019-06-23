package api

import (
	"fmt"
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
	}
}

func (api *AdminAPI) getCategoryEndpoint(w http.ResponseWriter, req *http.Request) {
	db := di.GetSQLDB()
	//db.Query("select * from public.ab")
	isDB := db.IsDBExist()

	w.Write([]byte(fmt.Sprintf("hi..%t", isDB)))
}
