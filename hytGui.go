package main

import (
	"encoding/json"
	"fmt"
	"image"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/AllenDang/giu"
	"github.com/sqweek/dialog"
)

type launcherCommune struct {
	Patchline int32 `json:"last_patchline"`
	Username string `json:"last_username"`
	SelectedVersion int32 `json:"last_version"`
	LatestVersions map[string]int `json:"last_version_scan_result"`
	Mode int32 `json:"mode"`

	// authentication
	AuthTokens *accessTokens `json:"token"`
	Profiles *[]accountInfo `json:"profiles"`
	SelectedProfile int32 `json:"selected_profile"`

	// settings
	GameFolder string `json:"install_directory"`
	UserDataFolder string `json:"userdata_directory"`
	JreFolder string `json:"jre_directory"`
	UUID string `json:"uuid_override"`
}


const DEFAULT_USERNAME = "TransRights";
const DEFAULT_PATCHLINE = E_PATCH_RELEASE;


const E_MODE_OFFLINE = 0;
const E_MODE_FAKEONLINE = 1;
const E_MODE_AUTHENTICATED = 2;

const E_PATCH_RELEASE = 0
const E_PATCH_PRE_RELEASE = 1

var (
	wMainWin *giu.MasterWindow
	wCommune = launcherCommune {
		Patchline: DEFAULT_PATCHLINE,
		Username: DEFAULT_USERNAME,
		LatestVersions: map[string]int{
			"release": 7,
			"pre-release": 17,
		},
		SelectedVersion: 4,
		Mode: E_MODE_FAKEONLINE,
		AuthTokens: nil,
		Profiles: nil,
		SelectedProfile: 0,

		GameFolder: DefaultGameFolder(),
		UserDataFolder: DefaultUserDataFolder(),
		JreFolder: DefaultJreFolder(),
		UUID: "",
	};
	wProgress float32 = 0.0
	wDisabled = false
	wSelectedTab = 0
	wGameRunning = false
	wInstalledVersions map[string]map[int]bool = map[string]map[int]bool{
		"release" : map[int]bool{},
		"pre-release" : map[int]bool{},
	}
	wImGuiWindow *giu.WindowWidget = nil;
)

func cacheVersionList() {

	channel := "release"
	latest := wCommune.LatestVersions[channel];

	for i := range latest {
		wInstalledVersions[channel][i+1] = isGameVersionInstalled(i+1, channel)
	}

	channel = "pre-release"
	latest = wCommune.LatestVersions[channel];

	for i := range latest {
		wInstalledVersions[channel][i+1] = isGameVersionInstalled(i+1, channel)
	}
}

func getWindowWidth() float32 {
	vec2 := imgui.ContentRegionAvail();
	return vec2.X; //float32(w) - float32(padX*2);
}


func doAuthentication() {
	aTokens, err := getAuthTokens(wCommune.AuthTokens);

	if err != nil {
		showErrorDialog(fmt.Sprintf("Failed to get auth tokens: %s", err), "Auth failed.");
		wCommune.AuthTokens = nil;
		wCommune.Mode = E_MODE_FAKEONLINE;
		writeSettings();
		//loop.Do(updateWindow);
	}

	wCommune.AuthTokens = &aTokens;

	// get profile list ..
	authenticatedCheckForUpdatesAndGetProfileList();

}


func checkForUpdates() {
	if wCommune.Mode != E_MODE_AUTHENTICATED {
		lastRelease := wCommune.LatestVersions["release"]
		lastPreRelease := wCommune.LatestVersions["pre-release"]

		latestRelease := findLatestVersionNoAuth(lastRelease, runtime.GOARCH, runtime.GOOS, "release");
		latestPreRelease := findLatestVersionNoAuth(lastPreRelease, runtime.GOARCH, runtime.GOOS, "pre-release");

		fmt.Printf("latestRelease: %d\n", latestRelease);
		fmt.Printf("latestPreRelease: %d\n", latestPreRelease);

		if latestRelease > lastRelease {
			fmt.Printf("Found new release version: %d\n", latestRelease);
			wCommune.LatestVersions["release"] = latestRelease;
		}

		if latestPreRelease > lastPreRelease {
			fmt.Printf("Found new pre-release version: %d\n", latestPreRelease);
			wCommune.LatestVersions["pre-release"] = latestPreRelease;
		}

		writeSettings();
		cacheVersionList();
	}
}

