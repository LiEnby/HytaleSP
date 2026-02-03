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

func download(targetUrl string, saveFilename string, progress func(done int64, total int64)) error {
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
		progress(done, total);
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

func getJreDownloadPath(operatingSystem string, architecture string, downloadUrl string) string {
	u, _ := url.Parse(downloadUrl);
	fp := filepath.Join(JreFolder(), "download", operatingSystem, architecture, path.Base(u.Path));
	return fp;
}


func downloadLatestVersion(atokens accessTokens, architecture string, operatingSystem string, channel string, fromVersion int, progress func(done int64, total int64)) error {
	fmt.Printf("Start version: %d\n", fromVersion);
	manifest, err := getVersionManifest(atokens, architecture, operatingSystem, channel, fromVersion);

	if(err != nil) {
		return err;
	}

	for _, step := range manifest.Steps {
		save := getVersionDownloadPath(step.From, step.To, channel);
		return download(step.Pwr, save, progress);
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

func installJre(progress func(done int64, total int64)) error{

	if isJreInstalled() {
		return nil;
	}

	jres, err := getJres("release");
	if err != nil {
		return err;
	}


	var downloadUrl string;

	switch(runtime.GOOS) {
		case "windows":
			downloadUrl = jres.DownloadUrls.Windows.Amd64.URL;
		case "linux":
			downloadUrl = jres.DownloadUrls.Linux.Amd64.URL;
		case "darwin":
			downloadUrl = jres.DownloadUrls.Darwin.Amd64.URL;

	}

	save := getJreDownloadPath(runtime.GOOS, runtime.GOARCH, downloadUrl);
	unpack := getJrePath(runtime.GOOS, runtime.GOARCH);

	err = download(downloadUrl, save, progress);
	defer os.Remove(save);

	if err != nil {
		return err;
	}

	valid := false;

	// validate jre
	switch(runtime.GOOS) {
		case "windows":
			valid = verifyFileSha256(save, jres.DownloadUrls.Windows.Amd64.Sha256);
		case "linux":
			valid = verifyFileSha256(save, jres.DownloadUrls.Linux.Amd64.Sha256);
		case "darwin":
			valid = verifyFileSha256(save, jres.DownloadUrls.Darwin.Arm64.Sha256);
	}

	if valid == false {
		return fmt.Errorf("Could not validate the SHA256 hash for the JRE runtime.");
	}

	os.MkdirAll(unpack, 0775);

	f, err := os.Open(save);
	if err != nil {
		os.RemoveAll(unpack);
		return err;
	}

	err = unpackit.Unpack(f, unpack);

	if(err != nil) {
		os.RemoveAll(unpack);
		return err;
	}

	os.Remove(save);
	os.RemoveAll(filepath.Dir(save));
	return nil;

}

func findClosestVersion(targetVersion int, channel string) int {
	installFolder := getVersionsFolder(channel);

	fVersion := 0;

	d, err := os.ReadDir(installFolder);
	if err != nil {
		return fVersion;
	}

	for _, e := range d {
		if !e.IsDir() {
			continue;
		}

		ver, err := strconv.Atoi(e.Name());

		if err != nil {
			continue;
		}

		if ver > fVersion && ver < targetVersion {
			fVersion = ver;
		}
	}

	return fVersion;

}

func installGame(version int, channel string, progress func(done int64, total int64)) error {


	if !isGameVersionInstalled(version, channel) {
		closestVersion := findClosestVersion(version, channel);
		srcPath := getVersionInstallPath(closestVersion, channel);

		fmt.Printf("Closest version: %d\n", closestVersion);
		fmt.Printf("Src Path: %s\n", srcPath);

		downloadUrl := guessPatchUrlNoAuth(runtime.GOARCH, runtime.GOOS, channel, closestVersion, version);
		downloadSig := guessPatchSigUrlNoAuth(runtime.GOARCH, runtime.GOOS, channel, closestVersion, version);

		unpack := getVersionInstallPath(version, channel);
		save := getVersionDownloadPath(closestVersion, version, channel);

		// check if this patch exists, if not fallback on the 0 patch.
		if !checkVerExist(closestVersion, version, runtime.GOARCH, runtime.GOOS, channel) {
			downloadUrl = guessPatchUrlNoAuth(runtime.GOARCH, runtime.GOOS, channel, 0, version);
			downloadSig = guessPatchSigUrlNoAuth(runtime.GOARCH, runtime.GOOS, channel, 0, version);
			save = getVersionDownloadPath(0, version, channel);
		}

		saveSig := save + ".pwr";

		err := download(downloadUrl, save, progress);
		defer os.Remove(save);
		defer os.RemoveAll(getVersionDownloadsFolder());

		if err != nil {
			return err;
		}

		err = download(downloadSig, saveSig, progress);
		defer os.Remove(saveSig);

		if err != nil {
			return err;
		}
		os.MkdirAll(unpack, 0775);

		err = applyPatch(srcPath, unpack, save, saveSig);
		if err != nil {
			err := os.RemoveAll(unpack);
			return err;
		}

		return nil;
	}
	return nil;
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
