package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/lib/pq"
)

//"github.com/94peter/pica/util"

//"golang.org/x/net/context"

//"google.golang.org/api/option"

type SqlDB struct {
	Ctx context.Context

	c        string
	Clinet   *sql.DB
	Port     int
	Dburl    string
	User     string
	Password string
	Db       string
}

// edb=# -- 依照 Session 設定
// edb=# set timezone to 'ROC';
// SET
// edb=# -- 設定某帳號預設時區
// edb=# ALTER ROLE enterprisedb SET timezone TO 'ROC';
// ALTER ROLE
// edb=# -- 設定某 Database 的預設時區
// edb=# ALTER DATABASE edb SET timezone TO 'ROC';
// ALTER DATABASE

func (sdb *SqlDB) ConnectSQLDB() (*sql.DB, error) {
	fmt.Println(fmt.Sprintf("ConnectSQLDB:[%s]", sdb.Db))
	//完整的資料格式連線如下
	var connectionString string = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", sdb.Dburl, sdb.Port, sdb.User, sdb.Password, sdb.Db)
	//var connectionString string = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=ak47 sslmode=disable", sdb.Dburl, sdb.port, sdb.User, sdb.password)
	db, err := sql.Open("postgres", connectionString)

	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	//check database connection
	err = db.Ping()
	if err != nil {
		fmt.Println("[Failed] ping:" + err.Error())

		//switch database method not found
		//error message :[pq: database "dbname" does not exist]
		foo := (strings.Index(err.Error(), "does not exist"))
		if foo > 0 {
			fmt.Println("database2 [" + sdb.Db + "] does not exist")
			var connectionString string = fmt.Sprintf("host=%s port=%d user=%s password=%s sslmode=disable", sdb.Dburl, sdb.Port, sdb.User, sdb.Password)
			db, err := sql.Open("postgres", connectionString)
			err = db.Ping()
			if err != nil {
				fmt.Println("[Failed] ping2:" + err.Error())
			} else {
				fmt.Println("無指定DB 連線成功")
				sdb.Clinet = db
				return sdb.Clinet, err
			}
		}
		return nil, err
	}
	fmt.Println(sdb.Db + " 連線成功")
	sdb.Clinet = db
	return sdb.Clinet, err
}

func (sdb *SqlDB) C(c string) InterSQLDB {
	sdb.c = c
	return sdb
}

func (sdb *SqlDB) Close() error {

	if sdb.Clinet == nil {
		return nil
	}
	fmt.Println("close db")
	return sdb.Close()
}

func (sdb *SqlDB) SQLCommand(cmd string) (*sql.Rows, error) {

	db, err := sdb.ConnectSQLDB()

	if err != nil {
		return nil, err
	}

	rows, err := db.Query(cmd)

	//fmt.Println("Query " + cmd)

	if err != nil {
		return nil, err
	}

	defer db.Close()

	if rows.Err() != nil {
		fmt.Println(rows.Err())
		return nil, rows.Err()
	}

	return rows, err
}

func (sdb *SqlDB) CreateDB() error {

	_, err := sdb.SQLCommand(fmt.Sprintf(
		"CREATE DATABASE %s "+
			"WITH "+
			"OWNER = %s "+
			"ENCODING = 'UTF8' "+
			"CONNECTION LIMIT = -1;", sdb.Db, sdb.User))

	if err != nil {
		fmt.Println("CreateDB:" + err.Error())
		return err
	}
	return nil
}

func (sdb *SqlDB) CreateDeductTable() error {

	_, err := sdb.SQLCommand(fmt.Sprintf(
		"CREATE TABLE public.Deduct "+
			"( "+
			"Did character varying(50) not NULL ,"+
			"ARid character varying(50) not NULL ,"+
			"date timestamp(0) without time zone DEFAULT NULL, "+
			"status character varying(50) DEFAULT '未支付',"+
			"item character varying(50) ,"+
			"description character varying(50) ,"+
			"Fee  integer DEFAULT 0, "+
			"Rid  character varying(50)  ,"+
			"Checknumber  character varying(50) DEFAULT '',"+
			"PRIMARY KEY (Did) "+
			") "+
			"WITH ( OIDS = FALSE);"+ //))
			"ALTER TABLE public.Deduct "+
			"OWNER to %s; ", sdb.User))

	if err != nil {
		fmt.Println("CreateDeductTable:" + err.Error())
		return err
	}
	fmt.Println("CreateDeductTable Done")
	return nil
}

