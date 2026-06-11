package domain

import (
	"fmt"
	"net/url"
	"os"
	"slices"
	"sync"

	"github.com/ibu-i/webgraph-crawler/internal/utils"
	"gonum.org/v1/gonum/graph/simple"
)

const defaultCrawlConcurrency = 8

type WebGraph struct {
	*simple.DirectedGraph

	mu sync.Mutex

	crawlSem chan struct{}

	baseHost string

	urlToID   map[string]int64
	idToURL   map[int64]string
	idToDepth map[int64]int
	nextID    int64
	maxDepth  int
}

func NewWebGraph(maxDepth int) *WebGraph {
	return &WebGraph{
		DirectedGraph: simple.NewDirectedGraph(),
		crawlSem:      make(chan struct{}, defaultCrawlConcurrency),
		urlToID:       make(map[string]int64),
		idToURL:       make(map[int64]string),
		idToDepth:     make(map[int64]int),
		nextID:        1,
		maxDepth:      maxDepth,
	}
}

func (wg *WebGraph) AddPage(url string, depth int) (int64, bool) {
	wg.mu.Lock()
	defer wg.mu.Unlock()

	if id, exists := wg.urlToID[url]; exists {
		if currentDepth, ok := wg.idToDepth[id]; !ok || depth < currentDepth {
			wg.idToDepth[id] = depth
		}
		return id, true
	}

	id := wg.nextID
	wg.nextID++

	wg.urlToID[url] = id
	wg.idToURL[id] = url
	wg.idToDepth[id] = depth
	wg.AddNode(simple.Node(id))

	return id, false
}

func (wg *WebGraph) ensureBaseHost(webPageURL string) bool {
	wg.mu.Lock()
	defer wg.mu.Unlock()

	if wg.baseHost != "" {
		return true
	}

	host, err := utils.GetHost(webPageURL)
	if err != nil {
		return false
	}

	wg.baseHost = host
	return true
}

func (wg *WebGraph) isBaseHost(host string) bool {
	wg.mu.Lock()
	defer wg.mu.Unlock()

	return host == wg.baseHost
}

func (wg *WebGraph) addEdge(fromID, toID int64) {
	wg.mu.Lock()
	defer wg.mu.Unlock()

	wg.SetEdge(wg.NewEdge(simple.Node(fromID), simple.Node(toID)))
}

func (wg *WebGraph) acquireCrawlSlot() {
	wg.crawlSem <- struct{}{}
}

func (wg *WebGraph) releaseCrawlSlot() {
	<-wg.crawlSem
}

func (wg *WebGraph) crawl(webPageURL string, depth int, crawlWG *sync.WaitGroup) {
	defer crawlWG.Done()

	wg.acquireCrawlSlot()
	defer wg.releaseCrawlSlot()

	if depth > wg.maxDepth {
		return
	}

	if !wg.ensureBaseHost(webPageURL) {
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

	pageID, _ := wg.AddPage(webPage.URL, depth)

	for link := range webPage.Links {
		host, err := utils.GetHost(link)
		if err != nil || !wg.isBaseHost(host) {
			continue
		}

		linkID, visited := wg.AddPage(link, depth+1)
		if linkID != pageID {
			wg.addEdge(pageID, linkID)
		}

		if visited {
			continue
		}

		crawlWG.Add(1)
		go wg.crawl(link, depth+1, crawlWG)
	}
}

func (wg *WebGraph) Make(webPageURL string, depth int) {
	var crawlWG sync.WaitGroup
	crawlWG.Add(1)
	go wg.crawl(webPageURL, depth, &crawlWG)
	crawlWG.Wait()
}

func (wg *WebGraph) ExportDOT(filename string) error {
	if err := os.MkdirAll("build/dot", 0755); err != nil {
		return err
	}

	f, err := os.Create("build/dot/" + filename)
	if err != nil {
		return err
	}
	defer f.Close()

	fmt.Fprintln(f, "digraph WebGraph {")
	fmt.Fprintln(f)
	fmt.Fprintln(f, "    rankdir=TB")
	fmt.Fprintln(f, "    node [shape=box]")
	fmt.Fprintln(f)

	nodes := wg.Nodes()
	depthToNodes := make(map[int][]int64)

	for nodes.Next() {
		node := nodes.Node()

		nodeURL := wg.idToURL[node.ID()]

		u, err := url.Parse(nodeURL)
		if err != nil {
			continue
		}

		label := u.Path
		if label == "" {
			label = "/"
		}

		depth := wg.idToDepth[node.ID()]
		depthToNodes[depth] = append(depthToNodes[depth], node.ID())

		fmt.Fprintf(
			f,
			`    N%d [label="%s"]`+"\n",
			node.ID(),
			label,
		)
	}

	fmt.Fprintln(f)

	depths := make([]int, 0, len(depthToNodes))
	for depth := range depthToNodes {
		depths = append(depths, depth)
	}
	slices.Sort(depths)

	for _, depth := range depths {
		nodeIDs := depthToNodes[depth]
		slices.Sort(nodeIDs)

		fmt.Fprintln(f, "    {")
		fmt.Fprintln(f, "        rank=same")
		for _, nodeID := range nodeIDs {
			fmt.Fprintf(f, "        N%d\n", nodeID)
		}
		fmt.Fprintln(f, "    }")
	}

	fmt.Fprintln(f)

	nodes = wg.Nodes()

	for nodes.Next() {
		from := nodes.Node()

		toNodes := wg.From(from.ID())

		for toNodes.Next() {
			to := toNodes.Node()

			fmt.Fprintf(
				f,
				"    N%d -> N%d\n",
				from.ID(),
				to.ID(),
			)
		}
	}

	fmt.Fprintln(f, "}")

	return nil
}
