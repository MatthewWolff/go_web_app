package main

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"html/template"
	"io"
	"net/http"
	"os"
	"strings"
)

// index_handler Basic index handler for root
func index_handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<h1>Whoa </h1> Go is neat!")
}

// getUrl_handler Handles get_url request
// the UI is in minskew.template
func getUrl_handler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method:", r.Method) //get request method
	if r.Method == "GET" {
		t, err := template.ParseFiles("minskew.template")
		if err != nil {
			panic(err)
		}
		t.Execute(w, nil)
	}
}

// minskew_handler Handles the /minskew request
func minskew_handler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method:", r.Method) //get request method
	if r.Method == "GET" {
		// parse the query that /should/ be attached to the URL
		r.ParseForm()

		// check if a URL was supplied
		if url, ok := r.Form["url"]; ok {
			if err := DownloadFile("genome.fa.gz", url[0], w); err != nil {
				panic(err)
			}
		} else {
			fmt.Fprintln(w, "<h1> Error </h1>"+
				"<h2> you must supply a '?url=' field in the URL </h2>")
		}
	}
}

// DownloadFile Download the fa.gz file from the URL and save it on disk
func DownloadFile(filepath string, url string, w http.ResponseWriter) error {
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

	ReadFile(filepath, w)

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
// Since genomes are large, limit to first 1000 lines
func ReadFile(filename string, w http.ResponseWriter) error {
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
	err = CalculateMinSkew(genome, w)

	return err
}

// CalculateMinSkew Calculate MinSkew and save the values to a slice
func CalculateMinSkew(genome string, w http.ResponseWriter) error {
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

	err := PlotGraph(skewList, w)
	return err
}

// PlotGraph Plot the graph for min skew and send it as response to the html page
func PlotGraph(skew_list []int, w http.ResponseWriter) error {
	p, err := plot.New()
	if err != nil {
		return err
	}

	pts := make(plotter.XYs, len(skew_list))
	for i := range pts {
		pts[i].X = float64(i)
		pts[i].Y = float64(skew_list[i])
	}

	l, err := plotter.NewLine(pts)
	if err != nil {
		return err
	}

	p.Add(l)
	if err := p.Save(50*vg.Inch, 10*vg.Inch, "assets/skew.jpeg"); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(w, "<h1>Skew graph:</h1>")
	fmt.Fprintf(w, "<img src='../assets/skew.jpeg' alt='skew' style='width:1400px;height:500px;'>")

	return err
}

// main function holds the information about function handling based on path of the request
// Similar to routing in Ruby on Rails
func main() {
	addr := "0.0.0.0:8000"
	http.HandleFunc("/", index_handler)
	http.HandleFunc("/get_url/", getUrl_handler)
	http.HandleFunc("/minskew/", minskew_handler)
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("./assets"))))
	fmt.Println("Now running on http://" + addr)
	http.ListenAndServe(addr, nil)
}
