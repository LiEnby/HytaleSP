GOC := go
RCC := windres
LDFLAGS := -X main.wVersion=$(shell git describe --abbrev=0 --exclude=continuous --tags) -s -w
GOFLAGS := -trimpath

AURORA := Aurora/Build/Aurora
BINARY := HytaleSP
DELCMD := $(if $(filter-out $(OS),Windows_NT),rm -rf,del /Q /F /S)

DLL := .so
EXE :=
OBJ :=

TARGET ?= $(if $(filter-out $(OS),Windows_NT),$(shell uname),Windows)

ifeq ($(TARGET),Windows)
	DLL := .dll
	EXE := .exe
	OBJ += resources.syso
	LDFLAGS += -H=windowsgui -extldflags=-static
endif

$(BINARY)$(EXE): setup $(AURORA)$(DLL) $(OBJ)
	$(GOC) build -ldflags="$(LDFLAGS)" $(GOFLAGS) -o $@ .

$(AURORA)$(DLL):
ifeq ($(VSCMD_VER),)
	make -C Aurora
else
	msbuild Aurora/Aurora.slnx /p:Configuration=Release
endif
	
ifeq ($(TARGET),Windows)
resources.syso:
	$(RCC) Resources/res.rc -O coff -o $@
endif

ifeq ($(TARGET),Linux)
# TODO: Move flatpak build outside of shell scripts.
@PHONY: flatpak
flatpak: $(BINARY)$(EXE)
	flatpak install org.freedesktop.Sdk//25.08 org.flatpak.Builder --system -y
	cd flatpak && ./buildpak.sh
	mv flatpak/$(BINARY).flatpak ./$(BINARY).flatpak
endif

@PHONY: setup
setup:
ifeq ($(TARGET),Windows)
	-win7go
endif
	go mod tidy

@PHONY: clean
clean:
	-$(DELCMD) $(BINARY)$(EXE)
	-$(DELCMD) $(OBJ)
	-$(DELCMD) $(BINARY).flatpak
	-make -C Aurora clean