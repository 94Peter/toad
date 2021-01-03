package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"toad/model"
	"toad/permission"
	"toad/util"
)

type ARAPI bool

var ACTION_BUY = "buy"
var ACTION_SELL = "sell"

func (api ARAPI) Enable() bool {
	return bool(api)
}

type inputAR struct {
	//ARid     string    `json:"id"`
	Date     time.Time      `json:"completionDate"` //成交日期
	CNo      string         `json:"contractNo"`
	Customer model.Customer `json:"customer"`
	CaseName string         `json:"caseName"`
	Amount   int            `json:"amount"`
	//Fee      int            `json:"fee"`
	Sales []*model.MAPSaler `json:"sales"`
}

type inputUpdateAR struct {
	Amount    int               `json:"amount"`
	SalerList []*model.MAPSaler `json:"salerList"`
}

type inputhouseGoAR struct {
	ARid     int       `json:"id"`
	Date     time.Time `json:"completionDate"` //成交日期
	CNo      string    `json:"contractNo"`
	CaseName string    `json:"caseName"`

	Sell struct {
		Amount int    `json:"amount"`
		Name   string `json:"name"`
	} `json:"sell"`

	Buyer struct {
		Amount int    `json:"amount"`
		Name   string `json:"name"`
	} `json:"buyer"`
	Sales []*model.HouseGoMAPSaler `json:"sales"`
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
	Fee    int `json:"fee"`    //扣款金額
	//InvoiceNo     string    `json:"invoiceNo"` //發票號碼
}

func (api ARAPI) GetAPIs() *[]*APIHandler {
	return &[]*APIHandler{
		&APIHandler{Path: "/v1/housego", Next: api.CreateHouseGoEndpoint, Method: "POST", Auth: true, Group: permission.All},
		&APIHandler{Path: "/v1/housego", Next: api.getHouseGoEndpoint, Method: "GET", Auth: true, Group: permission.All},
		&APIHandler{Path: "/v1/housego/{ID}", Next: api.UpgradeARInfoWithHouseGoEndpoint, Method: "PUT", Auth: true, Group: permission.All},

		&APIHandler{Path: "/v1/receivable", Next: api.getAccountReceivableEndpoint, Method: "GET", Auth: true, Group: permission.All},
		&APIHandler{Path: "/v1/receivable", Next: api.createAccountReceivableEndpoint, Method: "POST", Auth: true, Group: permission.All},
		&APIHandler{Path: "/v1/receivable/{ID}", Next: api.deleteAccountReceivableEndpoint, Method: "DELETE", Auth: true, Group: permission.All},
		&APIHandler{Path: "/v1/receivable/{ID}", Next: api.updateAccountReceivableEndpoint, Method: "PUT", Auth: true, Group: permission.All},

		&APIHandler{Path: "/v1/receipt", Next: api.createReceiptEndpoint, Method: "POST", Auth: true, Group: permission.All},
		&APIHandler{Path: "/v1/deduct", Next: api.createDeductEndpoint, Method: "POST", Auth: true, Group: permission.All},

		&APIHandler{Path: "/v1/receivable/saler", Next: api.getSalerDataEndpoint, Method: "GET", Auth: true, Group: permission.All},
		//&APIHandler{Path: "/v1/receivable/recount", Next: api.getReCountEndpoint, Method: "GET", Auth: true, Group: permission.All},

		// &APIHandler{Path: "/v1/category", Next: api.createCategoryEndpoint, Method: "POST", Auth: true, Group: permission.Backend},
		// &APIHandler{Path: "/v1/category/{NAME}", Next: api.deleteCategoryEndpoint, Method: "DELETE", Auth: true, Group: permission.Backend},

		// &APIHandler{Path: "/v1/user", Next: api.getUserEndpoint, Method: "GET", Auth: true, Group: permission.Backend},
		// &APIHandler{Path: "/v1/user/category", Next: api.updateUserCategoryEndpoint, Method: "PUT", Auth: true, Group: permission.Backend},
		// &APIHandler{Path: "/v1/user/permission", Next: api.updateUserPemissionEndpoint, Method: "PUT", Auth: true, Group: permission.Backend},
		// &APIHandler{Path: "/v1/user/{PHONE}", Next: api.deleteUserEndpoint, Method: "DELETE", Auth: true, Group: permission.Backend},
	}
}

