package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type testTokenRotator struct{ calls int }

func (r *testTokenRotator) ForceRegenerate() error { r.calls++; return nil }

func TestPasswordChangeRequiresCurrentPasswordAndRotatesTokens(t *testing.T) {
	credentials, err := newOwnerCredentialStore(t.TempDir(), "bootstrap-password")
	if err != nil {
		t.Fatal(err)
	}
	rotator := &testTokenRotator{}
	h := passwordChangeHandler(credentials, rotator)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "http://100.64.0.1:2003/password", strings.NewReader(`{"current_password":"wrong-password","new_password":"new-memorable-password","confirm_password":"new-memorable-password"}`))
	r.Header.Set("Origin", "http://100.64.0.1:2003")
	h(w, r)
	if w.Code != http.StatusUnauthorized || rotator.calls != 0 {
		t.Fatalf("wrong current password: status=%d rotations=%d", w.Code, rotator.calls)
	}
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodPost, "http://100.64.0.1:2003/password", strings.NewReader(`{"current_password":"bootstrap-password","new_password":"new-memorable-password","confirm_password":"new-memorable-password"}`))
	r.Header.Set("Origin", "http://100.64.0.1:2003")
	h(w, r)
	if w.Code != http.StatusOK || rotator.calls != 1 || !credentials.validate("new-memorable-password") {
		t.Fatalf("valid change failed: status=%d rotations=%d", w.Code, rotator.calls)
	}
}

func TestPasswordChangeRejectsCrossOriginAndMismatch(t *testing.T) {
	credentials, err := newOwnerCredentialStore(t.TempDir(), "bootstrap-password")
	if err != nil {
		t.Fatal(err)
	}
	rotator := &testTokenRotator{}
	h := passwordChangeHandler(credentials, rotator)
	for _, tc := range []struct {
		origin, body string
		want         int
	}{
		{"https://untrusted.example", `{"current_password":"bootstrap-password","new_password":"new-memorable-password","confirm_password":"new-memorable-password"}`, http.StatusForbidden},
		{"http://100.64.0.1:2003", `{"current_password":"bootstrap-password","new_password":"new-memorable-password","confirm_password":"different-password"}`, http.StatusBadRequest},
	} {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "http://100.64.0.1:2003/password", strings.NewReader(tc.body))
		r.Header.Set("Origin", tc.origin)
		h(w, r)
		if w.Code != tc.want {
			t.Fatalf("origin=%q: got %d want %d", tc.origin, w.Code, tc.want)
		}
	}
	if rotator.calls != 0 {
		t.Fatal("rejected requests rotated tokens")
	}
}
