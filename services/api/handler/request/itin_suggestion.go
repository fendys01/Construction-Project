package request

import (
	"database/sql"
	"math/rand"
	"panorama/lib/utils"
	"panorama/services/api/model"
	"time"
)

type SugItinReq struct {
	Title       string                   `json:"title" validate:"required"`
	Content     string                   `json:"content" validate:"required"`
	Destination string                   `json:"destination" validate:"required"`
	Img         string                   `json:"img"`
	Details     []map[string]interface{} `json:"details" validate:"required"`
	Tags        []int                    `json:"tags"`
	NewTags     []string                 `json:"new_tags"`
}

func (req SugItinReq) ToSugItinEnt(isNew bool) (model.ItinSugEnt, error) {
	var code string

	if isNew {
		rand.Seed(time.Now().UnixNano())
		code, _ = utils.Generate(`SGIT-[a-z0-9]{3}`)
	}

	return model.ItinSugEnt{
		ItinCode:    code,
		Title:       req.Title,
		Content:     req.Content,
		Img:         sql.NullString{String: req.Img, Valid: true},
		Details:     req.Details,
		Destination: req.Destination,
	}, nil
}

// Transform Sug Itin
func (req SugItinReq) Transform(m model.ItinSugEnt) model.ItinSugEnt {

	if len(req.Title) > 0 {
		m.Title = req.Title
	}

	if len(req.Content) > 0 {
		m.Content = req.Content
	}

	if len(req.Img) > 0 {
		m.Img.String = req.Img
	}

	if len(req.Details) > 0 {
		m.Details = req.Details
	}

	return m
}
