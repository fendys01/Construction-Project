package citcall

import (
	"encoding/json"
	"fmt"
	"net/http"
	"panorama/bootstrap"
	"panorama/lib/utils"
)

const (
	APP_NAME      = "Panorama"
	URL_HOST      = "http://104.199.196.122/gateway"
	VERSION       = "v3"
	STATUS_OK     = 0
	METHOD_SMS    = "sms"
	METHOD_SMSOTP = "smsotp"
)

type service struct {
	app *bootstrap.App
}

type ResponseStatus struct {
	RC   int
	Info string
}

func New(app *bootstrap.App) *service {
	return &service{app}
}

func getURLByType(typePrefix string) string {
	url := fmt.Sprintf("%s/%s/%s", URL_HOST, VERSION, typePrefix)
	return url
}

func getOTPMessage(token string, expiredTime int) string {
	return fmt.Sprintf("Your One Time Password (OTP) is %s and valid for %d minutes. If you did not initiate this, call our Call Center at (021) 2556 5000", token, expiredTime)
}

func (s *service) getApiKey() string {
	return s.app.Config.GetString("citcall.api_key")
}

func (s *service) SendOTP(phone, token string, expiredTime int) (ResponseStatus, error) {
	responseStatus := ResponseStatus{}
	textMessage := getOTPMessage(token, expiredTime)
	// url := getURLByType(s.app.Config.GetString("citcall.sms_type"))
	url := getURLByType(METHOD_SMSOTP)

	values := map[string]interface{}{
		"msisdn":   phone,
		"senderid": APP_NAME,
		"text":     textMessage,
	}

	request, err := utils.RequestHandler(values, url, http.MethodPost)
	if err != nil {
		return responseStatus, err
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Apikey "+s.getApiKey())

	response, err := utils.ResponseAsyncHandler(request)
	if err != nil {
		return responseStatus, err
	}

	json.Unmarshal([]byte(response[0]), &responseStatus)

	return responseStatus, err
}
