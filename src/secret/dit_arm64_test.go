package secret

import "testing"

func TestDIT(t *testing.T) {
	WithDIT(func() (any, error) {
		t.Helper()

		if !ditEnabled() {
			t.Fatal("dit not enabled within WithDIT closure")
		}

		return nil, nil
	})
}
