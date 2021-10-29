package model

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

// ChatGroupEnt ...
type ChatGroupEnt struct {
	ID                       int32
	Member                   MemberEnt
	MemberItin               MemberItinEnt
	User                     UserEnt
	ChatGroupType            string
	ChatGroupCode            string
	Name                     string
	CreatedDate              time.Time
	ChatGroupRelation        []string
	ChatGroupTotal           int32
	ChatGroupLastMessage     string
	ChatGroupUnreadTotal     int32
	Order                    OrderEnt
	TCReplacementDescription string
	ChatMessagesEnt          []ChatMessagesEnt
	TotalMember              int32
	Status                   bool `db:"status"`
	OrderHistory             []OrderEnt
	ListUser                 []ChatMessagesEnt
	TcAssigned               bool
	ChatGroupLastMessageDate time.Time
	ChatGroupStatus          string
}

// ChatMessagesEnt ...
type ChatMessagesEnt struct {
	ID          int32
	ChatGroupID int32
	UserID      int32
	UserCode    string
	Name        string
	Message     string
	Role        string
	Image       sql.NullString
	Email       string
	IsRead      bool
	CreatedDate time.Time
}

// CreateChatGroup ...
func (c *Contract) CreateChatGroup(tx pgx.Tx, ctx context.Context, cg ChatGroupEnt) (ChatGroupEnt, error) {

	sql := "insert into chat_groups (created_by, member_itin_id, tc_id, chat_group_code, name, status, created_date,chat_group_type) values($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id"

	var lastInsID int32

	err := tx.QueryRow(context.Background(), sql,
		cg.Member.ID, cg.MemberItin.ID, cg.User.ID, cg.ChatGroupCode, cg.Name, cg.Status, time.Now().In(time.UTC), cg.ChatGroupType,
	).Scan(&lastInsID)

	cg.ID = lastInsID

	return cg, err
}

// UpdateItinMemberToChat update itin member id
func (c *Contract) UpdateItinMemberToChat(ctx context.Context, tx pgx.Tx, miID, tcID int32, code string) error {
	var ID int32

	sql := `UPDATE chat_groups SET member_itin_id=$1, tc_id=$2 WHERE chat_group_code=$3 RETURNING id`

	err := tx.QueryRow(ctx, sql, miID, tcID, code).Scan(&ID)

	return err
}

// InviteTcToChatGroup ...
func (c *Contract) InviteTcToChatGroup(ctx context.Context, tx pgx.Tx, tcID int32, code string) error {
	var ID int32

	sql := `UPDATE chat_groups SET tc_id=$1 WHERE chat_group_code=$2  RETURNING id`

	err := tx.QueryRow(ctx, sql, tcID, code).Scan(&ID)

	return err
}

// CreateChatMessage ...
func (c *Contract) CreateChatMessage(tx pgx.Tx, ctx context.Context, cm ChatMessagesEnt) (ChatMessagesEnt, error) {

	sql := "insert into chat_messages (chat_group_id, user_id, role, messages, is_read, created_date) values($1,$2,$3,$4,$5,$6) RETURNING id"

	var lastInsID int32

	err := tx.QueryRow(context.Background(), sql,
		cm.ChatGroupID, cm.UserID, cm.Role, cm.Message, cm.IsRead, time.Now().In(time.UTC),
	).Scan(&lastInsID)

	cm.ID = lastInsID
	cm.CreatedDate = time.Now().In(time.UTC)

	return cm, err
}

// GetIDGroupChatByCode ...
func (c *Contract) GetIDGroupChatByCode(db *pgxpool.Conn, ctx context.Context, code string) (int32, error) {
	var id int32
	sql := `
		select id FROM chat_groups WHERE chat_group_code = $1`

	err := db.QueryRow(ctx, sql, code).Scan(&id)
	return id, err
}

// AddChatMemberTempBatch add member temporary to chat
func (c *Contract) AddChatMemberTempBatch(ctx context.Context, tx pgx.Tx, arrStr string) error {
	var lastInsID int32

	sql := `INSERT INTO chat_member_temporaries(email, chat_group_id, created_date) VALUES ` + arrStr + ` RETURNING id`
	err := tx.QueryRow(ctx, sql).Scan(&lastInsID)

	return err
}

