GOC=go
LDFLAGS = -X main.wVersion=$(shell git describe --abbrev=0 --exclude=continuous --tags) -s -w
GOFLAGS = -trimpath
AURORA = Aurora/Build/Aurora
BINARY = HytaleSP
DELCMD = rm -rf
DLL = .so
EXE =
SYSO =

ifeq ($(OS),Windows_NT)
	DLL := .dll
	EXE := .exe
	DELCMD := del /Q /F /S
	SYSO += resources.syso
	LDFLAGS += -H=windowsgui -extldflags=-static
endif

$(BINARY)$(EXE): setup $(AURORA)$(DLL) $(SYSO)
	$(GOC) build -ldflags="$(LDFLAGS)" $(GOFLAGS) -o $@ .

$(AURORA)$(DLL):
ifeq ($(VSCMD_VER),)
	make -C Aurora
else
	msbuild Aurora/Aurora.slnx /p:Configuration=Release
endif
	
ifeq ($(OS),Windows_NT)
resources.syso:
	windres Resources/res.rc -O coff -o $@
endif

# TODO: Move flatpak build outside of shell scripts.
@PHONY: flatpak
flatpak:
	flatpak install org.freedesktop.Sdk//25.08 org.flatpak.Builder --system -y
	cd flatpak || exit
	./buildpak.sh

@PHONY: setup
setup:
	-win7go
	go mod tidy

@PHONY: clean
clean:
	-$(DELCMD) $(BINARY)$(EXE)
	-$(DELCMD) $(SYSO)
	-make -C Aurora clean