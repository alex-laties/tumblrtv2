package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image/gif"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"

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
	rand.Seed(time.Now().Unix())
	maddenGIFBytes := gifBox.Bytes("madden.gif")
	maddenReader := bytes.NewBuffer(maddenGIFBytes)
	maddenGIF, err := gif.DecodeAll(maddenReader)
	if err != nil {
		panic(err)
	}

	gifPipeline <- maddenGIF

	for {
		tag := tags[rand.Intn(len(tags))]
		resp, err := http.Get(fmt.Sprintf("https://api.tumblr.com/v2/gif/search/%s?api_key=%s", tag, tumblrAPIKey))
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

		// shuffle GIFs
		rand.Shuffle(len(t.Response.GIFs), func(i, j int) {
			t.Response.GIFs[i], t.Response.GIFs[j] = t.Response.GIFs[j], t.Response.GIFs[i]
		})

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
