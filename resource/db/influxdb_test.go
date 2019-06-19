package db

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_SaveInflux(t *testing.T) {
	dbConf := &DBConf{
		InfluxDBConf: &tsdbConf{
			DB:   "test",
			Host: "localhost:8086",
		},
	}
	idb := dbConf.GetTSDB()

	err := idb.Save(&testPoint{})
	fmt.Println(err)
	assert.True(t, false)
}

type testPoint struct {
}

func (tp *testPoint) GetMeasurement() string {
	return "testM"
}
func (tp *testPoint) GetTags() map[string]string {
	return nil
}
func (tp *testPoint) GetFields() map[string]interface{} {
	return map[string]interface{}{
		"value": 10,
	}
}
func (tp *testPoint) GetTime() time.Time {
	return time.Now()
}
