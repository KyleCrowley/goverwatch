package main

import (
	"github.com/PuerkitoBio/goquery"
	"encoding/json"
	"net/http"
	"github.com/gorilla/mux"
	"strings"
	"io/ioutil"
)

type Achievement struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	ImageURL    string `json:"image_url"`
	Finished    bool `json:"finished"`
}

type Stat struct {
	Name        string `json:"name"`
	Value       string `json:"value"`
	SectionName string `json:"section_name"`
}

type Mode struct {
	Won    int `json:"won"`
	Lost   int `json:"lost"`
	Played int `json:"played"`
	Time   string `json:"time"`
}

type Profile struct {
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
	Level    map[string]interface{} `json:"level"`
	Modes struct {
		Quickplay   map[string]interface{} `json:"quickplay"`
		Competitive map[string]interface{} `json:"competitive"`
	} `json:"modes"`
	Competitive map[string]interface{} `json:"competitive"`
}

type Account struct {
	CareerLink          string `json:"careerLink"`
	PlatformDisplayName string `json:"platformDisplayName"`
	Level               int `json:"level"`
	Portrait            string `json:"portrait"`
}

type HeroBreakdown struct {
	Hero       string `json:"hero"`
	Image      string `json:"image"`
	Value      string `json:"value"`
	Percentage float64 `json:"percentage"`
}

// GetAccountByName returns a list of matching profiles, in particular, profiles that match the given tag name.
func (p *Player) GetAccountByName() []Account {
	res, err := http.Get(p.formatSearchURL())
	if err != nil {
		panic(err.Error())
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err.Error())
	}

	var a = new([]Account)

	e := json.Unmarshal([]byte(body), &a)
	if e != nil {
		panic(e.Error())
	}

	return *a
}

// GetHeroHexMap returns a map of hero names and their associated hex value.
// The hex values can be used later as an id to navigate the DOM.
func GetHeroHexMap(d *goquery.Document) map[string]string {
	heroMap := map[string]string{}

	// Each child of the <select> element is an <option> with two notable attributes:
	// "option-id": hero name
	// "value": hero hex value
	d.Find("select[data-group-id='stats']").Children().Each(func(i int, s *goquery.Selection) {
		k, _ := s.Attr("option-id")
		v, _ := s.Attr("value")

		// For consistency, make all the keys (hero names) lowercase
		heroMap[strings.ToLower(k)] = v
	})

	return heroMap
}

// GetStatGUIDMap returns a map of stat names and their associated GUID.
// The GUID can be used later as an id to navigate the DOM.
func GetStatGUIDMap(d *goquery.Document) map[string]string {
	statCategoryMap := make(map[string]string)

	// Each child of the <select> element is an <option> with two notable attributes:
	// "option-id": stat name
	// "value": stat GUID
	d.Find("select[data-group-id='comparisons']").Children().Each(func(i int, s *goquery.Selection) {
		k, _ := s.Attr("option-id")
		v, _ := s.Attr("value")

		statCategoryMap[k] = v
	})

	return statCategoryMap
}

// SearchHandler retrieves all platform, region and tag combinations for the tag supplied and returns a JSON array of
// combinations.
// This is useful to see if a player has multiple profiles, or even, if the player exists.
func SearchHandler(w http.ResponseWriter, r *http.Request) {
	// Get tag from request URL.
	// Pack into a new Player for future use.
	vars := mux.Vars(r)
	p := Player{Tag: vars["tag"]}

	// Call helper method to get all matching profiles by account name (tag).
	searchResults := p.GetAccountByName()

	// Iterate over each search result.
	profiles := []map[string]string{}
	for _, v := range searchResults {
		tempMap := map[string]string{}

		// Career link will follow this format:
		// /career/pc/us/TAG
		// Strip off the initial "/" and then split the string at each "/".
		parts := strings.Split(v.CareerLink[1:], "/")

		// Pack the platform, region and tag into the tempMap.
		tempMap["platform"] = parts[1]
		tempMap["region"] = parts[2]
		tempMap["tag"] = parts[3]

		profiles = append(profiles, tempMap)
	}

	// Call helper function to marshal the slice to JSON.
	MarshalAndHandleErrors(w, r, profiles)
}

