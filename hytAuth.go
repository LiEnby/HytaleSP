package main

import (
	cryptorand "crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

const OAUTH_URL = "https://oauth.accounts.hytale.com/oauth2/auth";
const TOKEN_URL = "https://oauth.accounts.hytale.com/oauth2/token";
const REDIRECT_URI = "https://accounts.hytale.com/consent/client";


func openUrl(url string) {
	fmt.Printf("Opening url: %s\n", url);

	switch runtime.GOOS {
		case "linux":
			exec.Command("xdg-open", url).Start()
		case "windows":
			exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start();
		case "darwin":
			exec.Command("open", url).Start()
		default:
			fmt.Printf("Please open the url: \"%s\" in your browser.", url);
	}
}
func b64json(data any) string {
	jsonEnc, _ := json.Marshal(data);
	b64 := base64.URLEncoding.EncodeToString(jsonEnc);
	return b64;
}

func randomStr(l int) string {
	chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890";
	sb := strings.Builder{};

	for range l {
		sb.WriteByte(chars[rand.N(len(chars))]);
	}

	return sb.String();
}
func genVerifier() string {
	data := make([]byte, 32)
	if _, err := cryptorand.Read(data); err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(data)
}

func genChallengeS256(verifier string) string {
	sha := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sha[:])
}


func openOauthPage(portNum int) string  {
	oauthPage, _ := url.Parse(OAUTH_URL);

	state := oauthState {
		State: randomStr(26),
		Port: strconv.Itoa(portNum),
	}

	verifier := genVerifier();
	challenge := genChallengeS256(verifier);

	query := make(url.Values);
	query.Add("access_type", "offline");
	query.Add("client_id", "hytale-launcher");
	query.Add("code_challenge", challenge);
	query.Add("code_challenge_method", "S256");
	query.Add("redirect_uri", REDIRECT_URI);
	query.Add("response_type", "code");
	query.Add("scope", "openid offline auth:launcher");
	query.Add("state", b64json(state));
	oauthPage.RawQuery = query.Encode();

	urlToOpen := oauthPage.String();
	openUrl(urlToOpen);

	return verifier;
}


func doOauth() (code string, verifier string, err error) {

	c := make(chan string)

	server := http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

			code := req.URL.Query().Get("code");

			w.WriteHeader(200);
			w.Write([]byte("Trans rights! (you are very cute, oh also authenticated ig?)"));

			c <- code;

		}),
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", "", err;
	}
	go server.Serve(ln)

	port := ln.Addr().(*net.TCPAddr).Port;

	verifier = openOauthPage(port);
	code = <-c;

	return code, verifier, nil;
}


func verifyToken(verifier string, code string) (accessTokens, error) {

	authStr := "Basic " + base64.URLEncoding.EncodeToString([]byte("hytale-launcher:"));

	data := url.Values{};
	data.Add("code", code);
	data.Add("code_verifier", verifier);
	data.Add("grant_type", "authorization_code");
	data.Add("redirect_uri", REDIRECT_URI);

	// create new request
	req, err := http.NewRequest("POST", TOKEN_URL, strings.NewReader(data.Encode()));
	if err != nil {
		return accessTokens{}, err;
	}

	req.Header.Add("Authorization", authStr);
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded");
	req.Header.Set("User-Agent", "hytale-launcher/2026.02.12-54e579b");
	req.Header.Set("x-hytale-launcher-version", "2026.02.12-54e579b");
	req.Header.Set("x-hytale-launcher-branch", "release");

	// send response
	resp, err := http.DefaultClient.Do(req);
	if err != nil {
		return accessTokens{}, err;
	}
	if resp.StatusCode != 200 {
		return accessTokens{}, fmt.Errorf("%s got non-200 status: %d", TOKEN_URL, resp.StatusCode);
	}

	tokens := accessTokens{};
	json.NewDecoder(resp.Body).Decode(&tokens);

	return tokens, nil;

}

func refreshToken(refreshToken string) (aToken accessTokens, err error){

	data := url.Values{}
	data.Set("client_id", "hytale-launcher")
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)

	req, _ := http.NewRequest("POST", TOKEN_URL, strings.NewReader(data.Encode()));
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded");
	req.Header.Set("User-Agent", "hytale-launcher/2026.02.12-54e579b");
	req.Header.Set("x-hytale-launcher-version", "2026.02.12-54e579b");
	req.Header.Set("x-hytale-launcher-branch", "release");

	resp, err := http.DefaultClient.Do(req);
	if err != nil {
		return accessTokens{}, err;
	}
	if resp.StatusCode != 200 {
		return accessTokens{}, fmt.Errorf("%s got non-200 status: %d", TOKEN_URL, resp.StatusCode);
	}

	newTokens := accessTokens{};
	json.NewDecoder(resp.Body).Decode(&newTokens);

	return newTokens, nil;
}

func getAuthTokens(previousTokens any) (atoken accessTokens, err error) {
	prevTokens, ok := previousTokens.(accessTokens);

	if ok {
		aToken, err := refreshToken(prevTokens.RefreshToken);
		if err != nil {
			return accessTokens{}, err;
		}
		return aToken, nil;
	} else {
		fmt.Println("Getting new tokens");
		code, verifier, err := doOauth();
		if err != nil {
			return accessTokens{}, err;
		}

		aToken, err := verifyToken(verifier, code);
		if err != nil {
			return accessTokens{}, err;
		}
		return aToken, err;
	}
}


