//go:generate goversioninfo

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"

	"bitbucket.org/rj/goey"
	"bitbucket.org/rj/goey/base"
	"bitbucket.org/rj/goey/loop"
	"bitbucket.org/rj/goey/windows"
	"github.com/sqweek/dialog"
)

type launcherCommune struct {
	Patchline string `json:"last_patchline"`
	Username string `json:"last_username"`
	SelectedVersion int `json:"last_version"`
	LatestVersions map[string]int `json:"last_version_scan_result"`
	GameFolder string `json:"install_directory"`
	Mode string `json:"mode"`
}


var(
	wMainWin *windows.Window
	wCommune = launcherCommune {
		Patchline: "release",
		Username: "TransRights",
		LatestVersions: map[string]int{
			"release": 4,
			"pre-release": 8,
		},
		SelectedVersion: 4,
		GameFolder: DefaultGameFolder(),
		Mode: "fakeonline",
	};
	wProgress = 0
	wDisabled = false
)



func checkForUpdates() {
	lastRelease := wCommune.LatestVersions["release"]
	lastPreRelease := wCommune.LatestVersions["pre-release"]

	latestRelease := findLatestVersionNoAuth(lastRelease, runtime.GOARCH, runtime.GOOS, "release");
	latestPreRelease := findLatestVersionNoAuth(lastPreRelease, runtime.GOARCH, runtime.GOOS, "pre-release");

	fmt.Printf("latestRelease: %d\n", latestRelease);
	fmt.Printf("latestPreRelease: %d\n", latestPreRelease);

	if latestRelease > lastRelease {
		fmt.Printf("Found new release version: %d", latestRelease);
		wCommune.LatestVersions["release"] = latestRelease;
	}

	if latestPreRelease > lastPreRelease {
		fmt.Printf("Found new pre-release version: %d", latestPreRelease);
		wCommune.LatestVersions["pre-release"] = latestPreRelease;
	}

	if wMainWin != nil {
		updateWindow();
		writeSettings();
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
		fmt.Printf("patchline: %s\n", wCommune.Patchline);
		fmt.Printf("last used version: %d\n", wCommune.SelectedVersion);
		fmt.Printf("newest known release: %d\n", wCommune.LatestVersions["release"])
		fmt.Printf("newest known pre-release: %d\n", wCommune.LatestVersions["pre-release"])

		// check for updates in the background
		go checkForUpdates();
	}
}


func valToChannel(vchl int) string {
	switch vchl {
		case 0:
			return "release";
		case 1:
			return "pre-release";
		default:
			return "release";
	}
}

func channelToVal(channel string) int {
	switch channel {
		case "release":
			return 0;
		case "pre-release":
			return 1;
		default:
			return 0;
	}
}

func patchLineMenu() base.Widget {
	return &goey.VBox{
		AlignMain: goey.SpaceBetween,
		Children: []base.Widget{
			&goey.Label{Text: "Patchline:"},
			&goey.SelectInput{
				Items: []string {
					"release",
					"pre-release",
				},
				Value: channelToVal(wCommune.Patchline),
				Disabled: wDisabled,
				OnChange: func(v int) {
					if wCommune.Patchline != valToChannel(v) {
						wCommune.Patchline = valToChannel(v);
						updateWindow();
					}
				},
			},
		},
	};
}


func versionMenu() base.Widget {
	versions := goey.SelectInput {
		OnChange: func(v int) {
			wCommune.SelectedVersion = v+1;
			updateWindow()
		},
		Disabled: wDisabled,
	};
	latest := wCommune.LatestVersions[wCommune.Patchline];

	for i := range latest {
		txt := "Version "+strconv.Itoa(i+1);
		if isGameVersionInstalled(i+1, wCommune.Patchline) {
			txt += " - installed";
		} else {
			txt += " - not installed";
		}

		versions.Items = append(versions.Items, txt);
	}

	selectedVersion := wCommune.SelectedVersion;
	selectedChannel := wCommune.Patchline;

	versions.Value = (selectedVersion-1);
	disabled := !isGameVersionInstalled(selectedVersion, selectedChannel) || wDisabled;

	return &goey.VBox{
		AlignMain: goey.SpaceBetween,
		Children: []base.Widget{
			&goey.Label{Text: "Version:"},

			&goey.HBox{
				AlignCross: goey.CrossCenter,
				Children: []base.Widget{
					&goey.Expand{
						Child: &versions,
					},
					&goey.Button {
						Text: "Delete",
						Disabled: disabled,
						OnClick: func() {
							wDisabled = true;
							updateWindow();

							go func() {
								installDir := getVersionInstallPath(selectedVersion, wCommune.Patchline);
								err := os.RemoveAll(installDir);
								if err != nil {
									fmt.Printf("failed to remove: %s", err);
									wMainWin.Message(fmt.Sprintf("failed to remove: %s", err)).WithError().WithTitle("Failed to remove").Show();
								}
								wDisabled = false;
								updateWindow();
							}();
						},
					},
				},
			},
		},
	};
}

