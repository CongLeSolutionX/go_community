package syscall

import "testing"

func TestTOKEN_ALL_ACCESS(t *testing.T) {
	if TOKEN_ALL_ACCESS != 0xF01FF {
		t.Errorf("TOKEN_ALL_ACCESS = %x, want 0xF01FF", TOKEN_ALLACCESS)
	}
}
