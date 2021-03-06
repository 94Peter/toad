package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"toad/excel"
	"toad/model"
	"toad/pdf"
	"toad/permission"
	"toad/txt"
	"toad/util"
)

type SalaryAPI bool

type inputBranchSalary struct {
	Branch string       `json:"branch"`
	Date   time.Time    `json:"date"`
	Name   string       `json:"name"`
	CList  []*model.Cid `json:"commissionList"`
	//Total  string `json:"total"`
	//Lock   int    `json:"Lock"`
}

// type cid struct {
// 	sid string `json:"sid"`
// 	rid string `json:"rid"`
// }

type inputSalerSalary struct {
	//PBonus string   `json:"pbonus"`
	Sid         string `json:"sid"`
	Lbonus      int    `json:"lbonus"`
	Abonus      int    `json:"abonus"`
	Tax         int    `json:"tax"`
	Other       int    `json:"other"`
	Welfare     int    `json:"welfare"`
	SP          int    `json:"sp"`
	Workday     int    `json:"workday"`
	Description string `json:"description"`
	//2020-12-08新增
	Salary    int `json:"salary"`
	LaborFee  int `json:"laborFee"`
	HealthFee int `json:"healthFee"`
	//Total  string `json:"total"`
	//Lock   int    `json:"Lock"`
}

type inputIncomeExpense struct {
	SalerFee    int     `json:"salerFee"`
	EarnAdjust  int     `json:"earnAdjust"`
	AnnualRatio float64 `json:"annualRatio"`
}

type salarylock struct {
	Lock string `json:"lock"`
}

type exportBranchId struct {
	BSidList []struct {
		BSid string `json:"bsid"`
	} `json:"idList"`
}

type inputCloseAccount struct {
	CloseDate time.Time `json:"closeDate"`
	Uid       string    `json:"uid"`
}

func (api SalaryAPI) Enable() bool {
	return bool(api)
}

func (api SalaryAPI) GetAPIs() *[]*APIHandler {
	return &[]*APIHandler{
		&APIHandler{Path: "/v1/download", Next: api.DownloadTest, Method: "GET", Auth: true, Group: permission.All},
		&APIHandler{Path: "/v1/test", Next: api.test, Method: "GET", Auth: true, Group: permission.All},

		&APIHandler{Path: "/v1/salary", Next: api.getBranchSalaryEndpoint, Method: "GET", Auth: true, Group: permission.All},
		&APIHandler{Path: "/v1/salary", Next: api.createBranchSalaryEndpoint, Method: "POST", Auth: true, Group: permission.All},
		&APIHandler{Path: "/v1/salary/{ID}", Next: api.lockSalaryEndpoint, Method: "PUT", Auth: true, Group: permission.All},
		&APIHandler{Path: "/v1/salary/{ID}", Next: api.deleteSalaryEndpoint, Method: "DELETE", Auth: true, Group: permission.All},

		&APIHandler{Path: "/v1/salary/export/{bsID}", Next: api.exportBranchSalaryEndpoint, Method: "GET", Auth: true, Group: permission.All},
		&APIHandler{Path: "/v1/salary/export", Next: api.exportBranchSalaryEndpoint, Method: "POST", Auth: true, Group: permission.All},

		&APIHandler{Path: "/v1/salary/detail/{bsID}", Next: api.getSalerSalaryEndpoint, Method: "GET", Auth: true, Group: permission.All},
		&APIHandler{Path: "/v1/salary/detail/{bsID}", Next: api.updateSalerSalaryEndpoint, Method: "PUT", Auth: true, Group: permission.All},
		&APIHandler{Path: "/v1/salary/detail/refresh/{bsID}", Next: api.refreshSalerSalaryEndpoint, Method: "GET", Auth: true, Group: permission.All},

		&APIHandler{Path: "/v1/NHIsalary/{bsID}", Next: api.getNHISalaryEndpoint, Method: "GET", Auth: true, Group: permission.All},

		&APIHandler{Path: "/v1/managerBonus/{bsID}", Next: api.getManagerBonusEndpoint, Method: "GET", Auth: true, Group: permission.All},
		&APIHandler{Path: "/v1/managerBonus/{bsID}", Next: api.updateManagerBonusEndpoint, Method: "PUT", Auth: true, Group: permission.All},

		&APIHandler{Path: "/v1/closeAccountSettlement", Next: api.closeAccountSettlementEndpoint, Method: "POST", Auth: true, Group: permission.All},
		&APIHandler{Path: "/v1/closeAccountSettlement", Next: api.getCloseAccountSettlementEndpoint, Method: "GET", Auth: true, Group: permission.All},
		&APIHandler{Path: "/v1/closeAccountSettlement", Next: api.deleteCloseAccountSettlementEndpoint, Method: "DELETE", Auth: true, Group: permission.All},
	}

}
func (api *SalaryAPI) test(w http.ResponseWriter, req *http.Request) {

	//dbname := req.Header.Get("dbname")
	//SalaryM := model.GetSalaryModel(di)
	//SalaryM.SetReturnsCommissionBSid(nil, dbname, nil)

	w.Write([]byte("OK"))

}

