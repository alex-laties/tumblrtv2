package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image/gif"
	"io/ioutil"
	"net/http"

	"github.com/gobuffalo/packr"
)

const tumblrAPIKey = "oK7KDFFbmTXCDKyPoehhKMHlbWMGOZVOWejcSuNLSJGYunjdkN"

var gifBox = packr.NewBox("./assets/gifs")

type TumblrAPIMeta struct {
	Status int    `json:"status"`
	Msg    string `json:"msg"`
}

type TumblrAPIResponse struct {
	GIFs []TumblrGIFResponse
}

type TumblrGIFResponse struct {
	MediaURL string `json:"media_url"`
}

type TumblrGIFSearchAPIResponse struct {
	Meta     TumblrAPIMeta     `json:"meta"`
	Response TumblrAPIResponse `json:"response"`
}

var (
	gifPipeline = make(chan *gif.GIF, 10)
)

func fetchGIFs(tags ...string) {
	maddenGIFBytes := gifBox.Bytes("madden.gif")
	maddenReader := bytes.NewBuffer(maddenGIFBytes)
	maddenGIF, err := gif.DecodeAll(maddenReader)
	if err != nil {
		panic(err)
	}

	gifPipeline <- maddenGIF
	for {
		resp, err := http.Get(fmt.Sprintf("https://api.tumblr.com/v2/gif/search/cat?api_key=%s", tumblrAPIKey))
		if err != nil {
			panic(err)
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}

		var t TumblrGIFSearchAPIResponse
		err = json.Unmarshal(body, &t)
		if err != nil {
			panic(err)
		}

		for _, gifr := range t.Response.GIFs {
			resp, err = http.Get(gifr.MediaURL)
			if err != nil {
				panic(err)
			}

			g, err := gif.DecodeAll(resp.Body)
			if err != nil {
				panic(err)
			}
			gifPipeline <- g
		}
	}
}
