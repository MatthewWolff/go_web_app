package main

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"html/template"
	"image/color"
	"io"
	"net/http"
	"os"
	"strings"
)

// Page a small struct to help us use HTML Templates
type Page struct {
	Title    string
	Contents template.HTML
}

// indexHandler Basic index handler for root
func indexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method:", r.Method, "on", "index") //get request method
	if r.Method == "GET" {
		t, err := template.ParseFiles("site/templates/index.html")
		if err != nil {
			panic(err)
		}
		t.Execute(w, nil) // we could pass a struct in to apply formatting if we wanted
	}
}

// getUrlHandler Handles get_url request
func getUrlHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method:", r.Method, "on", "get_url") //get request method
	if r.Method == "GET" {
		t, err := template.ParseFiles("site/templates/get_url.html")
		if err != nil {
			panic(err)
		}
		t.Execute(w, nil) // we could pass a struct in to apply formatting if we wanted
	}
}

// minskewHandler Handles the /minskew request
func minskewHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method:", r.Method, "on", "minSkew") //get request method
	if r.Method == "GET" {
		r.ParseForm() // parse the query that /should/ be attached to the URL

		// grab template
		t, err := template.ParseFiles("site/templates/minskew.html")
		if err != nil {
			panic(err)
		}

		var page Page
		// check if a URL was supplied
		if url, ok := r.Form["url"]; ok {
			if err := DownloadFile("genome.fa.gz", url[0]); err != nil {
				panic(err)
			}
			page = Page{
				Title:    "Resulting Skew Plot",
				Contents: "<img src='/plots/skew.jpg' class='rounded' alt='skew' style='width:90%;height:50%;'>",
			}
		} else {
			page = Page{
				Title:    "Error",
				Contents: "<h3 style='color: black'> You must supply a '?url=' parameter in the URL!</h3>",
			}
		}
		// push our page into the template
		t.Execute(w, page)
	}
}

// DownloadFile Download the fa.gz file from the URL and save it on disk
func DownloadFile(filepath string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	ReadFile(filepath)

	return err
}

// isGzipped checks the first two bytes to see if it's a gzipped file
func isGzipped(file io.Reader) bool {
	reader := bufio.NewReader(file)
	testBytes, err := reader.Peek(2)
	if err != nil {
		panic(err)
	}
	return testBytes[0] == 31 && testBytes[1] == 139
}

// ReadFile Read the fa.gz file using gzip package
// Note: Since genomes are large, limit to first 1000 lines
func ReadFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	var reader io.Reader
	if isGzipped(file) {
		reader, err = gzip.NewReader(file)
		if err != nil {
			panic(err)
		}
	} else {
		reader = bufio.NewReader(file)
	}
	scanner := bufio.NewScanner(reader)

	genome := ""
	count := 0
	for scanner.Scan() && count < 1000 {
		currentLine := scanner.Text()
		if currentLine[0:1] != ">" {
			genome += strings.Trim(currentLine, "\n ")
			count += 1
		}
	}
	err = CalculateMinSkew(genome)
	return err
}

// CalculateMinSkew Calculate MinSkew and save the values to a slice
func CalculateMinSkew(genome string) error {
	c := 0
	g := 0
	skewList := make([]int, 0)
	for i := range genome {
		if genome[i] == 'c' || genome[i] == 'C' {
			c += 1
		}
		if genome[i] == 'g' || genome[i] == 'G' {
			g += 1
		}
		skewList = append(skewList, g-c)
	}

	err := PlotGraph(skewList)
	return err
}

// PlotGraph Plot the graph for min skew and send it as response to the html page
func PlotGraph(skewList []int) error {
	p, err := plot.New()
	if err != nil {
		return err
	}

	p.Title.Text = "G-C Skew Along The Genome"
	p.X.Label.Text = "Position in Genome (bp)"
	p.Y.Label.Text = "Skew (G-count minus C-count)"

	// create all of our coordinates
	pts := make(plotter.XYs, len(skewList))
	for i := range pts {
		pts[i].X = float64(i)
		pts[i].Y = float64(skewList[i])
	}

	// put them into a line plot
	l, err := plotter.NewLine(pts)
	if err != nil {
		return err
	}
	l.Color = color.RGBA{R: 45, G: 77, B: 240, A: 1}
	p.Add(l) // add linePlot to plot

	if err := p.Save(12.5*vg.Inch, 5*vg.Inch, "plots/skew.jpg"); err != nil {
		return err
	}

	return err
}

// main function holds the information about function handling based on path of the request
// Similar to routing in Ruby on Rails
func main() {
	// URL handlers
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/get_url/", getUrlHandler)
	http.HandleFunc("/minskew/", minskewHandler)

	// file server: provides CSS/JS/img files + plots generated by the server
	http.Handle("/plots/", http.StripPrefix("/plots/", http.FileServer(http.Dir("./plots"))))
	http.Handle("/site/", http.StripPrefix("/site/", http.FileServer(http.Dir("./site"))))

	// output
	port, ok := os.LookupEnv("PORT")
	if !ok {
		port = "8080"
	}
	port = ":" + port

	fmt.Println("Now running on http://localhost" + port)
	fmt.Println()
	fmt.Println("HTTP Actions:")

	// start the web server
	http.ListenAndServe(port, nil)
}
