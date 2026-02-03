package main

import (
	"archive/zip"
	"crypto/ed25519"
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

var ENTITLEMENTS = []string {"game.base", "game.deluxe", "game.founder"};

const DEFAULT_COSMETICS = "{\"bodyCharacteristic\":[\"Default\",\"Muscular\"],\"cape\":[\"Cape_Royal_Emissary\",\"Cape_New_Beginning\",\"Cape_Forest_Guardian\",\"Cape_PopStar\",\"Cape_Scavenger\",\"Cape_Knight\",\"Cape_Seasons\",\"Hope_Of_Gaia_Cape\",\"Cape_Blazen_Wizard\",\"Cape_King\",\"Cape_Void_Hero\",\"Cape_Featherbound\",\"FrostwardenSet_Cape\",\"Cape_Bannerlord\",\"Cape_Wasteland_Marauder\"],\"earAccessory\":[\"EarHoops\",\"SimpleEarring\",\"DoubleEarrings\",\"SilverHoopsBead\",\"SpiralEarring\",\"AcornEarrings\"],\"ears\":[\"Default\",\"Elf_Ears\",\"Elf_Ears_Large\",\"Elf_Ears_Large_Down\",\"Elf_Ears_Small\",\"Ogre_Ears\"],\"eyebrows\":[\"Medium\",\"Thin\",\"Thick\",\"Bushy\",\"Shaved\",\"SmallRound\",\"Large\",\"RoundThin\",\"Angry\",\"Plucked\",\"Square\",\"Serious\",\"BushyThin\",\"Heavy\"],\"eyes\":[\"Medium_Eyes\",\"Large_Eyes\",\"Plain_Eyes\",\"Almond_Eyes\",\"Square_Eyes\",\"Reptile_Eyes\",\"Cat_Eyes\",\"Demonic_Eyes\",\"Goat_Eyes\"],\"face\":[\"Face_Neutral\",\"Face_Neutral_Freckles\",\"Face_Sunken\",\"Face_Tired_Eyes\",\"Face_Stubble\",\"Face_Scar\",\"Face_Aged\",\"Face_Older2\",\"Face_Almond_Eyes\",\"Face_MakeUp\",\"Face_Make_Up_2\",\"Face_MakeUp_Freckles\",\"Face_MakeUp_Highlight\",\"Face_MakeUp_6\",\"Face_MakeUp_Older\",\"Face_MakeUp_Older2\"],\"faceAccessory\":[\"EyePatch\",\"Glasses\",\"LargeGlasses\",\"MedicalEyePatch\",\"MouthCover\",\"MouthWheat\",\"ColouredGlasses\",\"CrazyGlasses\",\"RoundGlasses\",\"HeartGlasses\",\"AgentGlasses\",\"SunGlasses\",\"AviatorGlasses\",\"BusinessGlasses\",\"Plaster\",\"Glasses_Monocle\",\"GlassesTiny\",\"Goggles_Wasteland_Marauder\"],\"facialHair\":[\"Medium\",\"Beard_Large\",\"Goatee\",\"Chin_Curtain\",\"Moustache\",\"VikingBeard\",\"TwirlyMoustache\",\"SoulPatch\",\"PirateBeard\",\"TripleBraid\",\"DoubleBraid\",\"GoateeLong\",\"PirateGoatee\",\"Soldier\",\"Hip\",\"Trimmed\",\"Handlebar\",\"Groomed\",\"Stylish\",\"ThinGoatee\",\"Short_Trimmed\",\"Groomed_Large\",\"WavyLongBeard\",\"CurlyLongBeard\"],\"gloves\":[\"BasicGloves_Basic\",\"BoxingGloves\",\"FlowerBracer\",\"MiningGloves\",\"GoldenBracelets\",\"LeatherMittens\",\"Straps_Leather\",\"Shackles_Feran\",\"CatacombCrawler_Gloves\",\"Hope_Of_Gaia_Gloves\",\"Gloves_Void_Hero\",\"LongGloves_Popstar\",\"Gloves_Medium_Featherbound\",\"Arctic_Scout_Gloves\",\"Scavenger_Gloves\",\"Bracer_Daisy\",\"LongGloves_Savanna\",\"Gloves_Wasteland_Marauder\",\"Gloves_Blazen_Wizard\",\"Merchant_Gloves\",\"Battleworn_Gloves\"],\"haircut\":[\"Morning\",\"Bangs\",\"Quiff\",\"Lazy\",\"BobCut\",\"Messy\",\"Viking\",\"Fringe\",\"PonyTail\",\"Bun\",\"Braid\",\"BraidDouble\",\"ShortDreads\",\"Undercut\",\"Samurai\",\"DoublePart\",\"Rustic\",\"RoseBun\",\"SideBuns\",\"SmallPigtails\",\"Stylish\",\"Mohawk\",\"BowlCut\",\"Emo\",\"Pigtails\",\"Sideslick\",\"SingleSidePigtail\",\"Slickback\",\"WavyPonytail\",\"Wings\",\"ChopsticksPonyTail\",\"Curly\",\"MessyBobcut\",\"Simple\",\"WidePonytail\",\"RaiderMohawk\",\"MidSinglePart\",\"AfroPuffs\",\"PuffyQuiff\",\"GenericPuffy\",\"PuffyPonytail\",\"FighterBuns\",\"MaleElf\",\"Windswept\",\"SidePonytail\",\"PonyBuns\",\"ElfBackBun\",\"BraidedPonytail\",\"ThickBraid\",\"WavyBraids\",\"VikinManBun\",\"Witch\",\"FrizzyLong\",\"WavyLong\",\"SuperSlickback\",\"Cat\",\"Scavenger_Hair\",\"LongTied\",\"LongBangs\",\"BantuKnot\",\"Berserker\",\"CuteEmoBangs\",\"CutePart\",\"LongPigtails\",\"FeatheredHair\",\"LongHairPigtail\",\"StraightHairBun\",\"SuperSideSlick\",\"FrontTied\",\"EmoWavy\",\"MessyMop\",\"EmoBangs\",\"BowHair\",\"Greaser\",\"FrontFlick\",\"Long\",\"WavyShort\",\"GenericLong\",\"GenericMedium\",\"GenericShort\",\"CurlyShort\",\"LongCurly\",\"MorningLong\",\"CentrePart\",\"VikingWarrior\",\"MediumCurly\",\"SpikedUp\",\"Cowlick\",\"MessyWavy\",\"BuzzCut\",\"QuiffLeft\",\"StylishWindswept\",\"SuperShirt\",\"StylishQuiff\",\"BangsShavedBack\",\"FrizzyVolume\",\"Cornrows\",\"Balding\",\"Dreadlocks\"],\"headAccessory\":[\"Goggles\",\"Hoodie\",\"GiHeadband\",\"ForeheadProtector\",\"FlowerCrown\",\"Bandana\",\"FloppyBeanie\",\"BunnyBeanie\",\"Headband\",\"CatBeanie\",\"FrogBeanie\",\"WorkoutCap\",\"HeadDaliah\",\"HairRose\",\"HairPeony\",\"HairDaisy\",\"Logo_Cap\",\"BanjoHat\",\"WitchHat\",\"StrawHat\",\"PirateBandana\",\"HairHibiscus\",\"SantaHat\",\"ElfHat\",\"Head_Crown\",\"HeadphonesDadCap\",\"Headphones\",\"Beanie\",\"BandanaSkull\",\"StripedBeanie\",\"Head_Tiara\",\"Viking_Helmet\",\"Pirate_Captain_Hat\",\"TopHat\",\"CowboyHat\",\"RusticBeanie\",\"LeatherCap\",\"Ribbon\",\"Bunny_Ears\",\"Head_Bandage\",\"AcornNecktie\",\"AcornHairclip\",\"Forest_Guardian_Hat\",\"Hoodie_Feran\",\"ExplorerGoggles\",\"Hope_Of_Gaia_Crown\",\"FrostwardenSet_Hat\",\"Arctic_Scout_Hat\",\"Savanna_Scout_Hat\",\"BulkyBeanie\",\"Hat_Popstar\",\"ShapedCap_Chill\",\"Hoodie_Ornated\",\"Headband_Void_Hero\",\"Hood_Blazen_Wizard\",\"Merchant_Beret\",\"Battleworn_Helm\"],\"mouth\":[\"Mouth_Default\",\"Mouth_Makeup\",\"Mouth_Thin\",\"Mouth_Long\",\"Mouth_Tiny\"],\"overpants\":[\"KneePads\",\"LongSocks_Plain\",\"LongSocks_BasicWrap\",\"LongSocks_School\",\"LongSocks_Striped\",\"LongSocks_Bow\",\"LongSocks_Torn\"],\"overtop\":[\"PuffyJacket\",\"Tartan\",\"BunnyHoody\",\"StylishJacket\",\"LongBeltedJacket\",\"RobeOvertops\",\"HeroShirt\",\"ThreadedOvertops\",\"RaggedVest\",\"Winter_Jacket\",\"Suit_Jacket\",\"Wool_Jersey\",\"Chest_PuffyJersey\",\"Tunic_Weathered\",\"JacketShort\",\"JacketLong\",\"Coat\",\"TrenchCoat\",\"VikingVest\",\"GiShirt\",\"ShortTartan\",\"BulkyShirtLong\",\"MiniLeather\",\"Fantasy\",\"Pirate\",\"BulkyShirt_Scarf\",\"Scarf_Large_Stripped\",\"Scarf_Large\",\"BulkyShirtLong_LeatherJacket\",\"ForestVest\",\"BulkyShirt_StomachWrap\",\"FantasyShawl\",\"LeatherVest\",\"BulkyShirt_RoyalRobe\",\"Jinbaori\",\"Ronin\",\"MessyShirt\",\"StitchedShirt\",\"OpenShirtBand\",\"BulkyShirt_RuralShirt\",\"BulkyShirt_RuralPattern\",\"HeartNecklace\",\"Shark_Tooth_Necklace\",\"Pookah_Necklace\",\"Golden_Bangles\",\"BulkyShirt_FancyWaistcoat\",\"LetterJacket\",\"PinstripeJacket\",\"DoubleButtonJacket\",\"Polarneck\",\"FlowyHalf\",\"FurLinedJacket\",\"PlainHoodie\",\"LooseSweater\",\"SimpleDress\",\"SleevedDress\",\"SleevedDresswJersey\",\"Tunic_Long\",\"Scarf\",\"TracksuitJacket\",\"Jacket\",\"KhakiShirt\",\"LongCardigan\",\"GoldtrimJacket\",\"Cheststrap\",\"SantaJacket\",\"ElfJacket\",\"FarmerVest\",\"AviatorJacket\",\"QuiltedTop\",\"Jinbaori_Wave\",\"Jinbaori_Flower\",\"FloppyBunnyJersey\",\"PlainJersey\",\"Tunic_Villager\",\"RoughFabricBand\",\"Arm_Bandage\",\"Farmer_Dress\",\"OnePiece_SchoolDress\",\"OnePiece_ApronDress\",\"Noble_Beige\",\"Fancy_Coat\",\"Adventurer_Dress\",\"Oasis_Dress\",\"PuffyBomber\",\"Jacket_Voyager\",\"AlpineExplorerJumper\",\"Hope_Of_GaiaOvertop\",\"DaisyTop\",\"Arctic_Scout_Jacket\",\"Collared_Cool\",\"NeckHigh_Savanna\",\"Scavenger_Poncho\",\"NeckHigh_LeatherClad\",\"Jacket_Popstar\",\"Voidbearer_Top\",\"Featherbound_Tunic\",\"Forest_Guardian_Poncho\",\"Jacket_Void_Hero\",\"Straps_Wasteland_Marauder\",\"Robe_Blazen_Wizard\",\"Merchant_Tunic\",\"Battleworn_Tunic\",\"Bannerlord_Tunic\"],\"pants\":[\"ApprenticePants\",\"LeatherPants\",\"SurvivorPants\",\"StripedPants\",\"CostumePants\",\"ShortyRolled\",\"Jeans\",\"GiPants\",\"Forest_Bermuda\",\"BulkySuede\",\"Pants_Straight_WreckedJeans\",\"Pants_Slim\",\"Dungarees\",\"StylishShorts\",\"JeansStrapped\",\"Villager_Bermuda\",\"ExplorerShorts\",\"Explorer_Trousers\",\"PinstripeTrousers\",\"Pants_Slim_Faded\",\"Pants_Slim_Tracksuit\",\"LongDungarees\",\"KhakiShorts\",\"ColouredKhaki\",\"Leggings\",\"Colored_Trousers\",\"Slim_Short\",\"Shorty_Rotten\",\"SimpleSkirt\",\"DenimSkirt\",\"GoldtrimSkirt\",\"DesertDress\",\"Skirt\",\"Frilly_Skirt\",\"Crinkled_Skirt\",\"Icecream_Skirt\",\"Bermuda_Rolled\",\"Long_Dress\",\"Shorty_Mossy\",\"DaisySkirt\",\"CatacombCrawler_Shorts\",\"FrostwardenSet_Skirt\",\"Scavenger_Pants\",\"HighSkirt_Popstar\",\"SkaterShorts_Chunky\",\"Voidbearer_Pants\",\"Short_Ample\",\"Forest_Guardian\",\"Pants_Arctic_Scout\",\"Pants_Void_Hero\",\"Hope_Of_Gaia_Skirt\",\"Skirt_Savanna\",\"Pants_Wasteland_Marauder\",\"Merchant_Pants\",\"BannerlordQuilted\"],\"shoes\":[\"BasicBoots\",\"ScavenverLeatherBoots\",\"Boots_Thick\",\"BasicSandals\",\"BasicShoes\",\"SnowBoots\",\"Arctic\",\"HeavyLeather\",\"ThickSandals\",\"Sneakers_Sneakers\",\"HiBoots\",\"AdventurerBoots\",\"BannerlordBoots\",\"DesertBoots\",\"SlipOns\",\"MinerBoots\",\"Wellies\",\"Trainers\",\"SantaBoots\",\"ElfBoots\",\"GoldenBangle\",\"Boots_Long\",\"LeatherBoots\",\"Gem_Shoes\",\"FashionableBoots\",\"Icecream_Shoes\",\"BasicShoes_Shiny\",\"BasicShoes_Buckle\",\"BasicShoes_Strap\",\"BasicShoes_Sandals\",\"Boots_Voyager\",\"Hope_Of_Gaia_Boots\",\"DaisyShoes\",\"CatacombCrawler_Boots\",\"FrostwardenSet_Boots\",\"Arctic_Scout_Boots\",\"HeeledBoots_Savanna\",\"HeeledBoots_Popstar\",\"Scavenger_HeeledBoots\",\"Slipons_CoolGaia\",\"Voidbearer_Boots\",\"Shoes_Ornated\",\"Forest_Guardian_Boots\",\"Boots_Void_Hero\",\"Sneakers_Wasteland_Marauder\",\"Boots_Blazen_Wizard\",\"Merchant_Boots\",\"Battleworn_Boots\"],\"skinFeature\":[],\"undertop\":[\"SurvivorShirtBoy\",\"Wide_Neck_Shirt\",\"VNeck_Shirt\",\"Belt_Shirt\",\"Short_Sleeves_Shirt\",\"LongSleeveShirt\",\"VikingShirt\",\"LongSleeveShirt_GoldTrim\",\"LongSleeveShirt_ButtonUp\",\"HeartCamisole\",\"DoubleShirt\",\"DipCut\",\"Tshirt_Logo\",\"ColouredSleeves\",\"SmartShirt\",\"RibbedLongShirt\",\"StripedLong\",\"Undertops_Tubetop\",\"SpaghettiStrap\",\"ColouredStripes\",\"TieShirt\",\"FarmerTop\",\"LongSleevePeasantTop\",\"PaintSpillShirt\",\"FlowerShirt\",\"PastelFade\",\"PastelTracksuit\",\"CostumeShirt\",\"School_Shirt\",\"Frilly_Shirt\",\"School_Ribbon_Shirt\",\"School_Blazer_Shirt\",\"Crinkled_Top\",\"Flowy_Shirt\",\"Stylish_Belt_Shirt\",\"Amazon_Top\",\"Mercenary_Top\",\"Forest_Guardian_LongShirt\",\"CatacombCrawler_Undertop\",\"FrostwardenSet_Top\",\"Voidbearer_CursedArm\",\"Top_Wasteland_Marauder\",\"Bannerlord_Chainmail\"],\"underwear\":[\"Suit\",\"Bandeau\",\"Boxer\",\"Bra\"]}";

var wSkin = "{\"bodyCharacteristic\":\"Default.11\",\"underwear\":\"Bra.Blue\",\"face\":\"Face_Neutral\",\"ears\":\"Ogre_Ears\",\"mouth\":\"Mouth_Makeup\",\"haircut\":\"SideBuns.Black\",\"facialHair\":null,\"eyebrows\":\"RoundThin.Black\",\"eyes\":\"Plain_Eyes.Green\",\"pants\":\"Icecream_Skirt.Strawberry\",\"overpants\":\"LongSocks_Bow.Lime\",\"undertop\":\"VNeck_Shirt.Black\",\"overtop\":\"NeckHigh_Savanna.Pink\",\"shoes\":\"Wellies.Orange\",\"headAccessory\":null,\"faceAccessory\":null,\"earAccessory\":null,\"skinFeature\":null,\"gloves\":null,\"cape\":null}"

const SERVER_PROTOCOL =  "http://"
const SERVER_URI = "127.0.0.1:59313"

var wPublic, wPrivate, _ = ed25519.GenerateKey(rand.Reader);

func getSkinJsonPath() string {
	return filepath.Join(ServerDataFolder(), "skin.json");
}

func readSkinData() {
	load := getSkinJsonPath();
	os.MkdirAll(filepath.Dir(load), 0666);

	_, err := os.Stat(load);
	if err != nil {
		return;
	}
	skinData, _ := os.ReadFile(load);
	wSkin = string(skinData);

}

func writeSkinData(newData string) {
	save := getSkinJsonPath();
	os.MkdirAll(filepath.Dir(save), 0666);
	fmt.Printf("Writing skin data %s\n", save);


	os.WriteFile(save, []byte(newData), 0666);
	wSkin = newData;
}

func readCosmeticsIdFromAssets(zf *zip.ReadCloser, zpath string ) []string {
	defs := []cosmeticDefinition{};
	fmt.Printf("Opening: %s\n", zpath);

	f, err := zf.Open(zpath);
	if err != nil{
		fmt.Printf("err: %s\n", err);
		panic("failed to open cosmetic json file!");
	}
	defer f.Close();


	err = json.NewDecoder(f).Decode(&defs);
	if err != nil {
		panic("Failed to decode cosmetic json file!");
	}

	ids := []string {};
	for _, def := range defs {
		ids = append(ids, def.Id);
	}

	return ids;
}

func readCosmetics() string {

	// get currently installed gane folder ...

	patchline := valToChannel(int(wCommune.Patchline));
	gotVersion := int(wCommune.SelectedVersion+1);

	assetsZip := filepath.Join(getVersionInstallPath(gotVersion, patchline), "Assets.zip" );

	zf, err := zip.OpenReader(assetsZip);
	if err != nil {
		fmt.Printf("err: %s\n", err);
		return DEFAULT_COSMETICS;
	}
	defer zf.Close();

	ccFolder := path.Join("Cosmetics", "CharacterCreator");
	inventory := cosmeticsInventory{};

	inventory.BodyCharacteristic = readCosmeticsIdFromAssets(zf, path.Join(ccFolder, "BodyCharacteristics.json"));
	inventory.Cape = readCosmeticsIdFromAssets(zf, path.Join(ccFolder, "Capes.json"));
	inventory.EarAccessory = readCosmeticsIdFromAssets(zf, path.Join(ccFolder, "EarAccessory.json"));
	inventory.Ears = readCosmeticsIdFromAssets(zf, path.Join(ccFolder, "Ears.json"));
	inventory.Eyebrows = readCosmeticsIdFromAssets(zf, path.Join(ccFolder, "Eyebrows.json"));
	inventory.Eyes = readCosmeticsIdFromAssets(zf, path.Join(ccFolder, "Eyes.json"));
	inventory.Face = readCosmeticsIdFromAssets(zf, path.Join(ccFolder, "Faces.json"));
	inventory.FaceAccessory = readCosmeticsIdFromAssets(zf, path.Join(ccFolder, "FaceAccessory.json"));
	inventory.FacialHair = readCosmeticsIdFromAssets(zf, path.Join(ccFolder, "FacialHair.json"));
	inventory.Gloves = readCosmeticsIdFromAssets(zf, path.Join(ccFolder, "Gloves.json"));
	inventory.Haircut = readCosmeticsIdFromAssets(zf, path.Join(ccFolder, "Haircuts.json"));
	inventory.HeadAccessory = readCosmeticsIdFromAssets(zf, path.Join(ccFolder, "HeadAccessory.json"));
	inventory.Mouth = readCosmeticsIdFromAssets(zf, path.Join(ccFolder, "Mouths.json"));
	inventory.Overpants = readCosmeticsIdFromAssets(zf, path.Join(ccFolder, "Overpants.json"));
	inventory.Overtop = readCosmeticsIdFromAssets(zf, path.Join(ccFolder, "Overtops.json"));
	inventory.Pants = readCosmeticsIdFromAssets(zf, path.Join(ccFolder, "Pants.json"));
	inventory.Shoes = readCosmeticsIdFromAssets(zf, path.Join(ccFolder, "Shoes.json"));
	inventory.SkinFeature = readCosmeticsIdFromAssets(zf, path.Join(ccFolder, "SkinFeatures.json"));
	inventory.Undertop = readCosmeticsIdFromAssets(zf, path.Join(ccFolder, "Undertops.json"));
	inventory.Underwear = readCosmeticsIdFromAssets(zf, path.Join(ccFolder, "Underwear.json"));

	cosmeticsJson, err := json.Marshal(inventory);

	if err != nil {
		fmt.Printf("err: %s\n", err);
		return DEFAULT_COSMETICS;
	}

	fmt.Printf("Full Cosmetics List: %s\n", cosmeticsJson);

	return string(cosmeticsJson);
}

func genAccountInfo() accountInfo {
	readSkinData();
	return accountInfo{
		Username: wCommune.Username,
		UUID: getUUID(),
		Entitlements: ENTITLEMENTS,
		CreatedAt: time.Now(),
		NextNameChangeAt: time.Now(),
		Skin: wSkin,
	};
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
			json.NewEncoder(w).Encode(genAccountInfo());
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
				genAccountInfo(),
			},
		}
		w.Header().Add("Content-Type", "application/json");
		w.WriteHeader(200);
		json.NewEncoder(w).Encode(data);
	}
}


