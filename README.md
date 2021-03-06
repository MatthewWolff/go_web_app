# Running a Web App using Go
Once you've cloned this repository, simply run the setup script:
```bash
bash setup.sh  # like this
./setup.sh  # or like this
go_web_app/setup.sh # or like this
```

It should then tell you that the server is live on your local machine (`localhost`)!

Test it out by going to the address it outputs. When asked, try these URLs:
```
https://www.ncbi.nlm.nih.gov/sviewer/viewer.fcgi?id=1061361601&db=nuccore&report=fasta&retmode=text&withmarkup=on&tool=portal&log$=seqview&maxdownloadsize=100000000

https://raw.githubusercontent.com/MatthewWolff/go_web_app/master/src/app/test_data/rickettsia_prowazekii.txt

https://raw.githubusercontent.com/MatthewWolff/go_web_app/master/src/app/test_data/thermotoga_petrophila.txt
```

---

# What is a server?
We often hear about servers in movies and everyday life:
"he hacked the server" or "500 internal server error"—but
what exactly is it? You might think that it's just some giant
computer or multi-room machine, but in reality a server can 
just be a small monitor-less computer sitting on a shelf. Your
computer issues a request to it for information, and it
"serves" the response. 

More formally, a server is a piece of hardware or software 
that provides a service to "clients". In our context, the
server is serving web **requests** to it.

So let's chat about requests.

# Requests
## Issuing Requests: cURL
For the purpose of demonstration, we'll use a command line tool that is 
on every one of your machines: cURL AKA `curl`. From their `man` pages:

> curl  is  a  tool to transfer data from or to a server

It's used like this: 
```bash
$ curl --request [TYPE (default: GET)] [URL]
```

## Request Types
* **GET**: 
    - The most common type of request!
    - You use it every day
    - Every time you put a URL in your browser, it makes a GET request
    - What does a GET request bring back? Usually HTML!
    - You can add parameters to the URL like this: `curl url.com?key1=val&key2=val`
    ```html
    $ curl --request GET skew-web-server.herokuapp.com --silent | tail -n 20
    <header class="intro">
        <div class="intro-body">
            <div class="container">
                <div class="row">
                    <div class="col-lg-12">
                        <h2 class="brand-heading"> Welcome. </h2>
                        <p class="intro-text"> Let's learn about Web Apps </p>
                        <form action="/get_url" method="get">
                            <button class="btn btn-circle" type="submit">
                                <i class="fa animated"> ok </i>
                            </button>
                        </form>
                    </div>
                </div>
            </div>
        </div>
    </header>
    </body>
    
    </html>
    ```
* **POST**
    - Another VERY common type of request
    - Like a GET request, you can transfer information
    - The difference: POST is more secure, because it doesn't encode the information in the URL
    - Usually the response is JSON (JavaScript Object Notation)
        - the `--silent` flag gets rid of extra output
        - we'll use the tool `jq` to format the JSON to be prettier 
    ```bash
    $ curl --request POST skew-web-server.herokuapp.com \
        --data '{"Key1": "key1", "key2": "val2"}' \
        --silent | jq
    {
      "Msg": "Hello! Thanks for the POST request.",
      "Keys": "{Key1:key1 Key2:val2}"
    }
    ```
    - if you have some trouble running, that try it all on one line:
    ```bash
    curl --request POST skew-web-server.herokuapp.com --data '{"Key1": "key1", "key2": "val2"}' --silent
    ``` 
* Others
    - Delete, Put, Patch, etc.
    
# Handling Requests in Go
When you go to a website, you usually visit multiple pages. How does the website
know which page to return to your GET request? Well, obviously there's some 
important information in the URL!

Go has the `net/http` module, which provides a very nice way to run a web server. 
You simply specify a "request handler" for the supported pages. In the following 
example, we have 3 supported pages (with index being the default page that you see).
```go
// URL handlers
http.HandleFunc("/", indexHandler)
http.HandleFunc("/get_url/", getUrlHandler)
http.HandleFunc("/minskew/", minskewHandler)

http.ListenAndServe(":8080", nil)
```

