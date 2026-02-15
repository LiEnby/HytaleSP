GOC ?= go
RCC ?= windres
LDFLAGS ?= -X main.wVersion=$(TAG) -s -w
GOFLAGS ?= -trimpath

TAG := $(shell git describe --abbrev=0 --exclude=continuous --tags)
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
# use msbuild if vscmd otherwise use make
ifeq ($(VSCMD_VER),)
	make -C Aurora
else
	msbuild Aurora/Aurora.slnx /p:Configuration=Release
endif

ifeq ($(TARGET),Windows)
COMMA := ,
NORM_VER := $(subst $(subst ,, ),.,$(foreach IDX,1 2 3 4,$(if $(word $(IDX),$(subst ., ,$(subst v,,$(TAG)))),$(word $(IDX),$(subst ., ,$(subst v,,$(TAG)))),0)))

Resources/version.h:
	echo #define VER $(subst .,$(COMMA),$(NORM_VER)) >$@
	echo #define VERSTR "$(NORM_VER)" >>$@
resources.syso: Resources/version.h
	$(RCC) Resources/res.rc -O coff -o $@
endif

# TODO: Move flatpak build outside of shell scripts.
ifeq ($(TARGET),Linux)
.PHONY: flatpak
flatpak: $(BINARY)$(EXE)
	flatpak install org.freedesktop.Sdk//25.08 org.flatpak.Builder --system -y
	cd flatpak && ./buildpak.sh
	mv flatpak/$(BINARY).flatpak ./$(BINARY).flatpak
endif

.PHONY: setup
setup:
ifeq ($(TARGET),Windows)
	-win7go
endif
	go mod tidy

.PHONY: clean
clean:
	-$(DELCMD) $(BINARY)$(EXE)
	-$(DELCMD) $(OBJ)
	-$(DELCMD) $(BINARY).flatpak
	-$(DELCMD) Resources\version.h
	-make -C Aurora clean
