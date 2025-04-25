package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"

	"github.com/itsmenewbie03/go-animegg/scrapers/animegg"
	"github.com/itsmenewbie03/go-animegg/ui/mainmenu"
)

var (
	searchQuery  = flag.String("search", "", "Search query")
	popularFlag  = flag.Bool("popular", false, "Show popular anime")
	latestFlag   = flag.Bool("latest", false, "Show latest updates")
	episodesFlag = flag.String("episodes", "", "Get episodes for anime URL")
	videosFlag   = flag.String("videos", "", "Get videos for episode URL")

	lang    = flag.String("lang", "subbed", "Preferred language [subbed|dubbed|raw]")
	quality = flag.String("quality", "1080", "Preferred quality")
	server  = flag.String("server", "AnimeGG", "Preferred server")
)

func main() {
	flag.Parse()

	client := animegg.NewClient(*lang, *quality, *server)

	switch {
	case *searchQuery != "":
		results, err := client.Search(*searchQuery)
		handleResult(results, err)
	case *popularFlag:
		results, err := client.Popular()
		handleResult(results, err)
	case *latestFlag:
		results, err := client.Latest()
		handleResult(results, err)
	case *episodesFlag != "":
		results, err := client.Episodes(*episodesFlag)
		handleResult(results, err)
	case *videosFlag != "":
		results, err := client.Videos(*videosFlag)
		handleResult(results, err)
	default:
		mainmenu.Render(client)
	}
}

func handleResult(data any, err error) {
	if err != nil {
		log.Fatal(err)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(data); err != nil {
		log.Fatal(err)
	}
}
