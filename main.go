package main

import (
	"appinstaller/pkg/desktop"
	"appinstaller/pkg/fileutil"
	"appinstaller/pkg/manager"
	"appinstaller/pkg/types"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"strconv"
)

func checkDependencies() error {
	dependencies := []struct {
		lib     string
		pkg     string
		desc    string
	}{
		{"libfuse.so.2", "libfuse2", "FUSE library"},
		{"libz.so.1", "zlib1g", "zlib compression library"},
	}

	for _, dep := range dependencies {
		cmd := exec.Command("ldconfig", "-p")
		output, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("failed to check libraries: %v", err)
		}

		if !strings.Contains(string(output), dep.lib) {
			return fmt.Errorf("%s (%s) is missing. Please install it using:\nsudo apt-get install %s", dep.desc, dep.lib, dep.pkg)
		}
	}
	return nil
}

func extractApp(appPath string) {
	if err := os.Chmod(appPath, 0755); err != nil {
		log.Fatal("failed to set executable permissions: ", err)
	}

	file, err := os.Open(appPath)
	if err != nil {
		log.Fatal("failed to open file: ", err)
	}
	defer file.Close()

	magic := make([]byte, 16)
	if _, err := file.Read(magic); err != nil {
		log.Fatal("failed to read file header: ", err)
	}

	if string(magic[:4]) != "\x7FELF" {
		log.Fatal("not a valid AppImage file (missing ELF header)")
	}

	extractionMethods := []func(string) error{
		tryExtractWithUnsquashfs,
		tryExtractWithAppImage,
		tryManualExtract,
	}

	for _, method := range extractionMethods {
		if err := method(appPath); err == nil {
			return
		} else {
			fmt.Printf("Extraction method failed: %v\nTrying next method...\n", err)
		}
	}

	log.Fatal("all extraction methods failed")
}

func tryExtractWithUnsquashfs(appPath string) error {
	fmt.Println("Trying extraction with unsquashfs...")
	
	if _, err := exec.LookPath("unsquashfs"); err != nil {
		return fmt.Errorf("unsquashfs not found: %v", err)
	}

	offsetCmd := exec.Command("sh", "-c", fmt.Sprintf("dd if=%s bs=1 skip=0 count=100000 2>/dev/null | grep -a -b -o hsqs | cut -d ':' -f 1", appPath))
	output, err := offsetCmd.Output()
	if err != nil || len(output) == 0 {
		return fmt.Errorf("failed to find squashfs offset: %v", err)
	}

	offset, err := strconv.Atoi(strings.TrimSpace(string(output)))
	if err != nil {
		return fmt.Errorf("invalid offset value: %v", err)
	}

	squashFile := fmt.Sprintf("%s.squashfs", appPath)
	ddCmd := exec.Command("dd", fmt.Sprintf("if=%s", appPath), fmt.Sprintf("of=%s", squashFile), fmt.Sprintf("bs=1", appPath), fmt.Sprintf("skip=%d", offset))
	if err := ddCmd.Run(); err != nil {
		return fmt.Errorf("failed to extract squashfs part: %v", err)
	}
	defer os.Remove(squashFile)

	unsquashCmd := exec.Command("unsquashfs", "-f", "-d", "squashfs-root", squashFile)
	unsquashCmd.Stdout = os.Stdout
	unsquashCmd.Stderr = os.Stderr
	if err := unsquashCmd.Run(); err != nil {
		return fmt.Errorf("unsquashfs failed: %v", err)
	}

	return nil
}

func tryExtractWithAppImage(appPath string) error {
	fmt.Println("Trying native AppImage extraction...")

	cmd := exec.Command(appPath, "--appimage-extract")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	env := os.Environ()
	env = append(env, "APPIMAGE_EXTRACT_AND_RUN=1", "NO_CLEANUP=1")
	cmd.Env = env

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("native extraction failed: %v", err)
	}
	return nil
}

