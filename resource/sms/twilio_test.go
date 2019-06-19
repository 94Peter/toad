package sms

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_SendMessage(t *testing.T) {
	conf := SMSConf{
		Account: "ACef7217fbd8c378d59d45becd7b75bb3d",
		Token:   "ce647fcea7a272cf7c3cff9ff8b5459e",
		Number:  "+12063397574",
	}

	sms := conf.GetSMS()
	err := sms.Message("+886919966667", "this is test.中文 http://tw.yahoo.com")
	fmt.Println(err)
	assert.True(t, false)
}
