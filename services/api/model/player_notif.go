package model

import (
	"context"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

// Player represents a OneSignal player.
type PlayerEnt struct {
	ID                int32            
	AppID			  string			
	Identifier        string            
	Language          string           
	Timezone          int32
	GameVersion		  string               
	DeviceOS          string            
	DeviceType        int32              
	DeviceModel       string            
	CreatedDate       time.Time
	TotalCount 		  int32 
	Offset     		  int32 			
	Limit      		  int32
	Success 		  bool
	IDSuccess		  string 		
}

func (c *Contract) AddPlayers(db *pgxpool.Conn, ctx context.Context, p PlayerEnt) (PlayerEnt, error)  {
	var lastInsID int32
	err := db.QueryRow(ctx, `insert into device_list (app_id, identifier, language, game_version, timezone, device_os, device_type, device_model, created_date) 
		values($1, $2, $3, $4, $5, $6, $7,$8, $9) RETURNING id`,
		p.AppID, p.Identifier, p.Language,p.GameVersion,  p.Timezone, p.DeviceOS, p.DeviceType, p.DeviceModel, time.Now().In(time.UTC),
	).Scan(&lastInsID)

	p.ID = lastInsID

	return p, err
}

