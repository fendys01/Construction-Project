package bootstrap

import (
	"context"
	"fmt"
	"net/http"
	"panorama/lib/utils"
	"runtime/debug"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
)

// CustomClaims JWT custom claims
type CustomClaims struct {
	MemberCode string `json:"member_code"`
	Channel    string `json:"channel"`
	Role       string `json:"role"`
	jwt.StandardClaims
}

// RegisterClaims JWT custom claims
type RegisterClaims struct {
	Token string `json:"token_register"`
	jwt.StandardClaims
}

const (
	ChannelCustApp = "cust_mobile_app"
	ChannelCMS     = "cms"
	ChannelTCApp   = "tc_mobile_app"
)

var (
	mustHeader = []string{"X-Channel", "Content-Type"}
	headerVal  = []string{ChannelCustApp, ChannelCMS, ChannelTCApp, "application/json"}
)

func userContext(ctx context.Context, subject, id interface{}) context.Context {
	return context.WithValue(ctx, subject, id)
}

const pingReqURI string = "/v1/ping"

func isPingRequest(r *http.Request) bool {
	return r.RequestURI == pingReqURI
}

// NotfoundMiddleware A custom not found response.
func (app *App) NotfoundMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tctx := chi.NewRouteContext()
		rctx := chi.RouteContext(r.Context())

		if !rctx.Routes.Match(tctx, r.Method, r.URL.Path) {
			app.SendNotfound(w, "Sorry. We couldn't find that page")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Recoverer ...
func (app *App) Recoverer(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rvr := recover(); rvr != nil {
				logEntry := middleware.GetLogEntry(r)
				if logEntry != nil {
					logEntry.Panic(rvr, debug.Stack())
				} else {
					debug.PrintStack()
				}

				app.Log.FromDefault().WithFields(logrus.Fields{
					"Panic": rvr,
				}).Errorf("Panic: %v \n %v", rvr, string(debug.Stack()))

				app.SendBadRequest(w, "Something error with our system. Please contact our administrator")
				return
			}
		}()

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

// VerifyJwtToken ...
func (app *App) VerifyJwtToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := &CustomClaims{}
		tokenAuth := r.Header.Get("Authorization")
		_, err := jwt.ParseWithClaims(tokenAuth, claims, func(token *jwt.Token) (interface{}, error) {
			if jwt.SigningMethodHS256 != token.Method {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}

			secret := app.Config.GetString("app.key")
			return []byte(secret), nil
		})

		if err != nil {
			msg := "token is invalid"
			if mErr, ok := err.(*jwt.ValidationError); ok {
				if mErr.Errors == jwt.ValidationErrorExpired {
					msg = "token is expired"
				}
			}

			app.SendAuthError(w, msg)
			return
		}

		// need to check is token has a valid channel
		if !utils.Contains([]string{ChannelCMS, ChannelCustApp, ChannelTCApp}, claims.Channel) {
			app.SendBadRequest(w, "invalid token channel")
			return
		}

		// TODO: should check to redis/db is token expired or not

		ctx := userContext(r.Context(), "identifier", map[string]string{
			"mcode": claims.MemberCode,
			"role":  claims.Role,
		})

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// HeaderCheckerMiddleware check the necesarry headers
func (app *App) HeaderCheckerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, v := range mustHeader {
			if len(r.Header.Get(v)) == 0 || !utils.Contains(headerVal, r.Header.Get(v)) {
				app.SendBadRequest(w, fmt.Sprintf("undefined %s header or wrong value of header", v))
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

// CMSonly check that route only for cms/admin user
func (app *App) CMSonly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	})
}
