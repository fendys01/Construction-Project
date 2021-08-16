ckage model

import (
	"context"
	"database/sql"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

type MemberItinEnt struct {
	ID          int32
	Code        string
	Title       string
	CreatedBy   int32
	EstPrice    sql.NullInt64
	StartDate   time.Time
	EndDate     time.Time
	Details     []map[string]interface{}
	CreatedDate time.Time
	UpdatedDate time.Time
	DeletedDate time.Time
}

// AddMemberItin add new itinerary by members
func (c *Contract) AddMemberItin(db *pgxpool.Conn, ctx context.Context, m MemberItinEnt) (int32, error) {
	var (
		err       error
		lastInsID int32
	)

	sql := `insert into member_itins(itin_code, title, created_by, start_date, end_date, details, created_at) 
		values($1, $2, $3, $4, $5, $6, $7) RETURNING id`
	err = db.QueryRow(context.Background(), sql,
		m.Code, m.Title, m.CreatedBy, m.StartDate, m.EndDate, m.Details, m.CreatedDate,
	).Scan(&lastInsID)

	return lastInsID, err
}
