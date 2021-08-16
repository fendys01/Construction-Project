package request

import (
	"database/sql"
	"math/rand"
	"panorama/lib/utils"
	"panorama/services/api/model"
	"time"
)

type MemberItinReq struct {
	Title     string                   `json:"title"`
	Content   string                   `json:"content"`
	EstPrice  int                      `json:"price"`
	StartDate string                   `json:"start_date"`
	EndDate   string                   `json:"end_date"`
	Details   []map[string]interface{} `json:"details"`
	Tags      []int                    `json:"tags"`
	NewTags   []string                 `json:"new_tags"`
}

func (req MemberItinReq) ToMemberItinEnt(isNew bool) (model.MemberItinEnt, error) {
	var code string

	if isNew {
		rand.Seed(time.Now().UnixNano())
		code, _ = utils.Generate(`admin[a-z0-9]{3}`)
	}
	sDate, err := time.Parse("2006-01-02 15:04:05", req.StartDate)
	if err != nil {
		return model.MemberItinEnt{}, err
	}
	eDate, err := time.Parse("2006-01-02 15:04:05", req.EndDate)
	if err != nil {
		return model.MemberItinEnt{}, err
	}
	return model.MemberItinEnt{
		Code:      code,
		Title:     req.Title,
		EstPrice:  sql.NullInt64{Int64: int64(req.EstPrice), Valid: true},
		StartDate: sDate,
		EndDate:   eDate,
		Details:   req.Details,
	}, nil
}
