package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/94peter/toad/model"
	"github.com/94peter/toad/permission"
)

type ARAPI bool

func (api ARAPI) Enable() bool {
	return bool(api)
}

func (api ARAPI) GetAPIs() *[]*APIHandler {
	return &[]*APIHandler{
		&APIHandler{Path: "/v1/receivable", Next: api.getAccountReceivableEndpoint, Method: "GET", Auth: false, Group: permission.All},
		&APIHandler{Path: "/v1/receivable", Next: api.createAccountReceivableEndpoint, Method: "POST", Auth: false, Group: permission.All},
		&APIHandler{Path: "/v1/receipt", Next: api.createReceiptEndpoint, Method: "POST", Auth: false, Group: permission.All},

		// &APIHandler{Path: "/v1/category", Next: api.createCategoryEndpoint, Method: "POST", Auth: true, Group: permission.Backend},
		// &APIHandler{Path: "/v1/category/{NAME}", Next: api.deleteCategoryEndpoint, Method: "DELETE", Auth: true, Group: permission.Backend},
		// &APIHandler{Path: "/v1/user", Next: api.createUserEndpoint, Method: "POST", Auth: true, Group: permission.Backend},
		// &APIHandler{Path: "/v1/user", Next: api.getUserEndpoint, Method: "GET", Auth: true, Group: permission.Backend},
		// &APIHandler{Path: "/v1/user/category", Next: api.updateUserCategoryEndpoint, Method: "PUT", Auth: true, Group: permission.Backend},
		// &APIHandler{Path: "/v1/user/permission", Next: api.updateUserPemissionEndpoint, Method: "PUT", Auth: true, Group: permission.Backend},
		// &APIHandler{Path: "/v1/user/{PHONE}", Next: api.deleteUserEndpoint, Method: "DELETE", Auth: true, Group: permission.Backend},
	}
}

func (api *ARAPI) getAccountReceivableEndpoint(w http.ResponseWriter, req *http.Request) {

	am := model.GetARModel(di)
	var queryDate time.Time
	today := time.Date(queryDate.Year(), queryDate.Month(), 1, 0, 0, 0, 0, queryDate.Location())
	end := time.Date(queryDate.Year(), queryDate.Month()+1, 1, 0, 0, 0, 0, queryDate.Location())

	am.GetARData(today, end)
	//data, err := json.Marshal(result)
	data, err := am.Json()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (api *ARAPI) createAccountReceivableEndpoint(w http.ResponseWriter, req *http.Request) {
	//Get params from body
	ar := model.AR{}
	err := json.NewDecoder(req.Body).Decode(&ar)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid JSON format"))
		return
	}

	am := model.GetARModel(di)
	_err := am.CreateAccountReceivable(&ar)
	if _err != nil {
		w.Write([]byte("Error"))
	} else {
		w.Write([]byte("OK"))
	}

}

func (api *ARAPI) createReceiptEndpoint(w http.ResponseWriter, req *http.Request) {
	//Get params from body
	rt := model.Receipt{}
	err := json.NewDecoder(req.Body).Decode(&rt)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid JSON format"))
		return
	}

	am := model.GetARModel(di)
	_err := am.CreateReceipt(&rt)
	if _err != nil {
		w.Write([]byte("Error"))
	} else {
		w.Write([]byte("OK"))
	}

}
