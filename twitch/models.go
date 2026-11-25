package twitch

import (
	"strings"
	"time"
)

type Page[T any] struct {
	Items   []T    `json:"items"`
	Cursor  string `json:"cursor,omitempty"`
	HasMore bool   `json:"hasMore"`
}

type Game struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Slug     string `json:"slug"`
	BoxArt   string `json:"boxArtUrl,omitempty"`
	Viewers  int64  `json:"viewers,omitempty"`
	Channels int64  `json:"channels,omitempty"`
}

type User struct {
	ID          string `json:"id"`
	Login       string `json:"login"`
	DisplayName string `json:"displayName"`
	Avatar      string `json:"avatarUrl"`
}

type Stream struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	User      User      `json:"user"`
	Game      *Game     `json:"game,omitempty"`
	Viewers   int64     `json:"viewers"`
	Preview   string    `json:"previewUrl"`
	StartedAt time.Time `json:"startedAt"`
	Tags      []string  `json:"tags,omitempty"`
	// Language is the broadcaster language Twitch lists the stream under, such as "DE".
	Language string `json:"language,omitempty"`
	// Archive is the past broadcast Twitch records while the channel is live, when it keeps one.
	Archive *StreamArchive `json:"archive,omitempty"`
}

type StreamArchive struct {
	ID        string    `json:"id"`
	StartedAt time.Time `json:"startedAt"`
}

type Channel struct {
	User
	Description string  `json:"description"`
	Banner      string  `json:"bannerUrl,omitempty"`
	Followers   int64   `json:"followers"`
	Partner     bool    `json:"partner"`
	Live        *Stream `json:"live,omitempty"`
}

type Video struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Owner     User      `json:"owner"`
	Game      *Game     `json:"game,omitempty"`
	Length    int64     `json:"lengthSeconds"`
	Views     int64     `json:"views"`
	Preview   string    `json:"previewUrl"`
	CreatedAt time.Time `json:"createdAt"`
	Type      string    `json:"type"`
	// Recording marks the past broadcast of a live still running: its length keeps growing.
	Recording bool `json:"recording,omitempty"`

	Storyboard string `json:"-"`
}

type Clip struct {
	Slug        string    `json:"slug"`
	Title       string    `json:"title"`
	Broadcaster User      `json:"broadcaster"`
	Game        *Game     `json:"game,omitempty"`
	Duration    float64   `json:"durationSeconds"`
	Views       int64     `json:"views"`
	Thumbnail   string    `json:"thumbnailUrl"`
	CreatedAt   time.Time `json:"createdAt"`
	VideoID     string    `json:"videoId,omitempty"`
	VideoOffset int64     `json:"videoOffsetSeconds,omitempty"`
}

type ClipQuality struct {
	Quality   string  `json:"quality"`
	FrameRate float64 `json:"frameRate"`
	URL       string  `json:"url"`
}

const imageHost = "https://static-cdn.jtvnw.net/"

func image(url string) string {
	if !strings.HasPrefix(url, imageHost) || strings.Contains(url, "404_processing") {
		return ""
	}
	return url
}

type gqlGame struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
	Slug        string `json:"slug"`
	BoxArtURL   string `json:"boxArtURL"`
	Viewers     int64  `json:"viewersCount"`
	Channels    int64  `json:"broadcastersCount"`
}

func (g *gqlGame) model() *Game {
	if g == nil || g.ID == "" {
		return nil
	}
	return &Game{ID: g.ID, Name: g.DisplayName, Slug: g.Slug, BoxArt: image(g.BoxArtURL), Viewers: g.Viewers, Channels: g.Channels}
}

type gqlUser struct {
	ID          string `json:"id"`
	Login       string `json:"login"`
	DisplayName string `json:"displayName"`
	Avatar      string `json:"profileImageURL"`
}

func (u gqlUser) model() User {
	u.Avatar = image(u.Avatar)
	return User(u)
}

type gqlStream struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Viewers     int64    `json:"viewersCount"`
	Preview     string   `json:"previewImageURL"`
	CreatedAt   string   `json:"createdAt"`
	Game        *gqlGame `json:"game"`
	Broadcaster *struct {
		gqlUser
		Settings struct {
			Title    string `json:"title"`
			Language string `json:"language"`
		} `json:"broadcastSettings"`
	} `json:"broadcaster"`
	Archive *struct {
		ID        string `json:"id"`
		CreatedAt string `json:"createdAt"`
	} `json:"archiveVideo"`
	Tags []struct {
		Name string `json:"name"`
	} `json:"freeformTags"`
}

