package domain

import (
	"github.com/ibu-i/webgraph-crawler/internal/utils"
	"gonum.org/v1/gonum/graph/simple"
	"sync"
)

type WebGraph struct {
	*simple.DirectedGraph

	baseHost string

	urlToID  map[string]int64
	idToURL  map[int64]string
	nextID   int64
	maxDepth int

	mu sync.Mutex
}

func NewWebGraph(maxDepth int) *WebGraph {
	return &WebGraph{
		DirectedGraph: simple.NewDirectedGraph(),
		urlToID:       make(map[string]int64),
		idToURL:       make(map[int64]string),
		nextID:        1,
		maxDepth:      maxDepth,
	}
}

func (wg *WebGraph) AddPage(url string) (int64, bool) {
	wg.mu.Lock()
	defer wg.mu.Unlock()

	if id, exists := wg.urlToID[url]; exists {
		return id, true
	}

	id := wg.nextID
	wg.nextID++

	wg.urlToID[url] = id
	wg.idToURL[id] = url
	wg.AddNode(simple.Node(id))

	return id, false
}

func (wg *WebGraph) ensureBaseHost(webPageURL string) (string, bool) {
	wg.mu.Lock()
	defer wg.mu.Unlock()

	if wg.baseHost != "" {
		return wg.baseHost, true
	}

	host, err := utils.GetHost(webPageURL)
	if err != nil {
		return "", false
	}

	wg.baseHost = host
	return host, true
}

func (wg *WebGraph) addEdge(fromID, toID int64) {
	wg.mu.Lock()
	defer wg.mu.Unlock()

	wg.SetEdge(wg.NewEdge(simple.Node(fromID), simple.Node(toID)))
}

func (wg *WebGraph) Make(webPageURL string, depth int) {
	if depth > wg.maxDepth {
		return
	}

	baseHost, ok := wg.ensureBaseHost(webPageURL)
	if !ok {
		return
	}

	c := NewCollyCrawler(webPageURL)
	if c == nil {
		return
	}

	webPage, err := c.Crawl()

	if err != nil {
		return
	}

	pageID, _ := wg.AddPage(webPage.URL)

	var tasks sync.WaitGroup
	for link := range webPage.Links {
		host, err := utils.GetHost(link)
		if err != nil || host != baseHost {
			continue
		}

		linkID, visited := wg.AddPage(link)
		if linkID != pageID {
			wg.addEdge(pageID, linkID)
		}

		if !visited {
			tasks.Add(1)
			go func(link string) {
				defer tasks.Done()
				wg.Make(link, depth+1)
			}(link)
		}
	}

	tasks.Wait()
}
