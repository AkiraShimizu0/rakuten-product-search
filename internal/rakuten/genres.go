package rakuten

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
)

const GenreEndpoint = "https://openapi.rakuten.co.jp/ichibagt/api/IchibaGenre/Search/20260701"

type Genre struct {
	ID    json.Number `json:"genreId"`
	Name  string      `json:"nameJa"`
	Level int         `json:"level"`
}
type GenrePage struct {
	Genre    Genre   `json:"genre"`
	Children []Genre `json:"children"`
}

func (c *Client) Genres(ctx context.Context, id string) (GenrePage, error) {
	u, _ := url.Parse(GenreEndpoint)
	q := u.Query()
	q.Set("applicationId", c.cfg.AppID)
	q.Set("genreId", id)
	q.Set("formatVersion", "2")
	u.RawQuery = q.Encode()
	b, e := c.getJSON(ctx, u, 0)
	if e != nil {
		return GenrePage{}, e
	}
	var p GenrePage
	if json.Unmarshal(b, &p) != nil {
		return p, errors.New("invalid genre JSON")
	}
	return p, nil
}

// Availability refers to the API's explicit purchasability flag, never stock quantity.
func Availability(raw json.RawMessage) string {
	var v map[string]json.RawMessage
	if json.Unmarshal(raw, &v) != nil {
		return "unknown"
	}
	if b, ok := v["availability"]; ok {
		s := string(b)
		if s == "1" || s == `"1"` {
			return "api_available"
		}
		if s == "0" || s == `"0"` {
			return "api_unavailable"
		}
	}
	return "unknown"
}
func GenreID(g Genre) string { return fmt.Sprint(g.ID) }
