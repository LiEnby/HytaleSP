package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/c4milo/unpackit"
)


func urlToPath(targetUrl string) string {
	nurl, _ := url.Parse(targetUrl);
	npath := strings.TrimPrefix(nurl.Path, "/");
	return npath;
}

func download(targetUrl string, saveFilename string, onProgress func(done int64, total int64)) error {
	fmt.Printf("Downloading %s\n", targetUrl);

	os.MkdirAll(filepath.Dir(saveFilename), 0775);
	resp, err := http.Get(targetUrl);
	if err != nil {
		return err;
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("%s got non-200 status: %d", targetUrl, resp.StatusCode);
	}

	f, err := os.Create(saveFilename);
	if err != nil {
		return err;
	}
	defer f.Close();

	total := resp.ContentLength;
	done := int64(0);
	buffer := make([]byte, 0x8000);

	for done < total {
		rd, err := resp.Body.Read(buffer);
		if err != nil {
			return err;
		}
		done += int64(rd);
		f.Write(buffer[:rd]);
		onProgress(done, total);
	}

	return nil;
}

func getVersionDownloadsFolder() string {
	fp := filepath.Join(GameFolder(), "download");
	return fp;
}

func getVersionDownloadPath(startVersion int, endVersion int, channel string) string {
	fp := filepath.Join(getVersionDownloadsFolder(), channel, strconv.Itoa(endVersion), strconv.Itoa(startVersion) + "-" + strconv.Itoa(endVersion)+".pwr");
	return fp;
}

func getVersionsFolder(channel string) string {
	fp := filepath.Join(GameFolder(), channel);
	return fp;
}

func getVersionInstallPath(endVersion int, channel string) string {
	fp := filepath.Join(getVersionsFolder(channel), strconv.Itoa(endVersion));
	return fp;
}

func getJrePath(operatingSystem string, architecture string) string {
	fp := filepath.Join(JreFolder(), operatingSystem, architecture);
	return fp;
}

func getJreDownloadPath(operatingSystem string) string {
	ext := ".zip";
	if operatingSystem == "linux" || operatingSystem == "darwin" {
		ext = ".tar.gz";
	}

	fp := filepath.Join(JreFolder(), "download", "jre-download"+ext);
	return fp;
}


func downloadLatestVersion(atokens accessTokens, architecture string, operatingSystem string, channel string, fromVersion int, onProgress func(done int64, total int64)) error {
	fmt.Printf("Start version: %d\n", fromVersion);
	manifest, err := getVersionManifest(atokens, architecture, operatingSystem, channel, fromVersion);

	if(err != nil) {
		return err;
	}

	for _, step := range manifest.Steps {
		save := getVersionDownloadPath(step.From, step.To, channel);
		return download(step.Pwr, save, onProgress);
	}
	return errors.New("Could not locate latest version");
}


func isJreInstalled() bool {
	javaBin, ok := findJavaBin().(string);
	if ok {
		_, err := os.Stat(javaBin);
		if err != nil {
			return false;
		}
		return true;
	} else {
		return false;
	}
}

func isGameVersionInstalled(version int, channel string) bool {
	gameDir := findClientBinary(version, channel);
	_, err := os.Stat(gameDir);
	if err != nil {
		return false;
	}
	return true;
}


func verifyFileSha256(fp string, expected string) bool {
	file, err := os.Open(fp)
	if err != nil {
		return false;
	}
	defer file.Close();

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return false;
	}
	digest := hash.Sum(nil);

	return strings.EqualFold(hex.EncodeToString(digest), strings.ToLower(expected));
}

func installJreEx(operatingSystem string, out string, temp string, onProgress func(done int64, total int64)) error {
	jres, err := getJres("release");
	if err != nil {
		return err;
	}

	var downloadUrl string;

	switch(operatingSystem) {
		case "windows":
			downloadUrl = jres.DownloadUrls.Windows.Amd64.URL;
		case "linux":
			downloadUrl = jres.DownloadUrls.Linux.Amd64.URL;
		case "darwin":
			downloadUrl = jres.DownloadUrls.Darwin.Amd64.URL;

	}

	err = download(downloadUrl, temp, onProgress);
	defer os.Remove(temp);
	defer os.RemoveAll(filepath.Dir(temp));


	if err != nil {
		return err;
	}

	valid := false;

	// validate jre
	switch(operatingSystem) {
		case "windows":
			valid = verifyFileSha256(temp, jres.DownloadUrls.Windows.Amd64.Sha256);
		case "linux":
			valid = verifyFileSha256(temp, jres.DownloadUrls.Linux.Amd64.Sha256);
		case "darwin":
			valid = verifyFileSha256(temp, jres.DownloadUrls.Darwin.Arm64.Sha256);
	}

	if valid == false {
		return fmt.Errorf("Could not validate the SHA256 hash for the JRE runtime.");
	}

	os.MkdirAll(out, 0775);

	f, err := os.Open(temp);
	if err != nil {
		os.RemoveAll(out);
		return err;
	}

	err = unpackit.Unpack(f, out);

	if(err != nil) {
		os.RemoveAll(out);
		return err;
	}

	return nil;
}

