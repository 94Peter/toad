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
			"fee integer DEFAULT 0, "+
			"RA integer DEFAULT 0, "+
			"balance integer DEFAULT 0, "+
			"sales character varying(600), "+ //json[]
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
func (sdb *sqlDB) CreateSalaryTable() error {
	_, err := sdb.SQLCommand(fmt.Sprintf(
		"CREATE TABLE public.Salary "+
			"( "+
			"Bid character varying(50) ,"+
			"date timestamp(0) without time zone not NULL, "+
			"cNo character varying(50) not NULL, "+
			"caseName character varying(50) not NULL, "+
			"type character varying(50) not NULL, "+
			"BName character varying(50) not NULL, "+
			"amount integer not NULL, "+
			"invoiceNo character varying(50) , "+
			"ARid character varying(50) not NULL, "+
			"PRIMARY KEY (Rid) "+
			") "+
			"WITH ( OIDS = FALSE);"+ //))
			"ALTER TABLE public.receipt "+
			"OWNER to %s; ", sdb.user))

	if err != nil {
		fmt.Println("CreateSalaryTable:" + err.Error())
		return err
	}
	fmt.Println("CreateSalaryTable Done")
	return nil
}

func (sdb *sqlDB) CreatePrePayTable() error {

	_, err := sdb.SQLCommand(fmt.Sprintf(
		"CREATE TABLE public.PrePay "+
			"( "+
			"PPid character varying(50) ,"+
			"date timestamp(0) without time zone not NULL, "+
			"itemname character varying(50) ,"+
			"describe character varying(50) ,"+
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
			"describe character varying(50) ,"+
			"Income  integer DEFAULT 0, "+
			"Fee  integer DEFAULT 0, "+
			"Balance  integer not NULL, "+
			"PRIMARY KEY (Pid) "+
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

func (sdb *sqlDB) CreateReceiptTable() error {

	_, err := sdb.SQLCommand(fmt.Sprintf(
		"CREATE TABLE public.receipt "+
			"( "+
			"Rid character varying(50) ,"+
			"date timestamp(0) without time zone not NULL, "+
			"cNo character varying(50) not NULL, "+
			"caseName character varying(50) not NULL, "+
			"type character varying(50) not NULL, "+
			"name character varying(50) not NULL, "+
			"amount integer not NULL, "+
			"invoiceNo character varying(50) , "+
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
			"param character varying(50) ,"+
			"value double precision DEFAULT 0, "+
			"PRIMARY KEY (param) "+
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

func (sdb *sqlDB) CreateConfigBusinessTable() error {

	_, err := sdb.SQLCommand(fmt.Sprintf(
		"CREATE TABLE public.ConfigBusiness "+
			"( "+
			"Bid character varying(50) ,"+
			"bname character varying(50) ,"+
			"ZeroDate timestamp(0) without time zone not NULL, "+
			"ValidDate  timestamp(0) without time zone not NULL, "+
			"Title character varying(50) ,"+
			"Percent  double precision DEFAULT 0, "+
			"Salary integer DEFAULT 0, "+
			"Pay integer DEFAULT 0, "+
			"PRIMARY KEY (Bid) "+
			") "+
			"WITH ( OIDS = FALSE);"+ //))
			"ALTER TABLE public.ConfigBusiness "+
			"OWNER to %s; ", sdb.user))

	if err != nil {
		fmt.Println("CreateConfigBusinessTable:" + err.Error())
		return err
	}
	fmt.Println("CreateConfigBusinessTable Done")
	return nil
}

func (sdb *sqlDB) CreateCommissionTable() error {

	_, err := sdb.SQLCommand(fmt.Sprintf(
		"CREATE TABLE public.commission "+
			"( "+
			"Bid character varying(50) ,"+
			"Rid character varying(50) ,"+
			"date timestamp(0) without time zone not NULL, "+
			"item character varying(50) not NULL, "+
			"bname character varying(50) , "+
			"amount integer not NULL, "+
			"fee integer DEFAULT 0, "+
			//"percent double precision DEFAULT 0 , "+
			"percent integer DEFAULT 0 , "+
			"SR integer DEFAULT 0 ,"+
			"bouns integer DEFAULT 0 ,"+
			"PRIMARY KEY (Bid,Rid)"+
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
		"commission":      false,
		"amortization":    false,
		"pocket":          false,
		"branchprepay":    false,
		"prepay":          false,
		"accountitem":     false,
		"configbranch":    false,
		"configbusiness":  false,
		"configparameter": false,
	}

	for rows.Next() {
		var tName string
		if err := rows.Scan(&tName); err != nil {
			fmt.Println("err Scan %s\n", err)
		}

		switch tName {
		case "ar":
			mT["ar"] = true
		case "receipt":
			mT["receipt"] = true
		case "commission":
			mT["commission"] = true
		case "amortization":
			mT["amortization"] = true
		case "pocket":
			mT["pocket"] = true
		case "branchprepay":
			mT["branchprepay"] = true
		case "prepay":
			mT["prepay"] = true
		case "accountitem":
			mT["accountitem"] = true
		case "configbranch":
			mT["configbranch"] = true
		case "configbusiness":
			mT["configbusiness"] = true
		case "configparameter":
			mT["configparameter"] = true

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
			case "receipt":
				err = sdb.CreateReceiptTable()
			case "commission":
				err = sdb.CreateCommissionTable()
			case "amortization":
				err = sdb.CreateAmortizationTable()
			case "accountitem":
				err = sdb.CreateAccountItemTable()
			case "pocket":
				err = sdb.CreatePocketTable()
			case "prepay":
				err = sdb.CreatePrePayTable()
			case "branchprepay":
				err = sdb.CreateBranchPrePayTable()
			case "configbranch":
				err = sdb.CreateConfigBranchTable()
			case "configbusiness":
				err = sdb.CreateConfigBusinessTable()
			case "configparameter":
				err = sdb.CreateConfigParameterTable()
			default:
				fmt.Printf("unknown table %s.\n", tableName)
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
