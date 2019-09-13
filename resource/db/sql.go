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

type sqlDB struct {
	ctx context.Context

	c        string
	clinet   *sql.DB
	port     int
	dburl    string
	user     string
	password string
	db       string
}

func (sdb *sqlDB) ConnectSQLDB() (*sql.DB, error) {

	//完整的資料格式連線如下
	var connectionString string = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", sdb.dburl, sdb.port, sdb.user, sdb.password, sdb.db)
	//var connectionString string = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=ak47 sslmode=disable", sdb.dburl, sdb.port, sdb.user, sdb.password)
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
			fmt.Println("database2 " + sdb.db + " does not exist")
			var connectionString string = fmt.Sprintf("host=%s port=%d user=%s password=%s sslmode=disable", sdb.dburl, sdb.port, sdb.user, sdb.password)
			db, err := sql.Open("postgres", connectionString)
			err = db.Ping()
			if err != nil {
				fmt.Println("[Failed] ping2:" + err.Error())
			} else {
				fmt.Println("無指定DB 連線成功")
				sdb.clinet = db
				return sdb.clinet, err
			}
		}
		return nil, err
	}
	fmt.Println("toad 連線成功")
	sdb.clinet = db
	return sdb.clinet, err
}

func (sdb *sqlDB) C(c string) InterSQLDB {
	sdb.c = c
	return sdb
}

func (sdb *sqlDB) Close() error {

	if sdb.clinet == nil {
		return nil
	}
	fmt.Println("close db")
	return sdb.Close()
}

func (sdb *sqlDB) SQLCommand(cmd string) (*sql.Rows, error) {

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
		return nil, rows.Err()
	}

	return rows, err
}

func (sdb *sqlDB) CreateDB() error {

	_, err := sdb.SQLCommand(fmt.Sprintf(
		"CREATE DATABASE %s "+
			"WITH "+
			"OWNER = %s "+
			"ENCODING = 'UTF8' "+
			"CONNECTION LIMIT = -1;", sdb.db, sdb.user))

	if err != nil {
		fmt.Println("CreateDB:" + err.Error())
		return err
	}
	return nil
}

func (sdb *sqlDB) CreateDeductTable() error {

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
			"Rid  character varying(50) ,"+
			"PRIMARY KEY (Did) "+
			") "+
			"WITH ( OIDS = FALSE);"+ //))
			"ALTER TABLE public.Deduct "+
			"OWNER to %s; ", sdb.user))

	if err != nil {
		fmt.Println("CreateDeductTable:" + err.Error())
		return err
	}
	fmt.Println("CreateDeductTable Done")
	return nil
}

func (sdb *sqlDB) CreateARTable() error {

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
			"OWNER to %s; ", sdb.user))
	//"alter table public.ar alter column ra set default 0;"+
	//"alter table public.ar alter column balance set default 0;"+

	if err != nil {
		fmt.Println("CreateTable:" + err.Error())
		return err
	}
	fmt.Println("CreateARTable Done")
	return nil
}
func (sdb *sqlDB) CreateSalerSalaryTable() error {
	_, err := sdb.SQLCommand(fmt.Sprintf(
		"CREATE TABLE public.SalerSalary "+
			"( "+
			"Sid character varying(50) ,"+
			"Date timestamp(0) without time zone not NULL, "+
			"Branch character varying(50),"+
			"Pbonus integer DEFAULT 0, "+
			"Lbonus integer DEFAULT 0, "+
			"Abonus integer DEFAULT 0, "+
			"SP integer DEFAULT 0, "+
			"Tax integer DEFAULT 0, "+
			"LaborFee  integer DEFAULT 0, "+
			"HealthFee  integer DEFAULT 0, "+
			"Welfare  integer DEFAULT 0, "+
			"Org integer DEFAULT 0, "+
			"Other  integer DEFAULT 0, "+
			"TAmount integer DEFAULT 0, "+
			//"Lock integer DEFAULT 0, "+
			"PRIMARY KEY (Sid,Date,Branch) "+
			") "+
			"WITH ( OIDS = FALSE);"+ //))
			"ALTER TABLE public.SalerSalary "+
			"OWNER to %s; ", sdb.user))

	if err != nil {
		fmt.Println("CreateSalerSalaryTable:" + err.Error())
		return err
	}
	fmt.Println("CreateSalerSalaryTable Done")
	return nil
}

