package feed

import (
	"encoding/xml"
	"fmt"
	"html"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

// Item represents a news item from an RSS/Atom feed
type Item struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Published time.Time `json:"published"`
	Link      string    `json:"link"`
	FeedName  string    `json:"feed_name"`
}

// RSS structs for parsing RSS feeds
type RSS struct {
	Channel Channel `xml:"channel"`
}

type Channel struct {
	Items []RSSItem `xml:"item"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
	Link        string `xml:"link"`
	GUID        string `xml:"guid"`
}

// Atom structs for parsing Atom feeds
type Feed struct {
	Entries []AtomEntry `xml:"entry"`
}

type AtomEntry struct {
	ID      string `xml:"id"`
	Title   string `xml:"title"`
	Summary struct {
		Content string `xml:",chardata"`
		Type    string `xml:"type,attr"`
	} `xml:"summary"`
	Content struct {
		Content string `xml:",chardata"`
		Type    string `xml:"type,attr"`
	} `xml:"content"`
	Published string     `xml:"published"`
	Updated   string     `xml:"updated"`
	Links     []AtomLink `xml:"link"`
}

type AtomLink struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr"`
}

// Storage interface for caching (to avoid circular imports)
type CacheStorage interface {
	GetCacheFile(url string, maxAge time.Duration) ([]byte, bool)
	SetCacheFile(url string, data []byte) error
}

// ParseFeed fetches and parses an RSS or Atom feed (no caching)
func ParseFeed(url string) ([]Item, error) {
	return ParseFeedWithStorage(url, nil)
}

// ParseFeedWithStorage fetches and parses an RSS or Atom feed with optional caching
func ParseFeedWithStorage(url string, storage CacheStorage) ([]Item, error) {
	var body []byte

	// Try to get from cache first if storage is provided
	if storage != nil {
		if cachedData, found := storage.GetCacheFile(url, 15*time.Minute); found {
			body = cachedData
		}
	}

	// If we don't have cached data, fetch from HTTP
	if body == nil {
		resp, err := http.Get(url)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch feed: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("HTTP error: %d", resp.StatusCode)
		}

		// Read response body
		buf := make([]byte, 1024)
		for {
			n, err := resp.Body.Read(buf)
			if n > 0 {
				body = append(body, buf[:n]...)
			}
			if err != nil {
				break
			}
		}

		// Cache the data if storage is provided
		if storage != nil {
			if err := storage.SetCacheFile(url, body); err != nil {
				// Don't fail on cache errors, just log and continue
				fmt.Fprintf(os.Stderr, "Warning: Failed to cache feed data: %v\n", err)
			}
		}
	}

	// Try to determine if it's RSS or Atom by looking at the content
	bodyStr := string(body)
	if strings.Contains(bodyStr, "<rss") || strings.Contains(bodyStr, "<channel") {
		return parseRSS(body)
	} else if strings.Contains(bodyStr, "<feed") || strings.Contains(bodyStr, "atom") {
		return parseAtom(body)
	}

	// Default to trying RSS first, then Atom
	if items, err := parseRSS(body); err == nil && len(items) > 0 {
		return items, nil
	}

	return parseAtom(body)
}

func parseRSS(data []byte) ([]Item, error) {
	var rss RSS
	if err := xml.Unmarshal(data, &rss); err != nil {
		return nil, fmt.Errorf("failed to parse RSS: %w", err)
	}

	var items []Item
	for _, rssItem := range rss.Channel.Items {
		// Parse publication date
		pubTime, err := parseTime(rssItem.PubDate)
		if err != nil {
			// Skip items with invalid dates
			continue
		}

		// Clean up description/content
		content := cleanHTML(rssItem.Description)

		// Use GUID as ID, fallback to link
		id := rssItem.GUID
		if id == "" {
			id = rssItem.Link
		}

		item := Item{
			ID:        id,
			Title:     html.UnescapeString(rssItem.Title),
			Content:   content,
			Published: pubTime,
			Link:      rssItem.Link,
		}

		items = append(items, item)
	}

	return items, nil
}

func parseAtom(data []byte) ([]Item, error) {
	var feed Feed
	if err := xml.Unmarshal(data, &feed); err != nil {
		return nil, fmt.Errorf("failed to parse Atom: %w", err)
	}

	var items []Item
	for _, entry := range feed.Entries {
		// Parse publication date
		dateStr := entry.Published
		if dateStr == "" {
			dateStr = entry.Updated
		}
		pubTime, err := parseTime(dateStr)
		if err != nil {
			continue
		}

		// Get content - prefer content over summary
		content := entry.Content.Content
		if content == "" {
			content = entry.Summary.Content
		}
		content = cleanHTML(content)

		// Get link
		var link string
		for _, atomLink := range entry.Links {
			if atomLink.Rel == "alternate" || atomLink.Rel == "" {
				link = atomLink.Href
				break
			}
		}

		item := Item{
			ID:        entry.ID,
			Title:     html.UnescapeString(entry.Title),
			Content:   content,
			Published: pubTime,
			Link:      link,
		}

		items = append(items, item)
	}

	return items, nil
}

// parseTime attempts to parse various time formats commonly used in feeds
func parseTime(timeStr string) (time.Time, error) {
	timeStr = strings.TrimSpace(timeStr)

	// Common RSS date format: "Mon, 02 Jan 2006 15:04:05 MST"
	if t, err := time.Parse(time.RFC1123, timeStr); err == nil {
		return t, nil
	}

	// Alternative RSS format: "Mon, 02 Jan 2006 15:04:05 -0700"
	if t, err := time.Parse(time.RFC1123Z, timeStr); err == nil {
		return t, nil
	}

	// Atom format: "2006-01-02T15:04:05Z"
	if t, err := time.Parse(time.RFC3339, timeStr); err == nil {
		return t, nil
	}

	// Alternative format: "2006-01-02T15:04:05-07:00"
	if t, err := time.Parse("2006-01-02T15:04:05-07:00", timeStr); err == nil {
		return t, nil
	}

	// Simple format: "2006-01-02 15:04:05"
	if t, err := time.Parse("2006-01-02 15:04:05", timeStr); err == nil {
		return t, nil
	}

	return time.Time{}, fmt.Errorf("unable to parse time: %s", timeStr)
}

// cleanHTML removes HTML tags and cleans up content for display
func cleanHTML(content string) string {
	// Remove HTML tags
	re := regexp.MustCompile(`<[^>]*>`)
	content = re.ReplaceAllString(content, "")

	// Unescape HTML entities
	content = html.UnescapeString(content)

	// Clean up whitespace
	content = strings.TrimSpace(content)

	// Replace multiple newlines with double newline
	re = regexp.MustCompile(`\n\s*\n\s*\n`)
	content = re.ReplaceAllString(content, "\n\n")

	return content
}
