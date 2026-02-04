#!/bin/env sh
# just to script to help me not have to ctrl + r the command lol
echo this flatpak is for building and packaging the launcher
set -x
FPBUILD=
if command -v flatpak-builder; then
  FPBUILD=$(command -v flatpak-builder)
else
  FPBUILD="flatpak run org.flatpak.Builder"
fi

$FPBUILD --disable-rofiles-fuse --force-clean build --repo=repo hytaleSP.yaml --user
flatpak build-bundle repo HytaleSP.flatpak trans.hytaleSP.hytaleSP
