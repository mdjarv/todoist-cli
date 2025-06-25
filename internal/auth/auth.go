package auth

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mdjarv/todoist-cli/internal/todoist"
)

const (
	AuthURL      = "https://todoist.com/oauth/authorize"
	Scopes       = "data:read_write,data:delete"
	CallbackPort = 47829
)

var (
	ClientID     string
	ClientSecret string
)

type Credentials struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

func Login() error {
	if err := loadEnv(); err != nil {
		return fmt.Errorf("failed to load .env file: %w", err)
	}

	if ClientID == "" || ClientSecret == "" {
		return fmt.Errorf("TODOIST_CLIENT_ID and TODOIST_CLIENT_SECRET must be set in .env file")
	}

	state, err := generateRandomState()
	if err != nil {
		return fmt.Errorf("failed to generate state: %w", err)
	}

	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", CallbackPort))
	if err != nil {
		return fmt.Errorf("failed to start local server on port %d: %w", CallbackPort, err)
	}
	defer listener.Close()

	redirectURI := fmt.Sprintf("http://localhost:%d/callback", CallbackPort)
	authURL := buildAuthURL(redirectURI, state)

	fmt.Printf("Please visit the following URL to authorize the application:\n\n%s\n\n", authURL)
	fmt.Println("Waiting for authorization...")

	codeChan := make(chan string, 1)
	errChan := make(chan error, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		handleCallback(w, r, state, codeChan, errChan)
	})

	server := &http.Server{Handler: mux}
	go func() {
		if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- fmt.Errorf("server error: %w", err)
		}
	}()

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

	server.Shutdown(context.Background())

	// Use the client for token exchange
	client := todoist.NewClient("") // Empty token for OAuth calls
	tokenResp, err := client.ExchangeCodeForToken(code, redirectURI, ClientID, ClientSecret)
	if err != nil {
		return fmt.Errorf("failed to exchange code for token: %w", err)
	}

	creds := &Credentials{
		AccessToken: tokenResp.AccessToken,
		TokenType:   tokenResp.TokenType,
	}

	if err := saveCredentials(creds); err != nil {
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
	if errParam := r.URL.Query().Get("error"); errParam != "" {
		errChan <- fmt.Errorf("authorization error: %s", errParam)
		http.Error(w, "Authorization failed", http.StatusBadRequest)
		return
	}

	state := r.URL.Query().Get("state")
	if state != expectedState {
		errChan <- fmt.Errorf("invalid state parameter")
		http.Error(w, "Invalid state parameter", http.StatusBadRequest)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		errChan <- fmt.Errorf("missing authorization code")
		http.Error(w, "Missing authorization code", http.StatusBadRequest)
		return
	}

	fmt.Fprintf(w, "Authorization successful! You can close this window.")
	codeChan <- code
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
