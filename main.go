package main

// [START import]
import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/94peter/toad/resource"

	"github.com/gorilla/mux"

	mapi "github.com/94peter/toad/api"
	"github.com/94peter/toad/middle"
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

	log.Printf("Listening on port %s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), router))
	// [END setting_port]
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