## HTML Templating
Let's take a look at a request handler for the `mysite.com/get_url/` page
```go
// getUrlHandler Handles get_url request
func getUrlHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		t, err := template.ParseFiles(path.Join(TEMPLATES, "get_url.html"))
		if err != nil { panic(err) }
		t.Execute(w, nil) // we could pass a struct in to apply formatting if we wanted
	}
}
```
When a request is made, we can check the type and respond accordingly.  

In our example, we'll send back a `.html` page that we've designed. Go
also supports something called templating:
```go
type Page struct {
	Title    string
	Contents template.HTML
}

func myHandler(w http.ResponseWriter, r *http.Request) {
    // Create a simple HTML template!
    t, err := template.New("foo").Parse("<h1>{{.Title}}</h1> {{.Contents}}")
    if err != nil { panic(err) }
    
    // Fill out our template
    page := Page{Title: "My Title", Contents: "These are my contents"}
    t.Execute(w, page) // put the contents of page into the template
}
```

### HTML
While learning HTML is outside the scope of this recitation, it can be a
useful talent, albeit one which you can get away without having (i.e., 
if you know markdown you can use `pandoc` to convert it into an HTML page!).
Here's a [tutorial](https://www.w3schools.com/html/) on HTML. It's pretty basic:
```HTML
<html lang="en">
<head>
    <!-- you import things here, like Cascading Style Script (CSS) or Javascript --> 
</head>
<body>
    <h1> a header </h1>
    <h2> a smaller header </h2>
    <h3> even smaller header </h3>
    <p> a paragraph </p>
    
    My Unordered List:
    <ul>
        <li> list item 1 </li>
        <li> list item 2 </li>
    </ul>
    My Ordered List:
    <ol>
        <li> list item 1 </li>
        <li> list item 2 </li>
    </ol>
       <form action="/minskew" method="get">
            URL: <input type="text" name="url">
            <button type="submit" value="submit"> Submit </button>
       </form>
</body>
</html>
``` 

<img alt="example" src="./src/app/site/img/example.png" width="600"/>


    
# FileServer
What do you do if you want to use the file system on your server? By default,
anyone who accesses the server can't just see all its files. This makes sense.
But if you want to use CSS/JS files on your webpages, you'll actually need to 
provide access to those somehow! Same case for if you want to show stored images.  

Let's imagine the case where we have a bunch of images in a folder called `plots/`
```go 
http.Handle("/plots/", http.StripPrefix("/plots/", http.FileServer(http.Dir("./plots"))))
```

Now, any HTTP request to `oursite.com/plots` will be able to access the files stored there!

---


# Hosting the Web Server
[Heroku](https://devcenter.heroku.com/articles/getting-started-with-go)
makes it very easy to host your app!  

1. Create `server.go`
    - for this part, we don't want it to be in a project under `src/`!
    - this is just because Heroku uses some outdated tech ([`GB`](https://github.com/constabulary/gb)) when it senses a `src/` folder
        - it took me like 2 hours to debug, and I wanted to die
    - we want the file hierarchy here to be flat, so make a new folder:
    ```bash
    mkdir ~/herokuServer
    cp server.go ~/herokuServer # let's assume `server.go` exists
    cp -r site/ ~/herokuServer # assume your site with the templates exists
    cd ~/herokuServer
    ```
2. Make the code check the environment for a `$PORT`
    ```go
   	port, ok := os.LookupEnv("PORT")
   	if !ok {
   		port = "8080" // set a default
   	}
   	port = ":" + port
   
   	// start the web server
   	http.ListenAndServe(port, nil)
    ```
4. Prepare it for Heroku
    ```bash
    go mod init go_web_app/skew # create mod file
    go test  # this builds the go.mod file
    go mod vendor # this then generates the files you need to build
   
    # build a linux-specific binary
    env GOOS=linux GOARCH=amd64 go build -o bin/server -v .
    echo 'web: bin/server' > Procfile # tells Heroku how to run the file
    ```
5. Deploy it on Heroku!
    ```bash
    # create a commit
    git init
    git add --all
    git commit -m "heroku files"
    
    # create the heroku app
    heroku login 
    heroku create skew-web-server # give it a name
    git push heroku master
    echo 'we did it!'
    curl skew-web-server.herokuapp.com # check it out
    ```
