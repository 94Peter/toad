package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/94peter/toad/util"

	"github.com/94peter/toad/resource/sms"

	"github.com/gorilla/mux"

	"dforcepro.com/resource/logger"
	"github.com/94peter/toad/middle"
)

type APIconf struct {
	Router      *mux.Router
	MiddleWares *[]middle.Middleware
}

const DateFormat = "2006-01-02"

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
	GetLog() logger.Logger
	GetLoginURL() string
	GetSMS() sms.InterSMS
	GetJWTConf() *util.JwtConf
	GetLocation() *time.Location
}

var (
	di AppRes
)

func SetDI(c AppRes) {
	di = c
}

func getLog() logger.Logger {
	return di.GetLog()
}
