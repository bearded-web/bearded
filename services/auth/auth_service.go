package auth

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/asaskevich/govalidator"
	"github.com/emicklei/go-restful"
	"github.com/facebookgo/stackerr"

	"github.com/bearded-web/bearded/models/user"
	"github.com/bearded-web/bearded/pkg/email"
	"github.com/bearded-web/bearded/pkg/filters"
	"github.com/bearded-web/bearded/pkg/passlib/reset"
	"github.com/bearded-web/bearded/pkg/validate"
	"github.com/bearded-web/bearded/services"
)

type AuthService struct {
	*services.BaseService
}

func New(base *services.BaseService) *AuthService {
	return &AuthService{
		BaseService: base,
	}
}

func addDefaults(r *restful.RouteBuilder) {
	r.Do(services.ReturnsE(
		http.StatusUnauthorized,
		http.StatusInternalServerError,
	))
}

func (s *AuthService) Register(container *restful.Container) {
	authRequired := filters.AuthRequiredFilter(s.BaseManager())

	ws := &restful.WebService{}
	ws.Path("/api/v1/auth")
	ws.Doc("Authorization management")
	ws.Consumes(restful.MIME_JSON)
	ws.Produces(restful.MIME_JSON)

	r := ws.POST("").To(s.login)
	r.Doc("login")
	r.Operation("login")
	r.Reads(authEntity{})
	r.Returns(http.StatusCreated, "Session created", sessionEntity{})
	r.Do(services.ReturnsE(http.StatusBadRequest))
	addDefaults(r)
	ws.Route(r)

	r = ws.DELETE("").To(s.logout)
	r.Doc("logout")
	r.Operation("logout")
	r.Filter(authRequired)
	r.Do(services.Returns(http.StatusNoContent))
	addDefaults(r)
	ws.Route(r)

	r = ws.GET("").To(s.status)
	r.Doc("status")
	r.Operation("status")
	r.Notes("Returns 200 ok if user is authenticated")
	r.Filter(authRequired)
	r.Do(services.Returns(http.StatusOK))
	addDefaults(r)
	ws.Route(r)

	// registration actions
	r = ws.POST("reset-password").To(s.resetPassword)
	r.Doc("reset user password")
	r.Operation("resetPassword")
	r.Reads(resetPasswordEntity{})
	r.Returns(http.StatusCreated, "Token created", "")
	r.Do(services.ReturnsE(http.StatusBadRequest))
	addDefaults(r)
	ws.Route(r)

	// registration actions
	r = ws.GET("reset-password").To(s.checkResetToken)
	r.Param(ws.QueryParameter("token", "token for password reset"))
	r.Doc("check reset token")
	r.Operation("checkResetToken")
	r.Returns(http.StatusTemporaryRedirect, "Status", "")
	addDefaults(r)
	ws.Route(r)

	// registration actions
	r = ws.POST("register").To(s.register)
	r.Doc("register")
	r.Operation("register")
	r.Reads(registerEntity{})
	r.Returns(http.StatusCreated, "User registered", sessionEntity{})
	r.Do(services.ReturnsE(http.StatusBadRequest))
	addDefaults(r)
	ws.Route(r)

	container.Add(ws)
}

func (s *AuthService) login(req *restful.Request, resp *restful.Response) {
	session := filters.GetSession(req)

	raw := &authEntity{}

	if err := req.ReadEntity(raw); err != nil {
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusBadRequest, services.WrongEntityErr)
		return
	}
	// TODO (m0sth8): extract user login and logout, this helps to login in other services
	// TODO (m0sth8): validate password and email, type, max length etc
	if raw.Email == "" {
		resp.WriteServiceError(http.StatusBadRequest, services.NewBadReq("password shouldn't be empty"))
		return
	}
	if raw.Password == "" {
		resp.WriteServiceError(http.StatusBadRequest, services.NewBadReq("password shouldn't be empty"))
		return
	}

	mgr := s.Manager()
	defer mgr.Close()

	// get user
	u, err := mgr.Users.GetByEmail(raw.Email)
	if err != nil {
		if mgr.IsNotFound(err) {
			// TODO (m0sth8): add captcha to protect against bruteforce
			resp.WriteServiceError(http.StatusUnauthorized, services.AuthFailedErr)
			return
		}
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusInternalServerError, services.DbErr)
		return
	}
	// users without password can't login
	if u.Password == "" {
		resp.WriteServiceError(http.StatusUnauthorized, services.AuthFailedErr)
		return
	}

	// verify password
	verified, err := s.PassCtx().Verify(raw.Password, u.Password)
	if err != nil {
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusInternalServerError, services.AppErr)
		return
	}
	if !verified {
		resp.WriteServiceError(http.StatusUnauthorized, services.AuthFailedErr)
		return
	}

	// TODO (m0sth8): extract auth methods, like login or logout.
	// set user id to session
	session.Set(filters.SessionUserKey, u.Id.Hex())
	resp.WriteHeader(http.StatusCreated)
	resp.WriteEntity(sessionEntity{Token: "not ready"})
}

func (s *AuthService) status(_ *restful.Request, _ *restful.Response) {
	// do nothing, just return 200 ok, cause authorization was checked in filter
}

