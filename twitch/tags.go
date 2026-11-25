package twitch

import (
	"context"
	"unicode"
	"unicode/utf8"
)

// Twitch tags are one word of letters and digits in any script, at most 25 characters.
const maxTagLength = 25

func ValidTag(tag string) bool {
	if tag == "" || utf8.RuneCountInString(tag) > maxTagLength {
		return false
	}
	for _, r := range tag {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

func (c *Client) TagStreams(ctx context.Context, tag string, languages []string, first int, after string) (Page[Stream], error) {
	vars := languageVars(pageVars(first, after), languages)
	vars["tags"] = []string{tag}
	var data struct {
		Streams connection[gqlStream] `json:"streams"`
	}
	if err := c.raw(ctx, "", queryTagStreams, vars, &data); err != nil {
		return Page[Stream]{}, err
	}
	return toPage(data.Streams, func(s gqlStream) Stream { return s.model(nil) }), nil
}
