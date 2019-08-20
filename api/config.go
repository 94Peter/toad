package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/94peter/pica/permission"
	"github.com/94peter/toad/model"
	"github.com/94peter/toad/util"
)

type ConfigAPI bool

type inputAccountItem struct {
	ItemName string `json:"itemName"`
}
type inputConfigBusiness struct {
	Bid       string    `json:"id"`
	BName     string    `json:"name"`
	ZeroDate  time.Time `json:"zeroDate"`
	ValidDate time.Time `json:"validDate"`
	Title     string    `json:"title"`
	Salary    int       `json:"salary"`
	Pay       int       `json:"pay"`
	Percent   float64   `json:"percent"`
}

// type inputConfigParameter struct {
// 	Param string  `json:"param"`
// 	Value float64 `json:"value"`
// }
type inputConfigBranch struct {
	Branch    string `json:"branch"`
	Rent      int    `json:"rent"`
	AgentSign int    `json:"agentSign"`
}

func (api ConfigAPI) Enable() bool {
	return bool(api)
}

func (api ConfigAPI) GetAPIs() *[]*APIHandler {
	return &[]*APIHandler{
		&APIHandler{Path: "/v1/config/item", Next: api.getAccountItemEndpoint, Method: "GET", Auth: false, Group: permission.All},
		&APIHandler{Path: "/v1/config/item", Next: api.createAccountItemEndpoint, Method: "POST", Auth: false, Group: permission.All},
		&APIHandler{Path: "/v1/config/item/{ItemName}", Next: api.updateAccountItemEndpoint, Method: "PUT", Auth: false, Group: permission.All},
		&APIHandler{Path: "/v1/config/item/{ItemName}", Next: api.deleteAccountItemEndpoint, Method: "DELETE", Auth: false, Group: permission.All},

		&APIHandler{Path: "/v1/config/branch", Next: api.getConfigBranchEndpoint, Method: "GET", Auth: false, Group: permission.All},
		&APIHandler{Path: "/v1/config/branch", Next: api.createConfigBranchEndpoint, Method: "POST", Auth: false, Group: permission.All},
		&APIHandler{Path: "/v1/config/branch", Next: api.updateConfigBranchEndpoint, Method: "PUT", Auth: false, Group: permission.All},
		&APIHandler{Path: "/v1/config/parameter", Next: api.getConfigParameterEndpoint, Method: "GET", Auth: false, Group: permission.All},
		&APIHandler{Path: "/v1/config/parameter", Next: api.createConfigParameterEndpoint, Method: "POST", Auth: false, Group: permission.All},
		&APIHandler{Path: "/v1/config/parameter", Next: api.updateConfigParameterEndpoint, Method: "PUT", Auth: false, Group: permission.All},
		&APIHandler{Path: "/v1/config/business", Next: api.getConfigBusinessEndpoint, Method: "GET", Auth: false, Group: permission.All},
		&APIHandler{Path: "/v1/config/business", Next: api.createConfigBusinessEndpoint, Method: "POST", Auth: false, Group: permission.All},
		&APIHandler{Path: "/v1/config/business", Next: api.updateConfigBusinessEndpoint, Method: "PUT", Auth: false, Group: permission.All},
	}
}

func (api *ConfigAPI) getAccountItemEndpoint(w http.ResponseWriter, req *http.Request) {

	configM := model.GetConfigModel(di)
	var queryDate time.Time
	today := time.Date(queryDate.Year(), queryDate.Month(), 1, 0, 0, 0, 0, queryDate.Location())
	end := time.Date(queryDate.Year(), queryDate.Month()+1, 1, 0, 0, 0, 0, queryDate.Location())

	configM.GetAccountItemData(today, end)
	//data, err := json.Marshal(result)
	data, err := configM.Json("AccountItem")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (api *ConfigAPI) createAccountItemEndpoint(w http.ResponseWriter, req *http.Request) {
	//Get params from body

	iAItem := inputAccountItem{}
	err := json.NewDecoder(req.Body).Decode(&iAItem)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid JSON format"))
		return
	}

	if ok, err := iAItem.isAItemValid(); !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	configM := model.GetConfigModel(di)

	_err := configM.CreateAccountItem(iAItem.GetAccountItem())
	if _err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error,maybe already exist"))
	} else {
		w.Write([]byte("OK"))
	}

}

