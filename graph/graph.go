// Package graph handles the generate of a graph HTML file from listing data
package graph

import (
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/types"
	"github.com/pr1sm-dev/harbor-price-analytics/tori"
)

type GeoRoamVisitor struct {
	charts.BaseConfigurationVisitor
}

func (g GeoRoamVisitor) VisitGeo(geo opts.GeoComponent) any {
	type MyGeo struct {
		opts.GeoComponent

		Roam types.Bool `json:"roam"`
	}

	return &MyGeo{geo, opts.Bool(true)}
}

func createFinlandMap(listings tori.ToriQueryListings, searchQuery string) *charts.Geo {
	sc := charts.NewGeo()

	title := fmt.Sprintf("Location of %d items from search %q", len(listings), searchQuery)

	sc.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: title}),

		charts.WithGeoComponentOpts(opts.GeoComponent{Map: "芬兰"}),
	)

	geoData := make([]opts.GeoData, 0)
	for _, l := range listings {
		geoData = append(geoData, opts.GeoData{Name: l.Title, Value: []float64{l.Coordinates.Longitude, l.Coordinates.Latitude, float64(l.Price.Amount)}})
	}

	sc.AddSeries("tori.fi", types.ChartScatter, geoData).SetSeriesOptions(
		charts.WithSeriesOpts(func(s *charts.SingleSeries) {
			s.Roam = opts.Bool(true)
			s.CoordSystem = "geo"
			s.SymbolSize = 15
			s.Color = "#FF0000"
		}),
	)
	sc.Accept(&GeoRoamVisitor{})
	sc.JSAssets.Add("finland.js")

	return sc
}

func createPriceGraph(listings tori.ToriQueryListings) *charts.Line {
	lc := charts.NewLine()

	subtitle := fmt.Sprintf("Mean Price %.2f%s | Median Price %.2f%s", listings.MeanPrice(), listings[0].Price.Unit, listings.MedianPrice(), listings[0].Price.Unit)
	title := "Price over time of items"
	lc.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: title, Subtitle: subtitle}),
		charts.WithXAxisOpts(opts.XAxis{Name: "Time of Listing", NameLocation: "middle", Type: "time", NameGap: 40}),
		charts.WithYAxisOpts(opts.YAxis{Name: "Price in EUR", NameLocation: "middle", NameGap: 40}),
	)

	lineData := make([]opts.LineData, 0)
	sort.Sort(listings)

	for _, l := range listings {
		lineData = append(lineData, opts.LineData{Name: l.Title, Value: []int64{l.Timestamp / 1000, int64(l.Price.Amount)}})
	}

	lc.AddSeries("tori.fi", lineData).SetSeriesOptions(
		charts.WithSeriesOpts(func(s *charts.SingleSeries) {
			s.Color = "#FF0000"
		}),
	)

	return lc
}

func GenerateGraphs(listings tori.ToriQueryListings, search, outputPath string) {
	p := components.NewPage()

	pageTitle := fmt.Sprintf("Results from search %q, %d items found", search, len(listings))
	p.SetPageTitle(pageTitle)

	p.AddCharts(createFinlandMap(listings, search), createPriceGraph(listings))
	f, err := os.Create(outputPath)
	if err != nil {
		panic(err)
	}

	p.SetAssetsHost("./")

	p.Render(io.MultiWriter(f))
}