// AchievementsHandler retrieves all achievements for the given player and returns a JSON array of all achievements.
// This method will return all achievements, completed or not, but contains a field ("finished") to determine if the
// player completed the achievement.
func AchievementsHandler(w http.ResponseWriter, r *http.Request) {
	achievements := []Achievement{}

	// Get the platform, region and tag from the request URL.
	// Pack these into a new Player for future use.
	vars := mux.Vars(r)
	p := getPlayer(vars)

	// Find the parent achievement section, and iterate over all children (each achievement).
	p.GetProfileDoc().Find("#achievements-section .toggle-display .media-card").Each(func(i int, s *goquery.Selection) {
		imageURL, _ := s.ChildrenFiltered("img").Attr("src")
		title, _ := s.ChildrenFiltered(".media-card-caption").ChildrenFiltered(".media-card-title").Html()
		finished := s.HasClass("m-disabled")

		dataTooltip, _ := s.Attr("data-tooltip")
		description, _ := s.Parent().ChildrenFiltered("#" + dataTooltip).ChildrenFiltered("p").Html()

		achievement := Achievement{
			Title:       title,
			Description: description,
			ImageURL:    imageURL,
			Finished:    finished,
		}

		achievements = append(achievements, achievement)
	})

	// Call helper function to marshal the slice to JSON.
	MarshalAndHandleErrors(w, r, achievements)
}

// ProfileHandler retrieves the player's profile "overview", with statistics like player level, playtime, wins, etc.
// Returns a JSON object with this info and more.
func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	// Get the platform, region and tag from the request URL.
	// Pack these into a new Player for future use.
	vars := mux.Vars(r)
	p := getPlayer(vars)

	profile := Profile{}

	// Call helper method to get all matching profiles by account name (tag).
	accounts := p.GetAccountByName()

	// NOTE: GetAccountByName will return multiple results, so we need to iterate over the results to find the
	// matching profile
	matchingProfile := accounts[0]
	for _, account := range accounts {
		// Construct the 'careerLink' for the player searched for
		careerLink := "/career/" + p.Platform + "/" + p.Region + "/" + p.Tag

		// If the careerLink constructed matches the careerLink of the current account, we found the matching
		// account
		if account.CareerLink == careerLink {
			matchingProfile = account
		}
	}

	d := p.GetProfileDoc()

	// Maps to hold various data, broken down by logical sections.
	quickplayMap := make(map[string]interface{})
	competitiveMap := make(map[string]interface{})
	compRankMap := make(map[string]interface{})
	levelMap := make(map[string]interface{})

	username := d.Find(".header-masthead").Text()
	avatar, _ := d.Find(".player-portrait").Attr("src")

	quickplayGamesWon, _ := d.Find("#quickplay td:contains('Games Won')").Next().Html()
	if quickplayGamesWon != "" {
		quickplayMap["won"] = TrimToInt(quickplayGamesWon)
	}

	quickplayGamesPlayed, _ := d.Find("#quickplay td:contains('Games Played')").Next().Html()
	if quickplayGamesPlayed != "" {
		quickplayMap["played"] = TrimToInt(quickplayGamesPlayed)
	}

	quickplayTimePlayed, _ := d.Find("#quickplay td:contains('Time Played')").Next().Html()
	if quickplayTimePlayed != "" {
		quickplayMap["time"] = TrimToString(quickplayTimePlayed)
	}

	if quickplayGamesPlayed != "" && quickplayGamesWon != "" {
		quickplayMap["lost"] = TrimToInt(quickplayGamesPlayed) - TrimToInt(quickplayGamesWon)
	}

	competitiveGamesWon, _ := d.Find("#competitive td:contains('Games Won')").Next().Html()
	if competitiveGamesWon != "" {
		competitiveMap["won"] = TrimToInt(competitiveGamesWon)
	}

	competitiveGamesPlayed, _ := d.Find("#competitive td:contains('Games Played')").Next().Html()
	if competitiveGamesPlayed != "" {
		competitiveMap["played"] = TrimToInt(competitiveGamesPlayed)
	}

	competitiveTimePlayed, _ := d.Find("#competitive td:contains('Time Played')").Next().Html()
	if competitiveTimePlayed != "" {
		competitiveMap["time"] = TrimToString(competitiveTimePlayed)
	}

	if competitiveGamesPlayed != "" && competitiveGamesWon != "" {
		competitiveMap["lost"] = TrimToInt(competitiveGamesPlayed) - TrimToInt(competitiveGamesWon)
	}

	competitiveRankElm := d.Find(".competitive-rank")
	if competitiveRankElm != nil {
		rank, _ := d.Find(".competitive-rank div").Html()
		rankImg, _ := d.Find(".competitive-rank img").Attr("src")

		compRankMap["rank"] = TrimToString(rank)
		compRankMap["rank_img"] = TrimToString(rankImg)
	}

	levelElm := d.Find(".player-level")
	level := levelElm.First().Text()

	levelPortrait, _ := levelElm.Attr("style")
	// Format of the style is "background-image:url({URL})", so only take the slice with the actual URL in it.
	levelPortrait = levelPortrait[21:109]

	// TODO: Star Portrait

	levelMap["displayed"] = TrimToString(level)
	levelMap["actual"] = matchingProfile.Level
	levelMap["stars"] = CalculateStars(matchingProfile.Level)
	levelMap["portrait"] = levelPortrait

	profile.Username = username

	// If the player is on pc, their username is their BattleTag.
	// NOTE: BattleTags are sanitized ("#" -> "-") so we need to display the de-sanitized version.
	if p.Platform == "pc" {
		profile.Username = strings.Replace(p.Tag, "-", "#", 1)
	}

	profile.Avatar = avatar
	profile.Level = levelMap
	profile.Modes.Quickplay = quickplayMap
	profile.Modes.Competitive = competitiveMap
	profile.Competitive = compRankMap

	// Call helper function to marshal the slice to JSON.
	MarshalAndHandleErrors(w, r, profile)
}

