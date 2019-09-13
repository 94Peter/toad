package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

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

type inputInvoice struct {
	Rid string `json:"id"`
	//Date   time.Time `json:"date"`
	Title       string `json:"title"`
	GUI         string `json:"GUI"`
	InvoiceType string `json:"invoiceType"`
}

func (api ReceiptAPI) GetAPIs() *[]*APIHandler {
	return &[]*APIHandler{
		&APIHandler{Path: "/v1/receipt", Next: api.getReceiptEndpoint, Method: "GET", Auth: false, Group: permission.All},
		&APIHandler{Path: "/v1/receipt/{ID}", Next: api.deleteReceiptEndpoint, Method: "DELETE", Auth: false, Group: permission.All},
		&APIHandler{Path: "/v1/receipt/{ID}", Next: api.updateReceiptEndpoint, Method: "PUT", Auth: false, Group: permission.All},
		&APIHandler{Path: "/v1/invoice", Next: api.createInvoiceEndpoint, Method: "POST", Auth: false, Group: permission.All},
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
	//var queryDate time.Time
	//today := time.Date(queryDate.Year(), queryDate.Month(), 1, 0, 0, 0, 0, queryDate.Location())
	//end := time.Date(queryDate.Year(), queryDate.Month()+1, 1, 0, 0, 0, 0, queryDate.Location())

	queryVar := util.GetQueryValue(req, []string{"begin", "end"}, true)
	begin := (*queryVar)["begin"].(string)
	end := (*queryVar)["end"].(string)
	if begin == "" {
		begin = "2000-01-01T00:00:00Z"
	}
	if end == "" {
		end = "now()"
	}

	rm.GetReceiptData(begin, end)
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

func (api *ReceiptAPI) createInvoiceEndpoint(w http.ResponseWriter, req *http.Request) {
	//Get params from body
	iInovice := inputInvoice{}

	err := json.NewDecoder(req.Body).Decode(&iInovice)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid JSON format"))
		return
	}

	if ok, err := iInovice.isInvoiceValid(); !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	im := model.GetInvoiceModel(di)

	_err := im.CreateInvoice(iInovice.GetInvoice())
	if _err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(_err.Error()))
	} else {
		w.Write([]byte("OK"))
	}

}

func (iInovice *inputInvoice) GetInvoice() *model.Invoice {
	return &model.Invoice{
		GUI:   iInovice.GUI,
		Title: iInovice.Title,
		Rid:   iInovice.Rid,
	}
}

func (iInovice *inputInvoice) isInvoiceValid() (bool, error) {
	if iInovice.Rid == "" {
		return false, errors.New("id is empty")
	}

	if iInovice.InvoiceType == "TUI" {
		if iInovice.GUI == "" {
			return false, errors.New("gui is empty")
		}
		if iInovice.Title == "" {
			return false, errors.New("title is empty")
		}

	} else if iInovice.InvoiceType == "" {
		return false, errors.New("invoiceType is empty")
	} else if iInovice.InvoiceType == "DUI" {
		return true, nil
	} else {
		return false, errors.New("invoiceType is not vaild")
	}

	return true, nil
}