func handleSessionChild(w http.ResponseWriter, req *http.Request) {

	sessionRequest := sessionChild{};
	json.NewDecoder(req.Body).Decode(&sessionRequest);

	session := sessionNew{
		ExpiresAt: time.Now().Add(time.Hour*10),
		IdentityToken: generateIdentityJwt(sessionRequest.Scopes),
		SessionToken: generateSessionJwt(sessionRequest.Scopes),
	}

	w.Header().Add("Content-Type", "application/json");
	w.WriteHeader(200);
	json.NewEncoder(w).Encode(session);

}

func handleBugReport(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(204);
}

func handleFeedbacksReport(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(204);
}

func handleManifest(w http.ResponseWriter, req *http.Request) {

	target := req.PathValue("target")
	arch := req.PathValue("arch")
	branch := req.PathValue("branch")
	patch := req.PathValue("patch")
	fmt.Printf("target: %s\narch: %s\nbranch: %s\npatch: %s\n", target, arch, branch, patch);

	p := filepath.Join("patches", target, arch, branch, patch, "manifest.json");

	http.ServeFile(w, req, p);

}

func handlePatches(w http.ResponseWriter, req *http.Request) {

	fp := req.PathValue("filepath");
	p := filepath.Join("patches", fp);

	_, err := os.Stat(p);
	if err != nil {
		return;
	}

	http.ServeFile(w, req, p);
}

