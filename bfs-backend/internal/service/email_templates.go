package service

import _ "embed"

//go:embed templates/login_otp.html
var loginEmailHTML string

//go:embed templates/login_otp.txt
var loginEmailText string

//go:embed templates/email_change_otp.html
var emailChangeHTML string

//go:embed templates/email_change_otp.txt
var emailChangeText string

//go:embed templates/admin_invite.html
var adminInviteHTML string

//go:embed templates/admin_invite.txt
var adminInviteText string
