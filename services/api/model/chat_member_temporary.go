package model

import (
	"context"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type ChatMemberTemporary struct {
	ID          int32
	Email       string
	ChatGroupID int32
	CreatedDate time.Time
}

// GetListChatMemberTempByEmail Get member temporary list by email
func (c *Contract) GetListChatMemberTempByEmail(db *pgxpool.Conn, ctx context.Context, email string) ([]ChatMemberTemporary, error) {
	var mList []ChatMemberTemporary

	query := `SELECT id, email, chat_group_id, created_date FROM chat_member_temporaries WHERE email = $1`

	rows, err := db.Query(ctx, query, email)
	if err != nil {
		return mList, err
	}

	for rows.Next() {
		var m ChatMemberTemporary
		err = rows.Scan(&m.ID, &m.Email, &m.ChatGroupID, &m.CreatedDate)
		if err != nil {
			return mList, err
		}

		mList = append(mList, m)
	}

	return mList, err
}

// DeleteChatMemberTempByEmail delete chat member temporary by email
func (c *Contract) DeleteChatMemberTempByEmail(ctx context.Context, tx pgx.Tx, email string) error {

	sql := `delete from chat_member_temporaries as cmt where email = $1`

	pgx, err := tx.Query(ctx, sql, email)

	defer pgx.Close()

	return err
}
