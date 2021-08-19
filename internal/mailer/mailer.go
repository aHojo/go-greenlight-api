package mailer

import (
	"bytes"
	"embed"
	"text/template"
	"time"

	"github.com/go-mail/mail/v2"
)

// Below we declare a new variable with the type embed.FS (embedded file system) to hold
// our email templates. This has a comment directive in the format `//go:embed <path>`
// IMMEDIATELY ABOVE it, which indicates to Go that we want to store the contents of the
// ./templates directory in the templateFS embedded file system variable.
// ↓↓↓

//go:embed "templates"
var templateFS embed.FS

// Mailer contains mail.Dialer instance
// used to connect to a SMTP server)
// And the sender information for the emails
type Mailer struct {
	dialer *mail.Dialer
	sender string
}

func New(host string, port int, username, password, sender string) Mailer {
	// Initialize a new mail.Dialer instance with the given SMTP server settings
	// We also configure this to use a 5-second timeout whenever we send an email
	dialer := mail.NewDialer(host, port, username, password)
	dialer.Timeout = 5 * time.Second

	return Mailer{
		dialer: dialer,
		sender: sender,
	}
}

// Send() takes a recipient email address
// name of the file with the templates,
// any dynamic data for the templates as an interface{}

func (m Mailer) Send(recipient, templateFile string, data interface{}) error {

	// Use the ParseFS() mthod to pars the required template file from the embedded file system
	tmpl, err := template.New("email").ParseFS(templateFS, "templates/"+templateFile)
	if err != nil {
		return err
	}

	// Execute the named template "subject", passing in the dynamic data and sorting the result in a buffer
	subject := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		return err
	}

	// Follow the same pattern to execute the "plainBody" template and store the result
	plainBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(plainBody, "plainBody", data)
	if err != nil {
		return err
	}
	htmlBody	:= new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(htmlBody, "htmlBody", data)
	if err != nil {
		return err
	}

  // Use the mail.NewMessage() function to initialize a new mail.Message instance. 
  // Then we use the SetHeader() method to set the email recipient, sender and subject
  // headers, the SetBody() method to set the plain-text body, and the AddAlternative()
  // method to set the HTML body. It's important to note that AddAlternative() should
  // always be called *after* SetBody().
	msg := mail.NewMessage()
	msg.SetHeader("To", recipient)
	msg.SetHeader("From", m.sender)
	msg.SetHeader("Subject", subject.String())
	msg.SetBody("text/plain", plainBody.String())
	msg.SetBody("text/html", htmlBody.String())

	// Call tehe DialAndSend 
	// Opens a connection to the SMTPServer and sends the message.
	// it also closes the  connection.
	err = m.dialer.DialAndSend(msg)
	if err != nil {
		return err
	}
	return nil

}
