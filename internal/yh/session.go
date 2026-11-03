package yh

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Session is what the app keeps between runs: the account's token, and
// what it needs to sign in again.
//
// Name and Avatar are the account as the sign-in screen shows it while it
// is not signed in: the face and the name at the top of the panel are the
// account that was here last, not one that has been proved.
type Session struct {
	Token    string `json:"token"`
	Email    string `json:"email,omitempty"`
	Phone    string `json:"phone,omitempty"`
	DeviceID string `json:"deviceId"`
	Base     string `json:"base,omitempty"`
	Name     string `json:"name,omitempty"`
	Avatar   string `json:"avatar,omitempty"`
	// Password is kept when the reader asks to be remembered: signing in
	// again after the token has gone is then a press rather than a form.
	Password string `json:"password,omitempty"`
}

// ConfigPath is where the app keeps its session.
func ConfigPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "mygohu", "session.json"), nil
}

// LoadSession reads the saved session, if there is one.
func LoadSession() (*Session, error) {
	path, err := ConfigPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var s Session
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// SaveSession writes the session for the next run.
func SaveSession(s *Session) error {
	path, err := ConfigPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

// ClearSession forgets the saved session.
func ClearSession() error {
	path, err := ConfigPath()
	if err != nil {
		return err
	}
	err = os.Remove(path)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}
