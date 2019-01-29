package msg

import (
	"log"
	"testing"
)

func TestHTMLMail(t *testing.T) {
	sender := &Sender{
		User:   "xxxxxxxxx",
		Passwd: "xxxxxxxxx",
		Host:   "xxxxxxxxx",
		Port:   465,
	}
	sender.Configure()
	sw := sender.NewSendWorker(
		"xxx",
		"xxxx@xxxx",
		"testing",
	)

	err := sw.ParseTemplate(
		"activity.html",
		map[string]string{
			"year": "2019",
		},
	)

	if err != nil {
		log.Panicln(err)
	}

	err = sw.SendEmail()

	if err != nil {
		log.Panicln(err)
	}

	log.Println("mail module test passing")
}
