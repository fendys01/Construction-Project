package model

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"panorama/bootstrap"
	"panorama/lib/citcall"
	"panorama/lib/sendgrid"
	"panorama/lib/utils"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	mail "github.com/xhit/go-simple-mail/v2"
	"golang.org/x/crypto/bcrypt"
)

const (
	ActForgotPass            = "forgot_pass"
	ActChangeEmail           = "change_email"
	ActRegPhone              = "reg_phone"
	ActRegEmail              = "reg_email"
	ActChangePass            = "change_pass"
	ActChangePhone           = "change_phone"
	ActInviteGroupItinMember = "invite_group_itin_member"

	TokenViaEmail = "email"
	TokenViaPhone = "phone"

	tokenExpiredMin = 2 // in minutes
	urlCms          = "https://dev-panorama-cms.rebelworks.co"
)

var tokenMailSubj = map[string]string{
	ActForgotPass:  "[Panorama] Forgot Password",
	ActRegEmail:    "[Panorama] Email Verification",
	ActChangeEmail: "[Panorama] Change Email Verification",
	ActChangePass:  "[Panorama] Change Password Verification",
}

var urlCMSLink = map[string]string{
	ActForgotPass: "forgotpassword",
	ActRegEmail:   "verify",
}

var errInvalidCred error = fmt.Errorf("%s", "invalid user credential")
var errInactiveUser error = fmt.Errorf("%s", "inactive user")

type DataEmailInviteItinMember struct {
	Sender        string
	URL           string
	ItineraryName string
	EmailInvite   string
}

type DataEmailToken struct {
	TokenURL     string
	ExpiredTime  int
	Description  string
	IsChannelApp bool
	Title        string
}

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

