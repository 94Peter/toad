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

type DeductAPI bool

func (api DeductAPI) Enable() bool {
	return bool(api)
}

type inputDeduct struct {
	ARid string `json:"arid"`
	//Status      string    `json:"status"`
	//Date        time.Time `json:"date"`
	Fee         int               `json:"fee"`
	Description string            `json:"description"`
	Item        string            `json:"item"`
	SalerList   []*model.MAPSaler `json:"salerList"`
}
type inputUpdateDeduct struct {
	// Status string    `json:"status"`
	// Date   time.Time `json:"date"`
	Date        string `json:"date"`
	Status      string `json:"status"`
	CheckNumber string `json:"checkNumber"` //票號
	Fee         int    `json:"fee"`
}

type inputUpdateDeductItem struct {
	Item string `json:"item"`
}

type inputUpdateSales struct {
	SalerList []*model.MAPSaler `json:"salerList"`
}

func (api DeductAPI) GetAPIs() *[]*APIHandler {
	return &[]*APIHandler{
		&APIHandler{Path: "/v1/deduct", Next: api.getDeductEndpoint, Method: "GET", Auth: true, Group: permission.All},
		//&APIHandler{Path: "/v1/deduct", Next: api.createDeductEndpoint, Method: "POST", Auth: true, Group: permission.All},
		&APIHandler{Path: "/v1/deduct/{ID}", Next: api.deleteDeductEndpoint, Method: "DELETE", Auth: true, Group: permission.All},
		&APIHandler{Path: "/v1/deduct/{ID}", Next: api.updateDeductEndpoint, Method: "PUT", Auth: true, Group: permission.All},
		&APIHandler{Path: "/v1/deductFee/{ID}", Next: api.updateDeductFeeEndpoint, Method: "PUT", Auth: true, Group: permission.All},

		&APIHandler{Path: "/v1/deduct/item/{ID}", Next: api.updateDeductItemEndpoint, Method: "PUT", Auth: true, Group: permission.All},
		&APIHandler{Path: "/v1/deduct/sales/{ID}", Next: api.updateDeductSalesEndpoint, Method: "PUT", Auth: true, Group: permission.All},
	}
}

func (api *DeductAPI) getDeductEndpoint(w http.ResponseWriter, req *http.Request) {
	dbname := req.Header.Get("dbname")
	dm := model.GetDecuctModel(di)
	// var queryDate time.Time
	// today := time.Date(queryDate.Year(), queryDate.Month(), 1, 0, 0, 0, 0, queryDate.Location())
	// end := time.Date(queryDate.Year(), queryDate.Month()+1, 1, 0, 0, 0, 0, queryDate.Location())

	queryVar := util.GetQueryValue(req, []string{"date", "type"}, true)
	by_m := (*queryVar)["date"].(string)
	ey_m := (*queryVar)["date"].(string)
	mtype := (*queryVar)["type"].(string)
	if by_m == "" {
		by_m = "1980-01-01T00:00:00.000Z"
		ey_m = "2200-12-31T00:00:00.000Z"
	}

	if mtype == "" || mtype == "全部" || strings.ToLower(mtype) == "all" {
		mtype = "%"
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

	//fmt.Println("by_m:", by_m)
	dm.GetDeductData(b, e, mtype, dbname)
	//data, err := json.Marshal(result)
	data, err := dm.Json()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (api *DeductAPI) deleteDeductEndpoint(w http.ResponseWriter, req *http.Request) {
	//Get params from body
	vars := util.GetPathVars(req, []string{"ID"})
	ID := vars["ID"].(string)
	dbname := req.Header.Get("dbname")
	DM := model.GetDecuctModel(di)

	err := DM.DeleteDeduct(ID, dbname)
	if err != nil {
		switch err.Error() {
		case ERROR_CloseDate:
			w.WriteHeader(http.StatusLocked)
			break
		default:

			w.WriteHeader(http.StatusInternalServerError)
			break
		}
		w.Write([]byte("Error,maybe id is not exist or status is not accepted" + err.Error()))
	} else {
		w.Write([]byte("OK"))
	}

}

func (api *DeductAPI) updateDeductEndpoint(w http.ResponseWriter, req *http.Request) {
	//Get params from body
	vars := util.GetPathVars(req, []string{"ID"})
	ID := vars["ID"].(string)
	dbname := req.Header.Get("dbname")
	iUD := inputUpdateDeduct{}
	err := json.NewDecoder(req.Body).Decode(&iUD)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid JSON format"))
		return
	}

	if ok, err := iUD.isUpdateDeductValid(); !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	DM := model.GetDecuctModel(di)

	err = DM.UpdateDeduct(ID, iUD.Status, iUD.Date, iUD.CheckNumber, dbname)
	if err != nil {
		switch err.Error() {
		case ERROR_CloseDate:
			w.WriteHeader(http.StatusLocked)
			break
		default:
			w.WriteHeader(http.StatusInternalServerError)
			break
		}
		w.Write([]byte("Error" + err.Error()))
	} else {
		w.Write([]byte("OK"))
	}

}

func (api *DeductAPI) updateDeductFeeEndpoint(w http.ResponseWriter, req *http.Request) {
	//Get params from body
	vars := util.GetPathVars(req, []string{"ID"})
	ID := vars["ID"].(string)
	dbname := req.Header.Get("dbname")
	iUD := inputUpdateDeduct{}
	err := json.NewDecoder(req.Body).Decode(&iUD)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid JSON format"))
		return
	}

	if iUD.Fee < 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Fee is not vaild"))
		return
	}

	DM := model.GetDecuctModel(di)

	_err := DM.UpdateDeductFee(ID, dbname, iUD.Fee)
	if _err != nil {
		switch err.Error() {
		case ERROR_CloseDate:
			w.WriteHeader(http.StatusLocked)
			break
		default:

			w.WriteHeader(http.StatusInternalServerError)
			break
		}
		w.Write([]byte("[Error] " + _err.Error()))
	} else {
		w.Write([]byte("OK"))
	}

}

