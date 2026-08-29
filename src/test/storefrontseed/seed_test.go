package storefrontseed

import "testing"

func TestFixtureIdentifiersAreStable(t *testing.T) {
	t.Parallel()

	for name, value := range map[string]string{
		"product":  ProductID,
		"detail":   DetailID,
		"price A":  PriceAID,
		"price B":  PriceBID,
		"category": CategoryID,
	} {
		if len(value) != 36 {
			t.Fatalf("%s identifier length = %d, want UUID length", name, len(value))
		}
	}
	if PriceAID == PriceBID {
		t.Fatal("price identifiers must be distinct")
	}
}
