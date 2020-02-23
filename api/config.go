package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
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
const isDelete = 3

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
	Address     string `json:"address"`
	Birth       string `json:"birth"`
	IdentityNum string `json:"identityNum"`
	BankAccount string `json:"bankAccount"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	Remark      string `json:remark`
}

type inputConfigParameter struct {
	//Date   time.Time `json:"date"`
	Date string `json:"date"`
	//IT     float64 `json:"IT"`
	MMW    int     `json:"MMW"` //最低基本薪資
	NHI    float64 `json:"NHI"`
	LI     float64 `json:"LI"`
	NHI2nd float64 `json:"NHI2nd"`
	//AnnualRatio float64 `json:"annualRatio"`
}
type inputConfigBranch struct {
	Branch        string  `json:"branch"`
	Rent          int     `json:"rent"`
	AgentSign     int     `json:"agentSign"`
	CommercialFee float64 `json:"commercialFee"`
	Manager       string  `json:"manager"`
	AnnualRatio   float64 `json:"annualRatio"`
	Sid           string  `json:"sid"`
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
		//&APIHandler{Path: "/v1/config/branch", Next: api.createConfigBranchEndpoint, Method: "POST", Auth: false, Group: permission.All},
		&APIHandler{Path: "/v1/config/branch", Next: api.createConfigBranchEndpointWithStringArray, Method: "POST", Auth: false, Group: permission.All},
		&APIHandler{Path: "/v1/config/branch/{Branch}", Next: api.updateConfigBranchEndpoint, Method: "PUT", Auth: false, Group: permission.All},
		&APIHandler{Path: "/v1/config/branch/{Branch}", Next: api.deleteConfigBranchEndpoint, Method: "DELETE", Auth: false, Group: permission.All},

		&APIHandler{Path: "/v1/config/parameter", Next: api.getConfigParameterEndpoint, Method: "GET", Auth: false, Group: permission.All},
		&APIHandler{Path: "/v1/config/parameter", Next: api.createConfigParameterEndpoint, Method: "POST", Auth: false, Group: permission.All},
		&APIHandler{Path: "/v1/config/parameter/{id}", Next: api.updateConfigParameterEndpoint, Method: "PUT", Auth: false, Group: permission.All},
		&APIHandler{Path: "/v1/config/parameter/{id}", Next: api.deleteConfigParameterEndpoint, Method: "DELETE", Auth: false, Group: permission.All},

		&APIHandler{Path: "/v1/config/saler/check", Next: api.checkConfigSalerEndpoint, Method: "POST", Auth: false, Group: permission.All},

		&APIHandler{Path: "/v1/config/saler", Next: api.getConfigSalerEndpoint, Method: "GET", Auth: false, Group: permission.All},
		&APIHandler{Path: "/v1/config/saler", Next: api.createConfigSalerEndpoint, Method: "POST", Auth: false, Group: permission.All},
		&APIHandler{Path: "/v1/config/saler/{csID}", Next: api.updateConfigSalerEndpoint, Method: "PUT", Auth: false, Group: permission.All},
		&APIHandler{Path: "/v1/config/saler/{csID}", Next: api.deleteConfigSalerEndpoint, Method: "DELETE", Auth: false, Group: permission.All},

		&APIHandler{Path: "/v1/config/salary", Next: api.createConfigSalaryEndpoint, Method: "POST", Auth: false, Group: permission.All}, //內建PUT更改
		&APIHandler{Path: "/v1/config/salary", Next: api.getConfigSalaryEndpoint, Method: "GET", Auth: false, Group: permission.All},
		&APIHandler{Path: "/v1/config/salary/{id}", Next: api.deleteConfigSalaryEndpoint, Method: "DELETE", Auth: false, Group: permission.All},
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

func (api *ConfigAPI) createConfigBranchEndpointWithStringArray(w http.ResponseWriter, req *http.Request) {
	//取得body
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		http.Error(w, "can't read body", http.StatusBadRequest)
		return
	}
	//將body讀成string
	branch := fmt.Sprintf("%s", body)

	err2 := strings.ContainsAny(branch, "{:}") || (len(branch) <= 4) || !strings.ContainsAny(branch, "[\"]")
	if err2 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid string array format"))
		return
	}

	branchList := []string{}
	s := strings.Split(string(branch), "\"")
	for _, each := range s {
		//去除"字符 寫入 golang str array
		if each != "," && each != "[" && each != "]" {
			branchList = append(branchList, each)
		}
	}
	fmt.Println(branchList)

	configM := model.GetConfigModel(di)
	err = configM.CreateConfigBranch(branchList)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	} else {
		w.Write([]byte("OK"))
	}
}

func (api *ConfigAPI) deleteConfigBranchEndpoint(w http.ResponseWriter, req *http.Request) {
	vars := util.GetPathVars(req, []string{"Branch"})
	Branch := vars["Branch"].(string)

	configM := model.GetConfigModel(di)

	_err := configM.DeleteConfigBranch(Branch)
	if _err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(_err.Error() + ",maybe Branch is not exist"))
	} else {
		w.Write([]byte("OK"))
	}
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

	_err := configM.CreateConfigBranchWithManager(iCBranch.GetConfigBranch(isCreate))
	if _err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error,maybe already exist or Saler is not exist or Branch not match"))
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
		w.Write([]byte(_err.Error() + ",maybe Saler is not exist or Branch not match"))
	} else {
		w.Write([]byte("OK"))
	}

}

func (api *ConfigAPI) deleteConfigParameterEndpoint(w http.ResponseWriter, req *http.Request) {
	//Get params from body
	vars := util.GetPathVars(req, []string{"id"})
	id := vars["id"].(string)

	configM := model.GetConfigModel(di)

	// time, err := time.ParseInLocation("2006-01-02", Date, time.Local)
	// if err != nil {
	// 	w.WriteHeader(http.StatusInternalServerError)
	// 	w.Write([]byte("date is not valid, " + err.Error()))
	// 	return
	// }

	_err := configM.DeleteConfigParameter(id)
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
	vars := util.GetPathVars(req, []string{"id"})
	id := vars["id"].(string)

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

	_err := configM.UpdateConfigParameter(iCParam.GetConfigParameter(), id)
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
	// year, month, day := time.Now().Date()
	// if day >= 1 {
	// 	fmt.Println(year, month, day, "啟動WorkValidDate()，更新員工有效日薪水")
	// 	configM.WorkValidDate()
	// }
	//text := time.Now().Format("2006-01-02")

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

func (api *ConfigAPI) getConfigSalaryEndpoint(w http.ResponseWriter, req *http.Request) {

	configM := model.GetConfigModel(di)

	queryVar := util.GetQueryValue(req, []string{"id"}, true)

	sid := (*queryVar)["id"].(string)
	if sid == "" || sid == "全部" || strings.ToLower(sid) == "all" {
		sid = "%"
	}

	configM.GetConfigSalaryData(sid)
	//data, err := json.Marshal(result)
	data, err := configM.Json("ConfigSalary")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (api *ConfigAPI) checkConfigSalerEndpoint(w http.ResponseWriter, req *http.Request) {
	//Get params from body

	iCSaler := inputConfigSaler{}
	err := json.NewDecoder(req.Body).Decode(&iCSaler)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid JSON format"))
		return
	}

	if ok, err := iCSaler.checkConfigSalerVaild(); !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	configM := model.GetConfigModel(di)

	r, _err := configM.CheckConfigSaler(iCSaler.IdentityNum, iCSaler.ZeroDate)
	if _err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(_err.Error()))
	} else {
		w.Write([]byte(r))
	}

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

//借用同樣的結構(inputConfigSaler)建立ConfigSalary結構
func (api *ConfigAPI) createConfigSalaryEndpoint(w http.ResponseWriter, req *http.Request) {
	//Get params from body

	iCSaler := inputConfigSaler{}
	err := json.NewDecoder(req.Body).Decode(&iCSaler)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid JSON format"))
		return
	}

	if ok, err := iCSaler.isConfigSalaryValid(); !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	configM := model.GetConfigModel(di)

	_err := configM.CreateConfigSalary(iCSaler.GetConfigSalary())
	if _err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error"))
	} else {
		w.Write([]byte("OK"))
	}

}

func (api *ConfigAPI) updateConfigSalerEndpoint(w http.ResponseWriter, req *http.Request) {
	//Get params from body
	vars := util.GetPathVars(req, []string{"csID"})
	csID := vars["csID"].(string)

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

	_err := configM.UpdateConfigSaler(iCSaler.GetConfigSaler(), csID)
	if _err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error:" + _err.Error()))
	} else {
		w.Write([]byte("OK"))
	}

}

func (api *ConfigAPI) deleteConfigSalerEndpoint(w http.ResponseWriter, req *http.Request) {
	//Get params from body
	vars := util.GetPathVars(req, []string{"csID"})
	csID := vars["csID"].(string)

	configM := model.GetConfigModel(di)

	_err := configM.DeleteConfigSaler(csID)
	if _err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error:" + _err.Error()))
	} else {
		w.Write([]byte("OK"))
	}

}

func (api *ConfigAPI) deleteConfigSalaryEndpoint(w http.ResponseWriter, req *http.Request) {
	//Get path from body
	vars := util.GetPathVars(req, []string{"id"})
	id := vars["id"].(string)

	queryVar := util.GetQueryValue(req, []string{"zerodate"}, true)

	zerodate := (*queryVar)["zerodate"].(string)
	if zerodate == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("zerodate is empty"))
		return
	}

	configM := model.GetConfigModel(di)

	_err := configM.DeleteConfigSalary(id, zerodate)
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
	if iCBranch.Manager == "" {
		return false, errors.New("manager is empty")
	}
	if iCBranch.Sid == "" {
		return false, errors.New("sid is empty")
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
	if iCBranch.AnnualRatio < 0 || iCBranch.AnnualRatio > 100 {
		return false, errors.New("AnnualRatio is not valid")
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
			Manager:       iCBranch.Manager,
			AnnualRatio:   iCBranch.AnnualRatio,
			Sid:           iCBranch.Sid,
		}
	}
	if command == isUpdate {
		return &model.ConfigBranch{
			Rent:          iCBranch.Rent,
			AgentSign:     iCBranch.AgentSign,
			CommercialFee: iCBranch.CommercialFee,
			Manager:       iCBranch.Manager,
			AnnualRatio:   iCBranch.AnnualRatio,
			Sid:           iCBranch.Sid,
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

	if iCParam.MMW < 0 {
		return false, errors.New("MMW is not valid")
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

	// if iCParam.AnnualRatio < 0 || iCParam.AnnualRatio > 100 {
	// 	return false, errors.New("AnnualRatio is not valid")
	// }
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

func (iCSaler *inputConfigSaler) checkConfigSalerVaild() (bool, error) {

	_, err := time.ParseInLocation("2006-01-02", iCSaler.ZeroDate, time.Local)
	if err != nil {
		return false, errors.New("zeroDate is not valid, " + err.Error())
	}

	if iCSaler.IdentityNum == "" {
		return false, errors.New("IdentityNum is empty")
	}

	return true, nil
}

func (iCSaler *inputConfigSaler) isConfigSalaryValid() (bool, error) {

	_, err := time.ParseInLocation("2006-01-02", iCSaler.ZeroDate, time.Local)
	if err != nil {
		return false, errors.New("zeroDate is not valid, " + err.Error())
	}

	if iCSaler.Salary < 0 {
		return false, errors.New("salary is not valid")
	}
	if iCSaler.Percent < 0 {
		return false, errors.New("percent is not valid")
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
		return false, errors.New("association is not valid, only 0 or 1")
	}

	if iCSaler.Sid == "" {
		return false, errors.New("id(sid) is empty")
	}

	if iCSaler.SName == "" {
		return false, errors.New("name is empty")
	}

	if iCSaler.Branch == "" {
		return false, errors.New("branch is empty")
	}

	return true, nil
}

func (iCSaler *inputConfigSaler) isConfigSalerValid(command int) (bool, error) {
	// if !util.IsStrInList(iAR.Permission, permission.All...) {
	// 	return false, errors.New("permission error")
	// }
	if command == isCreate {
		// if iCSaler.Sid == "" {
		// 	return false, errors.New("id is empty")
		// }
		if iCSaler.SName == "" {
			return false, errors.New("name is empty")
		}
	}
	if iCSaler.ValidDate == "" {
		iCSaler.ValidDate = "0001-01-01"
	}

	_, err := time.ParseInLocation("2006-01-02", iCSaler.ZeroDate, time.Local)
	if err != nil {
		return false, errors.New("zeroDate is not valid, " + err.Error())
	}
	_, err = time.ParseInLocation("2006-01-02", iCSaler.ValidDate, time.Local)
	if err != nil {
		return false, errors.New("validDate is not valid, " + err.Error())
	}

	if iCSaler.Salary < 0 {
		return false, errors.New("salary is not valid")
	}
	if iCSaler.Percent < 0 {
		return false, errors.New("percent is not valid")
	}
	if iCSaler.ValidDate == "0001-01-01" {
		iCSaler.FPercent = -1
		iCSaler.Pay = -1
	}
	if iCSaler.FPercent < 0 && iCSaler.FPercent != -1 {
		return false, errors.New("fPercent is not valid")
	}
	if iCSaler.Pay < 0 && iCSaler.Pay != -1 {
		return false, errors.New("pay is not valid")
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
	if iCSaler.IdentityNum == "" {
		return false, errors.New("identityNum is empty")
	}
	iCSaler.Sid = iCSaler.IdentityNum

	if iCSaler.Branch == "" {
		return false, errors.New("branch is empty")
	}

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
		FPercent:       iCSaler.FPercent,
		Branch:         iCSaler.Branch,
		PayrollBracket: iCSaler.PayrollBracket,
		Enrollment:     iCSaler.Enrollment,
		Association:    iCSaler.Association,
		Address:        iCSaler.Address,
		Birth:          iCSaler.Birth,
		IdentityNum:    iCSaler.IdentityNum,
		BankAccount:    iCSaler.BankAccount,
		Email:          iCSaler.Email,
		Phone:          iCSaler.Phone,
		Remark:         iCSaler.Remark,
	}
}

func (iCSaler *inputConfigSaler) GetConfigSalary() *model.ConfigSalary {

	return &model.ConfigSalary{
		Sid:            iCSaler.Sid,
		SName:          iCSaler.SName,
		Salary:         iCSaler.Salary,
		Percent:        iCSaler.Percent,
		Title:          iCSaler.Title,
		ZeroDate:       iCSaler.ZeroDate,
		Branch:         iCSaler.Branch,
		PayrollBracket: iCSaler.PayrollBracket,
		Enrollment:     iCSaler.Enrollment,
		Association:    iCSaler.Association,
		Remark:         iCSaler.Remark,
	}
}

func (iCParam *inputConfigParameter) GetConfigParameter() *model.ConfigParameter {
	the_time, _ := time.ParseInLocation("2006-01-02", iCParam.Date, time.Local)
	return &model.ConfigParameter{
		Date:   the_time,
		NHI:    iCParam.NHI,
		NHI2nd: iCParam.NHI2nd,
		MMW:    iCParam.MMW,
		LI:     iCParam.LI,
		//AnnualRatio: iCParam.AnnualRatio,
	}
}
