package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/94peter/pica/permission"
	"github.com/94peter/toad/model"
)

type ConfigAPI bool

type inputAccountItem struct {
	ItemName string `json:"itemName"`
}

func (api ConfigAPI) Enable() bool {
	return bool(api)
}

func (api ConfigAPI) GetAPIs() *[]*APIHandler {
	return &[]*APIHandler{
		&APIHandler{Path: "/v1/config/item", Next: api.getAccountItemEndpoint, Method: "GET", Auth: false, Group: permission.All},
		&APIHandler{Path: "/v1/config/item", Next: api.createAccountItemEndpoint, Method: "POST", Auth: false, Group: permission.All},
	}
}

func (api *ConfigAPI) getAccountItemEndpoint(w http.ResponseWriter, req *http.Request) {

	aItemM := model.GetAccountItemModelModel(di)
	var queryDate time.Time
	today := time.Date(queryDate.Year(), queryDate.Month(), 1, 0, 0, 0, 0, queryDate.Location())
	end := time.Date(queryDate.Year(), queryDate.Month()+1, 1, 0, 0, 0, 0, queryDate.Location())

	aItemM.GetAccountItemData(today, end)
	//data, err := json.Marshal(result)
	data, err := aItemM.Json()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (api *ConfigAPI) createAccountItemEndpoint(w http.ResponseWriter, req *http.Request) {
	//Get params from body

	iAItem := inputAccountItem{}
	err := json.NewDecoder(req.Body).Decode(&iAItem)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid JSON format"))
		return
	}

	if ok, err := iAItem.isAItemValid(); !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	aItemM := model.GetAccountItemModelModel(di)

	_err := aItemM.CreateAccountItem(iAItem.GetAccountItem())
	if _err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error,maybe already exist"))
	} else {
		w.Write([]byte("OK"))
	}

}

func (iAItem *inputAccountItem) isAItemValid() (bool, error) {
	// if !util.IsStrInList(iAR.Permission, permission.All...) {
	// 	return false, errors.New("permission error")
	// }

	// if t := time.Now().Unix(); t <= iAR.Date.Unix() {
	// 	//未來的成交案 => 不成立
	// 	return false, errors.New("CompletionDate is not valid")
	// }
	if iAItem.ItemName == "" {
		return false, errors.New("itemName is empty")
	}

	return true, nil
}

func (iAItem *inputAccountItem) GetAccountItem() *model.AccountItem {
	return &model.AccountItem{
		ItemName: iAItem.ItemName,
	}
}
