#HytaleSP

An alternative launcher for "Hytale" with a fairly streightforward "native" UI
 features:
 
 - Multiple version management
 - Incremental patch from one version to another
 - Universal "online fix" or that works on all versions;
 - Auth server emulation is implemented locallyy 
 - Play local or online multiplayer
 - Completely standalone; a single executable- no butler or other external tools.
 - Fairly transparent online patch implementation.
 - Run offline and download without needing a hytale account.
 - The game can be played offline with all features no external auth server
 - Supports Windows 7+
 
 Currently for: Windows and Linux (i would do MacOS, but i dont have access to an ARM Mac right now.)

Design Philosophy:

	-> everything (except version downloads) is run locally on-device, (including the auth server for "fakeonline")
	-> telemetry is disabled for the game itself
	-> the software should not make use of Electron or any heavy UI frameworks (Currently: GIU/imgui)
	-> no hard modifications to the EXE, all client modifications are done via injecting DLLs
	
NOTE: you can even probably play online, but only if the servers your joining have ``--auth-mode=insecure`` set in the command line;
enabling this option would also mean anyone using the offical launcher cannot play 
because it doesn't allow insecure auth type outside singleplayer offline mode; .. for some reason

also if you use the "online play" feature in singleplayer should also work as long as other users also use hytLauncher ...

## Project layout

hytServer.go - hytale authentication server emulator "online fix"

hytAuth.go - implementation of hytale OAuth2.0, currently not used

hytFormats.go - most of the JSON structures used by hytale are here;

hytClient.go - downloading versions, etc

hytPatch.go - itch.io's 'apply patch' wharf code

hytJwt.go - JSON structures used in hytale auth tokens

hytLocations.go - default / Location resolvers for many folders and directories used by the game

hytGui.go - Graphical user interface and general code to start the game

hytCli.go - Command line handling

Resources/ - resource scripts, icons, images, manifest files, etc;

Flatpak

Aurora/ - c code for the dll or shared object, loaded with the game, it replaces http://hytale.com with http://localhost/

# Building

on windows, you first have to build the "Aurora.dll" using MSVC, 
and then you can use ``go build .``

or you can ``build-windows.bat`` within the VS2026 developer command prompt to do this;

on linux, you need ``build-essential`` and then you can build "Aurora.so" using its Makefile;

after that you can use ``go build .`` 

or you can use ``build-linux.sh``

# Online Multiplayer 

When using the "fake online" option, you CAN play online multiplayer; 
BUT only if the server is setup specifically so you can play on it; i.e "cracked servers"

if your using local multiplayer or "game codes", with other users of HytaleSP;
then that should just work,

you cannot play with users using "Authenticated" mode, or the offical launcher!
for more information about this, see [SERVER.md](SERVER.md)

An example server you can try is; ``server.diamondbyte.org:5521``

# Alternative names
Originally, i called it "hytLauncher", being a play on "TLauncher" for minecraft;
however i found that, alot of other alternative hytale launchers had a simular name, (eg HyTaLauncher, HyLauncher, etc.)

furthermore tLauncher is a bit sketchy in general and there are better options for minecraft too;
for now i have settled on "HytaleSP" .. a reference to an extremely OG minecraft launcher;

i may consider other names in the future;

other names i considered using :

- hyTLauncher
- HytaleSP
- AnjoCaidos Hytale Launcher
- HytaleForFree.com 

in all seriousness i kind of would want something a bit original xS
also this is not nessecarily planned to be a purely 'offline mode' launcher; 

i also want to add "premium" support as well-
mm the code for authentication flow is actually already here ..

but i wont remove the 'offline' or 'fakeonline' options for those who need them though .. 

# Screenshots 

![HytaleSP ui itself](https://git.silica.codes/Li/HytaleSP/raw/branch/main/images/screenshot1.png)
![skin selection screen](https://git.silica.codes/Li/HytaleSP/raw/branch/main/images/screenshot2.png)