func (api *ARAPI) deleteAccountReceivableEndpoint(w http.ResponseWriter, req *http.Request) {
	dbname := req.Header.Get("dbname")
	vars := util.GetPathVars(req, []string{"ID"})
	ID := vars["ID"].(string)

	am := model.GetARModel(di)
	if err := am.DeleteAccountReceivable(ID, dbname); err != nil {
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

func (api *ARAPI) updateAccountReceivableEndpoint(w http.ResponseWriter, req *http.Request) {

	vars := util.GetPathVars(req, []string{"ID"})
	ID := vars["ID"].(string)
	dbname := req.Header.Get("dbname")
	iUAR := inputUpdateAR{}
	err := json.NewDecoder(req.Body).Decode(&iUAR)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid JSON format"))
		return
	}

	out, err := json.Marshal(iUAR)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(out))

	if ok, err := iUAR.isARValid(); !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	am := model.GetARModel(di)
	model.GetRTModel(di) //
	model.GetCModel(di)
	model.GetDecuctModel(di)

	if err := am.UpdateAccountReceivable(iUAR.Amount, ID, dbname, iUAR.SalerList); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
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
	dbname := req.Header.Get("dbname")
	am := model.GetARModel(di)

	queryVar := util.GetQueryValue(req, []string{"key", "status", "export"}, true)
	key := (*queryVar)["key"].(string)
	status := (*queryVar)["status"].(string)
	if status == "" {
		status = "0"
	}
	am.GetARData(key, status, dbname)
	//data, err := json.Marshal(result)
	data, err := am.Json("ar")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (api *ARAPI) getSalerDataEndpoint(w http.ResponseWriter, req *http.Request) {

	am := model.GetARModel(di)

	queryVar := util.GetQueryValue(req, []string{"branch"}, true)
	branch := (*queryVar)["branch"].(string)
	dbname := req.Header.Get("dbname")
	if branch == "" || branch == "全部" || strings.ToLower(branch) == "all" {
		branch = "%"
	}

	am.GetSalerData(branch, dbname)
	//data, err := json.Marshal(result)
	data, err := am.Json("saler")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}
func (api *ARAPI) getReCountEndpoint(w http.ResponseWriter, req *http.Request) {

	am := model.GetARModel(di)
	am.ReCount()
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("done"))
}
func (api *ARAPI) getHouseGoEndpoint(w http.ResponseWriter, req *http.Request) {
	dbname := req.Header.Get("dbname")
	var queryDate time.Time
	today := time.Date(queryDate.Year(), queryDate.Month(), 1, 0, 0, 0, 0, queryDate.Location())
	end := time.Date(queryDate.Year(), queryDate.Month()+1, 1, 0, 0, 0, 0, queryDate.Location())

	queryVar := util.GetQueryValue(req, []string{"key", "export"}, true)
	key := (*queryVar)["key"].(string)

	am := model.GetARModel(di)
	am.GetHouseGoData(today, end, key, dbname)
	//data, err := json.Marshal(result)
	data, err := am.Json("housego")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (api *ARAPI) createAccountReceivableEndpoint(w http.ResponseWriter, req *http.Request) {

	//正常網站下的流程
	dbname := req.Header.Get("dbname")
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

	_err := am.CreateAccountReceivable(iAR.GetAR(), "nil", dbname)
	if _err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error"))
	} else {
		w.Write([]byte("OK"))
	}

}

func (api *ARAPI) CreateHouseGoEndpoint(w http.ResponseWriter, req *http.Request) {
	fmt.Println("houseGo inc.")
	dbname := req.Header.Get("dbname")
	iGoAR := inputhouseGoAR{}
	//err := json.NewDecoder(req.Body).Decode(&iGoAR)
	data, _ := ioutil.ReadAll(req.Body) //把  body 内容读入字符串
	//fmt.Println("data:", string(data))

	err := json.Unmarshal([]byte(data), &iGoAR)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid JSON format"))
		return
	}

	// out, err := json.Marshal(iGoAR)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(string(out))

	// if err != nil {
	// 	w.WriteHeader(http.StatusBadRequest)
	// 	w.Write([]byte("Invalid JSON format"))
	// 	return
	// }

	if ok, err := iGoAR.isGoARValid(); !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	am := model.GetARModel(di)

	_err := am.CreateAccountReceivable(iGoAR.GetAR(ACTION_BUY), string(data), dbname)
	if _err != nil {
		//不重複執行放入HouseGo Table
		if _err.Error() == "duplicate data" {
			w.Write([]byte("OK"))
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error." + _err.Error()))
		return
	}
	_err = am.CreateAccountReceivable(iGoAR.GetAR(ACTION_SELL), string(data), dbname)
	if _err != nil {
		// if _err.Error() == "duplicate data" {
		// 	w.WriteHeader(http.StatusBadRequest)
		// 	w.Write([]byte(_err.Error()))
		// 	return
		// }
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error." + _err.Error()))
	} else {
		w.Write([]byte("OK"))
	}

}