func (api *ConfigAPI) updateAccountItemEndpoint(w http.ResponseWriter, req *http.Request) {
	//Get params from body
	vars := util.GetPathVars(req, []string{"ItemName"})
	oldItemName := vars["ItemName"].(string)

	iAItem := inputAccountItem{}
	err := json.NewDecoder(req.Body).Decode(&iAItem)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid JSON format"))
		return
	}

	if ok, err := iAItem.isAItemValid(); !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	configM := model.GetConfigModel(di)

	_err := configM.UpdateAccountItem(oldItemName, iAItem.GetAccountItem())
	if _err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error,maybe already exist"))
	} else {
		w.Write([]byte("OK"))
	}

}

func (api *ConfigAPI) deleteAccountItemEndpoint(w http.ResponseWriter, req *http.Request) {
	//Get params from body
	vars := util.GetPathVars(req, []string{"ItemName"})
	oldItemName := vars["ItemName"].(string)

	configM := model.GetConfigModel(di)

	_err := configM.DeleteAccountItem(oldItemName)
	if _err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error,maybe is not exist"))
	} else {
		w.Write([]byte("OK"))
	}

}

func (api *ConfigAPI) getConfigBranchEndpoint(w http.ResponseWriter, req *http.Request) {

	configM := model.GetConfigModel(di)
	var queryDate time.Time
	today := time.Date(queryDate.Year(), queryDate.Month(), 1, 0, 0, 0, 0, queryDate.Location())
	end := time.Date(queryDate.Year(), queryDate.Month()+1, 1, 0, 0, 0, 0, queryDate.Location())

	configM.GetConfigBranchData(today, end)
	//data, err := json.Marshal(result)
	data, err := configM.Json("ConfigBranch")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (api *ConfigAPI) createConfigBranchEndpoint(w http.ResponseWriter, req *http.Request) {
	//Get params from body

	iCBranch := inputConfigBranch{}
	err := json.NewDecoder(req.Body).Decode(&iCBranch)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid JSON format"))
		return
	}

	if ok, err := iCBranch.isConfigBranchValid(); !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	configM := model.GetConfigModel(di)

	_err := configM.CreateConfigBranch(iCBranch.GetConfigBranch())
	if _err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error,maybe already exist"))
	} else {
		w.Write([]byte("OK"))
	}

}

func (api *ConfigAPI) updateConfigBranchEndpoint(w http.ResponseWriter, req *http.Request) {
	//Get params from body

	iCBranch := inputConfigBranch{}
	err := json.NewDecoder(req.Body).Decode(&iCBranch)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid JSON format"))
		return
	}

	if ok, err := iCBranch.isConfigBranchValid(); !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	configM := model.GetConfigModel(di)

	_err := configM.UpdateConfigBranch(iCBranch.GetConfigBranch())
	if _err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error:" + _err.Error()))
	} else {
		w.Write([]byte("OK"))
	}

}

