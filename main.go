package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/pr1sm-dev/harbor-price-analytics/graph"
	"github.com/pr1sm-dev/harbor-price-analytics/tori"
)

func main() {
	args := os.Args[1:]
	searchQuery := strings.Join(args, " ")

	if len(searchQuery) == 0 {
		panic("Invalid search query")
	}

	tClient := tori.CreateToriClient(10 * time.Second)
	listings, err := tClient.GetQueryListings(searchQuery)
	if err != nil {
		panic(err)
	}

	if len(listings) == 0 {
		fmt.Println("0 results found for this query")
		os.Exit(1)
	}

	fmt.Printf("%d results found\n", len(listings))

	outputName := strings.Join(strings.Split(strings.ToLower(searchQuery), " "), "-")
	outputPath := fmt.Sprintf("./out/%s.html", outputName)

	fmt.Printf("Generating graph at %s\n", outputPath)

	graph.GenerateGraphs(listings, searchQuery, outputPath)
}
