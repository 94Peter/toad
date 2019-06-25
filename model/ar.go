package model

import (
	"time"

	"github.com/94peter/toad/resource/db"
)

type AR struct {
	ARid         string
	Date         time.Time
	cNo          string
	caseName     string
	customertype string
	name         string
	amount       int
	fee          int
	RA           int
	balance      int
	sales        []*sale
}

type sale struct {
	bName   string
	percent float64
	Bid     string
}

type ARModel struct {
	db db.InterSQLDB
	//res interModelRes

	AR []*AR
}
