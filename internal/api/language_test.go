package api

import (
	"net/http"
	"testing"
)

func TestStreamListsRejectUnknownLanguages(t *testing.T) {
	e := newEnv(t)
	for _, path := range []string{
		"/v1/streams?language=xx",
		"/v1/games/chess/streams?language=de,xx",
		"/v1/tags/Speedrun/streams?language=english",
	} {
		if status, body := e.get(t, path, true); status != http.StatusBadRequest {
			t.Errorf("%s: status %d %s", path, status, body)
		}
	}
}
