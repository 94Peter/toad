package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/94peter/pica/permission"
	"github.com/94peter/toad/model"
	"github.com/94peter/toad/util"
)

type ConfigAPI bool

const isCreate = 1
const isUpdate = 2

type inputAccountItem struct {
	ItemName string `json:"itemName"`
}
type inputConfigSaler struct {
	Sid            string  `json:"id"`
	SName          string  `json:"name"`
	ZeroDate       string  `json:"zeroDate"`
	ValidDate      string  `json:"validDate"`
	Title          string  `json:"title"`
	Salary         int     `json:"salary"`
	Pay            int     `json:"pay"`
	Percent        float64 `json:"percent"`
	FPercent       float64 `json:"fPercent"`
	Branch         string  `json:"branch"`
	PayrollBracket int     `json:"payrollBracket"` //投保金額
	Enrollment     int     `json:"enrollment"`     //加保(眷屬人數)
	Association    int     `json:"association"`    //公會
	// ZeroDate       time.Time `json:"zeroDate"`
	// ValidDate      time.Time `json:"validDate"`
}

type inputConfigParameter struct {
	//Date   time.Time `json:"date"`
	Date   string  `json:"date"`
	IT     float64 `json:"IT"`
	NHI    float64 `json:"NHI"`
	LI     float64 `json:"LI"`
	NHI2nd float64 `json:"NHI2nd"`
}
type inputConfigBranch struct {
	Branch        string  `json:"branch"`
	Rent          int     `json:"rent"`
	AgentSign     int     `json:"agentSign"`
	CommercialFee float64 `json:"commercialFee"`
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
		&APIHandler{Path: "/v1/config/branch/{Branch}", Next: api.updateConfigBranchEndpoint, Method: "PUT", Auth: false, Group: permission.All},

		&APIHandler{Path: "/v1/config/parameter", Next: api.getConfigParameterEndpoint, Method: "GET", Auth: false, Group: permission.All},
		&APIHandler{Path: "/v1/config/parameter", Next: api.createConfigParameterEndpoint, Method: "POST", Auth: false, Group: permission.All},
		&APIHandler{Path: "/v1/config/parameter", Next: api.updateConfigParameterEndpoint, Method: "PUT", Auth: false, Group: permission.All},
		&APIHandler{Path: "/v1/config/parameter/{Date}", Next: api.deleteConfigParameterEndpoint, Method: "DELETE", Auth: false, Group: permission.All},

		&APIHandler{Path: "/v1/config/saler", Next: api.getConfigSalerEndpoint, Method: "GET", Auth: false, Group: permission.All},
		&APIHandler{Path: "/v1/config/saler", Next: api.createConfigSalerEndpoint, Method: "POST", Auth: false, Group: permission.All},
		&APIHandler{Path: "/v1/config/saler/{ID}", Next: api.updateConfigSalerEndpoint, Method: "PUT", Auth: false, Group: permission.All},
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

	if ok, err := iCBranch.isConfigBranchValid(isCreate); !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	configM := model.GetConfigModel(di)

	_err := configM.CreateConfigBranch(iCBranch.GetConfigBranch(isCreate))
	if _err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error,maybe already exist"))
	} else {
		w.Write([]byte("OK"))
	}

}

func (api *ConfigAPI) updateConfigBranchEndpoint(w http.ResponseWriter, req *http.Request) {
	//Get params from path

	vars := util.GetPathVars(req, []string{"Branch"})
	Branch := vars["Branch"].(string)

	iCBranch := inputConfigBranch{}
	err := json.NewDecoder(req.Body).Decode(&iCBranch)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid JSON format"))
		return
	}

	if ok, err := iCBranch.isConfigBranchValid(isUpdate); !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	configM := model.GetConfigModel(di)

	_err := configM.UpdateConfigBranch(Branch, iCBranch.GetConfigBranch(isUpdate))
	if _err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error:" + _err.Error()))
	} else {
		w.Write([]byte("OK"))
	}

}

