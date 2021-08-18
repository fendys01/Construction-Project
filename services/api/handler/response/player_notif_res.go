package response

import (
	"panorama/services/api/model"
)

// PlayerListResponse wraps the standard http.Response for the
// PlayersService.List method
type PlayerListResponse struct {
	TotalCount   int32  `json:"total_count"`
	Offset       int32  `json:"offset"`
	Limit        int32  `json:"limit"`
	ID           int32  `json:"id"`
	Identifier   string `json:"identifier"`
	SessionCount int    `json:"session_count"`
	Language     string `json:"language"`
	Timezone     int32  `json:"timezone"`
	DeviceOS     string `json:"device_os"`
}

func (r PlayerListResponse) Transform(i model.PlayerEnt) PlayerListResponse {

	r.TotalCount = i.TotalCount
	r.Offset = i.Offset
	r.Limit = i.Limit
	r.ID = i.ID
	r.Identifier = i.Identifier
	r.Language = i.Language
	r.Timezone = i.Timezone
	r.DeviceOS = i.DeviceOS

	return r
}

// PlayerCreateResponse wraps the standard http.Response for the
// PlayersService.Create method
type PlayerCreateResponse struct {
	// Success		bool   `json:"success"`
	// ID      	string `json:"id"`
	AppID       string `json:"app_id"`
	Language    string `json:"language"`
	Timezone    int32  `json:"timezone"`
	GameVersion string `json:"game_version"`
	DeviceOS    string `json:"device_os"`
	DeviceType  int32  `json:"device_type"`
	DeviceModel string `json:"device_model"`
}

func (r PlayerCreateResponse) Transform(m model.PlayerEnt) PlayerCreateResponse {
	// r.Success = m.Success
	// r.ID = m.IDSuccess
	r.AppID = m.AppID
	r.Language = m.Language
	r.Timezone = m.Timezone
	r.GameVersion = m.GameVersion
	r.DeviceOS = m.DeviceOS
	r.DeviceType = m.DeviceType
	r.DeviceModel = m.DeviceModel

	return r
}
