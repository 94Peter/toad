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
	fmt.Println("SQLCommand")
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
			"sales character varying(200), "+ //json[]
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
	fmt.Println("CreateTable Done")
	return nil
}

func (sdb *sqlDB) InitTable() error {

	rows, err := sdb.SQLCommand(`SELECT tablename FROM	pg_catalog.pg_tables 
								WHERE 
								schemaname != 'pg_catalog' AND schemaname != 'information_schema';`)

	if err != nil {
		return err
	}

	var mT = map[string]bool{
		"ar":      false,
		"receipt": false,
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
		default:
			fmt.Printf("unknown table %s.\n", tName)
		}
	}

	for t, s := range mT {
		//fmt.Println("i=", t, " s=", s)
		if !s {
			err = nil
			switch t {
			case "ar":
				err = sdb.CreateARTable()
			case "receipt":
				err = sdb.CreateReceiptTable()
			default:
				fmt.Printf("unknown table %s.\n", t)
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
		return true
	}

	if err := rows.Err(); err != nil {
		fmt.Println("err rows.Err() %s\n", err)
		return false
	}

	if err := sdb.CreateDB(); err != nil {
		fmt.Println("CreateDB Err() %s\n", err)
		return false
	}

	if err := sdb.InitTable(); err != nil {
		fmt.Println("InitTable Err() %s\n", err)
		return false
	}

	return false
}
