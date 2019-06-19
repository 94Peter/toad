package sms

import (
	"fmt"

	twilio "github.com/kevinburke/twilio-go"
)

type myTwilio struct {
	account string
	token   string
	number  string
	client  *twilio.Client
}

func (my *myTwilio) connet() *twilio.Client {
	if my.client != nil {
		return my.client
	}
	my.client = twilio.NewClient(my.account, my.token, nil)
	return my.client
}

func (my *myTwilio) Message(to string, msg string) error {
	c := my.connet()

	result, err := c.Messages.SendMessage(my.number, to, msg, nil)
	if err != nil {
		return err
	}
	fmt.Println(result.Sid, result.FriendlyPrice())
	return nil
}
