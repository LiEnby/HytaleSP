package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/google/uuid"
)
var SERVER_DATA_FOLDER = path.Join(MAIN_FOLDER, "serverdata");
var ACCOUNT_INFO = accountInfo{
	Username: "Anonymous",
	UUID: uuid.NewString(),
	Entitlements: []string {"game.base", "game.deluxe", "game.founder" },
	CreatedAt: time.Now(),
	NextNameChangeAt: time.Now(),
	Skin: "{\"bodyCharacteristic\":\"Default.11\",\"underwear\":\"Bra.Blue\",\"face\":\"Face_Neutral\",\"ears\":\"Ogre_Ears\",\"mouth\":\"Mouth_Makeup\",\"haircut\":\"SideBuns.Black\",\"facialHair\":null,\"eyebrows\":\"RoundThin.Black\",\"eyes\":\"Plain_Eyes.Green\",\"pants\":\"Icecream_Skirt.Strawberry\",\"overpants\":\"LongSocks_Bow.Lime\",\"undertop\":\"VNeck_Shirt.Black\",\"overtop\":\"NeckHigh_Savanna.Pink\",\"shoes\":\"Wellies.Orange\",\"headAccessory\":null,\"faceAccessory\":null,\"earAccessory\":null,\"skinFeature\":null,\"gloves\":null,\"cape\":null}",
};

func readSkinData() {
	load := path.Join(SERVER_DATA_FOLDER, "skin.json");
	os.MkdirAll(path.Dir(load), 0666);

	_, err := os.Stat(load);
	if err != nil {
		return;
	}
	skinData, _ := os.ReadFile(load);
	ACCOUNT_INFO.Skin = string(skinData);

}

func writeSkinData(newData string) {
	save := path.Join(SERVER_DATA_FOLDER, "skin.json");
	os.MkdirAll(path.Dir(save), 0666);
	fmt.Printf("Writing skin data %s\n", save);


	os.WriteFile(save, []byte(newData), 0666);
	ACCOUNT_INFO.Skin = newData;
}

func readCosmetics() string {
	load := path.Join(SERVER_DATA_FOLDER, "cosmetics.json");
	os.MkdirAll(path.Dir(load), 0666);

	_, err := os.Stat(load);
	if err != nil {
		return "{}";
	}

	data, err := os.ReadFile(load);

	if err != nil{
		return "{}";
	}

	return string(data);
}

func getAccountInfo() accountInfo {
	readSkinData();
	return ACCOUNT_INFO;
}

func handleMyAccountSkin(w http.ResponseWriter, req *http.Request) {

	switch req.Method {
		case "PUT":
			data, _ := io.ReadAll(req.Body);
			writeSkinData(string(data));
			w.WriteHeader(204);
	}
}

func handleMyAccountCosmetics(w http.ResponseWriter, req *http.Request) {

	switch req.Method {
		case "GET":
			w.Write([]byte(readCosmetics()));

	}
}

func handleMyAccountGameProfile(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
		case "GET":
			w.Header().Add("Content-Type", "application/json");
			w.WriteHeader(200);
			json.NewEncoder(w).Encode(getAccountInfo());
	}
}


func handleMyAccountLauncherData(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
		case "GET":
		data := launcherData {
			EulaAcceptedAt: time.Now(),
			Owner: uuid.NewString(),
			Patchlines: patchlines{
				PreRelease: gameVersion {
					BuildVersion: "2026.01.14-3e7a0ba6c",
					Newest: 4,
				},
				Release: gameVersion {
					BuildVersion: "2026.01.13-50e69c385",
					Newest: 3,
				},
			},
			Profiles: []accountInfo {
				getAccountInfo(),
			},
		}
		w.Header().Add("Content-Type", "application/json");
		w.WriteHeader(200);
		json.NewEncoder(w).Encode(data);
	}
}

func handleManifest(w http.ResponseWriter, req *http.Request) {

	target := req.PathValue("target")
	arch := req.PathValue("arch")
	branch := req.PathValue("branch")
	patch := req.PathValue("patch")
	fmt.Printf("target: %s\narch: %s\nbranch: %s\npatch: %s\n", target, arch, branch, patch);

	p := path.Join("patches", target, arch, branch, patch, "manifest.json");

	http.ServeFile(w, req, p);

}

func handlePatches(w http.ResponseWriter, req *http.Request) {

	filepath := req.PathValue("filepath");
	p := path.Join("patches", filepath);

	_, err := os.Stat(p);
	if err != nil {
		return;
	}

	http.ServeFile(w, req, p);
}

func runServer(name string, uid string) {

	ACCOUNT_INFO.UUID = uid;
	ACCOUNT_INFO.Username = name;

	mux := http.NewServeMux()
	mux.HandleFunc("/my-account/game-profile", handleMyAccountGameProfile);
	mux.HandleFunc("/my-account/skin", handleMyAccountSkin)
	mux.HandleFunc("/my-account/cosmetics", handleMyAccountCosmetics)
	mux.HandleFunc("/my-account/get-launcher-data", handleMyAccountLauncherData);

	mux.HandleFunc("/patches/{target}/{arch}/{branch}/{patch}", handleManifest);
	mux.HandleFunc("/patches/{filepath...}", handlePatches);

	var handler  http.Handler = mux
	handler = logRequestHandler(handler)

	err := http.ListenAndServe(":8080", handler)
	if err != nil {
		fmt.Printf("Failed to listen error=%s\n", err)
		os.Exit(1)
	}
}


func fakeSign() string {
	data := make([]byte, 0x40)
	if _, err := rand.Read(data); err != nil {
		panic(err)
	}

	return base64.URLEncoding.EncodeToString(data);
}

func generateSessionJwt() string {
	head := jwtHeader{
		Alg: "EdDSA",
		Kid: "2025-10-01",
		Typ: "JWT",
	};


	sesTok := sessionToken {
		Exp: int(time.Now().AddDate(0,0,1).Unix()),
		Iat: int(time.Now().Unix()),
		Iss: "https://sessions.hytale.com",
		Jti: uuid.NewString(),
		Scope: "hytale:client",
		Sub: uuid.NewString(),
	};

	return b64json(head) + "." + b64json(sesTok) + "." + fakeSign();
}


func generateIdentityJwt() string {
	head := jwtHeader{
		Alg: "EdDSA",
		Kid: "2025-10-01",
		Typ: "JWT",
	};

	idTok := identityToken {
		Exp: int(time.Now().AddDate(0,0,1).Unix()),
		Iat: int(time.Now().Unix()),
		Iss: "https://sessions.hytale.com",
		Jti: uuid.NewString(),
		Scope: "hytale:client",
		Sub: uuid.NewString(),
		Profile: profileInfo {
			Username: ACCOUNT_INFO.Username,
			Entitlements: ACCOUNT_INFO.Entitlements,
			Skin: ACCOUNT_INFO.Skin,
		},
	};

	return b64json(head) + "." + b64json(idTok) + "." + fakeSign();
}
