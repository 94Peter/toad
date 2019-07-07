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

type ReceiptAPI bool

func (api ReceiptAPI) Enable() bool {
	return bool(api)
}

type inputUpdateReceipt struct {
	Date   string `json:"date"`
	Amount int    `json:"amount"`
}

func (api ReceiptAPI) GetAPIs() *[]*APIHandler {
	return &[]*APIHandler{
		&APIHandler{Path: "/v1/receipt", Next: api.getReceiptEndpoint, Method: "GET", Auth: false, Group: permission.All},
		&APIHandler{Path: "/v1/receipt/{ID}", Next: api.deleteReceiptEndpoint, Method: "DELETE", Auth: false, Group: permission.All},
		&APIHandler{Path: "/v1/receipt/{ID}", Next: api.updateReceiptEndpoint, Method: "PUT", Auth: false, Group: permission.All},
		// &APIHandler{Path: "/v1/category", Next: api.createCategoryEndpoint, Method: "POST", Auth: true, Group: permission.Backend},
		// &APIHandler{Path: "/v1/category/{NAME}", Next: api.deleteCategoryEndpoint, Method: "DELETE", Auth: true, Group: permission.Backend},
		// &APIHandler{Path: "/v1/user", Next: api.createUserEndpoint, Method: "POST", Auth: true, Group: permission.Backend},
		// &APIHandler{Path: "/v1/user", Next: api.getUserEndpoint, Method: "GET", Auth: true, Group: permission.Backend},
		// &APIHandler{Path: "/v1/user/category", Next: api.updateUserCategoryEndpoint, Method: "PUT", Auth: true, Group: permission.Backend},
		// &APIHandler{Path: "/v1/user/permission", Next: api.updateUserPemissionEndpoint, Method: "PUT", Auth: true, Group: permission.Backend},
		// &APIHandler{Path: "/v1/user/{PHONE}", Next: api.deleteUserEndpoint, Method: "DELETE", Auth: true, Group: permission.Backend},
	}
}

func (api *ReceiptAPI) getReceiptEndpoint(w http.ResponseWriter, req *http.Request) {

	rm := model.GetRTModel(di)
	var queryDate time.Time
	today := time.Date(queryDate.Year(), queryDate.Month(), 1, 0, 0, 0, 0, queryDate.Location())
	end := time.Date(queryDate.Year(), queryDate.Month()+1, 1, 0, 0, 0, 0, queryDate.Location())

	rm.GetReceiptData(today, end)
	//data, err := json.Marshal(result)
	data, err := rm.Json()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}
func (api *ReceiptAPI) deleteReceiptEndpoint(w http.ResponseWriter, req *http.Request) {

	vars := util.GetPathVars(req, []string{"ID"})
	ID := vars["ID"].(string)
	fmt.Println(ID)
	rm := model.GetRTModel(di)
	if err := rm.DeleteReceiptData(ID); err != nil {
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
func (api *ReceiptAPI) updateReceiptEndpoint(w http.ResponseWriter, req *http.Request) {

	vars := util.GetPathVars(req, []string{"ID"})
	ID := vars["ID"].(string)
	fmt.Println(ID)
	iuRT := inputUpdateReceipt{}
	err := json.NewDecoder(req.Body).Decode(&iuRT)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid JSON format"))
		return
	}
	fmt.Println("iuRT.Amount", iuRT.Amount)
	fmt.Println("ID", ID)
	fmt.Println("iuRT.Datedate", iuRT.Date)

	rm := model.GetRTModel(di)
	if err := rm.UpdateReceiptData(iuRT.Amount, iuRT.Date, ID); err != nil {
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
