package mainmenu

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/itsmenewbie03/go-animegg/scrapers/animegg"
)

type MainMenuController struct {
	Client *animegg.Client
}

func NewMainMenuController(client *animegg.Client) *MainMenuController {
	return &MainMenuController{
		Client: client,
	}
}

func (c MainMenuController) Search(query string) (*[]list.Item, error) {
	res, err := c.Client.Search(query)
	if err != nil {
		return nil, err
	}
	var items []list.Item
	for _, anime := range res {
		desc := fmt.Sprintf("Status: %-9s • Episodes: %s", anime.Status, anime.TotalEpisodes)
		i := item{
			title: anime.Title,
			desc:  desc,
			link:  anime.URL,
		}
		items = append(items, i)
	}
	return &items, nil
}

func (c MainMenuController) GetEpisodes(url string) (*[]list.Item, error) {
	res, err := c.Client.Episodes(url)
	if err != nil {
		return nil, err
	}
	var items []list.Item
	for _, episode := range res {
		xs := strings.Split(episode.Title, " - ")
		title := xs[0]
		desc := xs[1]
		i := item{
			title: title,
			desc:  desc,
			link:  episode.URL,
		}
		items = append(items, i)
	}
	return &items, nil
}

func (c MainMenuController) GetVideos(url string) (*[]list.Item, error) {
	res, err := c.Client.Videos(url)
	if err != nil {
		return nil, err
	}
	var items []list.Item
	for _, video := range res {
		title := video.Title
		desc := fmt.Sprintf("%-5s • %s", video.Quality+"p", strings.ToUpper(video.Language))
		i := item{
			title: title,
			desc:  desc,
			link:  video.URL,
		}
		items = append(items, i)
	}
	return &items, nil
}
