package citcall

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"panorama/bootstrap"
)

const (
	APP_NAME = "Panorama"
	URL_HOST = "http://104.199.196.122/gateway"
	VERSION  = "v3"
)

type service struct {
	app *bootstrap.App
}

func New(app *bootstrap.App) *service {
	return &service{app}
}

func getURLByType(typePrefix string) string {
	url := fmt.Sprintf("%s/%s/%s", URL_HOST, VERSION, typePrefix)
	return url
}

func getOTPMessage(token string, expiredTime int) string {
	return fmt.Sprintf("Your One Time Password (OTP) for Panorama App is %s and valid for %d minutes. If you did not initiate this, call our Call Center at (021) 2556 5000", token, expiredTime)
}

func (s *service) SendOTP(phone, token string, expiredTime int) error {
	textMessage := getOTPMessage(token, expiredTime)
	apiKey := s.app.Config.GetString("citcall.api_key")
	url := getURLByType(s.app.Config.GetString("citcall.sms_type"))

	values := map[string]string{
		"msisdn":   phone,
		"senderid": APP_NAME,
		"text":     textMessage,
	}

	dataValues, err := json.Marshal(values)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(dataValues))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Apikey "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	log.Printf("Citcall API Key: %s\n", apiKey)
	log.Println("Response Status:", resp.Status)
	log.Println("Response Body:", string(body))

	return nil
}
