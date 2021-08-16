package model

import (
	"Contruction-Project/lib/utils"
	"context"
	"database/sql"
	"math/rand"
	"time"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4/pgxpool"
)

type UserEnt struct {
	ID          int32
	UserCode    string
	Name        string
	Email       string
	Phone       string
	Password    string
	Role        string // admin, tc
	Img         sql.NullString
	IsActive    bool
	CreatedDate time.Time
	UpdatedDate sql.NullTime
	DeletedDate sql.NullTime
	LastVisit   time.Time
	TotalClient int32
}

func (c *Contract) createUserCode() string {
	rand.Seed(time.Now().UnixNano())
	code, _ := utils.Generate(`u-[a-z0-9]{8}`)
	return code
}

// GetUser ...
func (c *Contract) GetUser(db *pgxpool.Conn, ctx context.Context) ([]UserEnt, error) {
	var u []UserEnt

	err := pgxscan.Select(ctx, db, &u, "select * from users order by id desc")
	return u, err
}

// GetUserByCode ...
func (c *Contract) GetUserByCode(db *pgxpool.Conn, ctx context.Context, code string) (UserEnt, error) {
	var u UserEnt

	err := pgxscan.Get(ctx, db, &u, "select * from users where user_code=$1 limit 1", code)
	return u, err
}

// GetUserByEmail ...
func (c *Contract) GetUserByEmail(db *pgxpool.Conn, ctx context.Context, email string) (UserEnt, error) {
	var u UserEnt

	err := pgxscan.Get(ctx, db, &u, "select * from users where email=$1 limit 1", email)
	return u, err
}

func (c *Contract) isUserExists(db *pgxpool.Conn, ctx context.Context, email string) bool {
	_, err := c.GetUserByEmail(db, ctx, email)
	if err != nil {
		return false
	}

	return true
}

// GetUserByID ...
func (c *Contract) GetUserByID(db *pgxpool.Conn, ctx context.Context, id int32) (UserEnt, error) {
	var u UserEnt

	err := pgxscan.Get(ctx, db, &u, "select * from users where id=$1", id)
	return u, err
}

// AddUser ...
func (c *Contract) AddUser(db *pgxpool.Conn, ctx context.Context, u UserEnt) error {

	_, err := db.Exec(context.Background(),
		"insert into users (name, user_code, phone, email, password, role, is_active, created_date) values($1,$2,$3,$4,$5,$6,$7,$8)",
		u.Name, c.createUserCode(), u.Phone, u.Email, u.Password, u.Role, u.IsActive, time.Now().In(time.UTC),
	)

	return err
}

// UpdateUser ...
func (c *Contract) UpdateUser(db *pgxpool.Conn, ctx context.Context, code string, u UserEnt) error {
	_, err := db.Exec(context.Background(),
		"update users set name=$1, role=$2, is_active=$3, phone=$4, updated_date=$5 where user_code=$6",
		u.Name, u.Role, u.IsActive, u.Phone, time.Now().In(time.UTC), code,
	)

	return err
}

// UpdateUserPass ...
func (c *Contract) UpdateUserPass(db *pgxpool.Conn, ctx context.Context, code, pass string) error {
	_, err := db.Exec(context.Background(),
		"update users set password=$1, updated_date=$2 where user_code=$3",
		pass, time.Now().In(time.UTC), code,
	)

	return err
}

// UpdateEmail ...
func (c *Contract) UpdateEmail(db *pgxpool.Conn, ctx context.Context, id int32, email string) error {
	_, err := db.Exec(context.Background(),
		"update users set email=$1, updated_date=$2 where id=$3",
		email, time.Now().In(time.UTC), id,
	)

	if err != nil {
		return err
	}

	// TODO: need to send email confirmation for change email

	return nil
}
