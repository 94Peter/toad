package model

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/94peter/pica/resource/db"
	"github.com/stretchr/testify/assert"
)

func Test_SaveReportItem(t *testing.T) {
	dbc := db.DBConf{}
	dbc.SetFirebase("firebaseServiceKey.json", "https://pica957.firebaseio.com/")
	var err error
	fireDB := dbc.GetDB()
	rim := &reportItemModel{
		db: fireDB,
	}
	for i := 0; i < 6; i++ {
		rim.Item = append(rim.Item,
			&ReportItem{
				Display: strconv.Itoa(i),
				Unit:    "件",
				HasText: i < 2,
				IsShow:  i != 5,
			},
		)
	}
	err = rim.Save()
	fmt.Println(err)
	assert.True(t, false)
}

func Test_LoadReportItem(t *testing.T) {
	dbc := db.DBConf{}
	dbc.SetFirebase("firebaseServiceKey.json", "https://pica957.firebaseio.com/")
	var err error
	fireDB := dbc.GetDB()
	rim := &reportItemModel{
		db: fireDB,
	}
	err = rim.Load()
	fmt.Println(err)
	j, _ := rim.Json()
	fmt.Println(string(j))
	assert.True(t, false)
}
