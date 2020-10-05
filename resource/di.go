package resource

import (
	"fmt"
	"io/ioutil"
	"time"

	mdb "toad/resource/db"
	mlog "toad/resource/log"
	"toad/resource/sms"
	"toad/util"

	yaml "gopkg.in/yaml.v2"
)

type APIConf struct {
	Port        string          `yaml:"port,omitempty"`
	AllowSystem string          `yaml:"allowSys,omitempty"`
	Middle      map[string]bool `yaml:"middle,omitempty"`
}

type DI struct {
	DBconf   *mdb.DBConf    `yaml:"dbConf"`
	Log      *mlog.Logger   `yaml:"log,omitempty"`
	APIConf  APIConf        `yaml:"api,omitempty"`
	Location *time.Location `yaml:"-"`
	SMSConf  *sms.SMSConf   `yaml:"smsConf"`
	JWTConf  *util.JwtConf  `yaml:"jwtConf"`
	LoginURL string         `yaml:"loginURL"`
	Init     struct {
		Email string `yaml:"email"`
		Name  string `yaml:"name"`
		Phone string `yaml:"phone"`
	} `yaml:"init"`

	SMTPConf util.SendMail `yaml:"smtpConf"`
}

func (d *DI) GetLoginURL() string {
	return d.LoginURL
}

func (d *DI) GetLog() *mlog.Logger {
	return d.Log
}

func (d *DI) GetAPIConf() APIConf {
	return d.APIConf
}

func (d *DI) GetSMS() sms.InterSMS {
	return d.SMSConf.GetSMS()
}

func (d *DI) GetJWTConf() *util.JwtConf {
	return d.JWTConf
}

func (d *DI) GetSMTPConf() util.SendMail {
	return d.SMTPConf
}

func (d *DI) GetLocation() *time.Location {
	return d.Location
}

func (d *DI) GetSQLDB() mdb.InterSQLDB {
	return d.DBconf.GetSQLDB()
}

func (d *DI) GetDB() mdb.InterDB {
	return d.DBconf.GetDB()
}

// 初始化設定檔，讀YAML檔
func GetConf(env string, timezone string) *DI {

	const confFileTpl = "conf/%s/config.yml"

	yamlFile, err := ioutil.ReadFile(fmt.Sprintf(confFileTpl, env))
	if err != nil {
		panic(err)
	}
	myDI := DI{}
	err = yaml.Unmarshal(yamlFile, &myDI)
	if err != nil {
		panic(err)
	}

	//没有 tzdata 就会从$GOROOT/中找。对于没有安装go环境的windows系统来说，就没办法通过 LoadLocation 设置时区。
	//// os.Setenv("ZONEINFO", "conf/%s/data.zip")
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		panic(err)
	}

	myDI.Location = loc

	myDI.Log.StartLog()
	myDI.GetSQLDB() //for quickly check DB schema

	// var queryDate time.Time
	// today := time.Date(queryDate.Year(), queryDate.Month(), 1, 0, 0, 0, 0, queryDate.Location())
	// end := time.Date(queryDate.Year(), queryDate.Month(), 1, 0, 0, 0, 0, queryDate.Location())
	// var result *[]_model.AR
	// mm := _model.GetARModel.Get

	return &myDI
}
