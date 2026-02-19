package main

import (
	"archive/zip"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

var ENTITLEMENTS = []string {"game.base", "game.deluxe", "game.founder"};

const DEFAULT_COSMETICS = "{\"bodyCharacteristic\":[\"Default\",\"Muscular\"],\"cape\":[\"Cape_Royal_Emissary\",\"Cape_New_Beginning\",\"Cape_Forest_Guardian\",\"Cape_PopStar\",\"Cape_Scavenger\",\"Cape_Knight\",\"Cape_Seasons\",\"Hope_Of_Gaia_Cape\",\"Cape_Blazen_Wizard\",\"Cape_King\",\"Cape_Void_Hero\",\"Cape_Featherbound\",\"FrostwardenSet_Cape\",\"Cape_Bannerlord\",\"Cape_Wasteland_Marauder\"],\"earAccessory\":[\"EarHoops\",\"SimpleEarring\",\"DoubleEarrings\",\"SilverHoopsBead\",\"SpiralEarring\",\"AcornEarrings\"],\"ears\":[\"Default\",\"Elf_Ears\",\"Elf_Ears_Large\",\"Elf_Ears_Large_Down\",\"Elf_Ears_Small\",\"Ogre_Ears\"],\"eyebrows\":[\"Medium\",\"Thin\",\"Thick\",\"Bushy\",\"Shaved\",\"SmallRound\",\"Large\",\"RoundThin\",\"Angry\",\"Plucked\",\"Square\",\"Serious\",\"BushyThin\",\"Heavy\"],\"eyes\":[\"Medium_Eyes\",\"Large_Eyes\",\"Plain_Eyes\",\"Almond_Eyes\",\"Square_Eyes\",\"Reptile_Eyes\",\"Cat_Eyes\",\"Demonic_Eyes\",\"Goat_Eyes\"],\"face\":[\"Face_Neutral\",\"Face_Neutral_Freckles\",\"Face_Sunken\",\"Face_Tired_Eyes\",\"Face_Stubble\",\"Face_Scar\",\"Face_Aged\",\"Face_Older2\",\"Face_Almond_Eyes\",\"Face_MakeUp\",\"Face_Make_Up_2\",\"Face_MakeUp_Freckles\",\"Face_MakeUp_Highlight\",\"Face_MakeUp_6\",\"Face_MakeUp_Older\",\"Face_MakeUp_Older2\"],\"faceAccessory\":[\"EyePatch\",\"Glasses\",\"LargeGlasses\",\"MedicalEyePatch\",\"MouthCover\",\"MouthWheat\",\"ColouredGlasses\",\"CrazyGlasses\",\"RoundGlasses\",\"HeartGlasses\",\"AgentGlasses\",\"SunGlasses\",\"AviatorGlasses\",\"BusinessGlasses\",\"Plaster\",\"Glasses_Monocle\",\"GlassesTiny\",\"Goggles_Wasteland_Marauder\"],\"facialHair\":[\"Medium\",\"Beard_Large\",\"Goatee\",\"Chin_Curtain\",\"Moustache\",\"VikingBeard\",\"TwirlyMoustache\",\"SoulPatch\",\"PirateBeard\",\"TripleBraid\",\"DoubleBraid\",\"GoateeLong\",\"PirateGoatee\",\"Soldier\",\"Hip\",\"Trimmed\",\"Handlebar\",\"Groomed\",\"Stylish\",\"ThinGoatee\",\"Short_Trimmed\",\"Groomed_Large\",\"WavyLongBeard\",\"CurlyLongBeard\"],\"gloves\":[\"BasicGloves_Basic\",\"BoxingGloves\",\"FlowerBracer\",\"MiningGloves\",\"GoldenBracelets\",\"LeatherMittens\",\"Straps_Leather\",\"Shackles_Feran\",\"CatacombCrawler_Gloves\",\"Hope_Of_Gaia_Gloves\",\"Gloves_Void_Hero\",\"LongGloves_Popstar\",\"Gloves_Medium_Featherbound\",\"Arctic_Scout_Gloves\",\"Scavenger_Gloves\",\"Bracer_Daisy\",\"LongGloves_Savanna\",\"Gloves_Wasteland_Marauder\",\"Gloves_Blazen_Wizard\",\"Merchant_Gloves\",\"Battleworn_Gloves\"],\"haircut\":[\"Morning\",\"Bangs\",\"Quiff\",\"Lazy\",\"BobCut\",\"Messy\",\"Viking\",\"Fringe\",\"PonyTail\",\"Bun\",\"Braid\",\"BraidDouble\",\"ShortDreads\",\"Undercut\",\"Samurai\",\"DoublePart\",\"Rustic\",\"RoseBun\",\"SideBuns\",\"SmallPigtails\",\"Stylish\",\"Mohawk\",\"BowlCut\",\"Emo\",\"Pigtails\",\"Sideslick\",\"SingleSidePigtail\",\"Slickback\",\"WavyPonytail\",\"Wings\",\"ChopsticksPonyTail\",\"Curly\",\"MessyBobcut\",\"Simple\",\"WidePonytail\",\"RaiderMohawk\",\"MidSinglePart\",\"AfroPuffs\",\"PuffyQuiff\",\"GenericPuffy\",\"PuffyPonytail\",\"FighterBuns\",\"MaleElf\",\"Windswept\",\"SidePonytail\",\"PonyBuns\",\"ElfBackBun\",\"BraidedPonytail\",\"ThickBraid\",\"WavyBraids\",\"VikinManBun\",\"Witch\",\"FrizzyLong\",\"WavyLong\",\"SuperSlickback\",\"Cat\",\"Scavenger_Hair\",\"LongTied\",\"LongBangs\",\"BantuKnot\",\"Berserker\",\"CuteEmoBangs\",\"CutePart\",\"LongPigtails\",\"FeatheredHair\",\"LongHairPigtail\",\"StraightHairBun\",\"SuperSideSlick\",\"FrontTied\",\"EmoWavy\",\"MessyMop\",\"EmoBangs\",\"BowHair\",\"Greaser\",\"FrontFlick\",\"Long\",\"WavyShort\",\"GenericLong\",\"GenericMedium\",\"GenericShort\",\"CurlyShort\",\"LongCurly\",\"MorningLong\",\"CentrePart\",\"VikingWarrior\",\"MediumCurly\",\"SpikedUp\",\"Cowlick\",\"MessyWavy\",\"BuzzCut\",\"QuiffLeft\",\"StylishWindswept\",\"SuperShirt\",\"StylishQuiff\",\"BangsShavedBack\",\"FrizzyVolume\",\"Cornrows\",\"Balding\",\"Dreadlocks\"],\"headAccessory\":[\"Goggles\",\"Hoodie\",\"GiHeadband\",\"ForeheadProtector\",\"FlowerCrown\",\"Bandana\",\"FloppyBeanie\",\"BunnyBeanie\",\"Headband\",\"CatBeanie\",\"FrogBeanie\",\"WorkoutCap\",\"HeadDaliah\",\"HairRose\",\"HairPeony\",\"HairDaisy\",\"Logo_Cap\",\"BanjoHat\",\"WitchHat\",\"StrawHat\",\"PirateBandana\",\"HairHibiscus\",\"SantaHat\",\"ElfHat\",\"Head_Crown\",\"HeadphonesDadCap\",\"Headphones\",\"Beanie\",\"BandanaSkull\",\"StripedBeanie\",\"Head_Tiara\",\"Viking_Helmet\",\"Pirate_Captain_Hat\",\"TopHat\",\"CowboyHat\",\"RusticBeanie\",\"LeatherCap\",\"Ribbon\",\"Bunny_Ears\",\"Head_Bandage\",\"AcornNecktie\",\"AcornHairclip\",\"Forest_Guardian_Hat\",\"Hoodie_Feran\",\"ExplorerGoggles\",\"Hope_Of_Gaia_Crown\",\"FrostwardenSet_Hat\",\"Arctic_Scout_Hat\",\"Savanna_Scout_Hat\",\"BulkyBeanie\",\"Hat_Popstar\",\"ShapedCap_Chill\",\"Hoodie_Ornated\",\"Headband_Void_Hero\",\"Hood_Blazen_Wizard\",\"Merchant_Beret\",\"Battleworn_Helm\"],\"mouth\":[\"Mouth_Default\",\"Mouth_Makeup\",\"Mouth_Thin\",\"Mouth_Long\",\"Mouth_Tiny\"],\"overpants\":[\"KneePads\",\"LongSocks_Plain\",\"LongSocks_BasicWrap\",\"LongSocks_School\",\"LongSocks_Striped\",\"LongSocks_Bow\",\"LongSocks_Torn\"],\"overtop\":[\"PuffyJacket\",\"Tartan\",\"BunnyHoody\",\"StylishJacket\",\"LongBeltedJacket\",\"RobeOvertops\",\"HeroShirt\",\"ThreadedOvertops\",\"RaggedVest\",\"Winter_Jacket\",\"Suit_Jacket\",\"Wool_Jersey\",\"Chest_PuffyJersey\",\"Tunic_Weathered\",\"JacketShort\",\"JacketLong\",\"Coat\",\"TrenchCoat\",\"VikingVest\",\"GiShirt\",\"ShortTartan\",\"BulkyShirtLong\",\"MiniLeather\",\"Fantasy\",\"Pirate\",\"BulkyShirt_Scarf\",\"Scarf_Large_Stripped\",\"Scarf_Large\",\"BulkyShirtLong_LeatherJacket\",\"ForestVest\",\"BulkyShirt_StomachWrap\",\"FantasyShawl\",\"LeatherVest\",\"BulkyShirt_RoyalRobe\",\"Jinbaori\",\"Ronin\",\"MessyShirt\",\"StitchedShirt\",\"OpenShirtBand\",\"BulkyShirt_RuralShirt\",\"BulkyShirt_RuralPattern\",\"HeartNecklace\",\"Shark_Tooth_Necklace\",\"Pookah_Necklace\",\"Golden_Bangles\",\"BulkyShirt_FancyWaistcoat\",\"LetterJacket\",\"PinstripeJacket\",\"DoubleButtonJacket\",\"Polarneck\",\"FlowyHalf\",\"FurLinedJacket\",\"PlainHoodie\",\"LooseSweater\",\"SimpleDress\",\"SleevedDress\",\"SleevedDresswJersey\",\"Tunic_Long\",\"Scarf\",\"TracksuitJacket\",\"Jacket\",\"KhakiShirt\",\"LongCardigan\",\"GoldtrimJacket\",\"Cheststrap\",\"SantaJacket\",\"ElfJacket\",\"FarmerVest\",\"AviatorJacket\",\"QuiltedTop\",\"Jinbaori_Wave\",\"Jinbaori_Flower\",\"FloppyBunnyJersey\",\"PlainJersey\",\"Tunic_Villager\",\"RoughFabricBand\",\"Arm_Bandage\",\"Farmer_Dress\",\"OnePiece_SchoolDress\",\"OnePiece_ApronDress\",\"Noble_Beige\",\"Fancy_Coat\",\"Adventurer_Dress\",\"Oasis_Dress\",\"PuffyBomber\",\"Jacket_Voyager\",\"AlpineExplorerJumper\",\"Hope_Of_GaiaOvertop\",\"DaisyTop\",\"Arctic_Scout_Jacket\",\"Collared_Cool\",\"NeckHigh_Savanna\",\"Scavenger_Poncho\",\"NeckHigh_LeatherClad\",\"Jacket_Popstar\",\"Voidbearer_Top\",\"Featherbound_Tunic\",\"Forest_Guardian_Poncho\",\"Jacket_Void_Hero\",\"Straps_Wasteland_Marauder\",\"Robe_Blazen_Wizard\",\"Merchant_Tunic\",\"Battleworn_Tunic\",\"Bannerlord_Tunic\"],\"pants\":[\"ApprenticePants\",\"LeatherPants\",\"SurvivorPants\",\"StripedPants\",\"CostumePants\",\"ShortyRolled\",\"Jeans\",\"GiPants\",\"Forest_Bermuda\",\"BulkySuede\",\"Pants_Straight_WreckedJeans\",\"Pants_Slim\",\"Dungarees\",\"StylishShorts\",\"JeansStrapped\",\"Villager_Bermuda\",\"ExplorerShorts\",\"Explorer_Trousers\",\"PinstripeTrousers\",\"Pants_Slim_Faded\",\"Pants_Slim_Tracksuit\",\"LongDungarees\",\"KhakiShorts\",\"ColouredKhaki\",\"Leggings\",\"Colored_Trousers\",\"Slim_Short\",\"Shorty_Rotten\",\"SimpleSkirt\",\"DenimSkirt\",\"GoldtrimSkirt\",\"DesertDress\",\"Skirt\",\"Frilly_Skirt\",\"Crinkled_Skirt\",\"Icecream_Skirt\",\"Bermuda_Rolled\",\"Long_Dress\",\"Shorty_Mossy\",\"DaisySkirt\",\"CatacombCrawler_Shorts\",\"FrostwardenSet_Skirt\",\"Scavenger_Pants\",\"HighSkirt_Popstar\",\"SkaterShorts_Chunky\",\"Voidbearer_Pants\",\"Short_Ample\",\"Forest_Guardian\",\"Pants_Arctic_Scout\",\"Pants_Void_Hero\",\"Hope_Of_Gaia_Skirt\",\"Skirt_Savanna\",\"Pants_Wasteland_Marauder\",\"Merchant_Pants\",\"BannerlordQuilted\"],\"shoes\":[\"BasicBoots\",\"ScavenverLeatherBoots\",\"Boots_Thick\",\"BasicSandals\",\"BasicShoes\",\"SnowBoots\",\"Arctic\",\"HeavyLeather\",\"ThickSandals\",\"Sneakers_Sneakers\",\"HiBoots\",\"AdventurerBoots\",\"BannerlordBoots\",\"DesertBoots\",\"SlipOns\",\"MinerBoots\",\"Wellies\",\"Trainers\",\"SantaBoots\",\"ElfBoots\",\"GoldenBangle\",\"Boots_Long\",\"LeatherBoots\",\"Gem_Shoes\",\"FashionableBoots\",\"Icecream_Shoes\",\"BasicShoes_Shiny\",\"BasicShoes_Buckle\",\"BasicShoes_Strap\",\"BasicShoes_Sandals\",\"Boots_Voyager\",\"Hope_Of_Gaia_Boots\",\"DaisyShoes\",\"CatacombCrawler_Boots\",\"FrostwardenSet_Boots\",\"Arctic_Scout_Boots\",\"HeeledBoots_Savanna\",\"HeeledBoots_Popstar\",\"Scavenger_HeeledBoots\",\"Slipons_CoolGaia\",\"Voidbearer_Boots\",\"Shoes_Ornated\",\"Forest_Guardian_Boots\",\"Boots_Void_Hero\",\"Sneakers_Wasteland_Marauder\",\"Boots_Blazen_Wizard\",\"Merchant_Boots\",\"Battleworn_Boots\"],\"skinFeature\":[],\"undertop\":[\"SurvivorShirtBoy\",\"Wide_Neck_Shirt\",\"VNeck_Shirt\",\"Belt_Shirt\",\"Short_Sleeves_Shirt\",\"LongSleeveShirt\",\"VikingShirt\",\"LongSleeveShirt_GoldTrim\",\"LongSleeveShirt_ButtonUp\",\"HeartCamisole\",\"DoubleShirt\",\"DipCut\",\"Tshirt_Logo\",\"ColouredSleeves\",\"SmartShirt\",\"RibbedLongShirt\",\"StripedLong\",\"Undertops_Tubetop\",\"SpaghettiStrap\",\"ColouredStripes\",\"TieShirt\",\"FarmerTop\",\"LongSleevePeasantTop\",\"PaintSpillShirt\",\"FlowerShirt\",\"PastelFade\",\"PastelTracksuit\",\"CostumeShirt\",\"School_Shirt\",\"Frilly_Shirt\",\"School_Ribbon_Shirt\",\"School_Blazer_Shirt\",\"Crinkled_Top\",\"Flowy_Shirt\",\"Stylish_Belt_Shirt\",\"Amazon_Top\",\"Mercenary_Top\",\"Forest_Guardian_LongShirt\",\"CatacombCrawler_Undertop\",\"FrostwardenSet_Top\",\"Voidbearer_CursedArm\",\"Top_Wasteland_Marauder\",\"Bannerlord_Chainmail\"],\"underwear\":[\"Suit\",\"Bandeau\",\"Boxer\",\"Bra\"]}";


const DEFAULT_SKIN = "{\"bodyCharacteristic\":\"Default.11\",\"underwear\":\"Bra.Blue\",\"face\":\"Face_Neutral\",\"ears\":\"Ogre_Ears\",\"mouth\":\"Mouth_Makeup\",\"haircut\":\"SideBuns.Black\",\"facialHair\":null,\"eyebrows\":\"RoundThin.Black\",\"eyes\":\"Plain_Eyes.Green\",\"pants\":\"Icecream_Skirt.Strawberry\",\"overpants\":\"LongSocks_Bow.Lime\",\"undertop\":\"VNeck_Shirt.Black\",\"overtop\":\"NeckHigh_Savanna.Pink\",\"shoes\":\"Wellies.Orange\",\"headAccessory\":null,\"faceAccessory\":null,\"earAccessory\":null,\"skinFeature\":null,\"gloves\":null,\"cape\":null}";

var SERVER_PROTOCOL =  "http://"
var SERVER_HOST = "127.0.0.1"
var SERVER_PORT = (1028) + rand.IntN(65535-1028);

const CURRENT_FMT_VERSION = 1;


var (
	sSkin = skinList{};
	sSkinLoaded = false;
)

func getSentryUrl() string {
	return SERVER_PROTOCOL + "transrights" + "@" + getServerHostPort() + "/2";
}

func getServerHostPort() string {
	return SERVER_HOST + ":" + strconv.Itoa(SERVER_PORT);
}

func getServerUrl() string {
	return SERVER_PROTOCOL + getServerHostPort();
}

func getOldSkinJsonPath() string {
	return filepath.Join(ServerDataFolder(), "skin.json");
}

func getNewSkinJsonPath() string {
	return filepath.Join(ServerDataFolder(), getUUID()+"_skin.json");
}

func readOldSkinData() string {
	load := getOldSkinJsonPath();
	os.MkdirAll(filepath.Dir(load), 0666);

	_, err := os.Stat(load);
	if err != nil {
		fmt.Printf("[Server] Failed to read old skinfile, falling back on default\n");
		return DEFAULT_SKIN;
	}
	skinData, _ := os.ReadFile(load);
	return string(skinData);
}


func writeSkinData() {
	save := getNewSkinJsonPath();
	os.MkdirAll(filepath.Dir(save), 0666);
	fmt.Printf("[Server] Writing skin data %s\n", save);

	newSkinData, err := json.Marshal(sSkin);
	if  err != nil {
		panic("failed to encode skin data!");
	}

	os.WriteFile(save, []byte(newSkinData), 0666);
}

func readSkinData() {
	if sSkinLoaded == true {
		return;
	}

	load := getNewSkinJsonPath();

	data, err := os.ReadFile(load);
	if err != nil {
		// first time?? write default skin ..
		initSkinData(DEFAULT_SKIN);
		sSkinLoaded = true;
		writeSkinData();
		return;
	} else {
		err = json.Unmarshal(data, &sSkin);

		if err != nil {
			panic("failed to decode skin data.");
		}

		sSkinLoaded = true;
	}

}

func readCosmeticsIdFromAssets(zf *zip.ReadCloser, zpath string ) []string {
	defs := []cosmeticDefinition{};
	fmt.Printf("[Server] Opening asesets file: %s\n", zpath);

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
		fmt.Printf("[Server] failed to read assets zip: %s, falling back on default cosmetics\n", err);
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
		fmt.Printf("[Server] failed to read cosmetics from assets, falling back on default: %s\n", err);
		return DEFAULT_COSMETICS;
	}

	fmt.Printf("[Server] Read Cosmetics List.\n");

	return string(cosmeticsJson);
}