func (api *SalaryAPI) lockSalaryEndpoint(w http.ResponseWriter, req *http.Request) {

	vars := util.GetPathVars(req, []string{"ID"})
	ID := vars["ID"].(string)
	fmt.Println(ID)
	dbname := req.Header.Get("dbname")
	SalaryM := model.GetSalaryModel(di)
	lock := salarylock{}
	err := json.NewDecoder(req.Body).Decode(&lock)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid JSON format"))
		return
	}

	if !(lock.Lock == "已完成" || lock.Lock == "未完成") {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("lock shoud be 已完成 or 未完成"))
	}
	if err := SalaryM.LockBranchSalary(ID, lock.Lock, dbname); err != nil {
		if strings.Contains(err.Error(), ERROR_CloseDate) {
			w.WriteHeader(http.StatusLocked)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
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

func (api *SalaryAPI) deleteSalaryEndpoint(w http.ResponseWriter, req *http.Request) {

	vars := util.GetPathVars(req, []string{"ID"})
	ID := vars["ID"].(string)
	SalaryM := model.GetSalaryModel(di)
	dbname := req.Header.Get("dbname")
	if err := SalaryM.DeleteSalary(ID, dbname); err != nil {
		if strings.Contains(err.Error(), ERROR_CloseDate) {
			w.WriteHeader(http.StatusLocked)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
		w.Write([]byte(err.Error()))
		return
	}

	w.Write([]byte("ok"))
	return
}

func (api *SalaryAPI) getBranchSalaryEndpoint(w http.ResponseWriter, req *http.Request) {
	dbname := req.Header.Get("dbname")
	queryVar := util.GetQueryValue(req, []string{"date"}, true)
	date := (*queryVar)["date"].(string)
	if date == "" {
		date = "%"
	}
	SalaryM := model.GetSalaryModel(di)

	SalaryM.GetBranchSalaryData(date, dbname)
	//data, err := json.Marshal(result)
	data, err := SalaryM.Json("BranchSalary")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (api *SalaryAPI) getNHISalaryEndpoint(w http.ResponseWriter, req *http.Request) {

	SalaryM := model.GetSalaryModel(di)
	vars := util.GetPathVars(req, []string{"bsID"})
	bsID := vars["bsID"].(string)
	dbname := req.Header.Get("dbname")
	SalaryM.GetNHISalaryData(bsID, dbname)
	//data, err := json.Marshal(result)
	data, err := SalaryM.Json("NHISalary")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (api *SalaryAPI) getSalerSalaryEndpoint(w http.ResponseWriter, req *http.Request) {

	SalaryM := model.GetSalaryModel(di)
	vars := util.GetPathVars(req, []string{"bsID"})
	bsID := vars["bsID"].(string)
	fmt.Println(bsID)
	dbname := req.Header.Get("dbname")
	SalaryM.GetSalerSalaryData(bsID, "%", dbname)
	//data, err := json.Marshal(result)
	data, err := SalaryM.Json("SalerSalary")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (api *SalaryAPI) refreshSalerSalaryEndpoint(w http.ResponseWriter, req *http.Request) {

	SalaryM := model.GetSalaryModel(di)
	vars := util.GetPathVars(req, []string{"bsID"})
	bsID := vars["bsID"].(string)
	fmt.Println(bsID)
	dbname := req.Header.Get("dbname")
	err := SalaryM.ReFreshSalerSalary(bsID, dbname)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.Write([]byte("ok"))
}

func (api *SalaryAPI) getManagerBonusEndpoint(w http.ResponseWriter, req *http.Request) {

	SalaryM := model.GetSalaryModel(di)
	vars := util.GetPathVars(req, []string{"bsID"})
	bsID := vars["bsID"].(string)
	fmt.Println(bsID)
	dbname := req.Header.Get("dbname")
	SalaryM.GetIncomeExpenseData(bsID, dbname)
	//data, err := json.Marshal(result)
	data, err := SalaryM.Json("ManagerBonus")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (api *SalaryAPI) DownloadTest(w http.ResponseWriter, req *http.Request) {
	ReceiveFile(w, req, "薪轉明細表.xlsx")

	//body := "testbody"
	//fname := "hello.pdf"
	conf := di.GetSMTPConf()
	fmt.Println(conf)
	//util.RunSendMail(conf.Host, conf.Port, conf.Password, conf.User, "geassyayaoo3@gmail.com", "subject", body, fname)

	util.GomailMailSend(conf.Host, conf.Port, conf.Password, conf.User, "geassyayaoo3@gmail.com", "subject", "body", "t.txt")
}
func (api *SalaryAPI) exportBranchSalaryEndpoint(w http.ResponseWriter, req *http.Request) {
	fmt.Println("exportBranchSalaryEndpoint")
	dbname := req.Header.Get("dbname")
	exportId := exportBranchId{}
	err := json.NewDecoder(req.Body).Decode(&exportId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid JSON format"))
		return
	}

	out, _ := json.Marshal(exportId)
	fmt.Println("exportId :", string(out))

	//Get params from body
	queryVar := util.GetQueryValue(req, []string{"pdf", "send"}, true)
	//vars := util.GetPathVars(req, []string{"bsID"})
	//bsID := vars["bsID"].(string)
	//bsID := ""
	param := (*queryVar)["pdf"].(string)
	//param_excel := (*queryVar)["excel"].(string)
	//sid := (*queryVar)["sid"].(string)
	send := (*queryVar)["send"].(string)

	var mExport = 0
	//有pdf參數
	if param != "" {

		typePdf, err := strconv.Atoi(param)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		mExport = typePdf

	}
	if !(send == "" || send == "true" || send == "false") {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("send should be true or false or empty"))
		return
	}

	fmt.Println(mExport)
	SalaryM := model.GetSalaryModel(di)

	model.GetCModel(di)            // 會使用到commission model函式，預防崩潰，所以要初始化
	model.GetSystemModel(di)       // 會使用到system model函式，預防崩潰，所以要初始化
	cm := model.GetConfigModel(di) // 會使用到system model函式，預防崩潰，所以要初始化
	pdf.GetNewPDF()                // to renew (不然會沿用到prepay pocket amor的pdf )
	switch mExport {
	case excel.PayrollTransfer: //2
		ex := excel.GetNewExcel()
		cm.ConfigSalerList = []*model.ConfigSaler{}
		for _, element := range exportId.BSidList {
			SalaryM.ExportPayrollTransfer(element.BSid, dbname)
		}
		SalaryM.EXCEL(mExport)

		ex.SaveFile("薪轉明細表")
		ReceiveFile(w, req, "薪轉明細表.xlsx")
		util.DeleteAllFile()
		return
	case excel.IncomeTaxReturn: //9
		ex := excel.GetNewExcel()
		cm.ConfigSalerList = []*model.ConfigSaler{}
		for _, element := range exportId.BSidList {
			SalaryM.ExportIncomeTaxReturn(element.BSid, dbname)
		}
		SalaryM.EXCEL(mExport)

		ex.SaveFile("年度所得申報")
		ReceiveFile(w, req, "年度所得申報.xlsx")
		util.DeleteAllFile()
		return
	case pdf.Commission: //4
		cm := model.GetCModel(di)
		for _, element := range exportId.BSidList {
			data := cm.ExportCommissiontDataByBSid(element.BSid, dbname)
			cm.PDF(pdf.OriPdf, data)
		}
		w.Write(cm.GetBytePDF())
		//w.Write(cm.PDF(false))
		return
	case pdf.SalerSalary: //7
		// if sid == "" {
		// 	w.WriteHeader(http.StatusBadRequest)
		// 	w.Write([]byte(strconv.Itoa(mPdf) + " type should be contain sid"))
		// 	return
		// }
		//SalaryM.GetSalerSalaryData(bsID, sid)
		SalaryM.SetSMTPConf(di.GetSMTPConf())
		for _, element := range exportId.BSidList {
			fmt.Println("element.BSid:", element.BSid)
			SalaryM.GetSalerSalaryData(element.BSid, "%", dbname)
			SalaryM.PDF(dbname, mExport, pdf.NewPdf, send)
		}
		/*
			1.寄送郵件的話，不輸出檔案
			2.個人傭金一起寄信
			3.根據店名、名稱、code綁定檔案 (重複姓名、code會有疑慮)。
		*/
		if send == "true" {
			//8 pdf運行
			for _, element := range exportId.BSidList {
				SalaryM.GetSalerCommission(element.BSid, dbname)
				//SalaryM.PDF(mExport, pdf.OriPdf)
				SalaryM.PDF(dbname, pdf.SalarCommission, pdf.NewPdf, send)
			}
			conf := SalaryM.SMTPConf
			fmt.Println(conf)
			//mailList已經於SalaryM.GetSalerCommission取得，但包含全部的店。
			var wg sync.WaitGroup //送信用背景執行。多執行序
			for _, saler := range SalaryM.MailList {
				//fmt.Println(saler)
				//f1 f2 預設空字串
				f1, f2, f3 := util.GetSameSalerFileName(saler.Branch + "-" + saler.SName + "-" + saler.Code)
				//fmt.Println("f1, f2, f3 ", f1, f2, f3)
				if f1 != "" && f2 != "" {
					// Add goroutine 1.
					wg.Add(1)
					go func() {
						defer wg.Done()
						//util.GomailMailSend(conf.Host, conf.Port, conf.Password, conf.User, saler.Email, "個人薪資(測試郵件)", "薪資表 <b>薪資測試 開啟若有密碼，則為000000或者您的身分證號碼</b>", f1, f2,f3)
						util.GomailMailSend(conf.Host, conf.Port, conf.Password, conf.User, "geassyayaoo3@gmail.com", "個人薪資(測試郵件)", "薪資表 <b>薪資測試 開啟若有密碼，則為000000或者您的身分證號碼</b>", f1, f2, f3)
					}()

				}
			}
			wg.Wait() //等送完信再砍檔案
			fmt.Println("wait all done to DeleteAllFile")
			util.DeleteAllFile()
			w.Write([]byte("OK"))
			return
		}
		util.CompressZip("download")
		ReceiveFile(w, req, "download.zip")
		util.DeleteAllFile()
		return
	case pdf.BranchSalary: //1
		for _, element := range exportId.BSidList {
			fmt.Println("element.BSid:", element.BSid)
			SalaryM.GetSalerSalaryData(element.BSid, "%", dbname)
			SalaryM.PDF(dbname, mExport, pdf.NewPdf)
		}
		util.CompressZip("download")
		ReceiveFile(w, req, "download.zip")
		util.DeleteAllFile()
		return

	case pdf.AgentSign: //5
		//SalaryM.SystemAccountList = []*model.SystemAccount{}
		SalaryM.CommissionList = []*model.Commission{}

		for _, element := range exportId.BSidList {
			SalaryM.GetAgentSign(element.BSid, dbname)
		}

		SalaryM.PDF(dbname, mExport, pdf.OriPdf)
		break

	case pdf.SalarCommission: //8 (在7[pdf.SalerSalary]的時候會執行 這邊應該用不到了!)
		SalaryM.SetSMTPConf(di.GetSMTPConf())
		for _, element := range exportId.BSidList {
			SalaryM.GetSalerCommission(element.BSid, dbname)

			//SalaryM.PDF(mExport, pdf.OriPdf)
			SalaryM.PDF(dbname, mExport, pdf.NewPdf, send)
		}
		//寄送郵件的話，不輸出檔案哦~
		if send == "true" {
			util.DeleteAllFile()
			w.Write([]byte("OK"))
			return
		}
		util.CompressZip("download")
		ReceiveFile(w, req, "download.zip")
		util.DeleteAllFile()
		break
	case pdf.SR: //6
		for _, element := range exportId.BSidList {
			SalaryM.ExportSR(element.BSid, dbname)
			SalaryM.PDF(dbname, mExport, pdf.OriPdf)
		}
		break
	case pdf.NHI: //3
		fmt.Println("NHI:", pdf.NHI)
		SalaryM.NHISalaryList = []*model.NHISalary{}
		for _, element := range exportId.BSidList {
			fmt.Println(element.BSid)
			SalaryM.ExportNHISalaryData(element.BSid, dbname)
		}
		SalaryM.PDF(dbname, mExport, pdf.OriPdf)
		break
	// case pdf.Amortization:
	// 	amor := model.GetAmortizationModel(di) // 會使用到system model函式，預防崩潰，所以要初始化
	// 	amor.GetAmortizationData("1980-01", "2280-01", "%")
	// 	break
	case txt.SalaryTransfer: //13
		fmt.Println("transfer:", mExport)
		//data := []*model.TransferSalary{}

		for _, element := range exportId.BSidList {
			fmt.Println(element.BSid)
			err := SalaryM.MakeTxtTransferSalary(element.BSid, dbname)
			if err != nil {
				fmt.Println(err)
				SalaryM.TransferSalaryList = []*model.TransferSalary{}
				util.DeleteAllFile()
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}
			SalaryM.TXT(mExport)
		}

		util.CompressTxtZip("download")
		ReceiveFile(w, req, "download.zip")
		util.DeleteAllFile()

		return

	case 0: //test

		SalaryM.PDF(dbname, mExport, pdf.OriPdf)
		break
	default:
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("unsupport " + strconv.Itoa(mExport) + " type "))
		return
	}
	//w.Write(SalaryM.PDF(mPdf, pdf.NewPdf))
	w.Write(SalaryM.GetPDFByte())
}

func (api *SalaryAPI) createBranchSalaryEndpoint(w http.ResponseWriter, req *http.Request) {
	//Get params from body
	per := req.Header.Get("AuthPerm")
	dbname := req.Header.Get("dbname")
	iBS := inputBranchSalary{}
	err := json.NewDecoder(req.Body).Decode(&iBS)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid JSON format"))
		return
	}

	if ok, err := iBS.isBranchSalaryValid(); !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	SalaryM := model.GetSalaryModel(di)

	for _, cid := range iBS.CList {
		fmt.Println("cid:", cid.Rid)
		fmt.Println("sid:", cid.Sid)
	}

	err = SalaryM.CreateSalary(iBS.GetBranchSalary(), iBS.CList, dbname, per)
	if err != nil {
		if strings.Contains(err.Error(), ERROR_CloseDate) {
			w.WriteHeader(http.StatusLocked)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		w.Write([]byte("[Error]" + err.Error()))
	} else {
		w.Write([]byte("OK"))
	}

}

func (iBS *inputBranchSalary) isBranchSalaryValid() (bool, error) {

	if iBS.Name == "" {
		return false, errors.New("name is empty")
	}
	if iBS.Branch == "" {
		return false, errors.New("branch is empty")
	}
	// if iBS.Date == "" {
	// 	return false, errors.New("date is empty")
	// }

	// _, err := time.ParseInLocation("2006-01-02", iBS.Date+"-01", time.Local)
	// if err != nil {
	// 	return false, errors.New("[" + iBS.Date + "]date is not valid" + err.Error())
	// }

	return true, nil
}
func (iSS *inputSalerSalary) inputSalerSalaryValid() (bool, error) {

	if iSS.Sid == "" {
		return false, errors.New("sid is empty")
	}
	if iSS.Abonus < 0 {
		return false, errors.New("abonus is not valid")
	}
	if iSS.Lbonus < 0 {
		return false, errors.New("lbonus is not valid")
	}
	if iSS.Workday < 0 {
		return false, errors.New("workday is not valid")
	}
	if iSS.Other < 0 {
		return false, errors.New("other is not valid")
	}
	if iSS.Tax < 0 {
		return false, errors.New("tax is not valid")
	}
	if iSS.SP < 0 {
		return false, errors.New("sp is not valid")
	}
	if iSS.Welfare < 0 {
		return false, errors.New("welfare is not valid")
	}
	if iSS.Salary < 0 {
		return false, errors.New("salary is not valid")
	}
	if iSS.LaborFee < 0 {
		return false, errors.New("laborFee is not valid")
	}
	if iSS.HealthFee < 0 {
		return false, errors.New("healthFee is not valid")
	}

	return true, nil
}

func (iIE *inputIncomeExpense) inpuIncomeExpenseValid() (bool, error) {

	if iIE.EarnAdjust < 0 {
		return false, errors.New("earnAdjust is not valid")
	}
	if iIE.SalerFee < 0 {
		return false, errors.New("salerFee is not valid")
	}
	if iIE.AnnualRatio < 0 || iIE.AnnualRatio > 100 {
		return false, errors.New("annualRatio is not valid")
	}

	return true, nil
}

func (iCA *inputCloseAccount) inputCloseAccountValid() (bool, error) {

	if iCA.Uid == "" {
		return false, errors.New("uid is empty")
	}

	return true, nil
}

func (iBS *inputBranchSalary) GetBranchSalary() *model.BranchSalary {
	loc, _ := time.LoadLocation("Asia/Taipei")
	t := iBS.Date.In(loc)
	y, m, d := t.Date()
	t = time.Date(y, m, d, 23, 59, 59, 99, loc)
	//time, _ := time.ParseInLocation("2006-01-02", iBS.Date, time.Local)
	strDate := fmt.Sprintf("%d-%02d-%02d", t.Year(), t.Month(), t.Day())
	if iBS.Branch == "ALL" || iBS.Branch == "all" {
		iBS.Branch = "%"
	}
	return &model.BranchSalary{
		Date:    t,
		Name:    iBS.Name,
		StrDate: strDate,
		Branch:  iBS.Branch,
	}
}

func (iSS *inputSalerSalary) GetSalerSalary() *model.SalerSalary {
	//time, _ := time.ParseInLocation("2006-01-02", iBS.Date, time.Local)
	return &model.SalerSalary{
		Sid:         iSS.Sid,
		Abonus:      iSS.Abonus,
		Lbonus:      iSS.Lbonus,
		Description: iSS.Description,
		Other:       iSS.Other,
		Tax:         iSS.Tax,
		Welfare:     iSS.Welfare,
		SP:          iSS.SP,
		Workday:     iSS.Workday,
		Salary:      iSS.Salary,
		LaborFee:    iSS.LaborFee,
		HealthFee:   iSS.HealthFee,
	}
}

func (iCA *inputCloseAccount) GetCloseAccount() *model.CloseAccount {
	//time, _ := time.ParseInLocation("2006-01-02", iBS.Date, time.Local)
	return &model.CloseAccount{
		CloseDate: iCA.CloseDate,
		Uid:       iCA.Uid,
	}
}

func (iIE *inputIncomeExpense) GetIncomeExpense() *model.IncomeExpense {
	//time, _ := time.ParseInLocation("2006-01-02", iBS.Date, time.Local)
	//var e model.Expense{SalerFee:iIE.SalerFee}

	return &model.IncomeExpense{
		EarnAdjust: iIE.EarnAdjust,

		Expense: model.Expense{
			SalerFee:    iIE.SalerFee,
			AnnualRatio: iIE.AnnualRatio,
		},
	}
}

func ReceiveFile(w http.ResponseWriter, r *http.Request, filename string) {

	fmt.Println(util.PdfDir + filename)

	// data, err := ioutil.ReadFile(util.PdfDir + filename)
	// if err != nil {
	// 	fmt.Println(err)
	// }

	// data, err := os.Open("MyPdf/a.zip")
	// if err != nil {
	// 	return
	// }
	// defer data.Close()

	// var zipFileStat os.FileInfo
	// zipFileStat, err = data.Stat()
	// if err != nil {
	// 	return
	// }

	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	http.ServeFile(w, r, util.PdfDir+filename)
	//w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
	//Set header
	//w.Header().Set("Content-type", "application/zip")
	w.Header().Set("Content-type", "application/x-download")
	//w.Header().Set("Content-Length", strconv.FormatInt(zipFileStat.Size(), 10))
	//io.Copy(w, data)

	//w.Write(data)

}

func (api *SalaryAPI) updateSalerSalaryEndpoint(w http.ResponseWriter, req *http.Request) {
	//Get params from body
	vars := util.GetPathVars(req, []string{"bsID"})
	bsID := vars["bsID"].(string)
	dbname := req.Header.Get("dbname")
	iSS := inputSalerSalary{}
	err := json.NewDecoder(req.Body).Decode(&iSS)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid JSON format"))
		return
	}

	if ok, err := iSS.inputSalerSalaryValid(); !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	SalaryM := model.GetSalaryModel(di)

	err = SalaryM.UpdateSalerSalaryData(iSS.GetSalerSalary(), bsID, dbname)
	if err != nil {
		if strings.Contains(err.Error(), ERROR_CloseDate) {
			w.WriteHeader(http.StatusLocked)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		w.Write([]byte("Error" + err.Error()))
	} else {
		w.Write([]byte("OK"))
	}

}

func (api *SalaryAPI) updateManagerBonusEndpoint(w http.ResponseWriter, req *http.Request) {
	//Get params from body
	vars := util.GetPathVars(req, []string{"bsID"})
	bsID := vars["bsID"].(string)
	dbname := req.Header.Get("dbname")
	iIE := inputIncomeExpense{}
	err := json.NewDecoder(req.Body).Decode(&iIE)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid JSON format"))
		return
	}

	if ok, err := iIE.inpuIncomeExpenseValid(); !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	SalaryM := model.GetSalaryModel(di)

	err = SalaryM.UpdateIncomeExpenseData(iIE.GetIncomeExpense(), bsID, dbname)
	if err != nil {
		if strings.Contains(err.Error(), ERROR_CloseDate) {
			w.WriteHeader(http.StatusLocked)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		w.Write([]byte("Error" + err.Error()))
	} else {
		w.Write([]byte("OK"))
	}

}

func (api *SalaryAPI) closeAccountSettlementEndpoint(w http.ResponseWriter, req *http.Request) {
	per := req.Header.Get("AuthPerm")
	dbname := req.Header.Get("dbname")
	fmt.Println(per)
	//Get params from body
	iCA := inputCloseAccount{}
	err := json.NewDecoder(req.Body).Decode(&iCA)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid JSON format"))
		return
	}

	if ok, err := iCA.inputCloseAccountValid(); !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	SalaryM := model.GetSalaryModel(di)

	err = SalaryM.CloseAccountSettlement(iCA.GetCloseAccount(), per, dbname)
	if err != nil {
		if strings.Contains(err.Error(), ERROR_CloseDate) {
			w.WriteHeader(http.StatusLocked)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		w.Write([]byte("[Error]" + err.Error()))
	} else {
		w.Write([]byte("OK"))
	}

}

func (api *SalaryAPI) getCloseAccountSettlementEndpoint(w http.ResponseWriter, req *http.Request) {
	dbname := req.Header.Get("dbname")
	SalaryM := model.GetSalaryModel(di)

	SalaryM.GetAccountSettlement(dbname, nil)
	//data, err := json.Marshal(result)
	data, err := SalaryM.Json("AccountSettlement")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (api *SalaryAPI) deleteCloseAccountSettlementEndpoint(w http.ResponseWriter, req *http.Request) {

	SalaryM := model.GetSalaryModel(di)
	dbname := req.Header.Get("dbname")
	SalaryM.DeleteAccountSettlement(dbname)

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("OK"))
}
