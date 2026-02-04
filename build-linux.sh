#!/bin/sh

make -C Aurora
go build -ldflags="-s -w -buildid=" -trimpath

./build-flatpak.sh