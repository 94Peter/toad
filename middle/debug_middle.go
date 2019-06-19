package middle

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/betacraft/yaag/middleware"
	"github.com/gorilla/mux"
)

type DebugMiddle bool

func (lm DebugMiddle) Enable() bool {
	return bool(lm)
}

func (lm DebugMiddle) GetMiddleWare() func(f http.HandlerFunc) http.HandlerFunc {
	return func(f http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Auth-Token, Authorization")
			if r.Method == "OPTIONS" {
				return
			}
			getLog().Debug("-------Debug Request-------")
			path, _ := mux.CurrentRoute(r).GetPathTemplate()
			path = fmt.Sprintf("%s,%s?%s", r.Method, r.URL.Path, r.URL.RawQuery)
			getLog().Debug("path: " + path)
			header, _ := json.Marshal(r.Header)
			getLog().Debug("header: " + string(header))
			b := middleware.ReadBody(r)
			out, _ := json.Marshal(b)
			getLog().Debug("body: " + string(out))

			start := time.Now()
			f(w, r)
			delta := time.Now().Sub(start)
			if delta.Seconds() < 3 {
				return
			} else {
				getLog().Debug("over 3 mins")
			}

			getLog().Debug("-------End Debug Request-------")
		}
	}
}