// AllHeroStatsHandler retrieves the stats for all hero's combined and returns a JSON array of all stats and their
// section name.
func AllHeroStatsHandler(w http.ResponseWriter, r *http.Request) {
	var stats []Stat

	// Get the platform, region and tag from the request URL.
	// Pack these into a new Player for future use.
	vars := mux.Vars(r)
	p := getPlayer(vars)

	// Get mode from request URL.
	mode := strings.ToLower(vars["mode"])

	// Get each stat card (stat section). s will be each card.
	p.GetProfileDoc().Find("#" + mode + " .career-stats-section div .row[data-category-id='0x02E00000FFFFFFFF'] div").Children().Each(func(i int, s *goquery.Selection) {
		// Get the section name (i.e. "Combat", "Assists", etc).
		sectionName := s.Find(".card-stat-block > table > thead > tr > th .stat-title").Text()

		// Iterate over each row in the the table (each row of the stat section).
		s.Find(".card-stat-block table > tbody > tr").Each(func(j int, t *goquery.Selection) {
			statName, _ := t.Find("td:nth-child(1)").Html()
			statName = TrimToString(statName)

			statValue, _ := t.Find("td:nth-child(2)").Html()
			statValue = TrimToString(statValue)

			// statName might match the format: "overwatch.guid.XXXX..."
			// In this case, skip the stat.
			if strings.HasPrefix(statName, "overwatch.guid") {
				return
			}

			// A trailing 's' is added if the value of the stat is greater than 1.
			// If there is a trailing "s", replace it with "(s)".
			if statName[len(statName)-1:] == "s" {
				statName = statName[:len(statName)-1] + "(s)"
			}

			stat := Stat{
				Name:        statName,
				Value:       statValue,
				SectionName: sectionName,
			}

			stats = append(stats, stat)
		})
	})

	// Call helper function to marshal the slice to JSON.
	MarshalAndHandleErrors(w, r, stats)
}

