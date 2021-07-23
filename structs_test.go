package eps

import (
	"testing"
)

func TestAddressRegexp(t *testing.T) {

	if groups := IDAddressRegexp.FindStringSubmatch("normal-name.f(acdds)"); groups == nil {
		t.Fatal("invalid match")
	} else if groups[1] != "normal-name" {
		t.Fatal("invalid operator name")
	} else if groups[2] != "f" {
		t.Fatalf("invalid method name: '%s'", groups[2])
	} else if groups[3] != "acdds" {
		t.Fatal("invalid ID")
	}

	if groups := IDAddressRegexp.FindStringSubmatch("a.b.c.d.e.f(acdds)"); groups == nil {
		t.Fatal("invalid match")
	} else if groups[1] != "a.b.c.d.e" {
		t.Fatal("invalid operator name")
	} else if groups[2] != "f" {
		t.Fatalf("invalid method name: '%s'", groups[2])
	} else if groups[3] != "acdds" {
		t.Fatal("invalid ID")
	}
}
