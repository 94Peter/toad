package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/94peter/pica/permission"
	"github.com/94peter/toad/model"
)

type AmortizationAPI bool

type inputAmortization struct {
	Branch                    string `json:"branch"`
	ItemName                  string `json:"itemName"`
	GainCost                  int    `json:"gainCost"`
	AmortizationYearLimit     int    `json:"amortizationYearLimit"`
	MonthlyAmortizationAmount int    `json:"monthlyAmortizationAmount"`
	FirstAmortizationAmount   int    `json:"firstAmortizationAmount"`
}

func (api AmortizationAPI) Enable() bool {
	return bool(api)
}

func (api AmortizationAPI) GetAPIs() *[]*APIHandler {
	return &[]*APIHandler{
		&APIHandler{Path: "/v1/amortization", Next: api.getAmortizationEndpoint, Method: "GET", Auth: false, Group: permission.All},
		&APIHandler{Path: "/v1/amortization", Next: api.createAmortizationEndpoint, Method: "POST", Auth: false, Group: permission.All},
	}
}

func (api *AmortizationAPI) getAmortizationEndpoint(w http.ResponseWriter, req *http.Request) {

	amorM := model.GetAmortizationModel(di)
	var queryDate time.Time
	today := time.Date(queryDate.Year(), queryDate.Month(), 1, 0, 0, 0, 0, queryDate.Location())
	end := time.Date(queryDate.Year(), queryDate.Month()+1, 1, 0, 0, 0, 0, queryDate.Location())

	amorM.GetAmortizationData(today, end)
	//data, err := json.Marshal(result)
	data, err := amorM.Json()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (api *AmortizationAPI) createAmortizationEndpoint(w http.ResponseWriter, req *http.Request) {
	//Get params from body

	iAmor := inputAmortization{}
	err := json.NewDecoder(req.Body).Decode(&iAmor)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid JSON format"))
		return
	}

	if ok, err := iAmor.isAmorValid(); !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	amor := model.GetAmortizationModel(di)

	_err := amor.CreateAmortization(iAmor.GetAmortization())
	if _err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error"))
	} else {
		w.Write([]byte("OK"))
	}

}

func (iAmor *inputAmortization) isAmorValid() (bool, error) {
	// if !util.IsStrInList(iAR.Permission, permission.All...) {
	// 	return false, errors.New("permission error")
	// }

	// if t := time.Now().Unix(); t <= iAR.Date.Unix() {
	// 	//未來的成交案 => 不成立
	// 	return false, errors.New("CompletionDate is not valid")
	// }
	if iAmor.Branch == "" {
		return false, errors.New("Branch is empty")
	}
	if iAmor.AmortizationYearLimit < 0 {
		return false, errors.New("AmortizationYearLimit is not valid")
	}
	if iAmor.FirstAmortizationAmount < 0 {
		return false, errors.New("FirstAmortizationAmount is not valid")
	}
	if iAmor.GainCost < 0 {
		return false, errors.New("GainCost is not valid")
	}
	if iAmor.ItemName == "" {
		return false, errors.New("Branch is empty")
	}
	if iAmor.MonthlyAmortizationAmount < 0 {
		return false, errors.New("MonthlyAmortizationAmount is not valid")
	}

	return true, nil
}

func (iAmor *inputAmortization) GetAmortization() *model.Amortization {
	return &model.Amortization{
		Branch:                    iAmor.Branch,
		Itemname:                  iAmor.ItemName,
		Gaincost:                  iAmor.GainCost,
		AmortizationYearLimit:     iAmor.AmortizationYearLimit,
		MonthlyAmortizationAmount: iAmor.MonthlyAmortizationAmount,
		FirstAmortizationAmount:   iAmor.FirstAmortizationAmount,
	}
}
