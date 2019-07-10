package api

import (
	"net/http"
	"time"

	"github.com/94peter/pica/permission"
	"github.com/94peter/toad/model"
)

type CommissionAPI bool

func (api CommissionAPI) Enable() bool {
	return bool(api)
}

type inputUpdateCommission struct {
	//	Date   string `json:"date"`
	//	Amount int    `json:"amount"`
}

func (api CommissionAPI) GetAPIs() *[]*APIHandler {
	return &[]*APIHandler{
		&APIHandler{Path: "/v1/commission", Next: api.getCommissionEndpoint, Method: "GET", Auth: false, Group: permission.All},
	}
}

func (api *CommissionAPI) getCommissionEndpoint(w http.ResponseWriter, req *http.Request) {

	cm := model.GetCModel(di)
	var queryDate time.Time
	today := time.Date(queryDate.Year(), queryDate.Month(), 1, 0, 0, 0, 0, queryDate.Location())
	end := time.Date(queryDate.Year(), queryDate.Month()+1, 1, 0, 0, 0, 0, queryDate.Location())

	cm.GetCommissiontData(today, end)
	//data, err := json.Marshal(result)
	data, err := cm.Json()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}
