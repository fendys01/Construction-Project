package model

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"panorama/lib/onesignal"
	"panorama/lib/utils"
	"strings"
	"time"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

const (
	NOTIF_TYPE_CHAT                  = 1
	NOTIF_TYPE_ORDER                 = 2
	NOTIF_TYPE_MBITIN                = 3
	NOTIF_TYPE_SUGGITIN              = 4
	NOTIF_TYPE_PROFILE               = 5
	NOTIF_TYPE_ADMIN                 = 6
	NOTIF_TYPE_TC                    = 7
	NOTIF_TYPE_CUSTOMER              = 8
	NOTIF_TYPE_STUFF            	 = 9
	NOTIF_SUBJ_CHAT_INCOME           = "Chat Incoming"
	NOTIF_SUBJ_CHAT_UNREAD           = "Chat Unread"
	NOTIF_SUBJ_CHAT_ROOM_ASSIGNED    = "New Chat Room Assigned"
	NOTIF_SUBJ_ORDER_INCOME          = "Payment Incoming"
	NOTIF_SUBJ_ORDER_VERIF           = "Verified Payment"
	NOTIF_SUBJ_ORDER_CANCEL          = "Cancelled Payment"
	NOTIF_SUBJ_ORDER_FAIL            = "Failed Payment"
	NOTIF_SUBJ_ORDER_HISTORY         = "Payment History"
	NOTIF_SUBJ_ORDER_CLIENT_COMPLETE = "Client Completed Payment"
	NOTIF_SUBJ_ORDER_CLIENT_FAIL     = "Client Failed Payment"
	NOTIF_SUBJ_MBITIN_PRE            = "Pre-trip"
	NOTIF_SUBJ_MBITIN_BEGIN          = "Trip Begins"
	NOTIF_SUBJ_SUGGITIN_NEW          = "New Suggested Itinerary"
	NOTIF_SUBJ_STUFF_NEW          	 = "New Stuff Added"
	NOTIF_SUBJ_PROFILE_CHANGE        = "Change Profile Info"
	NOTIF_SUBJ_ADMIN_ADD             = "New Admin Added"
	NOTIF_SUBJ_TC_INVITED            = "TC Invited"
	NOTIF_SUBJ_TC_CHANGED            = "TC Changed"
	NOTIF_SUBJ_TC_ADD                = "New TC Added"
	NOTIF_SUBJ_TC_REMOVE             = "TC Removed"
	NOTIF_SUBJ_CUSTOMER_BANNED       = "Customer Banned"
)

// NotificationEnt ...
type NotificationEnt struct {
	ID              int32
	Code            string
	UserID          int64
	Role            string
	Subject         string
	Type            int32
	TypeText        string
	Title           string
	Content         string
	Link            sql.NullString
	IsRead          bool
	AdditionalTitle sql.NullString
	CreatedDate     time.Time
	MemberItin      MemberItinEnt
	User            UserEnt
}

type NotificationContent struct {
	Subject       string
	ChatMessage   string
	ChatRoom      string
	RoomName      string
	ClientName    string
	OrderCode     string
	PaymentMethod string
	TripName      string
	StatusPayment string
	Day           int
	Info          string
	CustomerName  string
	TCName        string
	AdminName     string
	SugItinTitle  string
	StuffName	  string
}

func (c *Contract) SetNotificationCode() string {
	rand.Seed(time.Now().UnixNano())
	code, _ := utils.Generate(`[a-z0-9]{6}`)
	return fmt.Sprintf("NO-%s-%s", time.Now().In(time.Local).Format("060102"), code)
}

func (c *Contract) AddNotif(tx pgx.Tx, ctx context.Context, n NotificationEnt) (NotificationEnt, error) {
	var lastInsID int32

	n.Code = c.SetNotificationCode()

	err := tx.QueryRow(ctx, `insert into notifications (code, type, title, content, link, is_read, subject, user_id, role, created_date) values($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id`, n.Code, n.Type, n.Title, n.Content, n.Link, false, n.Subject, n.UserID, n.Role, time.Now().In(time.UTC)).Scan(&lastInsID)

	n.ID = lastInsID

	return n, err
}