// GetChatList ...
func (c *Contract) GetChatListBackupV1(db *pgxpool.Conn, ctx context.Context, param map[string]interface{}) ([]ChatGroupEnt, error) {
	list := []ChatGroupEnt{}
	var where []string
	var paramQuery []interface{}
	var lastMessage, tcCode, tcName, itinDestination, itinTitle, itinCode, orderCode, orderStatus, orderType, orderStatusDesc, tcReplacementDescription, memberName, memberCode, memberEmail sql.NullString
	var itinDuration sql.NullInt32
	var statusSession sql.NullBool

	query := `select
		cg.chat_group_code,
		cg.name,
		cg.chat_group_type,
		cg.status,
		gc.chat_group_total,
		case
			when gcml_unread.last_message is null or gcml_unread.last_message = '' then gcml.last_message
			when gcml.last_message is null or gcml.last_message = '' then ''
			else gcml_unread.last_message
		end chat_group_last_message,
		case
			when gcm.chat_unread_total is null then 1
			else gcm.chat_unread_total
		end chat_group_unread_total,
		mc.member_code,
		mc.name member_name,
		mc.email member_email,
		mc.member_img,
		u.user_code tc_code,
		u.name tc_name,
		mi.destination,
		mi.title,
		mi.itin_code,
		extract(day from mi.end_date::timestamp - mi.start_date::timestamp) trip_day_duration,
		o.order_code,
		o.order_status,
		o.order_type,
		case
			when o.order_status = 'P' then 'Waiting For Payment'
			when o.order_status = 'C' then 'Completed'
			when o.order_status = 'X' then 'Cancel'
			else o.order_status
		end order_status_description,
		case
			when gcmic.role = 'tc' then 'New Request From Tc Replacement'
			when gcmic.role = 'admin' then 'New Request From System Assignment'
			else gcmic.role
		end tc_replacement_description,
		cg.created_date
	from (
		select
			cg.chat_group_code,
			m.member_code,
			m.name,
			m.username,
			m.email,
			m.img member_img
		from members m
		join chat_groups cg on cg.created_by = m.id
		union
		select
			cg.chat_group_code,
			m.member_code,
			m.name,
			m.username,
			m.email,
			m.img member_img
		from chat_group_relations cgr
		join chat_groups cg on cg.id = cgr.chat_group_id
		join members m on m.id = cgr.member_id
		where cgr.deleted_date is null
		union
		select
			cg.chat_group_code,
			m.member_code,
			m.name,
			m.username,
			cmt.email,
			m.img member_img
		from chat_member_temporaries cmt
		left join chat_groups cg on cg.id = cmt.chat_group_id
		left join members m on m.email = cmt.email
	) mc
	join chat_groups cg on cg.chat_group_code = mc.chat_group_code
	left join users u on u.id = cg.tc_id and u.deleted_date is null
	left join member_itins mi on mi.id = cg.member_itin_id and mi.deleted_date is null
	left join orders o on (cg.id = o.chat_id)
	left outer join orders o2 on (
		cg.id = o2.chat_id
		and (
			o.created_date < o2.created_date or (
				o.created_date = o2.created_date and o.id < o2.id
			)
		)
	)
	left join (
		select
			groups_chat_messages.chat_group_code,
			count(groups_chat_messages.chat_group_code) chat_unread_total
		from (
			select
				cg.chat_group_code,
				case when (cm.messages is null) then '' else cm.messages end messages,
				case when (cm.is_read is null or cm.is_read = false) then false else true end is_read
			from chat_groups cg
			left join chat_messages cm on cm.chat_group_id = cg.id
			where cm.is_read = false
		) groups_chat_messages
		group by groups_chat_messages.chat_group_code
	) gcm on gcm.chat_group_code = cg.chat_group_code
	left join (
		select
			distinct on(cg.id)
			cg.chat_group_code,
			case when (cm.messages is null) then '' else cm.messages end last_message
		from chat_groups cg
		left join chat_messages cm on cm.chat_group_id = cg.id
		order by cg.id, cm.created_date desc
	) gcml on gcml.chat_group_code = cg.chat_group_code
	left join (
		select
			distinct on(cg.id)
			cg.chat_group_code,
			case when (cm.messages is null) then '' else cm.messages end last_message
		from chat_groups cg
		left join chat_messages cm on cm.chat_group_id = cg.id
		where cm.is_read = false
		order by cg.id, cm.created_date desc
	) gcml_unread on gcml_unread.chat_group_code = cg.chat_group_code
	left join (
		select
			groups_chat.chat_group_code,
			count(groups_chat.member_code) chat_group_total
		from (
			select
				cg.chat_group_code,
				m.member_code,
				m.name,
				m.username,
				m.email,
				m.img member_img
			from members m
			join chat_groups cg on cg.created_by = m.id
			union
			select
				cg.chat_group_code,
				m.member_code,
				m.name,
				m.username,
				m.email,
				m.img member_img
			from chat_group_relations cgr
			join chat_groups cg on cg.id = cgr.chat_group_id
			join members m on m.id = cgr.member_id
			where cgr.deleted_date is null
			union
			select
				cg.chat_group_code,
				m.member_code,
				m.name,
				m.username,
				cmt.email,
				m.img member_img
			from chat_member_temporaries cmt
			left join chat_groups cg on cg.id = cmt.chat_group_id
			left join members m on m.email = cmt.email
		) groups_chat
		group by groups_chat.chat_group_code
	) gc on gc.chat_group_code = cg.chat_group_code
	left join (
		select
			distinct on(cg.id)
			cg.chat_group_code,
			u.role
		from chat_groups cg
		left join member_itin_changes mic on mic.member_itin_id = cg.member_itin_id
		left join users u on u.id = mic.changed_user_id
		order by cg.id, mic.created_date desc
	) gcmic on gcmic.chat_group_code = cg.chat_group_code `

	var queryString string = query

	where = append(where, "o2.id is null")

	if len(param["member_code"].(string)) > 0 {
		var orWhere []string
		orWhere = append(orWhere, "mc.member_code = '"+param["member_code"].(string)+"'")

		where = append(where, strings.Join(orWhere, " AND "))
	}

	if len(param["tc_code"].(string)) > 0 {
		var orWhere []string
		orWhere = append(orWhere, "u.user_code = '"+param["tc_code"].(string)+"'")

		where = append(where, strings.Join(orWhere, " AND "))
	}

	if len(param["keyword"].(string)) > 0 {
		var orWhere []string
		orWhere = append(orWhere, "lower(cg.name) like lower('%"+param["keyword"].(string)+"%')")
		orWhere = append(orWhere, "lower(gcml.last_message) like lower('%"+param["keyword"].(string)+"%')")
		orWhere = append(orWhere, "lower(mc.name) like lower('%"+param["keyword"].(string)+"%')")
		orWhere = append(orWhere, "lower(mi.destination) like lower('%"+param["keyword"].(string)+"%')")
		orWhere = append(orWhere, "lower(mi.title) like lower('%"+param["keyword"].(string)+"%')")
		orWhere = append(orWhere, "lower(mi.itin_code) like lower('%"+param["keyword"].(string)+"%')")

		where = append(where, "("+strings.Join(orWhere, " OR ")+")")
	}

	if len(where) > 0 {
		queryString += "WHERE " + strings.Join(where, " AND ")
	}

	{
		var count int
		newQcount := `SELECT COUNT(*) FROM ( ` + queryString + ` ) AS data`

		err := db.QueryRow(ctx, newQcount, paramQuery...).Scan(&count)
		if err != nil {
			return list, err
		}
		param["count"] = count
	}

	if param["is_paging"].(bool) {
		// Select Max Page
		if param["count"].(int) > param["limit"].(int) && param["page"].(int) > int(param["count"].(int)/param["limit"].(int)) {
			param["page"] = int(math.Ceil(float64(param["count"].(int)) / float64(param["limit"].(int))))
		}
		param["offset"] = (param["page"].(int) - 1) * param["limit"].(int)
	}

	if param["limit"].(int) == -1 {
		queryString += " ORDER BY " + param["order"].(string) + " " + param["sort"].(string)
	} else {
		queryString += " ORDER BY " + param["order"].(string) + " " + param["sort"].(string) + " offset $1 limit $2 "
		paramQuery = append(paramQuery, param["offset"])
		paramQuery = append(paramQuery, param["limit"])
	}

	rows, err := db.Query(ctx, queryString, paramQuery...)
	if err != nil {
		return list, err
	}
	defer rows.Close()

	for rows.Next() {
		var c ChatGroupEnt
		err = rows.Scan(&c.ChatGroupCode, &c.Name, &c.ChatGroupType, &statusSession, &c.ChatGroupTotal, &lastMessage, &c.ChatGroupUnreadTotal, &memberCode, &memberName, &memberEmail, &c.Member.Img, &tcCode, &tcName, &itinDestination, &itinTitle, &itinCode, &itinDuration, &orderCode, &orderStatus, &orderType, &orderStatusDesc, &tcReplacementDescription, &c.CreatedDate)
		if err != nil {
			return list, err
		}
		// mapping null value
		c.Status = statusSession.Bool
		c.ChatGroupLastMessage = lastMessage.String
		c.Member.MemberCode = memberCode.String
		c.Member.Name = memberName.String
		c.Member.Email = memberEmail.String
		c.User.UserCode = tcCode.String
		c.User.Name = tcName.String
		c.MemberItin.Destination = itinDestination.String
		c.MemberItin.Title = itinTitle.String
		c.MemberItin.ItinCode = itinCode.String
		c.MemberItin.DayPeriod = itinDuration.Int32
		c.Order.OrderCode = orderCode.String
		c.Order.OrderStatus = orderStatus.String
		c.Order.OrderType = orderType.String
		c.Order.OrderStatusDescription = orderStatusDesc.String
		c.TCReplacementDescription = tcReplacementDescription.String

		list = append(list, c)
	}

	return list, nil
}

