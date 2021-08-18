package model

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"panorama/lib/utils"
	"strings"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

// NotificationEnt ...
type NotificationEnt struct {
	ID                int32
	Code              string
	MemberCode        string
	Type              int32
	Title             string
	Content           string
	Link              string
	IsRead            int
	CreatedDate       time.Time
	Member            MemberEnt
	MemberItin        MemberItinEnt
	User              UserEnt
	MemberItinChanges MemberItinChangesEnt
}

func (c *Contract) SetNotificationCode() string {
	rand.Seed(time.Now().UnixNano())
	code, _ := utils.Generate(`[a-z0-9]{6}`)
	return fmt.Sprintf("NO-%s-%s", time.Now().In(time.Local).Format("060102"), code)
}

func (c *Contract) AddNotif(db *pgxpool.Conn, ctx context.Context, n NotificationEnt) (NotificationEnt, error) {
	var lastInsID int32
	n.Code = c.SetNotificationCode()
	err := db.QueryRow(ctx, `insert into notifications (code, member_code, type, title, content, link, is_read, created_date) 
		values($1, $2, $3, $4, $5, $6, $7,$8) RETURNING id`,
		n.Code, n.MemberCode, n.Type, n.Title, n.Content, n.Link, n.IsRead, time.Now().In(time.UTC),
	).Scan(&lastInsID)

	n.ID = lastInsID

	return n, err
}

// GetListNotification ...
func (c *Contract) GetListNotification(db *pgxpool.Conn, ctx context.Context, param map[string]interface{}) ([]NotificationEnt, error) {
	listNotification := []NotificationEnt{}
	var where []string
	var paramQuery []interface{}
	var itinCode, itinTitle, tcCode, tcName, tcCodeChanged, tcNameChanged sql.NullString
	var itinStartDate, itinEndDate sql.NullTime

	query := `select distinct on(n.created_date)
		n.code,
		m.member_code,
		m.name member_name,
		n.type,
		n.title,
		n.content,
		n.link,
		n.is_read,
		n.created_date,
		mi.itin_code,
		mi.title itin_title,
		mi.start_date,
		mi.end_date,
		mi.details,
		u.user_code tc_code,
		u.name tc_name,
		max_member_itin_changes.user_code tc_changed_code,
		max_member_itin_changes.name tc_changed_name
	from notifications n 
	join members m on m.member_code = n.member_code  and m.is_active = 'true'
	left join member_itins mi on mi.created_by = m.id and mi.deleted_date is null
	left join orders o on o.member_itin_id = mi.id 
	left join users u on u.id = o.tc_id and u.is_active = 'true'
	left join member_itin_relations mir on mir.member_itin_id = mi.id and mir.deleted_date is null
	left join (
		select mic.member_itin_id, u2.user_code, u2.name, MAX(mic.created_date) max_date
		from member_itin_changes mic
		join users u2 on u2.id = mic.changed_user_id
		group by mic.member_itin_id, u2.user_code, u2.name
	) max_member_itin_changes on mi.id = max_member_itin_changes.member_itin_id`

	var queryString string = query

	if len(param["member_code"].(string)) > 0 {
		var orWhere []string
		orWhere = append(orWhere, "n.member_code = '"+param["member_code"].(string)+"'")

		where = append(where, strings.Join(orWhere, " AND "))
	}

	if len(param["keyword"].(string)) > 0 {
		var orWhere []string
		orWhere = append(orWhere, "lower(n.title) like lower('%"+param["keyword"].(string)+"%')")
		orWhere = append(orWhere, "lower(n.content) like lower('%"+param["keyword"].(string)+"%')")

		where = append(where, "("+strings.Join(orWhere, " OR ")+")")
	}

	if len(param["created_by"].(string)) > 0 {
		var orWhere []string
		orWhere = append(orWhere, "mi.created_by = "+param["created_by"].(string))

		where = append(where, "("+strings.Join(orWhere, " AND ")+")")
	}

	if len(where) > 0 {
		queryString += " WHERE " + strings.Join(where, " AND ")
	}

	{
		var count int
		newQcount := `SELECT COUNT(*) FROM ( ` + queryString + ` ) AS data`
		err := db.QueryRow(ctx, newQcount, paramQuery...).Scan(&count)
		if err != nil {
			return listNotification, err
		}
		param["count"] = count
	}

	// Select Max Page
	if param["count"].(int) > param["limit"].(int) && param["page"].(int) > int(param["count"].(int)/param["limit"].(int)) {
		param["page"] = int(math.Ceil(float64(param["count"].(int)) / float64(param["limit"].(int))))
	}

	param["offset"] = (param["page"].(int) - 1) * param["limit"].(int)

	if param["limit"].(int) == -1 {
		queryString += " ORDER BY " + param["order"].(string) + " " + param["sort"].(string)
	} else {
		queryString += " ORDER BY " + param["order"].(string) + " " + param["sort"].(string) + " offset $1 limit $2 "
		paramQuery = append(paramQuery, param["offset"])
		paramQuery = append(paramQuery, param["limit"])
	}

	rows, err := db.Query(ctx, queryString, paramQuery...)
	if err != nil {
		return listNotification, err
	}

	defer rows.Close()
	for rows.Next() {
		var notif NotificationEnt
		err = rows.Scan(&notif.Code, &notif.MemberCode, &notif.Member.Name, &notif.Type, &notif.Title, &notif.Content, &notif.Link, &notif.IsRead, &notif.CreatedDate, &itinCode, &itinTitle, &itinStartDate, &itinEndDate, &notif.MemberItin.Details, &tcCode, &tcName, &tcCodeChanged, &tcNameChanged)
		if err != nil {
			return listNotification, err
		}

		notif.MemberItin.ItinCode = itinCode.String
		notif.MemberItin.Title = itinTitle.String
		notif.MemberItin.StartDate = itinStartDate.Time
		notif.MemberItin.EndDate = itinEndDate.Time
		notif.User.UserCode = tcCode.String
		notif.User.Name = tcName.String
		notif.MemberItinChanges.User.UserCode = tcCodeChanged.String
		notif.MemberItinChanges.User.Name = tcNameChanged.String

		for _, v := range notif.MemberItin.Details {
			c, err := json.Marshal(v["visit_list"])
			if err != nil {
				return listNotification, err
			}
			notif.MemberItin.DayPeriod = int32(strings.Count(string(c), "]"))
		}

		listNotification = append(listNotification, notif)
	}
	return listNotification, err
}
