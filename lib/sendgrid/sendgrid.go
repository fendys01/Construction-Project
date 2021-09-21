package sendgrid

import (
	"encoding/json"
	"fmt"
	"net/http"

	"panorama/bootstrap"
	"panorama/lib/utils"
)

const (
	URL_HOST = "https://api.sendgrid.com"
	VERSION  = "v3"
)

type service struct {
	app *bootstrap.App
}

func New(app *bootstrap.App) *service {
	return &service{app}
}

func getURLByURI(URI string) string {
	url := fmt.Sprintf("%s/%s/%s", URL_HOST, VERSION, URI)
	return url
}

func (s *service) getApiKey() string {
	return s.app.Config.GetString("sendgrid.api_key")
}

func (s *service) SetMailSenderContent(subject, to, htmlTemplate string) MailPersonalizations {
	from := fmt.Sprintf("%s <%s>", s.app.Config.GetString("mail.mail_name"), s.app.Config.GetString("mail.mail_from"))

	return MailPersonalizations{
		Subject: subject,
		From: MailToRequest{
			Email: from,
		},
		Personalizations: []MailSendRequest{
			{
				To: []MailToRequest{
					{
						Email: to,
					},
				},
			},
		},
		Content: []MailContentRequest{
			{
				Type:  "text/html",
				Value: htmlTemplate,
			},
		},
	}
}

func (s *service) MailSender(subject, to, htmlTemplate string) (map[string]interface{}, error) {
	var callback map[string]interface{}

	url := getURLByURI("/mail/send")
	values := s.SetMailSenderContent(subject, to, htmlTemplate)

	request, err := utils.RequestHandlerEntity(values, url, http.MethodPost)
	if err != nil {
		return callback, err
	}
	request.Header.Add("Authorization", "Bearer "+s.getApiKey())
	request.Header.Add("Content-Type", "application/json")

	response, err := utils.ResponseAsyncHandler(request)
	if err != nil {
		return callback, err
	}

	json.Unmarshal([]byte(response[0]), &callback)

	utils.ResponseHandler(request)

	return callback, err
}