func (c *Contract) GetNotifByCode(db *pgxpool.Conn, ctx context.Context, code string) (NotificationEnt, error) {
	var n NotificationEnt

	err := pgxscan.Get(ctx, db, &n, "select * from notifications where code = $1 limit 1", code)
	return n, err
}

func (c *Contract) GetNotifAll(db *pgxpool.Conn, ctx context.Context) (NotificationEnt, error) {
	var n NotificationEnt

	err := pgxscan.Get(ctx, db, &n, "select * from notifications")
	return n, err
}

// UpdateIsReadNotificationsByCode ...
func (c *Contract) UpdateIsReadNotificationByCode(tx pgx.Tx, ctx context.Context, code string, n NotificationEnt) error {

	var ID int32

	sql := `UPDATE notifications SET is_read= true WHERE code=$1 RETURNING id`

	err := tx.QueryRow(ctx, sql, code).Scan(&ID)

	return err
}

// Force Deleted Notification.
func (c *Contract) ForceDeleteNotification(tx pgx.Tx, ctx context.Context, code string, n NotificationEnt) error {

	var ID int32

	sql := `delete from notifications where code=$1 RETURNING id`

	err := tx.QueryRow(ctx, sql, code).Scan(&ID)

	n.ID = ID

	return err
}

// Force Deleted Notification.
func (c *Contract) ForceDeleteAllNotification(tx pgx.Tx, ctx context.Context, n NotificationEnt) error {

	var ID int32

	sql := `delete from notifications RETURNING id`

	err := tx.QueryRow(ctx, sql).Scan(&ID)

	n.ID = ID

	return err
}

