package model

import (
	"context"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

// ChatGroupEnt ...
type ChatGroupEnt struct {
	ID            int32
	Member        MemberEnt
	MemberItin    MemberItinEnt
	User          UserEnt
	ChatGroupType string
	ChatGroupCode string
	Name          string
	CreatedDate   time.Time
}

// ChatMessagesEnt ...
type ChatMessagesEnt struct {
	ID          int32
	ChatGroupID int32
	UserID      int32
	Name        string
	Message     string
	Role        string
	IsRead      bool
	CreatedDate time.Time
}

// CreateChatGroup ...
func (c *Contract) CreateChatGroup(tx pgx.Tx, ctx context.Context, cg ChatGroupEnt) (ChatGroupEnt, error) {

	sql := "insert into chat_groups (created_by, member_itin_id, tc_id, chat_group_code, name, created_date,chat_group_type) values($1,$2,$3,$4,$5,$6,$7) RETURNING id"

	var lastInsID int32

	err := tx.QueryRow(context.Background(), sql,
		cg.Member.ID, cg.MemberItin.ID, cg.User.ID, cg.ChatGroupCode, cg.Name, time.Now().In(time.UTC), cg.ChatGroupType,
	).Scan(&lastInsID)

	cg.ID = lastInsID

	return cg, err
}

// UpdateItinMemberToChat update itin member id
func (c *Contract) UpdateItinMemberToChat(ctx context.Context, tx pgx.Tx, miID int32, code string) error {
	var ID int32

	sql := `UPDATE chat_groups SET member_itin_id=$1 WHERE chat_group_code=$2 RETURNING id`

	err := tx.QueryRow(ctx, sql, miID, code).Scan(&ID)

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
