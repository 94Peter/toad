package api

import (
	"net/http"

	"github.com/94peter/pica/permission"
	"github.com/94peter/toad/model"
	"github.com/94peter/toad/util"
)

type SystemAPI bool

func (api SystemAPI) Enable() bool {
	return bool(api)
}

func (api SystemAPI) GetAPIs() *[]*APIHandler {
	return &[]*APIHandler{
		&APIHandler{Path: "/v1/system/account", Next: api.getAccountDataEndpoint, Method: "GET", Auth: false, Group: permission.All},
		&APIHandler{Path: "/v1/system/branch", Next: api.getBranchDataEndpoint, Method: "GET", Auth: false, Group: permission.All},
	}
}

func (api *SystemAPI) getAccountDataEndpoint(w http.ResponseWriter, req *http.Request) {

	systemM := model.GetSystemModel(di)
	queryVar := util.GetQueryValue(req, []string{"branch"}, true)
	branch := (*queryVar)["branch"].(string)
	// if branch == "" {
	// 	branch = "all"
	// }

	//data, err := json.Marshal(result)
	systemM.GetAccountData(branch)
	data, err := systemM.Json("account")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}
func (api *SystemAPI) getBranchDataEndpoint(w http.ResponseWriter, req *http.Request) {

	systemM := model.GetSystemModel(di)

	data, err := systemM.GetBranchData()
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
