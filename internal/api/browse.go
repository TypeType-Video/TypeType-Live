package api

import (
	"net/http"
	"strings"

	"github.com/TypeType-Video/TypeType-Live/internal/live"
	"github.com/TypeType-Video/TypeType-Live/twitch"
)

func respond[T any](w http.ResponseWriter, r *http.Request, value T, err error) {
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, value)
}

func loginParam(w http.ResponseWriter, r *http.Request) (string, bool) {
	login := strings.ToLower(r.PathValue("login"))
	if !twitch.ValidLogin(login) {
		writeJSON(w, http.StatusBadRequest, errorBody{Error: "invalid_login"})
		return "", false
	}
	return login, true
}

func languageParam(w http.ResponseWriter, r *http.Request) ([]string, bool) {
	languages, ok := twitch.ParseLanguages(r.URL.Query().Get("language"))
	if !ok {
		writeJSON(w, http.StatusBadRequest, errorBody{Error: "invalid_language"})
	}
	return languages, ok
}

func (s *Server) topStreams(w http.ResponseWriter, r *http.Request) {
	languages, ok := languageParam(w, r)
	if !ok {
		return
	}
	first, after := pageParams(r)
	page, err := s.twitch.TopStreams(r.Context(), languages, first, after)
	respond(w, r, page, err)
}

func (s *Server) topGames(w http.ResponseWriter, r *http.Request) {
	first, after := pageParams(r)
	page, err := s.twitch.TopGames(r.Context(), first, after)
	respond(w, r, page, err)
}

type gameStreamsBody struct {
	Game    *twitch.Game               `json:"game"`
	Streams twitch.Page[twitch.Stream] `json:"streams"`
}

func (s *Server) gameStreams(w http.ResponseWriter, r *http.Request) {
	languages, ok := languageParam(w, r)
	if !ok {
		return
	}
	first, after := pageParams(r)
	game, page, err := s.twitch.GameStreams(r.Context(), r.PathValue("slug"), languages, first, after)
	respond(w, r, gameStreamsBody{Game: game, Streams: page}, err)
}

func (s *Server) tagStreams(w http.ResponseWriter, r *http.Request) {
	tag := r.PathValue("tag")
	if !twitch.ValidTag(tag) {
		writeJSON(w, http.StatusBadRequest, errorBody{Error: "invalid_tag"})
		return
	}
	languages, ok := languageParam(w, r)
	if !ok {
		return
	}
	first, after := pageParams(r)
	page, err := s.twitch.TagStreams(r.Context(), tag, languages, first, after)
	respond(w, r, page, err)
}

func (s *Server) channel(w http.ResponseWriter, r *http.Request) {
	login, ok := loginParam(w, r)
	if !ok {
		return
	}
	ch, err := s.channels.Channel(r.Context(), login)
	respond(w, r, ch, err)
}

func (s *Server) channelVideos(w http.ResponseWriter, r *http.Request) {
	login, ok := loginParam(w, r)
	if !ok {
		return
	}
	first, after := pageParams(r)
	page, err := s.twitch.ChannelVideos(r.Context(), login, first, after)
	respond(w, r, page, err)
}

func (s *Server) channelClips(w http.ResponseWriter, r *http.Request) {
	login, ok := loginParam(w, r)
	if !ok {
		return
	}
	first, after := pageParams(r)
	page, err := s.twitch.ChannelClips(r.Context(), login, first, after)
	respond(w, r, page, err)
}

type liveBody struct {
	Channel  twitch.Channel `json:"channel"`
	HLSURL   string         `json:"hlsUrl"`
	AudioURL string         `json:"audioUrl"`
}

func (s *Server) liveStream(w http.ResponseWriter, r *http.Request) {
	login, ok := loginParam(w, r)
	if !ok {
		return
	}
	ch, err := s.channels.Channel(r.Context(), login)
	if err != nil {
		writeError(w, r, err)
		return
	}
	if ch.Live == nil {
		writeError(w, r, twitch.ErrOffline)
		return
	}
	link, audio := s.live.PrepareMaster(r.Context(), live.KindLive, login, r.Header.Get(tokenHeader))
	writeJSON(w, http.StatusOK, liveBody{Channel: ch, HLSURL: s.hlsLink(link), AudioURL: s.hlsLink(audio)})
}

type videoBody struct {
	Video    twitch.Video `json:"video"`
	HLSURL   string       `json:"hlsUrl"`
	AudioURL string       `json:"audioUrl"`
}

func (s *Server) video(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !twitch.ValidID(id) {
		writeJSON(w, http.StatusBadRequest, errorBody{Error: "invalid_id"})
		return
	}
	v, err := s.twitch.Video(r.Context(), id)
	if err != nil {
		writeError(w, r, err)
		return
	}
	link, audio := s.live.PrepareMaster(r.Context(), live.KindVOD, id, r.Header.Get(tokenHeader))
	writeJSON(w, http.StatusOK, videoBody{Video: v, HLSURL: s.hlsLink(link), AudioURL: s.hlsLink(audio)})
}

type clipBody struct {
	Clip      twitch.Clip          `json:"clip"`
	Qualities []twitch.ClipQuality `json:"qualities"`
}

func (s *Server) clip(w http.ResponseWriter, r *http.Request) {
	clip, qualities, err := s.twitch.Clip(r.Context(), r.PathValue("slug"))
	if err != nil {
		writeError(w, r, err)
		return
	}
	for i := range qualities {
		qualities[i].URL = s.hlsLink(s.live.SignedClipLink(qualities[i].URL))
	}
	writeJSON(w, http.StatusOK, clipBody{Clip: clip, Qualities: qualities})
}

func (s *Server) search(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" || len(q) > 100 {
		writeJSON(w, http.StatusBadRequest, errorBody{Error: "invalid_query"})
		return
	}
	first, after := pageParams(r)
	switch r.URL.Query().Get("type") {
	case "games":
		page, err := s.twitch.SearchGames(r.Context(), q, first, after)
		respond(w, r, page, err)
	case "streams":
		page, err := s.twitch.SearchStreams(r.Context(), q, first, after)
		respond(w, r, page, err)
	default:
		page, err := s.twitch.SearchChannels(r.Context(), q, first, after)
		respond(w, r, page, err)
	}
}
