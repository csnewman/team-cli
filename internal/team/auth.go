package team

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const localhostRedir = "http://localhost:43672/"

type AuthToken struct {
	IdToken      string    `json:"id_token"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	TokenType    string    `json:"token_type"`
}

type rawAuthToken struct {
	IdToken      string `json:"id_token"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

func FetchToken(ctx context.Context, cfg *RemoteConfig) (*AuthToken, error) {
	slog.Info("Fetching authentication token")

	now := time.Now()

	codeChan := make(chan string, 1)

	hs := &http.Server{
		Addr: ":43672",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			params := r.URL.Query()

			code := params.Get("code")
			if code != "" {
				slog.Info("Got code from challenge", "code", code)

				select {
				case codeChan <- code:
				default:
					slog.Warn("Failed to send code")
				}
			}

			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`<html>
<head>
</head>
<body>
You can close this window now.

<script>
  setTimeout(function() {
      window.close()
  }, 1000);
</script>
</body>
</html>
`))
		}),
	}

	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		defer cancel()

		if err := hs.Shutdown(ctx); err != nil {
			slog.Warn("failed to shutdown http server", "err", err)
		}
	}()

	ctx, cancel := context.WithCancelCause(ctx)

	go func() {
		cancel(hs.ListenAndServe())
	}()

	state := randomCharacters(32)
	pkceKey, challenge := generateChallenge()

	params := url.Values{
		"redirect_uri":  {localhostRedir},
		"response_type": {cfg.OAuthResponseType},
		"client_id":     {cfg.UserPoolClientID},
		"scope":         {strings.Join(cfg.OAuthScopes, " ")},
		"state":         {state},
	}

	if cfg.OAuthResponseType == "code" {
		params.Add("code_challenge", challenge)
		params.Add("code_challenge_method", "S256")
	}

	u := url.URL{
		Scheme:   "https",
		Host:     cfg.OAuthDomain,
		Path:     "/oauth2/authorize",
		RawQuery: params.Encode(),
	}

	fmt.Println("\nPlease visit the following URL in your browser to authenticate:")
	fmt.Println(u.String())

	var code string

	select {
	case code = <-codeChan:
		// ok
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(time.Minute * 5):
		slog.Info("Timeout waiting for challenge")

		return nil, errors.New("timeout waiting for challenge")
	}

	ctx, cancelTimeout := context.WithTimeout(ctx, time.Second*30)
	defer cancelTimeout()

	u = url.URL{
		Scheme: "https",
		Host:   cfg.OAuthDomain,
		Path:   "/oauth2/token",
	}

	data := make(url.Values)
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("client_id", cfg.UserPoolClientID)
	data.Set("redirect_uri", localhostRedir)
	data.Set("code_verifier", pkceKey)

	r, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}

	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return nil, fmt.Errorf("failed to send token request: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: unexpected token status code: %d", ErrUnexpected, resp.StatusCode)
	}

	rawEnc, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read token body: %w", err)
	}

	var token *rawAuthToken

	if err := json.Unmarshal(rawEnc, &token); err != nil {
		return nil, fmt.Errorf("failed to unmarshal token body: %w", err)
	}

	return &AuthToken{
		IdToken:      token.IdToken,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresAt:    now.Add(time.Duration(token.ExpiresIn) * time.Millisecond),
		TokenType:    token.TokenType,
	}, nil
}

var randChars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randomCharacters(l int) string {
	out := make([]byte, l)

	for i := 0; i < l; i++ {
		out[i] = randChars[rand.IntN(len(randChars))]
	}

	return string(out)
}

func generateChallenge() (string, string) {
	chars := randomCharacters(32)

	challenge := base64.RawURLEncoding.EncodeToString([]byte(chars))

	hash := sha256.Sum256([]byte(challenge))

	encoded := base64.RawURLEncoding.EncodeToString(hash[:])

	return challenge, encoded
}