// GetListNotification ...
func (c *Contract) GetListNotification(db *pgxpool.Conn, ctx context.Context, param map[string]interface{}) ([]NotificationEnt, error) {
	listNotification := []NotificationEnt{}
	var where []string
	var paramQuery []interface{}

	query := `select distinct on(n.id)
		n.code notif_code,
		n.subject notif_subject,
		n.type notif_type,
		case 
			when n.type = 1 then 'Chat'
			when n.type = 2 then 'Payment'
			when n.type = 3 then 'Trip'
			when n.type = 4 then 'Itinerary'
			when n.type = 5 then 'User Activity'
			when n.type = 6 then 'Admin'
			when n.type = 7 then 'TC'
			when n.type = 8 then 'Customer'
			else null 
		end text_type,
		n.title notif_title,
		n.content notif_content,
		case 
			when n.link is null or n.link = '' then null
			else n.link
		end notif_link,
		n.is_read,
		uapp.user_code,
		uapp.user_name,
		uapp.user_email,
		uapp.user_phone,
		uapp.user_img,
		uapp.role,
		case 
			when n.type = 3 then its.title 
			when n.type = 4 then mi.title 
			when n.type = 5 then uapp.user_name
			when n.type = 6 and uapp.role = 'admin' then uapp.user_name
			when n.type = 7 and uapp.role = 'tc' then uapp.user_name
			when n.type = 8 and uapp.role = 'customer' then uapp.user_name
		end additional_title,
		case 
			when n.type = 3 then mi.details
			else null
		end itin_details,
		n.created_date
	from notifications n 
	join (
		select
			case 
				when m.id is not null then m.id
				when u.id is not null then u.id 
				else null 
			end user_id,
			case 
				when m.member_code is not null then m.member_code
				when u.user_code is not null then u.user_code 
				else null 
			end user_code,
			case 
				when m.name is not null then m.name
				when u.name is not null then u.name 
				else null 
			end user_name,
			case 
				when m.email is not null then m.email
				when u.email is not null then u.email 
				else null 
			end user_email,
			case 
				when m.phone is not null then m.phone
				when u.phone is not null then u.phone 
				else null 
			end user_phone,
			case 
				when m.img is not null then m.img
				when u.img is not null then u.img 
				else null 
			end user_img,
			n.role
		from notifications n 
		left join members m on m.id = n.user_id and n.role = 'customer' and m.deleted_date is null
		left join users u on u.id = n.user_id and n.role != 'customer' and u.deleted_date is null	
	) uapp on uapp.user_id = n.user_id
	left join member_itins mi on mi.created_by = uapp.user_id and uapp.role = 'customer' and mi.deleted_date is null
	left join itin_suggestions its on its.created_by = uapp.user_id and uapp.role != 'customer' and its.deleted_date is null `

	var queryString string = query

	if len(param["user_code"].(string)) > 0 {
		var orWhere []string
		orWhere = append(orWhere, "uapp.user_code = '"+param["user_code"].(string)+"'")

		where = append(where, strings.Join(orWhere, " AND "))
	}

	if len(param["keyword"].(string)) > 0 {
		var orWhere []string
		orWhere = append(orWhere, "lower(n.code) like lower('%"+param["keyword"].(string)+"%')")
		orWhere = append(orWhere, "lower(n.title) like lower('%"+param["keyword"].(string)+"%')")
		orWhere = append(orWhere, "lower(n.content) like lower('%"+param["keyword"].(string)+"%')")
		orWhere = append(orWhere, "lower(n.subject) like lower('%"+param["keyword"].(string)+"%')")

		where = append(where, "("+strings.Join(orWhere, " OR ")+")")
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
		return listNotification, err
	}

	defer rows.Close()
	for rows.Next() {
		var n NotificationEnt
		err = rows.Scan(&n.Code, &n.Subject, &n.Type, &n.TypeText, &n.Title, &n.Content, &n.Link, &n.IsRead, &n.User.UserCode, &n.User.Name, &n.User.Email, &n.User.Phone, &n.User.Img, &n.Role, &n.AdditionalTitle, &n.MemberItin.Details, &n.CreatedDate)
		if err != nil {
			return listNotification, err
		}

		for _, v := range n.MemberItin.Details {
			c, err := json.Marshal(v["visit_list"])
			if err != nil {
				return listNotification, err
			}
			n.MemberItin.DayPeriod = int32(strings.Count(string(c), "]"))
		}

		listNotification = append(listNotification, n)
	}
	return listNotification, err
}

// GetNotification ...
func (c *Contract) GetNotificationByCode(db *pgxpool.Conn, ctx context.Context, code string) (NotificationEnt, error) {
	var n NotificationEnt

	sql := `select distinct on(n.id)
		n.id,
		n.code notif_code,
		n.subject notif_subject,
		n.type notif_type,
		case 
			when n.type = 1 then 'Chat'
			when n.type = 2 then 'Payment'
			when n.type = 3 then 'Trip'
			when n.type = 4 then 'Itinerary'
			when n.type = 5 then 'User Activity'
			when n.type = 6 then 'Admin'
			when n.type = 7 then 'TC'
			when n.type = 8 then 'Customer'
			else null 
		end text_type,
		n.title notif_title,
		n.content notif_content,
		case 
			when n.link is null or n.link = '' then null
			else n.link
		end notif_link,
		n.is_read,
		uapp.user_code,
		uapp.user_name,
		uapp.user_email,
		uapp.user_phone,
		uapp.user_img,
		uapp.role,
		case 
			when n.type = 3 then its.title 
			when n.type = 4 then mi.title 
			when n.type = 5 then uapp.user_name
			when n.type = 6 and uapp.role = 'admin' then uapp.user_name
			when n.type = 7 and uapp.role = 'tc' then uapp.user_name
			when n.type = 8 and uapp.role = 'customer' then uapp.user_name
		end additional_title,
		case 
			when n.type = 3 then mi.details
			else null
		end itin_details,
		n.created_date
	from notifications n 
	join (
		select
			case 
				when m.id is not null then m.id
				when u.id is not null then u.id 
				else null 
			end user_id,
			case 
				when m.member_code is not null then m.member_code
				when u.user_code is not null then u.user_code 
				else null 
			end user_code,
			case 
				when m.name is not null then m.name
				when u.name is not null then u.name 
				else null 
			end user_name,
			case 
				when m.email is not null then m.email
				when u.email is not null then u.email 
				else null 
			end user_email,
			case 
				when m.phone is not null then m.phone
				when u.phone is not null then u.phone 
				else null 
			end user_phone,
			case 
				when m.img is not null then m.img
				when u.img is not null then u.img 
				else null 
			end user_img,
			n.role
		from notifications n 
		left join members m on m.id = n.user_id and n.role = 'customer' and m.deleted_date is null
		left join users u on u.id = n.user_id and n.role != 'customer' and u.deleted_date is null	
	) uapp on uapp.user_id = n.user_id
	left join member_itins mi on mi.created_by = uapp.user_id and uapp.role = 'customer' and mi.deleted_date is null
	left join itin_suggestions its on its.created_by = uapp.user_id and uapp.role != 'customer' and its.deleted_date is null
	where n.code = $1`

	err := db.QueryRow(ctx, sql, code).Scan(&n.ID, &n.Code, &n.Subject, &n.Type, &n.TypeText, &n.Title, &n.Content, &n.Link, &n.IsRead, &n.User.UserCode, &n.User.Name, &n.User.Email, &n.User.Phone, &n.User.Img, &n.Role, &n.AdditionalTitle, &n.MemberItin.Details, &n.CreatedDate)

	return n, err
}

