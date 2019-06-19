package model

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/94peter/pica/permission"
	"github.com/94peter/pica/resource/db"
	"github.com/94peter/pica/resource/sms"
	"github.com/94peter/pica/util"

	"errors"
)

const (
	categoryC = "category"
	userC     = "user"

	UserPerSales = "sales"

	UserStateInit   = "init"
	UserStateNormal = "normal"
	UserStateReset  = "reset"
)

type interModelRes interface {
	GetDB() db.InterDB
	GetAuth() db.InterAuth
	GetSMS() sms.InterSMS
	GetLoginURL() string
	GetTSDB() db.InterTSDB
	GetTrendItems() []string
	GetLocation() *time.Location
}

var (
	categoryModel *dictionaryCategory
)

func GetDictionaryCategory(db db.InterDB) *dictionaryCategory {
	if categoryModel != nil {
		return categoryModel
	}
	categoryModel = &dictionaryCategory{
		db: db,
	}
	categoryModel.load()
	return categoryModel
}

func GetMemberModel(mr interModelRes) *memberModel {
	cu := &categoryUser{
		db: mr.GetDB(),
	}
	cu.load()
	return &memberModel{
		auth:     mr.GetAuth(),
		sms:      mr.GetSMS(),
		loginURL: mr.GetLoginURL(),

		cu: cu,
	}
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func GetRandPwd(n int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

type User struct {
	Account     string
	Pwd         string
	Name        string
	Email       string
	Phone       string
	Category    string
	Stock       float32
	Performance float32
	Permission  string
	State       string
	Medal       [4]int
}

func (u *User) GetMedalToString() string {
	var result string
	for i, m := range u.Medal {
		if i == 0 {
			result = strconv.Itoa(m)
		}
		result = util.StrAppend(result, ",", strconv.Itoa(m))
	}
	return result
}

func (u *User) GetToken(jwtConf *util.JwtConf) (string, error) {
	token, err := jwtConf.GetToken(map[string]interface{}{
		"sub": u.Account,
		"nam": u.Name,
		"per": u.Permission,
		"cat": u.Category,
	})
	if err != nil {
		return "", err
	}
	return *token, nil
}

func (u *User) IsValid() bool {
	if !util.IsStrInList(u.Permission, permission.All...) {
		return false
	}
	if u.Phone == "" || u.Email == "" || u.Category == "" {
		return false
	}
	return true
}

func (u *User) ToDisplay() map[string]interface{} {
	return map[string]interface{}{
		"account":    u.Account,
		"name":       u.Name,
		"phone":      u.Phone,
		"email":      u.Email,
		"permission": u.Permission,
	}
}

type categoryUser struct {
	db             db.InterDB
	DictionaryUser map[string]*User   `json:"-"`
	CategoryUsers  map[string][]*User `json:"c"`
}

func (dc *categoryUser) isExist(account string) bool {
	if dc.DictionaryUser == nil {
		return false
	}
	_, ok := dc.DictionaryUser[account]
	return ok
}

func (dc *categoryUser) get(acc string) *User {
	return dc.DictionaryUser[acc]
}

func (dc *categoryUser) add(u *User) bool {
	if dc.DictionaryUser == nil {
		dc.DictionaryUser = make(map[string]*User)
	}
	if dc.CategoryUsers == nil {
		dc.CategoryUsers = make(map[string][]*User)
	}
	if dc.isExist(u.Account) {
		return false
	}
	dc.DictionaryUser[u.Account] = u
	ulist, _ := dc.CategoryUsers[u.Category]
	ulist = append(ulist, u)
	dc.CategoryUsers[u.Category] = ulist
	return true
}

func (dc *categoryUser) changePermission(acc, newPermission string) error {
	if !util.IsStrInList(newPermission, permission.All...) {
		return errors.New("permission setting error")
	}
	u := dc.get(acc)
	if u == nil {
		return errors.New("user not exist")
	}
	u.Permission = newPermission
	return dc.save()
}

func (dc *categoryUser) changeCategory(acc, newCategory string) error {
	cm := GetDictionaryCategory(dc.db)
	if !cm.isExist(newCategory) {
		return errors.New("category not exist")
	}
	u := dc.get(acc)
	if u == nil {
		return errors.New("user not exist")
	}
	users := dc.CategoryUsers[u.Category]
	userLen := len(users)
	// remove user from origin categry list
	for i := 0; i < userLen; i++ {
		if users[i].Account == u.Account {
			users = append(users[:i], users[i+1:]...)
			dc.CategoryUsers[u.Category] = users
			break
		}
	}
	// change user category
	u.Category = newCategory
	dc.DictionaryUser[u.Account] = u
	// add user to new category
	users = dc.CategoryUsers[u.Category]
	users = append(users, u)
	dc.CategoryUsers[u.Category] = users
	return dc.save()
}

func (dc *categoryUser) update(u *User) error {
	users := dc.CategoryUsers[u.Category]
	userLen := len(users)
	for i := 0; i < userLen; i++ {
		if users[i].Account == u.Account {
			users[i] = u
			dc.CategoryUsers[u.Category] = users
			return dc.save()
		}
	}
	return nil
}

func (dc *categoryUser) remove(u *User) error {
	if dc.DictionaryUser == nil || dc.CategoryUsers == nil || !dc.isExist(u.Account) {
		return nil
	}
	obj, _ := dc.DictionaryUser[u.Account]

	order := dc.CategoryUsers[u.Category]
	l := len(order)
	var i int
	for i = 0; i < l; i++ {
		if order[i].Account == obj.Account {
			break
		}
	}
	if i < l {
		order = append(order[:i], order[i+1:]...)
		dc.CategoryUsers[u.Category] = order
	}

	delete(dc.DictionaryUser, u.Account)
	return dc.save()
}

func (dc *categoryUser) json() ([]byte, error) {
	return json.Marshal(dc.DictionaryUser)
}

func (dc *categoryUser) GetID() string {
	const id = "1"
	return id
}

func (dc *categoryUser) save() error {
	if dc.db == nil {
		return errors.New("db not set")
	}
	return dc.db.C(userC).Save(dc)
}

func (dc *categoryUser) load() error {
	if dc.db == nil {
		return errors.New("db not set")
	}
	err := dc.db.C(userC).GetByID(dc.GetID(), dc)
	if err != nil {
		return err
	}
	dc.DictionaryUser = make(map[string]*User)
	for _, s := range dc.CategoryUsers {
		for _, u := range s {
			dc.DictionaryUser[u.Account] = u
		}
	}
	return nil
}

type memberModel struct {
	auth     db.InterAuth
	sms      sms.InterSMS
	loginURL string

	cu *categoryUser
}

func (mm *memberModel) GetCategoryStock(c string) float32 {
	totalStock := float32(0.0)
	users, ok := mm.cu.CategoryUsers[c]
	if !ok {
		return totalStock
	}
	for _, u := range users {
		totalStock = totalStock + u.Stock
	}
	return totalStock
}

func (mm *memberModel) GetCategoryMaxPerformace(c string) float32 {
	max := float32(0.0)
	users, ok := mm.cu.CategoryUsers[c]
	if !ok {
		return max
	}
	for _, u := range users {
		if u.Performance > max {
			max = u.Performance
		}
	}
	return max
}

// 離職
func (mm *memberModel) Quit(acc string) error {
	u := mm.cu.get(acc)
	if u == nil {
		return errors.New("user not exist")
	}

	if err := mm.cu.remove(u); err != nil {
		return err
	}
	return mm.auth.DeleteUser(acc)
}

func (mm *memberModel) ChangePermission(acc, permission string) error {
	return mm.cu.changePermission(acc, permission)
}

func (mm *memberModel) ChangeCategory(account, newCategory string) error {
	return mm.cu.changeCategory(account, newCategory)
}

func (mm *memberModel) ChangePwd(account, pwd string) error {
	return mm.auth.ChangePwd(account, pwd)
}

func (mm *memberModel) UpdateUser(u *User) error {
	return mm.cu.update(u)
}

func (mm *memberModel) VerifyToken(ftoken string) *User {
	uid, err := mm.auth.VerifyToken(ftoken)
	if err != nil {
		return nil
	}
	u, ok := mm.cu.DictionaryUser[uid]
	if !ok {
		return nil
	}
	return u
}

func (mm *memberModel) GetMember(account string) *User {
	u, ok := mm.cu.DictionaryUser[account]
	if !ok {
		return nil
	}
	return u
}

func (mm *memberModel) GetUserByCategory(c string) []map[string]interface{} {
	userList, ok := mm.cu.CategoryUsers[c]
	if !ok {
		return nil
	}
	size := len(userList)
	if size == 0 {
		return nil
	}
	result := make([]map[string]interface{}, size)
	for i := 0; i < size; i++ {
		result[i] = userList[i].ToDisplay()
	}
	return result
}

func (mm *memberModel) CreateUser(u *User) error {
	if mm.cu.isExist(u.Account) {
		return errors.New("account exist")
	}
	opwd := u.Pwd
	u.Pwd = util.MD5(u.Pwd)

	// 建立firebaseAuth User
	err := mm.auth.CreateUser(u.Phone, u.Name, u.Email, opwd, u.Permission)
	if err != nil {
		return err
	}

	// 存入firebaseDB
	if !mm.cu.add(u) {
		return errors.New("account exist")
	}

	err = mm.cu.save()
	if err != nil {
		return err
	}

	if mm.sms == nil {
		return nil
	}
	// 簡訊通知使用者登入
	const msgTpl = "業績管理平台帳號設定完成，您的帳號：%s，密碼：%s。登入網址為%s"
	toPhone := util.StrAppend("+886", u.Phone[1:])
	mm.sms.Message(toPhone, fmt.Sprintf(msgTpl, u.Email, opwd, mm.loginURL))
	return nil
}

type category struct {
	Name string
	Sub  *dictionaryCategory
}

type dictionaryCategory struct {
	db         db.InterDB
	Dictionary map[string]*category `json:"-"`
	Order      []*category          `json:"o"`
}

func (dc *dictionaryCategory) isExist(name string) bool {
	if dc.Dictionary == nil {
		return false
	}
	_, ok := dc.Dictionary[name]
	return ok
}

func (dc *dictionaryCategory) IsEmpty() bool {
	return len(dc.Order) == 0
}

func (dc *dictionaryCategory) Add(name string) bool {
	if dc.Dictionary == nil {
		dc.Dictionary = make(map[string]*category)
	}
	if dc.isExist(name) {
		return false
	}
	cg := category{
		Name: name,
	}
	dc.Dictionary[name] = &cg
	dc.Order = append(dc.Order, &cg)
	return true
}

func (dc *dictionaryCategory) Remove(name string) error {
	if dc.Dictionary == nil || !dc.isExist(name) {
		return errors.New("not found")
	}
	obj, _ := dc.Dictionary[name]
	l := len(dc.Order)
	var i int
	for i = 0; i < l; i++ {
		if dc.Order[i].Name == obj.Name {
			break
		}
	}
	if i < l {
		dc.Order = append(dc.Order[:i], dc.Order[i+1:]...)
	}
	delete(dc.Dictionary, name)
	return nil
}

func (dc *dictionaryCategory) Json() ([]byte, error) {
	var result []string
	for _, o := range dc.Order {
		if o == nil {
			continue
		}
		result = append(result, o.Name)
	}
	return json.Marshal(result)
}

func (dc *dictionaryCategory) GetID() string {
	const id = "1"
	return id
}

func (dc *dictionaryCategory) Save() error {
	if dc.db == nil {
		return errors.New("db not set")
	}
	return dc.db.C(categoryC).Save(dc)
}

func (dc *dictionaryCategory) load() error {
	if dc.db == nil {
		return errors.New("db not set")
	}

	err := dc.db.C(categoryC).GetByID(dc.GetID(), dc)
	if err != nil {
		return err
	}
	dc.Dictionary = make(map[string]*category)
	for _, c := range dc.Order {
		if c == nil {
			continue
		}
		dc.Dictionary[c.Name] = c
	}
	return nil
}
