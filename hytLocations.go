package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// we will still call it this internally for now;
// just for backwards compat with pre-installed versions.
var LAUNCHER_NAME = "hytLauncher";

func MainFolder() string {
	// look for explicit override of the main folder
	hytSpDir, valid := os.LookupEnv("HYTALESP_DIR");
	if valid == true {
		stat, err := os.Stat(hytSpDir);
		if err == nil && stat.IsDir() {
			return hytSpDir;
		}
	}

	// also check inside the same folder as the executable.
	exe, err := os.Executable();
	if err == nil {
		portable := filepath.Join(filepath.Dir(exe), LAUNCHER_NAME);
		stat, err := os.Stat(portable);

		if err == nil && stat.IsDir() {
			return portable;
		}
	}

	// this sucks why did i ever use this as the folder ?
	// check old directory that i used to use in v0.5 and older ..
	home, err :=  os.UserHomeDir();
	if err != nil {
		panic("Cannot find the home directory.");
	}
	oldFolder := filepath.Join(home, LAUNCHER_NAME);


	// if not found then use the new "app data" directory;
	_, err = os.Stat(oldFolder);
	if err != nil {
		switch(runtime.GOOS) {
			case "windows":
				appdata, valid := os.LookupEnv("APPDATA");
				if valid == true {
					stat, err := os.Stat(appdata);
					if err == nil && stat.IsDir() {
						return filepath.Join(appdata, LAUNCHER_NAME);
					}
				}
				return oldFolder;
			case "linux":
				config, valid := os.LookupEnv("XDG_CONFIG_HOME");
				if valid == true {
					stat, err := os.Stat(config);
					if err == nil && stat.IsDir() {
						return filepath.Join(config, LAUNCHER_NAME);
					}
				}
				return filepath.Join(home, ".config", LAUNCHER_NAME);
			case "darwin":
				return filepath.Join(home, "Library", LAUNCHER_NAME);
		}
	}

	return oldFolder;

}

func DefaultGameFolder() string {
	return filepath.Join(MainFolder(), "game", "versions");
}
func DefaultUserDataFolder() string {
	return filepath.Join(MainFolder(), "userdata");
}
func DefaultJreFolder() string {
	return filepath.Join(MainFolder(), "jre");
}

func DefaultLocalStoreFolder() string {
	return filepath.Join(MainFolder(), "localStore");
}


func LocalStoreFolder() string {
	if strings.Trim(wCommune.LocalStoreFolder, " ") == "" {
		return DefaultLocalStoreFolder();
	}

	return wCommune.LocalStoreFolder;
}

func GameFolder() string {
	if strings.Trim(wCommune.GameFolder, " ") == "" {
		return DefaultGameFolder();
	}

	return wCommune.GameFolder;
}

func UserDataFolder() string {
	if strings.Trim(wCommune.UserDataFolder, " ") == "" {
		return DefaultUserDataFolder();
	}

	return wCommune.UserDataFolder;
}

func JreFolder() string {
	if strings.Trim(wCommune.JreFolder, " ") == "" {
		return DefaultJreFolder();
	}

	return wCommune.JreFolder;
}

func LauncherFolder() string {
	return filepath.Join(MainFolder(), "launcher");
}

func ServerDataFolder() string {
	return filepath.Join(MainFolder(), "serverdata");
}


func LauncherJson() string {
	return filepath.Join(LauncherFolder(), "launcher.json");
}
