package model

import (
	"context"
	"database/sql"

	"github.com/jackc/pgx/v4/pgxpool"
)

type DailyVisitsEnt struct {
	LastActiveDate sql.NullTime
	TotalVisited   sql.NullInt32
}

// Get Daily Visits
func (c *Contract) GetDailyVisitsAct(db *pgxpool.Conn, ctx context.Context, param map[string]interface{}) ([]DailyVisitsEnt, error) {
	list := []DailyVisitsEnt{}
	var paramQuery []interface{}

	sql := `select
				last_active_date, SUM(total_visited) as total_visits
			FROM log_visit_app	
			where role = 'customer'
			group by last_active_date`

	rows, err := db.Query(ctx, sql, paramQuery...)
	if err != nil {
		return list, err
	}

	defer rows.Close()

	for rows.Next() {
		var d DailyVisitsEnt
		err = rows.Scan(&d.LastActiveDate, &d.TotalVisited)
		if err != nil {
			return list, err
		}

		list = append(list, d)
	}
	return list, err
}