func tryManualExtract(appPath string) error {
	fmt.Println("Trying manual extraction...")

	if err := os.MkdirAll("squashfs-root", 0755); err != nil {
		return fmt.Errorf("failed to create extraction directory: %v", err)
	}

	file, err := os.Open(appPath)
	if err != nil {
		return fmt.Errorf("failed to open AppImage: %v", err)
	}
	defer file.Close()

	buf := make([]byte, 4096)
	desktopData := ""
	var pos int64 = 0

	for {
		n, err := file.Read(buf)
		if err != nil || n == 0 {
			break
		}

		content := string(buf[:n])
		if idx := strings.Index(content, "[Desktop Entry]"); idx >= 0 {
			startPos := pos + int64(idx)
			file.Seek(startPos, 0)
			
			deskBuf := make([]byte, 4096)
			n, _ := file.Read(deskBuf)
			desktopData = string(deskBuf[:n])
			break
		}

		pos += int64(n)
	}

	if desktopData == "" {
		return fmt.Errorf("failed to find desktop entry in AppImage")
	}

	deskPath := filepath.Join("squashfs-root", "test.desktop")
	if err := os.WriteFile(deskPath, []byte(desktopData), 0644); err != nil {
		return fmt.Errorf("failed to write desktop file: %v", err)
	}

	dst := filepath.Join("squashfs-root", filepath.Base(appPath))
	if err := fileutil.Copy(appPath, dst); err != nil {
		return fmt.Errorf("failed to copy AppImage: %v", err)
	}

	return nil
}

func createDirectories(config types.Config) error {
	err := os.MkdirAll(config.ExtractDir, os.ModePerm)
	if err != nil {
		return err
	}
	err = os.MkdirAll(config.ExecDir, os.ModePerm)
	if err != nil {
		return err
	}
	err = os.MkdirAll(config.ImgPath, os.ModePerm)
	if err != nil {
		return err
	}
	err = os.Chmod(config.ExecDir, 0777)
	if err != nil {
		return err
	}
	return nil
}

func setConfig(path string) types.Config {
	config := types.Config{
		AppExtractDir:   "squashfs-root",
		ExtractDir:      "/tmp/appInstaller",
		ExecDir:         "/usr/share/appImages/",
		GnomeDesktopDir: "/usr/share/applications/",
		Debug:           false,
		ImgPath:         "/usr/share/pixmaps/",
		InputPath:       path,
	}
	config.ExecPath = filepath.Join(config.ExecDir, filepath.Base(config.InputPath))
	config.AppExtractDir = filepath.Join(config.ExtractDir, "squashfs-root")
	config.InputDir = filepath.Dir(config.InputPath)
	config.InputFileName = filepath.Base(config.InputPath)
	return config
}

func preInstall(config types.Config) error {
	err := createDirectories(config)
	if err != nil {
		log.Fatal("failed to create directories: ", err)
	}
	err = os.Chmod(config.InputPath, 0777)
	if err != nil {
		return err
	}
	err = fileutil.Copy(config.InputPath, config.ExecDir)
	if err != nil {
		return err
	}
	return nil
}

func findInternalDesktop(path string) string {
	desktopPath, err := fileutil.FindFile(path, []string{".desktop"})
	if err != nil {
		log.Fatal("failed to find desktop file: ", err)
	}
	return desktopPath
}

func generateDesktopFile(path string) (*desktop.DesktopFile, error) {
	deskFile := desktop.New()
	err := deskFile.FromFile(path)
	if err != nil {
		return nil, err
	}
	return deskFile, nil
}

func editDesktop(deskFile *desktop.DesktopFile, config types.Config) {
	execPath := filepath.Join(config.ExecDir, config.InputFileName)
	deskFile.Category("Desktop Entry").Set("Exec", execPath)
}

