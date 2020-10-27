package api

import (
	"fmt"
	"net/http"
	"time"

	"toad/middle"
	"toad/resource/db"
	mlog "toad/resource/log"
	"toad/resource/sms"
	"toad/util"

	"github.com/gorilla/mux"
)

type APIconf struct {
	Router      *mux.Router
	MiddleWares *[]middle.Middleware
}

const DateFormat = "2006-01-02"
const ERROR_CloseDate = "關帳日期錯誤"

type APIHandler struct {
	Path       string
	Next       func(http.ResponseWriter, *http.Request)
	Method     string
	Auth       bool
	AllowMulti bool // 允許Token驗證就存取，不用判定裝置重覆登入使用
	System     []string
	Group      []string
}

type API interface {
	GetAPIs() *[]*APIHandler
	Enable() bool
}

func InitAPI(conf *APIconf, apis ...API) {
	for _, myapi := range apis {
		if myapi.Enable() {
			addHandler(conf, myapi.GetAPIs())
		}
	}
}

func addHandler(conf *APIconf, apiHandlers *[]*APIHandler) {
	router := conf.Router
	for _, handler := range *apiHandlers {
		middle.AddAuthPath(fmt.Sprintf("%s:%s", handler.Path, handler.Method), handler.Auth, handler.Group)
		router.HandleFunc(handler.Path, middle.BuildChain(handler.Next, *conf.MiddleWares...)).Methods(handler.Method)
		router.HandleFunc(handler.Path, middle.BuildChain(handler.Next, *conf.MiddleWares...)).Methods("OPTIONS")
	}
}

type AppRes interface {
	GetLog() *mlog.Logger
	GetLoginURL() string
	GetSMS() sms.InterSMS
	GetJWTConf() *util.JwtConf
	GetLocation() *time.Location
	GetDB() db.InterDB
	GetSQLDB() db.InterSQLDB
	GetSQLDBwithDbname(string) db.InterSQLDB
	GetSMTPConf() util.SendMail
}

var (
	di AppRes
)

func SetDI(c AppRes) {
	di = c
}

func getLog() *mlog.Logger {
	return di.GetLog()
}
