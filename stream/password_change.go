package main

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
)

type tokenRotator interface {
	ForceRegenerate() error
}

func passwordChangeHandler(credentials *ownerCredentialStore, tokens tokenRotator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if r.Method != http.MethodPost {
			http.Error(w, "POST required", http.StatusMethodNotAllowed)
			return
		}
		if origin := r.Header.Get("Origin"); origin != "" && origin != "https://"+r.Host && origin != "http://"+r.Host {
			http.Error(w, "Origin rejected", http.StatusForbidden)
			return
		}
		if credentials == nil || tokens == nil {
			http.Error(w, "Password changes unavailable", http.StatusServiceUnavailable)
			return
		}
		var req struct {
			Current string `json:"current_password"`
			New     string `json:"new_password"`
			Confirm string `json:"confirm_password"`
		}
		if json.NewDecoder(io.LimitReader(r.Body, 4096)).Decode(&req) != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}
		if req.New != req.Confirm {
			http.Error(w, "New passwords do not match", http.StatusBadRequest)
			return
		}
		if err := credentials.change(req.Current, req.New); err != nil {
			if errors.Is(err, errInvalidCurrentPassword) {
				http.Error(w, "Current password is incorrect", http.StatusUnauthorized)
				return
			}
			if err := validateNewPassword(req.New); err != nil || req.Current == req.New {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			log.Printf("Owner password update failed: %v", err)
			http.Error(w, "Password update failed", http.StatusInternalServerError)
			return
		}
		if err := tokens.ForceRegenerate(); err != nil {
			log.Printf("Owner password changed but session invalidation failed: %v", err)
			http.Error(w, "Password changed; restart the sharing service to invalidate old sessions", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"changed": true})
	}
}