func installJre(onProgress func(done int64, total int64)) error{
	temp := getJreDownloadPath(runtime.GOOS);
	out := getJrePath(runtime.GOOS, runtime.GOARCH);

	return installJreEx(runtime.GOOS, out, temp, onProgress);
}

func findClosestVersion(targetVersion int, channel string) int {
	installFolder := getVersionsFolder(channel);

	cloestVersion := 0;

	d, err := os.ReadDir(installFolder);
	if err != nil {
		return cloestVersion;
	}

	for _, e := range d {
		if !e.IsDir() {
			continue;
		}

		ver, err := strconv.Atoi(e.Name());

		if err != nil {
			continue;
		}

		if ver > cloestVersion && ver < targetVersion {
			cloestVersion = ver;
		}
	}

	return cloestVersion;

}

func downloadAndRunGameEx(version int, channel string,
			  onJreInstalled func(),
			  onNewVersionInstalled func(channel string, ver int),
			  onGameStarting func(),
			  onGameClosing func(),
			  onProgress func(done int64, total int64)) error {
	// download new jre if not installed ..

	if !isJreInstalled() {
		err := installJre(updateProgress);

		if err != nil {
			return fmt.Errorf("Error getting the JRE: %s", err);
		};

		onJreInstalled();
	}

	// download new version if not installed ..
	if !isGameVersionInstalled(version, channel) {
		err := installGame(version, channel, onProgress);

		if err != nil {
			return fmt.Errorf("Error getting %s version %d: %s", channel, version, err);
		};

		onNewVersionInstalled(channel, version);
	}

	// start the game
	onGameStarting();
	defer onGameClosing();

	err := launchGame(version, channel, wCommune.Username, getUUID());

	if err != nil {
		return fmt.Errorf("Error starting the game: %s", err);
	};

	return nil;

}


func downloadAndRunGame(version int, channel string, onProgress func(done int64, total int64)) error {
	return downloadAndRunGameEx(version, channel, func(){}, func(string, int){}, func(){}, func(){}, cliProgressUpdate);
}

func installGameEx(startVersion int, endVersion int, channel string, operatingSystem string, architecture string, src string, out string, temp string, onProgress func(done int64, total int64)) error {

	downloadUrl := guessPatchUrlNoAuth(architecture, operatingSystem, channel, startVersion, endVersion);
	downloadSig := guessPatchSigUrlNoAuth(architecture, operatingSystem, channel, startVersion, endVersion);

	sigPath := temp + ".sig";

	err := download(downloadUrl, temp, onProgress);
	defer os.Remove(temp);
	defer os.Remove(sigPath);
	defer os.RemoveAll(getVersionDownloadsFolder());

	if err != nil {
		return err;
	}

	err = download(downloadSig, sigPath, onProgress);
	defer os.Remove(sigPath);

	if err != nil {
		return err;
	}
	os.MkdirAll(out, 0775);

	fmt.Printf("Applying patch %s, using source: %s to: %s\n", temp, src, out);
	err = applyPatch(src, out, temp, sigPath, onProgress);
	if err != nil {
		return err;
	}

	return nil;
}

func deleteVersion(version int, channel string) error{

	installDir := getVersionInstallPath(version, channel);
	err := os.RemoveAll(installDir);
	if err != nil {
		return err;
	}

	return nil;
}

func workaround(version int, channel string, onProgress func(done int64, total int64)) bool {
	if version == 21 && channel == "pre-release" && !isGameVersionInstalled(20, "pre-release") {

		err := installGame(20, "pre-release", onProgress);
		if err != nil {
			return false;
		}
		defer deleteVersion(20, channel);


		err = installGame(21, "pre-release", onProgress);
		if err != nil {
			return false;
		}
		return true;
	}

	return false;
}

func installGame(version int, channel string, onProgress func(done int64, total int64)) error {
	if(workaround(version, channel, onProgress)) {
		return nil;
	}


	closest := findClosestVersion(version, channel);
	src := getVersionInstallPath(closest, channel);
	apply := getVersionInstallPath(version, channel);
	temp := getVersionDownloadPath(closest, version, channel);

	// check if incremental patch file exists ...
	if !checkVerExist(closest, version, runtime.GOARCH, runtime.GOOS, channel) {
		closest = 0
		src = getVersionInstallPath(closest, channel);
		apply = getVersionInstallPath(version, channel);
		temp = getVersionDownloadPath(closest, version, channel);
	}


	return installGameEx(closest, version, channel, runtime.GOOS, runtime.GOARCH, src, apply, temp, onProgress);
}

func findJavaBin() any {
	jrePath := getJrePath(runtime.GOOS, runtime.GOARCH);

	d, err := os.ReadDir(jrePath);
	if err != nil {
		return nil;
	}

	for _, e := range d {
		if !e.IsDir() {
			continue;
		}

		if runtime.GOOS == "windows" {
			return filepath.Join(jrePath, e.Name(), "bin", "java.exe");
		} else {
			return filepath.Join(jrePath, e.Name(), "bin", "java");
		}
	}

	return nil;
}


