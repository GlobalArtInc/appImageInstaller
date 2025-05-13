#!/bin/bash

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Starting Flatpak package build for AppImageInstaller${NC}"

for cmd in flatpak flatpak-builder go; do
  if ! command -v $cmd &>/dev/null; then
    echo -e "${RED}Error: $cmd is not installed${NC}"
    echo "Please install the required tools:"
    echo "flatpak, flatpak-builder, go"
    exit 1
  fi
done

echo -e "${YELLOW}Checking Flatpak runtimes...${NC}"
if ! flatpak info org.freedesktop.Platform//22.08 &>/dev/null; then
  echo -e "${YELLOW}Installing Freedesktop Platform...${NC}"
  flatpak install -y flathub org.freedesktop.Platform//22.08
fi

if ! flatpak info org.freedesktop.Sdk//22.08 &>/dev/null; then
  echo -e "${YELLOW}Installing Freedesktop SDK...${NC}"
  flatpak install -y flathub org.freedesktop.Sdk//22.08
fi

echo -e "${YELLOW}Compiling the application...${NC}"
cd ..
go build -o flatpak/appinstaller .
chmod +x flatpak/appinstaller
cd flatpak

echo -e "${YELLOW}Building Flatpak package...${NC}"
flatpak-builder --force-clean build-dir com.globalart.appinstaller.yml

echo -e "${YELLOW}Creating repository and .flatpak file...${NC}"
flatpak-builder --repo=repo --force-clean build-dir com.globalart.appinstaller.yml
flatpak build-bundle repo appinstaller.flatpak com.globalart.appinstaller

echo -e "${GREEN}Build completed successfully!${NC}"
echo -e "Build results:"
echo -e "- Flatpak package: ${YELLOW}appinstaller.flatpak${NC}"
echo -e "- Repository: ${YELLOW}repo/${NC}"
echo ""
echo -e "To install the package, run: ${YELLOW}flatpak install --user appinstaller.flatpak${NC}"
echo -e "To run the application, execute: ${YELLOW}flatpak run com.globalart.appinstaller${NC}"
