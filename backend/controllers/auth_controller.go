package controllers

import (
	"net/http"

	"backend/services"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

type AuthController struct {
	svc *services.AuthService
}

func NewAuthController(svc *services.AuthService) *AuthController {
	return &AuthController{svc: svc}
}

// GET /login
func (a *AuthController) Login(c echo.Context) error {
	state := "anything-random" // ideally a crypto-rand string, omitted for brevity
	return c.Redirect(http.StatusTemporaryRedirect, a.svc.LoginURL(state))
}

// GET /callback
func (a *AuthController) Callback(c echo.Context) error {
	if errMsg := c.QueryParam("error"); errMsg != "" {
		return c.String(http.StatusBadRequest, errMsg)
	}

	code := c.QueryParam("code")
	if code == "" {
		return c.String(http.StatusBadRequest, "missing code")
	}

	claims, err := a.svc.HandleCallback(c.Request().Context(), code)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	sess, _ := session.Get("auth-session", c)
	sess.Values["profile"] = claims
	sess.Save(c.Request(), c.Response())

	return c.Redirect(http.StatusSeeOther, "/user")
}

// GET /user
func (a *AuthController) User(c echo.Context) error {
	sess, _ := session.Get("auth-session", c)
	if profile, ok := sess.Values["profile"]; ok {
		return c.JSON(http.StatusOK, profile)
	}
	return c.NoContent(http.StatusUnauthorized)
}

// GET /logout
func (a *AuthController) Logout(c echo.Context) error {
	sess, _ := session.Get("auth-session", c)
	sess.Options.MaxAge = -1 // delete cookie
	sess.Save(c.Request(), c.Response())

	return c.Redirect(http.StatusSeeOther, a.svc.LogoutURL("/"))
}
