package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOwnerCredentialStoreChangePersistsHash(t *testing.T) {
	dir := t.TempDir()
	s, err := newOwnerCredentialStore(dir, "bootstrap-password")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.change("bootstrap-password", "memorable-password"); err != nil {
		t.Fatal(err)
	}
	if s.validate("bootstrap-password") || !s.validate("memorable-password") {
		t.Fatal("credential change did not take effect")
	}
	b, err := os.ReadFile(filepath.Join(dir, ownerPasswordFilename))
	if err != nil {
		t.Fatal(err)
	}
	if string(b) == "memorable-password" {
		t.Fatal("password stored in plaintext")
	}
	info, err := os.Stat(filepath.Join(dir, ownerPasswordFilename))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0600 {
		t.Fatalf("password hash mode %o, want 600", info.Mode().Perm())
	}
	reloaded, err := newOwnerCredentialStore(dir, "ignored-bootstrap")
	if err != nil {
		t.Fatal(err)
	}
	if !reloaded.validate("memorable-password") || reloaded.validate("ignored-bootstrap") {
		t.Fatal("persisted hash did not override bootstrap password")
	}
}

func TestOwnerCredentialStoreRejectsUnsafeChanges(t *testing.T) {
	s, err := newOwnerCredentialStore(t.TempDir(), "bootstrap-password")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ current, next string }{
		{"wrong-password", "long-enough-password"},
		{"bootstrap-password", "short"},
		{"bootstrap-password", "bootstrap-password"},
	} {
		if err := s.change(tc.current, tc.next); err == nil {
			t.Fatalf("accepted unsafe change %#v", tc)
		}
	}
}