func (sdb *SqlDB) CreateHouseGoTable() error {

	_, err := sdb.SQLCommand(fmt.Sprintf(
		"CREATE TABLE public.housego "+
			"( "+
			"ARid character varying(50) not NULL,"+
			"id character varying(50) not NULL,"+
			"data character varying(500) ,"+
			// "date timestamp(0) without time zone not NULL, "+
			// "cNo character varying(50) not NULL, "+
			// "caseName character varying(50) not NULL, "+
			// "type character varying(50) not NULL, "+
			// "name character varying(50) not NULL, "+
			// "amount integer not NULL, "+
			//"fee integer DEFAULT 0, "+ //應扣費用
			//"RA integer DEFAULT 0, "+ //已收金額
			//"balance integer DEFAULT 0, "+
			//"sales character varying(600), "+ //json[]
			"PRIMARY KEY (ARid) "+
			") "+
			"WITH ( OIDS = FALSE);"+ //))
			" ALTER TABLE public.housego "+
			"OWNER to %s; ", sdb.User))
	//"alter table public.ar alter column ra set default 0;"+
	//"alter table public.ar alter column balance set default 0;"+

	if err != nil {
		fmt.Println("CreateTable:" + err.Error())
		return err
	}
	fmt.Println("CreatehousegoTable Done")
	return nil
}

func (sdb *SqlDB) CreateEventLogTable() error {

	_, err := sdb.SQLCommand(fmt.Sprintf(
		"CREATE TABLE public.eventlog "+
			"( "+
			"id character varying(50) not NULL,"+
			"account character varying(50) not NULL,"+
			"name character varying(50) not NULL,"+
			"auth character varying(50) not NULL, "+
			"msg character varying(50) not NULL, "+
			"type character varying(50) not NULL, "+
			"date timestamp(0) without time zone not NULL, "+
			"PRIMARY KEY (id) "+
			") "+
			"WITH ( OIDS = FALSE);"+ //))
			" ALTER TABLE public.eventlog "+
			"OWNER to %s; ", sdb.User))
	//"alter table public.ar alter column ra set default 0;"+
	//"alter table public.ar alter column balance set default 0;"+

	if err != nil {
		fmt.Println("CreateEventLogTable:" + err.Error())
		return err
	}
	fmt.Println("CreateEventLogTable Done")
	return nil
}

func (sdb *SqlDB) CreateAccountTable() error {

	_, err := sdb.SQLCommand(fmt.Sprintf(
		"CREATE TABLE public.account "+
			"( "+
			"account character varying(50) not NULL,"+
			//"passoword character varying(50) not NULL,"+
			"name character varying(50) not NULL,"+
			"permission character varying(50) not NULL, "+
			//"email character varying(50) not NULL, "+ // 信箱
			//"phone character varying(50) DEFAULT '', "+ //
			"branch character varying(50) DEFAULT '', "+ // 店家
			"createdate timestamp(0) DEFAULT now(), "+
			"lasttime timestamp(0) DEFAULT NULL, "+
			"state character varying(50) not NULL, "+ // 狀態
			"disable integer DEFAULT 0, "+ // 啟用
			"PRIMARY KEY (account) "+
			") "+
			"WITH ( OIDS = FALSE);"+ //))
			" ALTER TABLE public.account "+
			"OWNER to %s; ", sdb.User))
	//"alter table public.ar alter column ra set default 0;"+
	//"alter table public.ar alter column balance set default 0;"+

	if err != nil {
		fmt.Println("CreateTable:" + err.Error())
		return err
	}
	fmt.Println("CreateAccountTable Done")
	return nil
}

func (sdb *SqlDB) CreateARTable() error {

	// CREATE SEQUENCE public."generateID"
	// INCREMENT 1
	// START 1
	// MINVALUE 1
	// MAXVALUE 99999999
	// CACHE 1;

	// ALTER SEQUENCE public."generateID"
	// 	OWNER TO postgres;

	_, err := sdb.SQLCommand(fmt.Sprintf(
		"CREATE TABLE public.AR "+
			"( "+
			"ARid character varying(50) ,"+
			"date timestamp(0) without time zone not NULL, "+
			"cNo character varying(50) not NULL, "+
			"caseName character varying(50) not NULL, "+
			"type character varying(50) not NULL, "+
			"name character varying(50) not NULL, "+
			"amount integer not NULL, "+
			//"fee integer DEFAULT 0, "+ //應扣費用
			//"RA integer DEFAULT 0, "+ //已收金額
			//"balance integer DEFAULT 0, "+
			//"sales character varying(600), "+ //json[]
			"PRIMARY KEY (ARid) "+
			") "+
			"WITH ( OIDS = FALSE);"+ //))
			" ALTER TABLE public.AR "+
			"OWNER to %s; ", sdb.User))
	//"alter table public.ar alter column ra set default 0;"+
	//"alter table public.ar alter column balance set default 0;"+

	if err != nil {
		fmt.Println("CreateTable:" + err.Error())
		return err
	}
	fmt.Println("CreateARTable Done")
	return nil
}

