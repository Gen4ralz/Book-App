package main

import "net/http"

func (app *application) AuthTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		_, err := app.models.Token.AuthenticateToken(req)
		if err != nil {
			payload := jsonResponse{
				Error: true,
				Message: "invalid authentication credentials",
			}

			_ = app.writeJSON(res, http.StatusUnauthorized, payload)
			return
		}
		next.ServeHTTP(res, req)
	})
}