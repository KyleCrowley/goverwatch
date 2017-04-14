package main

const BASE_URL = "https://playoverwatch.com/en-us/career/"
const SEARCH_URL = "https://playoverwatch.com/search/account-by-name/"
const PATCH_NOTE_URL = "https://cache-eu.battle.net/system/cms/oauth/api/patchnote/list?program=pro&region=US&locale=enUS&type=RETAIL&page=1&pageSize=5&orderBy=buildNumber&buildNumberMin=0"

const NOT_FOUND_TEXT = "Could not find a user with that platform, region and BattleTag combination."

const ERROR_BAD_PLATFORM = "Invalid platform supplied. Must be one of the following: pc, psn, xbl."
const ERROR_BAD_REGION = "Invalid region supplied. Must be one of the following: us, eu, cn, kr, global."
const ERROR_BAD_TAG = "Invalid tag supplied."
const ERROR_BAD_MODE = "Invalid mode. Must be one of the following: quickplay, competitive."

var PLATFORMS = map[string]bool{"pc": true, "psn": true, "xbl": true}
var REGIONS = map[string]bool{"us": true, "eu": true, "cn": true, "kr": true, "global": true}
var MODES = map[string]bool{"quickplay": true, "competitive": true}
