package sendgrid

type MailSendRequest struct {
	To []MailToRequest `json:"to"`
}

type MailToRequest struct {
	Email string `json:"email"`
}

type MailContentRequest struct {
	Type  string `json:"type"`
	Value string `value:"value"`
}

type MailPersonalizations struct {
	Subject          string               `json:"subject"`
	From             MailToRequest        `json:"from"`
	Personalizations []MailSendRequest    `json:"personalizations"`
	Content          []MailContentRequest `json:"content"`
}