func (sdb *sqlDB) CreateBranchSalaryTable() error {
	_, err := sdb.SQLCommand(fmt.Sprintf(
		"CREATE TABLE public.BranchSalary "+
			"( "+
			"Date timestamp(0) without time zone not NULL, "+
			"Branch character varying(50),"+
			"Name character varying(50),"+
			"Total integer DEFAULT 0, "+
			"Lock integer DEFAULT 0, "+
			"PRIMARY KEY (Date,Branch) "+
			") "+
			"WITH ( OIDS = FALSE);"+ //))
			"ALTER TABLE public.BranchSalary "+
			"OWNER to %s; ", sdb.user))

	if err != nil {
		fmt.Println("CreateBranchSalaryTable:" + err.Error())
		return err
	}
	fmt.Println("CreateBranchSalaryTable Done")
	return nil
}

func (sdb *sqlDB) CreatePrePayTable() error {

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
			"OWNER to %s; ", sdb.user))

	if err != nil {
		fmt.Println("CreatePrePayTable:" + err.Error())
		return err
	}
	fmt.Println("CreatePrePayTable Done")
	return nil
}

func (sdb *sqlDB) CreateBranchPrePayTable() error {

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
			"OWNER to %s; ", sdb.user))

	if err != nil {
		fmt.Println("CreateBranchPrePayTable:" + err.Error())
		return err
	}
	fmt.Println("CreateBranchPrePayTable Done")
	return nil
}

func (sdb *sqlDB) CreatePocketTable() error {

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
			"OWNER to %s; ", sdb.user))

	if err != nil {
		fmt.Println("CreatePocketTable:" + err.Error())
		return err
	}
	fmt.Println("CreatePocketTable Done")
	return nil
}

func (sdb *sqlDB) CreateAccountItemTable() error {

	_, err := sdb.SQLCommand(fmt.Sprintf(
		"CREATE TABLE public.AccountItem "+
			"( "+
			"AccountItemName character varying(50) ,"+
			"Valid integer DEFAULT 1, "+
			"PRIMARY KEY (AccountItemName) "+
			") "+
			"WITH ( OIDS = FALSE);"+ //))
			"ALTER TABLE public.AccountItem "+
			"OWNER to %s; ", sdb.user))

	if err != nil {
		fmt.Println("CreateAccountItemTable:" + err.Error())
		return err
	}
	fmt.Println("CreateAccountItemTable Done")
	return nil
}

func (sdb *sqlDB) CreateAmortizationTable() error {

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
			"OWNER to %s; ", sdb.user))

	if err != nil {
		fmt.Println("CreateAmortizationTable:" + err.Error())
		return err
	}
	fmt.Println("CreateAmortizationTable Done")
	return nil
}

func (sdb *sqlDB) CreateInvoiceTable() error {

	_, err := sdb.SQLCommand(fmt.Sprintf(
		"CREATE TABLE public.Invoice "+
			"( "+
			"Rid character varying(50) ,"+
			"invoice character varying(50) ,"+
			"Title character varying(50) DEFAULT NULL,"+
			"Date timestamp(0) without time zone not NULL, "+
			"GUI  character varying(50) DEFAULT NULL, "+
			"Amount  integer not NULL, "+
			"PRIMARY KEY (Rid) "+
			") "+
			"WITH ( OIDS = FALSE);"+ //))
			"ALTER TABLE public.Invoice "+
			"OWNER to %s; ", sdb.user))

	if err != nil {
		fmt.Println("CreateInvoiceTable:" + err.Error())
		return err
	}
	fmt.Println("CreateInvoiceTable Done")
	return nil
}

func (sdb *sqlDB) CreateARMAPTable() error {

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
			"OWNER to %s; ", sdb.user))

	if err != nil {
		fmt.Println("CreateARMAPTable:" + err.Error())
		return err
	}
	fmt.Println("CreateInvoiceTable Done")
	return nil
}

func (sdb *sqlDB) CreateReceiptTable() error {

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
			"OWNER to %s; ", sdb.user))

	if err != nil {
		fmt.Println("CreateReceiptTable:" + err.Error())
		return err
	}
	fmt.Println("CreateReceiptTable Done")
	return nil
}

func (sdb *sqlDB) CreateConfigBranchTable() error {

	_, err := sdb.SQLCommand(fmt.Sprintf(
		"CREATE TABLE public.ConfigBranch "+
			"( "+
			"Branch character varying(50) ,"+
			"Rent integer DEFAULT 0, "+
			"AgentSign integer DEFAULT 0, "+
			"CommercialFee double precision DEFAULT 0,"+
			"PRIMARY KEY (Branch) "+
			") "+
			"WITH ( OIDS = FALSE);"+ //))
			"ALTER TABLE public.ConfigBranch "+
			"OWNER to %s; ", sdb.user))

	if err != nil {
		fmt.Println("CreateConfigBranchTable:" + err.Error())
		return err
	}
	fmt.Println("CreateConfigBranchTable Done")
	return nil
}

