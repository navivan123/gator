package main

import (
    "net/http"
    "encoding/xml"
    "io"
    "context"
    "fmt"
    "html"
)



func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
    req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
    if err != nil {
        return nil, fmt.Errorf("Error formulating request: %v", err)
    }

    req.Header.Add("User-Agent", "gator")

    client    := http.DefaultClient
    resp, err := client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("Error while fetching aggregation data: %v", err)
    }
    defer resp.Body.Close()

    data, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("Error while reading aggregation data: %v", err)
    }

    rss := RSSFeed{}
    err = xml.Unmarshal(data, &rss)
    if err != nil {
        return nil, fmt.Errorf("Error while parsing xml to struct: %v", err)
    }
    
    // Unescape HTML entities
    rss.Channel.Title       = html.UnescapeString(rss.Channel.Title)
    rss.Channel.Description = html.UnescapeString(rss.Channel.Description)
    for _, item := range rss.Channel.Item {
        item.Title       = html.UnescapeString(item.Title)
        item.Description = html.UnescapeString(item.Description)
    }

    return &rss, nil

}