func (s *AuthService) logout(req *restful.Request, resp *restful.Response) {
	session := filters.GetSession(req)
	session.Del(filters.SessionUserKey)
	resp.WriteHeader(http.StatusNoContent)
}

func (s *AuthService) register(req *restful.Request, resp *restful.Response) {
	session := filters.GetSession(req)

	raw := &registerEntity{}

	if err := req.ReadEntity(raw); err != nil {
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusBadRequest, services.WrongEntityErr)
		return
	}

	// TODO (m0sth8): add captcha support

	// check email
	if valid, err := govalidator.ValidateStruct(raw); !valid {
		resp.WriteServiceError(http.StatusBadRequest, services.NewBadReq(err.Error()))
		return
	}
	// check password
	if valid, reason := validate.Password(raw.Password); !valid {
		resp.WriteServiceError(http.StatusBadRequest, services.NewBadReq("Password %s", reason))
		return
	}

	mgr := s.Manager()
	defer mgr.Close()

	pass, err := s.PassCtx().Encrypt(raw.Password)
	if err != nil {
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusInternalServerError, services.AppErr)
		return
	}

	u := &user.User{
		Email:    raw.Email,
		Password: pass,
	}

	u, err = mgr.Users.Create(u)
	if err != nil {
		if mgr.IsDup(err) {
			resp.WriteServiceError(
				http.StatusConflict,
				services.NewError(services.CodeDuplicate, "user with this email is existed"))
			return
		}
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusInternalServerError, services.DbErr)
		return
	}

	session.Set(filters.SessionUserKey, u.Id.Hex())
	resp.WriteHeader(http.StatusCreated)
	resp.WriteEntity(sessionEntity{Token: "not ready"})
}

func (s *AuthService) resetPassword(req *restful.Request, resp *restful.Response) {

	raw := &resetPasswordEntity{}

	if err := req.ReadEntity(raw); err != nil {
		logrus.Warn(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusBadRequest, services.WrongEntityErr)
		return
	}

	if ok, err := govalidator.ValidateStruct(raw); !ok {
		resp.WriteServiceError(http.StatusBadRequest, services.NewBadReq(err.Error()))
		return
	}

	mgr := s.Manager()
	defer mgr.Close()

	// TODO (m0sth8): add captcha support
	u, err := mgr.Users.GetByEmail(raw.Email)
	if err != nil {
		if mgr.IsNotFound(err) {
			resp.WriteServiceError(http.StatusBadRequest, services.NewBadReq("Email is not found"))
			return
		}
		logrus.Error(stackerr.Wrap(err))
		resp.WriteServiceError(http.StatusInternalServerError, services.DbErr)
		return
	}

	reqUrl := req.Request.URL
	// TODO (m0sth8): send email in worker
	go func() {
		cfg := s.ApiCfg()
		dur := time.Second * time.Duration(cfg.ResetPasswordDuration)
		token := reset.NewToken(u.Email, dur, []byte(u.Password), []byte(cfg.ResetPasswordSecret))
		msg := email.NewMessage()
		msg.SetHeader("From", cfg.SystemEmail)
		msg.SetHeader("To", u.Email, u.Email)
		msg.SetHeader("Subject", "Reset password in bearded-web service")
		reqUrlVal := reqUrl.Query()
		reqUrlVal.Add("token", token)
		reqUrl.RawQuery = reqUrlVal.Encode()
		msg.SetBody("text/html", fmt.Sprintf("Please go to url: %s%s", cfg.Host, reqUrl.String()))
		if err := s.Mailer().Send(msg); err != nil {
			logrus.Error(err)
		}
	}()

	resp.ResponseWriter.WriteHeader(http.StatusCreated)

}

func (s *AuthService) checkResetToken(req *restful.Request, resp *restful.Response) {
	token := req.QueryParameter("token")

	mgr := s.Manager()
	defer mgr.Close()

	// TODO (m0sth8: take variables from config

	redirectErr := func(req *restful.Request, resp *restful.Response, errMsg string) {
		http.Redirect(resp.ResponseWriter, req.Request, fmt.Sprintf("/#/reset-end?error=%s", errMsg), http.StatusTemporaryRedirect)
	}

	redirect := func(req *restful.Request, resp *restful.Response, token string) {
		http.Redirect(resp.ResponseWriter, req.Request, fmt.Sprintf("/#/reset-end?token=%s", token), http.StatusTemporaryRedirect)
	}
	var u *user.User
	getUser := func(email string) ([]byte, error) {
		var err error
		u, err = mgr.Users.GetByEmail(email)
		if err != nil {
			return nil, err
		}
		return []byte(u.Password), err
	}

	cfg := s.ApiCfg()

	_, err := reset.VerifyToken(token, getUser, []byte(cfg.ResetPasswordSecret))
	if err != nil {
		if err == reset.ErrExpiredToken {
			redirectErr(req, resp, "Token expired, try again")
		}
		redirectErr(req, resp, "Wrong token, try again")
		return
	}
	if u == nil {
		resp.WriteServiceError(http.StatusInternalServerError, services.AppErr)
		return
	}
	// is that a good way to login user here?
	session := filters.GetSession(req)
	session.Set(filters.SessionUserKey, u.Id.Hex())

	redirect(req, resp, token)
}
