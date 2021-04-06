package api

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"toad/model"
	"toad/permission"
	"toad/util"
)

type ReturnsAPI bool

func (api ReturnsAPI) Enable() bool {
	return bool(api)
}

type createReturns struct {
	//Date        time.Time `json:"date"` //成交日期
	Description string `json:"description"`
	Amount      int    `json:"amount"`
	//Status      string    `json:"status"`
	Arid string `json:"arid"`
}

type updateReturns struct {
	//Date        time.Time `json:"date"` //成交日期
	Return_id   string
	Description string `json:"description"`
	Amount      int    `json:"amount"`
	//Status      string    `json:"status"`
	Arid       string                   `json:"arid"`
	Sales      []*model.ReturnMAPSaler  `json:"sales"`
	BranchList []*model.ReturnMAPBranch `json:"branchList"`
}

func (api ReturnsAPI) GetAPIs() *[]*APIHandler {
	return &[]*APIHandler{
		&APIHandler{Path: "/v1/returns", Next: api.createReturnsEndpoint, Method: "POST", Auth: true, Group: permission.All},
		&APIHandler{Path: "/v1/returns", Next: api.getReturnsEndpoint, Method: "GET", Auth: true, Group: permission.All},
		&APIHandler{Path: "/v1/returns/{ID}", Next: api.deleteReturnsEndpoint, Method: "DELETE", Auth: true, Group: permission.All},
		&APIHandler{Path: "/v1/returns/{ID}", Next: api.updateReturnsEndpoint, Method: "PUT", Auth: true, Group: permission.All},
	}
}

func (api *ReturnsAPI) createReturnsEndpoint(w http.ResponseWriter, req *http.Request) {
	dbname := req.Header.Get("dbname")
	cR := createReturns{}
	data, _ := ioutil.ReadAll(req.Body) //把  body 内容读入字符串

	err := json.Unmarshal([]byte(data), &cR)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid JSON format"))
		return
	}

	if ok, err := cR.isValid(); !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	model.GetDecuctModel(di)
	rm := model.GetReturnsModel(di)

	return_id, err := rm.CreateReturns(cR.GetReturns(), dbname)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error." + err.Error()))
	} else {
		if return_id == "" {
			w.Write([]byte("OK"))
			return
		}
		tmap := make(map[string]interface{})
		tmap["return_id"] = return_id
		out, err := json.MarshalIndent(tmap, "", "  ")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(out))
		}

	}

}
func (api *ReturnsAPI) getReturnsEndpoint(w http.ResponseWriter, req *http.Request) {
	dbname := req.Header.Get("dbname")
	rm := model.GetReturnsModel(di)

	// queryVar := util.GetQueryValue(req, []string{"key", "branch", "date", "status", "export"}, true)
	// key := (*queryVar)["key"].(string)
	// branch := (*queryVar)["branch"].(string)
	// status := (*queryVar)["status"].(string)
	// date := (*queryVar)["date"].(string)
	// if status == "" {
	// 	status = "0"
	// }
	// if branch == "all" || branch == "ALL" {
	// 	branch = ""
	// }
	// if date == "" {
	// 	date = "1980-12-31T00:00:00.000Z"
	// }
	// t, err := time.Parse(time.RFC3339, date)
	// if err != nil {
	// 	w.WriteHeader(http.StatusInternalServerError)
	// 	w.Write([]byte(fmt.Sprintf("date is not valid, %s", err.Error())))
	// }

	resultdata := rm.GetReturnsData("", dbname)
	//data, err := json.Marshal(result)
	data, err := json.Marshal(resultdata)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (api *ReturnsAPI) updateReturnsEndpoint(w http.ResponseWriter, req *http.Request) {
	dbname := req.Header.Get("dbname")
	uR := updateReturns{}

	vars := util.GetPathVars(req, []string{"ID"})
	ID := vars["ID"].(string)

	rm := model.GetReturnsModel(di)
	model.GetConfigModel(di)

	data, _ := ioutil.ReadAll(req.Body) //把  body 内容读入字符串

	err := json.Unmarshal([]byte(data), &uR)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid JSON format"))
		return
	}

	if ok, err := uR.isValid(); !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	uR.Return_id = ID
	if err := rm.UpdateReturns(uR.GetReturns(), dbname); err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(err.Error()))
		return
	}
	w.Write([]byte("ok"))
}

func (api *ReturnsAPI) deleteReturnsEndpoint(w http.ResponseWriter, req *http.Request) {
	dbname := req.Header.Get("dbname")

	vars := util.GetPathVars(req, []string{"ID"})
	ID := vars["ID"].(string)

	rm := model.GetReturnsModel(di)
	if err := rm.DeleteReturns(ID, dbname); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	w.Write([]byte("ok"))

}

func (cR *createReturns) isValid() (bool, error) {
	// if !util.IsStrInList(iAR.Permission, permission.All...) {
	// 	return false, errors.New("permission error")
	// }

	// if t := time.Now().Unix(); t <= iAR.Date.Unix() {
	// 	//未來的成交案 => 不成立
	// 	return false, errors.New("CompletionDate is not valid")
	// }

	if cR.Arid == "" {
		return false, errors.New("arid is empty")
	}

	return true, nil
}
func (uR *updateReturns) isValid() (bool, error) {

	for _, element := range uR.Sales {
		// if element.Percent < 0 {
		// 	return false, errors.New("proportion is not valid")
		// }
		if element.Sid == "" {
			return false, errors.New("account is empty")
		}
		if element.SName == "" {
			return false, errors.New("name is empty")
		}
		if element.Branch == "" {
			return false, errors.New("branch is empty")
		}

		if element.SR < 0 {
			return false, errors.New("SR is not valid")
		}

	}
	if len(uR.Sales) == 0 {
		return false, errors.New("Sales is empty")
	}
	for _, element := range uR.BranchList {

		if element.Branch == "" {
			return false, errors.New("branch is empty")
		}

		if element.SR < 0 {
			return false, errors.New("SR is not valid")
		}

	}
	return true, nil
}

func (cR *createReturns) GetReturns() *model.Returns {
	return &model.Returns{
		Amount: cR.Amount,
		//Date:        cR.Date,
		Description: cR.Description,
		Arid:        cR.Arid,
	}
}

func (uR *updateReturns) GetReturns() *model.Returns {
	return &model.Returns{
		Return_id: uR.Return_id,
		Amount:    uR.Amount,
		//Date:        cR.Date,
		Description: uR.Description,
		Sales:       uR.Sales,
		BranchList:  uR.BranchList,
	}
}
