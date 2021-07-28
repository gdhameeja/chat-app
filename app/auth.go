package app

import (
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/stretchr/gomniauth"
	"github.com/stretchr/objx"
)

type authHandler struct {
	next http.Handler
}

func (h *authHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("auth")
	// val == "" because when logging out, we're setting the value in cookie to ""
	if err == http.ErrNoCookie || cookie.Value == "" {
		// not authenticated
		w.Header().Set("Location", "/login")
		w.WriteHeader(http.StatusTemporaryRedirect)
		return
	}
	if err != nil {
		// some other error
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// success - call the next handler
	h.next.ServeHTTP(w, r)
}

func MustAuth(handler http.Handler) http.Handler {
	return &authHandler{next: handler}
}

// LoginHandler handles the third-party login process.
// format: /auth/{action}/{provider}
func LoginHandler(wr http.ResponseWriter, req *http.Request) {
	segs := strings.Split(req.URL.Path, "/")
	action := segs[2]
	provider := segs[3]
	switch action {
	case "login":
		provider, err := gomniauth.Provider(provider)
		if err != nil {
			http.Error(wr, fmt.Sprintf("Error when trying to get provider %s:%s",
				provider, err), http.StatusBadRequest)
			return
		}

		loginUrl, err := provider.GetBeginAuthURL(nil, nil)
		if err != nil {
			http.Error(wr, fmt.Sprintf("Error when trying to GetBeginAuthURL %s:%s",
				provider, err), http.StatusInternalServerError)
			return
		}

		wr.Header().Set("Location", loginUrl)
		wr.WriteHeader(http.StatusTemporaryRedirect)
	case "callback":
		provider, err := gomniauth.Provider(provider)
		if err != nil {
			http.Error(wr, fmt.Sprintf("Error when trying to get provider %s:%s",
				provider, err), http.StatusBadRequest)
			return
		}

		creds, err := provider.CompleteAuth(objx.MustFromURLQuery(req.URL.RawQuery))
		if err != nil {
			http.Error(wr, fmt.Sprintf("Error when trying complete auth for %s:%s",
				provider, err), http.StatusBadRequest)
			return
		}

		user, err := provider.GetUser(creds)
		if err != nil {
			http.Error(wr, fmt.Sprintf("Error when trying to get user from %s:%s",
				provider, err), http.StatusInternalServerError)
		}

		m := md5.New()
		io.WriteString(m, strings.ToLower(user.Email()))
		userId := fmt.Sprintf("%x", m.Sum(nil))
		authCookieValue := objx.New(map[string]interface{}{
			"userId":     userId,
			"name":       user.Name(),
			"avatar_url": user.AvatarURL(),
			"email":      user.Email(),
		}).MustBase64()
		http.SetCookie(wr, &http.Cookie{
			Name:  "auth",
			Value: authCookieValue,
			Path:  "/",
		})
		wr.Header().Set("Location", "/chat")
		wr.WriteHeader(http.StatusTemporaryRedirect)
	default:
		wr.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(wr, "Auth action %s not supported", action)
	}
}