func (c *Contract) getMemberField(username string, isMember bool) string {
	var field string = "phone"
	_, _, validPhone := utils.IsPhone(username)

	if validPhone {
		field = "phone"
	} else if utils.IsEmail(username) {
		field = "email"
	} else if utils.IsUsernameTag(username) && isMember {
		field = "username"
	} else if utils.IsCodeTag(username) && isMember {
		field = "member_code"
	} else if utils.IsCodeTag(username) && !isMember {
		field = "user_code"
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

func (c *Contract) sendDataMail(usedFor, subject, to string, dataMail interface{}) {
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
	tpl, err := utils.ParseTpl(fn, dataMail)
	if err != nil {
		fmt.Println(err.Error())
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
func (c *Contract) SendToken(db *pgxpool.Conn, ctx context.Context, ch, usedFor, via, username, role, tokenParam string) (string, error) {
	if !c.isValidTokenAction(ch, via, username) {
		return "", fmt.Errorf("%s", "send token invalid action")
	}

	if !c.isUsernameExists(db, ctx, ch, username) {
		return "", fmt.Errorf("user doesn't exists")
	}

	var token string
	dataMail := DataEmailToken{
		ExpiredTime: tokenExpiredMin,
	}

	switch ch {
	case ChannelCustApp:
		newToken, err := c.addNewToken(db, ctx, ch, usedFor, via, username, tokenParam)
		if err != nil {
			return "", err
		}

		token = newToken
		dataMail.Title = "OTP"
		dataMail.TokenURL = token
		dataMail.IsChannelApp = true
		dataMail.Description = "Please input the 4 digit code"

		if via == TokenViaEmail {
			go c.sendDataMail(usedFor, tokenMailSubj[usedFor], username, dataMail)
		}

		if via == TokenViaPhone {
			// Send SMS with token
			sms, err := citcall.New(c.App).SendOTP(username, token, tokenExpiredMin)
			if err != nil {
				return "", err
			}
			if sms.RC != citcall.STATUS_OK {
				return "", fmt.Errorf("citcall response = code: %d, description: %s", sms.RC, sms.Info)
			}
		}

	case ChannelCMS:
		tokenJWT, _, _ := c.generateJWT(ch, username, role, c.Config.GetString("app.key"))
		token = tokenJWT

		dataMail.Title = "Link"
		dataMail.IsChannelApp = false
		dataMail.TokenURL = fmt.Sprintf("%s/%s?token=%s", urlCms, urlCMSLink[usedFor], tokenJWT)
		dataMail.Description = "Please click the link"

		if via == TokenViaEmail {
			go c.sendDataMail(usedFor, tokenMailSubj[usedFor], username, dataMail)
		}
	}

	return token, nil
}

func (c *Contract) isUsernameExists(db *pgxpool.Conn, ctx context.Context, ch, username string) bool {
	switch ch {
	case ChannelCustApp:
		return c.isMemberExists(db, ctx, c.getMemberField(username, true), username)

	case ChannelCMS:
		return c.isUserExists(db, ctx, username)
	}

	return false
}

// addNewToken ...
func (c *Contract) addNewToken(db *pgxpool.Conn, ctx context.Context, ch, usedFor, via, username, tokenParam string) (string, error) {
	token := tokenParam
	if tokenParam == "" {
		rand.Seed(time.Now().UnixNano())
		token, _ = utils.Generate(`[\d]{4}`)
	}

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
	var t tokenLogEnt

	sql := `select token, exp_date from token_logs where channel=$1 and used_for=$2 and via=$3 and username=$4 order by id desc limit 1`
	err := db.QueryRow(ctx, sql, ch, usedFor, via, username).Scan(&t.Token, &t.ExpDate)

	return t, err
}

// ValidateToken ...
func (c *Contract) ValidateToken(db *pgxpool.Conn, ctx context.Context, ch, usedFor, username, token string) error {
	via := c.Via(username)
	if via == "" {
		return fmt.Errorf("%s", "email / phone requests is invalid.")
	}

	t, err := c.getLatestToken(db, ctx, ch, usedFor, via, username)
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

func (c *Contract) AuthLogin(db *pgxpool.Conn, ctx context.Context, ch string, username, pass string) (map[string]interface{}, error) {
	var userID int32
	var userCode, userRole, userName, userPhone, userEmail string
	var userStatus bool
	switch ch {
	case ChannelCustApp:
		// from customer app we can login with 3 type of auth method:
		// username(@username), email(user@email.com), phone(+6288899999999)
		// need to check which type to follow
		m, err := c.GetMemberBy(db, ctx, c.getMemberField(username, true), username)
		if err != nil {
			return nil, err
		}

		if !c.isValidPass(pass, m.Password) {
			return nil, errInvalidCred
		}

		// if !m.IsActive {
		// 	return nil, errInactiveUser
		// }

		//check last active visit app
		id, date, i, err := c.GetLogVisitApp(db, ctx, m.ID, "customer")
		if err != nil && err == sql.ErrNoRows {
			fmt.Println(err)
			return nil, err
		}

		// add log last visit
		if id > 0 {
			if date.IsZero() || DateEqual(date, time.Now()) {
				err = c.UpdateTotalVisited(db, ctx, m.ID, i+1, "customer", id)
				if err != nil {
					fmt.Println(err)
					return nil, err
				}
			} else {

				err = c.AddLogVisitApp(db, ctx, m.ID, "customer")
				if err != nil {
					fmt.Println(err)
					return nil, err
				}
			}
		}

		userID = m.ID
		userCode = m.MemberCode
		userRole = "customer"
		userName = m.Name
		userStatus = m.IsActive
		userPhone = m.Phone
		userEmail = m.Email

	case ChannelCMS:
		// from customer app we can login only with 1 type of auth method:
		// email(user@email.com)
		u, err := c.GetUserByEmail(db, ctx, username)
		if err != nil {
			return nil, errInvalidCred
		}

		if !c.isValidPass(pass, u.Password) {
			return nil, errInvalidCred
		}

		//check last active visit app
		id, date, i, err := c.GetLogVisitApp(db, ctx, u.ID, u.Role)
		if err != nil && err == sql.ErrNoRows {
			fmt.Println(err)
			return nil, err
		}

		// add log last visit
		if id > 0 {
			if date.IsZero() || DateEqual(date, time.Now()) {
				err = c.UpdateTotalVisited(db, ctx, u.ID, i+1, u.Role, id)
				if err != nil {
					fmt.Println(err)
					return nil, err
				}
			} else {

				err = c.AddLogVisitApp(db, ctx, u.ID, u.Role)
				if err != nil {
					fmt.Println(err)
					return nil, err
				}
			}
		}

		userID = u.ID
		userCode = u.UserCode
		userRole = u.Role
		userName = u.Name
		userStatus = u.IsActive
		userPhone = u.Phone
		userEmail = u.Email
	default:
		return nil, fmt.Errorf("%s", "invalid channel")
	}

	// TODO: need to save all user data to redis cache
	token, _, _ := c.generateJWT(ch, userCode, userRole, c.Config.GetString("app.key"))
	result := map[string]interface{}{
		"token":       token,
		"user_id":     userID,
		"user_code":   userCode,
		"user_role":   userRole,
		"user_name":   userName,
		"user_status": userStatus,
		"user_phone":  userPhone,
		"user_email":  userEmail,
	}

	return result, nil
}

// SendingMail sending email into email to with data mail
func (c *Contract) SendingMail(usedFor, subject, emailTo string, dataMail interface{}) error {
	if utils.IsEmail(emailTo) {
		go c.sendDataMail(usedFor, subject, emailTo, dataMail)

		return nil
	}

	return nil
}

// check date
func DateEqual(date1, date2 time.Time) bool {
	y1, m1, d1 := date1.Date()
	y2, m2, d2 := date2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

// UpdatePassword update password all roles ...
func (c *Contract) UpdatePassword(tx pgx.Tx, ctx context.Context, ch, username, newPass string) error {
	pass, err := bcrypt.GenerateFromPassword([]byte(newPass), 10)
	if err != nil {
		return err
	}

	isMember := false
	tableName := "users"
	if ch == ChannelCustApp {
		isMember = true
		tableName = "members"
	}

	sql := fmt.Sprintf("update %s set password=$1, updated_date=$2 where %s=$3", tableName, c.getMemberField(username, isMember))
	exec, err := tx.Exec(ctx, sql, string(pass), time.Now().In(time.UTC), username)
	if exec.RowsAffected() == 0 {
		return fmt.Errorf("update password %s failed", tableName)
	}

	return err
}

func (c *Contract) SendingMailWSG(usedFor, subject, emailTo string, dataMail interface{}) error {
	var err error

	if utils.IsEmail(emailTo) {
		err := c.sendDataMailWSG(usedFor, subject, emailTo, dataMail)
		if err != nil {
			return err
		}
	}

	return err
}

func (c *Contract) sendDataMailWSG(usedFor, subject, to string, dataMail interface{}) error {
	var err error

	// Parsing data into html
	fn := fmt.Sprintf("%s/%s.html", c.Config.GetString("resource_path"), usedFor)
	template, err := utils.ParseTpl(fn, dataMail)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	// Send mail with sendgrid
	_, err = sendgrid.New(c.App).MailSender(subject, to, template)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	log.Println("Email Sent")

	return err
}
