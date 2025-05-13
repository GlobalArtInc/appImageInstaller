# Instructions for creating a Flatpak package for AppImageInstaller

## Prerequisites

Make sure you have the following tools installed:

1. Flatpak: `sudo apt install flatpak` (or equivalent command for your distribution)
2. Flatpak Builder: `sudo apt install flatpak-builder` (or equivalent command for your distribution)
3. Required SDKs and runtimes: `flatpak install flathub org.freedesktop.Platform//22.08 org.freedesktop.Sdk//22.08`

## Steps to create a Flatpak package

### 1. Build the appinstaller binary

First, build the binary file using GoReleaser or manually:

```bash
# Using GoReleaser
goreleaser build --snapshot --rm-dist

# Or manually
go build -o flatpak/appinstaller .
```

### 2. Check the manifest

Make sure that the `flatpak/com.globalart.appinstaller.yml` file exists and contains the correct settings. Edit it if necessary.

### 3. Build the Flatpak package

```bash
cd flatpak
flatpak-builder --user --install --force-clean build-dir com.globalart.appinstaller.yml
```

### 4. Create a repository and .flatpak file (optional)

If you want to create a repository and a separate .flatpak file for distribution:

```bash
# Create a repository
flatpak-builder --repo=repo --force-clean build-dir com.globalart.appinstaller.yml

# Create a .flatpak package
flatpak build-bundle repo appinstaller.flatpak com.globalart.appinstaller
```

### 5. Install the Flatpak package

```bash
# If you created a .flatpak file
flatpak install --user appinstaller.flatpak

# If you installed from a local repository
flatpak install --user --reinstall com.globalart.appinstaller
```

### 6. Run the application

```bash
flatpak run com.globalart.appinstaller
```

## Possible issues and solutions

1. **Filesystem access error**: If the application doesn't have access to the filesystem, check the finish-args in the manifest and make sure the necessary permissions are included.

2. **Dependency problems**: If your application requires special dependencies, add them to the modules in the manifest.

## Publishing on Flathub

To publish on Flathub, you will need to:

1. Fork https://github.com/flathub/flathub
2. Add the manifest to the repository
3. Submit a pull request to the main Flathub repository

For details, see the website: https://docs.flathub.org/docs/for-app-authors/submission 