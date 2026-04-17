package httputil

import (
	"net/http"
)

func SetCookie(w http.ResponseWriter, name, value string, maxAge int, httpOnly bool) {
	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		MaxAge:   maxAge,
		Path:     "/",
		HttpOnly: httpOnly,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, cookie)
}

func ClearCookie(w http.ResponseWriter, name string) {
	SetCookie(w, name, "", -1, true)
}