// UpdateNotificationByCode edit notification
func (c *Contract) UpdateNotificationByCode(tx pgx.Tx, ctx context.Context, n NotificationEnt, code string) (NotificationEnt, error) {
	var ID int32
	var paramQuery []interface{}

	sql := `UPDATE notifications SET type=$1, title=$2, content=$3, link=$4, is_read=$5, subject=$6 WHERE code=$7 RETURNING id`
	paramQuery = append(paramQuery, n.Type, n.Title, n.Content, n.Link, n.IsRead, n.Subject, code)

	err := tx.QueryRow(ctx, sql, paramQuery...).Scan(&ID)

	n.ID = ID

	return n, err
}

func (c *Contract) SetNotifContent(userID int64, typeSubj int, role, subject, title, desc, link string) NotificationEnt {
	notif := NotificationEnt{
		UserID:  userID,
		Type:    int32(typeSubj),
		Title:   title,
		Content: desc,
		Subject: subject,
		Role:    role,
	}
	if len(link) > 0 {
		notif.Link = sql.NullString{String: link, Valid: true}
	}

	return notif
}

func (c *Contract) GetNotifChatIncome(userID int64, role, msg, chatRoom string) NotificationEnt {
	title := fmt.Sprintf("You have %s unread chat from %s", msg, chatRoom)
	desc := fmt.Sprintf("Go to %s to reply them", chatRoom)

	return c.SetNotifContent(userID, NOTIF_TYPE_CHAT, role, NOTIF_SUBJ_CHAT_INCOME, title, desc, "")
}

func (c *Contract) GetNotifChatUnread(userID int64, role, roomName string) NotificationEnt {
	title := fmt.Sprintf("You have unread chats from %s", roomName)
	desc := "Check them out!"

	return c.SetNotifContent(userID, NOTIF_TYPE_CHAT, role, NOTIF_SUBJ_CHAT_UNREAD, title, desc, "")
}

func (c *Contract) GetNotifChatRoomAssigned(userID int64, role, roomName string) NotificationEnt {
	title := fmt.Sprintf(`Congratulation! You have assigned to a new chat room "%s"`, roomName)
	desc := "Say hi to your new chat room"

	return c.SetNotifContent(userID, NOTIF_TYPE_CHAT, role, NOTIF_SUBJ_CHAT_ROOM_ASSIGNED, title, desc, "")
}

