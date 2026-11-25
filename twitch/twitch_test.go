package twitch

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func fakeClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	c := New(srv.Client(), "")
	c.GQLURL, c.UsherURL, c.IDURL = srv.URL+"/gql", srv.URL, srv.URL
	return c
}

func TestLiveTokenSendsPersistedQueryAndDeviceID(t *testing.T) {
	c := fakeClient(t, func(w http.ResponseWriter, r *http.Request) {
		var body gqlRequest
		_ = json.NewDecoder(r.Body).Decode(&body)
		if r.Header.Get("Client-Id") != DefaultClientID || len(r.Header.Get("X-Device-Id")) != 32 {
			http.Error(w, "headers", http.StatusBadRequest)
			return
		}
		if body.Extensions == nil || body.Extensions.PersistedQuery.SHA256Hash != playbackAccessTokenHash || body.Variables["login"] != "chan" {
			http.Error(w, "body", http.StatusBadRequest)
			return
		}
		io.WriteString(w, `{"data":{"streamPlaybackAccessToken":{"value":"{\"expires\":1794696000,\"server_ads\":true}","signature":"abc"}}}`)
	})
	tok, err := c.LiveToken(context.Background(), "chan", DefaultPlayer, "")
	if err != nil {
		t.Fatal(err)
	}
	if tok.Signature != "abc" || !tok.ServerAds || tok.Expires.Unix() != 1794696000 {
		t.Fatalf("token %+v", tok)
	}
	master, _ := url.Parse(c.LiveMasterURL("chan", tok, "web"))
	q := master.Query()
	if master.Path != "/api/v2/channel/hls/chan.m3u8" || q.Get("sig") != "abc" || q.Get("fast_bread") != "true" || q.Get("allow_source") != "true" {
		t.Fatalf("master URL %s", master)
	}
}

func TestGQLErrorsAreClassified(t *testing.T) {
	c := fakeClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "OAuth bad" {
			if r.Header.Get("Client-Id") != AuthClientID {
				http.Error(w, "client", http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		io.WriteString(w, `{"errors":[{"message":"failed integrity check"}],"data":null}`)
	})
	if _, err := c.TopStreams(context.Background(), nil, 5, ""); !errors.Is(err, ErrIntegrity) {
		t.Fatalf("err = %v", err)
	}
	if _, err := c.CurrentUser(context.Background(), "bad"); !errors.Is(err, ErrBadToken) {
		t.Fatalf("err = %v", err)
	}
}

func TestTopStreamsMapsEdges(t *testing.T) {
	c := fakeClient(t, func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, `{"data":{"streams":{"edges":[{"cursor":"c1","node":{"id":"1","viewersCount":5,"createdAt":"2026-11-14T20:00:00Z",
			"game":{"id":"9","displayName":"Chess","slug":"chess"},"broadcaster":{"id":"2","login":"a","displayName":"A","broadcastSettings":{"title":"Title"}}}}],
			"pageInfo":{"hasNextPage":true}}}}`)
	})
	page, err := c.TopStreams(context.Background(), nil, 5, "")
	if err != nil {
		t.Fatal(err)
	}
	s := page.Items[0]
	if page.Cursor != "c1" || !page.HasMore || s.Title != "Title" || s.User.Login != "a" || s.Game.Name != "Chess" || s.StartedAt.IsZero() {
		t.Fatalf("page %+v", page)
	}
}

func TestDeviceFlow(t *testing.T) {
	polls := 0
	c := fakeClient(t, func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		if r.Form.Get("client_id") != AuthClientID {
			http.Error(w, "invalid client", http.StatusBadRequest)
			return
		}
		switch r.URL.Path {
		case "/oauth2/device":
			if !strings.Contains(r.Form.Get("scopes"), "chat:read") {
				http.Error(w, "scopes", http.StatusBadRequest)
				return
			}
			io.WriteString(w, `{"device_code":"dev","user_code":"ABCD1234","verification_uri":"https://www.twitch.tv/activate","interval":5,"expires_in":1800}`)
		case "/oauth2/token":
			polls++
			if polls == 1 {
				w.WriteHeader(http.StatusBadRequest)
				io.WriteString(w, `{"status":400,"message":"authorization_pending"}`)
				return
			}
			io.WriteString(w, `{"access_token":"tok","refresh_token":"ref","expires_in":14000}`)
		}
	})
	code, err := c.StartDevice(context.Background())
	if err != nil || code.UserCode != "ABCD1234" || code.Interval != 5 {
		t.Fatalf("code %+v err %v", code, err)
	}
	if _, err := c.PollDevice(context.Background(), "dev"); !errors.Is(err, ErrAuthPending) {
		t.Fatalf("first poll err = %v", err)
	}
	tok, err := c.PollDevice(context.Background(), "dev")
	if err != nil || tok.AccessToken != "tok" || tok.RefreshToken != "ref" {
		t.Fatalf("token %+v err %v", tok, err)
	}
}

func TestValidators(t *testing.T) {
	if !ValidLogin("riot_games1") || ValidLogin("../x") || ValidLogin("") || ValidLogin(strings.Repeat("a", 26)) {
		t.Fatal("login validation")
	}
	if !ValidID("123") || ValidID("12a") {
		t.Fatal("id validation")
	}
}

func TestImagesOnlyKeepTheTwitchImageCDN(t *testing.T) {
	for in, want := range map[string]string{
		"https://static-cdn.jtvnw.net/previews-ttv/live_user_a-640x360.jpg": "https://static-cdn.jtvnw.net/previews-ttv/live_user_a-640x360.jpg",
		"https://vod-secure.twitch.tv/_404/404_processing_640x360.png":      "",
		"https://static-cdn.jtvnw.net/ttv-static/404_processing.jpg":        "",
		"https://evil.example/x.png":                                        "",
	} {
		if got := image(in); got != want {
			t.Errorf("image(%q) = %q", in, got)
		}
	}
}