func (s gqlStream) model(owner *gqlUser) Stream {
	out := Stream{ID: s.ID, Title: s.Title, Viewers: s.Viewers, Preview: image(s.Preview), Game: s.Game.model()}
	out.StartedAt, _ = time.Parse(time.RFC3339, s.CreatedAt)
	if s.Broadcaster != nil {
		out.User = s.Broadcaster.model()
		if out.Title == "" {
			out.Title = s.Broadcaster.Settings.Title
		}
		out.Language = s.Broadcaster.Settings.Language
	} else if owner != nil {
		out.User = owner.model()
	}
	for _, t := range s.Tags {
		out.Tags = append(out.Tags, t.Name)
	}
	if s.Archive != nil && s.Archive.ID != "" {
		out.Archive = &StreamArchive{ID: s.Archive.ID}
		out.Archive.StartedAt, _ = time.Parse(time.RFC3339, s.Archive.CreatedAt)
	}
	return out
}

type gqlVideo struct {
	ID        string   `json:"id"`
	Title     string   `json:"title"`
	Length    int64    `json:"lengthSeconds"`
	Views     int64    `json:"viewCount"`
	Preview   string   `json:"previewThumbnailURL"`
	CreatedAt string   `json:"createdAt"`
	Type      string   `json:"broadcastType"`
	Status    string   `json:"status"`
	Animated  string   `json:"animatedPreviewURL"`
	Game      *gqlGame `json:"game"`
	Owner     *gqlUser `json:"owner"`
}

func (v gqlVideo) model(owner *gqlUser) Video {
	out := Video{
		ID: v.ID, Title: v.Title, Length: v.Length, Views: v.Views, Preview: image(v.Preview),
		Type: v.Type, Game: v.Game.model(), Storyboard: v.Animated, Recording: v.Status == "RECORDING",
	}
	out.CreatedAt, _ = time.Parse(time.RFC3339, v.CreatedAt)
	switch {
	case v.Owner != nil:
		out.Owner = v.Owner.model()
	case owner != nil:
		out.Owner = owner.model()
	}
	return out
}

type gqlClip struct {
	Slug        string   `json:"slug"`
	Title       string   `json:"title"`
	Duration    float64  `json:"durationSeconds"`
	Views       int64    `json:"viewCount"`
	Thumbnail   string   `json:"thumbnailURL"`
	CreatedAt   string   `json:"createdAt"`
	Offset      int64    `json:"videoOffsetSeconds"`
	Game        *gqlGame `json:"game"`
	Broadcaster *gqlUser `json:"broadcaster"`
	Video       *struct {
		ID string `json:"id"`
	} `json:"video"`
}

func (c gqlClip) model(owner *gqlUser) Clip {
	out := Clip{Slug: c.Slug, Title: c.Title, Duration: c.Duration, Views: c.Views, Thumbnail: image(c.Thumbnail), VideoOffset: c.Offset, Game: c.Game.model()}
	out.CreatedAt, _ = time.Parse(time.RFC3339, c.CreatedAt)
	if c.Video != nil {
		out.VideoID = c.Video.ID
	}
	switch {
	case c.Broadcaster != nil:
		out.Broadcaster = c.Broadcaster.model()
	case owner != nil:
		out.Broadcaster = owner.model()
	}
	return out
}

type pageInfo struct {
	HasNextPage bool `json:"hasNextPage"`
}

type connection[T any] struct {
	Edges []struct {
		Cursor string `json:"cursor"`
		Node   T      `json:"node"`
	} `json:"edges"`
	PageInfo pageInfo `json:"pageInfo"`
}

func toPage[N, M any](c connection[N], convert func(N) M) Page[M] {
	page := Page[M]{Items: make([]M, 0, len(c.Edges)), HasMore: c.PageInfo.HasNextPage}
	for _, e := range c.Edges {
		page.Items = append(page.Items, convert(e.Node))
		page.Cursor = e.Cursor
	}
	if !page.HasMore {
		page.Cursor = ""
	}
	return page
}