func authenticatedCheckForUpdatesAndGetProfileList() {
	if wCommune.AuthTokens == nil {
		return;
	}
	if(wCommune.Mode != E_MODE_AUTHENTICATED) {
		return;
	}

	lData, err := getLauncherData(*wCommune.AuthTokens, runtime.GOARCH, runtime.GOOS);

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
		fmt.Printf("found new release: %d\n", lastReleaseVersion)
		wCommune.LatestVersions["release"] = latestReleaseVersion;
	}
	if latestPreReleaseVersion > lastPreReleaseVersion {
		fmt.Printf("found new release: %d\n", lastPreReleaseVersion)
		wCommune.LatestVersions["pre-release"] = latestPreReleaseVersion;
	}

	wCommune.Profiles = &lData.Profiles;

	writeSettings();
	cacheVersionList();
}

func reAuthenticate() {
	if wCommune.AuthTokens != nil && wCommune.Mode == E_MODE_AUTHENTICATED {
		aTokens, err:= getAuthTokens(*wCommune.AuthTokens);

		if err != nil {
			showErrorDialog(fmt.Sprintf("Failed to authenticate: %s", err), "Auth failed.");
			wCommune.AuthTokens = nil;
			wCommune.Mode = E_MODE_FAKEONLINE;
			writeSettings();
		}

		wCommune.AuthTokens = &aTokens;
		authenticatedCheckForUpdatesAndGetProfileList();
	}
}

func writeSettings() {
	fmt.Printf("Saving settings ...\n");
	jlauncher, _ := json.Marshal(wCommune);

	err := os.MkdirAll(filepath.Dir(getLauncherJson()), 0666);
	if err != nil {
		fmt.Printf("error writing settings: %s\n", err);
		return;
	}

	err = os.WriteFile(getLauncherJson(), jlauncher, 0666);
	if err != nil {
		fmt.Printf("error writing settings: %s\n", err);
		return;
	}
}

func getDefaultSettings() {
	writeSettings();
	go checkForUpdates();

}

func getLauncherJson() string {
	return filepath.Join(LauncherFolder(), "launcher.json");
}

func readSettings() {
	_, err := os.Stat(getLauncherJson())
	if err != nil {
		getDefaultSettings();
	} else {
		data, err := os.ReadFile(getLauncherJson());
		if err != nil{
			getDefaultSettings();
			return;
		}
		json.Unmarshal(data, &wCommune);

		if wCommune.GameFolder != GameFolder() {
			wCommune.GameFolder = GameFolder();
		}

		fmt.Printf("Reading last settings: \n");
		fmt.Printf("username: %s\n", wCommune.Username);
		fmt.Printf("patchline: %s\n", valToChannel(int(wCommune.Patchline)));
		fmt.Printf("last used version: %d\n", wCommune.SelectedVersion+1);
		fmt.Printf("newest known release: %d\n", wCommune.LatestVersions["release"])
		fmt.Printf("newest known pre-release: %d\n", wCommune.LatestVersions["pre-release"])

	}
}


func valToChannel(vchl int) string {
	switch vchl {
		case E_PATCH_RELEASE:
			return "release";
		case E_PATCH_PRE_RELEASE:
			return "pre-release";
		default:
			return "release";
	}
}

func channelToVal(channel string) int {
	switch channel {
		case "release":
			return E_PATCH_RELEASE;
		case "pre-release":
			return E_PATCH_PRE_RELEASE;
		default:
			return DEFAULT_PATCHLINE;
	}
}

