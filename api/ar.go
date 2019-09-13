package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/94peter/toad/model"
	"github.com/94peter/toad/permission"
	"github.com/94peter/toad/util"
)

type ARAPI bool

func (api ARAPI) Enable() bool {
	return bool(api)
}

type inputAR struct {
	//ARid     string    `json:"id"`
	Date     time.Time `json:"completionDate"` //成交日期
	CNo      string    `json:"contractNo"`
	Customer struct {
		Action string `json:"type"`
		Name   string `json:"name"`
	} `json:"customer"`
	CaseName string `json:"caseName"`
	Amount   int    `json:"amount"`
	//Fee      int            `json:"fee"`
	Sales []*model.Saler `json:"sales"`
}
type inputReceipt struct {
	//Rid           string    `json:"-"` //no return this key
	ARid string    `json:"id"`
	Date time.Time `json:"date"`
	// CNo           string    `json:"contractNo"`
	// CaseName      string    `json:"caseName"`
	// CustomertType string    `json:"customertType"`
	// Name          string    `json:"customerName"`
	Amount int `json:"amount"` //收款
	//InvoiceNo     string    `json:"invoiceNo"` //發票號碼
}

func (api ARAPI) GetAPIs() *[]*APIHandler {
	return &[]*APIHandler{
		&APIHandler{Path: "/v1/receivable", Next: api.getAccountReceivableEndpoint, Method: "GET", Auth: false, Group: permission.All},
		&APIHandler{Path: "/v1/receivable", Next: api.createAccountReceivableEndpoint, Method: "POST", Auth: false, Group: permission.All},
		&APIHandler{Path: "/v1/receivable/{ID}", Next: api.deleteAccountReceivableEndpoint, Method: "DELETE", Auth: false, Group: permission.All},
		&APIHandler{Path: "/v1/receipt", Next: api.createReceiptEndpoint, Method: "POST", Auth: false, Group: permission.All},
		&APIHandler{Path: "/v1/deduct", Next: api.createDeductEndpoint, Method: "POST", Auth: false, Group: permission.All},
		// &APIHandler{Path: "/v1/category", Next: api.createCategoryEndpoint, Method: "POST", Auth: true, Group: permission.Backend},
		// &APIHandler{Path: "/v1/category/{NAME}", Next: api.deleteCategoryEndpoint, Method: "DELETE", Auth: true, Group: permission.Backend},
		// &APIHandler{Path: "/v1/user", Next: api.createUserEndpoint, Method: "POST", Auth: true, Group: permission.Backend},
		// &APIHandler{Path: "/v1/user", Next: api.getUserEndpoint, Method: "GET", Auth: true, Group: permission.Backend},
		// &APIHandler{Path: "/v1/user/category", Next: api.updateUserCategoryEndpoint, Method: "PUT", Auth: true, Group: permission.Backend},
		// &APIHandler{Path: "/v1/user/permission", Next: api.updateUserPemissionEndpoint, Method: "PUT", Auth: true, Group: permission.Backend},
		// &APIHandler{Path: "/v1/user/{PHONE}", Next: api.deleteUserEndpoint, Method: "DELETE", Auth: true, Group: permission.Backend},
	}
}

func (api *ARAPI) deleteAccountReceivableEndpoint(w http.ResponseWriter, req *http.Request) {

	vars := util.GetPathVars(req, []string{"ID"})
	ID := vars["ID"].(string)

	am := model.GetARModel(di)
	if err := am.DeleteAccountReceivable(ID); err != nil {
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

func (api *ARAPI) getAccountReceivableEndpoint(w http.ResponseWriter, req *http.Request) {

	am := model.GetARModel(di)
	var queryDate time.Time
	today := time.Date(queryDate.Year(), queryDate.Month(), 1, 0, 0, 0, 0, queryDate.Location())
	end := time.Date(queryDate.Year(), queryDate.Month()+1, 1, 0, 0, 0, 0, queryDate.Location())

	queryVar := util.GetQueryValue(req, []string{"key", "export"}, true)
	key := (*queryVar)["key"].(string)

	am.GetARData(today, end, key)
	//data, err := json.Marshal(result)
	data, err := am.Json()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (api *ARAPI) createAccountReceivableEndpoint(w http.ResponseWriter, req *http.Request) {
	//Get params from body

	iAR := inputAR{}
	err := json.NewDecoder(req.Body).Decode(&iAR)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid JSON format"))
		return
	}

	if ok, err := iAR.isARValid(); !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	am := model.GetARModel(di)

	_err := am.CreateAccountReceivable(iAR.GetAR())
	if _err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error"))
	} else {
		w.Write([]byte("OK"))
	}

}

func (api *ARAPI) createDeductEndpoint(w http.ResponseWriter, req *http.Request) {
	//Get params from body

	iDeduct := inputDeduct{}
	err := json.NewDecoder(req.Body).Decode(&iDeduct)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid JSON format"))
		return
	}

	if ok, err := iDeduct.isDeductValid(); !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	DM := model.GetDecuctModel(di)

	_err := DM.CreateDeduct(iDeduct.GetDeduct())
	if _err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error"))
	} else {
		w.Write([]byte("OK"))
	}

}

