package smtpsrv_test

import (
	"fmt"
	"github.com/phires/go-guerrilla"
	"github.com/phires/go-guerrilla/backends"
	"github.com/phires/go-guerrilla/log"
	"github.com/phires/go-guerrilla/mail"
	"net/smtp"
	"strings"
	"testing"
	"time"
)

func TestServer(t *testing.T) {

	d, err := StartServer(&guerrilla.AppConfig{
		LogFile: log.OutputStdout.String(),
		Servers: []guerrilla.ServerConfig{
			{
				ListenInterface: "127.0.0.1:2526",
				IsEnabled:       true,
			},
		},
		BackendConfig: backends.BackendConfig{
			"save_workers_size":   3,
			"save_process":        "HeadersParser|Header|Hasher|Debugger|SmimeExtract",
			"log_received_mails":  true,
			"primary_mail_host":   "example.com",
			"decryption_key_path": "/tmp/cica",
		},
	})
	if err != nil {
		t.Fatalf("%+v", err)
	}
	defer d.Shutdown()

	d.AddProcessor("smimeExtract", MyFooProcessor)

	time.Sleep(time.Second * 2)

}

func TestSendMail(t *testing.T) {
	// Start the SMTP server
	d, err := StartServer(&guerrilla.AppConfig{
		LogFile: log.OutputStdout.String(),
		Servers: []guerrilla.ServerConfig{
			{
				ListenInterface: "127.0.0.1:2526",
				IsEnabled:       true,
			},
		},
		BackendConfig: backends.BackendConfig{
			"save_workers_size":   3,
			"save_process":        "HeadersParser|Header|Hasher|Debugger|SmimeExtract",
			"log_received_mails":  true,
			"primary_mail_host":   "example.com",
			"decryption_key_path": "/tmp/cica",
		},
		AllowedHosts: []string{"127.0.0.1", "example.com"},
	})
	if err != nil {
		t.Fatalf("%+v", err)
	}
	defer d.Shutdown()

	// Give the server some time to start
	time.Sleep(time.Second * 2)

	// Set up the email parameters
	from := "sender@example.com"
	to := []string{"recipient@example.com"}
	smtpHost := "127.0.0.1"
	smtpPort := "2526"

	// Compose the message
	subject := "Test Email"
	body := "This is a test email sent to the SMTP server."
	message := []byte(fmt.Sprintf("To: %s\r\nFrom: %s\r\nSubject: %s\r\n\r\n%s",
		strings.Join(to, ", "), from, subject, body))

	// Connect to the SMTP server and send the email
	err = smtp.SendMail(
		smtpHost+":"+smtpPort,
		nil, // No authentication
		from,
		to,
		message,
	)
	if err != nil {
		t.Fatalf("Failed to send email: %v", err)
	}

	// Give the server some time to process the email
	time.Sleep(time.Second * 2)

	t.Log("Email sent successfully")
}

func ProcessMail(envelop *mail.Envelope, selectTask backends.SelectTask) (backends.Result, error) {
	return nil, nil
}

func StartServer(cfg *guerrilla.AppConfig) (*guerrilla.Daemon, error) {
	d := guerrilla.Daemon{Config: cfg}
	d.AddProcessor("SmimeExtract", MyFooProcessor)

	if err := d.Start(); err != nil {
		return nil, err
	}
	return &d, nil
}

var MyFooProcessor = func() backends.Decorator {
	var decryptionKey string

	// our initFunc will load the config.
	initFunc := backends.InitializeWith(func(backendConfig backends.BackendConfig) error {
		if str, ok := backendConfig["decryption_key_path"].(string); ok {
			decryptionKey = str

			return nil
		} else {
			return fmt.Errorf("decryption_key_path is not a string")
		}
	})
	// register our initializer
	backends.Svc.AddInitializer(initFunc)

	return func(p backends.Processor) backends.Processor {
		return backends.ProcessWith(
			func(e *mail.Envelope, task backends.SelectTask) (backends.Result, error) {
				if task == backends.TaskValidateRcpt {
					_ = decryptionKey
					// optionally, validate recipient
					return p.Process(e, task)
				} else if task == backends.TaskSaveMail {
					/*
						do some work here..
						if want to stop processing, return

						 errors.New("Something went wrong")
						 return backends.NewBackendResult(fmt.Sprintf("554 Error: %s", err)), err
					*/

					// call the next processor in the chain
					return p.Process(e, task)
				}
				return p.Process(e, task)
			},
		)
	}
}
