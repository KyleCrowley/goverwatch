package main

import (
	"strings"
	"github.com/PuerkitoBio/goquery"
	"log"
)

type Player struct {
	Platform string
	Region   string
	Tag      string
}

// getPlayer returns a player object from a map of vars.
// Used in routes that contain "/{platform}/{region}/{tag}"
func getPlayer(vars map[string]string) *Player {
	return &Player{strings.ToLower(vars["platform"]), strings.ToLower(vars["region"]), vars["tag"]}
}

func (p *Player) platformIsValid() bool {
	if !platforms[p.Platform] {
		return false
	}

	return true
}

func (p *Player) regionIsValid() bool {
	if !regions[p.Region] {
		return false
	}

	return true
}

// GetProfileDoc gets the matching player's HTML document.
// TODO: Better error handling
func (p *Player) GetProfileDoc() *goquery.Document {
	d, err := goquery.NewDocument(p.formatProfileURL())
	if err != nil {
		log.Fatal(err)
	}

	return d
}

// formatProfileURL constructs and returns the profile URL of the player.
// PC players require a region in the URL, while PSN/XBL players do not.
// Also calls helper methods to sanitize BattleTags.
func (p *Player) formatProfileURL() string {
	// PSN/XBL: https://playoverwatch.com/en-us/career/${platform}/${tag}
	// PC: https://playoverwatch.com/en-us/career/${platform}/${region}/${tag}

	if p.Platform == "pc" {
		return BASE_URL + p.Platform + "/" + p.Region + "/" + p.sanitizeBattleTag()
	} else {
		return BASE_URL + p.Platform + "/" + p.sanitizeBattleTag()
	}

}

// formatSearchURL constructs and returns the search URL for the given tag.
func (p *Player) formatSearchURL() string {
	return SEARCH_URL + p.sanitizeBattleTag()
}

// sanitizeBattleTag returns a sanitized BattleTag.
// Ex: "#" cannot be used in URL's, unless they are encoded.
// The official Overwatch site simply replaces "#" with "-", so this function does just that.
func (p *Player) sanitizeBattleTag() string {
	return strings.Replace(p.Tag, "#", "-", 1)
}
