package middlewares

import (
	"InternalAssetManagement/database/dbhelper"
	"InternalAssetManagement/handler"
	"InternalAssetManagement/models"
	"InternalAssetManagement/utils"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-chi/chi/v5"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")

		claims := models.Claims{}

		tkn, err1 := jwt.ParseWithClaims(token, &claims, func(t *jwt.Token) (interface{}, error) {
			return handler.JwtKey, nil
		})
		if err1 != nil {
			if err1 == jwt.ErrSignatureInvalid {
				utils.RespondError(w, http.StatusUnauthorized, err1, "AuthMiddleware: Signature invalid.")
				return
			}
			utils.RespondError(w, http.StatusUnauthorized, err1, "AuthMiddleware: ParseErr.")
			return
		}

		if !tkn.Valid {
			w.WriteHeader(http.StatusUnauthorized)
			logrus.Printf("token is invalid")
			return
		}

		_, err := dbhelper.CheckSession(claims.ID)
		if err != nil {
			logrus.Printf("session expired:%v", err)
			return
		}
		userID := claims.ID

		// value := models.ContextValues{ID: userID}
		ctx := context.WithValue(r.Context(), utils.UserContextKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

var MaxAge = 300

// corsOptions setting up routes for cors
func corsOptions() *cors.Cors {
	return cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "Access-Token", "importDate", "X-Client-Version", "Cache-Control", "Pragma", "x-started-at", "x-api-key", "token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           MaxAge,
	})
}

// CommonMiddlewares middleware common for all routes
func CommonMiddlewares() chi.Middlewares {
	return chi.Chain(
		corsOptions().Handler,
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Add("Content-Type", "application/json")
				next.ServeHTTP(w, r)
			})
		},
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer func() {
					err := recover()
					if err != nil {
						logrus.Errorf("Request Panic err: %v", err)
						jsonBody, _ := json.Marshal(map[string]string{
							"error": "There was an internal server error",
							"trace": fmt.Sprintf("%+v", err),
							"stack": string(debug.Stack()),
						})
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusInternalServerError)
						_, err := w.Write(jsonBody)
						if err != nil {
							logrus.Errorf("Failed to send response from middleware with error: %+v", err)
						}
					}
				}()
				next.ServeHTTP(w, r)
			})
		},
	)
}