func (api *ConfigAPI) getConfigParameterEndpoint(w http.ResponseWriter, req *http.Request) {

	configM := model.GetConfigModel(di)
	var queryDate time.Time
	today := time.Date(queryDate.Year(), queryDate.Month(), 1, 0, 0, 0, 0, queryDate.Location())
	end := time.Date(queryDate.Year(), queryDate.Month()+1, 1, 0, 0, 0, 0, queryDate.Location())

	configM.GetConfigParameterData(today, end)
	//data, err := json.Marshal(result)
	data, err := configM.Json("ConfigParameter")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (api *ConfigAPI) createConfigParameterEndpoint(w http.ResponseWriter, req *http.Request) {
	//Get params from body

	//iCParam := []*inputConfigParameter{}
	iCParam := model.ConfigParameter{}
	err := json.NewDecoder(req.Body).Decode(&iCParam)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid JSON format"))
		return
	}

	// if ok, err := isConfigParameterValid(iCParam); !ok {
	// 	w.WriteHeader(http.StatusBadRequest)
	// 	w.Write([]byte(err.Error()))
	// 	return
	// }

	configM := model.GetConfigModel(di)

	_err := configM.CreateConfigParameter(iCParam)
	if _err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error,maybe already exist"))
	} else {
		w.Write([]byte("OK"))
	}

}

func (api *ConfigAPI) updateConfigParameterEndpoint(w http.ResponseWriter, req *http.Request) {
	//Get params from body

	//iCParam := []*inputConfigParameter{}
	iCParam := []*model.ConfigParameter{}
	err := json.NewDecoder(req.Body).Decode(&iCParam)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid JSON format"))
		return
	}

	if ok, err := isConfigParameterValid(iCParam); !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	configM := model.GetConfigModel(di)

	_err := configM.UpdateConfigParameter(iCParam)
	if _err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error:" + _err.Error()))
	} else {
		w.Write([]byte("OK"))
	}

}

func (api *ConfigAPI) getConfigBusinessEndpoint(w http.ResponseWriter, req *http.Request) {

	configM := model.GetConfigModel(di)
	var queryDate time.Time
	today := time.Date(queryDate.Year(), queryDate.Month(), 1, 0, 0, 0, 0, queryDate.Location())
	end := time.Date(queryDate.Year(), queryDate.Month()+1, 1, 0, 0, 0, 0, queryDate.Location())

	configM.GetConfigBusinessData(today, end)
	//data, err := json.Marshal(result)
	data, err := configM.Json("ConfigBusiness")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (api *ConfigAPI) createConfigBusinessEndpoint(w http.ResponseWriter, req *http.Request) {
	//Get params from body

	iCBusiness := inputConfigBusiness{}
	err := json.NewDecoder(req.Body).Decode(&iCBusiness)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid JSON format"))
		return
	}

	if ok, err := iCBusiness.isConfigBusinessValid(); !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	configM := model.GetConfigModel(di)

	_err := configM.CreateConfigBusiness(iCBusiness.GetConfigBusiness())
	if _err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error,maybe already exist"))
	} else {
		w.Write([]byte("OK"))
	}

}

func (api *ConfigAPI) updateConfigBusinessEndpoint(w http.ResponseWriter, req *http.Request) {
	//Get params from body

	iCBusiness := inputConfigBusiness{}
	err := json.NewDecoder(req.Body).Decode(&iCBusiness)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid JSON format"))
		return
	}

	if ok, err := iCBusiness.isConfigBusinessValid(); !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	configM := model.GetConfigModel(di)

	_err := configM.UpdateConfigBusiness(iCBusiness.GetConfigBusiness())
	if _err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error:" + _err.Error()))
	} else {
		w.Write([]byte("OK"))
	}

}

func (iAItem *inputAccountItem) isAItemValid() (bool, error) {
	// if !util.IsStrInList(iAR.Permission, permission.All...) {
	// 	return false, errors.New("permission error")
	// }

	// if t := time.Now().Unix(); t <= iAR.Date.Unix() {
	// 	//未來的成交案 => 不成立
	// 	return false, errors.New("CompletionDate is not valid")
	// }
	if iAItem.ItemName == "" {
		return false, errors.New("itemName is empty")
	}

	return true, nil
}

func (iAItem *inputAccountItem) GetAccountItem() *model.AccountItem {
	return &model.AccountItem{
		ItemName: iAItem.ItemName,
	}
}

