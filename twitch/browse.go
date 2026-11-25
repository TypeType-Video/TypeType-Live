package twitch

import "context"

const maxPage = 50

func pageVars(first int, after string) map[string]any {
	if first <= 0 || first > maxPage {
		first = 24
	}
	vars := map[string]any{"first": first}
	if after != "" {
		vars["after"] = after
	}
	return vars
}

func (c *Client) raw(ctx context.Context, userToken, query string, vars map[string]any, out any) error {
	return c.gql(ctx, userToken, nil, gqlRequest{Query: query, Variables: vars}, out)
}

func (c *Client) TopStreams(ctx context.Context, languages []string, first int, after string) (Page[Stream], error) {
	var data struct {
		Streams connection[gqlStream] `json:"streams"`
	}
	if err := c.raw(ctx, "", queryTopStreams, languageVars(pageVars(first, after), languages), &data); err != nil {
		return Page[Stream]{}, err
	}
	return toPage(data.Streams, func(s gqlStream) Stream { return s.model(nil) }), nil
}

func (c *Client) TopGames(ctx context.Context, first int, after string) (Page[Game], error) {
	var data struct {
		Games connection[gqlGame] `json:"games"`
	}
	if err := c.raw(ctx, "", queryTopGames, pageVars(first, after), &data); err != nil {
		return Page[Game]{}, err
	}
	return toPage(data.Games, func(g gqlGame) Game { return *g.model() }), nil
}

func (c *Client) GameStreams(ctx context.Context, slug string, languages []string, first int, after string) (*Game, Page[Stream], error) {
	vars := languageVars(pageVars(first, after), languages)
	vars["slug"] = slug
	var data struct {
		Game *struct {
			gqlGame
			Streams connection[gqlStream] `json:"streams"`
		} `json:"game"`
	}
	if err := c.raw(ctx, "", queryGameStreams, vars, &data); err != nil {
		return nil, Page[Stream]{}, err
	}
	if data.Game == nil {
		return nil, Page[Stream]{}, ErrNotFound
	}
	return data.Game.model(), toPage(data.Game.Streams, func(s gqlStream) Stream { return s.model(nil) }), nil
}

type gqlChannel struct {
	gqlUser
	Description string `json:"description"`
	Banner      string `json:"bannerImageURL"`
	Followers   struct {
		TotalCount int64 `json:"totalCount"`
	} `json:"followers"`
	Roles struct {
		IsPartner bool `json:"isPartner"`
	} `json:"roles"`
	Stream *gqlStream `json:"stream"`
}

func (ch gqlChannel) model() Channel {
	out := Channel{User: ch.gqlUser.model(), Description: ch.Description, Banner: image(ch.Banner), Followers: ch.Followers.TotalCount, Partner: ch.Roles.IsPartner}
	if ch.Stream != nil && ch.Stream.ID != "" {
		live := ch.Stream.model(&ch.gqlUser)
		out.Live = &live
	}
	return out
}

func (c *Client) Channel(ctx context.Context, login string) (Channel, error) {
	var data struct {
		User *gqlChannel `json:"user"`
	}
	if err := c.raw(ctx, "", queryChannel, map[string]any{"login": login}, &data); err != nil {
		return Channel{}, err
	}
	if data.User == nil {
		return Channel{}, ErrNotFound
	}
	return data.User.model(), nil
}

func (c *Client) ChannelVideos(ctx context.Context, login string, first int, after string) (Page[Video], error) {
	vars := pageVars(first, after)
	vars["login"] = login
	var data struct {
		User *struct {
			gqlUser
			Videos connection[gqlVideo] `json:"videos"`
		} `json:"user"`
	}
	if err := c.raw(ctx, "", queryChannelVideos, vars, &data); err != nil {
		return Page[Video]{}, err
	}
	if data.User == nil {
		return Page[Video]{}, ErrNotFound
	}
	owner := data.User.gqlUser
	return toPage(data.User.Videos, func(v gqlVideo) Video { return v.model(&owner) }), nil
}

func (c *Client) ChannelClips(ctx context.Context, login string, first int, after string) (Page[Clip], error) {
	vars := pageVars(first, after)
	vars["login"] = login
	var data struct {
		User *struct {
			gqlUser
			Clips connection[gqlClip] `json:"clips"`
		} `json:"user"`
	}
	if err := c.raw(ctx, "", queryChannelClips, vars, &data); err != nil {
		return Page[Clip]{}, err
	}
	if data.User == nil {
		return Page[Clip]{}, ErrNotFound
	}
	owner := data.User.gqlUser
	return toPage(data.User.Clips, func(cl gqlClip) Clip { return cl.model(&owner) }), nil
}

func (c *Client) Video(ctx context.Context, id string) (Video, error) {
	var data struct {
		Video *gqlVideo `json:"video"`
	}
	if err := c.raw(ctx, "", queryVideo, map[string]any{"id": id}, &data); err != nil {
		return Video{}, err
	}
	if data.Video == nil || data.Video.ID == "" {
		return Video{}, ErrNotFound
	}
	return data.Video.model(nil), nil
}
