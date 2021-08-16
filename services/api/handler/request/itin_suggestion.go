package request

import (
	"database/sql"
	"math/rand"
	"panorama/lib/utils"
	"panorama/services/api/model"
	"time"
)

type SugItinReq struct {
	Title     string                   `json:"title"`
	Content   string                   `json:"content"`
	Price     int                      `json:"price"`
	StartDate string                   `json:"start_date"`
	EndDate   string                   `json:"end_date"`
	Img       string                   `json:"img"`
	Details   []map[string]interface{} `json:"details"`
	Tags      []int                    `json:"tags"`
	NewTags   []string                 `json:"new_tags"`
}

func (req SugItinReq) ToSugItinEnt(isNew bool) (model.ItinSugEnt, error) {
	var code string

	if isNew {
		rand.Seed(time.Now().UnixNano())
		code, _ = utils.Generate(`admin[a-z0-9]{3}`)
	}
	sDate, err := time.Parse("2006-01-02 15:04:05", req.StartDate)
	if err != nil {
		return model.ItinSugEnt{}, err
	}
	eDate, err := time.Parse("2006-01-02 15:04:05", req.EndDate)
	if err != nil {
		return model.ItinSugEnt{}, err
	}

	return model.ItinSugEnt{
		ItinCode:  code,
		Title:     req.Title,
		Content:   req.Content,
		Price:     int64(req.Price),
		StartDate: sDate,
		EndDate:   eDate,
		Img:       sql.NullString{String: req.Img, Valid: true},
		Details:   req.Details,
	}, nil
}
