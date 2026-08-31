package integrationseed

import "testing"

func TestRootFixtureContract(t *testing.T) {
	if RootEmail != "root@gmail.com" {
		t.Fatalf("RootEmail = %q", RootEmail)
	}
	if len(rootPasswordHash) != 60 {
		t.Fatalf("root password hash length = %d, want bcrypt length", len(rootPasswordHash))
	}
}
