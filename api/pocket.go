package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/94peter/pica/permission"
	"github.com/94peter/toad/model"
	"github.com/94peter/toad/util"
)

type PocketAPI bool

type inputPocket struct {
	Branch string `json:"branch"`
	//Date        time.Time `json:"date"`
	Date        string `json:"date"`
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
		&APIHandler{Path: "/v1/pocket", Next: api.getPocketEndpoint, Method: "GET", Auth: false, Group: permission.All},
		&APIHandler{Path: "/v1/pocket", Next: api.createPocketEndpoint, Method: "POST", Auth: false, Group: permission.All},
		&APIHandler{Path: "/v1/pocket/{ID}", Next: api.updatePocketEndpoint, Method: "PUT", Auth: false, Group: permission.All},
		&APIHandler{Path: "/v1/pocket/{ID}", Next: api.deletePocketEndpoint, Method: "DELETE", Auth: false, Group: permission.All},
	}
}

func (api *PocketAPI) deletePocketEndpoint(w http.ResponseWriter, req *http.Request) {

	vars := util.GetPathVars(req, []string{"ID"})
	ID := vars["ID"].(string)
	fmt.Println(ID)
	PocketM := model.GetPocketModel(di)
	if err := PocketM.DeletePocket(ID); err != nil {
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

func (api *PocketAPI) getPocketEndpoint(w http.ResponseWriter, req *http.Request) {

	PocketM := model.GetPocketModel(di)
	//var queryDate time.Time

	queryVar := util.GetQueryValue(req, []string{"date", "branch"}, true)
	by_m := (*queryVar)["date"].(string)
	ey_m := by_m
	branch := (*queryVar)["branch"].(string)
	if by_m == "" {
		by_m = "2000-01"
		ey_m = "2200-01"
	}
	if branch == "" || branch == "全部" || strings.ToLower(branch) == "all" {
		branch = "%"
	}

	// today := time.Date(queryDate.Year(), queryDate.Month(), 1, 0, 0, 0, 0, queryDate.Location())
	// end := time.Date(queryDate.Year(), queryDate.Month()+1, 1, 0, 0, 0, 0, queryDate.Location())

	PocketM.GetPocketData(by_m, ey_m, branch)
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

	_err := PocketM.UpdatePocket(ID, iPocket.GetPocket())
	if _err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(_err.Error()))
	} else {
		w.Write([]byte("OK"))
	}

}

func (api *PocketAPI) createPocketEndpoint(w http.ResponseWriter, req *http.Request) {
	//Get params from body

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

	_err := PocketM.CreatePocket(iPocket.GetPocket())
	if _err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error"))
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

	_, err := time.ParseInLocation("2006-01-02", iPocket.Date, time.Local)
	if err != nil {
		return false, errors.New("date is not valid, " + err.Error())
	}

	return true, nil
}

func (iPocket *inputPocket) GetPocket() *model.Pocket {
	the_time, _ := time.ParseInLocation("2006-01-02", iPocket.Date, time.Local)
	return &model.Pocket{
		Branch:      iPocket.Branch,
		ItemName:    iPocket.ItemName,
		Description: iPocket.Description,
		Fee:         iPocket.Fee,
		Income:      iPocket.Income,
		Date:        the_time,
		CircleID:    iPocket.Date[:7],
	}
}
