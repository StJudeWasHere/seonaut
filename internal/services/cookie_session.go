package services

import (
	"context"
	"encoding/gob"
	"net/http"

	"github.com/stjudewashere/seonaut/internal/models"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
)

const (
	UserKey     contextKey = "user"
	SessionName string     = "SESSION_ID"
)

type (
	contextKey string

	CookieSessionRepository interface {
		FindUserByEmail(email string) (*models.User, error)
	}

	CookieSession struct {
		repository CookieSessionRepository
		cookie     *sessions.CookieStore
	}
)

func NewCookieSession(r CookieSessionRepository) *CookieSession {
	authKeyOne := securecookie.GenerateRandomKey(64)
	encryptionKeyOne := securecookie.GenerateRandomKey(32)

	cookie := sessions.NewCookieStore(
		authKeyOne,
		encryptionKeyOne,
	)

	cookie.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   60 * 15,
		HttpOnly: true,
	}

	gob.Register(models.User{})

	return &CookieSession{
		repository: r,
		cookie:     cookie,
	}
}

// requireAuth is a middleware function that wraps the provided handler function and enforces authentication.
// It checks if the user is authenticated based on the session data.
func (s *CookieSession) Auth(f func(w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := s.cookie.Get(r, SessionName)
		if err != nil {
			session.Options.MaxAge = -1
			session.Save(r, w)
			http.Redirect(w, r, "/signin", http.StatusSeeOther)
			return
		}

		authenticated, ok := session.Values["authenticated"].(bool)
		if !ok || !authenticated {
			session.Options.MaxAge = -1
			session.Save(r, w)
			http.Redirect(w, r, "/signin", http.StatusSeeOther)
			return
		}

		email, ok := session.Values["email"].(string)
		if !ok {
			session.Options.MaxAge = -1
			session.Save(r, w)
			http.Redirect(w, r, "/signin", http.StatusSeeOther)
			return
		}

		session.Options.MaxAge = 60 * 60 * 48
		if err := session.Save(r, w); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		user, err := s.repository.FindUserByEmail(email)
		if err != nil {
			http.Redirect(w, r, "/signin", http.StatusSeeOther)
			return
		}

		ctx := s.setUser(user, r.Context())
		req := r.WithContext(ctx)
		f(w, req)
	}
}

// SetUserToContext takes a User and a context as input and returns a new context with the given
// user value set.
func (s *CookieSession) setUser(user *models.User, c context.Context) context.Context {
	return context.WithValue(c, UserKey, user)
}

// GetUserFromContext takes a context as input and retrieves the associated User value from it, if present.
func (s *CookieSession) GetUser(c context.Context) (*models.User, bool) {
	v := c.Value(UserKey)
	user, ok := v.(*models.User)
	return user, ok
}

// Sets a user authentication session with the user Id.
func (s *CookieSession) SetSession(user *models.User, w http.ResponseWriter, r *http.Request) error {
	session, err := s.cookie.Get(r, SessionName)
	if err != nil {
		return err
	}

	session.Values["authenticated"] = true
	session.Values["email"] = user.Email
	return session.Save(r, w)
}

// Destroys a user authentication session to deauthenticate a user.
func (s *CookieSession) DestroySession(w http.ResponseWriter, r *http.Request) error {
	session, err := s.cookie.Get(r, SessionName)
	if err != nil {
		return err
	}

	session.Values["authenticated"] = false
	session.Values["email"] = nil
	session.Options.MaxAge = -1
	return session.Save(r, w)
}