// GetChatHistoryByGroupCode ...
func (c *Contract) GetChatHistoryByGroupCode(db *pgxpool.Conn, ctx context.Context, code string, param map[string]interface{}) (ChatGroupEnt, error) {

	var gc ChatGroupEnt
	var messages, roles, messageID, nameUsers, chatDates, isReads, userCodes string
	var itinTitle, itinCode, orderType, orderCode sql.NullString
	var paramQuery []interface{}

	{
		var count int
		newQcount := `
				SELECT COUNT(*) FROM ( 
					SELECT 
						cg.id cg_id, cg.member_itin_id, cg.name, cg.chat_group_code, cg.chat_group_type,
						cm.messages,cm.user_id, cm.role ,
						CASE WHEN cm.role = 'customer' THEN m.name else us.name end as user_name,
						CASE WHEN cm.role = 'customer' THEN m.member_code else us.user_code end as user_code,
						cm.created_date chat_date, cm.is_read, cm.id as message_id
					FROM chat_groups cg
					left join chat_messages cm on cm.chat_group_id = cg.id 
					left join members m on m.id = cm.user_id
					left join users us on us.id = cm.user_id
					where cg.chat_group_code = $1
					order by cm.id 
				) AS data`

		err := db.QueryRow(ctx, newQcount, code).Scan(&count)
		if err != nil {
			return gc, err
		}
		param["count"] = count
	}

	var q string

	if param["limit"].(int) == -1 {
		q += "ORDER BY " + param["order"].(string) + " " + param["sort"].(string)
	} else {
		q += "ORDER BY " + param["order"].(string) + " " + param["sort"].(string) + " offset $1 limit $2 "
		paramQuery = append(paramQuery, param["offset"])
		paramQuery = append(paramQuery, param["limit"])
	}

	query := `
			select 
				cg_id, a.name, chat_group_code, chat_group_type, m.id, m.name, m.email, m.img, m.member_code,
				CASE WHEN cgr.total_member != null THEN cgr.total_member + 1 else 1  end as total_member,
				mi.title, mi.itin_code, 
				-- o.order_type, o.order_code,
				json_agg(messages),json_agg(role), json_agg(message_id), json_agg(user_name),
				json_agg(chat_date), json_agg(is_read), json_agg(user_code)
			from (
				SELECT 
					cg.id cg_id, cg.member_itin_id, cg.name, cg.chat_group_code, cg.chat_group_type, cg.created_by createdby,
					cm.messages,cm.user_id, cm.role,
					CASE WHEN cm.role = 'customer' THEN m.name else us.name end as user_name,
					CASE WHEN cm.role = 'customer' THEN m.member_code else us.user_code end as user_code,
					cm.created_date chat_date, cm.is_read, cm.id as message_id
				FROM chat_groups cg
				left join chat_messages cm on cm.chat_group_id = cg.id 
				left join members m on m.id = cm.user_id
				left join users us on us.id = cm.user_id
				where chat_group_code = '` + code + `'
				` + q + ` ) as a
			left join ( 
				select cgr.chat_group_id, count(cgr.member_id) total_member
				from chat_group_relations cgr 
				group by chat_group_id) as cgr on cgr.chat_group_id = cg_id
			left join member_itins mi on mi.id = a.member_itin_id
			-- join orders o on o.chat_id = a.cg_id
			join members m on m.id = a.createdby
			group by a.name, chat_group_code, a.cg_id, mi.id,cgr.total_member, chat_group_type, m.id `

	err := db.QueryRow(ctx, query, paramQuery...).Scan(&gc.ID, &gc.Name, &gc.ChatGroupCode, &gc.ChatGroupType,
		&gc.Member.ID, &gc.Member.Name, &gc.Member.Email, &gc.Member.Img, &gc.Member.MemberCode,
		&gc.TotalMember, &itinTitle, &itinCode,
		// &orderType, &orderCode,
		&messages, &roles, &messageID, &nameUsers, &chatDates, &isReads, &userCodes)
	if err != nil {
		fmt.Println(err)
		return gc, err
	}

	gc.MemberItin.Title = itinTitle.String
	gc.MemberItin.ItinCode = itinCode.String
	gc.Order.OrderType = orderType.String
	gc.Order.OrderCode = orderCode.String

	var message []string
	err = json.Unmarshal([]byte(messages), &message)
	if err != nil {
		return gc, err
	}

	var role []string
	err = json.Unmarshal([]byte(roles), &role)
	if err != nil {
		return gc, err
	}

	var mID []int32
	err = json.Unmarshal([]byte(messageID), &mID)
	if err != nil {
		fmt.Println(err)
		return gc, err
	}

	var nameUser []string
	err = json.Unmarshal([]byte(nameUsers), &nameUser)
	if err != nil {
		return gc, err
	}

	var chatDate []time.Time
	err = json.Unmarshal([]byte(chatDates), &chatDate)
	if err != nil {
		return gc, err
	}

	var isRead []bool
	err = json.Unmarshal([]byte(isReads), &isRead)
	if err != nil {
		return gc, err
	}

	var userCode []string
	err = json.Unmarshal([]byte(userCodes), &userCode)
	if err != nil {
		return gc, err
	}

	for i, v := range message {
		gc.ChatMessagesEnt = append(gc.ChatMessagesEnt, ChatMessagesEnt{ID: mID[i], Message: v, Role: role[i], Name: nameUser[i], CreatedDate: chatDate[i], IsRead: isRead[i], UserCode: userCode[i]})
	}

	return gc, err
}