func (api *ARAPI) createReceiptEndpoint(w http.ResponseWriter, req *http.Request) {
	//Get params from body
	irt := inputReceipt{}

	err := json.NewDecoder(req.Body).Decode(&irt)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid JSON format"))
		return
	}

	if ok, err := irt.isReceiptValid(); !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	rm := model.GetRTModel(di)
	_ = model.GetCModel(di)      //init Commission Model for create commission
	_ = model.GetDecuctModel(di) //init Deduct Model for update DeductRid
	_err := rm.CreateReceipt(irt.GetReceipt())
	if _err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(_err.Error()))
	} else {
		w.Write([]byte("OK"))
	}

}

func (iAR *inputAR) isARValid() (bool, error) {
	// if !util.IsStrInList(iAR.Permission, permission.All...) {
	// 	return false, errors.New("permission error")
	// }
	fmt.Println("iAR.Date:", iAR.Date)

	if t := time.Now().Unix(); t <= iAR.Date.Unix() {
		//未來的成交案 => 不成立
		return false, errors.New("CompletionDate is not valid")
	}
	if iAR.CNo == "" {
		return false, errors.New("ContractNo is empty")
	}
	if iAR.Customer.Action == "" {
		return false, errors.New("Customer type is empty")
	} else {
		if !(iAR.Customer.Action == "sell" || iAR.Customer.Action == "buy") {
			return false, errors.New("Customer type should be 'sell' or 'buy'")
		}
	}
	if iAR.Customer.Name == "" {
		return false, errors.New("Customer name is empty")
	}
	if iAR.CaseName == "" {
		return false, errors.New("Case name is empty")
	}
	//有0元的成交案例嗎?
	if iAR.Amount < 0 {
		return false, errors.New("Amount is not valid")
	}
	// if iAR.Fee < 0 || iAR.Fee > iAR.Amount {
	// 	return false, errors.New("Fee is not valid")
	// }
	for _, element := range iAR.Sales {
		if element.Percent < 0 {
			return false, errors.New("Percent is not valid")
		}
		if element.Sid == "" {
			return false, errors.New("account is empty")
		}
		if element.SName == "" {
			return false, errors.New("name is empty")
		}
	}
	if len(iAR.Sales) == 0 {
		return false, errors.New("Sales is empty")
	}

	return true, nil
}

func (irt *inputReceipt) isReceiptValid() (bool, error) {
	if irt.ARid == "" {
		return false, errors.New("id is empty")
	}
	//收0元成立嗎?
	if irt.Amount < 1 {
		return false, errors.New("Amount is not valid")
	}

	if t := time.Now().Unix(); t <= irt.Date.Unix() {
		//未來的成交案 => 不成立
		return false, errors.New("Date is not valid")
	}

	return true, nil
}

func (iAR *inputAR) GetAR() *model.AR {
	return &model.AR{
		Amount:   iAR.Amount,
		Date:     iAR.Date,
		CNo:      iAR.CNo,
		CaseName: iAR.CaseName,
		Customer: iAR.Customer,
		//Fee:      iAR.Fee,
		Sales: iAR.Sales,
	}
}

func (irt *inputReceipt) GetReceipt() *model.Receipt {
	return &model.Receipt{
		Amount: irt.Amount,
		Date:   irt.Date,
		ARid:   irt.ARid,
	}
}