func copyImage(deskFile *desktop.DesktopFile, config types.Config) error {
	icon, err := deskFile.Category("Desktop Entry").Get("Icon")
	if err != nil || icon == "" {
		priorityDirs := []string{
			filepath.Join(config.AppExtractDir, "usr/share/icons"),
			filepath.Join(config.AppExtractDir, "usr/share/pixmaps"),
			filepath.Join(config.AppExtractDir, ".DirIcon"),
			config.AppExtractDir,
		}
		
		for _, dir := range priorityDirs {
			if _, statErr := os.Stat(dir); statErr == nil {
				iconFiles, _ := fileutil.FindFiles(dir, []string{".png", ".svg", ".xpm", ".ico"})
				if len(iconFiles) > 0 {
					icon = iconFiles[0]
					break
				}
			} else if strings.HasSuffix(dir, ".DirIcon") {
				if _, statErr := os.Stat(filepath.Join(config.AppExtractDir, ".DirIcon")); statErr == nil {
					icon = filepath.Join(config.AppExtractDir, ".DirIcon")
					break
				}
			}
		}
		
		if icon == "" {
			iconFiles, _ := fileutil.FindFiles(config.AppExtractDir, []string{".png", ".svg", ".xpm", ".ico"})
			if len(iconFiles) > 0 {
				icon = iconFiles[0]
			} else {
				return fmt.Errorf("no icon found in desktop file or directory")
			}
		}
	}

	newPath := filepath.Join(config.ImgPath, filepath.Base(icon))
	
	if filepath.IsAbs(icon) {
		iconPath := filepath.Join(config.AppExtractDir, icon)
		err = fileutil.Copy(iconPath, newPath)
		if err != nil {
			iconName := filepath.Base(icon)
			possiblePaths := []string{
				filepath.Join(config.AppExtractDir, iconName),
				filepath.Join(config.AppExtractDir, "usr/share/icons", iconName),
				filepath.Join(config.AppExtractDir, "usr/share/pixmaps", iconName),
				filepath.Join(config.AppExtractDir, ".DirIcon"),
			}
			
			for _, path := range possiblePaths {
				if _, statErr := os.Stat(path); statErr == nil {
					err = fileutil.Copy(path, newPath)
					if err == nil {
						deskFile.Category("Desktop Entry").Set("Icon", newPath)
						return nil
					}
				}
			}
			
			if _, err := os.Stat(icon); err == nil {
				err = fileutil.Copy(icon, newPath)
				if err == nil {
					deskFile.Category("Desktop Entry").Set("Icon", newPath)
					return nil
				}
			}
			
			return fmt.Errorf("failed to copy icon: %v", err)
		}
		deskFile.Category("Desktop Entry").Set("Icon", newPath)
	} else {
		iconPath := ""
		possiblePaths := []string{
			filepath.Join(config.AppExtractDir, icon),
			filepath.Join(config.AppExtractDir, "usr/share/icons", icon),
			filepath.Join(config.AppExtractDir, "usr/share/pixmaps", icon),
		}
		
		extensions := []string{"", ".png", ".svg", ".xpm", ".ico"}
		
		for _, path := range possiblePaths {
			for _, ext := range extensions {
				fullPath := path + ext
				if _, statErr := os.Stat(fullPath); statErr == nil {
					iconPath = fullPath
					break
				}
			}
			if iconPath != "" {
				break
			}
		}
		
		if iconPath != "" {
			err = fileutil.Copy(iconPath, newPath)
			if err == nil {
				deskFile.Category("Desktop Entry").Set("Icon", newPath)
				return nil
			}
		} else {
			err = fileutil.Copy(icon, newPath)
			if err == nil {
				deskFile.Category("Desktop Entry").Set("Icon", newPath)
				return nil
			}
		}
	}
	
	if icon != "" && err == nil {
		deskFile.Category("Desktop Entry").Set("Icon", newPath)
		return nil
	}
	
	return fmt.Errorf("failed to find or copy icon: %v", err)
}

func install(config types.Config) {
	err := preInstall(config)
	if err != nil {
		log.Fatal("setup failed: ", err)
	}
	os.Chdir(config.ExtractDir)
	extractApp(config.InputPath)
	desktopPath := findInternalDesktop(config.AppExtractDir)
	deskFile, err := generateDesktopFile(desktopPath)
	if err != nil {
		log.Fatal("failed to parse desktop file: ", err)
	}
	editDesktop(deskFile, config)
	err = copyImage(deskFile, config)
	if err != nil {
		fmt.Println(err)
	}
	err = deskFile.ToFile(filepath.Join(config.GnomeDesktopDir, filepath.Base(desktopPath)))
	if err != nil {
		log.Fatal("failed to write desktop file ", err)
	}
}

func runInstallScript(appPath string) error {
	if err := checkDependencies(); err != nil {
		return err
	}

	path, _ := filepath.Abs(appPath)
	_, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("install %s: %w", appPath, err)
	}
	config := setConfig(path)
	err = os.RemoveAll(config.ExtractDir)
	if err != nil {
		return fmt.Errorf("removing extract directory: %w", err)
	}
	install(config)
	err = os.RemoveAll(config.ExtractDir)
	if err != nil {
		return fmt.Errorf("removing extract directory: %w", err)
	}
	return nil
}

func help() {
	fmt.Println("Usage: sudo appinstaller [OPTIONS] [path/to/app.AppImage]")
	fmt.Println("(after install use sudo update-desktop-database to reload gnome icons)")
	fmt.Println("\nOptions:")
	fmt.Println("  -l, --list            List installed apps (from this tool only)")
	fmt.Println("  -d, --delete <name>   Delete the specified app (installed by this tool)")
	fmt.Println("  -h, --help            Show this help message")
	fmt.Println("  -v, --version         Show version information")
}