func (sdb *SqlDB) CreateNHISalaryTable() error {
	_, err := sdb.SQLCommand(fmt.Sprintf(
		"CREATE TABLE public.NHISalary "+
			"( "+
			"Sid character varying(50) ,"+
			"BSid character varying(50) ,"+
			"SName character varying(50) ,"+
			"PayrollBracket integer DEFAULT 0, "+
			"Salary integer DEFAULT 0, "+
			"Pbonus integer DEFAULT 0, "+
			"Bonus integer DEFAULT 0, "+
			"Total integer DEFAULT 0, "+
			"PD integer DEFAULT 0, "+ // 累進差額 (Progressive Difference)
			"SalaryBalance integer DEFAULT 0, "+
			"FourBouns integer DEFAULT 0, "+
			"SP integer DEFAULT 0, "+
			"FourSP integer DEFAULT 0, "+
			"PTSP integer DEFAULT 0, "+
			"PRIMARY KEY (BSid,Sid) "+
			") "+
			"WITH ( OIDS = FALSE);"+ //))
			"ALTER TABLE public.NHISalary "+
			"OWNER to %s; ", sdb.User))

	if err != nil {
		fmt.Println("CreateNHISalaryTable:" + err.Error())
		return err
	}
	fmt.Println("CreateNHISalaryTable Done")
	return nil
}

func (sdb *SqlDB) CreateSalerSalaryTable() error {
	_, err := sdb.SQLCommand(fmt.Sprintf(
		"CREATE TABLE public.SalerSalary "+
			"( "+
			"Sid character varying(50) ,"+
			"BSid character varying(50) ,"+
			"SName character varying(50) ,"+
			//"Date timestamp(0) without time zone not NULL, "+
			"Date character varying(10) not NULL, "+
			"Branch character varying(50),"+
			"Salary integer DEFAULT 0, "+
			"Pbonus integer DEFAULT 0, "+
			"Lbonus integer DEFAULT 0, "+
			"Abonus integer DEFAULT 0, "+
			"Total integer DEFAULT 0, "+
			"SP integer DEFAULT 0, "+
			"Tax integer DEFAULT 0, "+
			"LaborFee  integer DEFAULT 0, "+
			"HealthFee  integer DEFAULT 0, "+
			"Welfare  integer DEFAULT 0, "+
			"CommercialFee integer DEFAULT 0, "+ //商耕費(原本組織費)
			"Other  integer DEFAULT 0, "+
			"TAmount integer DEFAULT 0, "+
			"Description character varying(50) ,"+
			"WorkDay   integer DEFAULT 30, "+
			"Year character varying(4),"+
			"PRIMARY KEY (BSid,Sid,Date,Branch) "+
			") "+
			"WITH ( OIDS = FALSE);"+ //))
			"ALTER TABLE public.SalerSalary "+
			"OWNER to %s; ", sdb.User))

	if err != nil {
		fmt.Println("CreateSalerSalaryTable:" + err.Error())
		return err
	}
	fmt.Println("CreateSalerSalaryTable Done")
	return nil
}

func (sdb *SqlDB) CreateBranchSalaryTable() error {
	_, err := sdb.SQLCommand(fmt.Sprintf(
		"CREATE TABLE public.BranchSalary "+
			"( "+
			"BSid character varying(50), "+
			// "Date timestamp(0) without time zone not NULL, "+
			"Date character varying(50) not NULL, "+
			"Branch character varying(50),"+
			"Name character varying(50),"+
			"Total integer DEFAULT 0, "+
			"Lock character varying(50) DEFAULT '未完成', "+
			"PRIMARY KEY (BSid,Branch) "+
			") "+
			"WITH ( OIDS = FALSE);"+ //))
			"ALTER TABLE public.BranchSalary "+
			"OWNER to %s; ", sdb.User))

	if err != nil {
		fmt.Println("CreateBranchSalaryTable:" + err.Error())
		return err
	}
	fmt.Println("CreateBranchSalaryTable Done")
	return nil
}

func (sdb *SqlDB) CreatePrePayTable() error {

	_, err := sdb.SQLCommand(fmt.Sprintf(
		"CREATE TABLE public.PrePay "+
			"( "+
			"PPid character varying(50) ,"+
			"date timestamp(0) without time zone not NULL, "+
			"itemname character varying(50) ,"+
			"description character varying(50) ,"+
			"Fee  integer DEFAULT 0, "+
			"PRIMARY KEY (PPid) "+
			") "+
			"WITH ( OIDS = FALSE);"+ //))
			"ALTER TABLE public.PrePay "+
			"OWNER to %s; ", sdb.User))

	if err != nil {
		fmt.Println("CreatePrePayTable:" + err.Error())
		return err
	}
	fmt.Println("CreatePrePayTable Done")
	return nil
}

