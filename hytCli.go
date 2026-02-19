package main

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"sort"
	"strconv"
	"strings"
)

type commandLineArgument struct {
	arg_long string
	arg string
	help string
	cmd_func func(arg string) error
	priority int
}
var commands = []commandLineArgument{};

var(
	cOs string
	cArch string
)

func loadDefaults() {
	readSettings();

	// select latest version by default ?
	wCommune.SelectedVersion = int32(wCommune.LatestVersions[valToChannel(int(wCommune.Patchline))]);

	cOs = runtime.GOOS
	cArch = runtime.GOARCH
}
func cliProgressUpdate(done int64, total int64) {

	progress := int(math.Floor((float64(done) / float64(total) * 100.0)));

	pBar := "["
	pBarM := 100 / 2;
	pBarV := progress / 2;

	for range pBarV {
		pBar += "=";
	}

	for range (pBarM - pBarV) {
		pBar += " ";
	}

	pBar += "]";


	fmt.Printf("Progress %s %d%% \r", pBar, progress);

}
func unimplementedCommand(arg string) error {
	fmt.Printf("Unimplemented command!\n");
	return nil;
}

func handleDownloadClient(arg string) error {

	// get channel as a string ..
	channelStr := valToChannel(int(wCommune.Patchline));
	version := int(wCommune.SelectedVersion)+1;

	// find the closest version
	closest := 0;
	src := getVersionInstallPath(closest, channelStr);
	apply := getVersionInstallPath(version, channelStr);
	temp := getVersionDownloadPath(closest, version, channelStr);

	// if version matches the actual os version,
	// then look for any previously installed version to do an incremental patch from ..
	if cOs == runtime.GOOS && cArch == runtime.GOARCH {
		closest = findClosestVersion(version, channelStr);
		if checkVerExist(closest, version, cArch, cOs, channelStr) {
			src = getVersionInstallPath(closest, channelStr);
			apply = getVersionInstallPath(version, channelStr);
			temp = getVersionDownloadPath(closest, version, channelStr);
		}
	}

	if arg != "" {
		apply = arg;
		temp = filepath.Join(filepath.Dir(apply), "patch.pwr");
	}


	err := installGameEx(closest, version, channelStr, cOs, cArch, src, apply, temp, cliProgressUpdate);
	if err != nil{
		return err;
	}

	return nil;
}

func handleDownloadJre(arg string) error {
	out := getJrePath(cOs, cArch);
	temp := getJreDownloadPath(cOs);

	if arg != "" {
		out = arg;
		temp = filepath.Join(filepath.Dir(temp), "jre-download."+filepath.Ext(temp));
	}

	return installJreEx(cOs, out, temp, cliProgressUpdate);
}

func handleDownloadServer(arg string) error {

	if arg == "" {
		return fmt.Errorf("No output directory specified, call with --download-server=\"./path/to/directory\"");
	}

	err := handleDownloadClient(arg);
	if err != nil {
		return err;
	}


	err = os.RemoveAll(filepath.Join(arg, "Client"));
	if err != nil {
		return err;
	}

	return nil;

}

func handleOperatingSystem(arg string) error {
	switch(arg) {
		case "macos":
		case "mac":
		case "osx":
		case "darwin":
			cOs = "darwin";
			cArch = "arm64";
			return nil;
		case "linux":
			cOs = "linux";
			cArch = "amd64";
			return nil;
		case "windows":
		case "win32":
			cOs = "windows";
			cArch = "amd64"
			return nil;
		default:
			break;
	}
	return fmt.Errorf("Expected one of 3 options: \"windows\", \"linux\", or \"darwin\", got: "+arg);
}

func handleLaunchMode(arg string) error {
	switch(arg) {
		case "offline":
			wCommune.Mode = E_MODE_OFFLINE;
			return nil;
		case "onlinefix":
			fallthrough;
		case "cracked":
			fallthrough;
		case "onlinepatch":
			fallthrough;
		case "fakeonline":
			wCommune.Mode = E_MODE_FAKEONLINE;
			return nil;
		case "online":
			fallthrough;
		case "offical":
			fallthrough;
		case "authenticated":
			wCommune.Mode = E_MODE_AUTHENTICATED;
			return nil;
		default:
			return fmt.Errorf("Expected one of 3 options: \"offline\", \"fakeonline\" or \"authenticated\" got: "+arg);

	}

}