func (api *ConfigAPI) deleteConfigParameterEndpoint(w http.ResponseWriter, req *http.Request) {
	//Get params from body
	vars := util.GetPathVars(req, []string{"Date"})
	Date := vars["Date"].(string)

	configM := model.GetConfigModel(di)

	time, err := time.ParseInLocation("2006-01-02", Date, time.Local)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("date is not valid, " + err.Error()))
		return
	}

	_err := configM.DeleteConfigParameter(time)
	if _err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error,maybe is not exist"))
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
	iCParam := inputConfigParameter{}
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

	_err := configM.CreateConfigParameter(iCParam.GetConfigParameter())
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
	iCParam := inputConfigParameter{}
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

	_err := configM.UpdateConfigParameter(iCParam.GetConfigParameter())
	if _err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error:" + _err.Error()))
	} else {
		w.Write([]byte("OK"))
	}

}

func (api *ConfigAPI) getConfigSalerEndpoint(w http.ResponseWriter, req *http.Request) {

	configM := model.GetConfigModel(di)
	// var queryDate time.Time
	// today := time.Date(queryDate.Year(), queryDate.Month(), 1, 0, 0, 0, 0, queryDate.Location())
	// end := time.Date(queryDate.Year(), queryDate.Month()+1, 1, 0, 0, 0, 0, queryDate.Location())

	queryVar := util.GetQueryValue(req, []string{"branch"}, true)
	branch := (*queryVar)["branch"].(string)
	if branch == "" || branch == "全部" || strings.ToLower(branch) == "all" {
		branch = "%"
	}

	configM.GetConfigSalerData(branch)
	//data, err := json.Marshal(result)
	data, err := configM.Json("ConfigSaler")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (api *ConfigAPI) createConfigSalerEndpoint(w http.ResponseWriter, req *http.Request) {
	//Get params from body

	iCSaler := inputConfigSaler{}
	err := json.NewDecoder(req.Body).Decode(&iCSaler)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid JSON format"))
		return
	}

	if ok, err := iCSaler.isConfigSalerValid(isCreate); !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	configM := model.GetConfigModel(di)

	_err := configM.CreateConfigSaler(iCSaler.GetConfigSaler())
	if _err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error,maybe already exist"))
	} else {
		w.Write([]byte("OK"))
	}

}

func (api *ConfigAPI) updateConfigSalerEndpoint(w http.ResponseWriter, req *http.Request) {
	//Get params from body
	vars := util.GetPathVars(req, []string{"ID"})
	Sid := vars["ID"].(string)

	iCSaler := inputConfigSaler{}
	err := json.NewDecoder(req.Body).Decode(&iCSaler)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid JSON format"))
		return
	}

	if ok, err := iCSaler.isConfigSalerValid(isUpdate); !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	configM := model.GetConfigModel(di)

	_err := configM.UpdateConfigSaler(iCSaler.GetConfigSaler(), Sid)
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

func (iCBranch *inputConfigBranch) isConfigBranchValid(command int) (bool, error) {
	// if !util.IsStrInList(iAR.Permission, permission.All...) {
	// 	return false, errors.New("permission error")
	// }

	// if t := time.Now().Unix(); t <= iAR.Date.Unix() {
	// 	//未來的成交案 => 不成立
	// 	return false, errors.New("CompletionDate is not valid")
	// }
	if command == isCreate {
		if iCBranch.Branch == "" {
			return false, errors.New("branch is empty")
		}
	}
	if iCBranch.Rent < 0 {
		return false, errors.New("rent is not valid")
	}

	if iCBranch.AgentSign < 0 {
		return false, errors.New("agentSign is not valid")
	}

	if iCBranch.CommercialFee < 0 || iCBranch.CommercialFee > 100 {
		return false, errors.New("commercialFee is not valid")
	}

	return true, nil
}

func (iCBranch *inputConfigBranch) GetConfigBranch(command int) *model.ConfigBranch {
	if command == isCreate {
		return &model.ConfigBranch{
			Branch:        iCBranch.Branch,
			Rent:          iCBranch.Rent,
			AgentSign:     iCBranch.AgentSign,
			CommercialFee: iCBranch.CommercialFee,
		}
	}
	if command == isUpdate {
		return &model.ConfigBranch{
			Rent:          iCBranch.Rent,
			AgentSign:     iCBranch.AgentSign,
			CommercialFee: iCBranch.CommercialFee,
		}
	}
	return nil
}