func startGame() {
	// disable the current window
	wDisabled = true;

	// enable the window again once done
	defer func() {
		wDisabled = false;
		wGameRunning = false;
	}();

	ver := int(wCommune.SelectedVersion+1);
	channel := valToChannel(int(wCommune.Patchline));

	err := installJre(updateProgress);

	if err != nil {
		showErrorDialog(fmt.Sprintf("Error getting the JRE: %s", err), "Install JRE failed.");
		return;
	};

	if !wInstalledVersions[channel][ver] {
		err = installGame(ver, valToChannel(int(wCommune.Patchline)), updateProgress);

		if err != nil {
			showErrorDialog(fmt.Sprintf("Error getting the game: %s", err), "Install game failed.");
			return;
		};

		wInstalledVersions[channel][ver] = true;
	}

	// set game running flag
	wGameRunning = true;
	err = launchGame(ver, channel, wCommune.Username, getUUID());

	if err != nil {
		showErrorDialog(fmt.Sprintf("Error running the game: %s", err), "Run game failed.");
		return;
	};
}

func patchLineMenu() giu.Widget {
	return giu.Layout{
		giu.Label("Patchline: "),
		giu.Row(
			giu.Combo("##patchline", valToChannel(int(wCommune.Patchline)), []string{"release", "pre-release"}, &wCommune.Patchline).OnChange(func() {
				wCommune.SelectedVersion = int32(wCommune.LatestVersions[valToChannel(int(wCommune.Patchline))]-1);
			}).Size(getWindowWidth()),
		),
	}
}


func versionMenu() giu.Widget {
	versions := []string {};
	selectedChannel := valToChannel(int(wCommune.Patchline));
	selectedVersion := int(wCommune.SelectedVersion+1);

	latest := wCommune.LatestVersions[selectedChannel];
	for i := range latest {
		txt := "Version "+strconv.Itoa(i+1);
		if wInstalledVersions[selectedChannel][i+1] {
			txt += " - installed";
		} else {
			txt += " - not installed";
		}
		versions = append(versions, txt);
	}
	buttondisabled := !wInstalledVersions[selectedChannel][selectedVersion] || wDisabled;

	bSize := getButtonSize("Delete")
	padX, _ := giu.GetWindowPadding();


	return giu.Layout{
		giu.Label("Version: "),
		giu.Row(
			giu.Combo("##version", versions[int(wCommune.SelectedVersion) % len(versions)], versions, &wCommune.SelectedVersion).Size(getWindowWidth() - (bSize + padX)),
			giu.Button("Delete").Disabled(buttondisabled).OnClick(func() {
				wDisabled = true;

				go func() {
					defer func() { wDisabled = false; }();

					installDir := getVersionInstallPath(selectedVersion, selectedChannel);
					err := os.RemoveAll(installDir);
					if err != nil {
						showErrorDialog(fmt.Sprintf("failed to remove: %s", err), "failed to remove");
						return;
					}

					wInstalledVersions[selectedChannel][selectedVersion] = false;
				}();
			}),
		),
	};
}


func labeledTextInput(label string, value *string, disabled bool) giu.Widget {
	if value == nil {
		panic("failed to initalize browse button");
	}

	return giu.Style().SetDisabled(wDisabled || disabled).To(
		giu.Label(label+": "),
		giu.Row(
			giu.InputText(value).Hint(label).Label("##"+label).Size(getWindowWidth()),
		),
	);

}

func browseButton(label string, value *string, callback func()) giu.Widget {
	if value == nil {
		panic("failed to initalize browse button");
	}

	button := giu.Button("Browse").OnClick(func() {
		dir, err := dialog.Directory().Title("Select "+label).Browse();
		if err != nil {
			if err != dialog.ErrCancelled {
				showErrorDialog(fmt.Sprintf("Failed: %s", err), "Error reading directory");
			}
		}
		*value = dir;
		if callback != nil { callback(); }
	})

	// surely there has got to be a better way to do this ..?
	// it literally tells me not to use this function lol
	bSize := getButtonSize("Browse");
	padX, _ := giu.GetWindowPadding();

	return giu.Layout{
		giu.Label(label + ": "),
		giu.Row(
			giu.InputText(value).Hint(label).Size(getWindowWidth() - (bSize + padX)).OnChange(func() { if callback != nil { callback(); } }),
			button,
		),
	};

}