func (c *Contract) GetNotifChatClientCompletedPayment(userID int64, role, roomName, clientName, orderCode string) NotificationEnt {
	title := fmt.Sprintf("%s assigne %s has completed the payment", roomName, clientName)
	desc := fmt.Sprintf("%s has completed the payment for order ID %s", clientName, orderCode)

	return c.SetNotifContent(userID, NOTIF_TYPE_CHAT, role, NOTIF_SUBJ_ORDER_CLIENT_COMPLETE, title, desc, "")
}

func (c *Contract) GetNotifChatClientFailedPayment(userID int64, role, roomName, clientName, orderCode string) NotificationEnt {
	title := fmt.Sprintf("%s assigne %s failed to complete the payment", roomName, clientName)
	desc := fmt.Sprintf("%s failed to complete the payment for order ID %s", clientName, orderCode)

	return c.SetNotifContent(userID, NOTIF_TYPE_CHAT, role, NOTIF_SUBJ_ORDER_CLIENT_FAIL, title, desc, "")
}

func (c *Contract) GetNotifPaymentIncome(userID int64, role, tripName, orderCode string) NotificationEnt {
	title := fmt.Sprintf("Waiting payment for %s (%s)", tripName, orderCode)
	desc := "Please make payment for your order immediately"

	return c.SetNotifContent(userID, NOTIF_TYPE_ORDER, role, NOTIF_SUBJ_ORDER_INCOME, title, desc, "")
}

func (c *Contract) GetNotifPaymentVerified(userID int64, role, tripName, orderCode, paymentMethod string) NotificationEnt {
	title := fmt.Sprintf("Payment %s has been verified", paymentMethod)
	desc := fmt.Sprintf("Thank you we have received your payment for %s and %s, please wait for further notification.", tripName, orderCode)

	return c.SetNotifContent(userID, NOTIF_TYPE_ORDER, role, NOTIF_SUBJ_ORDER_VERIF, title, desc, "")
}

func (c *Contract) GetNotifPaymentCancelled(userID int64, role, tripName, orderCode, paymentMethod string) NotificationEnt {
	title := fmt.Sprintf("Payment %s has been cancelled", paymentMethod)
	desc := fmt.Sprintf("Payment for %s and %s has been cancelled", tripName, orderCode)

	return c.SetNotifContent(userID, NOTIF_TYPE_ORDER, role, NOTIF_SUBJ_ORDER_CANCEL, title, desc, "")
}

func (c *Contract) GetNotifPaymentFailed(userID int64, role, tripName, orderCode, paymentMethod string) NotificationEnt {
	title := fmt.Sprintf("Payment %s failed", paymentMethod)
	desc := fmt.Sprintf("Payment for %s and %s is failed, please reach us for immediate assistance", tripName, orderCode)

	return c.SetNotifContent(userID, NOTIF_TYPE_ORDER, role, NOTIF_SUBJ_ORDER_FAIL, title, desc, "")
}

func (c *Contract) GetNotifPaymentHistory(userID int64, role, tripName, statusPayment string) NotificationEnt {
	title := "Payment History"
	desc := fmt.Sprintf("%s status payment is %s", tripName, statusPayment)

	return c.SetNotifContent(userID, NOTIF_TYPE_ORDER, role, NOTIF_SUBJ_ORDER_HISTORY, title, desc, "")
}

func (c *Contract) GetNotifMbitinPre(userID int64, role string, day int) NotificationEnt {
	title := fmt.Sprintf("Yay! your trip will begin in %d days", day)
	desc := fmt.Sprintf("Prepare for the best possible times of your life in %d days", day)

	return c.SetNotifContent(userID, NOTIF_TYPE_MBITIN, role, NOTIF_SUBJ_MBITIN_PRE, title, desc, "")
}

func (c *Contract) GetNotifMbitinBegin(userID int64, role, tripName string) NotificationEnt {
	title := "Trip day is here!"
	desc := fmt.Sprintf(`Please enjoy your "%s" trip and happy vacation!`, tripName)

	return c.SetNotifContent(userID, NOTIF_TYPE_MBITIN, role, NOTIF_SUBJ_MBITIN_BEGIN, title, desc, "")
}