func checkFzf() bool {
	_, err := exec.LookPath("fzf")
	return err == nil
}

func listingWithFzf(m *manager.Manager, entries []*desktop.DesktopFile) error {
	var items []string
	for _, entry := range entries {
		name, _ := entry.Category("Desktop Entry").Get("Name")
		exec, _ := entry.Category("Desktop Entry").Get("Exec")
		execPath := strings.Split(exec, " ")[0]
		items = append(items, fmt.Sprintf("%-30s | %s", name, execPath))
	}

	cmd := exec.Command("fzf", "--header=Select application to delete (ESC to exit)", "--height=40%")
	cmd.Stderr = os.Stderr
	
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	
	go func() {
		defer stdin.Close()
		for _, item := range items {
			fmt.Fprintln(stdin, item)
		}
	}()

	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 130 {
			return nil
		}
		return err
	}

	selected := strings.TrimSpace(string(output))
	if selected == "" {
		return nil
	}

	name := strings.TrimSpace(strings.Split(selected, "|")[0])
	fmt.Printf("Deleting application '%s'... ", name)
	
	if err := m.Delete(name); err != nil {
		fmt.Printf("error: %v\n", err)
		return err
	}
	fmt.Println("success")
	return nil
}

func listing() error {
	config := setConfig("")
	m := manager.New(config)
	entries := m.List()
	
	if len(entries) == 0 {
		fmt.Println("No installed applications found")
		return nil
	}

	if checkFzf() {
		return listingWithFzf(m, entries)
	}

	fmt.Printf("\n%-4s | %-30s | %-50s\n", "#", "Name", "Executable Path")
	fmt.Println(strings.Repeat("-", 87))

	for i, entry := range entries {
		name, _ := entry.Category("Desktop Entry").Get("Name")
		exec, _ := entry.Category("Desktop Entry").Get("Exec")
		execPath := strings.Split(exec, " ")[0]
		fmt.Printf("%-4d | %-30s | %-50s\n", i+1, name, execPath)
	}
	fmt.Println(strings.Repeat("-", 87))

	fmt.Print("\nEnter application number to delete (or 'q' to exit): ")
	var input string
	fmt.Scanln(&input)

	if input == "q" {
		return nil
	}

	if num, err := strconv.Atoi(input); err == nil && num > 0 && num <= len(entries) {
		name, _ := entries[num-1].Category("Desktop Entry").Get("Name")
		fmt.Printf("Deleting application '%s'... ", name)
		
		if err := m.Delete(name); err != nil {
			fmt.Printf("error: %v\n", err)
			return err
		}
		fmt.Println("success")
	} else {
		fmt.Println("Invalid input")
	}

	return nil
}

func deleteApp(appName string) error {
	config := setConfig("")
	m := manager.New(config)
	return m.Delete(appName)
}

func chooseScript() error {
	if len(os.Args) < 2 {
		help()
		return nil
	}

	switch os.Args[1] {
	case "-l", "--list":
		return listing()
	case "-d", "--delete":
		if len(os.Args) >= 3 {
			return deleteApp(os.Args[2])
		}
		fmt.Println("Error: Application name required for delete operation")
		help()
		return fmt.Errorf("missing application name")
	case "-h", "--help":
		help()
		return nil
	case "-v", "--version":
		fmt.Println("1.0")
		return nil
	default:
		if len(os.Args) == 2 {
			return runInstallScript(os.Args[1])
		}
		help()
		return nil
	}
}

func checkSuperuser() bool {
	testDirs := []string{
		"/usr/share/applications",
		"/usr/share/pixmaps",
		"/usr/share/appImages",
	}

	for _, dir := range testDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			err = os.MkdirAll(dir, 0755)
			if err != nil {
				return false
			}
			continue
		}

		testFile := filepath.Join(dir, ".write_test")
		f, err := os.OpenFile(testFile, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return false
		}
		f.Close()
		os.Remove(testFile)
	}
	
	return true
}

func main() {
	if !checkSuperuser() {
		fmt.Println("Error: This application requires superuser privileges")
		fmt.Println("Please run with sudo: sudo appinstaller [options]")
		os.Exit(1)
	}

	if len(os.Args) < 2 {
		help()
		os.Exit(0)
	}

	switch os.Args[1] {
	case "-h", "--help":
		help()
		os.Exit(0)
	case "-v", "--version":
		fmt.Println("1.0")
		os.Exit(0)
	}

	err := chooseScript()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
