package model

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_GetAnswer(t *testing.T) {
	sm := GetSalesModel(&testDI{})
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	result := sm.GetAnswer(today, "Peter")
	jsonByte, _ := json.Marshal(result)
	fmt.Println(string(jsonByte))
	assert.True(t, false)
}

func Test_addReportState(t *testing.T) {
	rs := reportState{}
	rs.addState("Peter", time.Now())
	fmt.Println(rs)

	newDate := time.Now().Add(time.Hour * 72)
	rs.addState("Chen", newDate)
	rs.addState("Peter", newDate)

	newDate = time.Now().Add(time.Hour * -72)
	fmt.Println(rs.addState("Chen", newDate))
	fmt.Println(rs.addState("Peter", newDate))
	fmt.Println(rs)
	assert.True(t, false)
}