func (c *Contract) GetNotifProfilChanged(userID int64, role, info string) NotificationEnt {
	title := fmt.Sprintf("Your %s profile has been updated", info)
	desc := "If this is not you, you should check this activity and secure your account"

	return c.SetNotifContent(userID, NOTIF_TYPE_PROFILE, role, NOTIF_SUBJ_PROFILE_CHANGE, title, desc, "")
}

func (c *Contract) GetNotifCustomerBanned(userID int64, role, custName string) NotificationEnt {
	title := fmt.Sprintf("%s has been banned from the Panorama", custName)
	desc := "If this is a mistake, you should check this activity"

	return c.SetNotifContent(userID, NOTIF_TYPE_CUSTOMER, role, NOTIF_SUBJ_CUSTOMER_BANNED, title, desc, "")
}

func (c *Contract) GetNotifTCAdded(userID int64, role, tcName string) NotificationEnt {
	title := "New travel consultant added"
	desc := fmt.Sprintf("Say hi to our new travel consultant %s", tcName)

	return c.SetNotifContent(userID, NOTIF_TYPE_TC, role, NOTIF_SUBJ_TC_ADD, title, desc, "")
}

func (c *Contract) GetNotifTCInvited(userID int64, role, chatRoom string) NotificationEnt {
	title := fmt.Sprintf(`Someone invited Travel Consultant to "%s"`, chatRoom)
	desc := fmt.Sprintf(`A travel consultant has been invited to "%s", please ask everything or book accommodations to our TC`, chatRoom)

	return c.SetNotifContent(userID, NOTIF_TYPE_TC, role, NOTIF_SUBJ_TC_INVITED, title, desc, "")
}

func (c *Contract) GetNotifTCChanged(userID int64, role, chatRoom string) NotificationEnt {
	title := fmt.Sprintf(`your "%s" travel consultant has changed`, chatRoom)
	desc := fmt.Sprintf(`"%s" 's travel consultant has been changed, say hi to new travel consultant!`, chatRoom)

	return c.SetNotifContent(userID, NOTIF_TYPE_TC, role, NOTIF_SUBJ_TC_CHANGED, title, desc, "")
}

func (c *Contract) GetNotifTcRemoved(userID int64, role, tcName string) NotificationEnt {
	title := "Travel consultant access removed"
	desc := fmt.Sprintf("%s access to Panorama has been removed", tcName)

	return c.SetNotifContent(userID, NOTIF_TYPE_TC, role, NOTIF_SUBJ_TC_REMOVE, title, desc, "")
}

func (c *Contract) GetNotifAdminAdded(userID int64, role, adminName string) NotificationEnt {
	title := "New Panorama administrator added"
	desc := fmt.Sprintf("Say hi to our new admin %s", adminName)

	return c.SetNotifContent(userID, NOTIF_TYPE_ADMIN, role, NOTIF_SUBJ_ADMIN_ADD, title, desc, "")
}

func (c *Contract) GetNotifSuggitinNew(userID int64, role, sugItinTitle, adminName string) NotificationEnt {
	var byAdminName string
	if len(adminName) > 0 {
		byAdminName = fmt.Sprintf(`by %s`, adminName)
	}
	title := "New suggested itinerary has been published"
	desc := fmt.Sprintf(`Checkout "%s", a new suggested itinerary %s`, sugItinTitle, byAdminName)

	return c.SetNotifContent(userID, NOTIF_TYPE_SUGGITIN, role, NOTIF_SUBJ_SUGGITIN_NEW, title, desc, "")
}

func (c *Contract) GetNotifStuffNew(userID int64, role, StuffName, adminName string) NotificationEnt {
	var byAdminName string
	if len(adminName) > 0 {
		byAdminName = fmt.Sprintf(`by %s`, adminName)
	}
	title := "New stuff has been published"
	desc := fmt.Sprintf(`Checkout "%s", a new stuff %s`, StuffName, byAdminName)

	return c.SetNotifContent(userID, NOTIF_TYPE_STUFF, role, NOTIF_SUBJ_STUFF_NEW, title, desc, "")
}

