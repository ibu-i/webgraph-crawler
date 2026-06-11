package domain

import (
	"github.com/gocolly/colly/v2"
	"github.com/ibu-i/webgraph-crawler/internal/utils"
)

type WebPage struct {
	URL   string
	Title string
	Links map[string]struct{}
}

type CollyCrawler struct {
	URL string
}

func NewCollyCrawler(url string) *CollyCrawler {
	sanitizedURL, err := utils.SanitizeURL(url)
	if err != nil {
		return nil
	}
	url = sanitizedURL
	return &CollyCrawler{URL: url}
}

func (crawler *CollyCrawler) Crawl() (WebPage, error) {
	c := colly.NewCollector()
	var webPage WebPage

	c.OnHTML("html", func(e *colly.HTMLElement) {
		webPage.URL = crawler.URL
		webPage.Title = e.ChildText("title")
		webPage.Links = make(map[string]struct{})

		e.ForEach("a[href]", func(_ int, el *colly.HTMLElement) {
			link := e.Request.AbsoluteURL(el.Attr("href"))
			sanitizedLink, err := utils.SanitizeURL(link)
			if err != nil {
				return
			}
			webPage.Links[sanitizedLink] = struct{}{}
		})
	})

	err := c.Visit(crawler.URL)
	if err != nil {
		return WebPage{}, err
	}

	return webPage, nil
}
