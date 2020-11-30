package api

import (
	"fmt"
	"net/http"
	"time"

	"toad/model"
	"toad/permission"
	"toad/util"
)

type IndexAPIV2 bool

func (api IndexAPIV2) Enable() bool {
	return bool(api)
}

func (api IndexAPIV2) GetAPIs() *[]*APIHandler {
	return &[]*APIHandler{
		&APIHandler{Path: "/v2/index/info", Next: api.getInfoDataEndpoint, Method: "GET", Auth: true, Group: permission.All},
	}
}

func (api *IndexAPIV2) getInfoDataEndpoint(w http.ResponseWriter, req *http.Request) {
	queryVar := util.GetQueryValue(req, []string{"date"}, true)
	by_m := (*queryVar)["date"].(string)
	dbname := req.Header.Get("dbname")
	date, err := time.Parse(time.RFC3339, by_m)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("date is not valid, %s", err.Error())))
		return
	}

	indexM := model.GetIndexModelV2(di)
	indexM.GetInfoData(date, dbname)
	data, err := indexM.Json("info")
	//data, err := json.Marshal(result)
	//data, err := systemM.Json("branch")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}
