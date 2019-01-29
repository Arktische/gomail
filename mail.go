package msg

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/smtp"
)

// Sender global Sender config
type Sender struct {
	User   string
	Passwd string
	Host   string
	Port   int
	auth   smtp.Auth
}

// SendWorker worker process
type SendWorker struct {
	*Sender
	from    string
	to      string
	subject string
	body    string
	head    string
}

// NewSendWorker return a initialized SendWorker struct
func (s *Sender) NewSendWorker(from, to, subject string) *SendWorker {
	header := make(map[string]string)
	header["From"] = fmt.Sprintf("%s %s", from, s.User)
	header["To"] = to
	header["Subject"] = subject
	header["Content-Type"] = "text/html; charset=UTF-8"
	var message string
	for k, v := range header {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	return &SendWorker{
		Sender:  s,
		from:    s.User,
		to:      to,
		subject: subject,
		head:    message,
	}
}

// Configure make smtp plain auth
func (s *Sender) Configure() {
	s.auth = smtp.PlainAuth(
		"",
		s.User,
		s.Passwd,
		s.Host,
	)
	return
}

func (sw *SendWorker) ParseTemplate(fileName string, data interface{}) error {
	t, err := template.ParseFiles(fileName)
	if err != nil {
		return err
	}
	buffer := new(bytes.Buffer)
	if err = t.Execute(buffer, data); err != nil {
		return err
	}
	sw.body = buffer.String()
	return nil
}

// SendEmail send email
func (sw *SendWorker) SendEmail() error {
	url := fmt.Sprintf("%s:%d", sw.Sender.Host, sw.Port)

	err := SendMailUsingTLS(
		url,
		sw.auth,
		sw.from,
		[]string{sw.to},
		[]byte(sw.head+sw.body),
	)
	if err != nil {
		fmt.Println(err)
		return err
	}

	fmt.Println("Send mail success!")

	return nil
}

// Dial return a smtp client
func Dial(addr string) (*smtp.Client, error) {
	conn, err := tls.Dial("tcp", addr, nil)
	if err != nil {
		log.Println("Dialing Error:", err)
		return nil, err
	}
	//分解主机端口字符串
	host, _, _ := net.SplitHostPort(addr)
	return smtp.NewClient(conn, host)
}

//SendMailUsingTLS 参考net/smtp的func SendMail()
//使用net.Dial连接tls(ssl)端口时,smtp.NewClient()会卡住且不提示err
//len(to)>1时,to[1]开始提示是密送
func SendMailUsingTLS(addr string, auth smtp.Auth, from string,
	to []string, msg []byte) (err error) {

	//create smtp client
	c, err := Dial(addr)
	if err != nil {
		log.Println("Create smpt client error:", err)
		return err
	}
	defer c.Close()

	if auth != nil {
		if ok, _ := c.Extension("AUTH"); ok {
			if err = c.Auth(auth); err != nil {
				log.Println("Error during AUTH", err)
				return err
			}
		}
	}

	if err = c.Mail(from); err != nil {
		return err
	}

	for _, addr := range to {
		if err = c.Rcpt(addr); err != nil {
			return err
		}
	}

	w, err := c.Data()
	if err != nil {
		return err
	}

	_, err = w.Write(msg)
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	return c.Quit()
}