func (c *Contract) SendNotifications(tx pgx.Tx, db *pgxpool.Conn, ctx context.Context, players []DeviceListEnt, content NotificationContent) ([]NotificationEnt, error) {
	var listPlayerID []string
	var notifications []NotificationEnt

	// 	Send notification - Set notif to players
	if len(players) > 0 {
		// Send notification - Set notification temp for content notif
		var notification NotificationEnt

		for _, p := range players {

			if len(p.PlayerID) <= 0 {
				continue
			}
			// Send notification - Grouping player id into list
			listPlayerID = append(listPlayerID, p.PlayerID)

			// Send notification - Set user name player
			var userName string
			userID := fmt.Sprintf("%d", p.UserID)
			if p.Role == "customer" {
				member, _ := c.GetMemberBy(db, ctx, "id", userID)
				userName = member.Name
			} else {
				user, _ := c.GetUserBy(db, ctx, "id", userID)
				userName = user.Name
			}

			// Send notification - Get data notif
			var notifContent NotificationEnt

			switch content.Subject {
			case NOTIF_SUBJ_CHAT_INCOME:
			case NOTIF_SUBJ_CHAT_UNREAD:
			case NOTIF_SUBJ_CHAT_ROOM_ASSIGNED:
				notifContent = c.GetNotifChatRoomAssigned(p.UserID, p.Role, content.RoomName)
			case NOTIF_SUBJ_ORDER_INCOME:
				notifContent = c.GetNotifPaymentIncome(p.UserID, p.Role, content.TripName, content.OrderCode)
			case NOTIF_SUBJ_ORDER_VERIF:
				notifContent = c.GetNotifPaymentVerified(p.UserID, p.Role, content.TripName, content.OrderCode, content.PaymentMethod)
			case NOTIF_SUBJ_ORDER_CANCEL:
				notifContent = c.GetNotifPaymentCancelled(p.UserID, p.Role, content.TripName, content.OrderCode, content.PaymentMethod)
			case NOTIF_SUBJ_ORDER_FAIL:
				notifContent = c.GetNotifPaymentFailed(p.UserID, p.Role, content.TripName, content.OrderCode, content.PaymentMethod)
			case NOTIF_SUBJ_ORDER_HISTORY:
				notifContent = c.GetNotifPaymentHistory(p.UserID, p.Role, content.TripName, content.StatusPayment)
			case NOTIF_SUBJ_ORDER_CLIENT_COMPLETE:
				notifContent = c.GetNotifChatClientCompletedPayment(p.UserID, p.Role, content.RoomName, content.ClientName, content.OrderCode)
			case NOTIF_SUBJ_ORDER_CLIENT_FAIL:
				notifContent = c.GetNotifChatClientFailedPayment(p.UserID, p.Role, content.RoomName, content.ClientName, content.OrderCode)
			case NOTIF_SUBJ_MBITIN_PRE:
			case NOTIF_SUBJ_MBITIN_BEGIN:
			case NOTIF_SUBJ_SUGGITIN_NEW:
				notifContent = c.GetNotifSuggitinNew(p.UserID, p.Role, content.SugItinTitle, content.AdminName)
			case NOTIF_SUBJ_PROFILE_CHANGE:
				notifContent = c.GetNotifProfilChanged(p.UserID, p.Role, userName)
			case NOTIF_SUBJ_ADMIN_ADD:
				notifContent = c.GetNotifAdminAdded(p.UserID, p.Role, content.AdminName)
			case NOTIF_SUBJ_TC_INVITED:
			case NOTIF_SUBJ_TC_CHANGED:
			case NOTIF_SUBJ_TC_ADD:
				notifContent = c.GetNotifTCAdded(p.UserID, p.Role, content.TCName)
			case NOTIF_SUBJ_TC_REMOVE:
				notifContent = c.GetNotifTcRemoved(p.UserID, p.Role, content.TCName)
			case NOTIF_SUBJ_CUSTOMER_BANNED:
				notifContent = c.GetNotifCustomerBanned(p.UserID, p.Role, content.CustomerName)
			case NOTIF_SUBJ_STUFF_NEW:
				notifContent = c.GetNotifStuffNew(p.UserID, p.Role, content.StuffName, content.AdminName)
			default:
				return nil, fmt.Errorf("%s", "invalid subject")
			}

			notificationSaved, err := c.AddNotif(tx, ctx, notifContent)
			if err != nil {
				return notifications, err
			}
			notification = notificationSaved
			notifications = append(notifications, notificationSaved)
		}

		// Send notification - Send blast data notif into players
		if len(listPlayerID) > 0 {
			_, err := onesignal.New(c.App).PushNotification(notification.Title, notification.Content, listPlayerID)
			if err != nil {
				return notifications, err
			}
		}
	}

	return notifications, nil
}

