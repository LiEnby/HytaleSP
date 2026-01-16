package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strconv"
	"strings"
)

var HOME_FOLDER, _ = os.UserHomeDir();
var MAIN_FOLDER = path.Join(HOME_FOLDER, "hytLauncher");
var GAME_FOLDER = path.Join(MAIN_FOLDER, "game", "versions");
var USERDATA_FOLDER = path.Join(MAIN_FOLDER, "userdata");
var JRE_FOLDER = path.Join(MAIN_FOLDER, "jre");

func urlToPath(targetUrl string) string {
	nurl, _ := url.Parse(targetUrl);
	npath := strings.TrimPrefix(nurl.Path, "/");
	return npath;
}

func download(targetUrl string, saveFilename string) any {
	fmt.Printf("Downloading %s\n", targetUrl);

	os.MkdirAll(path.Dir(saveFilename), 0666);
	_, err := os.Stat(saveFilename);

	if err != nil {
		resp, err := http.Get(targetUrl);
		if err != nil {
			return nil;
		}

		if resp.StatusCode == 200 {
			f, _ := os.Create(saveFilename);
			io.Copy(f, resp.Body);
			defer f.Close();
		}

	}

	return saveFilename;
}


func getVersionDownloadPath(startVersion int, endVersion int, channel string) string {
	filepath := path.Join(GAME_FOLDER, channel, strconv.Itoa(endVersion), "download", strconv.Itoa(startVersion) + "-" + strconv.Itoa(endVersion)+".pwr");
	return filepath;
}

func getVersionInstallPath(endVersion int, channel string) string {
	filepath := path.Join(GAME_FOLDER, channel, strconv.Itoa(endVersion));
	return filepath;
}

func getJrePath(operatingSystem string, architecture string) string {
	filepath := path.Join(JRE_FOLDER, operatingSystem, architecture);
	return filepath;
}

func getJreDownloadPath(operatingSystem string, architecture string, downloadUrl string) string {
	u, _ := url.Parse(downloadUrl);
	filepath := path.Join(getJrePath(operatingSystem, architecture), "download", path.Base(u.Path));
	return filepath;
}


func downloadLatestVersion(atokens accessTokens, architecture string, operatingSystem string, channel string, fromVersion int) any {
	fmt.Printf("Start version: %d\n", fromVersion);
	manifest := getVersionManifest(atokens, architecture, operatingSystem, channel, fromVersion);
	for _, step := range manifest.Steps {
		save := getVersionDownloadPath(step.From, step.To, channel);
		return download(step.Pwr, save);
	}
	return nil;
}


func installJre() any {
	jres := getJres("release");

	if runtime.GOOS == "windows" && runtime.GOARCH == "amd64" {
		downloadUrl := jres.DownloadUrls.Windows.Amd64.URL;
		save := getJreDownloadPath(runtime.GOOS, runtime.GOARCH, downloadUrl);
		unpack := getJrePath(runtime.GOOS, runtime.GOARCH);

		_, ok := os.Stat(unpack);

		if ok != nil {
			_, ok := download(downloadUrl, save).(string);
			if ok {
				unzip(save, unpack);
				os.RemoveAll(path.Dir(save));
				return unpack;
			}
		}

	} else {
		fmt.Println("There is no version of Hytale for your operating system :C");
	}
	return nil;
}

func installGame(version int, channel string) any {
	save := getVersionDownloadPath(0, version, channel);
	unpack := getVersionInstallPath(version, channel);

	if runtime.GOOS == "windows" && runtime.GOARCH == "amd64" {
		_, ok := os.Stat(unpack);

		if ok != nil {
			downloadUrl := guessPatchUrlNoAuth(runtime.GOARCH, runtime.GOOS, channel, 0, version);
			pwr := download(downloadUrl, save);

			if _, ok := pwr.(string); ok {
				applyPatch(unpack, unpack, save);

				// delete pwr file ..
				os.Remove(save);
				// delete download folder
				os.RemoveAll(path.Dir(save));
				return unpack;
			}
		}
	}
	return nil;
}

func findJavaBin() any {
	jrePath := getJrePath(runtime.GOOS, runtime.GOARCH);

	d, err := os.ReadDir(jrePath);
	if err != nil {
		fmt.Printf("err: %s\n", err);
		os.Exit(0);
	}

	for _, e := range d {
		if !e.IsDir() {
			continue;
		}

		if runtime.GOOS == "windows" {
			return path.Join(jrePath, e.Name(), "bin", "java.exe");
		} else {
			return path.Join(jrePath, e.Name(), "bin", "java");
		}
	}

	return nil;
}

func launchGame(version int, channel, username string, uuid string) {

	j, _ := findJavaBin().(string);


	if runtime.GOOS == "windows" {
		appDir := strings.ReplaceAll(getVersionInstallPath(version, channel), "\\", "/");
		userDir := strings.ReplaceAll(USERDATA_FOLDER, "\\", "/");
		javaBin := strings.ReplaceAll(j, "\\", "/");
		hytaleClientBin := strings.ReplaceAll(path.Join(appDir, "Client", "HytaleClient.exe"), "/", "\\");

		e := exec.Command(hytaleClientBin,
				"--app-dir",
				appDir,
				"--user-dir",
				userDir,
				"--java-exec",
				javaBin,
				"--auth-mode",
				"authenticated",
				"--uuid",
				uuid,
				"--name",
				username,
				"--identity-token",
				generateIdentityJwt(),
				"--session-token",
				generateSessionJwt());

		e.Start();

		runServer(username, uuid);

	}



}

func usernameToUuid(username string) string {
	m := md5.New();
	m.Write([]byte(username));
	h := hex.EncodeToString(m.Sum(nil));

	return h[:8]+"-"+h[8:12]+"-"+h[12:16]+"-"+h[16:20]+"-"+h[20:32];
}

func main() {

	os.MkdirAll(MAIN_FOLDER, 0666);
	os.MkdirAll(GAME_FOLDER, 0666);
	os.MkdirAll(USERDATA_FOLDER, 0666);
	os.MkdirAll(JRE_FOLDER, 0666);

	installJre();

	if len(os.Args) >= 4 {
		version, err := strconv.Atoi(os.Args[1]);
		channel := os.Args[2];
		username := os.Args[3];
		uuid := usernameToUuid(username);

		if err != nil {
			fmt.Printf("err parsing version: %s\n", err);
			os.Exit(0);
		}

		installGame(version, channel);
		launchGame(version, channel, username, uuid);
	}

}
