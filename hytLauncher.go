package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"bitbucket.org/rj/goey/loop"
	"github.com/c4milo/unpackit"
)


func urlToPath(targetUrl string) string {
	nurl, _ := url.Parse(targetUrl);
	npath := strings.TrimPrefix(nurl.Path, "/");
	return npath;
}

func download(targetUrl string, saveFilename string, progress func(done int64, total int64)) any {
	fmt.Printf("Downloading %s\n", targetUrl);

	os.MkdirAll(filepath.Dir(saveFilename), 0775);
	resp, err := http.Get(targetUrl);
	if err != nil {
		return nil;
	}

	if resp.StatusCode == 200 {
		f, _ := os.Create(saveFilename);

		defer f.Close();

		total := resp.ContentLength;
		done := int64(0);
		buffer := make([]byte, 0x8000);

		for done < total {
			rd, _ := resp.Body.Read(buffer);
			done += int64(rd);
			f.Write(buffer[:rd]);
			progress(done, total);
		}
	}

	return saveFilename;
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


func downloadLatestVersion(atokens accessTokens, architecture string, operatingSystem string, channel string, fromVersion int, progress func(done int64, total int64)) any {
	fmt.Printf("Start version: %d\n", fromVersion);
	manifest := getVersionManifest(atokens, architecture, operatingSystem, channel, fromVersion);
	for _, step := range manifest.Steps {
		save := getVersionDownloadPath(step.From, step.To, channel);
		return download(step.Pwr, save, progress);
	}
	return nil;
}


func installJre(progress func(done int64, total int64)) any {
	jres, ok := getJres("release").(versionFeed);
	if ok {

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

		_, err := os.Stat(unpack);

		if err != nil {
			_, ok := download(downloadUrl, save, progress).(string);
			if ok {
				os.MkdirAll(unpack, 0775);

				f, err := os.Open(save);
				if err != nil {
					panic("failed to open jre download");
				}

				err = unpackit.Unpack(f, unpack);

				if(err != nil) {
					panic("failed to unpack jre");
				}

				os.Remove(save);
				os.RemoveAll(filepath.Dir(save));
				return unpack;
			} else {
				panic("Failed to download jre");
			}
		}

	}

	return nil;
}

func isGameVersionInstalled(version int, channel string) bool {
	gameDir := findClientBinary(version, channel);
	_, err := os.Stat(gameDir);
	if err != nil {
		return false;
	}
	return true;
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

func installGame(version int, channel string, progress func(done int64, total int64)) any {
	save := getVersionDownloadPath(0, version, channel);
	unpack := getVersionInstallPath(version, channel);

	closestVersion := findClosestVersion(version, channel);
	srcPath := getVersionInstallPath(closestVersion, channel);

	fmt.Printf("Closest version: %d\n", closestVersion);
	fmt.Printf("Src Path: %s\n", srcPath);


	if !isGameVersionInstalled(version, channel) {
		downloadUrl := guessPatchUrlNoAuth(runtime.GOARCH, runtime.GOOS, channel, closestVersion, version);

		// check if this patch exists, if not fallback on the 0 patch.
		if !checkVerExist(closestVersion, version, runtime.GOARCH, runtime.GOOS, channel) {
			downloadUrl = guessPatchUrlNoAuth(runtime.GOARCH, runtime.GOOS, channel, 0, version);
		}

		pwr := download(downloadUrl, save, progress);
		os.MkdirAll(unpack, 0775);

		if _, ok := pwr.(string); ok {
			applyPatch(srcPath, unpack, save);

			os.Remove(save);
			os.RemoveAll(getVersionDownloadsFolder());
			return unpack;
		} else {
			panic("Failed to download version");
		}
	}
	return nil;
}

func findJavaBin() any {
	jrePath := getJrePath(runtime.GOOS, runtime.GOARCH);

	d, err := os.ReadDir(jrePath);
	if err != nil {
		panic(err);
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
			fallthrough; // TODO: confirm this ..
		case "linux":
			return filepath.Join(clientFolder, "HytaleClient");
		default:
			panic("Hytale is not supported by your OS.");
	}
}

func launchGame(version int, channel string, username string, uuid string) {

	javaBin, _ := findJavaBin().(string);

	appDir := getVersionInstallPath(version, channel)
	userDir := UserDataFolder()
	clientBinary := findClientBinary(version, channel);

	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		os.Chmod(javaBin, 0775);
		os.Chmod(clientBinary, 0775);
	}

	if wCommune.Mode == "fakeonline" {

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
			panic("failed to read aurora dll");
		}
		os.WriteFile(dllName, data, 0777);
		defer os.Remove(dllName);

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
			generateIdentityJwt("hytale:client"),
			"--session-token",
			generateSessionJwt("hytale:client"));

		fmt.Printf("LD_PRELOAD=%s\n", dllName);

		if runtime.GOOS == "linux" {
			os.Setenv("LD_PRELOAD", dllName);
		}

		fmt.Printf("Running: %s\n", strings.Join(e.Args, " "))

		err = e.Start();
		if err != nil {
			wMainWin.Message(fmt.Sprintf("Failed to start %s", err)).WithError().WithTitle("Failed to start").Show();
		}

		if runtime.GOOS == "linux" {
			os.Unsetenv("LD_PRELOAD");
		}

		defer e.Process.Kill();
		e.Process.Wait();

	} else { // start in offline mode

		// remove fakeonline patch if present.
		dllName := filepath.Join(filepath.Dir(clientBinary), "Secur32.dll");
		os.Remove(dllName);

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
			loop.Do(func() error{
				wMainWin.Message(fmt.Sprintf("Failed to start %s", err)).WithError().WithTitle("Failed to start").Show();
				return nil;
			})
		}


		defer e.Process.Kill();
		e.Process.Wait();
	}
}