func (sdb *SqlDB) CreateBranchPrePayTable() error {

	_, err := sdb.SQLCommand(fmt.Sprintf(
		"CREATE TABLE public.BranchPrePay "+
			"( "+
			"PPid character varying(50) ,"+
			"Branch character varying(50) ,"+
			"Cost  integer DEFAULT 0, "+
			"PRIMARY KEY (PPid,Branch) "+
			") "+
			"WITH ( OIDS = FALSE);"+ //))
			"ALTER TABLE public.BranchPrePay "+
			"OWNER to %s; ", sdb.User))

	if err != nil {
		fmt.Println("CreateBranchPrePayTable:" + err.Error())
		return err
	}
	fmt.Println("CreateBranchPrePayTable Done")
	return nil
}

func (sdb *SqlDB) CreatePocketTable() error {

	_, err := sdb.SQLCommand(fmt.Sprintf(
		"CREATE TABLE public.Pocket "+
			"( "+
			"Pid character varying(50) ,"+
			"date timestamp(0) without time zone not NULL, "+
			"branch character varying(50) ,"+
			"itemname character varying(50) ,"+
			"description character varying(50) ,"+
			"CircleID character varying(50) ,"+
			"Income  integer DEFAULT 0, "+
			"Fee  integer DEFAULT 0, "+
			"Balance  integer DEFAULT NULL, "+
			"PRIMARY KEY (Pid,branch) "+
			") "+
			"WITH ( OIDS = FALSE);"+ //))
			"ALTER TABLE public.Pocket "+
			"OWNER to %s; ", sdb.User))

	if err != nil {
		fmt.Println("CreatePocketTable:" + err.Error())
		return err
	}
	fmt.Println("CreatePocketTable Done")
	return nil
}

func (sdb *SqlDB) CreateAccountItemTable() error {

	_, err := sdb.SQLCommand(fmt.Sprintf(
		"CREATE TABLE public.AccountItem "+
			"( "+
			"AccountItemName character varying(50) ,"+
			"Valid integer DEFAULT 1, "+
			"PRIMARY KEY (AccountItemName) "+
			") "+
			"WITH ( OIDS = FALSE);"+ //))
			"ALTER TABLE public.AccountItem "+
			"OWNER to %s; ", sdb.User))

	if err != nil {
		fmt.Println("CreateAccountItemTable:" + err.Error())
		return err
	}
	fmt.Println("CreateAccountItemTable Done")
	return nil
}

func (sdb *SqlDB) CreateAmortizationTable() error {

	_, err := sdb.SQLCommand(fmt.Sprintf(
		"CREATE TABLE public.amortization "+
			"( "+
			"Amorid character varying(50) ,"+
			"branch character varying(50) ,"+
			"date timestamp(0) without time zone not NULL, "+
			"itemName character varying(50) not NULL, "+
			"GainCost integer DEFAULT 0, "+
			"AmortizationYearLimit integer DEFAULT 0,  "+
			"MonthlyAmortizationAmount integer DEFAULT 0,  "+
			"FirstAmortizationAmount integer DEFAULT 0,  "+
			"HasAmortizationAmount integer DEFAULT 0,  "+
			"NotAmortizationAmount integer DEFAULT 0,  "+
			"isOver integer DEFAULT 0,  "+
			"PRIMARY KEY (Amorid) "+
			") "+
			"WITH ( OIDS = FALSE);"+ //))
			"ALTER TABLE public.amortization "+
			"OWNER to %s; ", sdb.User))

	if err != nil {
		fmt.Println("CreateAmortizationTable:" + err.Error())
		return err
	}
	fmt.Println("CreateAmortizationTable Done")
	return nil
}

func (sdb *SqlDB) CreateAmorMapTable() error {

	_, err := sdb.SQLCommand(fmt.Sprintf(
		"CREATE TABLE public.AmorMap "+
			"( "+
			"Amorid character varying(50) ,"+
			"Date character varying(10) ,"+
			//"Bsid character varying(50) ,"+
			"Cost integer DEFAULT 0, "+
			"PRIMARY KEY (Amorid,Date) "+
			") "+
			"WITH ( OIDS = FALSE);"+ //))
			"ALTER TABLE public.AmorMap "+
			"OWNER to %s; ", sdb.User))

	if err != nil {
		fmt.Println("CreateAmorMapTable:" + err.Error())
		return err
	}
	fmt.Println("CreateAmorMapTable Done")
	return nil
}

func (sdb *SqlDB) CreateInvoiceTable() error {

	_, err := sdb.SQLCommand(fmt.Sprintf(
		"CREATE TABLE public.Invoice "+
			"( "+
			"Rid character varying(50) ,"+
			"Sid character varying(50) ,"+
			"InvoiceNo character varying(20) ,"+
			"BuyerID character varying(20) ,"+
			"SellerID  character varying(20) ,"+
			"RandomNum character varying(10) ,"+
			"Title character varying(50) DEFAULT NULL,"+
			"Date  character varying(50) DEFAULT NULL, "+
			"Amount  integer not NULL, "+
			"left_qrcode  character varying(200) DEFAULT NULL, "+
			"right_qrcode character varying(200) DEFAULT NULL, "+
			"PRIMARY KEY (Rid,Sid) "+
			") "+
			"WITH ( OIDS = FALSE);"+ //))
			"ALTER TABLE public.Invoice "+
			"OWNER to %s; ", sdb.User))

	if err != nil {
		fmt.Println("CreateInvoiceTable:" + err.Error())
		return err
	}
	fmt.Println("CreateInvoiceTable Done")
	return nil
}