func handleArchitecture(arg string) error {
	arch := runtime.GOARCH;

	switch(arg) {
		case "intel":
			fallthrough;
		case "amd":
			fallthrough;
		case "x86":
			fallthrough;
		case "x86_64":
			fallthrough;
		case "amd64":
			arch = "amd64";
			break;
		case "arm":
			fallthrough;
		case "aarch64":
			fallthrough;
		case "arm64":
			arch = "arm64";
			break;
		default:
			return fmt.Errorf("Expected one of 2 options: \"amd64\" or \"arm64\", got: "+arg);
	}

	// check for arm linux or arm windows ...
	if arch == "arm64" && (cOs == "linux" || cOs == "windows") {
		return fmt.Errorf("Unsupported operating system for architecture %s: %s (expected: %s)", arch, cOs, "darwin");
	}

	// check for amd64 mac
	if arch == "amd64" && (cOs == "darwin") {
		return fmt.Errorf("Unsupported operating system for architecture %s: %s (expected: %s)", arch, cOs, "linux or windows");
	}

	cArch = arch;
	return nil;
}

func handleGameDirectory(arg string) error {
	if arg == "" {
		return fmt.Errorf("No game directory specified.");
	}

	wCommune.GameFolder = arg;
	return nil;
}

func handleJreDirectory(arg string) error {
	if arg == "" {
		return fmt.Errorf("No jre directory specified.");
	}

	wCommune.JreFolder = arg;
	return nil;
}

func handleUserdataDirectory(arg string) error {
	if arg == "" {
		return fmt.Errorf("No userdata directory specified.");
	}

	wCommune.UserDataFolder = arg;
	return nil;
}


func handleUsername(arg string) error {
	if wCommune.Mode == E_MODE_AUTHENTICATED {
		return fmt.Errorf("Cannot specify custom username when using authenticated mode.");
	}
	if arg == "" {
		return fmt.Errorf("No username specified.");
	}

	wCommune.Username = arg;
	return nil;
}

func handleUuid(arg string) error {
	if wCommune.Mode == E_MODE_AUTHENTICATED {
		return fmt.Errorf("Cannot specify custom uuid when using authenticated mode.");
	}

	if arg == "" {
		return fmt.Errorf("No UUID specified.");
	}

	wCommune.UUID = arg;
	return nil;
}

func handlePatchline(arg string) error {
	switch(arg) {
		case "release":
			wCommune.Patchline = E_PATCH_RELEASE;
			wCommune.SelectedVersion = int32(wCommune.LatestVersions["release"]-1);
			return nil;
		case "prerelease":
			fallthrough;
		case "pre-release":
			wCommune.Patchline = E_PATCH_PRE_RELEASE;
			wCommune.SelectedVersion = int32(wCommune.LatestVersions["pre-release"]-1);
			return nil;
		default:
			break;
	}
	return fmt.Errorf("Expected one of 2 options: \"release\", \"pre-release\", got: "+arg);
}

func handleRunClient(arg string) error {
	version := int(wCommune.SelectedVersion)+1;
	channel := valToChannel(int(wCommune.Patchline));

	downloadAndRunGame(version, channel, cliProgressUpdate);

	return nil;
}

func handleAuthetnicate(arg string) error {
	aTokens, err := getAuthTokens(wCommune.AuthTokens);

	if err != nil {
		return err;
	}

	wCommune.AuthTokens = &aTokens;

	authenticatedCheckForUpdatesAndGetProfileList();
	return nil;
}

func handleListVersion(arg string) error {
	fmt.Printf("-- Version List: --\n");

	fmt.Printf("pre-release: \n");
	for i := range wCommune.LatestVersions["pre-release"] {
		fmt.Printf("\tPre Release: %d\n", i+1);
	}
	fmt.Printf("release: \n");
	for i := range wCommune.LatestVersions["release"] {
		fmt.Printf("\tRelease: %d\n", i+1);
	}

	return nil;
}

func handleVersion(arg string) error {
	channelStr := valToChannel(int(wCommune.Patchline));

	if arg == "latest" {
		wCommune.SelectedVersion = int32(wCommune.LatestVersions[channelStr]-1);
		return nil;
	}

	ver, err := strconv.Atoi(arg);

	if err != nil {
		return err;
	}

	latest := wCommune.LatestVersions[channelStr];

	if ver > latest {
		return fmt.Errorf("Selected version %d is larger than the latest version %s %d\n", ver, channelStr, latest);
	}

	if ver <= 0 {
		return fmt.Errorf("Selected version %d is smaller than the first version %s %d\n", ver, channelStr, 1);
	}

	wCommune.SelectedVersion = int32(ver-1);

	return nil;

}