func (api *DeductAPI) updateDeductItemEndpoint(w http.ResponseWriter, req *http.Request) {
	//Get params from body
	vars := util.GetPathVars(req, []string{"ID"})
	ID := vars["ID"].(string)
	dbname := req.Header.Get("dbname")
	iUD := inputUpdateDeductItem{}
	err := json.NewDecoder(req.Body).Decode(&iUD)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid JSON format"))
		return
	}

	DM := model.GetDecuctModel(di)

	err = DM.UpdateDeductItem(ID, iUD.Item, dbname)
	if err != nil {
		switch err.Error() {
		case ERROR_CloseDate:
			w.WriteHeader(http.StatusLocked)
			break
		default:
			w.WriteHeader(http.StatusInternalServerError)
			break
		}
		w.Write([]byte("Error" + err.Error()))
	} else {
		w.Write([]byte("OK"))
	}

}

func (api *DeductAPI) updateDeductSalesEndpoint(w http.ResponseWriter, req *http.Request) {
	//Get params from body
	vars := util.GetPathVars(req, []string{"ID"})
	ID := vars["ID"].(string)
	dbname := req.Header.Get("dbname")
	iUD := inputUpdateSales{}
	err := json.NewDecoder(req.Body).Decode(&iUD)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid JSON format"))
		return
	}

	DM := model.GetDecuctModel(di)

	err = DM.UpdateDeductSales(ID, dbname, iUD.SalerList)
	if err != nil {
		switch err.Error() {
		case ERROR_CloseDate:
			w.WriteHeader(http.StatusLocked)
			break
		default:
			w.WriteHeader(http.StatusInternalServerError)
			break
		}
		w.Write([]byte("Error" + err.Error()))
	} else {
		w.Write([]byte("OK"))
	}

}

func (iUD *inputUpdateDeduct) isUpdateDeductValid() (bool, error) {
	// if !util.IsStrInList(iAR.Permission, permission.All...) {
	// 	return false, errors.New("permission error")
	// }

	if iUD.Status == "未支付" {
		if !(iUD.Date == "") {
			return false, errors.New("date should be empty")
		}
	} else if iUD.Status == "支付" {
		if iUD.Date == "" {
			return false, errors.New("date is empty")
		}
		//https://blog.csdn.net/tianzhaixing2013/article/details/74625906
		//the_time, err := time.ParseInLocation("2006-01-02T15:04:05Z", iUD.Date, time.Local)
		// the_time, err := time.ParseInLocation("2006-01-02", iUD.Date, time.Local)
		// if err == nil {
		// 	unix_time := the_time.Unix()
		// 	// fmt.Println("方法二 时间戳：", unix_time, reflect.TypeOf(unix_time))
		// 	// fmt.Println("now：", time.Now().Unix())
		// 	if t := time.Now().Unix(); t < unix_time {
		// 		//未來的成交案 => 不成立
		// 		return false, errors.New("date is not valid, future date")
		// 	}
		// } else {
		// 	return false, errors.New("date is not valid, " + err.Error())
		// }
	} else if iUD.Status == "" {
		return false, errors.New("status is empty")
	} else {
		return false, errors.New("status is not vaild")
	}

	return true, nil
}

func (iDeduct *inputDeduct) isDeductValid() (bool, error) {

	if iDeduct.ARid == "" {
		return false, errors.New("arid is empty")
	}
	if iDeduct.Item == "" {
		return false, errors.New("item is empty")
	}

	if iDeduct.Fee < 0 {
		return false, errors.New("fee is not valid")
	}
	return true, nil
}

func (iDeduct *inputDeduct) GetDeduct() *model.Deduct {
	return &model.Deduct{
		ARid:        iDeduct.ARid,
		Item:        iDeduct.Item,
		Description: iDeduct.Description,
		Fee:         iDeduct.Fee,
		Sales:       iDeduct.SalerList,
	}
}

func (iDeduct *inputUpdateDeductItem) GetDeduct() *model.Deduct {
	return &model.Deduct{
		Item: iDeduct.Item,
	}
}

// func (iUD *inputUpdateDeduct) GetDeduct() *model.Deduct {
// 	return &model.Deduct{
// 		Status: iUD.Status,
// 	}
// }
