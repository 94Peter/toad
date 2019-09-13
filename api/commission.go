package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/94peter/pica/permission"
	"github.com/94peter/toad/model"
	"github.com/94peter/toad/util"
)

type CommissionAPI bool

func (api CommissionAPI) Enable() bool {
	return bool(api)
}

type inputUpdateCommission struct {
	Percent float64 `json:"percent"`
}

func (api CommissionAPI) GetAPIs() *[]*APIHandler {
	return &[]*APIHandler{
		&APIHandler{Path: "/v1/commission", Next: api.getCommissionEndpoint, Method: "GET", Auth: false, Group: permission.All},
		&APIHandler{Path: "/v1/commission/{Rid}/{Sid}", Next: api.updateCommissionEndpoint, Method: "PUT", Auth: false, Group: permission.All},
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

func (api *CommissionAPI) updateCommissionEndpoint(w http.ResponseWriter, req *http.Request) {

	vars := util.GetPathVars(req, []string{"Rid"})
	vars2 := util.GetPathVars(req, []string{"Sid"})
	Rid := vars["Rid"].(string)
	Sid := vars2["Sid"].(string)
	fmt.Println("Rid" + Rid)
	fmt.Println("Sid" + Sid)
	iuC := inputUpdateCommission{}
	err := json.NewDecoder(req.Body).Decode(&iuC)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid JSON format"))
		return
	}

	cm := model.GetCModel(di)
	if err := cm.UpdateCommission(iuC.GetCommission(), Rid, Sid); err != nil {
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

func (iuC *inputUpdateCommission) GetCommission() *model.Commission {
	return &model.Commission{
		CPercent: iuC.Percent,
	}
}