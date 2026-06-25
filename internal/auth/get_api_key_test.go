package auth

import (
	"net/http"
	"testing"
)

// TestGetAPIKeySuccess prüft, dass ein korrekt formatierter Header das Token liefert.
func TestGetAPIKeySuccess(t *testing.T) {
	const expected = "test-api-key-123"

	// Header mit gültigem Authorization‑Wert erzeugen
	headers := http.Header{}
	headers.Set("Authorization", "ApiKey "+expected)

	got, err := GetAPIKey(headers)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != expected {
		t.Fatalf("expected %q, got %q", expected, got)
	}
}

// TestGetAPIKeyMissing prüft das Verhalten bei fehlendem Header.
func TestGetAPIKeyMissing(t *testing.T) {
	headers := http.Header{} // kein Authorization‑Header

	_, err := GetAPIKey(headers)
	if err == nil {
		t.Fatalf("expected an error when Authorization header is missing")
	}
	if err != ErrNoAuthHeaderIncluded {
		t.Fatalf("expected ErrNoAuthHeaderIncluded, got %v", err)
	}
}

// TestGetAPIKeyMalformed prüft das Verhalten bei einem falsch formatierten Header.
func TestGetAPIKeyMalformed(t *testing.T) {
	// Header mit falschem Präfix bzw. fehlendem Token
	cases := []string{
		"Bearer abcdef",    // falscher Scheme
		"ApiKey",           // kein Token
		"ApiKey ",          // leerer Token
		"ApiKey   abc def", // zu viele Leerzeichen (nur das erste Token wird verwendet, Rest ist ok – hier aber bewusst ungültig)
	}

	for _, val := range cases {
		headers := http.Header{}
		headers.Set("Authorization", val)

		_, err := GetAPIKey(headers)
		if err == nil {
			t.Fatalf("expected an error for malformed header %q", val)
		}
		// Der konkrete Fehlertyp ist nicht definiert – nur sicherstellen, dass ein Fehler kommt
	}
}
