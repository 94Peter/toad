package middle

import (
	"net/http"

	"toad/resource"
	mlog "toad/resource/log"
	"toad/util"
)

type middle interface {
	Enable() bool
	GetMiddleWare() func(f http.HandlerFunc) http.HandlerFunc
}

type middleRes interface {
	GetLog() *mlog.Logger
	GetAPIConf() resource.APIConf
	GetJWTConf() *util.JwtConf
}

type Middleware func(http.HandlerFunc) http.HandlerFunc

var (
	di middleRes
)

func SetDI(c middleRes) {
	if c.GetAPIConf().AllowSystem == "" {
		panic("config file missing allowSys")
	}
	di = c
}

func getLog() *mlog.Logger {
	return di.GetLog()
}

func GetMiddlewares(middles ...middle) *[]Middleware {
	var middlewares []Middleware
	for _, m := range middles {
		if m.Enable() {
			middlewares = append(middlewares, m.GetMiddleWare())
		}
	}
	return &middlewares
}

func BuildChain(f http.HandlerFunc, m ...Middleware) http.HandlerFunc {
	// if our chain is done, use the original handlerfunc
	if len(m) == 0 {
		return f
	}
	// otherwise nest the handlerfuncs
	return m[0](BuildChain(f, m[1:len(m)]...))
}
