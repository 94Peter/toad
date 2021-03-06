package main

// [START import]
import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"toad/model"
	"toad/resource"

	"github.com/gorilla/mux"

	mapi "toad/api"
	"toad/middle"
)

// [END import]
// [START main_func]

func main() {

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

	myDI := resource.GetConf(env, timezone)
	router := mux.NewRouter()
	//middleConf := di.APIConf.Middle
	middle.SetDI(myDI)
	middlewares := middle.GetMiddlewares(
		//middle.DBMiddle(true),
		//middle.DebugMiddle(true),
		middle.AuthMiddle(true),
		// middle.BasicAuthMiddle(middleConf.Auth),
	)

	apiConf := &mapi.APIconf{Router: router, MiddleWares: middlewares}
	mapi.SetDI(myDI)
	mapi.InitAPI(
		apiConf,
		mapi.AdminAPI(true),
		mapi.ARAPI(true),
		mapi.ReturnsAPI(true),
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
		mapi.IndexAPIV2(true),
		mapi.LogAPI(true),
		mapi.FileuploadAPI(true),
	)

	//init EventLogModel, to record event
	model.GetEventLogModel(myDI)

	// configM := model.GetConfigModel(myDI)
	// configM.WorkValidDate()
	startTimer(myDI)
	startTimerInvoice(myDI)
	log.Printf("Listening on port %s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), router))

	// [END setting_port]

}

//golang 定时器，启动的时候执行一次，以后每天晚上12点执行
func startTimer(myDI *resource.DI) {
	model.GetSalaryModel(myDI)
	const DATE_FORMAT = "2006-01-02"
	go func() {
		for {
			// 计算下一个月初
			now := time.Now()
			year, month, day := now.Date()
			nextMonth := time.Date(year, month+1, 1, 0, 0, 0, 0, time.Local)
			if day == 1 {
				fmt.Println("startTimer:", year, month, day)
				//更新業務薪資
				configM := model.GetConfigModel(myDI)
				configM.WorkValidDate("nil") //待調整
				//新增攤提費用
				amorM := model.GetAmortizationModel(myDI)
				amorM.RefreshAmortizationData("nil") //待調整
			}
			fmt.Println("WorkValidDate 距離下次執行時間:", nextMonth.Sub(now))

			t := time.NewTimer(nextMonth.Sub(now))
			<-t.C

			// 计算下一个凌晨零時
			// next := now.Add(time.Hour * 24)
			// next = time.Date(next.Year(), next.Month(), next.Day(), 0, 0, 0, 0, next.Location())
			// t := time.NewTimer(next.Sub(now))
			// <-t.C

			// timer1 := time.NewTimer(time.Second * 5)
			// <-timer1.C

		}
	}()
}

func startTimerInvoice(myDI *resource.DI) {
	ivM := model.GetInvoiceModel(myDI)
	const DATE_FORMAT = "2006-01-02"
	go func() {
		for {
			// 计算下一个月初
			now := time.Now()
			year, month, day := now.Date()
			nextDay := time.Date(year, month, day, 23, 50, 0, 0, time.Local)
			ivM.UpdateInvoiceStatus()
			if nextDay.Sub(now) > 0 {
				fmt.Println("startTimerInvoice 距離下次執行時間:", nextDay.Sub(now))
				t := time.NewTimer(nextDay.Sub(now))
				<-t.C
			} else {
				nextDay = time.Date(year, month, day+1, 23, 50, 0, 0, time.Local)
				fmt.Println("startTimerInvoice 距離下次執行時間:", nextDay.Sub(now))
				t := time.NewTimer(nextDay.Sub(now))
				<-t.C
			}
		}
	}()
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
