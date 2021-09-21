package model

import (
	"context"
	"database/sql"
	"fmt"
	"time"

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

	sql := `select last_active_date, SUM(total_visited) as total_visits FROM log_visit_app	
	where role = 'customer' and last_active_date between $1 and $2 group by last_active_date order by last_active_date asc`

	startDate := fmt.Sprintf("%v", time.Now().Format("2006-01-02")) + " 00:00:00"
	endDate := fmt.Sprintf("%v", time.Now().Format("2006-01-02")) + " 23:59:59"

	if len(param["start_date"].(string)) > 0 && len(param["end_date"].(string)) > 0 {
		startDate = fmt.Sprintf("%v %s", param["start_date"], "00:00:00")
		endDate = fmt.Sprintf("%v %s", param["end_date"], "23:59:59")
	}

	paramQuery = append(paramQuery, startDate)
	paramQuery = append(paramQuery, endDate)

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