func (sdb *SqlDB) CreateInvoiceConfigTable() error {

	_, err := sdb.SQLCommand(fmt.Sprintf(
		"CREATE TABLE public.InvoiceConfig "+
			"( "+
			"Branch character varying(50) ,"+
			"SellerID character varying(50) ,"+
			"Auth character varying(100) ,"+
			"PRIMARY KEY (Branch) "+
			") "+
			"WITH ( OIDS = FALSE);"+ //))
			"ALTER TABLE public.InvoiceConfig "+
			"OWNER to %s; ", sdb.User))

	if err != nil {
		fmt.Println("CreateInvoiceConfigTable:" + err.Error())
		return err
	}
	fmt.Println("CreateInvoiceConfigTable Done")
	return nil
}

func (sdb *SqlDB) CreateARMAPTable() error {

	_, err := sdb.SQLCommand(fmt.Sprintf(
		"CREATE TABLE public.ARMAP "+
			"( "+
			"ARid character varying(50) not NULL,"+
			"Sid character varying(50) not NULL,"+
			"Proportion double precision DEFAULT 0,"+
			"SName  character varying(50) not NULL,"+
			"PRIMARY KEY (ARid,Sid) "+
			") "+
			"WITH ( OIDS = FALSE);"+ //))
			"ALTER TABLE public.ARMAP "+
			"OWNER to %s; ", sdb.User))

	if err != nil {
		fmt.Println("CreateARMAPTable:" + err.Error())
		return err
	}
	fmt.Println("CreateARMapTable Done")
	return nil
}

func (sdb *SqlDB) CreateDeductMAPTable() error {

	_, err := sdb.SQLCommand(fmt.Sprintf(
		"CREATE TABLE public.DEDUCTMAP "+
			"( "+
			"Did character varying(50) not NULL,"+
			"Sid character varying(50) not NULL,"+
			"Proportion double precision DEFAULT 0,"+
			"SName  character varying(50) not NULL,"+
			"PRIMARY KEY (Did,Sid) "+
			") "+
			"WITH ( OIDS = FALSE);"+ //))
			"ALTER TABLE public.DEDUCTMAP "+
			"OWNER to %s; ", sdb.User))

	if err != nil {
		fmt.Println("CreateDEDUCTMAPTable:" + err.Error())
		return err
	}
	fmt.Println("CreateDEDUCTMAPTable Done")
	return nil
}

func (sdb *SqlDB) CreateReceiptTable() error {

	_, err := sdb.SQLCommand(fmt.Sprintf(
		"CREATE TABLE public.receipt "+
			"( "+
			"Rid character varying(50) ,"+
			"date timestamp(0) without time zone not NULL, "+
			//"cNo character varying(50) not NULL, "+
			//"caseName character varying(50) not NULL, "+
			//"type character varying(50) not NULL, "+
			//"name character varying(50) not NULL, "+
			"amount integer not NULL, "+
			//"invoiceNo character varying(50) , "+
			"ARid character varying(50) not NULL, "+
			"PRIMARY KEY (Rid) "+
			") "+
			"WITH ( OIDS = FALSE);"+ //))
			"ALTER TABLE public.receipt "+
			"OWNER to %s; ", sdb.User))

	if err != nil {
		fmt.Println("CreateReceiptTable:" + err.Error())
		return err
	}
	fmt.Println("CreateReceiptTable Done")
	return nil
}

func (sdb *SqlDB) CreateConfigBranchTable() error {

	_, err := sdb.SQLCommand(fmt.Sprintf(
		"CREATE TABLE public.ConfigBranch "+
			"( "+
			"Branch character varying(50) ,"+
			"Rent integer DEFAULT 0, "+
			"AgentSign integer DEFAULT 0, "+
			"CommercialFee double precision DEFAULT 0,"+
			"AnnualRatio double precision DEFAULT 0,"+
			"Manager character varying(50) ,"+
			"Sid character varying(50) ,"+
			"PRIMARY KEY (Branch) "+
			") "+
			"WITH ( OIDS = FALSE);"+ //))
			"ALTER TABLE public.ConfigBranch "+
			"OWNER to %s; ", sdb.User))

	if err != nil {
		fmt.Println("CreateConfigBranchTable:" + err.Error())
		return err
	}
	fmt.Println("CreateConfigBranchTable Done")
	return nil
}