// IsExistInGroupChat ...
func (c *Contract) IsExistInGroupChat(db *pgxpool.Conn, ctx context.Context, groupCode string, code string) (int32, error) {

	var id int32

	query := `
			select 
				count(*) 
			from (
				select 
					members.member_code, us.user_code, cg.chat_group_code 
				from chat_groups as cg
				left join chat_group_relations cgr on cgr.chat_group_id = cg.id
				left join members on members.id = cgr.member_id
				left join users us on us.id = cg.tc_id
				union
				select 
					members.member_code, us.user_code, cg.chat_group_code 
				from chat_groups as cg
				left join chat_group_relations cgr on cgr.chat_group_id = cg.id
				left join members on members.id = cg.created_by
				left join users us on us.id = cg.tc_id 
			) a
			where chat_group_code = $1 and member_code = $2 or user_code = $3
 		`
	err := db.QueryRow(ctx, query, groupCode, code, code).Scan(&id)
	if err != nil {
		return id, err
	}

	return id, nil
}

// UpdateIsRead ...
func (c *Contract) UpdateIsRead(ctx context.Context, tx pgx.Tx, cgID int32, userIDLogin int32) error {
	var ID int32
	sql := `UPDATE chat_messages SET is_read = true WHERE chat_group_id=$1 and user_id != $2  RETURNING id`

	err := tx.QueryRow(ctx, sql, cgID, userIDLogin).Scan(&ID)
	return err
}

