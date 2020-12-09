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

type PocketAPI bool

type inputPocket struct {
	Branch string    `json:"branch"`
	Date   time.Time `json:"date"`
	//Date        time.Time `json:"date"`
	ItemName    string `json:"itemName"`
	Description string `json:"description"`
	Income      int    `json:"income"`
	Fee         int    `json:"fee"`
}

func (api PocketAPI) Enable() bool {
	return bool(api)
}

func (api PocketAPI) GetAPIs() *[]*APIHandler {
	return &[]*APIHandler{
		&APIHandler{Path: "/v1/pocket", Next: api.getPocketEndpoint, Method: "GET", Auth: true, Group: permission.All},
		&APIHandler{Path: "/v1/pocket", Next: api.createPocketEndpoint, Method: "POST", Auth: true, Group: permission.All},
		&APIHandler{Path: "/v1/pocket/{ID}", Next: api.updatePocketEndpoint, Method: "PUT", Auth: true, Group: permission.All},
		&APIHandler{Path: "/v1/pocket/{ID}", Next: api.deletePocketEndpoint, Method: "DELETE", Auth: true, Group: permission.All},

		&APIHandler{Path: "/v1/pocket/export", Next: api.exportPocketEndpoint, Method: "GET", Auth: true, Group: permission.All},
	}
}

func (api *PocketAPI) exportPocketEndpoint(w http.ResponseWriter, req *http.Request) {

	PocketM := model.GetPocketModel(di)
	//var queryDate time.Time

	queryVar := util.GetQueryValue(req, []string{"date", "branch"}, true)
	by_m := (*queryVar)["date"].(string)
	ey_m := by_m
	branch := (*queryVar)["branch"].(string)

	if by_m == "" {
		by_m = "1980-01-01T00:00:00.000Z"
		ey_m = "2200-12-31T00:00:00.000Z"
	}

	b, err := time.Parse(time.RFC3339, by_m)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("date is not valid, %s", err.Error())))
		return
	}
	m, err := time.Parse(time.RFC3339, ey_m)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("date is not valid, %s", err.Error())))
		return
	}

	if branch == "" || branch == "全部" || strings.ToLower(branch) == "all" {
		branch = "%"
	}

	// today := time.Date(queryDate.Year(), queryDate.Month(), 1, 0, 0, 0, 0, queryDate.Location())
	// end := time.Date(queryDate.Year(), queryDate.Month()+1, 1, 0, 0, 0, 0, queryDate.Location())

	PocketM.GetPocketData(b, m, branch)
	w.Write(PocketM.PDF())
}

func (api *PocketAPI) deletePocketEndpoint(w http.ResponseWriter, req *http.Request) {
	dbname := req.Header.Get("dbname")
	vars := util.GetPathVars(req, []string{"ID"})
	ID := vars["ID"].(string)
	fmt.Println(ID)
	PocketM := model.GetPocketModel(di)
	if err := PocketM.DeletePocket(ID, dbname); err != nil {
		if strings.Contains(err.Error(), ERROR_CloseDate) {
			w.WriteHeader(http.StatusLocked)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
		w.Write([]byte(err.Error()))
		return
	}

	w.Write([]byte("ok"))
	return
}

func (api *PocketAPI) getPocketEndpoint(w http.ResponseWriter, req *http.Request) {

	PocketM := model.GetPocketModel(di)
	//var queryDate time.Time

	queryVar := util.GetQueryValue(req, []string{"date", "branch"}, true)
	by_m := (*queryVar)["date"].(string)
	ey_m := by_m
	branch := (*queryVar)["branch"].(string)

	if by_m == "" {
		by_m = "1980-01-01T00:00:00.000Z"
		ey_m = "2200-12-31T00:00:00.000Z"
	}

	b, err := time.Parse(time.RFC3339, by_m)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("date is not valid, %s", err.Error())))
	}
	m, err := time.Parse(time.RFC3339, ey_m)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("date is not valid, %s", err.Error())))
	}

	if branch == "" || branch == "全部" || strings.ToLower(branch) == "all" {
		branch = "%"
	}

	// today := time.Date(queryDate.Year(), queryDate.Month(), 1, 0, 0, 0, 0, queryDate.Location())
	// end := time.Date(queryDate.Year(), queryDate.Month()+1, 1, 0, 0, 0, 0, queryDate.Location())

	PocketM.GetPocketData(b, m, branch)
	//data, err := json.Marshal(result)
	data, err := PocketM.Json()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (api *PocketAPI) updatePocketEndpoint(w http.ResponseWriter, req *http.Request) {
	//Get params from body
	vars := util.GetPathVars(req, []string{"ID"})
	ID := vars["ID"].(string)
	dbname := req.Header.Get("dbname")
	iPocket := inputPocket{}
	err := json.NewDecoder(req.Body).Decode(&iPocket)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid JSON format"))
		return
	}

	if ok, err := iPocket.isPocketValid(); !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	PocketM := model.GetPocketModel(di)

	err = PocketM.UpdatePocket(ID, dbname, iPocket.GetPocket())
	if err != nil {
		if strings.Contains(err.Error(), ERROR_CloseDate) {
			w.WriteHeader(http.StatusLocked)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		w.Write([]byte(err.Error()))
	} else {
		w.Write([]byte("OK"))
	}

}

func (api *PocketAPI) createPocketEndpoint(w http.ResponseWriter, req *http.Request) {
	//Get params from body
	dbname := req.Header.Get("dbname")
	iPocket := inputPocket{}
	err := json.NewDecoder(req.Body).Decode(&iPocket)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid JSON format"))
		return
	}

	if ok, err := iPocket.isPocketValid(); !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	PocketM := model.GetPocketModel(di)

	err = PocketM.CreatePocket(iPocket.GetPocket(), dbname)
	if err != nil {
		if strings.Contains(err.Error(), ERROR_CloseDate) {
			w.WriteHeader(http.StatusLocked)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		w.Write([]byte("Error:" + err.Error()))
	} else {
		w.Write([]byte("OK"))
	}

}

func (iPocket *inputPocket) isPocketValid() (bool, error) {
	// if !util.IsStrInList(iAR.Permission, permission.All...) {
	// 	return false, errors.New("permission error")
	// }

	// if t := time.Now().Unix(); t <= iAR.Date.Unix() {
	// 	//未來的成交案 => 不成立
	// 	return false, errors.New("CompletionDate is not valid")
	// }
	if iPocket.Branch == "" {
		return false, errors.New("branch is empty")
	}
	if iPocket.ItemName == "" {
		return false, errors.New("itemName is empty")
	}

	if iPocket.Fee < 0 {
		return false, errors.New("fee is not valid")
	}

	if iPocket.Income < 0 {
		return false, errors.New("income is not valid")
	}

	return true, nil
}

func (iPocket *inputPocket) GetPocket() *model.Pocket {
	//GCP local time zone是+0時區，circleID月份強制改+8時間
	loc, _ := time.LoadLocation("Asia/Taipei")
	t := iPocket.Date.In(loc)
	circleid := fmt.Sprintf("%d-%02d", t.Year(), t.Month())
	fmt.Println("circleID:", circleid)
	return &model.Pocket{
		Branch:      iPocket.Branch,
		ItemName:    iPocket.ItemName,
		Description: iPocket.Description,
		Fee:         iPocket.Fee,
		Income:      iPocket.Income,
		Date:        iPocket.Date,
		CircleID:    circleid,
	}
}
