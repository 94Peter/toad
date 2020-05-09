package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"

	"github.com/94peter/toad/model"
	"github.com/94peter/toad/permission"
	"github.com/94peter/toad/util"
)

type AdminAPI bool

func (api AdminAPI) Enable() bool {
	return bool(api)
}

func (api AdminAPI) GetAPIs() *[]*APIHandler {
	return &[]*APIHandler{
		&APIHandler{Path: "/v1/token", Next: api.tokenEndpoint, Method: "GET", Auth: false},

		&APIHandler{Path: "/v1/refreshToken", Next: api.RefreshTokenEndpoint, Method: "GET", Auth: true},

		&APIHandler{Path: "/v1/category", Next: api.getCategoryEndpoint, Method: "GET", Auth: true, Group: permission.All},
		//&APIHandler{Path: "/v1/category", Next: api.t, Method: "POST", Auth: false, Group: permission.All},

		&APIHandler{Path: "/v1/user", Next: api.getUserEndPoint, Method: "GET", Auth: true, Group: permission.Backend},
		&APIHandler{Path: "/v1/user", Next: api.createUser, Method: "POST", Auth: true, Group: permission.Backend},
		&APIHandler{Path: "/v1/user/{ID}", Next: api.deleteUserEndPoint, Method: "DELETE", Auth: true, Group: permission.Backend},
		&APIHandler{Path: "/v1/user", Next: api.updateUserEndPoint, Method: "PUT", Auth: true, Group: permission.Backend},

		&APIHandler{Path: "/v1/user/pwd", Next: api.updatePwdEndPoint, Method: "PUT", Auth: true, Group: permission.Backend},
		//&APIHandler{Path: "/v1/user/pwd/{Email}", Next: api.resetPwdEndPoint, Method: "POST", Auth: false, Group: permission.All}, not work
		&APIHandler{Path: "/v1/user/disable", Next: api.disableUserEndPoint, Method: "PUT", Auth: true, Group: permission.Backend},
		&APIHandler{Path: "/v1/user/state", Next: api.updateStateEndPoint, Method: "PUT", Auth: true, Group: permission.Backend},
	}
}

var auth_token = "eyJhbGciOiJSUzI1NiIsImtpZCI6ImRldiIsInR5cCI6IkpXVCJ9.eyJleHAiOjEwNzkzOTAyMzIxLCJpYXQiOjE1NzA1MzAyODQsImlzcyI6InBpY2Fpc3MiLCJzeXMiOiJ0b2FkIn0.dCeCH2cYCm5MewP2lCpLGJV4ka4C8j4joHL23YlphRQJpOemKBRLReCXKFQh1GhdnFKXh6xh9ULox_BUBZxckdRDoJo5-R7fXM7eOy5hIRFyOwO8FOuKJ50QddR0qoLbuLbzIklJncxDRftBcujuOFFAFEBIkR5Nq9TyBEgIkSI"

type inputUser struct {
	Account    string `json:"account"`
	Name       string `json:"name"`
	Permission string `json:"permission"`
	Password   string `json:"password"`
}

type inputUpdateUser struct {
	Account    string `json:"account"`
	Name       string `json:"name"`
	Permission string `json:"permission"`
}

type inputPwd struct {
	Account  string `json:"account"`
	Password string `json:"password"`
}

type inputDisable struct {
	Account string `json:"account"`
	Disable bool   `json:"disable"`
}

type inputState struct {
	Account string `json:"account"`
	State   string `json:"state"`
}

func (api *AdminAPI) getUserEndPoint(w http.ResponseWriter, req *http.Request) {
	memM := model.GetMemberModel(di)

	//data, err := json.Marshal(result)
	data, err := memM.GetAccountUserData()

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	r, err := json.Marshal(data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Write(r)
}

func (api *AdminAPI) createUser(w http.ResponseWriter, req *http.Request) {
	// 取得IP
	ip, _, _ := net.SplitHostPort(req.RemoteAddr)
	// 印出IP
	fmt.Println(ip + "\n\n")

	user := inputUser{}
	err := json.NewDecoder(req.Body).Decode(&user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid JSON format"))
		return
	}
	if ok, err := user.isUserValid(); !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	memM := model.GetMemberModel(di)
	err = memM.CreateUser(user.GetUser())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	} else {
		w.Write([]byte("OK"))
	}
}

func (api *AdminAPI) deleteUserEndPoint(w http.ResponseWriter, req *http.Request) {

	vars := util.GetPathVars(req, []string{"ID"})
	ID := vars["ID"].(string)

	memM := model.GetMemberModel(di)

	if err := memM.DeleteUser(ID); err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(err.Error()))
		return
	}

	w.Write([]byte("ok"))
}

func (api *AdminAPI) updateUserEndPoint(w http.ResponseWriter, req *http.Request) {

	user := inputUpdateUser{}
	err := json.NewDecoder(req.Body).Decode(&user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid JSON format"))
		return
	}
	if ok, err := user.isUserValid(); !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	memM := model.GetMemberModel(di)
	if err := memM.UpdateUser(user.GetUser()); err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(err.Error()))
		return
	}

	w.Write([]byte("ok"))
}

