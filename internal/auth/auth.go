package auth

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	AuthURL      = "https://todoist.com/oauth/authorize"
	TokenURL     = "https://todoist.com/oauth/access_token"
	Scopes       = "data:read_write,data:delete" // Covers list, add, remove, complete tasks
	CallbackPort = 47829                         // Static port for OAuth callback
)

var (
	ClientID     string
	ClientSecret string
)

type Credentials struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Error       string `json:"error,omitempty"`
}

func Login() error {
	// Load environment variables from .env file
	if err := loadEnv(); err != nil {
		return fmt.Errorf("failed to load .env file: %w", err)
	}

	if ClientID == "" || ClientSecret == "" {
		return fmt.Errorf("TODOIST_CLIENT_ID and TODOIST_CLIENT_SECRET must be set in .env file")
	}

	// Generate random state for CSRF protection
	state, err := generateRandomState()
	if err != nil {
		return fmt.Errorf("failed to generate state: %w", err)
	}

	// Start local server for OAuth callback
	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", CallbackPort))
	if err != nil {
		return fmt.Errorf("failed to start local server on port %d: %w", CallbackPort, err)
	}
	defer listener.Close()

	redirectURI := fmt.Sprintf("http://localhost:%d/callback", CallbackPort)

	// Build authorization URL
	authURL := buildAuthURL(redirectURI, state)

	fmt.Printf("Please visit the following URL to authorize the application:\n\n%s\n\n", authURL)
	fmt.Println("Waiting for authorization...")

	// Channel to receive authorization code
	codeChan := make(chan string, 1)
	errChan := make(chan error, 1)

	// Start HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		handleCallback(w, r, state, codeChan, errChan)
	})

	server := &http.Server{Handler: mux}
	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("server error: %w", err)
		}
	}()

	// Wait for callback or timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	var code string
	select {
	case code = <-codeChan:
		// Success
	case err := <-errChan:
		return err
	case <-ctx.Done():
		return fmt.Errorf("authorization timeout")
	}

	// Shutdown server
	server.Shutdown(context.Background())

	// Exchange code for token
	token, err := exchangeCodeForToken(code, redirectURI)
	if err != nil {
		return fmt.Errorf("failed to exchange code for token: %w", err)
	}

	// Save credentials
	if err := saveCredentials(token); err != nil {
		return fmt.Errorf("failed to save credentials: %w", err)
	}

	fmt.Println("Successfully authenticated!")
	return nil
}

func generateRandomState() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func buildAuthURL(redirectURI, state string) string {
	params := url.Values{
		"client_id":     {ClientID},
		"scope":         {Scopes},
		"state":         {state},
		"redirect_uri":  {redirectURI},
		"response_type": {"code"},
	}
	return AuthURL + "?" + params.Encode()
}

func handleCallback(w http.ResponseWriter, r *http.Request, expectedState string, codeChan chan<- string, errChan chan<- error) {
	// Check for error parameter
	if errParam := r.URL.Query().Get("error"); errParam != "" {
		errChan <- fmt.Errorf("authorization error: %s", errParam)
		http.Error(w, "Authorization failed", http.StatusBadRequest)
		return
	}

	// Verify state parameter
	state := r.URL.Query().Get("state")
	if state != expectedState {
		errChan <- fmt.Errorf("invalid state parameter")
		http.Error(w, "Invalid state parameter", http.StatusBadRequest)
		return
	}

	// Get authorization code
	code := r.URL.Query().Get("code")
	if code == "" {
		errChan <- fmt.Errorf("missing authorization code")
		http.Error(w, "Missing authorization code", http.StatusBadRequest)
		return
	}

	// Send success response to browser
	fmt.Fprintf(w, "Authorization successful! You can close this window.")

	// Send code to main goroutine
	codeChan <- code
}

func exchangeCodeForToken(code, redirectURI string) (*Credentials, error) {
	data := url.Values{
		"client_id":     {ClientID},
		"client_secret": {ClientSecret},
		"code":          {code},
		"redirect_uri":  {redirectURI},
	}

	resp, err := http.PostForm(TokenURL, data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, err
	}

	if tokenResp.Error != "" {
		return nil, fmt.Errorf("token exchange error: %s", tokenResp.Error)
	}

	if tokenResp.AccessToken == "" {
		return nil, fmt.Errorf("empty access token received")
	}

	return &Credentials{
		AccessToken: tokenResp.AccessToken,
		TokenType:   tokenResp.TokenType,
	}, nil
}

func saveCredentials(creds *Credentials) error {
	configDir := filepath.Join(os.Getenv("HOME"), ".config", "todoist")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	credentialsPath := filepath.Join(configDir, "credentials.json")
	data, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(credentialsPath, data, 0600)
}

func LoadCredentials() (*Credentials, error) {
	credentialsPath := filepath.Join(os.Getenv("HOME"), ".config", "todoist", "credentials.json")
	data, err := os.ReadFile(credentialsPath)
	if err != nil {
		return nil, err
	}

	var creds Credentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, err
	}

	return &creds, nil
}

func loadEnv() error {
	file, err := os.Open(".env")
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "TODOIST_CLIENT_ID":
			ClientID = value
		case "TODOIST_CLIENT_SECRET":
			ClientSecret = value
		}
	}

	return scanner.Err()
}
