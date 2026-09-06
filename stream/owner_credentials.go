package main

import (
	"crypto/subtle"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"golang.org/x/crypto/bcrypt"
)

const ownerPasswordFilename = "owner_password.bcrypt"

var errInvalidCurrentPassword = errors.New("current password is incorrect")

type ownerCredentialStore struct {
	mu       sync.RWMutex
	path     string
	hash     []byte
	fallback []byte
}

func newOwnerCredentialStore(secretDir, fallbackPassword string) (*ownerCredentialStore, error) {
	s := &ownerCredentialStore{path: filepath.Join(secretDir, ownerPasswordFilename), fallback: []byte(fallbackPassword)}
	hash, err := os.ReadFile(s.path)
	if err == nil {
		if _, err := bcrypt.Cost(hash); err != nil {
			return nil, fmt.Errorf("invalid owner password hash: %w", err)
		}
		s.hash = hash
		s.fallback = nil
		return s, nil
	}
	if !os.IsNotExist(err) {
		return nil, fmt.Errorf("load owner password hash: %w", err)
	}
	return s, nil
}

func (s *ownerCredentialStore) validate(password string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.validateLocked(password)
}

func (s *ownerCredentialStore) validateLocked(password string) bool {
	if len(s.hash) != 0 {
		return bcrypt.CompareHashAndPassword(s.hash, []byte(password)) == nil
	}
	return subtle.ConstantTimeCompare(s.fallback, []byte(password)) == 1
}

func validateNewPassword(password string) error {
	if len(password) < 10 {
		return errors.New("new password must be at least 10 characters")
	}
	if len(password) > 72 {
		return errors.New("new password must be at most 72 UTF-8 bytes")
	}
	return nil
}

func (s *ownerCredentialStore) change(currentPassword, newPassword string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.validateLocked(currentPassword) {
		return errInvalidCurrentPassword
	}
	if err := validateNewPassword(newPassword); err != nil {
		return err
	}
	if currentPassword == newPassword {
		return errors.New("new password must be different")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash new password: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0700); err != nil {
		return fmt.Errorf("create credential directory: %w", err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(s.path), ".owner-password-*")
	if err != nil {
		return fmt.Errorf("create credential update: %w", err)
	}
	tmpPath := tmp.Name()
	ok := false
	defer func() {
		tmp.Close()
		if !ok {
			os.Remove(tmpPath)
		}
	}()
	if err := tmp.Chmod(0600); err != nil {
		return fmt.Errorf("protect credential update: %w", err)
	}
	if _, err := tmp.Write(hash); err != nil {
		return fmt.Errorf("write credential update: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		return fmt.Errorf("sync credential update: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close credential update: %w", err)
	}
	if err := os.Rename(tmpPath, s.path); err != nil {
		return fmt.Errorf("install credential update: %w", err)
	}
	ok = true
	s.hash = hash
	s.fallback = nil
	return nil
}
