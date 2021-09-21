package model

import (
	"context"

	"github.com/jackc/pgx/v4"
)

// AddChatGroupRelationBatch add new member to chat group
func (c *Contract) AddChatGroupRelationBatch(ctx context.Context, tx pgx.Tx, arrStr string) error {
	var lastInsID int32

	sql := `INSERT INTO chat_group_relations(member_id, chat_group_id, created_date) VALUES ` + arrStr + ` RETURNING id`
	err := tx.QueryRow(ctx, sql).Scan(&lastInsID)

	return err
}