func (c *Contract) GetCounterNotifByUserCode(db *pgxpool.Conn, ctx context.Context, userCode string, isRead bool) (map[string]int, error) {
	var notifAll, notifOrder, notifTrip int

	sql := `select (
		select
			count(notif.id) notif_all
		from (
			select 
				n.id,
				uapp.user_code
			from notifications n 
			join (
				select
					case 
						when m.id is not null then m.id
						when u.id is not null then u.id 
						else null 
					end user_id,
					case 
						when m.member_code is not null then m.member_code
						when u.user_code is not null then u.user_code 
						else null 
					end user_code
				from notifications n 
				left join members m on m.id = n.user_id and n.role = 'customer' and m.deleted_date is null
				left join users u on u.id = n.user_id and n.role != 'customer' and u.deleted_date is null	
			) uapp on uapp.user_id = n.user_id
			where n.is_read = $4
			group by n.id, uapp.user_code
		) notif
		where notif.user_code = $1
	),
	(
		select
			count(notif.id) notif_order
		from (
			select 
				n.id,
				uapp.user_code
			from notifications n 
			join (
				select
					case 
						when m.id is not null then m.id
						when u.id is not null then u.id 
						else null 
					end user_id,
					case 
						when m.member_code is not null then m.member_code
						when u.user_code is not null then u.user_code 
						else null 
					end user_code
				from notifications n 
				left join members m on m.id = n.user_id and n.role = 'customer' and m.deleted_date is null
				left join users u on u.id = n.user_id and n.role != 'customer' and u.deleted_date is null	
			) uapp on uapp.user_id = n.user_id
			where n.is_read = $4
			and n.type = $2
			group by n.id, uapp.user_code
		) notif
		where notif.user_code = $1
	),
	(
		select
			count(notif.id) notif_trip
		from (
			select 
				n.id,
				uapp.user_code
			from notifications n 
			join (
				select
					case 
						when m.id is not null then m.id
						when u.id is not null then u.id 
						else null 
					end user_id,
					case 
						when m.member_code is not null then m.member_code
						when u.user_code is not null then u.user_code 
						else null 
					end user_code
				from notifications n 
				left join members m on m.id = n.user_id and n.role = 'customer' and m.deleted_date is null
				left join users u on u.id = n.user_id and n.role != 'customer' and u.deleted_date is null	
			) uapp on uapp.user_id = n.user_id
			where n.is_read = $4
			and n.type = $3
			group by n.id, uapp.user_code
		) notif
		where notif.user_code = $1
	)`

	err := db.QueryRow(ctx, sql, userCode, NOTIF_TYPE_ORDER, NOTIF_TYPE_MBITIN, isRead).Scan(&notifAll, &notifOrder, &notifTrip)
	result := map[string]int{
		"notif_all":   notifAll,
		"notif_order": notifOrder,
		"notif_trip":  notifTrip,
	}

	return result, err
}
