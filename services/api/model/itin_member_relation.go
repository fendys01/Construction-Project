package model

import (
	"context"
	"database/sql"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type MemberItinRelationEnt struct {
	ID            int32
	MemberItinID  int32
	MemberID      int32
	CreatedDate   time.Time
	DeletedDate   sql.NullTime
	MemberItinEnt MemberItinEnt
	MemberEnt     MemberEnt
	ChatGroup     ChatGroupEnt
	IsOwner       bool
	IsTemporary   bool
}

// AddMemberItinRelation add new member itin relation
func (c *Contract) AddMemberItinRelation(tx pgx.Tx, ctx context.Context, m MemberItinRelationEnt) (MemberItinRelationEnt, error) {
	var lastInsID int32
	timeStamp := time.Now().In(time.UTC)

	sql := `INSERT INTO member_itin_relations(member_itin_id, member_id, created_date) VALUES($1, $2, $3) RETURNING id`

	err := tx.QueryRow(ctx, sql, m.MemberItinID, m.MemberID, timeStamp).Scan(&lastInsID)

	m.ID = lastInsID

	return m, err
}

// DeleteMemberItinRelation delete member itin relation
func (c *Contract) DeleteMemberItinRelation(tx pgx.Tx, ctx context.Context, memberID, itinID int32) error {
	sql := `delete from member_itin_relations where member_id = $1 and member_itin_id = $2`
	_, err := tx.Exec(ctx, sql, memberID, itinID)

	return err
}

// GetMemberRelationByMemberID Get member itin relation by Member ID
func (c *Contract) GetMemberItinRelationByMemberIDAndMemberItinID(db *pgxpool.Conn, ctx context.Context, memberID int32, memberItinID int32) (MemberItinRelationEnt, error) {
	var m MemberItinRelationEnt

	sqlM := `SELECT id, member_itin_id, member_id, created_date FROM member_itin_relations WHERE member_id = $1 AND member_itin_id = $2 AND deleted_date IS NULL`

	err := db.QueryRow(ctx, sqlM, memberID, memberItinID).Scan(&m.ID, &m.MemberItinID, &m.MemberID, &m.CreatedDate)

	return m, err
}

// AddMemberItinRelationBatch add new member temporary batch
func (c *Contract) AddMemberItinRelationBatch(ctx context.Context, tx pgx.Tx, arrTemp string) error {
	var lastInsID int32

	sql := `INSERT INTO member_itin_relations(member_itin_id, member_id, created_date) VALUES ` + arrTemp + ` RETURNING id`
	err := tx.QueryRow(ctx, sql).Scan(&lastInsID)

	return err
}

// GetListMemberItinRelationByItinID Get member relation list by itin ID
func (c *Contract) GetListMemberItinRelationByItinID(db *pgxpool.Conn, ctx context.Context, itinID int32) ([]MemberItinRelationEnt, error) {
	var mList []MemberItinRelationEnt

	query := `select 
		gm.itin_id,
		gm.itin_code,
		gm.itin_title,
		gm.member_id,
		gm.member_code,
		gm.member_name,
		gm.member_username,
		gm.member_email,
		gm.member_img,
		gm.is_owner,
		gm.is_temporary,
		cg.chat_group_code,
		cg.name chat_group_name,
		cg.chat_group_type,
		cg.status chat_group_status
	from (
		select
			mi.id itin_id,
			mi.itin_code,
			mi.title itin_title,
			m.id member_id,
			m.member_code, 
			m.name member_name, 
			m.username member_username, 
			m.email member_email,
			m.img member_img,
			case 
				when m.member_code is not null then true
				else true 
			end is_owner,
			case 
				when m.member_code is null then false
				else false 
			end is_temporary
		from members m
		join member_itins mi on mi.created_by = m.id 
		union
		select
			mi.id itin_id,
			mi.itin_code,
			mi.title itin_title,
			m.id member_id,
			m.member_code, 
			m.name member_name, 
			m.username member_username, 
			m.email member_email,
			m.img member_img,
			case 
				when m.member_code is not null then false
				else false 
			end is_owner,
			case 
				when m.member_code is null then false
				else false 
			end is_temporary
		from member_itin_relations mir
		join member_itins mi on mi.id = mir.member_itin_id 
		join members m on m.id = mir.member_id 
		where mir.deleted_date is null
		union
		select
			mi.id itin_id,
			mi.itin_code,
			mi.title itin_title,
			m.id member_id,
			m.member_code, 
			m.name member_name, 
			m.username member_username, 
			mt.email member_email,
			m.img member_img,
			case 
				when m.member_code is not null then false
				else false 
			end is_owner,
			case 
				when mt.email is not null then true
				else true 
			end is_temporary
		from member_temporaries mt
		left join member_itins mi on mi.id = mt.member_itin_id 
		left join members m on m.email = mt.email 
	) gm
	left join chat_groups cg on cg.member_itin_id = gm.itin_id
	where gm.itin_id = $1`

	rows, err := db.Query(ctx, query, itinID)
	if err != nil {
		return mList, err
	}

	var memberID sql.NullInt32
	var memberCode, memberName, memberUsername, memberEmail, chatGroupCode, chatGroupName, chatGroupType sql.NullString
	var chatGroupStatus sql.NullBool

	for rows.Next() {
		var m MemberItinRelationEnt
		err = rows.Scan(&m.MemberItinEnt.ID, &m.MemberItinEnt.ItinCode, &m.MemberItinEnt.Title, &memberID, &memberCode, &memberName, &memberUsername, &memberEmail, &m.MemberEnt.Img, &m.IsOwner, &m.IsTemporary, &chatGroupCode, &chatGroupName, &chatGroupType, &chatGroupStatus)
		if err != nil {
			return mList, err
		}

		m.MemberEnt.ID = memberID.Int32
		m.MemberEnt.MemberCode = memberCode.String
		m.MemberEnt.Name = memberName.String
		m.MemberEnt.Username = memberUsername.String
		m.MemberEnt.Email = memberEmail.String
		m.ChatGroup.ChatGroupCode = chatGroupCode.String
		m.ChatGroup.Name = chatGroupName.String
		m.ChatGroup.ChatGroupType = chatGroupType.String
		m.ChatGroup.Status = chatGroupStatus.Bool

		mList = append(mList, m)
	}

	return mList, err
}

// GetListItinMemberRelationByItinID Get list member itin relation by itin ID
func (c *Contract) GetListItinMemberRelationByItinID(db *pgxpool.Conn, ctx context.Context, itinID int32) ([]MemberItinRelationEnt, error) {
	var list []MemberItinRelationEnt

	sql := `SELECT id, member_itin_id, member_id, created_date FROM member_itin_relations WHERE member_itin_id = $1`

	rows, err := db.Query(ctx, sql, itinID)
	if err != nil {
		return list, err
	}

	for rows.Next() {
		var m MemberItinRelationEnt
		err = rows.Scan(&m.ID, &m.MemberItinID, &m.MemberID, &m.CreatedDate)
		if err != nil {
			return list, err
		}

		list = append(list, m)
	}

	return list, err
}
