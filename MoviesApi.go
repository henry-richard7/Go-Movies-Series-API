package main

import (
	"encoding/json"
	"fmt"
	"github.com/gocolly/colly/v2"
	"net/http"
	"regexp"
	"strings"
)

type item struct {
	Title  string
	Link   string
	Poster string
}

// Returns JSON Array for Movies
func getContents(cType string, pageNo string) string {
	c := colly.NewCollector()
	var tempString string

	c.OnHTML("#main_bg > div:nth-child(5) > div > div.vc_row.wpb_row.vc_row-fluid.vc_custom_1404913114846 > div.vc_col-sm-12.wpb_column.column_container > div > div > ul > li > a", func(e *colly.HTMLElement) {
		temp := item{Title: e.ChildText("div.name"), Link: "https://vidcloud9.com" + e.Attr("href"), Poster: e.ChildAttr("div.img>div.picture>img", "src")}
		b, _ := json.Marshal(temp)

		tempString += string(b) + ","

	})

	/* 	I know this is not Optimal solution for returning as JSON
	   	I am new to Go Language
	   	If you know how to return struct as JSON array please teach me !!
	*/

	c.Visit("https://vidcloud9.com/" + cType + "?page=" + pageNo)

	i := strings.LastIndex(tempString, ",")
	excludingLast := tempString[:i] + strings.Replace(tempString[i:], ",", "", 1)
	results := "[" + excludingLast + "]"
	return results

}

func fixStreamLink(streamLink string) string {
	var jsScript string
	var re = regexp.MustCompile(`(?m)file: '(.*?)'`)
	c := colly.NewCollector()
	c.OnHTML("body>div>div>script", func(element *colly.HTMLElement) {
		jsScript = element.Text

	})
	c.Visit(streamLink)
	var match = re.FindString(jsScript)
	fixString := strings.ReplaceAll(match, "file: '", "")
	fixString = strings.ReplaceAll(fixString, "'", "")
	return fixString
}

func getDirectLink(videoLink string) string {
	type directLink struct {
		Url string
	}
	var fixedUrl string

	c := colly.NewCollector()
	c.OnHTML("#main_bg > div:nth-child(5) > div > div.video-info-left > div.watch_play > div.play-video > iframe", func(element *colly.HTMLElement) {
		fixedUrl = fixStreamLink("https:" + strings.ReplaceAll(element.Attr("src"), "streaming.php", "loadserver.php"))

	})
	c.Visit(videoLink)

	temp := directLink{Url: fixedUrl}
	b, _ := json.Marshal(temp)

	return string(b)
}
func getEpisodes(videoLink string) string {
	type episodesJsonArray struct {
		EpisodeNumber string
		EpisodeUrl    string
	}
	var tempString string
	c := colly.NewCollector()
	c.OnHTML("#main_bg > div:nth-child(5) > div > div.video-info-left > ul > li > a", func(element *colly.HTMLElement) {
		temp := episodesJsonArray{EpisodeNumber: element.ChildText("div.name"), EpisodeUrl: "https://vidcloud9.com" + element.Attr("href")}
		b, _ := json.Marshal(temp)

		tempString += string(b) + ","
	})
	c.Visit(videoLink)
	i := strings.LastIndex(tempString, ",")
	excludingLast := tempString[:i] + strings.Replace(tempString[i:], ",", "", 1)
	results := "[" + excludingLast + "]"
	return results
}
func moviesHandler(w http.ResponseWriter, r *http.Request) {
	pageNo := r.URL.Query().Get("pageNo")
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	fmt.Fprintf(w, getContents("movies", pageNo))

}

func seriesHandler(w http.ResponseWriter, r *http.Request) {
	pageNo := r.URL.Query().Get("pageNo")
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	fmt.Fprintf(w, getContents("series", pageNo))

}

func directLinkHandler(w http.ResponseWriter, r *http.Request) {
	videoLink := r.URL.Query().Get("videoLink")
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	fmt.Fprintf(w, getDirectLink(videoLink))

}

func episodesHandler(w http.ResponseWriter, r *http.Request) {
	videoLink := r.URL.Query().Get("videoLink")
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	fmt.Fprintf(w, getEpisodes(videoLink))
}

func main() {
	
	http.HandleFunc("/movies", moviesHandler)
	http.HandleFunc("/series", seriesHandler)
	http.HandleFunc("/directLink", directLinkHandler)
	http.HandleFunc("/episodes", episodesHandler)
	http.ListenAndServe(":8080", nil)
}
