package middle

import (
	"fmt"
	"net/http"

	"github.com/94peter/pica/util"
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
			// ... pre handler functionality
			path, err := mux.CurrentRoute(r).GetPathTemplate()
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(err.Error()))
				return
			}
			auth := isAuth(path, r.Method)
			if auth {
				authToken := r.Header.Get("Auth-Token")
				if authToken == "" {
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte("miss token"))
					return
				}

				jwtToken, err := di.GetJWTConf().Parse(authToken)
				if err != nil {
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte(err.Error()))
					return
				}
				mapClaims := jwtToken.Claims.(jwt.MapClaims)

				permission := mapClaims["per"].(string)
				if hasPerm := hasPerm(path, r.Method, permission); !hasPerm {
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte("permission error"))
					return
				}

				r.Header.Set("isLogin", "true")
				r.Header.Set("AuthAccount", mapClaims["sub"].(string))
				r.Header.Set("AuthName", mapClaims["nam"].(string))
				r.Header.Set("AuthPerm", permission)
				r.Header.Set("AuthCategory", mapClaims["cat"].(string))

			}
			f(w, r)
		}
	}
}