func updateSkinDefinition(skinId string, skinData skinDefinition) {
	// ngl this is kinda bullshit, surely theres another way?

	for i, skin := range sSkin.Skins {
		if skin.ID == skinId {
			sSkin.Skins[i] = skinData;
		}
	}

	writeSkinData();
}

func updateSkin(skinId string, skinData string) {

	for i, skin := range sSkin.Skins {
		if skin.ID == skinId {
			sSkin.Skins[i].SkinData = skinData;
		}
	}

	writeSkinData();
}

func getSkinByName(name string) string {
	readSkinData();

	for _, skin := range sSkin.Skins {
		if skin.Name == name {
			return skin.SkinData;
		}
	}

	return DEFAULT_SKIN;
}
func getActiveSkin() string {
	readSkinData();

	activeSkin := sSkin.ActiveSkin;

	for _, skin := range sSkin.Skins {
		if skin.ID == activeSkin {
			return skin.SkinData;
		}
	}

	return DEFAULT_SKIN;
}

func delSkin(skinId string) {

	for i, skin := range sSkin.Skins {
		if skin.ID == skinId {
			sSkin.Skins[i] = sSkin.Skins[len(sSkin.Skins)-1];
			sSkin.Skins = sSkin.Skins[:len(sSkin.Skins)-1];

			// set active skin to the skin just before this one if this one is selected.
			if sSkin.ActiveSkin == skin.ID {
				sSkin.ActiveSkin = sSkin.Skins[len(sSkin.Skins)-1].ID;
			}

			writeSkinData();
			return;
		}
	}

}



