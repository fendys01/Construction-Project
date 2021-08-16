package model

import (
	"context"
	"strconv"
	"strings"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4/pgxpool"
)

type TagEnt struct {
	ID   int32
	Name string
}

type SugItinTagsEnt struct {
	ID        int32
	SugItinID int32
	TagID     int32
}

// GetTags ...
func (c *Contract) GetTags(db *pgxpool.Conn, ctx context.Context) ([]TagEnt, error) {
	var t []TagEnt
	err := pgxscan.Select(ctx, db, &t, "select * from tags")

	return t, err
}

// AddMutiTag ...
func (c *Contract) AddMutiTag(db *pgxpool.Conn, ctx context.Context, t []string) []int32 {
	var ids []int32
	var lastInsID int32
	for _, v := range t {
		db.QueryRow(context.Background(), "insert into tags (tag_name) values($1) RETURNING id;", v).Scan(&lastInsID)
		ids = append(ids, lastInsID)
	}

	return ids
}

// AddMultiSugItinTags ..
func (c *Contract) AddMultiSugItinTags(db *pgxpool.Conn, ctx context.Context, sug int32, tags []int32) error {
	sql := "insert into itin_suggestion_tags (itin_sug_id, tag_id) values "

	var arrStr []string
	for _, v := range tags {
		arrStr = append(arrStr, "("+strconv.Itoa(int(sug))+","+strconv.Itoa(int(v))+")")
	}

	sql = sql + strings.Join(arrStr, ",")

	_, err := db.Exec(context.Background(), sql)

	return err
}

// DelSugItinTagsBySugItinID delete all tags of suggestion by suggestion id
func (c *Contract) DelSugItinTagsBySugItinID(db *pgxpool.Conn, ctx context.Context, sugItinID int32) error {
	sql := `delete from itin_suggestion_tags where itin_sug_id=$1;`

	_, err := db.Exec(context.Background(), sql, sugItinID)

	return err
}
