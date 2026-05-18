package wgkeys

import "testing"

func TestGenerateProducesDifferentKeys(t *testing.T) {
	a, err := Generate()
	if err != nil {
		t.Fatal(err)
	}
	b, err := Generate()
	if err != nil {
		t.Fatal(err)
	}
	if a.Private == b.Private {
		t.Fatal("expected unique private keys")
	}
	if a.Public == "" || a.Private == "" {
		t.Fatal("keys must not be empty")
	}
}
