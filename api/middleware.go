package api

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/chuuch/go-banking/token"
	"github.com/gin-gonic/gin"
)

const (
	authorizationHeaderKey  = "authorization"
	authorizationTypeBearer = "bearer"
	authorizationPayloadKey = "authorization_payload"
)

func authMiddleware(tokenMaker token.Maker) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// We get the authorization header from the context
		authorizationHeader := ctx.GetHeader(authorizationHeaderKey)
		// we check if the authorization header is provided
		// if not we abort and send a JSON error response withg the status code
		if len(authorizationHeader) == 0 {
			err := errors.New("authorization header is not provided")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		// We split the authorization header by space
		fields := strings.Fields(authorizationHeader)
		// we check if there are at least 2 elements in our split header
		if len(fields) < 2 {
			err := errors.New("invalid authorization header format")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}
		// we get the authorization type from the header
		authorizationType := strings.ToLower(fields[0])
		// we compare the header authorization type
		if authorizationType != authorizationTypeBearer {
			// if not "bearer" we send a JSON error response with status unauthorized
			err := fmt.Errorf("unsupported authorization type %s", authorizationType)
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		// we get the access token from the 2nd part (1st index) of the header fields
		// and verify it with our token maker method
		accessToken := fields[1]
		payload, err := tokenMaker.VerifyToken(accessToken)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		ctx.Set(authorizationPayloadKey, payload)
		ctx.Next()
	}
}
