package model

import (
	"Contruction-Project/lib/utils"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4/pgxpool"
)

// SettingEnt ...
type SettingEnt struct {
	ID           int32
	SetGroup     string
	SetKey       string
	SetLabel     string
	SetOrder     int
	ContentType  string
	ContentValue string
	IsActive     bool
	CreatedDate  time.Time
	UpdatedDate  time.Time
}

var setType = []string{
	"json_arr",
	"json_obj",
	"bool",
	"str",
}

// GetByGroup ...
func (c *Contract) GetByGroup(db *pgxpool.Conn, ctx context.Context, gr string) ([]SettingEnt, error) {
	var s []SettingEnt
	var err error

	err = pgxscan.Select(ctx, db, &s, `select * from settings where set_group=$1`, gr)

	return s, err
}

// GetByGroupAndKey ...
func (c *Contract) GetByGroupAndKey(db *pgxpool.Conn, ctx context.Context, gr, key string) ([]SettingEnt, error) {
	var s []SettingEnt
	var err error

	err = pgxscan.Select(ctx, db, &s, `select * from settings where set_group=$1 AND set_key=$2`, gr, key)

	return s, err
}

// ToArray decode to array
func (s SettingEnt) ToArray() []map[string]interface{} {
	var data []map[string]interface{}
	_ = json.Unmarshal([]byte(s.ContentValue), &data)

	return data
}

// AddSetting
func (c *Contract) AddSetting(db *pgxpool.Conn, ctx context.Context, s SettingEnt) error {
	if !utils.Contains(setType, s.ContentType) {
		return fmt.Errorf("%s", "wrong content type")
	}

	sql := `insert into settings(set_group, set_key, set_label, set_order, content_type, content_value, is_active, created_date)
	values($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := db.Exec(context.Background(), sql,
		s.SetGroup, s.SetKey, s.SetLabel,
		s.ContentType, s.ContentValue,
		s.IsActive, time.Now().In(time.UTC),
	)

	return err
}

// UpdateSetting ...
func (c *Contract) UpdateSetting(db *pgxpool.Conn, ctx context.Context, id int32, s SettingEnt) error {
	sql := `update settings set set_label=$1, set_order=$2, content_type=$3, content_value=$4, is_active=$5, updated_date=$6
	where id=$`
	_, err := db.Exec(context.Background(), sql,
		s.SetLabel, s.SetOrder, s.ContentType, s.ContentValue,
		s.IsActive, time.Now().In(time.UTC),
	)

	return err
}

// ActivationSetting ...
func (c *Contract) ActivationSetting(db *pgxpool.Conn, ctx context.Context, id int32, isActive bool) error {
	bVal := utils.Bint(isActive)
	_, err := db.Exec(context.Background(), "update settings set is_active=$1 where id=$2", bVal, id)

	return err
}