func (api *AdminAPI) updatePwdEndPoint(w http.ResponseWriter, req *http.Request) {

	user := inputPwd{}
	err := json.NewDecoder(req.Body).Decode(&user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid JSON format"))
		return
	}

	memM := model.GetMemberModel(di)
	if err := memM.ChangePwd(user.Account, user.Password); err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(err.Error()))
		return
	}

	w.Write([]byte("ok"))
}
func (api *AdminAPI) resetPwdEndPoint(w http.ResponseWriter, req *http.Request) {

	vars := util.GetPathVars(req, []string{"Email"})
	Email := vars["Email"].(string)

	memM := model.GetMemberModel(di)

	if err := memM.ResetPwd(Email); err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(err.Error()))
		return
	}

	w.Write([]byte("ok"))
}
func (api *AdminAPI) disableUserEndPoint(w http.ResponseWriter, req *http.Request) {

	user := inputDisable{}
	err := json.NewDecoder(req.Body).Decode(&user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid JSON format"))
		return
	}

	memM := model.GetMemberModel(di)
	if err := memM.SetUserDisable(user.Account, user.Disable); err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(err.Error()))
		return
	}

	w.Write([]byte("ok"))
}

func (api *AdminAPI) updateStateEndPoint(w http.ResponseWriter, req *http.Request) {

	user := inputState{}
	err := json.NewDecoder(req.Body).Decode(&user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid JSON format"))
		return
	}

	memM := model.GetMemberModel(di)
	if err := memM.UpdateState(user.Account, user.State); err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(err.Error()))
		return
	}

	w.Write([]byte("ok"))
}

// 交換firebase token to pica token
func (api *AdminAPI) tokenEndpoint(w http.ResponseWriter, req *http.Request) {
	//ftoken := req.Header.Get("Auth-Token")
	ftoken := req.Header.Get("token")

	if ftoken == "" {

		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	mm := model.GetMemberModel(di)
	user := mm.VerifyToken(ftoken)
	if user == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	token, err := user.GetToken(di.GetJWTConf())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		di.GetLog().Err(err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"token":      token,
		"state":      user.State,
		"permission": user.Permission,
	})

}

// test
func (api *AdminAPI) RefreshTokenEndpoint(w http.ResponseWriter, req *http.Request) {
	//ftoken := req.Header.Get("Auth-Token")

	user := model.User{
		Account: req.Header.Get("AuthAccount"),
	}

	token, err := user.GetToken(di.GetJWTConf())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		di.GetLog().Err(err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"token": token,
	})
}

func (api *AdminAPI) getCategoryEndpoint(w http.ResponseWriter, req *http.Request) {
	//db := di.GetSQLDB()
	//db.Query("select * from public.ab")
	//isDB := db.InitDB()

	client := &http.Client{}

	req, err := http.NewRequest("GET", "https://pica957.appspot.com/v1/toad/category", nil)
	if err != nil {
		// handle error
		fmt.Println(err)
	}

	req.Header.Set("auth-token", auth_token)

	resp, err := client.Do(req)
	if err != nil {
		// handle error
		fmt.Println(err)
	}
	defer resp.Body.Close()

	sitemap, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
		return
	}

	fmt.Println(string(sitemap))

	w.Write(sitemap)
}

func (u *inputUser) isUserValid() (bool, error) {

	if u.Account == "" {
		return false, errors.New("account is empty")
	}

	if u.Password == "" {
		return false, errors.New("password is empty")
	}

	if u.Name == "" {
		return false, errors.New("name is empty")
	}

	err := permissionCheck(u.Permission)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (u *inputUpdateUser) isUserValid() (bool, error) {

	if u.Account == "" {
		return false, errors.New("account is empty")
	}

	if u.Name == "" {
		return false, errors.New("name is empty")
	}

	err := permissionCheck(u.Permission)
	if err != nil {
		return false, err
	}

	return true, nil
}

func permissionCheck(perm string) error {
	if perm == permission.Office {
		return nil
	}
	if perm == permission.Manager {
		return nil
	}
	if perm == permission.Admin {
		return nil
	}
	if perm == "" {
		return errors.New("permission is empty")
	}
	return errors.New(perm + " permission in unknown.")
}

func (user *inputUser) GetUser() *model.User {

	return &model.User{
		Password:   user.Password,
		Permission: user.Permission,
		Account:    user.Account,
		Name:       user.Name,
	}
}

func (user *inputUpdateUser) GetUser() *model.User {

	return &model.User{
		Permission: user.Permission,
		Account:    user.Account,
		Name:       user.Name,
	}
}

/*
DELETE FROM public.ar;
DELETE FROM public.armap;
DELETE FROM public.branchsalary;
DELETE FROM public.salersalary;
DELETE FROM public.incomeexpense;
DELETE FROM public.receipt;
DELETE FROM public.commission;
*/
