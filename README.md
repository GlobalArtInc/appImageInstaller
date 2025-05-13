# AppImage Installer

![CI](https://github.com/GlobalArtInc/appImageInstaller/actions/workflows/go.yml/badge.svg)
![Go](https://img.shields.io/badge/go-%2300ADD8.svg?style=for-the-badge&logo=go&logoColor=white)
![Release](https://img.shields.io/github/v/release/GlobalArtInc/appImageInstaller?style=for-the-badge&sort=semver)
![Issues](https://img.shields.io/github/issues-raw/GlobalArtInc/appImageInstaller?style=for-the-badge)

A lightweight tool for seamless integration of AppImage applications into Linux desktop environments.

## Features

- Automatic desktop entry creation
- System-wide installation support
- Icon extraction and integration
- Clean and simple command-line interface
- Proper file permissions handling

## Installation

### Using Package Managers

#### Debian/Ubuntu:
```bash
curl -L https://github.com/GlobalArtInc/appImageInstaller/releases/latest/download/appinstaller_linux_amd64.deb -o /tmp/appinstaller.deb && sudo dpkg -i /tmp/appinstaller.deb && rm /tmp/appinstaller.deb
```

#### RHEL/Fedora:
```bash
curl -L https://github.com/GlobalArtInc/appImageInstaller/releases/latest/download/appinstaller_linux_amd64.rpm -o /tmp/appinstaller.rpm && sudo rpm -i /tmp/appinstaller.rpm && rm /tmp/appinstaller.rpm
```

#### Snap:
```bash
# Install from Snap Store
sudo snap install appinstaller

# Or install from a local snap file
sudo snap install appinstaller_amd64.snap --dangerous
```

#### Flatpak:
```bash
# Install from a local Flatpak bundle
flatpak install --user appinstaller.flatpak

# After adding a repository
flatpak install --user com.globalart.appinstaller
```

See [flatpak/FLATPAK.md](flatpak/FLATPAK.md) for more information on building and installing the Flatpak package.

### Manual Installation

Download and install the latest binary release:

```bash
curl -L https://github.com/GlobalArtInc/appImageInstaller/releases/latest/download/appinstaller_linux_amd64.tar.gz -o /tmp/appinstaller.tar.gz && sudo tar xzf /tmp/appinstaller.tar.gz -C /usr/bin/ appinstaller && rm /tmp/appinstaller.tar.gz && sudo chmod +x /usr/bin/appinstaller
```

## Usage

Install an AppImage:
```bash
sudo appinstaller /path/to/your/application.AppImage
```

List installed applications:
```bash
sudo appinstaller -l
```

Remove an installed application:
```bash
sudo appinstaller -d "Application Name"
```

View help:
```bash
appinstaller -h
```

## How It Works

1. Extracts the AppImage in a temporary directory
2. Locates and processes the .desktop file
3. Copies the application to a system directory
4. Integrates icons and creates desktop entries
5. Cleans up temporary files

## Requirements

- Linux operating system
- Root privileges for installation
- GNOME-compatible desktop environment

## Contributing

1. Fork the repository
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

If you encounter any issues or have suggestions, please [open an issue](https://github.com/GlobalArtInc/appImageInstaller/issues).
