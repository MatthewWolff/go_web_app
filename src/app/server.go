package main

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"hash/fnv"
	"html/template"
	"image/color"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
)

const PLOTS = "plots"
const TEMPLATES = "site/templates"

// Page a small struct to help us use HTML Templates
type Page struct {
	Title    string
	Contents template.HTML
}

// PostResponse a small struct to output a response with
type PostResponse struct {
	Msg  string
	Keys string
}

// PostRequest the expected structure of the POST request. Non-conforming requests will be (mostly) ignored
type PostRequest struct {
	Key1 string
	Key2 string
}

// indexHandler Basic index handler for root
func indexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method:", r.Method, "on", "index") // print request method
	if r.Method == "GET" {
		t, err := template.ParseFiles(path.Join(TEMPLATES, "index.html"))
		if err != nil {
			panic(err)
		}
		t.Execute(w, nil) // we could pass a struct in to apply formatting if we wanted
	} else if r.Method == "POST" {
		servePostRequest(w, r)
	}
}

func servePostRequest(w http.ResponseWriter, r *http.Request) {
	// parse the post request
	var req PostRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// create response
	response := PostResponse{
		Msg:  "Hello! Thanks for the POST request.",
		Keys: fmt.Sprintf("%+v", req),
	}

	// issue response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// getUrlHandler Handles get_url request
func getUrlHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method:", r.Method, "on", "get_url") //get request method
	if r.Method == "GET" {
		t, err := template.ParseFiles(path.Join(TEMPLATES, "get_url.html"))
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
		r.ParseForm() // parse the query that *should* be attached to the URL

		// grab template
		t, err := template.ParseFiles(path.Join(TEMPLATES, "minskew.html"))
		if err != nil {
			panic(err)
		}

		var page Page
		// check if a URL was supplied
		if url, ok := r.Form["url"]; ok {
			_, ok := r.Form["overwrite"] // force plot regeneration if overwrite param is present
			plotFile := processRequest(url[0], ok)
			plotPath := path.Join(PLOTS, plotFile)
			page = Page{
				Title: "Resulting Skew Plot",
				Contents: template.HTML(fmt.Sprintf(
					"<img src='/%s' class='rounded' alt='skew' style='width:90%%;height:50%%;'>", plotPath,
				)),
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

// hash is a helper function for creating a unique string from a url. stolen from stack overflow
func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

// processRequest takes a fasta URL then generates a skew plot for it. Caches the plots for efficiency
func processRequest(url string, overwrite bool) string {
	hashed := hash(url) // don't regenerate plots we already have
	plotFile := fmt.Sprintf("skew_%d.jpg", hashed)
	outfile := fmt.Sprintf("genome_%d.fa.gz", hashed)

	if !plotExists(plotFile) {
		if err := downloadFile(outfile, url); err != nil {
			panic(err)
		}
		genome := readFile(outfile)
		skewList := calculateMinSkew(genome)
		p := plotGraph(skewList)
		plotPath := path.Join(PLOTS, plotFile)
		if err := p.Save(12.5*vg.Inch, 5*vg.Inch, plotPath); err != nil {
			panic(err)
		}
		// remove file when done to save disk space
		if err := os.Remove(outfile); err != nil {
			panic(err)
		}
	}
	return plotFile
}

// plotExists checks if a plot exists and returns a boolean accordingly
func plotExists(file string) bool {
	if _, err := os.Stat(path.Join(PLOTS, file)); err == nil {
		return true
	} else if os.IsNotExist(err) {
		return false
	} else {
		// Schrodinger: file may or may not exist. See err for details.
		panic(err)
	}
}

// downloadFile Download the fa.gz or fasta file from the URL and save it on disk
func downloadFile(filepath string, url string) error {
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

// readFile Read the fa.gz file using gzip package
// Note: Since genomes are large, limit to first 1000 lines
func readFile(filename string) string {
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
	return genome
}

// calculateMinSkew Calculate MinSkew and save the values to a slice
func calculateMinSkew(genome string) []int {
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

	return skewList
}

// plotGraph Plot the graph for min skew and send it as response to the html page
func plotGraph(skewList []int) *plot.Plot {
	p, err := plot.New()
	if err != nil {
		panic(err)
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
		panic(err)
	}
	l.Color = color.RGBA{R: 45, G: 77, B: 240, A: 1}
	p.Add(l) // add linePlot to plot

	return p
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
