package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
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


func run(e *exec.Cmd) error {

	stdout, oerr := e.StdoutPipe();
	stderr, eerr := e.StderrPipe();

	err := e.Start();

	if err != nil {
		return err;
	}

	if wVersion == "no-version" || (wCommune.Console && runtime.GOOS != "windows") {
		go func() {
			if oerr == nil {
				stdout_scan := bufio.NewScanner(stdout)
				stdout_scan.Split(bufio.ScanRunes)
				for stdout_scan.Scan() {
					r := stdout_scan.Text()
					os.Stdout.WriteString(r);
				}
			}
		}();

		go func() {
			if eerr == nil {
				stderr_scan := bufio.NewScanner(stderr)
				stderr_scan.Split(bufio.ScanRunes)
				for stderr_scan.Scan() {
					r := stderr_scan.Text()
					os.Stderr.WriteString(r);
				}
			}
		}();
	}

	proc, err := e.Process.Wait();
	if proc.ExitCode() != 0 {
		return fmt.Errorf("Process exited with non-0 exit code: %x", proc.ExitCode());
	}

	return nil;
}

func urlToPath(targetUrl string) string {
	nurl, _ := url.Parse(targetUrl);
	npath := strings.TrimPrefix(nurl.Path, "/");
	return npath;
}

func copyFile(sourcePath string, saveFilename string, onProgress func(done int64, total int64)) error {
	fmt.Printf("[Launcher] Copying %s to %s\n", sourcePath, saveFilename);
	os.MkdirAll(filepath.Dir(saveFilename), 0775);

	srcfd, err := os.OpenFile(sourcePath, os.O_RDONLY, 0777);
	if err != nil {
		return err;
	}
	defer srcfd.Close();

	dstfd, err := os.Create(saveFilename);
	if err != nil {
		return err;
	}
	defer dstfd.Close();

	stat, err := srcfd.Stat()
	if err != nil {
		return err;
	}

	total := stat.Size()
	done := int64(0);
	buffer := make([]byte, 0x8000);

	for done < total {
		rd, err := srcfd.Read(buffer);
		if err != nil {
			return err;
		}
		done += int64(rd);
		dstfd.Write(buffer[:rd]);
		onProgress(done, total);
	}

	return nil;
}