func (c *Contract) GetGroupChatsCreatedBy(db *pgxpool.Conn, ctx context.Context, code string) (ChatGroupEnt, error) {
	var cg ChatGroupEnt

	sql := `select 
			cg.id chat_group_id,
			cg.name chat_group_name, 
			m.name member_name, 
			cg.status chat_group_status 
		from chat_groups cg
		join members m on m.id = cg.created_by and m.deleted_date is null
		where chat_group_code = $1 limit 1`

	err := db.QueryRow(ctx, sql, code).Scan(&cg.ID, &cg.Name, &cg.Member.Name, &cg.Status)

	return cg, err
}

func (c *Contract) UpdateChatGroupStatusAndTC(tx pgx.Tx, ctx context.Context, tcID int32, status bool, code string) error {
	var ID int32
	sql := `UPDATE chat_groups SET tc_id = $1, status = $2 WHERE chat_group_code = $3 RETURNING id`
	err := tx.QueryRow(ctx, sql, tcID, status, code).Scan(&ID)

	return err
}

// UpdateIsReadMessageByUserID ...
func (c *Contract) UpdateIsReadMessageByUserID(tx pgx.Tx, ctx context.Context, userID int32, cgID int32) error {

	var ID int32

	sql := `UPDATE chat_messages SET is_read= true WHERE user_id=$1 and chat_group_id = $2  RETURNING id`

	err := tx.QueryRow(ctx, sql, userID, cgID).Scan(&ID)

	return err
}

