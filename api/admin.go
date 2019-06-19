package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/94peter/pica/model"
	"github.com/94peter/pica/permission"
	"github.com/94peter/pica/util"
)

type AdminAPI bool

func (api AdminAPI) Enable() bool {
	return bool(api)
}

func (api AdminAPI) GetAPIs() *[]*APIHandler {
	return &[]*APIHandler{
		&APIHandler{Path: "/v1/category", Next: api.getCategoryEndpoint, Method: "GET", Auth: true, Group: permission.All},
		&APIHandler{Path: "/v1/category", Next: api.createCategoryEndpoint, Method: "POST", Auth: true, Group: permission.Backend},
		&APIHandler{Path: "/v1/category/{NAME}", Next: api.deleteCategoryEndpoint, Method: "DELETE", Auth: true, Group: permission.Backend},
		&APIHandler{Path: "/v1/user", Next: api.createUserEndpoint, Method: "POST", Auth: true, Group: permission.Backend},
		&APIHandler{Path: "/v1/user", Next: api.getUserEndpoint, Method: "GET", Auth: true, Group: permission.Backend},
		&APIHandler{Path: "/v1/user/category", Next: api.updateUserCategoryEndpoint, Method: "PUT", Auth: true, Group: permission.Backend},
		&APIHandler{Path: "/v1/user/permission", Next: api.updateUserPemissionEndpoint, Method: "PUT", Auth: true, Group: permission.Backend},
		&APIHandler{Path: "/v1/user/{PHONE}", Next: api.deleteUserEndpoint, Method: "DELETE", Auth: true, Group: permission.Backend},
	}
}

type inputUser struct {
	Name       string `json:"name"`
	C          string `json:"category"`
	Phone      string `json:"phone"`
	Email      string `json:"email"`
	Permission string `json:"permission"`
}

func (u *inputUser) isDeleteValid() (bool, error) {
	if u.Phone == "" {
		return false, errors.New("Phone is empty")
	}
	return true, nil
}

func (u *inputUser) isChangePermissionValid() (bool, error) {
	if u.Phone == "" {
		return false, errors.New("Phone is empty")
	}
	if !util.IsStrInList(u.Permission, permission.Manager, permission.Sales) {
		return false, errors.New("permission error")
	}
	return true, nil
}

func (u *inputUser) isChangeCategoryValid() (bool, error) {
	if u.Phone == "" {
		return false, errors.New("Phone is empty")
	}
	if u.C == "" {
		return false, errors.New("Category is empty")
	}
	return true, nil
}

func (u *inputUser) isValid() (bool, error) {
	if !util.IsStrInList(u.Permission, permission.All...) {
		return false, errors.New("permission error")
	}
	if u.Phone == "" {
		return false, errors.New("Phone is empty")
	}
	if u.Email == "" {
		return false, errors.New("Email is empty")
	}
	if u.C == "" {
		return false, errors.New("Category is empty")
	}
	return true, nil
}

func (u *inputUser) GetUser() *model.User {
	return &model.User{
		Account:    u.Phone,
		Name:       u.Name,
		Email:      u.Email,
		Phone:      u.Phone,
		Category:   u.C,
		Permission: u.Permission,
		Pwd:        model.GetRandPwd(8),
		State:      model.UserStateInit,
	}
}

func (api *AdminAPI) updateUserPemissionEndpoint(w http.ResponseWriter, req *http.Request) {
	iu := inputUser{}
	err := json.NewDecoder(req.Body).Decode(&iu)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid JSON format"))
		return
	}
	if ok, err := iu.isChangePermissionValid(); !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	memberModel := model.GetMemberModel(di)
	if err := memberModel.ChangePermission(iu.Phone, iu.Permission); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Write([]byte("ok"))
}

func (api *AdminAPI) updateUserCategoryEndpoint(w http.ResponseWriter, req *http.Request) {
	iu := inputUser{}
	err := json.NewDecoder(req.Body).Decode(&iu)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid JSON format"))
		return
	}
	if ok, err := iu.isChangeCategoryValid(); !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	memberModel := model.GetMemberModel(di)
	if err := memberModel.ChangeCategory(iu.Phone, iu.C); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Write([]byte("ok"))
}

func (api *AdminAPI) deleteUserEndpoint(w http.ResponseWriter, req *http.Request) {
	vars := util.GetPathVars(req, []string{"PHONE"})
	phone := vars["PHONE"].(string)
	memberModel := model.GetMemberModel(di)
	if user := memberModel.GetMember(phone); user == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if err := memberModel.Quit(phone); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Write([]byte("ok"))
}

func (api *AdminAPI) createUserEndpoint(w http.ResponseWriter, req *http.Request) {
	iu := inputUser{}
	err := json.NewDecoder(req.Body).Decode(&iu)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid JSON format"))
		return
	}
	if ok, err := iu.isValid(); !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	memberModel := model.GetMemberModel(di)

	if err := memberModel.CreateUser(iu.GetUser()); err != nil {
		if strings.Contains(err.Error(), "account exist") {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
	}
	w.Write([]byte("ok"))
}

func (api *AdminAPI) getUserEndpoint(w http.ResponseWriter, req *http.Request) {
	queryVar := util.GetQueryValue(req, []string{"c"}, true)
	category := (*queryVar)["c"].(string)
	if category == "" {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	memberModel := model.GetMemberModel(di)
	result := memberModel.GetUserByCategory(category)
	if result == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (api *AdminAPI) createCategoryEndpoint(w http.ResponseWriter, req *http.Request) {
	data := make(map[string]interface{})
	err := json.NewDecoder(req.Body).Decode(&data)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid JSON format"))
		return
	}

	category, ok := data["name"].(string)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	categoryModel := model.GetDictionaryCategory(di.GetDB())
	ok = categoryModel.Add(category)
	if ok {
		err = categoryModel.Save()
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Write([]byte("ok"))
}

func (api *AdminAPI) getCategoryEndpoint(w http.ResponseWriter, req *http.Request) {
	categoryModel := model.GetDictionaryCategory(di.GetDB())
	if categoryModel.IsEmpty() {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	j, err := categoryModel.Json()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		di.GetLog().Err(err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(j)
}

func (api *AdminAPI) deleteCategoryEndpoint(w http.ResponseWriter, req *http.Request) {
	vars := util.GetPathVars(req, []string{"NAME"})
	name := vars["NAME"].(string)
	categoryModel := model.GetDictionaryCategory(di.GetDB())

	// 判斷類別底上是否有使用者
	mm := model.GetMemberModel(di)
	categoryUsers := mm.GetUserByCategory(name)
	if len(categoryUsers) > 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("found user in the category"))
		return
	}

	if err := categoryModel.Remove(name); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	w.Write([]byte("ok"))
}
