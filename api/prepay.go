package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/94peter/pica/permission"
	"github.com/94peter/toad/model"
	"github.com/94peter/toad/util"
)

type PrePayAPI bool

type inputPrePay struct {
	Date        time.Time             `json:"date"`
	ItemName    string                `json:"itemName"`
	Description string                `json:"description"`
	Fee         int                   `json:"fee"`
	PrePay      []*model.BranchPrePay `json:"prepay"`
}

type inputBranchPrePay struct {
	Branch string `json:"branch"`
	Cost   int    `json:"cost"`
}

func (api PrePayAPI) Enable() bool {
	return bool(api)
}

func (api PrePayAPI) GetAPIs() *[]*APIHandler {
	return &[]*APIHandler{
		&APIHandler{Path: "/v1/prepay", Next: api.getPrePayEndpoint, Method: "GET", Auth: false, Group: permission.All},
		&APIHandler{Path: "/v1/prepay", Next: api.createPrePayEndpoint, Method: "POST", Auth: false, Group: permission.All},
		&APIHandler{Path: "/v1/prepay/{ID}", Next: api.deletePrePayEndpoint, Method: "DELETE", Auth: false, Group: permission.All},
		&APIHandler{Path: "/v1/prepay/{ID}", Next: api.updatePrePayEndpoint, Method: "PUT", Auth: false, Group: permission.All},
	}
}

func (api *PrePayAPI) deletePrePayEndpoint(w http.ResponseWriter, req *http.Request) {

	vars := util.GetPathVars(req, []string{"ID"})
	ID := vars["ID"].(string)
	fmt.Println(ID)
	PrePayM := model.GetPrePayModel(di)
	if err := PrePayM.DeletePrePay(ID); err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(err.Error()))
		return
	}
	// if err := memberModel.Quit(phone); err != nil {
	// 	w.WriteHeader(http.StatusInternalServerError)
	// 	w.Write([]byte(err.Error()))
	// 	return
	// }
	w.Write([]byte("ok"))
	return
}

func (api *PrePayAPI) getPrePayEndpoint(w http.ResponseWriter, req *http.Request) {

	PrePayM := model.GetPrePayModel(di)
	var queryDate time.Time
	today := time.Date(queryDate.Year(), queryDate.Month(), 1, 0, 0, 0, 0, queryDate.Location())
	end := time.Date(queryDate.Year(), queryDate.Month()+1, 1, 0, 0, 0, 0, queryDate.Location())

	PrePayM.GetPrePayData(today, end)
	//data, err := json.Marshal(result)
	data, err := PrePayM.Json()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (api *PrePayAPI) createPrePayEndpoint(w http.ResponseWriter, req *http.Request) {
	//Get params from body

	iPrePay := inputPrePay{}
	err := json.NewDecoder(req.Body).Decode(&iPrePay)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid JSON format"))
		return
	}

	if ok, err := iPrePay.isPrePayValid(); !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	PrePayM := model.GetPrePayModel(di)

	_err := PrePayM.CreatePrePay(iPrePay.GetPrePay())
	if _err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error"))
	} else {
		w.Write([]byte("OK"))
	}

}

func (api *PrePayAPI) updatePrePayEndpoint(w http.ResponseWriter, req *http.Request) {
	//Get params from body
	vars := util.GetPathVars(req, []string{"ID"})
	ID := vars["ID"].(string)
	fmt.Println(ID)

	iPrePay := inputPrePay{}
	err := json.NewDecoder(req.Body).Decode(&iPrePay)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid JSON format"))
		return
	}

	if ok, err := iPrePay.isPrePayValid(); !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	PrePayM := model.GetPrePayModel(di)

	_err := PrePayM.UpdatePrePay(ID, iPrePay.GetPrePay())
	if _err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error"))
	} else {
		w.Write([]byte("OK"))
	}

}

func (iPrePay *inputPrePay) isPrePayValid() (bool, error) {
	// if !util.IsStrInList(iAR.Permission, permission.All...) {
	// 	return false, errors.New("permission error")
	// }

	// if t := time.Now().Unix(); t <= iAR.Date.Unix() {
	// 	//未來的成交案 => 不成立
	// 	return false, errors.New("CompletionDate is not valid")
	// }
	if iPrePay.Description == "" {
		return false, errors.New("description is empty")
	}
	if iPrePay.ItemName == "" {
		return false, errors.New("itemName is empty")
	}

	if iPrePay.Fee < 0 {
		return false, errors.New("fee is not valid")
	}

	if len(iPrePay.PrePay) <= 0 {
		return false, errors.New("prepay is not valid")
	}

	_, err := json.Marshal(iPrePay)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (iPrePay *inputPrePay) GetPrePay() *model.PrePay {
	return &model.PrePay{
		Date:        iPrePay.Date,
		ItemName:    iPrePay.ItemName,
		Description: iPrePay.Description,
		Fee:         iPrePay.Fee,
		PrePay:      iPrePay.PrePay,
	}
}
