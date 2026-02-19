package main

import (
	"fmt"
	"image"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strconv"
	"time"

	"git.silica.codes/Li/UpdateChecker"
	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/AllenDang/giu"
	"github.com/ncruces/zenity"
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
	LocalStoreFolder string `json:"local_store_directory"`

	AutoUpdates bool `json:"automatic_updates"`
	FormatVersion int `json:"fmt_version"`
	SpoofLauncherVersion string `json:"spoof_launcher_version"`

	// Debug Settings
	UUID string `json:"uuid_override"`
	MaxSkins int32 `json:"max_skins_override"`
	Console bool `json:"show_console"`
	AuroraEverywhere bool `json:"aurora_everywhere"`
}


const DEFAULT_USERNAME = "TransRights";
const DEFAULT_PATCHLINE = E_PATCH_RELEASE;

const E_MODE_OFFLINE = 0;
const E_MODE_FAKEONLINE = 1;
const E_MODE_AUTHENTICATED = 2;

const E_PATCH_RELEASE = 0;
const E_PATCH_PRE_RELEASE = 1;

var (
	wMainWin *giu.MasterWindow
	wCommune = launcherCommune {
		Patchline: DEFAULT_PATCHLINE,
		Username: DEFAULT_USERNAME,
		LatestVersions: map[string]int{
			"release": 0,
			"pre-release": 0,
		},
		SelectedVersion: -1,
		Mode: E_MODE_FAKEONLINE,
		AuthTokens: nil,
		Profiles: nil,
		SelectedProfile: 0,

		AutoUpdates: true,
		GameFolder: DefaultGameFolder(),
		UserDataFolder: DefaultUserDataFolder(),
		JreFolder: DefaultJreFolder(),
		LocalStoreFolder: DefaultLocalStoreFolder(),
		SpoofLauncherVersion: "2026.02.12-54e579b",
		FormatVersion: 0,

		MaxSkins: 5,
		UUID: "",
		Console: false,
		AuroraEverywhere: true,
	};
	wProgress float32 = 0.0
	wGotLatestLauncherVersion = false
	wDisabled = false
	wSelectedTab = 0
	wGameRunning = false
	wInstalledVersions map[string]map[int]bool = map[string]map[int]bool{
		"release" : map[int]bool{},
		"pre-release" : map[int]bool{},
	}
	wImGuiWindow *giu.WindowWidget = nil;
	wVersion = "no-version";
)


func getWindowWidth() float32 {
	vec2 := imgui.ContentRegionAvail();
	return vec2.X;
}


func doAuthentication() {
	aTokens, err := getAuthTokens(wCommune.AuthTokens);

	if err != nil {
		showErrorDialog(fmt.Sprintf("Failed to get auth tokens: %s", err), "Auth failed.");
		wCommune.AuthTokens = nil;
		wCommune.Mode = E_MODE_FAKEONLINE;
		writeSettings();
	}

	wCommune.AuthTokens = &aTokens;

	// get profile list ..
	authenticatedCheckForUpdatesAndGetProfileList();

}


func checkForLauncherUpdates() {
	// cleanup .old file
	exePath, err := os.Executable();
	oldPath := exePath + ".old";
	_, err = os.Stat(oldPath);
	if err == nil {
		os.Remove(oldPath);
	}


	// dont check for updates when using `go run .`
	if wVersion == "no-version" {
		return;
	}

	if wCommune.AutoUpdates {
		UpdateChecker.Init(UpdateChecker.Repository{
			Domain: "git.silica.codes",
			Owner: "Li",
			Name: "HytaleSP",
		}, wVersion);

		updateInfo, err :=  UpdateChecker.CheckForUpdate();

		if err != nil {
			return;
		}

		if updateInfo != nil {
			err := zenity.Question(fmt.Sprintf("A new version of HytaleSP was found!\nVersion: %s\n%s", updateInfo.NewVersion, updateInfo.Description),
					       zenity.Title("HytaleSP Update"), zenity.QuestionIcon, zenity.OKLabel("Install"), zenity.CancelLabel("Not Now"), zenity.ExtraButton("Never"));
			if err != nil {
				if err == zenity.ErrCanceled {
					return;
				}
				if(err == zenity.ErrExtraButton) {
					wCommune.AutoUpdates = false;
				}
			}


			if err != nil {
				showErrorDialog(fmt.Sprintf("Error getting update: %s", err), "Update Fail");
				return;
			}

			success := false;

			switch(runtime.GOOS) {
				case "windows":
					success = UpdateChecker.DownloadUpdateIfPresent("HytaleSP.exe", exePath);
				case "linux":
					success = UpdateChecker.DownloadUpdateIfPresent("HytaleSP", exePath);
			}

			if success {
				e := exec.Command(exePath);
				e.Args = os.Args;
				e.Start();
				os.Exit(0);
			} else {
				showErrorDialog("Error to download update.", "Update Fail");
				return;
			}

		}

		fmt.Printf("No update found.\n");
		return;


	}
}


