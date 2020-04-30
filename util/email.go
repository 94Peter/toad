package util

import (
	"fmt"
	"strconv"

	"gopkg.in/gomail.v2"
)

// define email interface, and implemented auth and send method
// type Mail interface {
// 	Auth()
// 	Send(message Message) error
// }

type SendMail struct {
	User     string
	Password string
	Host     string
	Port     string
	//auth     smtp.Auth
}

// type Attachment struct {
// 	name        string
// 	contentType string
// 	withFile    bool
// }

// type Message struct {
// 	from        string
// 	to          []string
// 	cc          []string
// 	bcc         []string
// 	subject     string
// 	body        string
// 	contentType string
// 	attachment  Attachment
// }

func GomailMailSend(smtpHost, smtpPort, smtpPass, smtpUser, to, subject, body string, fNames ...string) {
	m := gomail.NewMessage()
	m.SetHeader("From", "test@example.com")
	m.SetHeader("To", to)
	//m.SetHeader("To", to)
	//m.SetAddressHeader("Cc", "crgcrg0034@gmail.com", "system")
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)
	for _, f := range fNames {
		if f != "" {
			m.Attach(f)
		}
	}

	// smtpConf:
	// user: crgcrg0034@gmail.com
	// password: nqnbzmrmywrtvyyv
	// host: smtp.gmail.com
	// port: "587"
	//d := gomail.NewDialer("smtp.gmail.com", 587, "crgcrg0034@gmail.com", "nqnbzmrmywrtvyyv")
	port, _ := strconv.Atoi(smtpPort)
	d := gomail.NewDialer(smtpHost, port, smtpUser, smtpPass)

	// Send the email to Bob, Cora and Dan.
	if err := d.DialAndSend(m); err != nil {
		fmt.Println(err)
		return
	}
}

// func RunSendMail(smtpHost, smtpPort, smtpPass, smtpUser, to, subject, body, fname string) {
// 	return
// 	var mail Mail
// 	// smtpHost := "smtp.gmail.com"       // 你可以改为其他的
// 	// smtpPort := "587"                  // 端口
// 	// smtpPass := "nqnbzmrmywrtvyyv"     // 密码
// 	// smtpUser := "crgcrg0034@gmail.com" // 用户

// 	mail = &SendMail{User: smtpUser, Password: smtpPass, Host: smtpHost, Port: smtpPort}
// 	message := Message{from: "test@test.com",
// 		//to:          []string{to},
// 		to: []string{"geassyayaoo3@gmail.com"},
// 		//cc:          []string{"a3162858@gmail.com", "ch.focke@gmail.com", "crgcrg0034@gmail.com"}, //cc功能是無作用的，寄不出去...
// 		bcc:         []string{"crgcrg0034@gmail.com"},
// 		subject:     subject,
// 		body:        body,
// 		contentType: "text/plain;charset=utf-8",
// 		attachment: Attachment{
// 			name: PdfDir + fname,
// 			//contentType: "image/png",
// 			withFile: true,
// 		},
// 	}
// 	mail.Send(message)
// }

// func main() {
// 	var mail Mail
// 	smtpHost := "smtp.gmail.com"       // 你可以改为其他的
// 	smtpPort := "587"                  // 端口
// 	smtpPass := "nqnbzmrmywrtvyyv"     // 密码
// 	smtpUser := "crgcrg0034@gmail.com" // 用户

// 	mail = &SendMail{User: smtpUser, Password: smtpPass, Host: smtpHost, Port: smtpPort}
// 	message := Message{from: "test@test.com",
// 		to:          []string{"geassyayaoo3@gmail.com"},
// 		cc:          []string{},
// 		bcc:         []string{},
// 		subject:     "HELLO WORLD",
// 		body:        "",
// 		contentType: "text/plain;charset=utf-8",
// 		attachment: Attachment{
// 			name: "hello.pdf",
// 			//contentType: "image/png",
// 			withFile: true,
// 		},
// 	}
// 	mail.Send(message)
// }

// func (mail *SendMail) Auth() {
// 	mail.auth = smtp.PlainAuth("", mail.User, mail.Password, mail.Host)
// }

// func (mail SendMail) Send(message Message) error {
// 	mail.Auth()
// 	buffer := bytes.NewBuffer(nil)
// 	boundary := "GoBoundary"
// 	Header := make(map[string]string)
// 	Header["From"] = message.from
// 	Header["To"] = strings.Join(message.to, ";")
// 	Header["Cc"] = strings.Join(message.cc, ";")
// 	Header["Bcc"] = strings.Join(message.bcc, ";")
// 	Header["Subject"] = message.subject
// 	Header["Content-Type"] = "multipart/mixed;boundary=" + boundary
// 	Header["Mime-Version"] = "1.0"
// 	Header["Date"] = time.Now().String()
// 	mail.writeHeader(buffer, Header)

// 	body := "\r\n--" + boundary + "\r\n"
// 	body += "Content-Type:" + message.contentType + "\r\n"
// 	body += "\r\n" + message.body + "\r\n"
// 	buffer.WriteString(body)

// 	if message.attachment.withFile {
// 		attachment := "\r\n--" + boundary + "\r\n"
// 		attachment += "Content-Transfer-Encoding:base64\r\n"
// 		attachment += "Content-Disposition:attachment\r\n"
// 		attachment += "Content-Type:" + message.attachment.contentType + ";name=\"" + strings.Replace(message.attachment.name, PdfDir, "", -1) + "\"\r\n" //更改信箱中看到的檔案名稱
// 		buffer.WriteString(attachment)
// 		defer func() {
// 			if err := recover(); err != nil {
// 				log.Fatalln(err)
// 			}
// 		}()
// 		mail.writeFile(buffer, message.attachment.name)
// 	}

// 	buffer.WriteString("\r\n--" + boundary + "--")
// 	err := smtp.SendMail(mail.Host+":"+mail.Port, mail.auth, message.from, message.to, buffer.Bytes())
// 	if err != nil {
// 		fmt.Println(err)
// 	}
// 	return nil
// }

// func (mail SendMail) writeHeader(buffer *bytes.Buffer, Header map[string]string) string {
// 	header := ""
// 	for key, value := range Header {
// 		header += key + ":" + value + "\r\n"
// 	}
// 	header += "\r\n"
// 	buffer.WriteString(header)
// 	return header
// }

// // read and write the file to buffer
// func (mail SendMail) writeFile(buffer *bytes.Buffer, fileName string) {
// 	file, err := ioutil.ReadFile(fileName)
// 	if err != nil {
// 		panic(err.Error())
// 	}
// 	payload := make([]byte, base64.StdEncoding.EncodedLen(len(file)))
// 	base64.StdEncoding.Encode(payload, file)
// 	buffer.WriteString("\r\n")
// 	for index, line := 0, len(payload); index < line; index++ {
// 		buffer.WriteByte(payload[index])
// 		if (index+1)%76 == 0 {
// 			buffer.WriteString("\r\n")
// 		}
// 	}
// }
