package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"toad/model"
	"toad/permission"
	"toad/util"
)

type IndexAPI bool

func (api IndexAPI) Enable() bool {
	return bool(api)
}

func (api IndexAPI) GetAPIs() *[]*APIHandler {
	return &[]*APIHandler{
		&APIHandler{Path: "/v1/index/info", Next: api.getInfoDataEndpoint, Method: "GET", Auth: false, Group: permission.All},
		&APIHandler{Path: "/v1/index/incomeStatement", Next: api.getBranchDataEndpoint, Method: "GET", Auth: false, Group: permission.All},
	}
}

func (api *IndexAPI) getInfoDataEndpoint(w http.ResponseWriter, req *http.Request) {
	queryVar := util.GetQueryValue(req, []string{"date"}, true)
	by_m := (*queryVar)["date"].(string)
	date, err := time.Parse(time.RFC3339, by_m)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("date is not valid, %s", err.Error())))
		return
	}

	indexM := model.GetIndexModel(di)
	indexM.GetInfoData(date)
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
func (api *IndexAPI) getBranchDataEndpoint(w http.ResponseWriter, req *http.Request) {

	queryVar := util.GetQueryValue(req, []string{"branch", "date"}, true)

	by_m := (*queryVar)["date"].(string)

	date, err := time.Parse(time.RFC3339, by_m)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("date is not valid, %s", err.Error())))
		return
	}

	branch := (*queryVar)["branch"].(string)
	if branch == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("require branch param"))
		return
	}

	indexM := model.GetIndexModel(di)
	data, err := indexM.GetIncomeStatement(branch, date)
	//data, err := indexM.Json("incomeStatement")
	//data, err := json.Marshal(result)
	//data, err := systemM.Json("branch")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	result, _ := json.Marshal(data)
	w.Header().Set("Content-Type", "application/json")
	w.Write(result)
}