func (sdb *SqlDB) CreateConfigParameterTable() error {

	_, err := sdb.SQLCommand(fmt.Sprintf(
		"CREATE TABLE public.ConfigParameter "+
			"( "+
			"id character varying(50) ,"+
			"date timestamp(0) without time zone not NULL,"+
			"NHI double precision DEFAULT 0, "+
			"LI double precision DEFAULT 0, "+
			"NHI2nd double precision DEFAULT 0, "+
			"MMW  integer  DEFAULT 0, "+
			//"AnnualRatio double precision DEFAULT 0,"+
			//"IT double precision DEFAULT 0, "+
			"PRIMARY KEY (id) "+
			") "+
			"WITH ( OIDS = FALSE);"+ //))
			"ALTER TABLE public.ConfigParameter "+
			"OWNER to %s; ", sdb.User))

	if err != nil {
		fmt.Println("CreateConfigParameterTable:" + err.Error())
		return err
	}
	fmt.Println("CreateConfigParameterTable Done")
	return nil
}

func (sdb *SqlDB) CreateConfigSalerTable() error {

	_, err := sdb.SQLCommand(fmt.Sprintf(
		"CREATE TABLE public.ConfigSaler "+
			"( "+
			//"csid character varying(50) ,"+
			"Sid character varying(50) ,"+
			"Sname character varying(50) ,"+
			"Code integer DEFAULT 200,"+ //員工代號
			"Branch character varying(50) ,"+
			"ZeroDate timestamp(0) without time zone not NULL, "+
			//"ValidDate  timestamp(0) without time zone not NULL, "+
			"Title character varying(50) ,"+
			"Percent double precision DEFAULT 0, "+
			//"FPercent double precision DEFAULT 0, "+
			"Salary integer DEFAULT 0, "+
			//"Pay integer DEFAULT 0, "+ //未來薪資
			"PayrollBracket integer DEFAULT 0, "+ //健保 投保金額
			"InsuredAmount integer DEFAULT 0, "+ //勞保 投保金額
			"Enrollment integer DEFAULT 0, "+ //加保(眷屬人數)
			"Association integer DEFAULT 0, "+ // 公會
			"Address character varying(50) DEFAULT '', "+ // 地址
			"Birth character varying(50) DEFAULT '', "+ // 出生年月日
			"IdentityNum character varying(50) DEFAULT '', "+ // 身份證字號
			"Bankaccount character varying(50) DEFAULT '', "+ // 銀行帳戶
			"Email character varying(50) DEFAULT '', "+ // 信箱
			"Phone character varying(50) DEFAULT '', "+ // 電話
			"Remark character varying(50) DEFAULT '', "+ // 備註
			"PRIMARY KEY (sid) "+
			") "+
			"WITH ( OIDS = FALSE);"+ //))
			"ALTER TABLE public.ConfigSaler "+
			"OWNER to %s; ", sdb.User))

	if err != nil {
		fmt.Println("CreateConfigSalerTable:" + err.Error())
		return err
	}
	fmt.Println("CreateConfigSalerTable Done")
	return nil
}

func (sdb *SqlDB) CreateConfigSalaryTable() error {

	_, err := sdb.SQLCommand(fmt.Sprintf(
		"CREATE TABLE public.ConfigSalary "+
			"( "+
			"Sid character varying(50) ,"+
			"ZeroDate character varying(50), "+ // ☆這裡用string★
			"Sname character varying(50) ,"+
			"Branch character varying(50) ,"+
			"Title character varying(50) ,"+
			"Percent double precision DEFAULT 0, "+
			"Salary integer DEFAULT 0, "+
			"PayrollBracket integer DEFAULT 0, "+ //健保 投保金額
			"InsuredAmount integer DEFAULT 0, "+ //勞保 投保金額
			"Enrollment integer DEFAULT 0, "+ //加保(眷屬人數)
			"Association integer DEFAULT 0, "+ // 公會
			"Remark character varying(50) DEFAULT '', "+ // 備註
			"PRIMARY KEY (sid, ZeroDate) "+
			") "+
			"WITH ( OIDS = FALSE);"+ //))
			"ALTER TABLE public.ConfigSalary "+
			"OWNER to %s; ", sdb.User))

	if err != nil {
		fmt.Println("CreateConfigSalerTable:" + err.Error())
		return err
	}
	fmt.Println("CreateConfigSalerTable Done")
	return nil
}

