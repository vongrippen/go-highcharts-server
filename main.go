package main

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

func chart(w http.ResponseWriter, r *http.Request) {
	// Temp files
	var callbackfile *os.File
	chartfile, _ := ioutil.TempFile("./temp", "hc_chart_")
	infile, _ := ioutil.TempFile("./temp", "hc_input_")

	// Read the params
	input := r.FormValue("input")
	scale := r.FormValue("scale")
	cType := r.FormValue("type")
	if len(cType) == 0 {
		cType = "png"
	}
	width := r.FormValue("width")
	constr := r.FormValue("constr")
	callback := r.FormValue("callback")

	// Setup the options to pass to phantomjs
	var options []string
	infile.WriteString(input)
	infile.Close()
	options = append(options, "highcharts-convert.js")
	options = append(options, "-infile")
	options = append(options, infile.Name())
	if len(width) > 0 {
		options = append(options, "-scale")
		options = append(options, scale)
	}
	options = append(options, "-outfile")
	options = append(options, fmt.Sprintf("%s.%s", chartfile.Name(), cType))
	if len(width) > 0 {
		options = append(options, "-width")
		options = append(options, width)
	}
	if len(constr) > 0 {
		options = append(options, "-constr")
		options = append(options, constr)
	}
	if len(callback) > 0 {
		callbackfile, _ = ioutil.TempFile("./temp", "hc_callback_")
		callbackfile.WriteString(callback)
		callbackfile.Close()
		options = append(options, "-callback")
		options = append(options, callbackfile.Name())
	}

	// Run highcharts
	cmd := exec.Command("phantomjs", options...)
	out, err := cmd.Output()
	fmt.Printf("%s\n", out)
	if err != nil {
		log.Fatal(err)
	}

	//Serve the file
	http.ServeFile(w, r, fmt.Sprintf("%s.%s", chartfile.Name(), cType))

	// Cleanup files
	if len(callback) > 0 {
		os.Remove(callbackfile.Name())
	}
	os.Remove(infile.Name())
	// Don't delete the actual chart though, might not be finished downloading and
	// the daily Heroku reboot at midnight should clear this out.
	//os.Remove(chartfile.Name())
}

// Respond to ping requests (for keepalive on Heroku)
func pong(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "pong")
}

// Ping yourself (for keepalive on Heroku)
func ping(url string) {
	for {
		time.Sleep(10000 * time.Millisecond)
		http.Get(url)
	}
}

type handler func(w http.ResponseWriter, r *http.Request)

// HTTP Basic Auth wrapper
func BasicAuth(pass handler, u string, p string) handler {

	return func(w http.ResponseWriter, r *http.Request) {

		auth := strings.SplitN(r.Header["Authorization"][0], " ", 2)

		if len(auth) != 2 || auth[0] != "Basic" {
			http.Error(w, "bad syntax", http.StatusBadRequest)
			return
		}

		payload, _ := base64.StdEncoding.DecodeString(auth[1])
		pair := strings.SplitN(string(payload), ":", 2)

		if len(pair) != 2 || !Validate(pair[0], pair[1], u, p) {
			http.Error(w, "authorization failed", http.StatusUnauthorized)
			return
		}

		pass(w, r)
	}
}

func Validate(username, password, u, p string) bool {
	if username == u && password == p {
		return true
	}
	return false
}

func main() {
	port := "8080"
	ip := "0.0.0.0"

	if len(os.Getenv("PORT")) > 0 {
		port = os.Getenv("PORT")
	}
	if len(os.Getenv("IP")) > 0 {
		ip = os.Getenv("IP")
	}

	if len(os.Getenv("KEEPALIVE_URL")) > 0 {
		go ping(os.Getenv("KEEPALIVE_URL"))
	}

	u, p := os.Getenv("HTTP_BASIC_USERNAME"), os.Getenv("HTTP_BASIC_PASSWORD")

	// Only enable basic auth if the env vars are set
	if len(u) > 0 && len(p) > 0 {
		http.HandleFunc("/", BasicAuth(chart, u, p))
	} else {
		http.HandleFunc("/", chart)
	}
	http.HandleFunc("/ping", pong)
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%s", ip, port), nil))
}
