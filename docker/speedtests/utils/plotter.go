package main

import (
	"encoding/csv"
	"fmt"
	"image/color"
	"log"
	"os"
	"regexp"
	"sort"
	"strconv"
	"time"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
)

func main() {
	var filterRegex *regexp.Regexp
	var title string
	var fileSource string

	switch len(os.Args) {
	case 3:
		fileSource = os.Args[1]
		title = os.Args[2]
		filterRegex = nil
		break

	case 4:
		fileSource = os.Args[1]
		title = os.Args[2]
		filterRegex = regexp.MustCompile(os.Args[3])
		break

	default:
		panic("Usage: plotter <csv file> <title> [event-filter-regex]")
	}

	f, err := os.OpenFile(fileSource, os.O_RDONLY, 0777)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	traces := make(map[string]plotter.XYs)
	connections := []string{}

	rd := csv.NewReader(f)
	rd.Comma = ','

	// xticks defines how we convert and display time.Time values.
	xticks := plot.TimeTicks{
		Format: time.RFC3339Nano,
		Time: func(t float64) time.Time {
			return time.UnixMilli(int64(t))
		},
	}

	record, err := rd.Read()
	for err == nil {
		record, err = rd.Read()
		if len(record) != 3 {
			continue
		}

		tag := record[1]
		if filterRegex != nil && !filterRegex.MatchString(tag) {
			continue
		}
		timestamp, errParse := time.Parse(time.RFC3339Nano, record[0])
		if errParse != nil {
			continue
		}
		speedValue, errParse := strconv.ParseFloat(record[2], 64)
		if errParse != nil {
			continue
		}

		if traces[tag] == nil {
			traces[tag] = make(plotter.XYs, 0, 256)
			connections = append(connections, tag)
		}
		traces[tag] = append(traces[tag], plotter.XY{
			X: float64(timestamp.UnixMilli()),
			Y: speedValue / 1000.0,
		})
	}

	fmt.Printf("ordering data\n")
	for _, tag := range connections {
		sort.SliceStable(traces[tag], func(i, j int) bool {
			return traces[tag][i].X < traces[tag][j].X
		})
	}

	fmt.Printf("start plotter\n")
	p := plot.New()
	p.Title.Text = title
	p.X.Tick.Marker = xticks
	p.X.Label.Text = "Time"
	p.Y.Label.Text = "Connection download speed (KB/s)"
	p.Add(plotter.NewGrid())

	sort.Strings(connections)

	//legend := plot.NewLegend()
	//legend.Padding = vg.Millimeter
	//legend.YPosition = draw.PosTop
	//
	for _, key := range connections {
		fmt.Printf("add trace %v\n", key)
		line, points, err := plotter.NewLinePoints(traces[key])
		if err != nil {
			log.Panic(err)
		}

		line.Color = color.RGBA{G: 255, A: 255}
		points.Shape = draw.CircleGlyph{}
		points.Color = color.RGBA{R: 255, A: 255}

		//legend.Add("")

		p.Add(line, points)
		fmt.Printf("added trace %v\n", key)
	}

	fmt.Printf("saving\n")
	err = p.Save(50*vg.Centimeter, 40*vg.Centimeter, "data.png")
	if err != nil {
		log.Panic(err)
	}
	fmt.Printf("done\n")
}
