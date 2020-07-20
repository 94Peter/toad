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

type CommissionAPI bool

func (api CommissionAPI) Enable() bool {
	return bool(api)
}

type inputUpdateCommission struct {
	Percent float64 `json:"percent"`
}

type exportCommission struct {
	CList []struct {
		Sid string `json:"sid"`
		Rid string `json:"rid"`
	} `json:"commissionList"`
}

func (api CommissionAPI) GetAPIs() *[]*APIHandler {
	return &[]*APIHandler{
		&APIHandler{Path: "/v1/commission", Next: api.getCommissionEndpoint, Method: "GET", Auth: false, Group: permission.All},
		//&APIHandler{Path: "/v1/commission/export", Next: api.exportCommissionEndpoint, Method: "POST", Auth: true, Group: permission.All},
		&APIHandler{Path: "/v1/commission/{Rid}/{Sid}", Next: api.updateCommissionEndpoint, Method: "PUT", Auth: true, Group: permission.All},
		&APIHandler{Path: "/v1/commission/status/{Rid}/{Sid}", Next: api.updateCommissionStatusEndpoint, Method: "PUT", Auth: true, Group: permission.All},
		//更新Bonus使用
		&APIHandler{Path: "/v1/commission/bonus/{Rid}/{Sid}", Next: api.refreshCommissionBonusEndpoint, Method: "PUT", Auth: true, Group: permission.All},
	}
}

func (api *CommissionAPI) getCommissionEndpoint(w http.ResponseWriter, req *http.Request) {

	cm := model.GetCModel(di)

	queryVar := util.GetQueryValue(req, []string{"date", "status", "export"}, true)
	//export := (*queryVar)["export"].(string)
	status := (*queryVar)["status"].(string)
	by_m := (*queryVar)["date"].(string)
	ey_m := by_m

	if status == "" || status == "all" {
		status = "%"
	}

	if by_m == "" {
		by_m = "1980-01"
		ey_m = "2200-01"
	}
	b, err := time.ParseInLocation("2006-01-02", by_m+"-01", time.Local)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("date is not valid, %s", err.Error())))
		return
	}
	e, err := time.ParseInLocation("2006-01-02", ey_m+"-01", time.Local)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("date is not valid, %s", err.Error())))
		return
	}

	//cm.GetCommissiontData(by_m+"-01", ey_m+"-01", status)
	cm.GetCommissiontData(b, e, status)
	//data, err := json.Marshal(result)
	//if export == "" {
	data, err := cm.Json()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
	// } else {
	// 	data := cm.PDF()
	// 	//w.Header().Set("Content-Type", "application/json")
	// 	w.Write(data)
	// }
}

// func (api *CommissionAPI) exportCommissionEndpoint(w http.ResponseWriter, req *http.Request) {

// 	cm := model.GetCModel(di)

// 	exportC := exportCommission{}
// 	err := json.NewDecoder(req.Body).Decode(&exportC)
// 	if err != nil {
// 		w.WriteHeader(http.StatusBadRequest)
// 		w.Write([]byte("Invalid JSON format"))
// 		return
// 	}

// 	// data, err := json.Marshal(exportC)
// 	// fmt.Println(string(data))

// 	cm.ExportCommissiontData(exportC.GetCommission())
// 	w.Write(cm.PDF())

// }

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

	if ok, err := iuC.isCommissionValid(); !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
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

func (api *CommissionAPI) refreshCommissionBonusEndpoint(w http.ResponseWriter, req *http.Request) {

	queryVar := util.GetQueryValue(req, []string{"type"}, true)
	vars := util.GetPathVars(req, []string{"Rid", "Sid"})
	Rid := vars["Rid"].(string)
	Sid := vars["Sid"].(string)
	mtype := (*queryVar)["type"].(string)

	fmt.Println("Rid" + Rid + " Sid" + Sid + " type " + mtype)

	cm := model.GetCModel(di)
	if err := cm.RefreshCommissionBonus(Sid, Rid, mtype); err != nil {
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

func (api *CommissionAPI) updateCommissionStatusEndpoint(w http.ResponseWriter, req *http.Request) {

	vars := util.GetPathVars(req, []string{"Rid"})
	vars2 := util.GetPathVars(req, []string{"Sid"})
	Rid := vars["Rid"].(string)
	Sid := vars2["Sid"].(string)
	fmt.Println("Rid" + Rid)
	fmt.Println("Sid" + Sid)

	cm := model.GetCModel(di)
	if err := cm.UpdateCommissionStatus(Rid, Sid); err != nil {
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
func (exportC *exportCommission) GetCommission() []*model.Commission {
	data := []*model.Commission{}
	for _, element := range exportC.CList {
		c := &model.Commission{
			Sid: element.Sid,
			Rid: element.Rid,
		}
		data = append(data, c)
	}

	return data
}

func (iuC *inputUpdateCommission) isCommissionValid() (bool, error) {

	// if iuC.Status == "" {
	// 	return false, errors.New("Status is empty")
	// }

	// if !(iuC.Status == "remove" || iuC.Status == "normal") {
	// 	return false, errors.New("status should be remove or normal")
	// }

	if iuC.Percent < 0 || iuC.Percent > 100 {
		return false, errors.New("percent is not valid")
	}

	// _, err := time.ParseInLocation("2006-01-02", iBS.Date, time.Local)
	// if err != nil {
	// 	return false, errors.New("date is not valid, " + err.Error())
	// }

	return true, nil
}