func isConfigParameterValid(iCParam inputConfigParameter) (bool, error) {
	// if !util.IsStrInList(iAR.Permission, permission.All...) {
	// 	return false, errors.New("permission error")
	// }

	_, err := time.ParseInLocation("2006-01-02", iCParam.Date, time.Local)
	if err != nil {
		return false, errors.New("date is not valid, " + err.Error())
	}

	if iCParam.IT < 0 || iCParam.IT > 100 {
		return false, errors.New("IT is not valid")
	}
	if iCParam.NHI < 0 || iCParam.NHI > 100 {
		return false, errors.New("NHI is not valid")
	}
	if iCParam.NHI2nd < 0 || iCParam.NHI2nd > 100 {
		return false, errors.New("NHI2nd is not valid")
	}
	if iCParam.LI < 0 || iCParam.LI > 100 {
		return false, errors.New("LI is not valid")
	}
	// if t := time.Now().Unix(); t <= iAR.Date.Unix() {
	// 	//未來的成交案 => 不成立
	// 	return false, errors.New("CompletionDate is not valid")
	// }
	// for _, param := range iCParam {
	// 	if param.Param == "" {
	// 		return false, errors.New("param is empty")
	// 	}
	// 	if param.Value < 0 {
	// 		return false, errors.New("value is not valid")
	// 	}
	// }

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

func (iCSaler *inputConfigSaler) isConfigSalerValid(command int) (bool, error) {
	// if !util.IsStrInList(iAR.Permission, permission.All...) {
	// 	return false, errors.New("permission error")
	// }
	if command == isCreate {
		if iCSaler.Sid == "" {
			return false, errors.New("id is empty")
		}
		if iCSaler.SName == "" {
			return false, errors.New("name is empty")
		}
	}

	_, err := time.ParseInLocation("2006-01-02", iCSaler.ZeroDate, time.Local)
	if err != nil {
		return false, errors.New("zeroDate is not valid, " + err.Error())
	}
	_, err = time.ParseInLocation("2006-01-02", iCSaler.ValidDate, time.Local)
	if err != nil {
		return false, errors.New("validDate is not valid, " + err.Error())
	}

	if iCSaler.Pay < 0 {
		return false, errors.New("pay is not valid")
	}
	if iCSaler.Salary < 0 {
		return false, errors.New("salary is not valid")
	}
	if iCSaler.Percent < 0 {
		return false, errors.New("percent is not valid")
	}
	if iCSaler.FPercent < 0 {
		return false, errors.New("fPercent is not valid")
	}
	if iCSaler.Title == "" {
		return false, errors.New("title is empty")
	}
	if iCSaler.PayrollBracket < 0 {
		return false, errors.New("payrollBracket is not valid")
	}
	if iCSaler.Enrollment < 0 {
		return false, errors.New("enrollment is not valid")
	}
	if !(iCSaler.Association == 0 || iCSaler.Association == 1) {
		return false, errors.New("association is not valid")
	}

	// if iCSaler.ValidDate == "" {
	// 	return false, errors.New("validDate is empty")
	// }
	// if iCSaler.ZeroDate == "" {
	// 	return false, errors.New("zeroDate is empty")
	// }

	// if iCSaler.Branch == "" {
	// 	return false, errors.New("branch is empty")
	// }

	return true, nil
}

func (iCSaler *inputConfigSaler) GetConfigSaler() *model.ConfigSaler {
	zero_time, _ := time.ParseInLocation("2006-01-02", iCSaler.ZeroDate, time.Local)
	valid_time, _ := time.ParseInLocation("2006-01-02", iCSaler.ValidDate, time.Local)
	return &model.ConfigSaler{
		Sid:            iCSaler.Sid,
		SName:          iCSaler.SName,
		Pay:            iCSaler.Pay,
		Salary:         iCSaler.Salary,
		Percent:        iCSaler.Percent,
		Title:          iCSaler.Title,
		ValidDate:      valid_time,
		ZeroDate:       zero_time,
		Branch:         iCSaler.Branch,
		PayrollBracket: iCSaler.PayrollBracket,
		Enrollment:     iCSaler.Enrollment,
		Association:    iCSaler.Association,
	}
}

func (iCParam *inputConfigParameter) GetConfigParameter() *model.ConfigParameter {
	the_time, _ := time.ParseInLocation("2006-01-02", iCParam.Date, time.Local)
	return &model.ConfigParameter{
		Date:   the_time,
		NHI:    iCParam.NHI,
		NHI2nd: iCParam.NHI2nd,
		IT:     iCParam.IT,
		LI:     iCParam.LI,
	}
}
