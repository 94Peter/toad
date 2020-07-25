package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"toad/model"
	"toad/permission"
	"toad/util"
)

type AmortizationAPI bool

type inputAmortization struct {
	Date                  string `json:"date"`
	Branch                string `json:"branch"`
	ItemName              string `json:"itemName"`
	GainCost              int    `json:"gainCost"`
	AmortizationYearLimit int    `json:"amortizationYearLimit"`
	//MonthlyAmortizationAmount int    `json:"monthlyAmortizationAmount"`
	//FirstAmortizationAmount   int    `json:"firstAmortizationAmount"`
}

func (api AmortizationAPI) Enable() bool {
	return bool(api)
}

func (api AmortizationAPI) GetAPIs() *[]*APIHandler {
	return &[]*APIHandler{
		&APIHandler{Path: "/v1/amortization", Next: api.getAmortizationEndpoint, Method: "GET", Auth: false, Group: permission.All},
		&APIHandler{Path: "/v1/amortization", Next: api.createAmortizationEndpoint, Method: "POST", Auth: false, Group: permission.All},
		&APIHandler{Path: "/v1/amortization/{ID}", Next: api.deleteAmortizationEndpoint, Method: "DELETE", Auth: false, Group: permission.All},

		&APIHandler{Path: "/v1/amortization/export", Next: api.exportAmortizationEndpoint, Method: "GET", Auth: false, Group: permission.All},
	}
}

func (api *AmortizationAPI) exportAmortizationEndpoint(w http.ResponseWriter, req *http.Request) {
	//Get params from body
	//vars := util.GetPathVars(req, []string{"ID"})
	//amorID := vars["ID"].(string)
	amor := model.GetAmortizationModel(di)

	queryVar := util.GetQueryValue(req, []string{"begin", "end", "branch"}, true)
	by_m := (*queryVar)["begin"].(string)
	ey_m := (*queryVar)["end"].(string)
	branch := (*queryVar)["branch"].(string)
	if by_m == "" {
		by_m = "1980-01-01T00:00:00.000Z"
	}
	if ey_m == "" {
		ey_m = "2200-12-31T00:00:00.000Z"
	}
	if branch == "" || branch == "全部" || strings.ToLower(branch) == "all" {
		branch = "%"
	}
	b, err := time.Parse(time.RFC3339, by_m)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("date is not valid, %s", err.Error())))
	}

	e, err := time.Parse(time.RFC3339, ey_m)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("date is not valid, %s", err.Error())))
	}

	amor.GetAmortizationData(b, e, branch)

	w.Write(amor.PDF())

}

func (api *AmortizationAPI) deleteAmortizationEndpoint(w http.ResponseWriter, req *http.Request) {

	vars := util.GetPathVars(req, []string{"ID"})
	ID := vars["ID"].(string)

	amorM := model.GetAmortizationModel(di)
	if err := amorM.DeleteAmortization(ID); err != nil {
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
}

func (api *AmortizationAPI) getAmortizationEndpoint(w http.ResponseWriter, req *http.Request) {

	amorM := model.GetAmortizationModel(di)
	// var queryDate time.Time
	// today := time.Date(queryDate.Year(), queryDate.Month(), 1, 0, 0, 0, 0, queryDate.Location())
	// end := time.Date(queryDate.Year(), queryDate.Month()+1, 1, 0, 0, 0, 0, queryDate.Location())
	queryVar := util.GetQueryValue(req, []string{"begin", "end", "branch"}, true)
	by_m := (*queryVar)["begin"].(string)
	ey_m := (*queryVar)["end"].(string)
	branch := (*queryVar)["branch"].(string)
	if by_m == "" {
		by_m = "1980-01-01T00:00:00.000Z"
	}
	if ey_m == "" {
		ey_m = "2200-12-31T00:00:00.000Z"
	}

	if branch == "" || branch == "全部" || strings.ToLower(branch) == "all" {
		branch = "%"
	}

	b, err := time.Parse(time.RFC3339, by_m)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("date is not valid, %s", err.Error())))
	}

	e, err := time.Parse(time.RFC3339, ey_m)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("date is not valid, %s", err.Error())))
	}

	amorM.GetAmortizationData(b, e, branch)
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

	_, err := time.Parse(time.RFC3339, iAmor.Date)
	if err != nil {
		return false, errors.New("date is not valid, " + err.Error())
	}

	if iAmor.Branch == "" {
		return false, errors.New("Branch is empty")
	}
	if iAmor.AmortizationYearLimit < 0 {
		return false, errors.New("AmortizationYearLimit is not valid")
	}
	// if iAmor.FirstAmortizationAmount < 0 {
	// 	return false, errors.New("FirstAmortizationAmount is not valid")
	// }
	if iAmor.GainCost < 0 {
		return false, errors.New("GainCost is not valid")
	}
	if iAmor.ItemName == "" {
		return false, errors.New("Branch is empty")
	}
	// if iAmor.MonthlyAmortizationAmount < 0 {
	// 	return false, errors.New("MonthlyAmortizationAmount is not valid")
	// }

	return true, nil
}

func (iAmor *inputAmortization) GetAmortization() *model.Amortization {
	date, _ := time.Parse(time.RFC3339, iAmor.Date)
	var MonthlyAmortizationAmount = iAmor.GainCost / (iAmor.AmortizationYearLimit * 12)
	var FirstAmortizationAmount = MonthlyAmortizationAmount + iAmor.GainCost%(iAmor.AmortizationYearLimit*12)
	return &model.Amortization{
		Branch:                    iAmor.Branch,
		Itemname:                  iAmor.ItemName,
		Gaincost:                  iAmor.GainCost,
		Date:                      date,
		AmortizationYearLimit:     iAmor.AmortizationYearLimit,
		MonthlyAmortizationAmount: MonthlyAmortizationAmount,
		FirstAmortizationAmount:   FirstAmortizationAmount,
	}
}
