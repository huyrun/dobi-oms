package auth

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/google/uuid"

	"github.com/dgrijalva/jwt-go"

	"github.com/gin-gonic/gin"

	"github.com/pghuy/dobi-oms/pkg/http_response"
)

type header struct {
	Authorization string `header:"authorization"`
}

func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var err error

		defer func() {
			if err != nil {
				logrus.WithError(err).Error("Middleware")
			}
		}()

		jwtTokenKey := jwtTokenKey
		header := &header{}
		err = c.BindHeader(header)
		if err != nil {
			http_response.Abort(c, err)
			return
		}

		tokenString := strings.Replace(header.Authorization, "Bearer ", "", -1)

		if tokenString == "" {
			http_response.Abort(c, err)
			return
		}

		token, err := validateToken(tokenString, jwtTokenKey)
		if err != nil {
			http_response.Abort(c, err)
			return
		}
		if token == nil {
			http_response.Abort(c, err)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			logrus.Infof("Middleware, cast claims failed")
			http_response.Abort(c, err)
			return
		}

		if expiredTime, ok := claims["expired_time"]; ok {
			if t, ok := expiredTime.(float64); ok {
				if t-float64(time.Now().Unix()) < 0 {
					http_response.Abort(c, err)
					return
				}
			} else {
				logrus.Infof("Middleware, cast expired_time failed")
				http_response.Abort(c, err)
				return
			}
		}

		c = withUID(c, claims)
		c.Set(jwtTokenKey, token)

		c.Next()
	}
}

func validateToken(encodedToken string, secretJWT string) (*jwt.Token, error) {
	return jwt.Parse(encodedToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); ok {
			return []byte(secretJWT), nil
		}

		return nil, errors.New("Invalid token")
	})
}

func withUID(ctx *gin.Context, jwtClaims jwt.MapClaims) *gin.Context {
	uid := UID{}

	if userID, ok1 := jwtClaims["user_id"]; ok1 {
		uid.ID, _ = uuid.Parse(fmt.Sprintf("%s", userID))
	}

	if username, ok1 := jwtClaims["username"]; ok1 {
		if v, ok2 := username.(string); ok2 {
			uid.Username = v
		}
	}

	ctx.Set(uidKey, uid)

	return ctx
}
