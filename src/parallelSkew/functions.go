package main

import (
	"bufio"
	"bytes"
	"golang.org/x/exp/errors/fmt"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
	"os"
	"path"
)

// readFasta Returns the sequence from a FASTA file without comments and newlines.
func readFasta(filename string) (string, error) {
	buffer := bytes.NewBufferString("")
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" && line[0] != '>' { // discard comments
			buffer.WriteString(line)
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}
	return buffer.String(), nil
}

func SkewArraySerial(genome string) []int {
	a := make([]int, len(genome)+1)
	a[0] = 0
	for i := range genome {
		if genome[i] == 'C' {
			a[i+1] = a[i] - 1
		} else if genome[i] == 'G' {
			a[i+1] = a[i] + 1
		} else {
			a[i+1] = a[i]
		}
	}
	return a
}

func SkewArrayParallel(c chan Bacterium, bact Bacterium) {
	a := make([]int, len(bact.genome)+1)
	a[0] = 0
	for i := range bact.genome {
		if bact.genome[i] == 'C' {
			a[i+1] = a[i] - 1
		} else if bact.genome[i] == 'G' {
			a[i+1] = a[i] + 1
		} else {
			a[i+1] = a[i]
		}
	}
	bact.skewArray = a
	c <- bact // a channel
}

func (bact *Bacterium) ConvertToPNG() {
	skewPlot, _ := plot.New()

	skewPlot.Title.Text = "Skew array for " + bact.name
	skewPlot.X.Label.Text = "Genome position"
	skewPlot.Y.Label.Text = "Skew"

	_ = plotutil.AddLinePoints(skewPlot, "", PlotArray(bact.skewArray))
	skewPlot.Save(4*vg.Inch, 4*vg.Inch, path.Join("plots", bact.name+".png"))
	fmt.Println("  - generated plot for", bact.name)
}

func PlotArray(a []int) plotter.XYs {
	n := len(a)
	points := make(plotter.XYs, n)
	for i := range points {
		points[i].X = float64(i)
		points[i].Y = float64(a[i])
	}
	return points
}