func (sdb *sqlDB) CreateConfigParameterTable() error {

	_, err := sdb.SQLCommand(fmt.Sprintf(
		"CREATE TABLE public.ConfigParameter "+
			"( "+
			"date timestamp(0) without time zone not NULL,"+
			"NHI double precision DEFAULT 0, "+
			"LI double precision DEFAULT 0, "+
			"NHI2nd double precision DEFAULT 0, "+
			"IT double precision DEFAULT 0, "+
			"PRIMARY KEY (date) "+
			") "+
			"WITH ( OIDS = FALSE);"+ //))
			"ALTER TABLE public.ConfigParameter "+
			"OWNER to %s; ", sdb.user))

	if err != nil {
		fmt.Println("CreateConfigParameterTable:" + err.Error())
		return err
	}
	fmt.Println("CreateConfigParameterTable Done")
	return nil
}

func (sdb *sqlDB) CreateConfigSalerTable() error {

	_, err := sdb.SQLCommand(fmt.Sprintf(
		"CREATE TABLE public.ConfigSaler "+
			"( "+
			"Sid character varying(50) ,"+
			"Sname character varying(50) ,"+
			"Branch character varying(50) ,"+
			"ZeroDate timestamp(0) without time zone not NULL, "+
			"ValidDate  timestamp(0) without time zone not NULL, "+
			"Title character varying(50) ,"+
			"Percent double precision DEFAULT 0, "+
			"FPercent double precision DEFAULT 0, "+
			"Salary integer DEFAULT 0, "+
			"Pay integer DEFAULT 0, "+ //未來薪資
			"PayrollBracket integer DEFAULT 0, "+ //投保金額
			"Enrollment integer DEFAULT 0, "+ //加保(眷屬人數)
			"Association integer DEFAULT 0, "+
			"PRIMARY KEY (Sid, ZeroDate) "+
			") "+
			"WITH ( OIDS = FALSE);"+ //))
			"ALTER TABLE public.ConfigSaler "+
			"OWNER to %s; ", sdb.user))

	if err != nil {
		fmt.Println("CreateConfigSalerTable:" + err.Error())
		return err
	}
	fmt.Println("CreateConfigSalerTable Done")
	return nil
}

func (sdb *sqlDB) CreateCommissionTable() error {

	_, err := sdb.SQLCommand(fmt.Sprintf(
		"CREATE TABLE public.commission "+
			"( "+
			"Sid character varying(50) ,"+
			"Rid character varying(50) ,"+
			"ARid character varying(50) ,"+
			//"date timestamp(0) without time zone not NULL, "+
			"item character varying(50) not NULL, "+
			"SName character varying(50) , "+
			//"amount integer not NULL, "+
			//"fee integer DEFAULT 0, "+
			//"percent double precision DEFAULT 0 , "+
			"CPercent double precision DEFAULT 0 , "+
			"SR double precision DEFAULT 0 ,"+
			"bonus double precision DEFAULT 0 ,"+
			"PRIMARY KEY (Sid,Rid)"+
			") "+
			"WITH ( OIDS = FALSE);"+
			"ALTER TABLE public.commission "+
			"OWNER to %s; ", sdb.user))

	if err != nil {
		fmt.Println("CreateCommissionTable:" + err.Error())
		return err
	}
	fmt.Println("CreateCommissionTable Done")
	return nil
}

func (sdb *sqlDB) InitTable() error {

	rows, err := sdb.SQLCommand(`SELECT tablename FROM	pg_catalog.pg_tables 
								WHERE 
								schemaname != 'pg_catalog' AND schemaname != 'information_schema';`)

	if err != nil {
		return err
	}
	//mapTable
	var mT = map[string]bool{
		"ar":              false,
		"receipt":         false,
		"deduct":          false,
		"commission":      false,
		"amortization":    false,
		"pocket":          false,
		"branchprepay":    false,
		"prepay":          false,
		"accountitem":     false,
		"configbranch":    false,
		"configsaler":     false,
		"configparameter": false,
		"invoice":         false,
		"armap":           false,
		"salersalary":     false,
		"branchsalary":    false,
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
		case "configparameter":
			mT["configparameter"] = true
			break
		case "invoice":
			mT["invoice"] = true
			break
		case "armap":
			mT["armap"] = true
			break
		case "salersalary":
			mT["salersalary"] = true
			break
		case "branchsalary":
			mT["branchsalary"] = true
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
			case "configparameter":
				err = sdb.CreateConfigParameterTable()
				break
			case "invoice":
				err = sdb.CreateInvoiceTable()
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

func (sdb *sqlDB) InitDB() bool {

	rows, err := sdb.SQLCommand(fmt.Sprintf("SELECT datname FROM pg_database WHERE datname = '%s';", sdb.db))

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