func downloadFile(sourceUrl string, saveFilename string, onProgress func(done int64, total int64)) error {
	fmt.Printf("[Launcher] Downloading %s\n", saveFilename);

	os.MkdirAll(filepath.Dir(saveFilename), 0775);
	req, err := createRequest("GET", sourceUrl, nil);
	resp, err := http.DefaultClient.Do(req);

	if err != nil {
		return err;
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("%s got non-200 status: %s", sourceUrl, resp.Status);
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

func checkFileExistsOnWebOrLocalStore(downloadUrl string) bool {
	urlParsed, err := url.Parse(downloadUrl);
	if err != nil {
		return false;
	}

	_, err = os.Stat(filepath.Join(LocalStoreFolder(), urlParsed.Path));
	if err != nil {
		req, err := createRequest("HEAD", downloadUrl, nil);
		resp, err := http.DefaultClient.Do(req);
		if err != nil {
			return false;
		}

		switch(resp.StatusCode) {
			case 200:
				return true;
			case 404:
				return false;
			case 403:
				return false;
			default:
				return false;
		}
	} else {
		return true;
	}
}


func downloadOrCopyFromLocalStore(targetUrl string, saveFilename string, onProgress func(done int64, total int64)) error{
	parsedUrl, err := url.Parse(targetUrl);
	if err != nil {
		return err;
	}

	localStorePath := filepath.Join(LocalStoreFolder(), parsedUrl.Path);

	_, err = os.Stat(localStorePath);
	if err != nil {
		return downloadFile(targetUrl, saveFilename, onProgress);
	} else {
		return copyFile(localStorePath, saveFilename, onProgress);
	}
}

func checkVerExist(startVersion int, endVersion int, architecture string, operatingSystem string, channel string) bool {
	uri := guessPatchUrlNoAuth(architecture, operatingSystem, channel, startVersion, endVersion);
	return checkFileExistsOnWebOrLocalStore(uri);
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
	fmt.Printf("[Launcher] Start version: %d\n", fromVersion);
	manifest, err := getVersionManifest(atokens, architecture, operatingSystem, channel, fromVersion);

	if(err != nil) {
		return err;
	}

	for _, step := range manifest.Steps {
		save := getVersionDownloadPath(step.From, step.To, channel);
		return downloadOrCopyFromLocalStore(step.Pwr, save, onProgress);
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

	err = downloadOrCopyFromLocalStore(downloadUrl, temp, onProgress);
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

	err := downloadOrCopyFromLocalStore(downloadUrl, temp, onProgress);
	defer os.Remove(temp);
	defer os.Remove(sigPath);
	defer os.RemoveAll(getVersionDownloadsFolder());

	if err != nil {
		return err;
	}

	err = downloadOrCopyFromLocalStore(downloadSig, sigPath, onProgress);
	defer os.Remove(sigPath);

	if err != nil {
		return err;
	}
	os.MkdirAll(out, 0775);

	fmt.Printf("[Launcher] Applying patch %s, using source: %s to: %s\n", temp, src, out);
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

func writeAurora(dllName string, embedName string) error {

	// write aurora dll
	data, err := embeddedFiles.ReadFile(embedName);
	if err != nil {
		return fmt.Errorf("failed to read embedded Aurora dll: %m -- Maybe try offline mode?", err);
	}

	fmt.Printf("[Aurora] Writing aurora dll: %s\n", dllName);
	os.WriteFile(dllName, data, 0777);

	// ld preload aurora dll
	switch(runtime.GOOS) {
		case "linux":
			fmt.Printf("[Aurora] LD_PRELOAD=%s\n", dllName);
			os.Setenv("LD_PRELOAD", dllName);
		case "darwin":
			fmt.Printf("[Aurora] DYLD_INSERT_LIBRARIES=%s\n", dllName);
			os.Setenv("DYLD_INSERT_LIBRARIES ", dllName);
	}

	return nil;
}

func findAurora(clientPath string) (dllName string, embedName string) {

	// get aurora dll location ..
	if runtime.GOOS == "windows" {
		dllName = filepath.Join(clientPath, "Secur32.dll");
		embedName = path.Join("Aurora", "Build", "Aurora.dll");
	}

	if runtime.GOOS == "linux" {
		dllName = filepath.Join(os.TempDir(), "Aurora.so");
		embedName = path.Join("Aurora", "Build", "Aurora.so");
	}

	fmt.Printf("[Aurora] dll name: %s\n", dllName);
	fmt.Printf("[Aurora] embed name: %s\n", embedName);

	return dllName, embedName;
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

	dllName, embedName := findAurora(filepath.Dir(clientBinary));

	// clear out aurora incase its there,
	os.Remove(dllName);

	// remove auora dll once were done ..
	defer os.Remove(dllName);

	if wCommune.Mode == E_MODE_FAKEONLINE { // start with fake online mode

		// setup fake online patch
		go runServer();
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


		fmt.Printf("[Launcher] Running: %s\n", strings.Join(e.Args, " "))

		writeAurora(dllName, embedName);

		// configure aurora for fakeonline
		e.Env = e.Environ();
		if wCommune.Console {
			e.Env = append(e.Env, "AURORA_ENABLE_CONSOLE=true");
		}

		e.Env = append(e.Env, "AURORA_ENABLE_INSECURE_SERVERS=true");
		e.Env = append(e.Env, "AURORA_ENABLE_AUTH_SWAP=true");
		e.Env = append(e.Env, "AURORA_ENABLE_SINGLEPLAYER_AS_INSECURE=true");
		e.Env = append(e.Env, "AURORA_SESSIONS="+getServerUrl()[:14]);
		e.Env = append(e.Env, "AURORA_ACCOUNT_DATA="+getServerUrl()[:14]);
		e.Env = append(e.Env, "AURORA_TOOLS="+getServerUrl()[:14]);
		e.Env = append(e.Env, "AURORA_TELEMETRY="+getServerUrl()[:14]);
		e.Env = append(e.Env, "AURORA_HYTALE_COM="+getServerUrl()[14:]);
		e.Env = append(e.Env, "AURORA_SENTRY_URL="+getSentryUrl());

		err := run(e);

		if err != nil {
			return err;
		}

	} else if wCommune.Mode == E_MODE_AUTHENTICATED { // start authenticated
		if wCommune.AuthTokens == nil {
			return errors.New("No auth token found.");
		}

		if wCommune.Profiles == nil {
			return errors.New("Could not find a profile, does this account own hytale?");
		}

		// get currently selected profile
		profileList := *wCommune.Profiles;
		profile := profileList[wCommune.SelectedProfile];

		newSess, err := getNewSession(*wCommune.AuthTokens, profile.UUID);
		if(err != nil) {
			return err;
		}

		// write aurora dll

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

		fmt.Printf("[Launcher] Running: %s\n", strings.Join(e.Args, " "))

		if wCommune.AuroraEverywhere {
			err := writeAurora(dllName, embedName);
			if err == nil {
				e.Env = e.Environ();
				if wCommune.Console {
					e.Env = append(e.Env, "AURORA_ENABLE_CONSOLE=true");
				}

				e.Env = append(e.Env, "AURORA_ENABLE_INSECURE_SERVERS=true");
				e.Env = append(e.Env, "AURORA_ENABLE_AUTH_SWAP=true");
				e.Env = append(e.Env, "AURORA_ENABLE_SINGLEPLAYER_AS_INSECURE=false");

				// disable sentry & telemetry
				e.Env = append(e.Env, "AURORA_TELEMETRY=http://127.0.0.1/");
				e.Env = append(e.Env, "AURORA_SENTRY_URL=http://a@127.0.0.1/2");
			}
		}

		err = run(e);

		if err != nil {
			return err;
		}

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

		fmt.Printf("[Launcher] Running: %s %s\n", clientBinary, strings.Join(e.Args, " "))

		err := run(e);

		if err != nil {
			return err;
		}
	}

	// clear ld-preload
	switch(runtime.GOOS) {
		case "linux":
			os.Unsetenv("LD_PRELOAD");
		case "darwin":
			os.Unsetenv("DYLD_INSERT_LIBRARIES ");
	}

	return nil;
}



func writeSettings() {
	jlauncher, _ := json.Marshal(wCommune);

	err := os.MkdirAll(filepath.Dir(LauncherJson()), 0666);
	if err != nil {
		fmt.Printf("[Launcher] error writing settings: %s\n", err);
		return;
	}

	err = os.WriteFile(LauncherJson(), jlauncher, 0666);
	if err != nil {
		fmt.Printf("[Launcher] error writing settings: %s\n", err);
		return;
	}

	fmt.Printf("[Launcher] Writing settings successfully.\n");
}


func readSettings() {
	_, err := os.Stat(LauncherJson())
	if err != nil {
		getDefaultSettings();
	} else {
		data, err := os.ReadFile(LauncherJson());
		if err != nil{
			getDefaultSettings();
			return;
		}
		json.Unmarshal(data, &wCommune);

		if wCommune.GameFolder != GameFolder() {
			wCommune.GameFolder = GameFolder();
		}
	}
	fmt.Printf("[Launcher] Reading settings successfully.\n")
}

func getDefaultSettings() {
	writeSettings();
	go checkForGameUpdates();

}

func cacheVersionList(channel string) {
	latest := wCommune.LatestVersions[channel];

	for i := range latest {
		wInstalledVersions[channel][i+1] = isGameVersionInstalled(i+1, channel)
	}
}

func findInstalledVersionList(channel string) {
	entries, err := os.ReadDir(getVersionsFolder(channel));
	if err != nil {
		return;
	}

	for _, e := range entries {
		if e.IsDir() {
			versionNumber, err := strconv.Atoi(filepath.Base(e.Name()));
			if err != nil {
				return;
			}

			if isGameVersionInstalled(versionNumber, channel) {
				wInstalledVersions[channel][versionNumber] = true;
				if versionNumber > wCommune.LatestVersions[channel] {
					wCommune.LatestVersions[channel] = versionNumber;
				}
			}
		}
	}
}

func findAllInstalledVersions() {
	findInstalledVersionList("release");
	findInstalledVersionList("pre-release");
}

func cacheAllVersions() {
	findAllInstalledVersions();
	cacheVersionList("release");
	cacheVersionList("pre-release");
}



func updateSelectedVerison() {

	prevLatest := 1;

	if wCommune.Patchline == E_PATCH_PRE_RELEASE {
		prevLatest = int(wCommune.LatestVersions["pre-release"])-1;
	} else {
		prevLatest = int(wCommune.LatestVersions["release"])-1;
	}

	if wCommune.SelectedVersion < 0 || int(wCommune.SelectedVersion) >= prevLatest {
		wCommune.SelectedVersion = int32(prevLatest);
	}
}

func authenticatedCheckForUpdatesAndGetProfileList() {
	if wCommune.AuthTokens == nil {
		return;
	}
	if(wCommune.Mode != E_MODE_AUTHENTICATED) {
		return;
	}

	lData, err := getVersionInformation(*wCommune.AuthTokens, runtime.GOARCH, runtime.GOOS);

	if err != nil {
		showErrorDialog(fmt.Sprintf("Failed to get launcher data: %s", err), "Auth failed.");
		wCommune.AuthTokens = nil;
		wCommune.Mode = E_MODE_FAKEONLINE;
		writeSettings();
	}

	lastReleaseVersion := wCommune.LatestVersions["release"];
	latestReleaseVersion := lData.Patchlines.Release.Newest;

	lastPreReleaseVersion := wCommune.LatestVersions["pre-release"];
	latestPreReleaseVersion := lData.Patchlines.PreRelease.Newest;

	if latestReleaseVersion > lastReleaseVersion {
		fmt.Printf("[Launcher] [AUTH] found new release: %d\n", lastReleaseVersion)
		wCommune.LatestVersions["release"] = latestReleaseVersion;
	}
	if latestPreReleaseVersion > lastPreReleaseVersion {
		fmt.Printf("[Launcher] [AUTH] found new release: %d\n", lastPreReleaseVersion)
		wCommune.LatestVersions["pre-release"] = latestPreReleaseVersion;
	}

	wCommune.Profiles = &lData.Profiles;
	updateSelectedVerison();

	writeSettings();
	cacheAllVersions();

}


func checkForGameUpdates() {
	if !wGotLatestLauncherVersion {
		wCommune.SpoofLauncherVersion = getLatestLauncherVersion();
	}

	if wCommune.Mode != E_MODE_AUTHENTICATED {
		lastRelease := wCommune.LatestVersions["release"]
		lastPreRelease := wCommune.LatestVersions["pre-release"]

		latestRelease := findLatestVersionNoAuth(lastRelease, runtime.GOARCH, runtime.GOOS, "release");
		latestPreRelease := findLatestVersionNoAuth(lastPreRelease, runtime.GOARCH, runtime.GOOS, "pre-release");

		if latestRelease > lastRelease {
			fmt.Printf("[Launcher] Found new release version: %d\n", latestRelease);
			wCommune.LatestVersions["release"] = latestRelease;
		}

		if latestPreRelease > lastPreRelease {
			fmt.Printf("[Launcher] Found new pre-release version: %d\n", latestPreRelease);
			wCommune.LatestVersions["pre-release"] = latestPreRelease;
		}

		updateSelectedVerison();

	} else {
		authenticatedCheckForUpdatesAndGetProfileList();
	}

	writeSettings();
	cacheAllVersions();

}

