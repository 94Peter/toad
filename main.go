package main

// [START import]
import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/94peter/toad/model"
	"github.com/94peter/toad/resource/sms"
	"github.com/94peter/toad/util"

	"dforcepro.com/resource"
	"dforcepro.com/resource/logger"
	"github.com/gorilla/mux"
	yaml "gopkg.in/yaml.v2"

	mapi "github.com/94peter/toad/api"
	"github.com/94peter/toad/middle"
	mdb "github.com/94peter/toad/resource/db"
)

// [END import]
// [START main_func]

func main() {

	// smtpHost := "smtp.gmail.com"       // 你可以改为其他的
	// smtpPort := "587"                  // 端口
	// smtpPass := "nqnbzmrmywrtvyyv"     // 密码
	// smtpUser := "crgcrg0034@gmail.com" // 用户
	// subject := "test"
	// body := "testbody"
	// fname := "hello.pdf"
	// util.RunSendMail(smtpHost, smtpPort, smtpPass, smtpUser, "geassyayaoo3@gmail.com", subject, body, fname)

	//excel.GetExcel()

	// //f := excelize.NewFile()
	// ex := excel.GetExcel()
	// data := excel.GetDataTable(2)
	// data.RawData["A10"] = "ds"
	// ex.FillText(data)
	// // Create a new sheet.

	// // index := f.NewSheet("薪轉戶")
	// // f.DeleteSheet("Sheet1")
	// // f.SetCellValue("薪轉戶", "B2", 100)
	// // f.SetColWidth("薪轉戶", "C", "C", 35)
	// // f.SetColWidth("薪轉戶", "D", "E", 60)

	// // // Set active sheet of the workbook.
	// // f.SetActiveSheet(index)
	// // // Save xlsx file by the given path.
	// fakeId := fmt.Sprintf("%d", time.Now().Unix())

	// err := ex.File.SaveAs("./Book" + fakeId + ".xlsx")
	// if err != nil {
	// 	fmt.Println(err)
	// }

	//return

	// const dir = "pdf/"
	// //獲取原始檔列表
	// f, err := ioutil.ReadDir(dir)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// fzip, _ := os.Create("img-50.zip")
	// w := zip.NewWriter(fzip)
	// defer w.Close()
	// for _, file := range f {
	// 	fw, _ := w.Create(file.Name())
	// 	filecontent, err := ioutil.ReadFile(dir + file.Name())
	// 	if err != nil {
	// 		fmt.Println(err)
	// 	}
	// 	n, err := fw.Write(filecontent)
	// 	if err != nil {
	// 		fmt.Println(err)
	// 	}
	// 	fmt.Println(n)
	// }

	// fmt.Println(pdf.SayHelloTo("f"))
	// p := pdf.GetNewPDF()
	// p.DrawPDF(pdf.GetDataTable(""))
	//return

	// [START setting_port]
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	env := os.Getenv("ENV")
	if env == "" {
		env = "dev"
		log.Printf("Defaulting to ENV %s", env)
	}

	timezone := os.Getenv("TIMEZONE")
	if timezone == "" {
		timezone = "Asia/Taipei"
		log.Printf("Defaulting to timezone %s", timezone)
	}

	myDI := GetConf(env, timezone)

	router := mux.NewRouter()
	//middleConf := di.APIConf.Middle
	middle.SetDI(myDI)
	middlewares := middle.GetMiddlewares(
		// middle.DBMiddle(true),
		middle.DebugMiddle(true),
		middle.AuthMiddle(true),
		// middle.BasicAuthMiddle(middleConf.Auth),
	)

	apiConf := &mapi.APIconf{Router: router, MiddleWares: middlewares}
	mapi.SetDI(myDI)
	mapi.InitAPI(
		apiConf,
		mapi.AdminAPI(true),
		mapi.ARAPI(true),
		mapi.ReceiptAPI(true),
		mapi.DeductAPI(true),
		mapi.CommissionAPI(true),
		mapi.AmortizationAPI(true),
		mapi.PrePayAPI(true),
		mapi.PocketAPI(true),
		mapi.ConfigAPI(true),
		mapi.SalaryAPI(true),
		mapi.SystemAPI(true),
		mapi.IndexAPI(true),
	)
	initBranch(myDI)
	log.Printf("Listening on port %s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), router))
	// [END setting_port]
}

type di struct {
	// Mongodb  *db.Mongo         `yaml:"mongodb,omitempty"`
	// Redis    *db.Redis         `yaml:"redis,omitempty"`
	// Elastic  *db.Elastic       `yaml:"elastic,omitempty"`
	DBconf  *mdb.DBConf      `yaml:"dbConf"`
	Log     logger.Logger    `yaml:"log,omitempty"`
	APIConf resource.APIConf `yaml:"api,omitempty"`
	// SysMap   map[string]string `yaml:"sysMap,omitempty"`
	Location *time.Location `yaml:"-"`
	SMSConf  *sms.SMSConf   `yaml:"smsConf"`
	JWTConf  *util.JwtConf  `yaml:"jwtConf"`
	LoginURL string         `yaml:"loginURL"`
	Init     struct {
		Email string `yaml:"email"`
		Name  string `yaml:"name"`
		Phone string `yaml:"phone"`
	} `yaml:"init"`

	SMTPConf util.SendMail `yaml:"smtpConf"`
}

func (d *di) GetLoginURL() string {
	return d.LoginURL
}

func (d *di) GetLog() logger.Logger {
	return d.Log
}

func (d *di) GetAPIConf() resource.APIConf {
	return d.APIConf
}

func (d *di) GetSMS() sms.InterSMS {
	return d.SMSConf.GetSMS()
}

func (d *di) GetJWTConf() *util.JwtConf {
	return d.JWTConf
}

func (d *di) GetSMTPConf() util.SendMail {
	return d.SMTPConf
}

func (d *di) GetLocation() *time.Location {
	return d.Location
}

func (d *di) GetSQLDB() mdb.InterSQLDB {
	return d.DBconf.GetSQLDB()
}

func (d *di) GetDB() mdb.InterDB {
	return d.DBconf.GetDB()
}

// 初始化設定檔，讀YAML檔
func GetConf(env string, timezone string) *di {

	const confFileTpl = "conf/%s/config.yml"

	yamlFile, err := ioutil.ReadFile(fmt.Sprintf(confFileTpl, env))
	if err != nil {
		panic(err)
	}
	myDI := di{}
	err = yaml.Unmarshal(yamlFile, &myDI)
	if err != nil {
		panic(err)
	}
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		panic(err)
	}

	myDI.Location = loc
	myDI.Log.StartLog()
	myDI.GetSQLDB() //for quickly test DB

	// var queryDate time.Time
	// today := time.Date(queryDate.Year(), queryDate.Month(), 1, 0, 0, 0, 0, queryDate.Location())
	// end := time.Date(queryDate.Year(), queryDate.Month(), 1, 0, 0, 0, 0, queryDate.Location())
	// var result *[]_model.AR
	// mm := _model.GetARModel.Get

	return &myDI
}

func initBranch(myDI *di) {
	systemM := model.GetSystemModel(myDI)
	configM := model.GetConfigModel(myDI)
	branchbyte, err := systemM.GetBranchData()
	if err != nil {
		fmt.Println(err)
		return
	}

	branchList := []string{}
	s := strings.Split(string(branchbyte), "\"")
	for _, each := range s {
		fmt.Println(each)
		if each != "," && each != "[" && each != "]" {
			branchList = append(branchList, each)
		}
	}
	fmt.Println(branchList)

	configM.CreateConfigBranch(branchList)
}

// [END main_func]

// [START indexHandler]

// indexHandler responds to requests with our greeting.
// func indexHandler(w http.ResponseWriter, r *http.Request) {
// 	if r.URL.Path != "/" {
// 		http.NotFound(w, r)
// 		return
// 	}
// 	fmt.Fprint(w, "Hello, World!")
// }