func (sdb *SqlDB) CreateCommissionTable() error {

	_, err := sdb.SQLCommand(fmt.Sprintf(
		"CREATE TABLE public.commission "+
			"( "+
			"Sid character varying(50) ,"+
			"Rid character varying(50) ,"+
			"ARid character varying(50) ,"+
			"BSid character varying(50) ,"+
			//"date timestamp(0) without time zone not NULL, "+
			"item character varying(50) not NULL, "+
			"SName character varying(50) , "+
			//"amount integer not NULL, "+
			"fee integer DEFAULT 0, "+
			//"percent double precision DEFAULT 0 , "+
			"CPercent double precision DEFAULT 0 , "+
			"SR integer DEFAULT 0 ,"+
			"bonus integer DEFAULT 0 ,"+
			"status character varying(50) DEFAULT 'normal',"+
			//"bankaccount character varying(50) DEFAULT 'normal',"+
			"PRIMARY KEY (Sid,Rid)"+
			") "+
			"WITH ( OIDS = FALSE);"+
			"ALTER TABLE public.commission "+
			"OWNER to %s; ", sdb.User))

	if err != nil {
		fmt.Println("CreateCommissionTable:" + err.Error())
		return err
	}
	fmt.Println("CreateCommissionTable Done")
	return nil
}

func (sdb *SqlDB) CreateAccountSettlementTable() error {

	_, err := sdb.SQLCommand(fmt.Sprintf(
		"CREATE TABLE public.accountsettlement "+
			"( "+
			"id character varying(50) ,"+
			"uid character varying(50) ,"+
			"closedate timestamp(0) without time zone not NULL, "+ //關帳日
			"date timestamp(0) without time zone not NULL default now(), "+ //操作人員動作時間
			"status character varying(50) default '1',"+
			"PRIMARY KEY (id)"+
			") "+
			"WITH ( OIDS = FALSE);"+
			"ALTER TABLE public.commission "+
			"OWNER to %s; ", sdb.User))

	if err != nil {
		fmt.Println("CreateCommissionTable:" + err.Error())
		return err
	}
	fmt.Println("CreateCommissionTable Done")
	return nil
}

//收入支出 (紅利店長表)
func (sdb *SqlDB) CreateIncomeExpenseTable() error {

	_, err := sdb.SQLCommand(fmt.Sprintf(
		"CREATE TABLE public.IncomeExpense "+
			"( "+
			"BSid character varying(50) ,"+
			"SR integer DEFAULT 0,"+
			"BusinessTax integer DEFAULT 0,"+
			"SalesAmounts integer  DEFAULT 0,"+
			"PBonus integer  DEFAULT 0,"+ //獎金(績效))
			"LBonus integer  DEFAULT 0,"+ //組長(領導)
			"AmorCost integer  DEFAULT 0,"+
			"AgentSign integer  DEFAULT 0,"+
			"Rent integer  DEFAULT 0,"+
			"Commercialfee integer  DEFAULT 0,"+
			"Salary integer  DEFAULT 0,"+
			"Prepay integer  DEFAULT 0,"+
			"Pocket integer  DEFAULT 0,"+
			"AnnualBonus integer  DEFAULT 0,"+
			"SalerFee integer  DEFAULT 0,"+
			"PreTax integer  DEFAULT 0,"+
			"AfterTax integer  DEFAULT 0,"+
			"EarnAdjust  integer  DEFAULT 0,"+
			"LastLoss  integer  DEFAULT 0,"+
			"BusinessIncomeTax  integer  DEFAULT 0,"+
			"ManagerBonus  integer  DEFAULT 0,"+
			"AnnualRatio double precision DEFAULT 0,"+
			"PRIMARY KEY (BSid)"+
			") "+
			"WITH ( OIDS = FALSE);"+
			"ALTER TABLE public.IncomeExpense "+
			"OWNER to %s; ", sdb.User))

	if err != nil {
		fmt.Println("CreateIncomeExpenseTable:" + err.Error())
		return err
	}
	fmt.Println("CreateIncomeExpenseTable Done")
	return nil
}

