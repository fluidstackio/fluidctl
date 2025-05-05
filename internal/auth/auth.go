package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fluidstackio/fluidctl/internal/utils"
	"github.com/golang-jwt/jwt/v5"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/browser"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

var (
	defaultIssuer    = "https://fluidstack.us.auth0.com/"
	defaultAuthURL   = "https://fluidstack.us.auth0.com/authorize"
	defaultTokenURL  = "https://fluidstack.us.auth0.com/oauth/token"
	defaultClientID  = "diPhN35HH6jVXs615vsafkdIQM4Y5rF8"
	defaultAudience  = "https://api.fluidstack.io"
	defaultRedirect  = "http://localhost:5173"
	defaultTokenFile = "~/.fluidstack/token"
)

func isTokenExpired(tokenString string) (bool, error) {
	token, _, err := jwt.NewParser().ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return false, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if exp, ok := claims["exp"].(float64); ok {
			expirationTime := time.Unix(int64(exp), 0)
			if time.Now().After(expirationTime) {
				return true, nil
			} else {
				return false, nil
			}
		} else {
			return false, errors.New("token does not contain an expiration claim")
		}
	} else {
		return false, errors.New("invalid token claims")
	}
}

func readToken() (string, error) {
	tokenFile, err := homedir.Expand(defaultTokenFile)
	if err != nil {
		return "", fmt.Errorf("failed to expand token file path: %w", err)
	}

	token, err := os.ReadFile(tokenFile)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(token)), nil
}

func writeToken(token string) error {
	tokenFile, err := homedir.Expand(defaultTokenFile)
	if err != nil {
		return fmt.Errorf("failed to expand token file path: %w", err)
	}

	err = os.MkdirAll(filepath.Dir(tokenFile), 0700)
	if err != nil {
		return fmt.Errorf("failed to create token file directory: %w", err)
	}

	err = os.WriteFile(tokenFile, []byte(token), 0600)
	if err != nil {
		return fmt.Errorf("failed to write token file: %w", err)
	}

	return nil
}

func randomBytesInHex(count int) (string, error) {
	buf := make([]byte, count)
	_, err := io.ReadFull(rand.Reader, buf)
	if err != nil {
		return "", fmt.Errorf("Could not generate %d random bytes: %v", count, err)
	}

	return hex.EncodeToString(buf), nil
}

// GenerateCodeChallenge generates a code challenge from the code verifier.
func GenerateCodeChallenge(verifier string) string {
	hash := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}

func Login(cmd *cobra.Command) (string, error) {
	tokenString := utils.MustGetStringFlag(cmd, "token")
	if tokenString != "" {
		return tokenString, nil
	}

	tokenString, err := readToken()
	if err != nil {
		if !os.IsNotExist(err) {
			return "", fmt.Errorf("failed to read token file: %w", err)
		}
	} else if isExpired, err := isTokenExpired(tokenString); err == nil && !isExpired {
		return string(tokenString), nil
	}

	codeVerifier, verifierErr := randomBytesInHex(32)
	if verifierErr != nil {
		return "", fmt.Errorf("failed to create code verifier: %v", verifierErr)
	}

	hash := sha256.New()
	hash.Write([]byte(codeVerifier))
	codeChallenge := base64.RawURLEncoding.EncodeToString(hash.Sum(nil))

	config := &oauth2.Config{
		ClientID:    defaultClientID,
		RedirectURL: defaultRedirect,
		Scopes:      []string{"openid", "profile", "email", "offline_access"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  defaultAuthURL,
			TokenURL: defaultTokenURL,
		},
	}

	state, stateErr := randomBytesInHex(24)
	if stateErr != nil {
		return "", fmt.Errorf("failed to generate random state: %v", stateErr)
	}

	authURL := config.AuthCodeURL(
		state,
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		oauth2.SetAuthURLParam("code_challenge", codeChallenge),
		oauth2.SetAuthURLParam("audience", defaultAudience),
	)

	code, err := waitForAuthorizationCode(authURL)
	if err != nil {
		return "", err
	}

	token, err := config.Exchange(cmd.Context(), code, oauth2.SetAuthURLParam("code_verifier", codeVerifier))
	if err != nil {
		return "", fmt.Errorf("failed to exchange authorization code: %w", err)
	}

	if err := writeToken(token.AccessToken); err != nil {
		return "", fmt.Errorf("failed to save token: %w", err)
	}

	return token.AccessToken, nil
}

// waitForAuthorizationCode starts a local server to capture the authorization code.
func waitForAuthorizationCode(authURL string) (string, error) {
	server := &http.Server{Addr: ":5173"}
	defer server.Close()

	codeChan := make(chan string)
	http.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "Authorization code not found", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`
			<!DOCTYPE html>
			<html>
			<head>
				<title>Authorization Successful</title>
			</head>
			<body>
				<p>Authorization successful. You can close this window.</p>
				<script>
					window.close();
				</script>
			</body>
			</html>
		`))

		codeChan <- code
	})

	go server.ListenAndServe()

	browser.OpenURL(authURL)

	code := <-codeChan

	return code, nil
}
