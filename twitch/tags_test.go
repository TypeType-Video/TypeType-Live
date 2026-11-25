package twitch

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestValidTag(t *testing.T) {
	for _, tag := range []string{"English", "Speedrun", "日本語", "RTAイベント", "OCE", "ADHD2"} {
		if !ValidTag(tag) {
			t.Errorf("%q rejected", tag)
		}
	}
	for _, tag := range []string{"", "two words", "a/b", "../x", "tag?x=1", strings.Repeat("a", 26)} {
		if ValidTag(tag) {
			t.Errorf("%q accepted", tag)
		}
	}
}

func TestTagStreamsFiltersByTag(t *testing.T) {
	c := fakeClient(t, func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Variables struct {
				Tags []string `json:"tags"`
			} `json:"variables"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		if len(req.Variables.Tags) != 1 || req.Variables.Tags[0] != "Speedrun" {
			http.Error(w, "tags", http.StatusBadRequest)
			return
		}
		io.WriteString(w, `{"data":{"streams":{"edges":[{"cursor":"c1","node":{"id":"1","viewersCount":5,"createdAt":"2026-11-14T20:00:00Z",
			"freeformTags":[{"name":"Speedrun"}],"broadcaster":{"id":"2","login":"a","displayName":"A","broadcastSettings":{"title":"Run"}}}}],
			"pageInfo":{"hasNextPage":false}}}}`)
	})
	page, err := c.TagStreams(context.Background(), "Speedrun", nil, 24, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 1 || page.Items[0].User.Login != "a" || page.HasMore {
		t.Fatalf("page %+v", page)
	}
}