func handleJwksRequest(w http.ResponseWriter, req *http.Request) {

	keys := jwkKeyList{
		Keys: []jwkKey {
			{
				Alg: "EdDSA",
				Crv: "Ed25519",
				Kid: "2025-10-01",
				Kty: "OKP",
				Use: "sig",
				X: base64.RawURLEncoding.EncodeToString([]byte(wPublic)),
			},
		},
	};

	w.Header().Add("Content-Type", "application/json");
	w.WriteHeader(200);

	json.NewEncoder(w).Encode(keys);
}

func handleTelemetryRequest(w http.ResponseWriter, req *http.Request) {
	// send telemetry to /dev/null ..
	w.WriteHeader(204);
}


func logRequestHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("> %s %s\n", r.Method,  r.URL);
		h.ServeHTTP(w, r)
	})
}



func runServer() {

	mux := http.NewServeMux();

	// account-data.hytale.com
	mux.HandleFunc("/my-account/game-profile", handleMyAccountGameProfile);
	mux.HandleFunc("/my-account/skin", handleMyAccountSkin)
	mux.HandleFunc("/my-account/cosmetics", handleMyAccountCosmetics)
	mux.HandleFunc("/my-account/get-launcher-data", handleMyAccountLauncherData);

	mux.HandleFunc("/patches/{target}/{arch}/{branch}/{patch}", handleManifest);
	mux.HandleFunc("/patches/{filepath...}", handlePatches);

	// session.hytale.com
	mux.HandleFunc("/game-session/child", handleSessionChild);
	mux.HandleFunc("/.well-known/jwks.json", handleJwksRequest);

	// tools.hytale.com
	mux.HandleFunc("/bugs/create", handleBugReport);
	mux.HandleFunc("/feedback/create", handleFeedbacksReport);

	mux.HandleFunc("/telemetry/client", handleTelemetryRequest);


	var handler  http.Handler = mux;
	handler = logRequestHandler(handler);

	http.ListenAndServe(SERVER_URI, handler);
}

