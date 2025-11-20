package main

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mrbrist/go-blog-aggregator/internal/database"
)

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "gator")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, errors.New("Status code: " + strconv.Itoa(res.StatusCode))
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var feed RSSFeed
	err = xml.Unmarshal(data, &feed)
	if err != nil {
		return nil, err
	}

	feed.Channel.Title = html.UnescapeString(feed.Channel.Title)
	feed.Channel.Description = html.UnescapeString(feed.Channel.Description)

	for _, item := range feed.Channel.Item {
		item.Title = html.UnescapeString(item.Title)
		item.Description = html.UnescapeString(item.Description)
	}

	return &feed, nil
}

func scrapeFeeds(s *state) {
	next, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		fmt.Println(err)
		return
	}

	err = s.db.MarkFeedFetched(context.Background(), next.ID)
	if err != nil {
		fmt.Println(err)
		return
	}

	feed, err := fetchFeed(context.Background(), next.Url)
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, i := range feed.Channel.Item {
		pubTime, err := time.Parse(time.RFC1123Z, i.PubDate)
		if err != nil {
			return
		}
		post, err := s.db.CreatePost(context.Background(), database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Title:       i.Title,
			Url:         i.Link,
			Description: i.Description,
			PublishedAt: pubTime,
			FeedID:      next.ID,
		})
		if err != nil {
			if strings.Contains(err.Error(), "unique") ||
				strings.Contains(err.Error(), "duplicate") {
				continue
			}

			// Log all other errors
			fmt.Printf("error saving post %q: %v\n", i.Title, err)
			continue
		}
		fmt.Printf("Saved post: %s\n", post.Title)
	}
}
