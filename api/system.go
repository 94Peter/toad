package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/94peter/pica/permission"
	"github.com/94peter/toad/model"
	"github.com/94peter/toad/util"
)

type SystemAPI bool

type inputSystemAccount struct {
	Account  string `json:"account"`
	Name     string `json:"name"`
	Auth     string `json:"auth"`
	Password string `json:"password"`
}

func (api SystemAPI) Enable() bool {
	return bool(api)
}

func (api SystemAPI) GetAPIs() *[]*APIHandler {
	return &[]*APIHandler{
		&APIHandler{Path: "/v1/system/account", Next: api.getAccountDataEndpoint, Method: "GET", Auth: false, Group: permission.All},
		&APIHandler{Path: "/v1/system/branch", Next: api.getBranchDataEndpoint, Method: "GET", Auth: false, Group: permission.All},

		&APIHandler{Path: "/v1/system/account", Next: api.createAccountDataEndpoint, Method: "POST", Auth: false, Group: permission.All},
	}
}

func (api *SystemAPI) createAccountDataEndpoint(w http.ResponseWriter, req *http.Request) {

	iSA := inputSystemAccount{}
	err := json.NewDecoder(req.Body).Decode(&iSA)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid JSON format"))
		return
	}

	if ok, err := iSA.isSystemAccountValid(); !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	systemM := model.GetSystemModel(di)

	_err := systemM.CreateSystemAccount(iSA.GetAccount())
	if _err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error,maybe already exist"))
	} else {
		w.Write([]byte("OK"))
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
	w.Header().Set("Content-Type", "application/plain")
	w.Write(data)
}

func (iSA *inputSystemAccount) isSystemAccountValid() (bool, error) {
	// if !util.IsStrInList(iAR.Permission, permission.All...) {
	// 	return false, errors.New("permission error")
	// }

	// if t := time.Now().Unix(); t <= iAR.Date.Unix() {
	// 	//未來的成交案 => 不成立
	// 	return false, errors.New("CompletionDate is not valid")
	// }
	if iSA.Account == "" {
		return false, errors.New("account is empty")
	}
	if iSA.Password == "" {
		return false, errors.New("password is empty")
	}
	if iSA.Auth == "" {
		return false, errors.New("auth is empty")
	}
	if iSA.Name == "" {
		return false, errors.New("Name is empty")
	}

	return true, nil
}

func (iSA *inputSystemAccount) GetAccount() *model.SystemAccount {
	return &model.SystemAccount{
		Account:  iSA.Account,
		Name:     iSA.Name,
		Password: iSA.Password,
		Auth:     iSA.Auth,
	}
}
