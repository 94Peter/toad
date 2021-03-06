package middle

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"toad/resource"
	"toad/util"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
)

type AuthMiddle bool

const (
	authValue       = uint8(1 << iota)
	allowMultiValue = uint8(1 << iota)
)

var (
	authMap  map[string]uint8    = make(map[string]uint8)
	groupMap map[string][]string = make(map[string][]string)
	myDI                         = resource.GetConf("dev", os.Getenv("TIMEZONE"))
)

func AddAuthPath(path string, auth bool, group []string) {
	value := uint8(0)
	if auth {
		value = value | authValue
	}
	authMap[path] = uint8(value)
	groupMap[path] = group
}

func isAuth(path string, method string) bool {
	key := fmt.Sprintf("%s:%s", path, method)
	value, ok := authMap[key]

	if ok {
		return (value & authValue) > 0
	}
	return false
}

func hasPerm(path string, method string, perm string) bool {
	key := fmt.Sprintf("%s:%s", path, method)
	value, ok := groupMap[key]
	if len(value) == 0 {
		return true
	}
	if ok && util.IsStrInList(perm, value...) {
		return true
	}
	return false
}

func (am AuthMiddle) Enable() bool {
	return bool(am)
}

func (am AuthMiddle) GetMiddleWare() func(f http.HandlerFunc) http.HandlerFunc {
	return func(f http.HandlerFunc) http.HandlerFunc {
		// one time scope setup area for middleware
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Auth-Token, Authorization, Token")
			fmt.Println("GetMiddleWare")
			// ... pre handler functionality
			path, err := mux.CurrentRoute(r).GetPathTemplate()
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(err.Error()))
				return
			}
			auth := isAuth(path, r.Method)
			//fmt.Println(auth)
			if auth {
				authToken := r.Header.Get("Auth-Token")
				if authToken == "" {
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte("miss token"))
					return
				}

				jwtToken, err := di.GetJWTConf().Parse(authToken)
				if err != nil {
					fmt.Println("err:", err.Error())
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte(err.Error()))
					return
				}

				out, err := json.Marshal(jwtToken)
				if err != nil {
					panic(err)
				}
				fmt.Println("auth middle:", string(out))

				mapClaims := jwtToken.Claims.(jwt.MapClaims)

				permission := mapClaims["per"].(string)

				if hasPerm := hasPerm(path, r.Method, permission); !hasPerm {
					fmt.Println("permission error")
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte("permission error"))
					return
				}

				r.Header.Set("isLogin", "true")
				r.Header.Set("AuthAccount", mapClaims["sub"].(string))
				r.Header.Set("AuthName", mapClaims["nam"].(string))
				r.Header.Set("AuthPerm", permission)
				r.Header.Set("AuthCategory", mapClaims["cat"].(string))
				r.Header.Set("dbname", mapClaims["dbname"].(string))
				if mapClaims["dbname"].(string) == "" {
					fmt.Println("auth_middle dbname error")
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte("dbname error"))
					return
				}
			}
			f(w, r)
		}
	}
}
