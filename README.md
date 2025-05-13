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
# Download the latest .deb package from releases
wget https://github.com/GlobalArtInc/appImageInstaller/releases/latest/download/appinstaller_linux_amd64.deb
# Install the package
sudo dpkg -i appinstaller_linux_amd64.deb
```

#### RHEL/Fedora:
```bash
# Download the latest .rpm package from releases
wget https://github.com/GlobalArtInc/appImageInstaller/releases/latest/download/appinstaller_linux_amd64.rpm
# Install the package
sudo rpm -i appinstaller_linux_amd64.rpm
```

### Manual Installation

Download the latest binary release and install it system-wide:

```bash
# Download the latest release
wget https://github.com/GlobalArtInc/appImageInstaller/releases/latest/download/appinstaller_linux_amd64.tar.gz
# Extract the archive
tar xzf appinstaller_linux_amd64.tar.gz
# Install the binary
sudo cp appinstaller /usr/bin
sudo chmod +x /usr/bin/appinstaller
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
