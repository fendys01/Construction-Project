package model

import (
	"Contruction-Project/bootstrap"
	"Contruction-Project/lib/utils"
	"context"
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4/pgxpool"
	mail "github.com/xhit/go-simple-mail/v2"
	"golang.org/x/crypto/bcrypt"
)

const (
	ActForgotPass  = "forgot_pass"
	ActChangeEmail = "change_email"
	ActRegPhone    = "reg_phone"
	ActRegEmail    = "reg_email"

	TokenViaEmail = "email"
	TokenViaPhone = "phone"

	tokenExpiredMin = 2 // in minutes
)

var tokenMailSubj = map[string]string{
	ActForgotPass:  "[Panorama] Forgot Password",
	ActRegEmail:    "[Panorama] Email Verification",
	ActChangeEmail: "[Panorama] Change Email Verification",
}

var errInvalidCred error = fmt.Errorf("%s", "invalid user credential")
var errInactiveUser error = fmt.Errorf("%s", "inactive user")

func (c *Contract) generateJWT(ch, userID, role, key string) (string, int64, error) {
	expirationTime := time.Now().Add(2160 * time.Hour).Unix()
	claims := bootstrap.CustomClaims{
		userID,
		ch,
		role,
		jwt.StandardClaims{
			Issuer:    userID,
			ExpiresAt: expirationTime,
		},
	}
	rawToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := rawToken.SignedString([]byte(key))

	return token, expirationTime, err
}

func (c *Contract) getMemberField(username string) string {
	var field string = "phone"
	if utils.IsEmail(username) {
		field = "email"
	} else if utils.IsUsernameTag(username) {
		field = "username"
	}

	return field
}

func (c *Contract) Via(username string) string {
	if utils.IsEmail(username) {
		return TokenViaEmail
	}

	_, _, valid := utils.IsPhone(username)
	if valid {
		return TokenViaPhone
	}

	return ""
}

func (c *Contract) isValidTokenAction(ch, via, username string) bool {
	switch ch {
	case ChannelCustApp:
		if via == TokenViaEmail {
			if utils.IsEmail(username) {
				return true
			}
		}

		if via == TokenViaPhone {
			_, _, valid := utils.IsPhone(username)
			if valid {
				return true
			}
		}

	case ChannelCMS:
		if utils.IsEmail(username) {
			return true
		}
	}

	return false
}

func (c *Contract) sendTokenMail(usedFor, subject, to, token string) {
	fn := fmt.Sprintf("%s/%s.html", c.Config.GetString("resource_path"), usedFor)

	server := mail.NewSMTPClient()

	// SMTP Server
	server.Host = c.Config.GetString("mail.host")
	server.Port = c.Config.GetInt("mail.port")
	server.Username = c.Config.GetString("mail.username")
	server.Password = c.Config.GetString("mail.password")
	server.Encryption = mail.EncryptionSTARTTLS

	// SMTP client
	smtpClient, err := server.Connect()
	if err != nil {
		log.Fatal(err)
		return
	}

	// fill the html body
	type data struct {
		Token string
	}
	tpl, err := utils.ParseTpl(fn, data{Token: token})
	if err != nil {
		panic(1)
	}

	// New email simple html with inline and CC
	from := fmt.Sprintf("%s <%s>", c.Config.GetString("mail.mail_name"), c.Config.GetString("mail.mail_from"))
	email := mail.NewMSG()
	email.SetFrom(from).
		AddTo(to).
		SetSubject(subject)

	email.SetBody(mail.TextHTML, tpl)
	if email.Error != nil {
		log.Fatal(email.Error)
	}

	// Call Send and pass the client
	err = email.Send(smtpClient)
	if err != nil {
		log.Println(err)
	} else {
		log.Println("Email Sent")
	}
}

// SendToken sending token for multiple action, type and channel
func (c *Contract) SendToken(db *pgxpool.Conn, ctx context.Context, ch, usedFor, via, username string) error {
	if !c.isValidTokenAction(ch, via, username) {
		return fmt.Errorf("%s", "send token invalid action")
	}

	// TODO: need to implement this into specific logic
	// if !c.isUsernameExists(db, ctx, ch, username) {
	// 	return "", fmt.Errorf("username doesn't exists")
	// }

	token, err := c.addNewToken(db, ctx, ch, usedFor, via, username)
	if err != nil {
		return err
	}
	fmt.Printf("\nchannel: %s\nused for: %s\nvia: %s\nuser: %s\ntoken: %s\n", ch, usedFor, via, username, token)
	switch ch {
	case ChannelCustApp:
		if via == TokenViaEmail {
			go c.sendTokenMail(usedFor, tokenMailSubj[usedFor], username, token)
		}

		if via == TokenViaPhone {
			// send SMS with token
		}

	case ChannelCMS:
		if via == TokenViaEmail {
			go c.sendTokenMail(usedFor, tokenMailSubj[usedFor], username, token)
		}
	}

	return nil
}