func refreshAuthentication() {
	if wCommune.AuthTokens != nil && wCommune.Mode == E_MODE_AUTHENTICATED {
		_, session, err := unmakeJwt(wCommune.AuthTokens.AccessToken);
		if err != nil {
			showErrorDialog(fmt.Sprintf("Failed to parse JWT %s", err), "Auth failed.");
			wCommune.AuthTokens = nil;
			wCommune.Mode = E_MODE_FAKEONLINE;
			writeSettings();
			return;
		}

		if time.Now().Unix() > int64(session.Exp) {
			aTokens, err:= getAuthTokens(*wCommune.AuthTokens);


			if err != nil {
				showErrorDialog(fmt.Sprintf("Failed to authenticate: %s", err), "Auth failed.");
				wCommune.AuthTokens = nil;
				wCommune.Mode = E_MODE_FAKEONLINE;
				writeSettings();
				return;
			}

			wCommune.AuthTokens = &aTokens;
		}

		authenticatedCheckForUpdatesAndGetProfileList();
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

	if !isJreInstalled() {
		err := installJre(updateProgress);

		if err != nil {
			showErrorDialog(fmt.Sprintf("Error getting the JRE: %s", err), "Install JRE failed.");
			return;
		};
	}

	if !isGameVersionInstalled(ver, valToChannel(int(wCommune.Patchline))) {
		err := installGame(ver, valToChannel(int(wCommune.Patchline)), updateProgress);

		if err != nil {
			showErrorDialog(fmt.Sprintf("Error getting the game: %s", err), "Install game failed.");
			return;
		};

		wInstalledVersions[channel][ver] = true;
	}

	// set game running flag
	wGameRunning = true;
	err := launchGame(ver, channel, wCommune.Username, getUUID());

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
	selectedChannel := valToChannel(int(wCommune.Patchline));
	versions := []string {};
	/* for key := range wInstalledVersions[selectedChannel] {
		text := "Version " + strconv.Itoa(key);
		if wInstalledVersions[selectedChannel][key] {
			text += " - installed";
		} else {
			text += " - not installed";
		}
		versions = append(versions, text);
	} */


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


	bSize := getButtonSize("Delete")
	padX, _ := giu.GetWindowPadding();

	versionIndex := int(wCommune.SelectedVersion);
	if versionIndex < 0 {
		versionIndex = 0;
	} else if versionIndex >= len(versions) {
		versionIndex = len(versions)-1;
	}

	currentVersion := "No Versions Found.";
	if len(versions) > 0 {
		currentVersion = versions[versionIndex];
	}

	selectedVersion := versionIndex + 1;
	buttondisabled := !wInstalledVersions[selectedChannel][selectedVersion] || wDisabled;

	return giu.Layout{
		giu.Label("Version: "),
		giu.Row(
			giu.Combo("##version", currentVersion, versions, &wCommune.SelectedVersion).Size(getWindowWidth() - (bSize + padX)),
			giu.Button("Delete").Disabled(buttondisabled).OnClick(func() {
				wDisabled = true;

				go func() {
					defer func() { wDisabled = false; }();
					err := deleteVersion(selectedVersion, selectedChannel)

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

func labeledIntInput(label string, value *int32, disabled bool) giu.Widget {
	if value == nil {
		panic("failed to initalize browse button");
	}

	return giu.Style().SetDisabled(wDisabled || disabled).To(
		giu.Label(label+": "),
		giu.Row(
			giu.InputInt(value).Size(getWindowWidth()),
		),
	);

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
		dir, err := zenity.SelectFile(zenity.Directory());
		if err != nil {
			if err != zenity.ErrCanceled {
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
		giu.Combo("##launchMode", modes[wCommune.Mode], modes, &wCommune.Mode).Size(getWindowWidth()).OnChange(func() {
			if wCommune.Mode == E_MODE_AUTHENTICATED {
				go refreshAuthentication();
			}
		}),
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
				go func() {
					wDisabled = true;
					defer func() {wDisabled = false;}();
					if giu.IsKeyDown(giu.KeyLeftShift) {
						zenity.SelectFile()
					}

					err := downloadAndRunGameEx(int(wCommune.SelectedVersion)+1, valToChannel(int(wCommune.Patchline)),
							func() {},
							func(channel string, version int){wInstalledVersions[channel][version] = true},
							func() {wGameRunning = true},
							func() {wGameRunning = false},
							updateProgress);
					if err != nil {
						showErrorDialog(fmt.Sprintf("Error: %s", err), "Error launching the game");
					}

				}();
			}).Size(getWindowWidth(), 0),
	}
}

func drawSettings() giu.Widget{

	return giu.Style().SetDisabled(wDisabled).To(
		drawSeperator("Directories"),
		giu.Tooltip("The location that the game files are stored\n(they will be downloaded here, if it's not found)").To(browseButton("Game Location", &wCommune.GameFolder, cacheAllVersions)),
		giu.Tooltip("The location of the Java Runtime Environment that the game's server uses\n(it will be downloaded here, if it's not found)").To(browseButton("JRE Location", &wCommune.JreFolder, nil)),
		giu.Tooltip("The location that the games savedata will be stored,\n(worlds, mods, server list, log files, etc)").To(browseButton("User Data Location", &wCommune.UserDataFolder, nil)),
		drawSeperator("Launcher"),
		giu.Tooltip("Toggles wether or not to automatically check for updates to the launcher.\n").To(giu.Checkbox("Check For Updates", &wCommune.AutoUpdates)),

		giu.TreeNode("★Debug Settings").Layout(
			giu.Tooltip("Allows you to run the game spoofing a specific Universal Unique Identifier").To(labeledTextInput("★Override UUID", &wCommune.UUID, wCommune.Mode == E_MODE_AUTHENTICATED)),
			giu.Tooltip("Allows you to override the maximum amount of skin presets (default: 5)").To(labeledIntInput("★Override Max Skins", &wCommune.MaxSkins, wCommune.Mode == E_MODE_AUTHENTICATED)),
			giu.TreeNode("★Aurora Settings").Layout(
				giu.Tooltip("Display stdout/stderr output for HytaleClient.").To(giu.Style().SetDisabled(wDisabled).To(giu.Checkbox("★Enable Console", &wCommune.Console))),
				giu.Tooltip("This will allow you to join offline/insecure server and disable telemetry; even while using offical authentication").To(giu.Style().SetDisabled(wDisabled).To(giu.Checkbox("★Enable Aurora in Authenticated Mode", &wCommune.AuroraEverywhere))),
			),
		),
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

	wMainWin = giu.NewMasterWindow(fmt.Sprintf("HytaleSP %s", wVersion), 800, 360, 0);
	if wMainWin == nil {
		return fmt.Errorf("result from NewMasterWindow was nil");
	}

	io := imgui.CurrentIO();
	io.SetConfigFlags(io.ConfigFlags() & ^imgui.ConfigFlagsViewportsEnable);

	wMainWin.SetCloseCallback(func() bool {
		defer writeSettings();

		// warn about closing hytale while the auth server emulator is still running.
		if wGameRunning && wCommune.Mode == E_MODE_FAKEONLINE {
			err := zenity.Question("WAIT! Hytale is still running!!!\nso .. if you close HytaleSP now-\nThe game will report \"Session Expired\"\nand revert to \"Offline Mode\"\nand as such; you will not be able to join servers or edit your avatar\nuntil the game is restarted ..\nDo you really want to close HytaleSP ??", zenity.Title("Hytale is still running"), zenity.QuestionIcon, zenity.OKLabel("Yes"), zenity.CancelLabel("No"));

			if err == zenity.ErrCanceled {
				return false;
			} else {
				return true;
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
	zenity.Error(msg, zenity.Title(title), zenity.ErrorIcon);
}


func main() {

	if len(os.Args) > 1 {
		ret := handleCli(os.Args);
		if ret <= 0 {
			os.Exit(ret);
		}
	}

	go checkForLauncherUpdates();

	os.MkdirAll(MainFolder(), 0775);
	os.MkdirAll(LauncherFolder(), 0775);
	os.MkdirAll(ServerDataFolder(), 0775);

	readSettings();

	os.MkdirAll(UserDataFolder(), 0775);
	os.MkdirAll(JreFolder(), 0775);
	os.MkdirAll(GameFolder(), 0775);

	go refreshAuthentication();
	go checkForGameUpdates();

	dataFixerUpper(wCommune.FormatVersion);

	err := createWindow();
	if err != nil {
		showErrorDialog(fmt.Sprintf("Error occured while creating window: %s", err), "Error while creating window");
	}

}
