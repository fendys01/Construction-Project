package model

import (
	"context"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

// MemberTemporaryEnt ...
type MemberTemporaryEnt struct {
	ID           int32
	Email        string
	MemberItinID int32
	CreatedDate  time.Time
	MemberItin   MemberItinEnt
}

// AddMemberTemporary add new member temporary
func (c *Contract) AddMemberTemporary(tx pgx.Tx, ctx context.Context, m MemberTemporaryEnt) (MemberTemporaryEnt, error) {
	var lastInsID int32
	timeStamp := time.Now().In(time.UTC)

	sql := `INSERT INTO member_temporaries(email, member_itin_id, created_date) VALUES($1, $2, $3) RETURNING id`

	err := tx.QueryRow(ctx, sql, m.Email, m.MemberItinID, timeStamp).Scan(&lastInsID)

	m.ID = lastInsID
	m.CreatedDate = timeStamp

	return m, err
}

// GetMemberTemporaryByEmailAndItinID Get member itin relation by email & itin
func (c *Contract) GetMemberTemporaryByEmailAndItinID(db *pgxpool.Conn, ctx context.Context, email string, memberItinID int32) (MemberTemporaryEnt, error) {
	var m MemberTemporaryEnt

	sqlM := `SELECT id, email, member_itin_id, created_date FROM member_temporaries WHERE email = $1 AND member_itin_id = $2`

	err := db.QueryRow(ctx, sqlM, email, memberItinID).Scan(&m.ID, &m.Email, &m.MemberItinID, &m.CreatedDate)

	return m, err
}

// GetListMemberTemporaryByEmail Get member temporary list by email
func (c *Contract) GetListMemberTemporaryByEmail(db *pgxpool.Conn, ctx context.Context, email string) ([]MemberTemporaryEnt, error) {
	var mList []MemberTemporaryEnt

	query := `SELECT id, email, member_itin_id, created_date FROM member_temporaries WHERE email = $1`

	rows, err := db.Query(ctx, query, email)
	if err != nil {
		return mList, err
	}

	for rows.Next() {
		var m MemberTemporaryEnt
		err = rows.Scan(&m.ID, &m.Email, &m.MemberItinID, &m.CreatedDate)
		if err != nil {
			return mList, err
		}

		mList = append(mList, m)
	}

	return mList, err
}

// DeleteMemberTempByEmail delete member temporary by email
func (c *Contract) DeleteMemberTempByEmail(ctx context.Context, tx pgx.Tx, email string) error {

	sql := `delete from member_temporaries as mt where email = $1`

	pgx, err := tx.Query(ctx, sql, email)

	defer pgx.Close()

	return err
}

// GetListMemberTemporaryByItinID Get member temporary list by itin ID
func (c *Contract) GetListMemberTemporaryByItinID(db *pgxpool.Conn, ctx context.Context, memberItinID int32) ([]MemberTemporaryEnt, error) {
	var mList []MemberTemporaryEnt

	query := `SELECT id, email, member_itin_id, created_date FROM member_temporaries WHERE member_itin_id = $1`

	rows, err := db.Query(ctx, query, memberItinID)
	if err != nil {
		return mList, err
	}

	for rows.Next() {
		var m MemberTemporaryEnt
		err = rows.Scan(&m.ID, &m.Email, &m.MemberItinID, &m.CreatedDate)
		if err != nil {
			return mList, err
		}

		mList = append(mList, m)
	}

	return mList, err
}
