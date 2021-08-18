package request

// PlayerRequest represents a request to create/update a player.
type PlayerRequest struct {
	AppID             string            `json:"app_id"`
	Identifier        string            `json:"identifier,omitempty"`
	Language          string            `json:"language,omitempty"`
	Timezone          int32 			`json:"timezone" validate:"required"`
	GameVersion		  string			`json:"game_version"`
	DeviceType        int32             `json:"device_type"`
	DeviceOS          string            `json:"device_os,omitempty"`
	DeviceModel       string            `json:"device_model,omitempty"`
	Tags			  Tags				`json:"tags"`
}

type Tags struct {
	Foo string `json:"foo"`
}