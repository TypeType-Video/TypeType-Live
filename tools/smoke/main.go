package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/TypeType-Video/TypeType-Live/proxy"
	"github.com/TypeType-Video/TypeType-Live/twitch"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	c := twitch.New(proxy.NewHTTPClient(proxy.NewHosts(proxy.DefaultAllowedSuffixes)), "")
	login := "riotgames"
	if len(os.Args) > 1 {
		login = os.Args[1]
	}
	check := func(name string, err error, detail string) {
		if err != nil {
			fmt.Printf("FAIL %-16s %v\n", name, err)
			return
		}
		fmt.Printf("ok   %-16s %s\n", name, detail)
	}
	streams, err := c.TopStreams(ctx, nil, 3, "")
	check("TopStreams", err, fmt.Sprintf("%d items, more=%v", len(streams.Items), streams.HasMore))
	games, err := c.TopGames(ctx, 3, "")
	check("TopGames", err, fmt.Sprintf("%d items", len(games.Items)))
	if len(games.Items) > 0 {
		g, gs, err := c.GameStreams(ctx, games.Items[0].Slug, nil, 2, "")
		name := ""
		if g != nil {
			name = g.Name
		}
		check("GameStreams", err, fmt.Sprintf("%s %d items", name, len(gs.Items)))
	}
	ch, err := c.Channel(ctx, login)
	check("Channel", err, fmt.Sprintf("%s followers=%d live=%v", ch.DisplayName, ch.Followers, ch.Live != nil))
	videos, err := c.ChannelVideos(ctx, login, 2, "")
	check("ChannelVideos", err, fmt.Sprintf("%d items", len(videos.Items)))
	clips, err := c.ChannelClips(ctx, login, 2, "")
	check("ChannelClips", err, fmt.Sprintf("%d items", len(clips.Items)))
	if len(videos.Items) > 0 {
		v, err := c.Video(ctx, videos.Items[0].ID)
		check("Video", err, fmt.Sprintf("len=%ds type=%s", v.Length, v.Type))
	}
	if len(clips.Items) > 0 {
		_, qs, err := c.Clip(ctx, clips.Items[0].Slug)
		check("Clip", err, fmt.Sprintf("%d qualities", len(qs)))
	}
	sc, err := c.SearchChannels(ctx, "chess", 2, "")
	check("SearchChannels", err, fmt.Sprintf("%d items", len(sc.Items)))
	sg, err := c.SearchGames(ctx, "chess", 2, "")
	check("SearchGames", err, fmt.Sprintf("%d items", len(sg.Items)))
	ss, err := c.SearchStreams(ctx, "chess", 2, "")
	check("SearchStreams", err, fmt.Sprintf("%d items", len(ss.Items)))
	tok, err := c.LiveToken(ctx, login, twitch.DefaultPlayer, "")
	check("LiveToken", err, fmt.Sprintf("expires in %s, server_ads=%v", time.Until(tok.Expires).Round(time.Second), tok.ServerAds))
}