// HerosHandler retrieves the breakdown of each stat by hero. Each stat is the key, and the value is a JSON array
// containing the value & percentage for each hero.
// Essentially, this method breaks-down each stat on a per-hero basis.
func HerosHandler(w http.ResponseWriter, r *http.Request) {
	statMap := make(map[string][]HeroBreakdown)

	// Get the platform, region and tag from the request URL.
	// Pack these into a new Player for future use.
	vars := mux.Vars(r)
	p := getPlayer(vars)

	// Get mode from request URL.
	mode := vars["mode"]

	d := p.GetProfileDoc()

	// Call helper function to get the GUID of each stat.
	// The GUID will be used to find the HTML of each stat.
	statGUIDMap := GetStatGUIDMap(d)

	row := d.Find("body > div > .profile-background > #" + mode + " > .hero-comparison-section .row.column")

	// Iterate over the map (each stat), and find the associated HTML nodes.
	for k, v := range statGUIDMap {
		// Temp slice to hold each hero's breakdown.
		var breakdownList []HeroBreakdown

		row.Find("div[data-category-id='" + v + "']").Children().Each(func(i int, bar *goquery.Selection) {
			percent, _ := bar.Attr("data-overwatch-progress-percent")
			percent = TrimToString(percent)

			image, _ := bar.ChildrenFiltered("img").Attr("src")
			image = TrimToString(image)

			heroName := bar.Find(".bar-container .bar-text .title").Text()
			heroName = TrimToString(heroName)

			value := bar.Find(".bar-container .bar-text .description").Text()
			value = TrimToString(value)

			breakdownList = append(breakdownList, HeroBreakdown{heroName, image, value, TrimToFloat(percent)})
		})

		// Key = stat name
		// Value = stat's slice (slice of hero breakdowns)
		statMap[k] = breakdownList
	}

	// Call helper function to marshal the slice to JSON.
	MarshalAndHandleErrors(w, r, statMap)
}

// HeroHandler retrieves the stats for the given hero (by name) and returns a JSON array of all of the stats and
// their section name.
// This method is similar to AllHeroStatsHandler, with the except that the stats shown are for the hero itself,
// rather that a combined total.
func HeroHandler(w http.ResponseWriter, r *http.Request) {
	var stats []Stat

	// Get the platform, region and tag from the request URL.
	// Pack these into a new Player for future use.
	vars := mux.Vars(r)
	p := getPlayer(vars)

	// Get mode and the hero name from request URL.
	mode := strings.ToLower(vars["mode"])
	heroName := strings.ToLower(vars["name"])

	d := p.GetProfileDoc()

	// Call helper function to get the hero hex map.
	// The hex of the hero will be used as an id to find the matching HTML.
	heroMap := GetHeroHexMap(d)

	// Find the stat section for the matching mode.
	row := d.Find("body > div > .profile-background > #" + mode + " > .career-stats-section > div")

	// Get the hex for the hero the user supplied.
	// TODO: Handle hero name not found
	hex := heroMap[heroName]

	// Use the hex to find the matching stat section (hero's section).
	// Then iterate over stat card (stat section).
	row.ChildrenFiltered(".row[data-category-id='" + hex + "']").Children().Each(func(i int, s *goquery.Selection) {
		// Get the section name (i.e. "Combat", "Assists", etc).
		sectionName := s.Find(".card-stat-block > table > thead > tr > th > .stat-title").Text()

		// Similar to AllHeroStatsHandler, iterate over each stat in the section's table.
		s.Find(".card-stat-block table > tbody > tr").Each(func(j int, t *goquery.Selection) {
			statName, _ := t.Find("td:nth-child(1)").Html()
			statName = TrimToString(statName)

			statValue, _ := t.Find("td:nth-child(2)").Html()
			statValue = TrimToString(statValue)

			// statName might match the format: "overwatch.guid.XXXX..."
			// In this case, skip the stat.
			if strings.HasPrefix(statName, "overwatch.guid") {
				return
			}

			// A trailing 's' is added if the value of the stat is greater than 1.
			// If there is a trailing "s", replace it with "(s)".
			if statName[len(statName)-1:] == "s" {
				statName = statName[:len(statName)-1] + "(s)"
			}

			stat := Stat{
				Name:        statName,
				Value:       statValue,
				SectionName: sectionName,
			}

			stats = append(stats, stat)
		})
	})

	// Call helper function to marshal the slice to JSON.
	MarshalAndHandleErrors(w, r, stats)
}

// ErrorHandler is a generic error handler to respond to various HTTP stats codes.
func ErrorHandler(w http.ResponseWriter, r *http.Request, statusCode int) {
	w.WriteHeader(statusCode)
	if statusCode == http.StatusNotFound {
		w.Write([]byte(ERROR_NOT_FOUND))
	}
}

// MarshalAndHandleErrors is a helper function to marshal data to JSON, while catching errors along the way.
func MarshalAndHandleErrors(w http.ResponseWriter, r *http.Request, res interface{}) {
	response, err := json.Marshal(res)
	if err != nil {
		panic(err)
	}

	if string(response) == "null" {
		ErrorHandler(w, r, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}
