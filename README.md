# Running a Web Server using Go
Once you've cloned this repository, simply run the setup script:
```bash
bash setup.sh  # like this
./setup.sh  # or like this
go_web_app/setup.sh # or like this
```

---

# Hosting the server
[Heroku](https://devcenter.heroku.com/articles/getting-started-with-go) makes it very easy to host your app!  

1. Create `server.go`
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
    echo 'web: bin/server' > Procfile # tells Heroku how to run the file
    go mod init go_web_app/skew # create mod file
    go test  # this builds the go.mod file
    go mod vendor # this then generates the files you need to build
   
    # build a linux-specific binary
    env GOOS=linux GOARCH=amd64 go build -o bin/server -v .
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