// CheckOtherMessages ...
func (c *Contract) CheckOtherMessages(db *pgxpool.Conn, ctx context.Context, gcID int32, userID int32) (int32, error) {
	var id int32
	sql := `
		select id FROM chat_messages WHERE chat_group_id = $1 and user_id != $2`

	err := db.QueryRow(ctx, sql, gcID, userID).Scan(&id)
	return id, err
}

// GetListUserGroupChat ...
func (c *Contract) GetListUserGroupChat(db *pgxpool.Conn, ctx context.Context, gcID string) ([]ChatMessagesEnt, error) {

	list := []ChatMessagesEnt{}
	var name, email, code sql.NullString

	sql := `
		select 
			m.name, m.email, m.member_code as code, m.img, 'customer' as role
		from chat_group_relations as cgr
		left join members m on m.id = cgr.member_id
		where cgr.chat_group_id = ` + gcID + `
		union
		select 
			m.name, m.email, m.member_code as code, m.img, 'customer' as role 
		from chat_groups as cg
		left join members m on m.id = cg.created_by
		where cg.id = ` + gcID + `
		union
		select 
			us."name", us.email, us.user_code as code, us.img , 'tc' as role 
		from chat_groups as cg
		left join users us on us.id = cg.tc_id
		where cg.id = ` + gcID + `
		union
		select 
			m.name, cmt.email, m.member_code as code, m.img,  'customer' as role 
		from  chat_member_temporaries as cmt 
		left join members m on m.email = cmt.email
		where cmt.chat_group_id = ` + gcID + `
 `
	rows, err := db.Query(ctx, sql)
	if err != nil {
		return list, err
	}

	defer rows.Close()
	for rows.Next() {
		var a ChatMessagesEnt
		err = rows.Scan(&name, &email, &code, &a.Image, &a.Role)
		if err != nil {
			return list, err
		}
		a.Name = name.String
		a.Email = email.String
		a.UserCode = code.String

		list = append(list, a)
	}

	return list, err
}

// GetGroupChatByCode ...
func (c *Contract) GetGroupChatByCode(db *pgxpool.Conn, ctx context.Context, code string) (ChatGroupEnt, error) {
	var ch ChatGroupEnt
	sql := `select id, created_by, member_itin_id, tc_id, chat_group_code, name, created_date, chat_group_type, status FROM chat_groups WHERE chat_group_code = $1`

	err := db.QueryRow(ctx, sql, code).Scan(&ch.ID, &ch.Member.ID, &ch.MemberItin.ID, &ch.User.ID, &ch.ChatGroupCode, &ch.Name, &ch.CreatedDate, &ch.ChatGroupType, &ch.Status)
	return ch, err
}

