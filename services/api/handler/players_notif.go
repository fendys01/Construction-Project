package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"panorama/services/api/handler/request"
	"panorama/services/api/handler/response"
	"panorama/services/api/model"

	"github.com/go-playground/validator/v10"
)

// PlayerCreateResponse wraps the standard http.Response for the
// PlayersService.Create method
type PlayerCreateResponse struct {
	Success bool   `json:"success"`
	ID      string `json:"id"`
}

// Create a player.
//
// OneSignal API docs:
// https://documentation.onesignal.com/docs/players-add-a-device

func (h *Contract) AddPlayersAct(w http.ResponseWriter, r *http.Request) {
			
	var err error
	req := request.PlayerRequest{}
	if err = h.Bind(r, &req); err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	if err = h.Validator.Driver.Struct(req); err != nil {
		h.SendRequestValidationError(w, err.(validator.ValidationErrors))
		return
	}

	payloadBytes, err := json.Marshal(req)
	if err != nil {
		return
	}
	body := bytes.NewReader(payloadBytes)
	
	request, err := http.NewRequest("POST", "https://onesignal.com/api/v1/players", body)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	request.Header.Set("Content-Type", "application/json")

	ctx := context.Background()
	db, err := h.DB.Acquire(ctx)
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}
	defer db.Release()
	
	m := model.Contract{App: h.App}
	players, err := m.AddPlayers(db, ctx, model.PlayerEnt{
		AppID:  		req.AppID,
		Identifier: 	req.Identifier,
		Language:       req.Language,
		Timezone:     	req.Timezone,
		GameVersion:	req.GameVersion,
		DeviceOS: 		req.DeviceOS,
		DeviceType: 	req.DeviceType,
		DeviceModel:	req.DeviceModel,
		CreatedDate: 	time.Time{},
	})
	if err != nil {
		h.SendBadRequest(w, err.Error())
		return
	}

	// plResp := &PlayerCreateResponse{}
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return 
	}
	
	defer resp.Body.Close()

	var res response.PlayerCreateResponse
	res = res.Transform(players)

	h.SendSuccess(w, res, nil)
}
