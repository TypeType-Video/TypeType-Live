package twitch

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"testing"
)

func TestParseLanguages(t *testing.T) {
	cases := map[string][]string{
		"":          nil,
		" , ":       nil,
		"de":        {"DE"},
		"DE,en":     {"DE", "EN"},
		"de, de":    {"DE"},
		"zh-hk":     {"ZH_HK"},
		"other,asl": {"OTHER", "ASL"},
	}
	for raw, want := range cases {
		got, ok := ParseLanguages(raw)
		if !ok || !reflect.DeepEqual(got, want) {
			t.Errorf("ParseLanguages(%q) = %v, %v; want %v", raw, got, ok, want)
		}
	}
	for _, raw := range []string{"xx", "de,xx", "english", "de;en", "../x"} {
		if _, ok := ParseLanguages(raw); ok {
			t.Errorf("ParseLanguages(%q) accepted", raw)
		}
	}
}

func TestStreamsSendLanguagesAndKeepTheLanguage(t *testing.T) {
	var seen []string
	c := fakeClient(t, func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Variables map[string]json.RawMessage `json:"variables"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		seen = append(seen, string(req.Variables["languages"]))
		io.WriteString(w, `{"data":{"streams":{"edges":[{"cursor":"c1","node":{"id":"1","viewersCount":5,"createdAt":"2026-11-14T20:00:00Z",
			"broadcaster":{"id":"2","login":"a","displayName":"A","broadcastSettings":{"title":"Titel","language":"DE"}}}}],
			"pageInfo":{"hasNextPage":false}}}}`)
	})
	page, err := c.TopStreams(context.Background(), []string{"DE", "EN"}, 5, "")
	if err != nil {
		t.Fatal(err)
	}
	if page.Items[0].Language != "DE" {
		t.Fatalf("language %q", page.Items[0].Language)
	}
	if _, err := c.TagStreams(context.Background(), "Speedrun", []string{"DE"}, 5, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := c.TopStreams(context.Background(), nil, 5, ""); err != nil {
		t.Fatal(err)
	}
	want := []string{`["DE","EN"]`, `["DE"]`, ""}
	if !reflect.DeepEqual(seen, want) {
		t.Fatalf("languages sent %q, want %q", seen, want)
	}
}