func sign(j string) string {
	sig := ed25519.Sign(wPrivate, []byte(j));
	return base64.RawURLEncoding.EncodeToString(sig);
}

func make_jwt(body any) string {
	head := jwtHeader{
		Alg: "EdDSA",
		Kid: "2025-10-01",
		Typ: "JWT",
	};

	jHead, _ := json.Marshal(head);
	jBody, _ := json.Marshal(body);


	jwt := base64.RawURLEncoding.EncodeToString(jHead) + "." + base64.RawURLEncoding.EncodeToString(jBody)
	jwt += "." + sign(jwt);
	return jwt;
}

func getUUID() string{
	r, err := regexp.MatchString("[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}", strings.ToLower(wCommune.UUID));
	if err != nil || r == false{
		m := md5.New();
		m.Write([]byte(wCommune.Username));
		h := hex.EncodeToString(m.Sum(nil));

		return h[:8]+"-"+h[8:12]+"-"+h[12:16]+"-"+h[16:20]+"-"+h[20:32];
	}
	return wCommune.UUID;
}

func generateSessionJwt(scope []string) string {


	sesTok := sessionToken {
		Exp: int(time.Now().Add(time.Hour*200).Unix()),
		Iat: int(time.Now().Unix()),
		Iss: SERVER_PROTOCOL + SERVER_URI,
		Jti: uuid.NewString(),
		Scope: strings.Join(scope, " "),
		Sub: getUUID(),
	};
	fmt.Printf("[JWT] Generating new session JWT with scopes: %s\n", sesTok.Scope);

	return make_jwt(sesTok);
}



func generateIdentityJwt(scope []string) string {

	idTok := identityToken {
		Exp: int(time.Now().Add(time.Hour*200).Unix()),
		Iat: int(time.Now().Unix()),
		Iss: SERVER_PROTOCOL + SERVER_URI,
		Jti: uuid.NewString(),
		Scope: strings.Join(scope, " "),
		Sub: getUUID(),
		Profile: profileInfo {
			Username: wCommune.Username,
			Entitlements: ENTITLEMENTS,
			Skin: wSkin,
		},
	};

	fmt.Printf("[JWT] Generating new identity JWT with scopes: %s\n", idTok.Scope);
	return make_jwt(idTok);
}