func genAccountInfo() accountInfo {
	readSkinData();
	return accountInfo{
		Username: wCommune.Username,
		UUID: getUUID(),
		Entitlements: ENTITLEMENTS,
		CreatedAt: time.Now(),
		NextNameChangeAt: time.Now(),
		Skin: getActiveSkin(),
	};
}



func initSkinData(skinData string) {
	sSkin = skinList{};
	skinUuid := uuid.NewString();

	sSkin.ActiveSkin = skinUuid;
	sSkin.MaxSkins = int(wCommune.MaxSkins);

	skinDef := skinDefinition{
		ID: skinUuid,
		Name: "Default",
		SkinData: skinData,
	};

	sSkin.Skins = append(sSkin.Skins, skinDef);

	writeSkinData();
}

func migrateSkinData() {
	fmt.Printf("[Server] Migrating skin data ...\n");

	oldSkinData := readOldSkinData();
	initSkinData(oldSkinData);

	// remove old skin.json file
	os.Remove(getOldSkinJsonPath());

}

func handleMyAccountSkin(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
		case "PUT":
			data, _ := io.ReadAll(req.Body);
			updateSkin(sSkin.ActiveSkin, string(data));
			w.WriteHeader(204);

	}
}

func handleNewSkin(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
		case "GET":
			w.Header().Add("Content-Type", "application/json");
			w.WriteHeader(200);

			// set max skins :
			sSkin.MaxSkins = int(wCommune.MaxSkins);

			json.NewEncoder(w).Encode(sSkin);
		case "POST":
			def := skinDefinition{};
			json.NewDecoder(req.Body).Decode(&def);

			// create new skin entry ..
			def.ID = uuid.NewString();
			sSkin.Skins = append(sSkin.Skins, def);
			sSkin.ActiveSkin = def.ID;

			ident := skinIdentifier{ SkinID: def.ID };

			writeSkinData();

			w.WriteHeader(200);
			json.NewEncoder(w).Encode(ident);
	}
}

