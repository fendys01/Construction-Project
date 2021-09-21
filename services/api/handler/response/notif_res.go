package response

import (
	"fmt"
	"panorama/lib/utils"
	"panorama/services/api/model"
	"strconv"
	"time"

	"github.com/spf13/viper"
)

type NotifResponse struct {
	NotifCode        string    `json:"notif_code"`
	NotifSubject     string    `json:"notif_subject"`
	NotifType        int32     `json:"notif_type"`
	NotifTitle       string    `json:"notif_title"`
	NotifContent     string    `json:"notif_content"`
	NotifLink        string    `json:"notif_link"`
	NotifIsRead      bool      `json:"notif_is_read"`
	UserCode         string    `json:"user_code"`
	UserName         string    `json:"user_name"`
	UserEmail        string    `json:"user_email"`
	UserPhone        string    `json:"user_phone"`
	UserImg          string    `json:"user_img"`
	UserRole         string    `json:"user_role"`
	CreatedDate      time.Time `json:"created_date"`
	TimeElapsed      string    `json:"time_elapsed"`
	AdditionalTittle string    `json:"additional_title"`
}

// Transform from member model to member response
func (r NotifResponse) Transform(m model.NotificationEnt) NotifResponse {
	r.NotifCode = m.Code
	r.NotifSubject = m.Subject
	r.NotifType = m.Type
	r.NotifTitle = m.Title
	r.NotifTitle = m.Title
	r.NotifContent = m.Content
	r.NotifContent = m.Content
	r.NotifLink = m.Link.String
	r.NotifIsRead = m.IsRead
	r.UserCode = m.User.UserCode
	r.UserName = m.User.Name
	r.UserEmail = m.User.Email
	r.UserPhone = m.User.Phone
	r.UserRole = m.Role
	r.CreatedDate = m.CreatedDate
	r.TimeElapsed = utils.TimeElapsed(m.CreatedDate)

	var userImg string
	if len(m.User.Img.String) > 0 {
		if IsUrl(m.User.Img.String) {
			userImg = m.User.Img.String
		} else {
			userImg = viper.GetString("aws.s3.public_url") + m.User.Img.String
		}
	}
	r.UserImg = userImg

	var additionalTitle string
	if len(m.AdditionalTitle.String) > 0 {
		additionalTitle = m.AdditionalTitle.String
		if m.Type == model.NOTIF_TYPE_MBITIN && m.MemberItin.DayPeriod != 0 {
			dayPeriod := "(" + strconv.Itoa(int(m.MemberItin.DayPeriod)) + "D" + strconv.Itoa(int(m.MemberItin.DayPeriod-1)) + "N)"
			additionalTitle = fmt.Sprintf("%s:  %s %s", m.TypeText, m.AdditionalTitle.String, dayPeriod)
		}
	}
	r.AdditionalTittle = additionalTitle

	return r
}
