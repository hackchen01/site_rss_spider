package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// RSS数据结构定义
type RSSFeed struct {
	XMLName xml.Name `xml:"rss"`
	Version string   `xml:"version,attr"`
	Channel Channel  `xml:"channel"`
}

type Channel struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	Items       []Item `xml:"item"`
}

type Item struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate,omitempty"`
	GUID        string `xml:"guid"`
}

// 网站配置
type SiteConfig struct {
	Name          string
	URL           string
	ItemSelector  string
	TitleSelector string
	LinkSelector  string
	DescSelector  string
	DateSelector  string
	DateFormat    string
}

// 缓存结构
type FeedCache struct {
	Feed     RSSFeed
	ExpireAt time.Time
}

var (
	cache     = make(map[string]FeedCache)
	cacheLock sync.RWMutex
)

// 初始化缓存
func initCache() {
	var sites []string
	for s, _ := range getAllSiteConfig() {
		sites = append(sites, s)
	}

	for _, site := range sites {
		go refreshCache(site)
	}

	// 设置定时器，每10分钟刷新一次所有缓存
	ticker := time.NewTicker(10 * time.Minute)
	go func() {
		for range ticker.C {
			for _, site := range sites {
				go refreshCache(site)
			}
		}
	}()
}

// 刷新指定网站的缓存
func refreshCache(site string) {
	log.Printf("Refreshing cache for site: %s", site)

	feed, err := fetchAndGenerateRSS(site)
	if err != nil {
		log.Printf("Failed to refresh cache for %s: %v", site, err)
		return
	}

	cacheLock.Lock()
	cache[site] = FeedCache{
		Feed:     feed,
		ExpireAt: time.Now().Add(10 * time.Minute),
	}
	cacheLock.Unlock()

	log.Printf("Cache refreshed for site: %s", site)
}

// 生成RSS的HTTP处理函数
func generateRSSHandler(w http.ResponseWriter, r *http.Request) {
	site := r.URL.Query().Get("site")
	if site == "" {
		http.Error(w, "Missing 'site' parameter", http.StatusBadRequest)
		return
	}

	// 检查缓存
	cacheLock.RLock()
	cached, ok := cache[site]
	cacheLock.RUnlock()

	// 如果缓存存在且未过期，直接返回
	if ok && time.Now().Before(cached.ExpireAt) {
		w.Header().Set("Content-Type", "application/rss+xml")
		xml.NewEncoder(w).Encode(cached.Feed)
		return
	}

	// 如果缓存不存在或已过期，返回现有缓存（如果有）并异步刷新
	if ok {
		// 返回旧缓存
		go refreshCache(site)
		w.Header().Set("Content-Type", "application/rss+xml")
		xml.NewEncoder(w).Encode(cached.Feed)
		return
	}

	// 首次请求，同步获取
	feed, err := fetchAndGenerateRSS(site)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to generate RSS: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/rss+xml")
	xml.NewEncoder(w).Encode(feed)
}

// 获取所有网站配置
func getAllSiteConfig() map[string]SiteConfig {
	return map[string]SiteConfig{
		"example": {
			Name:          "示例网站",
			URL:           "https://example.com",
			ItemSelector:  "article h2",
			TitleSelector: "article h2",
			LinkSelector:  "article a",
			DescSelector:  "article p.summary",
			DateSelector:  "article time",
			DateFormat:    "2006-01-02",
		},
		"abc": {
			Name:          "abc网站",
			URL:           "https://www.abc.com/",
			ItemSelector:  ".content article",
			TitleSelector: "header a",
			LinkSelector:  "header a",
			DescSelector:  "p.note",
			DateSelector:  "div.meta time",
			DateFormat:    "2006-01-02",
		},
	}
}

// 获取网站配置
func getSiteConfig(site string) (SiteConfig, bool) {
	configs := getAllSiteConfig()

	config, exists := configs[site]
	return config, exists
}

// 抓取内容并生成RSS
func fetchAndGenerateRSS(site string) (RSSFeed, error) {
	config, exists := getSiteConfig(site)
	if !exists {
		return RSSFeed{}, fmt.Errorf("site configuration not found: %s", site)
	}

	doc, err := goquery.NewDocument(config.URL)
	if err != nil {
		return RSSFeed{}, err
	}

	var items []Item

	doc.Find(config.ItemSelector).Each(func(i int, s *goquery.Selection) {
		title := s.Find(config.TitleSelector).Text()

		link, _ := s.Find(config.LinkSelector).Attr("href")
		if !strings.HasPrefix(link, "http") {
			link = config.URL + link
		}

		desc := s.Find(config.DescSelector).Text()

		dateStr := s.Find(config.DateSelector).Text()
		var pubDate string
		if dateStr != "" {
			t, err := time.Parse(config.DateFormat, dateStr)
			if err == nil {
				pubDate = t.Format("2006-01-02 15:04:05")
			}
		}

		if title != "" && link != "" {
			items = append(items, Item{
				Title:       title,
				Link:        link,
				Description: desc,
				PubDate:     pubDate,
				GUID:        link,
			})
		}
	})

	feed := RSSFeed{
		Version: "2.0",
		Channel: Channel{
			Title:       config.Name,
			Link:        config.URL,
			Description: fmt.Sprintf("RSS feed for %s", config.Name),
			Items:       items,
		},
	}

	return feed, nil
}

func main() {
	// 解析命令行参数获取端口号
	port := flag.String("port", "8080", "Server port")
	flag.Parse()

	// 初始化缓存
	initCache()

	http.HandleFunc("/rss", generateRSSHandler)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "RSS生成服务已启动！\n使用方法: /rss?site=example")
	})

	log.Printf("Server started on :%s", *port)
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}
