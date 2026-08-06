package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"pos_bos/internal/core/domain"
	"pos_bos/pkg/response"

	"github.com/go-chi/chi/v5"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var googleOauthConfig *oauth2.Config
var oauthStateString = "random_state_string_for_security"

func init() {
	// Initialize oauth config when needed or in main, but let's do it lazy or dynamically based on env
}

func getGoogleOauthConfig() *oauth2.Config {
	if googleOauthConfig == nil {
		googleOauthConfig = &oauth2.Config{
			RedirectURL:  os.Getenv("OAUTH_REDIRECT_URL"),
			ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
			ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
			Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
			Endpoint:     google.Endpoint,
		}
	}
	return googleOauthConfig
}

func (handler *AuthHandler) RegisterOAuthRoutes(router chi.Router) {
	router.Get("/auth/google/login", handler.GoogleLogin)
	router.Get("/auth/google/callback", handler.GoogleCallback)
}

func (handler *AuthHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	url := getGoogleOauthConfig().AuthCodeURL(oauthStateString)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (handler *AuthHandler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	state := r.FormValue("state")
	if state != oauthStateString {
		response.Error(w, http.StatusBadRequest, "Invalid OAuth state")
		return
	}

	code := r.FormValue("code")
	token, err := getGoogleOauthConfig().Exchange(context.Background(), code)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Code exchange failed")
		return
	}

	resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to get user info")
		return
	}
	defer resp.Body.Close()

	var userInfo struct {
		ID    string `json:"id"`
		Email string `json:"email"`
		Name  string `json:"name"`
		Photo string `json:"picture"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to decode user info")
		return
	}

	// Login or create user
	loginReq := domain.OAuthLoginRequest{
		Email:      userInfo.Email,
		Name:       userInfo.Name,
		Photo:      userInfo.Photo,
		Provider:   "google",
		ProviderID: userInfo.ID,
	}

	loginRes, err := handler.service.OAuthLogin(r.Context(), loginReq)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "OAuth Login failed: "+err.Error())
		return
	}

	// Set HttpOnly Cookies
	http.SetCookie(w, &http.Cookie{
		Domain:   getCookieDomain(),
		Name:     "access_token",
		Value:    loginRes.Token,
		Expires:  time.Now().Add(15 * time.Minute),
		HttpOnly: true,
		Secure:   os.Getenv("APP_ENV") == "production",
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})

	http.SetCookie(w, &http.Cookie{
		Domain:   getCookieDomain(),
		Name:     "refresh_token",
		Value:    loginRes.RefreshToken,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
		HttpOnly: true,
		Secure:   os.Getenv("APP_ENV") == "production",
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})

	http.SetCookie(w, &http.Cookie{
		Domain:   getCookieDomain(),
		Name:     "is_logged_in",
		Value:    "1",
		Expires:  time.Now().Add(7 * 24 * time.Hour),
		HttpOnly: false,
		Secure:   os.Getenv("APP_ENV") == "production",
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})

	// Redirect back to frontend
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		if os.Getenv("APP_ENV") == "production" {
			frontendURL = "https://posbos.gofarputraperdana.my.id"
		} else {
			frontendURL = "http://localhost:3000"
		}
	}
	http.Redirect(w, r, frontendURL, http.StatusTemporaryRedirect)
}