func handleHelp(arg string) error {

	for _, command := range commands {
		fmt.Printf("\t%-20s %-10s %5s\n", "--"+command.arg_long, "-"+command.arg, command.help);
	}
	fmt.Println();
	fmt.Println();
	fmt.Println("Example 1: HytaleSP --run-client --launch-mode=fakeonline --version=latest --patchline=release --username=\"Hypixel\" ")
	fmt.Println("Example 2: HytaleSP --operating-system=linux --architecture=amd64 --version=latest --download-client=\"./game-client\"")
	return nil;
}

func getPrioList() []int {
	// get a list of all command priorities ...
	prioList := []int{};
	for _, command := range commands {
		if !slices.Contains(prioList, command.priority) {
			prioList = append(prioList, command.priority)
		}
	}

	sort.Slice(prioList, func(i, j int) bool {
		return prioList[i] > prioList[j]
	})

	return prioList;
}

func handleCli(argv []string) int{

	commands = []commandLineArgument {
		{"help", 		"h", 	"Shows this help message", 								handleHelp, 1000},
		{"operating-system", 	"os", 	"Specifies what operating system your targeting. (eg: windows, darwin, linux)",		handleOperatingSystem, 100},
		{"patchline", 		"p", 	"Specifies the patchline to run or download from. (eg: release, pre-release)", 		handlePatchline, 100},
		{"launch-mode", 	"m", 	"Specifies the launch mode to use the game, (eg: offline, fakeonline, authenticated)", 	handleLaunchMode, 100},
		{"version", 		"v", 	"Specifies the version to use while downloading or running the game (eg: latest, 8,)", 	handleVersion, 50},
		{"architecture", 	"arch", "Specifies what processor architecture your targeting. (eg: amd64, arm64)",		handleArchitecture, 50},
		{"game-directory", 	"gd", 	"Specifies the game directory",								handleGameDirectory, 20},
		{"jre-directory", 	"jd", 	"Specifies the jre directory",								handleJreDirectory, 20},
		{"userdata-directory", 	"ud", 	"Specifies the userdata directory",							handleUserdataDirectory, 20},
		{"username", 		"u", 	"Specifies the username to run the client with",					handleUsername, 20},
		{"uuid", 		"uu", 	"Specifies the universal unique identifier to run the client with",			handleUuid, 20},
		{"authenticate", 	"a", 	"Authenticate to hytale.com",		 						handleAuthetnicate, 10},
		{"list-version", 	"lv", 	"Lists all the versions for the specified patchline.", 					handleListVersion, 5},
		{"download-client", 	"dc", 	"Downloads a client patch and applies it to the specified directory", 			handleDownloadClient, 5},
		{"download-jre", 	"dj", 	"Downloads and extracts the JRE and puts it into the specified directory",		handleDownloadJre, 5},
		{"download-server", 	"ds", 	"Downloads or extracts the server and puts it into the specified directory",		handleDownloadServer, 5},
		{"run-client", 		"rc", 	"Starts a given version of hytale, or downloads it if its not found.", 			handleRunClient, 1},
	}


	os.MkdirAll(MainFolder(), 0775);
	os.MkdirAll(LauncherFolder(), 0775);
	os.MkdirAll(ServerDataFolder(), 0775);


	loadDefaults();
	refreshAuthentication();
	checkForGameUpdates();

	os.MkdirAll(UserDataFolder(), 0775);
	os.MkdirAll(JreFolder(), 0775);
	os.MkdirAll(GameFolder(), 0775);

	res := 1;

	for _, currentPriority := range getPrioList() {
		for _, arg := range argv {
			normArg := strings.Trim(arg, "-");
			subArgs := strings.Split(normArg, "=")

			for _, command := range commands {
				if command.priority != currentPriority {
					continue;
				}

				if subArgs[0] == command.arg || subArgs[0] == command.arg_long {
					err := command.cmd_func(strings.Join(subArgs[1:], " "));
					if err != nil {
						fmt.Printf("Error: %s\n", err);
						return -1;
					}
					res = 0;
				}
			}
		}

	}

	writeSettings();

	return res;
}
