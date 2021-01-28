package echoadapter

import (
	"net"
	"strings"

	"bitbucket.org/calmisland/go-server-logs/logger"
	"bitbucket.org/calmisland/go-server-requests/apierrors"
	"bitbucket.org/calmisland/go-server-requests/sessions"
	"bitbucket.org/calmisland/go-server-requests/tokens/accesstokens"
	"github.com/calmisland/go-errors"
	"github.com/labstack/echo/v4"
)

const (
	authorizationHeaderName = "Authorization"
	accessTokenHeaderName   = "X-Access-Token"

	bearerAuthorizationType = "Bearer"
)

type AuthContext struct {
	echo.Context
	Session *sessions.Session
}

// AuthMiddleware is a request middleware that validates the session provided in the request.
func AuthMiddleware(validator accesstokens.Validator, rejectRequest bool) echo.MiddlewareFunc {
	if validator == nil {
		panic(errors.New("The validator cannot be nil"))
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			cc := &AuthContext{c, nil}

			req := c.Request()
			clientIP := net.ParseIP(c.RealIP())
			clientUserAgent := req.UserAgent()
			accessToken, hasAccessToken := getAccessTokenFromRequest(c)
			if !hasAccessToken || len(accessToken) == 0 {
				if rejectRequest {
					// We reject the request since it doesn't have an access token
					logger.LogFormat("[SECURITY] A request that requires authentication lacked an access token from IP [%s] and UserAgent [%s]\n", clientIP, clientUserAgent)
					return SetClientError(c, apierrors.ErrorUnauthorized)
				}

				// Continue if we don't reject, since there is no session here
				return next(c)
			}

			sessionData, err := validator.ValidateAccessToken(accessToken)
			if err != nil {
				logger.LogFormat("[SECURITY] A request with from IP [%s] and UserAgent [%s] failed access token validation: %s\n", clientIP, clientUserAgent, err.Error())

				// If we don't want to reject the request, we just continue directly
				if !rejectRequest {
					return next(c)
				}

				// Return different errors based on the validation error
				switch err.(type) {
				case *accesstokens.InvalidFormatError:
					err = SetClientError(c, apierrors.ErrorUnauthorized)
				case *accesstokens.InvalidTokenTypeError:
					err = SetClientError(c, apierrors.ErrorUnauthorized)
				case *accesstokens.InvalidHashAlgError:
					err = SetClientError(c, apierrors.ErrorUnauthorized)
				case *accesstokens.InvalidSignatureError:
					err = SetClientError(c, apierrors.ErrorUnauthorized)
				case *accesstokens.ExpiredTokenError:
					err = SetClientError(c, apierrors.ErrorExpiredAccessToken)
				case *accesstokens.SuspiciousTokenError:
					err = SetClientError(c, apierrors.ErrorUnauthorized)
				default:
				}

				return err
			} else if sessionData == nil {
				logger.LogFormat("[SECURITY] A request with a valid access token was missing session data from IP [%s] and UserAgent [%s]\n", clientIP, clientUserAgent)

				// If we don't want to reject the request, we just continue directly
				if !rejectRequest {
					return next(c)
				}

				return errors.New("Session data is missing from a seemingly valid access token")
			}

			// Create the session struct
			session := &sessions.Session{}

			// Assign the session data, and go to the next middleware
			session.ID = sessionData.SessionID
			session.Data = sessionData

			cc.Session = session
			return next(cc)
		}
	}
}

func getAccessTokenFromRequest(c echo.Context) (string, bool) {
	// Check the Authorization HTTP header

	authValue := c.Request().Header.Get(authorizationHeaderName)
	if len(authValue) > 0 {
		authValues := strings.SplitN(authValue, " ", 2)

		if len(authValues) == 2 {
			authType := authValues[0]
			switch authType {
			case bearerAuthorizationType:
				// Bearer authorization contains the access token
				accessToken := authValues[1]
				return accessToken, true
			}
		}
	}

	// As a fallback, use the previous header for the access tokens for backwards compatibility
	accessToken := c.Request().Header.Get(accessTokenHeaderName)
	if len(accessToken) > 0 {
		return accessToken, true
	}

	return "", false
}
