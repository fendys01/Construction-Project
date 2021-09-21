package onesignal

import (
	"encoding/json"
	"fmt"
	"net/http"
	"panorama/bootstrap"
	"panorama/lib/utils"
)

// device_type list reference https://documentation.onesignal.com/reference/add-a-device
const (
	URL_HOST                      = "https://onesignal.com/api"
	VERSION                       = "v1"
	API_PREFIX_DEVICE             = "players"
	API_PREFIX_NOTIFICATION       = "notifications"
	DEVICE_TYPE_IOS               = 0
	DEVICE_TYPE_ANDROID           = 1
	DEVICE_TYPE_BROWSER_CHROMEAPP = 4
	DEVICE_TYPE_BROWSER_CHROMEWEB = 5
	DEVICE_TYPE_BROWSER_SAFARI    = 7
	DEVICE_TYPE_BROWSER_FIREFOX   = 8
)

type service struct {
	app *bootstrap.App
}

func New(app *bootstrap.App) *service {
	return &service{app}
}

func getURLByPrefix(prefix string) string {
	url := fmt.Sprintf("%s/%s/%s", URL_HOST, VERSION, prefix)
	return url
}

func (s *service) getApiKey() string {
	return s.app.Config.GetString("onesignal.api_key")
}

func (s *service) getAppID() string {
	return s.app.Config.GetString("onesignal.app_id")
}

// Player represents a OneSignal player.
type Player struct {
	ID                string            `json:"id"`
	Playtime          int               `json:"playtime"`
	SDK               string            `json:"sdk"`
	Identifier        string            `json:"identifier"`
	SessionCount      int               `json:"session_count"`
	Language          string            `json:"language"`
	Timezone          int               `json:"timezone"`
	GameVersion       string            `json:"game_version"`
	DeviceOS          string            `json:"device_os"`
	DeviceType        int               `json:"device_type"`
	DeviceModel       string            `json:"device_model"`
	AdID              string            `json:"ad_id"`
	Tags              map[string]string `json:"tags"`
	LastActive        int               `json:"last_active"`
	AmountSpent       float32           `json:"amount_spent"`
	CreatedAt         int               `json:"created_at"`
	InvalidIdentifier bool              `json:"invalid_identifier"`
	BadgeCount        int               `json:"badge_count"`
}

func (s *service) AddDevice(deviceType int) (string, error) {
	player := Player{}
	url := getURLByPrefix(API_PREFIX_DEVICE)

	values := map[string]interface{}{
		"app_id":      s.getAppID(),
		"device_type": deviceType,
	}

	request, err := utils.RequestHandler(values, url, http.MethodPost)
	if err != nil {
		return "", err
	}
	request.Header.Set("Content-Type", "application/json; charset=utf-8")

	response, err := utils.ResponseAsyncHandler(request)
	if err != nil {
		return "", err
	}

	json.Unmarshal([]byte(response[0]), &player)

	return player.ID, err
}

func (s *service) PushNotification(header, content string, playerID []string) (map[string]interface{}, error) {
	var callback map[string]interface{}
	url := getURLByPrefix(API_PREFIX_NOTIFICATION)

	values := map[string]interface{}{
		"app_id":             s.getAppID(),
		"headings":           map[string]interface{}{"en": header},
		"contents":           map[string]interface{}{"en": content},
		"include_player_ids": playerID,
	}

	request, err := utils.RequestHandler(values, url, http.MethodPost)
	if err != nil {
		return callback, err
	}
	request.Header.Set("Authorization", "Basic "+s.getApiKey())
	request.Header.Set("Content-Type", "application/json; charset=utf-8")

	response, err := utils.ResponseAsyncHandler(request)
	if err != nil {
		return callback, err
	}

	json.Unmarshal([]byte(response[0]), &callback)

	utils.ResponseHandler(request)

	return callback, err
}

func (s *service) GetPlayerDevice(playerID string) (Player, error) {
	player := Player{}
	url := fmt.Sprintf("%s/%s", getURLByPrefix(API_PREFIX_DEVICE), playerID)

	request, err := utils.RequestHandler(nil, url, http.MethodGet)
	if err != nil {
		return player, err
	}
	request.Header.Set("Authorization", "Basic "+s.getApiKey())
	request.Header.Set("Content-Type", "application/json; charset=utf-8")

	response, err := utils.ResponseHandler(request)
	if err != nil {
		return player, err
	}
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		return player, err
	}

	json.Unmarshal(jsonResponse, &player)

	return player, err
}
