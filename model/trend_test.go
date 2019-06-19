package model

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_GetDailyData(t *testing.T) {
	tm := GetTrendModel(&testDI{})
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day()-2, 0, 0, 0, 0, now.Location())
	result := tm.GetDailyData(today)
	jsonByte, _ := json.Marshal(result)
	fmt.Println(string(jsonByte))
	assert.True(t, false)
}

func Test_GetAreaData(t *testing.T) {
	tm := GetTrendModel(&testDI{})
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	end := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, now.Location())
	result := tm.GetAreaData(today, end)
	jsonByte, _ := result.ToJSON()
	fmt.Println(string(jsonByte))
	assert.True(t, false)
}

func Test_GetCategoryData(t *testing.T) {
	tm := GetTrendModel(&testDI{})
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	end := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, now.Location())
	result := tm.GetCategoryData("B", today, end)
	jsonByte, _ := result.ToJSON()
	fmt.Println(string(jsonByte))
	assert.True(t, false)
}
