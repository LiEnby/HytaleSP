package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

const ACCOUNT_DATA_URL = "https://account-data.hytale.com/";
// const GAME_PATCHES_URL = "https://game-patches.hytale.com/";
// no longer works :c

const GAME_PATCHES_URL = "\x68\x74\x74\x70\x73\x3a\x2f\x2f\x61\x63\x63\x6f\x75\x6e\x74\x2d\x64\x61\x74\x61\x2e\x68\x79\x74\x61\x6c\x65\x2e\x63\x6f\x6d\x2f";
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

	req, err:= createRequest("POST", fullUrl, &nSess);

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
		return sessionNew{}, fmt.Errorf("%s got non-200 status: %s", fullUrl, resp.Status);
	}

	json.NewDecoder(resp.Body).Decode(&n);
	return n, nil;


}


func getJres(channel string) (versionFeed, error) {
	fullUrl, _ := url.JoinPath(LAUNCHER_URL, "version", channel, "jre.json");

	req, err := createRequest("GET", fullUrl, nil);
	resp, err := http.DefaultClient.Do(req);

	if err != nil{
		return versionFeed{}, err;;
	}

	if resp.StatusCode != 200 {
		return versionFeed{}, fmt.Errorf("%s got non-200 status: %s", fullUrl, resp.Status);
	}

	feed := versionFeed{};
	json.NewDecoder(resp.Body).Decode(&feed);

	return feed, nil;
}

func getLatestLauncherVersion() string {
	latestLauncher, err := getLauncherInfo("release");
	if err != nil {
		return "2026.02.12-54e579b";
	}
	return latestLauncher.Version;
}

func getLauncherInfo(channel string) (versionFeed, error) {
	fullUrl, _ := url.JoinPath(LAUNCHER_URL, "version", channel, "launcher.json");

	req, err := createRequest("GET", fullUrl, nil);
	resp, err := http.DefaultClient.Do(req);

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

func getVersionInformation(atokens accessTokens, architecture string, operatingSystem string) (launcherData, error) {

	fullUrl, _ := url.JoinPath(ACCOUNT_DATA_URL, "my-account", "get-launcher-data");
	launcherDataUrl, _ := url.Parse(fullUrl);

	query := make(url.Values);
	query.Add("arch", architecture);
	query.Add("os", operatingSystem);

	launcherDataUrl.RawQuery = query.Encode();

	req, _:= createRequest("GET", launcherDataUrl.String(), nil);

	req.Header.Add("Authorization", "Bearer " + atokens.AccessToken);
	req.Header.Add("Content-Type", "application/json");

	resp, err := http.DefaultClient.Do(req);

	if err != nil {
		return launcherData{}, err;
	}

	if resp.StatusCode != 200 {
		return launcherData{}, fmt.Errorf("%s got non-200 status: %s", fullUrl, resp.Status);
	}


	ldata := launcherData{};
	json.NewDecoder(resp.Body).Decode(&ldata);

	return ldata, nil;
}

func getVersionManifest(atokens accessTokens, architecture string, operatingSystem string, channel string, gameVersion int) (versionManifest, error) {
	fullUrl, _ := url.JoinPath(ACCOUNT_DATA_URL, "patches", operatingSystem, architecture, channel, strconv.Itoa(gameVersion));

	req, _:= createRequest("GET", fullUrl, nil);

	req.Header.Add("Authorization", "Bearer " + atokens.AccessToken);
	req.Header.Add("Content-Type", "application/json");

	resp, err := http.DefaultClient.Do(req);

	if err != nil {
		return versionManifest{}, err;
	}
	if resp.StatusCode != 200 {
		return versionManifest{}, fmt.Errorf("%s got non-200 status: %s", fullUrl, resp.Status);
	}

	mdata := versionManifest{};
	json.NewDecoder(resp.Body).Decode(&mdata);

	return mdata, nil;
}


