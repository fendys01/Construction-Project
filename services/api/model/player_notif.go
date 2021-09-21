package model

import (
	"context"
	"fmt"
	"panorama/lib/onesignal"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

// DeviceListEnt.
type DeviceListEnt struct {
	ID          int32
	PlayerID    string
	UserID      int64
	Role        string
	DeviceModel string
	DeviceType  int32
	CreatedDate time.Time
}

func (c *Contract) AddDevice(tx pgx.Tx, ctx context.Context, p DeviceListEnt) (DeviceListEnt, error) {
	var lastInsID int32

	if len(p.PlayerID) <= 0 {
		playerID, err := onesignal.New(c.App).AddDevice(int(p.DeviceType))
		if err != nil {
			return p, err
		}
		p.PlayerID = playerID
	}

	sql := `insert into device_list (player_id, user_id, role, device_type, device_model, created_date) values($1, $2, $3, $4, $5, $6) RETURNING id;`
	err := tx.QueryRow(ctx, sql, p.PlayerID, p.UserID, p.Role, p.DeviceType, p.DeviceModel, time.Now().In(time.UTC)).Scan(&lastInsID)

	p.ID = lastInsID

	return p, err
}

// GetDeviceNotAppByUserIDAndRole get device not app mobile by user id and role
func (c *Contract) GetDeviceNotAppByUserIDAndRole(db *pgxpool.Conn, ctx context.Context, userID int32, role string) (DeviceListEnt, error) {
	var p DeviceListEnt

	sql := `select id, player_id, created_date, user_id, role, device_type, device_model 
			from device_list
			where user_id = $1 and role = $2 and device_type not in ($3, $4) `

	err := db.QueryRow(ctx, sql, userID, role, onesignal.DEVICE_TYPE_IOS, onesignal.DEVICE_TYPE_ANDROID).Scan(&p.ID, &p.PlayerID, &p.CreatedDate, &p.UserID, &p.Role, &p.DeviceType, &p.DeviceModel)

	return p, err
}

// GetDeviceByPlayerIDAndRole get device by player id and role
func (c *Contract) GetDeviceByPlayerIDAndRole(db *pgxpool.Conn, ctx context.Context, playerID, role string) (DeviceListEnt, error) {
	var p DeviceListEnt

	sql := `select id, player_id, created_date, user_id, role, device_type, device_model from device_list where player_id = $1 and role = $2 `

	err := db.QueryRow(ctx, sql, playerID, role).Scan(&p.ID, &p.PlayerID, &p.CreatedDate, &p.UserID, &p.Role, &p.DeviceType, &p.DeviceModel)

	return p, err
}

func (c *Contract) UpdatePlayer(tx pgx.Tx, ctx context.Context, p DeviceListEnt, deviceID int32) (DeviceListEnt, error) {
	var ID int32
	sql := `update device_list set player_id=$1, user_id=$2, role=$3, device_type=$4, device_model=$5 where id=$6 RETURNING id`
	err := tx.QueryRow(ctx, sql, p.PlayerID, p.UserID, p.Role, p.DeviceType, p.DeviceModel, deviceID).Scan(&ID)

	p.ID = ID

	return p, err
}

// GetListPlayerByUserCodeAndRole ...
func (c *Contract) GetListPlayerByUserCodeAndRole(db *pgxpool.Conn, ctx context.Context, code, role string) ([]DeviceListEnt, error) {
	var listPlayer []DeviceListEnt
	var paramQuery []interface{}

	// Default only user cms
	whereMember := `where dl.role != 'customer'`
	whereUser := `where dl.role != 'customer'`

	if len(code) > 0 && len(role) > 0 {
		whereMember = `where m.member_code = $1 and dl.role = $3`
		whereUser = `where us.user_code = $2 and dl.role = $4`
		paramQuery = append(paramQuery, code, code, role, role)
	} else if len(code) > 0 {
		whereMember = `where m.member_code = $1`
		whereUser = `where us.user_code = $2`
		paramQuery = append(paramQuery, code, code)
	} else if len(role) > 0 {
		whereMember = `where dl.role = $1`
		whereUser = `where dl.role = $2`
		paramQuery = append(paramQuery, role, role)
	}

	query := fmt.Sprintf(`
		select dl.player_id, dl.user_id, dl.role from device_list as dl join members m on m.id = dl.user_id %s union
		select dl.player_id, dl.user_id, dl.role from device_list as dl join users us on us.id = dl.user_id %s
	`, whereMember, whereUser)

	rows, err := db.Query(ctx, query, paramQuery...)
	if err != nil {
		return listPlayer, err
	}

	defer rows.Close()
	for rows.Next() {
		var p DeviceListEnt
		err = rows.Scan(&p.PlayerID, &p.UserID, &p.Role)
		if err != nil {
			return listPlayer, err
		}

		listPlayer = append(listPlayer, p)
	}

	return listPlayer, err
}

func (c *Contract) AddPlayer(tx pgx.Tx, db *pgxpool.Conn, ctx context.Context, userID int64, userRole, xPlayerID, ch string) (DeviceListEnt, error) {
	var err error
	var device DeviceListEnt

	// Set player formatter
	formatter := DeviceListEnt{
		UserID: int64(userID),
		Role:   userRole,
	}

	// Check device exist & Adjust device each channel & update player id
	if len(xPlayerID) > 0 {
		// Check if x-player header assigned all channel
		formatter.PlayerID = xPlayerID
		device, _ = c.GetDeviceByPlayerIDAndRole(db, ctx, xPlayerID, userRole)
		if device.ID != 0 && device.UserID != int64(userID) {
			device, err = c.UpdatePlayer(tx, ctx, formatter, device.ID)
			if err != nil {
				return device, err
			}
		}
	} else {
		// Check by user id role but not app channel
		device, _ = c.GetDeviceNotAppByUserIDAndRole(db, ctx, int32(userID), userRole)
	}
	if device.ID == 0 {
		// Set default device type chrome if cms browser
		if ch == ChannelCMS {
			formatter.DeviceType = onesignal.DEVICE_TYPE_BROWSER_CHROMEWEB
		}
		device, err = c.AddDevice(tx, ctx, formatter)
		if err != nil {
			return formatter, err
		}
	}
	xPlayerID = device.PlayerID

	// Check device by player id onesignal
	playerDevice, err := onesignal.New(c.App).GetPlayerDevice(xPlayerID)
	if err != nil {
		return device, err
	}
	formatter.PlayerID = playerDevice.ID
	formatter.DeviceType = int32(playerDevice.DeviceType)
	formatter.DeviceModel = playerDevice.DeviceModel

	// Update device data credential if device null
	if len(device.DeviceModel) <= 0 {
		deviceUpdated, err := c.UpdatePlayer(tx, ctx, formatter, device.ID)
		if err != nil {
			return device, err
		}
		deviceUpdated.ID = device.ID
		deviceUpdated.CreatedDate = device.CreatedDate
		device = deviceUpdated
	}

	return device, err
}
