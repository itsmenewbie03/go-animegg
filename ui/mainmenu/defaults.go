package mainmenu

import "github.com/charmbracelet/bubbles/list"

var mainMenuItems = []list.Item{
	item{ident: CURRENT, title: "Currently Watching", desc: "Animes you're currently watching"},
	item{ident: SHOWALL, title: "Show All", desc: "Display all animes in your list"},
	item{ident: UNTRACKED, title: "Untracked watching", desc: "Animes you're watching but not tracking"},
	item{ident: UPDATE, title: "Update", desc: "Update anime progress or metadata"},
	item{ident: CONTINUE, title: "Continue Last Session", desc: "Resume watching from where you left off"},
	item{ident: QUIT, title: "Quit", desc: "Exit the application"},
}
