package main

import (
	"fmt"

	"github.com/ibu-i/webgraph-crawler/internal/domain"
)

func main() {
	url := "https://ja.wikipedia.org/wiki/%E3%83%A1%E3%82%A4%E3%83%B3%E3%83%9A%E3%83%BC%E3%82%B8"
	maxDepth := 2

	webGraph := domain.NewWebGraph(maxDepth)
	webGraph.Make(url, 0)

	fmt.Printf("Crawled %d pages starting from %s\n", webGraph.Nodes().Len(), url)

	err := webGraph.ExportDOT("graph.dot")
	if err != nil {
		panic(err)
	}
}