func installLocation() base.Widget {
	return &goey.VBox {
			AlignMain: goey.SpaceBetween,
			Children: []base.Widget {
				&goey.HR{},
				&goey.Label{ Text: "Install Location:" },
				&goey.HBox{
					AlignCross: goey.CrossCenter,
					Children: []base.Widget {
						&goey.Expand{
							Child: &goey.TextInput{
								Placeholder: "Install Location",
								Disabled: wDisabled,
								Value: wCommune.GameFolder,
								OnChange: func(v string) {
									wCommune.GameFolder = v;
									updateWindow();
								},
							},
						},
						&goey.Button{
							Text: "Browse",
							Disabled: wDisabled,
							OnClick: func() {
								dir, err := dialog.Directory().Title("Select install location").Browse();
								if err != nil {
									if err != dialog.ErrCancelled {
										errorMsg := fmt.Sprintf("Failed: %s", err);
										wMainWin.Message(errorMsg).WithError().WithTitle("Error reading directory.").Show();
									}
								}

								wCommune.GameFolder = dir;
								updateWindow();
							},
						},
					},
				},
			},
	};
}

func modeSelector () base.Widget {
	var v int;
	if wCommune.Mode == "fakeonline" {
		v = 1;
	} else {
		v = 0;
	}

	return &goey.VBox {
		AlignMain: goey.SpaceBetween,
		Children: []base.Widget{
			&goey.Label{Text: "Launch Mode:"},
			&goey.SelectInput{
				Items: []string {
					"Offline Mode",
					"Fake Online Mode",
					//"Authenticated",
				},
				Value: v,
				Disabled: wDisabled,
				OnChange: func(v int) {
					if v == 1{
						wCommune.Mode = "fakeonline";
					} else {
						wCommune.Mode = "offline";
					}
				},
			},
		},
	}
}

func usernameBox() base.Widget {
	return &goey.VBox{
		AlignMain: goey.SpaceBetween,
		Children: []base.Widget{
			&goey.Label{Text: "Username:"},
			&goey.TextInput{
					Value: wCommune.Username,
					Placeholder: "Username",
					Disabled: wDisabled,
					OnChange: func(v string) {
						wCommune.Username = v;
					},
			},
		},
	};
}


func updateProgress(done int64, total int64) {
	lastProgress := wProgress;
	wProgress = int((float64(done) / float64(total)) * 100.0);
	if lastProgress != wProgress{
		updateWindow();
	}
}

func createWindow() error {
	w, err := windows.NewWindow("hytLauncher", renderWindow())
	if err != nil {
		return err
	}

	w.SetScroll(false, false);

	w.SetOnClosing(func() bool {
			writeSettings();
			return false;
	});

	wMainWin = w;
	return nil
}


func updateWindow() {
	err := wMainWin.SetChild(renderWindow())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
	}
}

func renderWindow() base.Widget {
	return &goey.Padding{
		Insets: goey.DefaultInsets(),
		Child: &goey.Align{
			Child: &goey.VBox{
				AlignMain: goey.MainStart,
				Children: []base.Widget{

					usernameBox(),
					patchLineMenu(),
					versionMenu(),

					&goey.Progress{
						Value: wProgress,
						Min: 0,
						Max: 100,
					},
					&goey.Button{
						Text: "Start Game",
						Disabled: wDisabled,
						OnClick: func() {
							go func() {
								wDisabled = true;
								updateWindow();

								installJre(updateProgress);
								installGame(wCommune.SelectedVersion, wCommune.Patchline, updateProgress);
								launchGame(wCommune.SelectedVersion, wCommune.Patchline, wCommune.Username, usernameToUuid(wCommune.Username));

								wDisabled = false;
								updateWindow();
							}();
						},
					},
					installLocation(),
					modeSelector(),
				},
			},
		},
	}
}


func main() {

	os.MkdirAll(MainFolder(), 0775);
	os.MkdirAll(LauncherFolder(), 0775);
	os.MkdirAll(UserDataFolder(), 0775);
	os.MkdirAll(JreFolder(), 0775);
	os.MkdirAll(ServerDataFolder(), 0775);
	readSettings();
	os.MkdirAll(GameFolder(), 0775);
;

	err := loop.Run(createWindow)
	if err != nil {
		fmt.Println("Error: ", err)
	}

}