func (api *ARAPI) UpgradeARInfoWithHouseGoEndpoint(w http.ResponseWriter, req *http.Request) {
	dbname := req.Header.Get("dbname")
	vars := util.GetPathVars(req, []string{"ID"})
	ID := vars["ID"].(string)

	am := model.GetARModel(di)
	if err := am.UpgradeARInfo(ID, dbname); err != nil {
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

func (api *ARAPI) createDeductEndpoint(w http.ResponseWriter, req *http.Request) {
	//Get params from body
	dbname := req.Header.Get("dbname")
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

	_err := DM.CreateDeduct(iDeduct.GetDeduct(), dbname)
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
	dbname := req.Header.Get("dbname")
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
	_err := rm.CreateReceipt(irt.GetReceipt(), dbname, nil)
	if _err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(_err.Error()))
	} else {
		w.Write([]byte("OK"))
	}

}

func (iGoAR *inputhouseGoAR) isGoARValid() (bool, error) {
	// if !util.IsStrInList(iAR.Permission, permission.All...) {
	// 	return false, errors.New("permission error")
	// }

	// if t := time.Now().Unix(); t <= iAR.Date.Unix() {
	// 	//未來的成交案 => 不成立
	// 	return false, errors.New("CompletionDate is not valid")
	// }

	if iGoAR.CNo == "" {
		return false, errors.New("ContractNo is empty")
	}
	if iGoAR.CaseName == "" {
		return false, errors.New("Case name is empty")
	}

	if iGoAR.Sell.Name == "" {
		return false, errors.New("Sell's name is empty")
	}
	if iGoAR.Buyer.Name == "" {
		return false, errors.New("Buyer's name is empty")
	}

	// if iAR.Fee < 0 || iAR.Fee > iAR.Amount {
	// 	return false, errors.New("Fee is not valid")
	// }
	for _, element := range iGoAR.Sales {
		if element.Percent < 0 {
			return false, errors.New("Proportion is not valid")
		}
		if element.Sid == "" {
			return false, errors.New("id is empty")
		}
		if element.SName == "" {
			return false, errors.New("name is empty")
		}
	}
	if len(iGoAR.Sales) == 0 {
		return false, errors.New("Sales is empty")
	}

	return true, nil
}

func (iAR *inputAR) isARValid() (bool, error) {
	// if !util.IsStrInList(iAR.Permission, permission.All...) {
	// 	return false, errors.New("permission error")
	// }
	fmt.Println("iAR.Date:", iAR.Date)

	// if t := time.Now().Unix(); t <= iAR.Date.Unix() {
	// 	//未來的成交案 => 不成立
	// 	return false, errors.New("CompletionDate is not valid")
	// }

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
	if irt.Fee <= -1 {
		return false, errors.New("Fee is not valid")
	}
	// if t := time.Now().Unix(); t <= irt.Date.Unix() {
	// 	//未來的成交案 => 不成立
	// 	return false, errors.New("Date is not valid")
	// }

	return true, nil
}

func (iUAR *inputUpdateAR) isARValid() (bool, error) {
	if iUAR.Amount <= 0 {
		return false, errors.New("amount is not vaild")
	}

	if len(iUAR.SalerList) <= 0 {
		return false, errors.New("SalerList is empty")
	}

	for _, element := range iUAR.SalerList {
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

func (iGoAR *inputhouseGoAR) GetAR(action string) *model.AR {
	var customer = model.Customer{}
	var amount = 0
	var arid string
	customer.Action = "none"
	if action == ACTION_BUY {
		customer.Action = ACTION_BUY
		amount = iGoAR.Buyer.Amount
		arid = strconv.Itoa(iGoAR.ARid) + "_b"
	} else if action == ACTION_SELL {
		arid = strconv.Itoa(iGoAR.ARid) + "_s"
		customer.Action = ACTION_SELL
		amount = iGoAR.Sell.Amount
	}

	customer.Name = iGoAR.Buyer.Name

	var sales []*model.MAPSaler
	for _, element := range iGoAR.Sales {
		var data = model.MAPSaler{}
		data.Percent = element.Percent
		data.SName = element.SName
		data.Sid = element.Sid
		sales = append(sales, &data)
	}

	return &model.AR{
		ARid:     arid,
		Amount:   amount,
		Date:     iGoAR.Date,
		CNo:      iGoAR.CNo,
		CaseName: iGoAR.CaseName,
		Customer: customer,
		//Fee:      iAR.Fee,
		Sales: sales,
	}
}

func (irt *inputReceipt) GetReceipt() *model.Receipt {
	return &model.Receipt{
		Amount: irt.Amount,
		Date:   irt.Date,
		ARid:   irt.ARid,
		Fee:    irt.Fee,
	}
}
