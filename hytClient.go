package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const ACCOUNT_DATA_URL = "https://account-data.hytale.com/";
// const GAME_PATCHES_URL = "https://game-patches.hytale.com/";
// no longer works :c

const GAME_PATCHES_URL = "https://vitatricks.xyz/hytale";
const LAUNCHER_URL     = "https://launcher.hytale.com/"
const SESSIONS_URL     = "https://sessions.hytale.com/"


func guessPatchSigUrlNoAuth(architecture string, operatingSystem string, channel string, startVersion int, targetVersion int) string{
	fullUrl, _ := url.JoinPath(GAME_PATCHES_URL, "patches", operatingSystem, architecture, channel, strconv.Itoa(startVersion), strconv.Itoa(targetVersion) + ".pwr.sig");
	return fullUrl;
}
func guessPatchUrlNoAuth(architecture string, operatingSystem string, channel string, startVersion int, targetVersion int) string{
	fullUrl, _ := url.JoinPath(GAME_PATCHES_URL, "patches", operatingSystem, architecture, channel, strconv.Itoa(startVersion), strconv.Itoa(targetVersion) + ".pwr");
	return fullUrl;
}

func getNewSession(atokens accessTokens, uuid string) (sessionNew, error) {
	fullUrl, _ := url.JoinPath(SESSIONS_URL, "game-session", "new");


	nSess := sessNewRequest{
		UUID: uuid,
	};

	j, err := json.Marshal(&nSess);

	if err != nil {
		return sessionNew{}, err;
	}

	req, err:= http.NewRequest("POST", fullUrl, bytes.NewReader(j));

	if err != nil {
		return sessionNew{}, err;
	}


	req.Header.Add("Authorization", "Bearer " + atokens.AccessToken);
	req.Header.Add("Content-Type", "application/json");

	n := sessionNew{};

	resp, err := http.DefaultClient.Do(req);

	if err != nil {
		return sessionNew{}, err;
	}

	if resp.StatusCode != 200 {
		return sessionNew{}, fmt.Errorf("%s got non-200 status: %d", fullUrl, resp.StatusCode);
	}

	json.NewDecoder(resp.Body).Decode(&n);
	return n, nil;


}


func getJres(channel string) (versionFeed, error) {
	fullUrl, _ := url.JoinPath(LAUNCHER_URL, "version", channel, "jre.json");

	resp, err := http.Get(fullUrl);
	if err != nil{
		return versionFeed{}, err;;
	}

	if resp.StatusCode != 200 {
		return versionFeed{}, fmt.Errorf("%s got non-200 status: %d", fullUrl, resp.StatusCode);
	}

	feed := versionFeed{};
	json.NewDecoder(resp.Body).Decode(&feed);

	return feed, nil;
}

func getLaunchers(channel string) (versionFeed, error) {
	fullUrl, _ := url.JoinPath(LAUNCHER_URL, "version", channel, "launcher.json");

	resp, err := http.Get(fullUrl);

	if err != nil{
		return versionFeed{}, err;;
	}
	if resp.StatusCode != 200 {
		return versionFeed{}, fmt.Errorf("%s got non-200 status: %d", fullUrl, resp.StatusCode);
	}

	feed := versionFeed{};
	json.NewDecoder(resp.Body).Decode(&feed);

	return feed, nil;

}

func getLauncherData(atokens accessTokens, architecture string, operatingSystem string) (launcherData, error) {

	fullUrl, _ := url.JoinPath(ACCOUNT_DATA_URL, "my-account", "get-launcher-data");
	launcherDataUrl, _ := url.Parse(fullUrl);

	query := make(url.Values);
	query.Add("arch", architecture);
	query.Add("os", operatingSystem);

	launcherDataUrl.RawQuery = query.Encode();

	req, _:= http.NewRequest("GET", launcherDataUrl.String(), nil);

	req.Header.Add("Authorization", "Bearer " + atokens.AccessToken);
	req.Header.Add("Content-Type", "application/json");

	resp, err := http.DefaultClient.Do(req);

	if err != nil {
		return launcherData{}, err;
	}

	if resp.StatusCode != 200 {
		return launcherData{}, fmt.Errorf("%s got non-200 status: %d", fullUrl, resp.StatusCode);
	}


	ldata := launcherData{};
	json.NewDecoder(resp.Body).Decode(&ldata);

	return ldata, nil;
}

func getVersionManifest(atokens accessTokens, architecture string, operatingSystem string, channel string, gameVersion int) (versionManifest, error) {
	fullUrl, _ := url.JoinPath(ACCOUNT_DATA_URL, "patches", operatingSystem, architecture, channel, strconv.Itoa(gameVersion));

	req, _:= http.NewRequest("GET", fullUrl, nil);

	req.Header.Add("Authorization", "Bearer " + atokens.AccessToken);
	req.Header.Add("Content-Type", "application/json");

	resp, err := http.DefaultClient.Do(req);

	if err != nil {
		return versionManifest{}, err;
	}
	if resp.StatusCode != 200 {
		return versionManifest{}, fmt.Errorf("%s got non-200 status: %d", fullUrl, resp.StatusCode);
	}

	mdata := versionManifest{};
	json.NewDecoder(resp.Body).Decode(&mdata);

	return mdata, nil;
}



func checkVerExist(startVersion int, endVersion int, architecture string, operatingSystem string, channel string) bool {
	for range 5 {
		uri := guessPatchUrlNoAuth(architecture, operatingSystem, channel, startVersion, endVersion);
		req, err := http.Head(uri);
		if err != nil {
			return false;
		}


		switch(req.StatusCode) {
			case 200:
				return true;
			case 404:
				return false;
			default:
				time.Sleep(time.Second * 3);
				continue;
		}
	}
	return false;
}


func findLatestVersionNoAuth(current int, architecture string, operatingSystem string, channel string) int {

	// obtaining the latest version from hytale CDN (as well as its 'pretty' name)
	// requires authentication to hytale servers,
	// however downloading versions does not,
	// this is an optimized search alogirithm to find the latest version
	//
	// it makes a few assumptions; mainly-
	// - there are never gaps in version numbers
	// - the url scheme of version downloads is .. os/arch/channel/startver/destver.pwr
	// if hytale ever changes how they handle this, then everything will break.


	if current <= 0 {
		current = 1;
	}

	lastVersion := current;
	curVersion := current;

	// check if has been updates since this; no point if no new versions are added
	if checkVerExist(0, current+1, architecture, operatingSystem, channel) {

		// multiply version number by 2 until a version is not found ..
		for checkVerExist(0, curVersion, architecture, operatingSystem, channel) {
			lastVersion = curVersion;
			curVersion *= 2;
		}

		// binary search from last valid, to largest invalid;
		for lastVersion+1 < curVersion {
			middle := (curVersion + lastVersion) /2;
			if checkVerExist(0, middle, architecture, operatingSystem,channel) {
				lastVersion = middle;
			} else {
				curVersion = middle;
			}
		}
	}


	return lastVersion;
}