func (iCBranch *inputConfigBranch) isConfigBranchValid() (bool, error) {
	// if !util.IsStrInList(iAR.Permission, permission.All...) {
	// 	return false, errors.New("permission error")
	// }

	// if t := time.Now().Unix(); t <= iAR.Date.Unix() {
	// 	//未來的成交案 => 不成立
	// 	return false, errors.New("CompletionDate is not valid")
	// }
	if iCBranch.Branch == "" {
		return false, errors.New("branch is empty")
	}
	if iCBranch.Rent < 0 {
		return false, errors.New("rent is not valid")
	}

	if iCBranch.AgentSign < 0 {
		return false, errors.New("agentSign is not valid")
	}

	return true, nil
}

func (iCBranch *inputConfigBranch) GetConfigBranch() *model.ConfigBranch {
	return &model.ConfigBranch{
		Branch:    iCBranch.Branch,
		Rent:      iCBranch.Rent,
		AgentSign: iCBranch.AgentSign,
	}
}

func isConfigParameterValid(iCParam []*model.ConfigParameter) (bool, error) {
	// if !util.IsStrInList(iAR.Permission, permission.All...) {
	// 	return false, errors.New("permission error")
	// }

	// if t := time.Now().Unix(); t <= iAR.Date.Unix() {
	// 	//未來的成交案 => 不成立
	// 	return false, errors.New("CompletionDate is not valid")
	// }
	for _, param := range iCParam {
		if param.Param == "" {
			return false, errors.New("param is empty")
		}
		if param.Value < 0 {
			return false, errors.New("value is not valid")
		}
	}

	// f := map[string]interface{}{
	// 	"Name": "Wednesday",
	// 	"Age":  6,
	// 	"Parents": []interface{}{
	// 		"Gomez",
	// 		"Morticia",
	// 	},
	// }

	// for k, v := range f {
	// 	switch vv := v.(type) {
	// 	case string:
	// 		fmt.Println(k, "is string", vv)
	// 	case float64:
	// 		fmt.Println(k, "is float64", vv)
	// 	case []interface{}:
	// 		fmt.Println(k, "is an array:")
	// 		for i, u := range vv {
	// 			fmt.Println(i, u)
	// 		}
	// 	default:
	// 		fmt.Println(k, "is of a type I don't know how to handle")
	// 	}
	// }

	return true, nil
}

func (iCBusiness *inputConfigBusiness) isConfigBusinessValid() (bool, error) {
	// if !util.IsStrInList(iAR.Permission, permission.All...) {
	// 	return false, errors.New("permission error")
	// }

	if iCBusiness.Bid == "" {
		return false, errors.New("id is empty")
	}
	if iCBusiness.BName == "" {
		return false, errors.New("name is empty")
	}

	if iCBusiness.Pay < 0 {
		return false, errors.New("pay is not valid")
	}
	if iCBusiness.Salary < 0 {
		return false, errors.New("salary is not valid")
	}
	if iCBusiness.Percent < 0 {
		return false, errors.New("percent is not valid")
	}
	if iCBusiness.Title == "" {
		return false, errors.New("title is empty")
	}
	// if iCBusiness.ValidDate == "" {
	// 	return false, errors.New("validDate is empty")
	// }
	// if iCBusiness.ZeroDate == "" {
	// 	return false, errors.New("zeroDate is empty")
	// }
	return true, nil
}

func (iCBusiness *inputConfigBusiness) GetConfigBusiness() *model.ConfigBusiness {
	return &model.ConfigBusiness{
		Bid:       iCBusiness.Bid,
		BName:     iCBusiness.BName,
		Pay:       iCBusiness.Pay,
		Salary:    iCBusiness.Salary,
		Percent:   iCBusiness.Percent,
		Title:     iCBusiness.Title,
		ValidDate: iCBusiness.ValidDate,
		ZeroDate:  iCBusiness.ZeroDate,
	}
}