// GetChatList ...
func (c *Contract) GetChatList(db *pgxpool.Conn, ctx context.Context, param map[string]interface{}) ([]ChatGroupEnt, error) {
	list := []ChatGroupEnt{}
	var where []string
	var paramQuery []interface{}
	var lastMessage, tcCode, tcName, memberName, memberCode, memberEmail sql.NullString
	var statusSession sql.NullBool
	var messagesDate sql.NullTime
	var addedFriend string

	if len(param["member_code"].(string)) > 0 {
		addedFriend = `
		union
			select
				cg.chat_group_code,
				m.member_code, 
				m.name, 
				m.username, 
				m.email,
				m.img member_img
			from chat_group_relations cgr
			join chat_groups cg on cg.id = cgr.chat_group_id 
			join members m on m.id = cgr.member_id 
			where cgr.deleted_date is null
			union
			select
				cg.chat_group_code,
				m.member_code, 
				m.name, 
				m.username, 
				cmt.email,
				m.img member_img
			from chat_member_temporaries cmt
			left join chat_groups cg on cg.id = cmt.chat_group_id 
			left join members m on m.email = cmt.email`
	}

	query := `
	select 
		cg.chat_group_code,
		cg.name,
		cg.chat_group_type,
		cg.status,
		gc.chat_group_total,
		case 
			when gcml_unread.last_message is null or gcml_unread.last_message = '' then gcml.last_message
			when gcml.last_message is null or gcml.last_message = '' then ''
			else gcml_unread.last_message 
		end chat_group_last_message,
		case 
			when gcm.chat_unread_total is null then 1 
			else gcm.chat_unread_total 
		end chat_group_unread_total, 
		gcml.tc_assigned,
		gcml.messages_date,
		mc.member_code,
		mc.name member_name,
		mc.email member_email,
		mc.member_img,
		gcml.tc_code,
		gcml.tc_name,
		cg.created_date
	from (
		select
			cg.chat_group_code,
			m.member_code, 
			m.name, 
			m.username, 
			m.email,
			m.img member_img
		from members m
		join chat_groups cg on cg.created_by = m.id
		` + addedFriend + `
	) mc
	join chat_groups cg on cg.chat_group_code = mc.chat_group_code
	left join users u on u.id = cg.tc_id and u.deleted_date is null
	left join (
		select
			groups_chat_messages.chat_group_code,
			count(groups_chat_messages.chat_group_code) chat_unread_total
		from (
			select 
				cg.chat_group_code,
				case when (cm.messages is null) then '' else cm.messages end messages, 
				case when (cm.is_read is null or cm.is_read = false) then false else true end is_read 
			from chat_groups cg
			left join chat_messages cm on cm.chat_group_id = cg.id
			where cm.is_read = false
		) groups_chat_messages
		group by groups_chat_messages.chat_group_code
	) gcm on gcm.chat_group_code = cg.chat_group_code
	left join (
		select 
			distinct on(cg.id)
			cg.chat_group_code,
			case when (cm.messages is null) then '' else cm.messages end last_message,
			case when (a.role = 'tc') then true else false end tc_assigned,
			case when (a.role = 'tc') then a.name else null end tc_name,
			case when (a.role = 'tc') then a.user_code else null end tc_code,
			cm.created_date messages_date
		from chat_groups cg
		left join chat_messages cm on cm.chat_group_id = cg.id
		left join (
		select 
			cg.chat_group_code, cm.role, user_id, us.user_code, us.name
		from chat_groups cg
		left join chat_messages cm on cm.chat_group_id = cg.id 
		join users us on us.id = cm.user_id
		where cm.role = 'tc') as a on a.chat_group_code = cg.chat_group_code
		order by cg.id, cm.created_date desc
	) gcml on gcml.chat_group_code = cg.chat_group_code
	left join (
		select 
			distinct on(cg.id)
			cg.chat_group_code,
			case when (cm.messages is null) then '' else cm.messages end last_message
		from chat_groups cg
		left join chat_messages cm on cm.chat_group_id = cg.id
		where cm.is_read = false
		order by cg.id, cm.created_date desc
	) gcml_unread on gcml_unread.chat_group_code = cg.chat_group_code
	left join (
		select 
			groups_chat.chat_group_code,
			count(groups_chat.member_code) chat_group_total
		from (
			select
				cg.chat_group_code,
				m.member_code, 
				m.name, 
				m.username, 
				m.email,
				m.img member_img
			from members m
			join chat_groups cg on cg.created_by = m.id
			union
			select
				cg.chat_group_code,
				m.member_code, 
				m.name, 
				m.username, 
				m.email,
				m.img member_img
			from chat_group_relations cgr
			join chat_groups cg on cg.id = cgr.chat_group_id 
			join members m on m.id = cgr.member_id 
			where cgr.deleted_date is null
			union
			select
				cg.chat_group_code,
				m.member_code, 
				m.name, 
				m.username, 
				cmt.email,
				m.img member_img
			from chat_member_temporaries cmt
			left join chat_groups cg on cg.id = cmt.chat_group_id 
			left join members m on m.email = cmt.email 
		) groups_chat
		group by groups_chat.chat_group_code
	) gc on gc.chat_group_code = cg.chat_group_code
	
	`

	var queryString string = query

	if len(param["member_code"].(string)) > 0 {
		var orWhere []string
		orWhere = append(orWhere, "mc.member_code = '"+param["member_code"].(string)+"'")

		where = append(where, strings.Join(orWhere, " AND "))
	}

	if len(param["tc_code"].(string)) > 0 {
		var orWhere []string
		orWhere = append(orWhere, "u.user_code = '"+param["tc_code"].(string)+"'")

		where = append(where, strings.Join(orWhere, " AND "))
	}

	if len(param["chat_status"].(string)) > 0 {

		if param["chat_status"].(string) == "new_chat" {
			var orWhere []string
			orWhere = append(orWhere, "gcml.tc_assigned = 'false'")
			where = append(where, strings.Join(orWhere, " AND "))
		} else if param["chat_status"].(string) == "active_chat" {
			var orWhere []string
			orWhere = append(orWhere, "gcml.tc_assigned = 'true'")
			where = append(where, strings.Join(orWhere, " AND "))
		}

	}

	if len(param["keyword"].(string)) > 0 {
		var orWhere []string
		orWhere = append(orWhere, "lower(cg.name) like lower('%"+param["keyword"].(string)+"%')")
		orWhere = append(orWhere, "lower(gcml.last_message) like lower('%"+param["keyword"].(string)+"%')")
		orWhere = append(orWhere, "lower(mc.name) like lower('%"+param["keyword"].(string)+"%')")
		where = append(where, "("+strings.Join(orWhere, " OR ")+")")
	}

	if len(where) > 0 {
		queryString += "WHERE " + strings.Join(where, " AND ")
	}

	{
		var count int
		newQcount := `SELECT COUNT(*) FROM ( ` + queryString + ` ) AS data`

		err := db.QueryRow(ctx, newQcount, paramQuery...).Scan(&count)
		if err != nil {
			return list, err
		}
		param["count"] = count
	}

	if param["is_paging"].(bool) {
		// Select Max Page
		if param["count"].(int) > param["limit"].(int) && param["page"].(int) > int(param["count"].(int)/param["limit"].(int)) {
			param["page"] = int(math.Ceil(float64(param["count"].(int)) / float64(param["limit"].(int))))
		}
		param["offset"] = (param["page"].(int) - 1) * param["limit"].(int)
	}

	if param["limit"].(int) == -1 {
		queryString += " ORDER BY " + param["order"].(string) + " " + param["sort"].(string)
	} else {
		queryString += " ORDER BY " + param["order"].(string) + " " + param["sort"].(string) + " offset $1 limit $2 "
		paramQuery = append(paramQuery, param["offset"])
		paramQuery = append(paramQuery, param["limit"])
	}
	rows, err := db.Query(ctx, queryString, paramQuery...)
	if err != nil {
		return list, err
	}
	defer rows.Close()

	for rows.Next() {
		var c ChatGroupEnt

		err = rows.Scan(&c.ChatGroupCode, &c.Name, &c.ChatGroupType, &statusSession, &c.ChatGroupTotal, &lastMessage, &c.ChatGroupUnreadTotal, &c.TcAssigned, &messagesDate, &memberCode, &memberName, &memberEmail, &c.Member.Img, &tcCode, &tcName, &c.CreatedDate)
		if err != nil {
			return list, err
		}
		// mapping null value
		c.Status = statusSession.Bool
		c.ChatGroupLastMessage = lastMessage.String
		c.Member.MemberCode = memberCode.String
		c.Member.Name = memberName.String
		c.Member.Email = memberEmail.String
		c.User.UserCode = tcCode.String
		c.User.Name = tcName.String
		c.ChatGroupLastMessageDate = messagesDate.Time

		if len(param["chat_status"].(string)) > 0 {

			if param["chat_status"].(string) == "new_chat" {
				c.ChatGroupStatus = "new_chat"
			} else if param["chat_status"].(string) == "active_chat" {
				c.ChatGroupStatus = "active_chat"
			}

		}

		list = append(list, c)
	}

	return list, nil
}
