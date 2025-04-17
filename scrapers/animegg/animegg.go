package animegg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

const baseURL = "https://www.animegg.org"

type Client struct {
	HTTPClient *http.Client
	Headers    http.Header
	Language   string
	Quality    string
	Server     string
}

type Anime struct {
	Title       string   `json:"title"`
	URL         string   `json:"url"`
	Thumbnail   string   `json:"thumbnail"`
	Description string   `json:"description,omitempty"`
	Status      string   `json:"status,omitempty"`
	Genres      []string `json:"genres,omitempty"`
}

type Episode struct {
	Number  float64  `json:"number"`
	Title   string   `json:"title"`
	URL     string   `json:"url"`
	Sources []string `json:"sources,omitempty"`
}

type Video struct {
	URL      string `json:"url"`
	Quality  string `json:"quality"`
	Language string `json:"language"`
	Server   string `json:"server"`
	Title    string `json:"title"`
}

type Link string

func (l Link) Canonical() string {
	return fmt.Sprintf("%s%s", baseURL, l)
}

func NewClient(lang, quality, server string) *Client {
	return &Client{
		HTTPClient: &http.Client{},
		Headers: http.Header{
			"User-Agent": []string{"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"},
			"Referer":    []string{baseURL},
		},
		Language: lang,
		Quality:  quality,
		Server:   server,
	}
}

func (c *Client) Search(query string) ([]Anime, error) {
	req, _ := http.NewRequest("GET", Link("/search").Canonical(), nil)
	q := req.URL.Query()
	q.Add("q", query)
	req.URL.RawQuery = q.Encode()
	req.Header = c.Headers

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	var results []Anime
	doc.Find(".mse").Each(func(i int, s *goquery.Selection) {
		title := s.Find(".first h2").Text()
		link, _ := s.Attr("href")
		img, _ := s.Find("img").Attr("src")

		results = append(results, Anime{
			Title:     strings.TrimSpace(title),
			URL:       Link(link).Canonical(),
			Thumbnail: img,
		})
	})

	return results, nil
}

func (c *Client) fixJSON(input string) (string, error) {
	// Regex to quote unquoted keys: foo: -> "foo":
	re := regexp.MustCompile(`(?m)(\s*)([a-zA-Z_][a-zA-Z0-9_]*)\s*:`)
	quoted := re.ReplaceAllString(input, `$1"$2":`)

	// Validate JSON
	var buf bytes.Buffer
	if err := json.Compact(&buf, []byte(quoted)); err != nil {
		return "", fmt.Errorf("invalid after fix: %w", err)
	}

	return buf.String(), nil
}

func (c *Client) Popular() ([]Anime, error) {
	return c.fetchAnimeList("hits", "DESC")
}

func (c *Client) Latest() ([]Anime, error) {
	return c.fetchAnimeList("createdAt", "DESC")
}

func (c *Client) fetchAnimeList(sortBy, direction string) ([]Anime, error) {
	req, _ := http.NewRequest("GET", Link("/popular-series").Canonical(), nil)
	q := req.URL.Query()
	q.Add("sortBy", sortBy)
	q.Add("sortDirection", direction)
	q.Add("ongoing", "")
	q.Add("limit", "50")
	q.Add("start", "0")
	req.URL.RawQuery = q.Encode()
	req.Header = c.Headers

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	var results []Anime
	doc.Find(".fea").Each(func(i int, s *goquery.Selection) {
		title := s.Find(".rightpop a").Text()
		link, _ := s.Find(".rightpop a").Attr("href")
		img, _ := s.Find("img").Attr("src")

		results = append(results, Anime{
			Title:     strings.TrimSpace(title),
			URL:       Link(link).Canonical(),
			Thumbnail: img,
		})
	})

	return results, nil
}

func (c *Client) Episodes(url string) ([]Episode, error) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header = c.Headers

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	var episodes []Episode
	doc.Find(".newmanga li div").Each(func(i int, s *goquery.Selection) {
		numStr := s.Find(".anm_det_pop strong").Text()
		num := parseEpisodeNumber(numStr)
		title := s.Find(".anititle").Text()
		link, _ := s.Find(".anm_det_pop").Attr("href")

		episodes = append(episodes, Episode{
			Number: num,
			Title:  fmt.Sprintf("Episode %s - %s", formatNumber(num), title),
			URL:    Link(link).Canonical(),
		})
	})

	return episodes, nil
}