func modeSelector () giu.Widget {
	modes := []string {"Offline Mode", "Fake Online Mode", "Authenticated"}


	return giu.Layout{
		giu.Label("Launch Mode: "),
		giu.Combo("##launchMode", modes[wCommune.Mode], modes, &wCommune.Mode).Size(getWindowWidth()),
	};
}


func drawProfileSelector() giu.Widget {
	profileList := []string{};

	if wCommune.Profiles != nil {
		for _, profile := range *wCommune.Profiles {
			profileList = append(profileList, profile.Username);
		}
	}

	profileListTxt := "Not logged in.";
	if len(profileList) > 0 {
		profileListTxt = profileList[int(wCommune.SelectedProfile) % len(profileList)];
	}

	return giu.Style().SetDisabled(wDisabled).To(
		giu.Label("Select profile"),
		giu.Combo("##selectProfile", profileListTxt, profileList, &wCommune.SelectedProfile).Size(getWindowWidth()),
	);
}

func drawAuthenticatedSettings() giu.Widget {

	if wCommune.Mode != E_MODE_AUTHENTICATED {
		return giu.Custom(func() {});
	}

	logoutDisabled := wDisabled || (wCommune.AuthTokens == nil);
	loginDisabled := wDisabled || (wCommune.AuthTokens != nil);

	padX, _ := giu.GetWindowPadding();

	return giu.Style().SetDisabled(wDisabled).To(
		drawSeperator("Authentication"),
		giu.Row(
			giu.Button("Login (OAuth 2.0)").Disabled(loginDisabled).OnClick(func() {
				go doAuthentication();
			}).Size((getWindowWidth() / 2) - padX, 0),
			giu.Button("Logout").Disabled(logoutDisabled).OnClick(func() {
				wCommune.AuthTokens = nil;
				wCommune.Profiles = nil;
				writeSettings();
			}).Size((getWindowWidth() / 2) - padX, 0),
		),
	);

}

func drawSeperator(label string) giu.Widget {
	return giu.Custom(func() {imgui.SeparatorText(label)});
}
func updateProgress(done int64, total int64) {
	wProgress = float32(float64(done) / float64(total));
}

func createDownloadProgress () giu.Widget {
	progress := (strconv.Itoa(int(wProgress * 100.0)) + "%");

	w, _ := giu.CalcTextSize(progress);
	padX, _ := giu.GetWindowPadding();

	return giu.Layout{
		giu.Row(
			giu.ProgressBar(float32(wProgress)).Size(getWindowWidth() - (w + padX), 0),
			giu.Label(progress),
		),
	}
}

func drawUserSelection() giu.Widget {
	if wCommune.Mode == E_MODE_AUTHENTICATED {
		return drawProfileSelector()
	} else {
		return labeledTextInput("Username", &wCommune.Username, wDisabled)
	}
}

func drawStartGame() giu.Widget{

	startGameDisabled := (wCommune.Mode == E_MODE_AUTHENTICATED && wCommune.Profiles == nil) || wDisabled

	return &giu.Layout {
			giu.Style().SetDisabled(wDisabled).To(
				drawUserSelection(),
				modeSelector(),
				// maybe should seperate these two  somehow (???)
				drawSeperator("Version"),
				patchLineMenu(),
				versionMenu(),
			),
			createDownloadProgress(),
			giu.Button("Start Game").Disabled(startGameDisabled).OnClick(func() {
				go startGame();
			}).Size(getWindowWidth(), 0),
	}
}

