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

// UpdateMemberItinRelation update member itin relation
func (c *Contract) SoftDeleteMemberItinRelation(db *pgxpool.Conn, ctx context.Context, tx pgx.Tx, m MemberItinRelationEnt, memberID int32) error {
	sql := `UPDATE member_itin_relations SET deleted_date=$1 WHERE member_id=$2`
	_, err := tx.Exec(ctx, sql, time.Now().In(time.UTC), memberID)

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

// GetListMemberItinRelationByMemberItinID Get member relation list by itin ID
func (c *Contract) GetListMemberItinRelationByMemberItinID(db *pgxpool.Conn, ctx context.Context, memberItinID int32) ([]MemberItinRelationEnt, error) {
	var mList []MemberItinRelationEnt

	query := `SELECT id, member_itin_id, member_id, created_date FROM member_itin_relations WHERE member_itin_id = $1 AND deleted_date IS NULL`

	rows, err := db.Query(ctx, query, memberItinID)
	if err != nil {
		return mList, err
	}

	for rows.Next() {
		var m MemberItinRelationEnt
		err = rows.Scan(&m.ID, &m.MemberItinID, &m.MemberID, &m.CreatedDate)
		if err != nil {
			return mList, err
		}

		mList = append(mList, m)
	}

	return mList, err
}