func (c *Client) Videos(episodeURL string) ([]Video, error) {
	req, _ := http.NewRequest("GET", episodeURL, nil)
	req.Header = c.Headers

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	title := doc.Find("div.titleep > h1").Text()
	if title == "" {
		title = episodeURL
	} else {
		title = strings.TrimSpace(title)
	}

	var wg sync.WaitGroup
	var videos []Video
	var mu sync.Mutex

	doc.Find("iframe").Each(func(i int, iframe *goquery.Selection) {
		wg.Add(1)
		go func(src string) {
			defer wg.Done()

			iframeReq, _ := http.NewRequest("GET", Link(src).Canonical(), nil)
			iframeReq.Header = c.Headers.Clone()
			iframeReq.Header.Set("Referer", baseURL)

			resp, err := c.HTTPClient.Do(iframeReq)
			if err != nil {
				return
			}
			defer resp.Body.Close()

			doc, err := goquery.NewDocumentFromReader(resp.Body)
			if err != nil {
				return
			}

			script := doc.Find("script:contains('videoSources')").Text()
			jsonStr := extractJSON(script)
			jsonStr, err = c.fixJSON(jsonStr)
			if err != nil {
				return
			}

			var ggVideos []struct {
				File  string `json:"file"`
				Label string `json:"label"`
			}

			if err := json.Unmarshal([]byte(jsonStr), &ggVideos); err != nil {
				return
			}

			mu.Lock()
			defer mu.Unlock()

			for _, v := range ggVideos {
				videoURL := Link(v.File).Canonical()
				lang := extractLanguage(iframe)

				videos = append(videos, Video{
					URL:      videoURL,
					Quality:  extractQuality(v.Label),
					Language: lang,
					Server:   "AnimeGG",
					Title:    title,
				})
			}
		}(iframe.AttrOr("src", ""))
	})

	wg.Wait()
	sortVideos(videos, c.Language, c.Server, c.Quality)
	return videos, nil
}

func sortVideos(videos []Video, lang, server, quality string) {
	sort.Slice(videos, func(i, j int) bool {
		a, b := videos[i], videos[j]

		if a.Language != b.Language {
			return a.Language == lang
		}
		if a.Server != b.Server {
			return a.Server == server
		}
		aQual, _ := strconv.Atoi(a.Quality)
		bQual, _ := strconv.Atoi(b.Quality)
		return aQual >= bQual
	})
}

func extractJSON(script string) string {
	re := regexp.MustCompile(`var videoSources = (\[.*?\]);`)
	match := re.FindStringSubmatch(script)
	if len(match) > 1 {
		return strings.ReplaceAll(match[1], "'", "\"")
	}
	return ""
}

func extractQuality(label string) string {
	re := regexp.MustCompile(`(\d+)p`)
	match := re.FindStringSubmatch(label)
	if len(match) > 1 {
		return match[1]
	}
	return "unknown"
}

func extractLanguage(iframe *goquery.Selection) string {
	parent := iframe.ParentsFiltered(".tab-pane")
	id, _ := parent.Attr("id")
	switch {
	case strings.Contains(id, "subbed"):
		return "subbed"
	case strings.Contains(id, "dubbed"):
		return "dubbed"
	case strings.Contains(id, "raw"):
		return "raw"
	}
	return "unknown"
}

func parseEpisodeNumber(s string) float64 {
	re := regexp.MustCompile(`(\d+(?:\.\d+)?)`)
	match := re.FindStringSubmatch(s)
	if len(match) == 0 {
		return 0
	}
	num, _ := strconv.ParseFloat(match[1], 64)
	return num
}

func formatNumber(n float64) string {
	if n == float64(int(n)) {
		return fmt.Sprintf("%d", int(n))
	}
	return fmt.Sprintf("%.1f", n)
}
