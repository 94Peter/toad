package sms

type InterSMS interface {
	Message(to string, msg string) error
}

type SMSConf struct {
	Account string
	Token   string
	Number  string
}

func (conf *SMSConf) GetSMS() InterSMS {
	return &myTwilio{
		account: conf.Account,
		token:   conf.Token,
		number:  conf.Number,
	}
}
