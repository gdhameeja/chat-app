package app

import (
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/stretchr/gomniauth"
	gomniauthcommon "github.com/stretchr/gomniauth/common"
	"github.com/stretchr/objx"
)

type ChatUser interface {
	UniqueID() string
	AvatarURL() string
}

type chatUser struct {
	// gomniauthcommon.User is an interface.
	// be embedding that interface here, `chatUser` automatically implements
	// gomniauthcommon.User interface. Hence the implementation of `AvatarURL()`
	// is provided by gomniauthcommon.User
	gomniauthcommon.User
	uniqueID string
}

func (u chatUser) UniqueID() string {
	return u.uniqueID
}

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
			http.Error(wr, fmt.Sprintf("Error when trying to complete auth for %s:%s",
				provider, err), http.StatusBadRequest)
			return
		}

		user, err := provider.GetUser(creds)
		if err != nil {
			http.Error(wr, fmt.Sprintf("Error when trying to get user from %s:%s",
				provider, err), http.StatusInternalServerError)
		}

		chatUser := &chatUser{User: user}
		m := md5.New()
		io.WriteString(m, strings.ToLower(user.Email()))

		chatUser.uniqueID = fmt.Sprintf("%x", m.Sum(nil))
		avatarURL, err := avatars.GetAvatarURL(chatUser)
		if err != nil {
			log.Fatalln("Error when trying to GetAvatarURL", "-", err)
		}

		authCookieValue := objx.New(map[string]interface{}{
			"userId":     chatUser.uniqueID,
			"name":       user.Name(),
			"avatar_url": avatarURL,
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
