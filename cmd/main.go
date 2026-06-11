package main

import (
	"fmt"

	"github.com/ibu-i/webgraph-crawler/internal/domain"
)

func main() {
	url := "https://visitor.gogatsusai.jp"
	maxDepth := 2

	webGraph := domain.NewWebGraph(maxDepth)
	webGraph.Make(url, 0)

	fmt.Printf("Crawled %d pages starting from %s\n", webGraph.Nodes().Len(), url)
}