func drawSettings() giu.Widget{

	return giu.Style().SetDisabled(wDisabled).To(
		drawSeperator("Directories"),
		giu.Tooltip("The location that the game files are stored\n(they will be downloaded here, if it's not found)").To(browseButton("Game Location", &wCommune.GameFolder, cacheVersionList)),
		giu.Tooltip("The location of the Java Runtime Environment that the game's server uses\n(it will be downloaded here, if it's not found)").To(browseButton("JRE Location", &wCommune.JreFolder, nil)),
		giu.Tooltip("The location that the games savedata will be stored,\n(worlds, mods, server list, log files, etc)").To(browseButton("User Data Location", &wCommune.UserDataFolder, nil)),
		giu.Tooltip("These are settings for advanced usecases that are not likely to be needed by most users.\nThe convention of prefixing them with a \"★\" is shamelessly stolen from PlayStation.").To(
			giu.TreeNode("★Debug Settings").Layout(
			giu.Tooltip("Allows you to run the game spoofing a specific Universal Unique Identifier (you probably dont need this)").To(labeledTextInput("★Override UUID", &wCommune.UUID, wCommune.Mode == E_MODE_AUTHENTICATED)),
		)),
	);
}




func getButtonSize(label string) float32 {
	padX, _:= giu.GetFramePadding();
	wPadX, _:= giu.GetWindowPadding();
	w, _ := giu.CalcTextSize(label)

	return (wPadX + padX + w);
}


func drawWidgets() {

	w, h := wMainWin.GetSize();
	imgui.SetWindowSizeVec2(imgui.Vec2{X: float32(w), Y: float32(h)});

	wImGuiWindow := giu.SingleWindow();

	wImGuiWindow.Layout(
		giu.TabBar().TabItems(
			giu.TabItem("Game").Layout(
				drawStartGame(),
				drawAuthenticatedSettings(),
			),
			giu.TabItem("Settings").Layout(
				drawSettings(),
			),
			/*giu.TabItem("Mods").Layout(
				giu.Custom(func(){}),
			),*/
		),
	)
}

func createWindow() error {

	wMainWin = giu.NewMasterWindow("HytaleSP", 800, 360, 0);
	if wMainWin == nil {
		return fmt.Errorf("result from NewMasterWindow was nil");
	}

	io := imgui.CurrentIO();
	io.SetConfigFlags(io.ConfigFlags() & ^imgui.ConfigFlagsViewportsEnable);

	wMainWin.SetCloseCallback(func() bool {
		defer writeSettings();

		// warn about closing hytale while the auth server emulator is still running.
		if wGameRunning && wCommune.Mode == E_MODE_FAKEONLINE {
			dlg := dialog.Message("WAIT! Hytale is still running!!!\nso .. if you close HytaleSP now-\nThe game will report \"Session Expired\" and revert to \"Offline Mode\"\nand as such; you will not be able to join servers or edit your avatar\nuntil the game is restarted ..\nDo you really want to close HytaleSP ??")
			dlg.Title("Hytale is still running!");
			if dlg.YesNo() {
				return true;
			} else {
				return false;
			}
		}

		return true;
	});


	f, err := embeddedImages.Open(path.Join("Resources", "icon.png"));
	if err != nil {
		return nil;
	}
	defer f.Close()

	image, _, err := image.Decode(f)


	wMainWin.SetIcon(image);
	wMainWin.Run(drawWidgets);


	return nil
}


func showErrorDialog(msg string, title string) {
		dlg := dialog.Message(msg);
		dlg.Title(title);
		dlg.Error();
}


func main() {

	os.MkdirAll(MainFolder(), 0775);
	os.MkdirAll(LauncherFolder(), 0775);
	os.MkdirAll(ServerDataFolder(), 0775);

	readSettings();

	os.MkdirAll(UserDataFolder(), 0775);
	os.MkdirAll(JreFolder(), 0775);
	os.MkdirAll(GameFolder(), 0775);

	cacheVersionList();
	go reAuthenticate();
	go checkForUpdates();

	err := createWindow();
	if err != nil {
		showErrorDialog(fmt.Sprintf("Error occured while creating window: %s", err), "Error while creating window");
	}

}