func (c *Contract) isUsernameExists(db *pgxpool.Conn, ctx context.Context, ch, username string) bool {
	switch ch {
	case ChannelCustApp:
		return c.isMemberExists(db, ctx, c.getMemberField(username), username)

	case ChannelCMS:
		return c.isUserExists(db, ctx, username)
	}

	return false
}

// addNewToken ...
func (c *Contract) addNewToken(db *pgxpool.Conn, ctx context.Context, ch, usedFor, via, username string) (string, error) {
	rand.Seed(time.Now().UnixNano())
	token, _ := utils.Generate(`[\d]{6}`)

	now := time.Now().In(time.UTC)

	_, err := db.Exec(context.Background(),
		`insert into token_logs(channel,used_for,via,username,token,exp_date,created_date)
		values($1,$2,$3,$4,$5,$6,$7)`,
		ch, usedFor, via, username, token, now.Add(time.Minute*tokenExpiredMin), now,
	)
	if err != nil {
		return "", err
	}

	return token, nil
}

type tokenLogEnt struct {
	ID          int32
	Channel     string
	UsedFor     string
	Via         string
	Username    string
	Token       string
	ExpDate     time.Time
	CreatedDate time.Time
}

// getLatestToken get latest token of user
func (c *Contract) getLatestToken(db *pgxpool.Conn, ctx context.Context, ch, usedFor, via, username string) (tokenLogEnt, error) {
	fmt.Println(ch, usedFor, via, username)
	var t tokenLogEnt
	err := pgxscan.Get(ctx, db, &t,
		`select * from token_logs 
		where channel=$1 and used_for=$2 and via=$3 and username=$4 order by id desc limit 1`,
		ch, usedFor, via, username,
	)

	return t, err
}

// ValidateToken ...
func (c *Contract) ValidateToken(db *pgxpool.Conn, ctx context.Context, ch, usedFor, username, token string) error {
	t, err := c.getLatestToken(db, ctx, ch, usedFor, c.Via(username), username)
	if err != nil {
		return err
	}

	now := time.Now().In(time.UTC)
	if now.After(t.ExpDate) {
		return fmt.Errorf("%s", "token expired")
	}

	if strings.Trim(t.Token, " ") != strings.Trim(token, " ") {
		return fmt.Errorf("%s", "invalid token")
	}

	return nil
}

// isValidPass
func (c *Contract) isValidPass(plain, enc string) bool {
	byteHash := []byte(enc)
	if err := bcrypt.CompareHashAndPassword(byteHash, []byte(plain)); err != nil {
		return false
	}

	return true
}

func (c *Contract) AuthLogin(db *pgxpool.Conn, ctx context.Context, ch string, username, pass string) (string, error) {
	var userID, userRole string
	switch ch {
	case ChannelCustApp:
		// from customer app we can login with 3 type of auth method:
		// username(@username), email(user@email.com), phone(+6288899999999)
		// need to check which type to follow
		m, err := c.GetMemberBy(db, ctx, c.getMemberField(username), username)
		if err != nil {
			return "", err
		}

		if !c.isValidPass(pass, m.Password) {
			return "", errInvalidCred
		}

		if !m.IsActive {
			return "", errInactiveUser
		}

		//check last active visit app
		s, i, err := c.GetLogVisitApp(db, ctx, m.ID)
		if err != nil && err == sql.ErrNoRows {
			fmt.Println(err)
			return "", err
		}
		// add log last visit
		if i <= 0 {
			err = c.AddNewLogVisitApp(db, ctx, m.ID)
			if err != nil {
				fmt.Println(err)
				return "", err
			}
		} else {
			// check if same date
			ck := s.Truncate(24 * time.Hour).Equal(time.Now().Truncate(24 * time.Hour))

			if !ck {
				err = c.UpdateTotalVisited(db, ctx, m.ID, i+1)
				if err != nil {
					fmt.Println(err)
					return "", err
				}
			}
		}
		// fmt.Println(i)
		// fmt.Println(prettytime.Format(s))

		userID = m.MemberCode
		userRole = "customer"
	case ChannelCMS:
		// from customer app we can login only with 1 type of auth method:
		// email(user@email.com)
		u, err := c.GetUserByEmail(db, ctx, username)
		if err != nil {
			return "", errInvalidCred
		}

		if !c.isValidPass(pass, u.Password) {
			return "", errInvalidCred
		}
		userID = strconv.Itoa(int(u.ID))
		userRole = u.Role
	default:
		return "", fmt.Errorf("%s", "invalid channel")
	}

	// TODO: need to save all user data to redis cache

	token, _, _ := c.generateJWT(ch, userID, userRole, c.Config.GetString("app.key"))
	return token, nil
}
