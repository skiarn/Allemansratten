package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"

	"github.com/elazarl/goproxy"
)

var logger *log.Logger
var errorlog *os.File

func main() {
	var iport = flag.Int("infop", 7866, "What port the information site should run on.")
	var pport = flag.Int("p", 7867, "What port the proxy should run on.")
	flag.Parse() // parse the flags
	errorlog, err := os.OpenFile("allemansratten.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("error opening file: %v", err)
		os.Exit(1)
	}
	defer errorlog.Close()
	logger = log.New(errorlog, "applog: ", log.Lshortfile|log.LstdFlags)

	logger.Printf("Information website server listen on port: %v \n", *iport)
	mux := http.NewServeMux()
	mux.HandleFunc("/firefox", func(w http.ResponseWriter, r *http.Request) {
		client := &http.Client{}
		request, err := http.NewRequest("GET", "https://download.mozilla.org/?product=firefox-stub&os=win&lang=en-US", nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		request.Header.Add("User-Agent", r.Header.Get("User-Agent"))
		resp, err := client.Do(request)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()
		contents, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(contents)

	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w,
			`<!DOCTYPE html>
<html>
<body>
<p>Hi there, I recomend downloading firefox to help use allemansratten </p>
      <a href="https://www.mozilla.org/en-US/firefox/new/?scene=2">Get firefox.. Direct link</a>
      <a href="/firefox">Download through me!</a>
      <p>
        1. Open Firefox
        2. Click upper right corner ≈ symbol.
        3. Choose Preferences
        4. Click Advanced, bottom left side list.
        5. Choose network —> Settings…
        6. Click Manuall proxy configuration: Host Proxy: %s Port: %s
      </p>
  </body>
</html>
`, r.Host, strconv.Itoa(*pport))
	})
	go func() {
		logger.Fatal(http.ListenAndServe(":"+strconv.Itoa(*iport), mux))
	}()

	logger.Printf("Proxy server to listen on a port: %v \n", *pport)
	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = true

	proxy.OnRequest(goproxy.ReqHostMatches(regexp.MustCompile(".*"))).DoFunc(
		func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
			ryoutube, _ := regexp.Compile("youtube.*")
			rfacebook, _ := regexp.Compile("facebook.*")
			rgoogle, _ := regexp.Compile("google.*")
			rwikipedia, _ := regexp.Compile("wikipedia.*")

			isyoutube := ryoutube.MatchString(r.Host)
			isfacebook := rfacebook.MatchString(r.Host)
			isgoogle := rgoogle.MatchString(r.Host)
			iswiki := rwikipedia.MatchString(r.Host)

			if !isyoutube && !isfacebook && !isgoogle && !iswiki {
				return r, goproxy.NewResponse(r,
					goproxy.ContentTypeText, http.StatusForbidden,
					"Sorry, I only allow, youtube, facebook, google, wikipedia!")
			}
			return r, nil
		})

	logger.Fatal(http.ListenAndServe(":"+strconv.Itoa(*pport), proxy))
}
