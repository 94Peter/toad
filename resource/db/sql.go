package db

import (
	"context"
	"database/sql"
	"fmt"

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

func (sdb *sqlDB) connectSQLDB() (*sql.DB, error) {
	//完整的資料格式連線如下
	//var connectionString string = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", sdb.dburl, sdb.port, sdb.user, sdb.password, sdb.db)
	var connectionString string = fmt.Sprintf("host=%s port=%d user=%s password=%s sslmode=disable", sdb.dburl, sdb.port, sdb.user, sdb.password)
	db, err := sql.Open("postgres", connectionString)

	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	//check database connection
	err = db.Ping()
	if err != nil {
		fmt.Println("[Failed] ping:" + err.Error())
		return nil, err
	}
	fmt.Println("連線成功")
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

func (sdb *sqlDB) Query(cmd string) (*sql.Rows, error) {

	c, err := sdb.connectSQLDB()

	if err != nil {
		return nil, err
	}

	rows, err := c.Query(cmd)
	fmt.Println("Query " + cmd)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return rows, err
}

func (sdb *sqlDB) CreateDB() error {

	_, err := sdb.Query(fmt.Sprintf(
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

func (sdb *sqlDB) IsDBExist() bool {

	rows, err := sdb.Query(fmt.Sprintf("SELECT datname FROM pg_database WHERE datname = '%s';", sdb.db))

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

	return false
}