func findClientBinary(version int, channel string) string {
	clientFolder := filepath.Join(getVersionInstallPath(version, channel), "Client");

	switch(runtime.GOOS) {
		case "windows":
			return filepath.Join(clientFolder, "HytaleClient.exe");
		case "darwin":
			return filepath.Join(clientFolder, "Hytale.app", "Contents", "MacOS", "HytaleCleint");
		case "linux":
			return filepath.Join(clientFolder, "HytaleClient");
		default:
			panic("Hytale is not supported by your OS.");
	}
}

func launchGame(version int, channel string, username string, uuid string) error{

	javaBin, _ := findJavaBin().(string);

	appDir := getVersionInstallPath(version, channel)
	userDir := UserDataFolder()
	clientBinary := findClientBinary(version, channel);

	// create user directory
	os.MkdirAll(userDir, 0775);



	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		os.Chmod(javaBin, 0775);
		os.Chmod(clientBinary, 0775);
	}

	// remove fakeonline patch if present.
	if runtime.GOOS == "windows" {
		dllName := filepath.Join(filepath.Dir(clientBinary), "Secur32.dll");
		os.Remove(dllName);
	}

	if wCommune.Mode == E_MODE_FAKEONLINE { // start with fake online mode

		// setup fake online patch
		go runServer();

		var dllName string;
		var embedName string;

		if runtime.GOOS == "windows" {
			dllName = filepath.Join(filepath.Dir(clientBinary), "Secur32.dll");
			embedName = path.Join("Aurora", "Build", "Aurora.dll");
		}

		if runtime.GOOS == "linux" {
			dllName = filepath.Join(os.TempDir(), "Aurora.so");
			embedName = path.Join("Aurora", "Build", "Aurora.so");
		}

		// write fakeonline dll
		data, err := embeddedFiles.ReadFile(embedName);
		if err != nil {
			return errors.New("read embedded Aurora dll -- Try offline mode.");
		}
		os.WriteFile(dllName, data, 0777);
		defer os.Remove(dllName);

		// start the client

		e := exec.Command(clientBinary,
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
			generateIdentityJwt([]string{"hytale:client"}),
			"--session-token",
			generateSessionJwt([]string{"hytale:client"}));



		switch(runtime.GOOS) {
			case "linux":
				os.Setenv("LD_PRELOAD", dllName);
			case "darwin":
				os.Setenv("DYLD_INSERT_LIBRARIES ", dllName);
		}

		fmt.Printf("Running: %s\n", strings.Join(e.Args, " "))

		e.Env = []string {
			"AURORA_ENABLE_INSECURE_SERVERS=true",
			"AURORA_ENABLE_AUTH_SWAP=true",
			"AURORA_ENABLE_SINGLEPLAYER_AS_INSECURE=true",
			"AURORA_SESSIONS=http://127.0.0",
			"AURORA_ACCOUNT_DATA=http://127.0.0",
			"AURORA_TOOLS=http://127.0.0",
			"AURORA_TELEMETRY=http://127.0.0",
			"AURORA_HYTALE_COM=.1:59313",
		}

		err = e.Start();

		if err != nil {
			return err;
		}

		switch(runtime.GOOS) {
			case "linux":
				os.Unsetenv("LD_PRELOAD");
			case "darwin":
				os.Unsetenv("DYLD_INSERT_LIBRARIES ");
		}

		e.Process.Wait();

	} else if wCommune.Mode == E_MODE_AUTHENTICATED { // start authenticated
		if wCommune.AuthTokens == nil {
			return errors.New("No auth token found.");
		}

		if wCommune.Profiles == nil {
			return errors.New("Could not find a profile");
		}

		// get currently selected profile
		profileList := *wCommune.Profiles;
		profile := profileList[wCommune.SelectedProfile];

		newSess, err := getNewSession(*wCommune.AuthTokens, profile.UUID);
		if(err != nil) {
			return err;
		}

		e := exec.Command(clientBinary,
			"--app-dir",
			appDir,
			"--user-dir",
			userDir,
			"--java-exec",
			javaBin,
			"--auth-mode",
			"authenticated",
			"--uuid",
			profile.UUID,
			"--name",
			profile.Username,
			"--identity-token",
			newSess.IdentityToken,
			"--session-token",
			newSess.SessionToken);

		fmt.Printf("Running: %s\n", strings.Join(e.Args, " "))

		err = e.Start();

		if err != nil {
			return err;
		}

		e.Process.Wait();
	} else { // start in offline mode

		e := exec.Command(clientBinary,
			"--app-dir",
			appDir,
			"--user-dir",
			userDir,
			"--java-exec",
			javaBin,
			"--auth-mode",
			"offline",
			"--uuid",
			uuid,
			"--name",
			username);

		fmt.Printf("Running: %s %s\n", clientBinary, strings.Join(e.Args, " "))

		err := e.Start();

		if err != nil {
			return err;
		}


		e.Process.Wait();
	}
	return nil;
}