func handleNewSkinActive(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
		case "PUT":
			ident := skinIdentifier{};
			json.NewDecoder(req.Body).Decode(&ident);

			// set new skin to skin id
			sSkin.ActiveSkin = ident.SkinID;
			writeSkinData();

			w.WriteHeader(204);
	}
}



func handleNewSkinLookup(w http.ResponseWriter, req *http.Request) {
	skinUuid := req.PathValue("uuid")

	switch req.Method {
		case "PUT":
			def := skinDefinition{};
			json.NewDecoder(req.Body).Decode(&def);

			def.ID = skinUuid;
			updateSkinDefinition(skinUuid, def)

			w.WriteHeader(204);
		case "DELETE":
			delSkin(skinUuid);

			writeSkinData();

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


func handleSession(w http.ResponseWriter, req *http.Request) {
	switch(req.Method) {
		case "DELETE":
			w.WriteHeader(204);
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


func handleJwksRequest(w http.ResponseWriter, req *http.Request) {
	fmt.Printf("[JWT] private key: %s\n", hex.EncodeToString(jwtPrivate));
	fmt.Printf("[JWT] public key: %s\n", hex.EncodeToString(jwtPublic));

	keys := jwkKeyList{
		Keys: []jwkKey {
			{
				Alg: "EdDSA",
				Crv: "Ed25519",
				Kid: "2025-10-01",
				Kty: "OKP",
				Use: "sig",
				X: base64.RawURLEncoding.EncodeToString([]byte(jwtPublic)),
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

func handleSentryRequest(w http.ResponseWriter, req *http.Request) {
	// send other telemetry to /dev/null ..

	w.Header().Add("Content-Type", "application/json");
	w.WriteHeader(200);

	w.Write([]byte("{}"));
}

func logRequestHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("[Server] [%s] %s\n", r.Method,  r.URL);
		h.ServeHTTP(w, r)
	})
}


func dataFixerUpper(oldVersion int) {
	if oldVersion <= 0 {
		migrateSkinData();
	}

	wCommune.FormatVersion = CURRENT_FMT_VERSION;
	writeSettings();
}

func runServer() {

	reloadSkin();

	mux := http.NewServeMux();

	// account-data.hytale.com
	mux.HandleFunc("/my-account/game-profile", handleMyAccountGameProfile);
	mux.HandleFunc("/my-account/skin", handleMyAccountSkin)
	mux.HandleFunc("/my-account/cosmetics", handleMyAccountCosmetics)
	mux.HandleFunc("/my-account/get-launcher-data", handleMyAccountLauncherData);

	// new skins
	mux.HandleFunc("/player-skins", handleNewSkin);
	mux.HandleFunc("/player-skins/active", handleNewSkinActive);
	mux.HandleFunc("/player-skins/{uuid}", handleNewSkinLookup);

	// session.hytale.com
	mux.HandleFunc("/game-session", handleSession);
	mux.HandleFunc("/game-session/child", handleSessionChild);
	mux.HandleFunc("/.well-known/jwks.json", handleJwksRequest);

	// tools.hytale.com
	mux.HandleFunc("/bugs/create", handleBugReport);
	mux.HandleFunc("/feedback/create", handleFeedbacksReport);

	// telemetry.hytale.com
	mux.HandleFunc("/telemetry/client", handleTelemetryRequest);

	// sentry.hytale.com
	mux.HandleFunc("/api/2/envelope", handleSentryRequest);


	// game-patches.hytale.com ..
	//mux.HandleFunc("/patches/{target}/{arch}/{branch}/{patch}", handleManifest);
	//mux.HandleFunc("/patches/{filepath...}", handlePatches);

	var handler  http.Handler = mux;
	handler = logRequestHandler(handler);

	err := http.ListenAndServe(getServerHostPort(), handler);
	if err != nil {
		fmt.Printf("[Server] error starting server: %s is the server already running?\n", err);
	} else {
		fmt.Printf("[Server] Starting server on $s\n", getServerHostPort());
	}
}



func reloadSkin() {
	sSkinLoaded = false;
	readSkinData();
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
		Iss: getServerUrl(),
		Jti: uuid.NewString(),
		Scope: strings.Join(scope, " "),
		Sub: getUUID(),
	};
	fmt.Printf("[JWT] Generating new session JWT with scopes: %s\n", sesTok.Scope);

	return makeJwt(sesTok);
}

func generateIdentityJwt(scope []string) string {

	idTok := identityToken {
		Exp: int(time.Now().Add(time.Hour*200).Unix()),
		Iat: int(time.Now().Unix()),
		Iss: getServerUrl(),
		Jti: uuid.NewString(),
		Scope: strings.Join(scope, " "),
		Sub: getUUID(),
		Profile: profileInfo {
			Username: wCommune.Username,
			Entitlements: ENTITLEMENTS,
			Skin: getActiveSkin(),
		},
	};

	fmt.Printf("[JWT] Generating new identity JWT with scopes: %s\n", idTok.Scope);
	return makeJwt(idTok);
}
