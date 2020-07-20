package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"toad/model"
	"toad/permission"
	"toad/util"
)

type SystemAPI bool

type inputSystemAccount struct {
	Account     string `json:"account"`
	Name        string `json:"name"`
	Auth        string `json:"auth"`
	Password    string `json:"password"`
	NewPassword string `json:"new_password"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
}

var UPDATE_PASSWORD = true

func (api SystemAPI) Enable() bool {
	return bool(api)
}

func (api SystemAPI) GetAPIs() *[]*APIHandler {
	return &[]*APIHandler{
		&APIHandler{Path: "/v1/system/branch", Next: api.getBranchDataEndpoint, Method: "GET", Auth: true, Group: permission.All},

		&APIHandler{Path: "/v1/system/account", Next: api.getAccountDataEndpoint, Method: "GET", Auth: true, Group: permission.All},
		//&APIHandler{Path: "/v1/system/account", Next: api.updateAccountDataEndpoint, Method: "PUT", Auth: true, Group: permission.All},
		//&APIHandler{Path: "/v1/system/account/password", Next: api.updateAccountPasswordEndpoint, Method: "PUT", Auth: true, Group: permission.All},
		//&APIHandler{Path: "/v1/system/account/{account}", Next: api.deleteAccountDataEndpoint, Method: "DELETE", Auth: true, Group: permission.All},
		//&APIHandler{Path: "/v1/system/account", Next: api.createAccountDataEndpoint, Method: "POST", Auth: true, Group: permission.All},
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

func (api *SystemAPI) deleteAccountDataEndpoint(w http.ResponseWriter, req *http.Request) {
	//Get params from body
	vars := util.GetPathVars(req, []string{"account"})
	account := vars["account"].(string)

	systemM := model.GetSystemModel(di)

	_err := systemM.DeleteSystemAccount(account)
	if _err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error,maybe is not exist"))
	} else {
		w.Write([]byte("OK"))
	}

}

func (api *SystemAPI) updateAccountDataEndpoint(w http.ResponseWriter, req *http.Request) {

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

	_err := systemM.UpdateSystemAccount(iSA.GetAccount())
	if _err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Error,maybe password wrong"))
	} else {
		w.Write([]byte("OK"))
	}
}

func (api *SystemAPI) updateAccountPasswordEndpoint(w http.ResponseWriter, req *http.Request) {

	iSA := inputSystemAccount{}
	err := json.NewDecoder(req.Body).Decode(&iSA)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid JSON format"))
		return
	}

	if ok, err := iSA.isSystemAccountValid(UPDATE_PASSWORD); !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	systemM := model.GetSystemModel(di)

	_err := systemM.UpdateSystemAccountPassword(iSA.NewPassword, iSA.GetAccount())
	if _err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Error,maybe old password wrong"))
	} else {
		w.Write([]byte("OK"))
	}
}

func (api *SystemAPI) getAccountDataEndpoint(w http.ResponseWriter, req *http.Request) {

	systemM := model.GetSystemModel(di)
	// queryVar := util.GetQueryValue(req, []string{"branch"}, true)
	// branch := (*queryVar)["branch"].(string)
	// if branch == "" {
	// 	branch = "all"
	// }

	//data, err := json.Marshal(result)
	systemM.GetAccountData()
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

func (iSA *inputSystemAccount) isSystemAccountValid(things ...interface{}) (bool, error) {
	// if !util.IsStrInList(iAR.Permission, permission.All...) {
	// 	return false, errors.New("permission error")
	// }

	// if t := time.Now().Unix(); t <= iAR.Date.Unix() {
	// 	//未來的成交案 => 不成立
	// 	return false, errors.New("CompletionDate is not valid")
	// }
	var pwd = false
	for _, params := range things {
		pwd = params.(bool)
	}

	if iSA.Account == "" {
		return false, errors.New("account is empty")
	}
	if iSA.Password == "" {
		return false, errors.New("password is empty")
	}
	if pwd {
		if iSA.NewPassword == "" {
			return false, errors.New("new_password is empty")
		}
	} else {

		if iSA.Auth == "" {
			return false, errors.New("auth is empty")
		}
		if iSA.Name == "" {
			return false, errors.New("name is empty")
		}
		if iSA.Email == "" {
			return false, errors.New("email is empty")
		}

	}

	return true, nil
}

func (iSA *inputSystemAccount) GetAccount() *model.SystemAccount {
	return &model.SystemAccount{
		Account:  iSA.Account,
		Name:     iSA.Name,
		Password: iSA.Password,
		Auth:     iSA.Auth,
		Email:    iSA.Email,
		Phone:    iSA.Phone,
	}
}
