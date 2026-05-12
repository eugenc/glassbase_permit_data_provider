package repair

import "testing"

func TestTruncateUTF8Runes_ascii(t *testing.T) {
	const want = "ab\n… [truncated for prompt size]"
	if got := truncateUTF8Runes("abcdef", 2); got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestTruncateUTF8Runes_noop(t *testing.T) {
	if got := truncateUTF8Runes("hi", 10); got != "hi" {
		t.Fatalf("got %q", got)
	}
}

func TestTruncateUTF8Runes_wide(t *testing.T) {
	const in = "あいう"
	got := truncateUTF8Runes(in, 2)
	if rs := []rune(got); len(rs) < 2 {
		t.Fatalf("expected at least 2 runes in output, got %q", got)
	}
}
