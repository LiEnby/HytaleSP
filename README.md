#HytaleSP

An alternative launcher for "Hytale" with a fairly streightforward "native" UI
 features:
 
 - Multiple version management
 - Incremental patch from one version to another
 - Universal "online fix" or that works on all versions;
 - Access all "Supporter" and "Cursebreaker" cosmetic items.
 - Play local or online multiplayer
 - Completely standalone; a single executable- no butler or other external tools.
 - Fairly transparent online patch implementation.
 - Run offline and download without needing a hytale account.
 - Supports Windows 7+
 
Currently for: Windows and Linux (i would do MacOS, but i dont have access to an ARM Mac right now.)


NOTE: you can even probably play online, but only if the servers your joining have ``--auth-mode=insecure`` set in the command line;
enabling this option would also mean anyone using the offical launcher cannot play 
because it doesn't allow insecure auth type outside singleplayer offline mode; .. for some reason

also if you use the "online play" feature in singleplayer should also work as long as other users also use hytLauncher ...

## Project layout

hytServer.go - hytale authentication server emulator "online fix"

hytAuth.go - implementation of hytale OAuth2.0, currently not used

hytFormats.go - most of the JSON structures used by hytale are here;

hytClient.go - downloading versions, etc

patch.go - itch.io's 'apply patch' wharf code

jwt.go - JSON structures used in hytale auth tokens

locations.go - default / location resolvers for many folders


Aurora/ - c code for the dll or shared object, loaded with the game, 
that replaces account-data.hytale.com with a custom server .. 

# Building

on windows, you first have to build the "Aurora.dll" using MSVC, 
and then you can use ``go build .``
or you can ``build-windows.bat`` within the VS2026 developer command prompt to do this;

on linux, you need ``build-essential`` and then you can build "Aurora.so" using its Makefile;
after that you can use ``go build .`` 
or you can use ``build-linux.sh``

# Online Multiplayer 

When using the "fake online" option, 
you CAN play online multiplayer; BUT you can only join servers with the command-line flag ``--auth-mode=insecure`` set; 
this is much simular to how minecraft works where you can only join servers with ``offline_mode=true`` in the server.properties

however unlike minecraft, players using the offical game are actually unable to join "insecure" servers,
meaning all players would have to use the 'fake online' option in this (or another) launcher to play on these servers.

there are server side plugins that can get around this restriction.

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