func (sdb *SqlDB) InitTable() error {

	rows, err := sdb.SQLCommand(`SELECT tablename FROM	pg_catalog.pg_tables 
								WHERE 
								schemaname != 'pg_catalog' AND schemaname != 'information_schema';`)

	if err != nil {
		return err
	}
	//mapTable
	var mT = map[string]bool{
		"ar":                false,
		"receipt":           false,
		"deduct":            false,
		"commission":        false,
		"amortization":      false,
		"amormap":           false,
		"pocket":            false,
		"branchprepay":      false,
		"prepay":            false,
		"accountitem":       false,
		"configbranch":      false,
		"configsaler":       false,
		"configsalary":      false,
		"configparameter":   false,
		"invoice":           false,
		"invoiceconfig":     false,
		"armap":             false,
		"deductmap":         false,
		"salersalary":       false,
		"branchsalary":      false,
		"nhisalary":         false,
		"incomeexpense":     false,
		"housego":           false,
		"account":           false,
		"eventlog":          false,
		"accountsettlement": false, //關帳日
	}

	for rows.Next() {
		var tName string
		if err := rows.Scan(&tName); err != nil {
			fmt.Println("err Scan %s\n", err)
		}

		switch tName {
		case "ar":
			mT["ar"] = true
			break
		case "receipt":
			mT["receipt"] = true
			break
		case "deduct":
			mT["deduct"] = true
			break
		case "commission":
			mT["commission"] = true
			break
		case "amortization":
			mT["amortization"] = true
			break
		case "amormap":
			mT["amormap"] = true
			break
		case "pocket":
			mT["pocket"] = true
			break
		case "branchprepay":
			mT["branchprepay"] = true
			break
		case "prepay":
			mT["prepay"] = true
			break
		case "accountitem":
			mT["accountitem"] = true
			break
		case "configbranch":
			mT["configbranch"] = true
			break
		case "configsaler":
			mT["configsaler"] = true
			break
		case "configsalary":
			mT["configsalary"] = true
			break
		case "configparameter":
			mT["configparameter"] = true
			break
		case "invoice":
			mT["invoice"] = true
			break
		case "invoiceconfig":
			mT["invoiceconfig"] = true
			break
		case "armap":
			mT["armap"] = true
			break
		case "deductmap":
			mT["deductmap"] = true
			break
		case "salersalary":
			mT["salersalary"] = true
			break
		case "branchsalary":
			mT["branchsalary"] = true
			break
		case "nhisalary":
			mT["nhisalary"] = true
			break
		case "incomeexpense":
			mT["incomeexpense"] = true
			break
		case "housego":
			mT["housego"] = true
			break
		case "account":
			mT["account"] = true
			break
		case "eventlog":
			mT["eventlog"] = true
			break
		case "accountsettlement":
			mT["accountsettlement"] = true
			break
		default:
			fmt.Printf("unknown table %s.\n", tName)
		}
	}

	for tableName, status := range mT {
		//fmt.Println("i=", t, " s=", s)
		if !status {
			err = nil
			switch tableName {
			case "ar":
				err = sdb.CreateARTable()
				break
			case "receipt":
				err = sdb.CreateReceiptTable()
				break
			case "deduct":
				err = sdb.CreateDeductTable()
				break
			case "commission":
				err = sdb.CreateCommissionTable()
				break
			case "amortization":
				err = sdb.CreateAmortizationTable()
				break
			case "amormap":
				err = sdb.CreateAmorMapTable()
				break
			case "accountitem":
				err = sdb.CreateAccountItemTable()
				break
			case "pocket":
				err = sdb.CreatePocketTable()
				break
			case "prepay":
				err = sdb.CreatePrePayTable()
				break
			case "branchprepay":
				err = sdb.CreateBranchPrePayTable()
				break
			case "configbranch":
				err = sdb.CreateConfigBranchTable()
				break
			case "configsaler":
				err = sdb.CreateConfigSalerTable()
				break
			case "configsalary":
				err = sdb.CreateConfigSalaryTable()
				break

			case "configparameter":
				err = sdb.CreateConfigParameterTable()
				break
			case "invoice":
				err = sdb.CreateInvoiceTable()
				break
			case "invoiceconfig":
				err = sdb.CreateInvoiceConfigTable()
				break
			case "armap":
				err = sdb.CreateARMAPTable()
				break
			case "salersalary":
				err = sdb.CreateSalerSalaryTable()
				break
			case "branchsalary":
				err = sdb.CreateBranchSalaryTable()
				break
			case "nhisalary":
				err = sdb.CreateNHISalaryTable()
				break
			case "incomeexpense":
				err = sdb.CreateIncomeExpenseTable()
				break
			case "housego":
				err = sdb.CreateHouseGoTable()
				break
			case "account":
				err = sdb.CreateAccountTable()
				break
			case "eventlog":
				err = sdb.CreateEventLogTable()
				break
			case "deductmap":
				err = sdb.CreateDeductMAPTable()
				break
			case "accountsettlement":
				err = sdb.CreateAccountSettlementTable()
				break
			default:
				fmt.Printf("unknown table %s.\n", tableName)
				break
			}
			if err != nil {
				fmt.Println(err)
			}
		}
	}

	return err
}

func (sdb *SqlDB) InitDB() bool {

	rows, err := sdb.SQLCommand(fmt.Sprintf("SELECT datname FROM pg_database WHERE datname = '%s';", sdb.Db))

	if err != nil {
		return false
	}

	for rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			fmt.Println("err Scan %s\n", err)
		}
		fmt.Println("Data Base " + value + " exist")
		//return false
	}

	if err := rows.Err(); err != nil {
		fmt.Println("err rows.Err() %s\n", err)
		return true
	}

	if err := sdb.CreateDB(); err != nil {
		if strings.Index(err.Error(), "already exists") > -1 {
			fmt.Println("toad already exists, init table")
			if err := sdb.InitTable(); err != nil {
				fmt.Println("InitTable Err() %s\n", err)
				return true
			}
		} else {
			fmt.Println("CreateDB Err() %s\n", err)
			return true
		}
	}

	if err := sdb.InitTable(); err != nil {
		fmt.Println("InitTable Err() %s\n", err)
		return true
	}

	return false
}
