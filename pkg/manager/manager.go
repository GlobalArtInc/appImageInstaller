package manager

import (
    "fmt"
    "log"
    "os"
    "os/exec"
    "path/filepath"
    "strings"

    "appinstaller/pkg/desktop"
    "appinstaller/pkg/types"
)

type Manager struct {
    config types.Config
}

func New(config types.Config) *Manager {
    return &Manager{
        config: config,
    }
}

func (m *Manager) Config() types.Config {
    return m.config
}

func (m *Manager) IsGeneratedDesktop(deskFile *desktop.DesktopFile) (bool, error) {
    isValid, err := m.IsValidDesktop(deskFile)
    if !isValid {
        return isValid, err
    }

    execPath, err := deskFile.Category("Desktop Entry").Get("Exec")
    if err != nil {
        return false, err
    }

    if !strings.Contains(execPath, m.config.ExecDir) {
        return false, fmt.Errorf("external executable path")
    }

    return true, nil
}

func (m *Manager) IsValidDesktop(deskFile *desktop.DesktopFile) (bool, error) {
    if !deskFile.HasValues("Desktop Entry", []string{"Name", "Exec"}) {
        return false, fmt.Errorf("missing basic values")
    }

    cmd, err := deskFile.Category("Desktop Entry").Get("Exec")
    if err != nil {
        return false, err
    }

    words := strings.Split(cmd, " ")
    if len(words) == 0 {
        return false, fmt.Errorf("missing executable path")
    }

    path := words[0]
    if _, err := os.Stat(path); err != nil {
        if _, err := exec.LookPath(path); err != nil {
            return false, fmt.Errorf("executable not found: %s", path)
        }
    }

    return true, nil
}

func (m *Manager) List() []*desktop.DesktopFile {
    var appList []*desktop.DesktopFile

    entries, err := os.ReadDir(m.config.GnomeDesktopDir)
    if err != nil {
        log.Fatal(err)
    }

    for _, e := range entries {
        deskFile := desktop.New()
        deskFilePath := filepath.Join(m.config.GnomeDesktopDir, e.Name())
        
        if err := deskFile.FromFile(deskFilePath); err != nil {
            continue
        }

        if _, err := m.IsValidDesktop(deskFile); err != nil {
            continue
        }

        if isGenerated, _ := m.IsGeneratedDesktop(deskFile); !isGenerated {
            continue
        }

        appList = append(appList, deskFile)
    }

    return appList
}

func (m *Manager) Delete(appName string) error {
    entries, err := os.ReadDir(m.config.GnomeDesktopDir)
    if err != nil {
        log.Fatal(err)
    }

    for _, e := range entries {
        deskFile := desktop.New()
        deskFilePath := filepath.Join(m.config.GnomeDesktopDir, e.Name())
        
        if err := deskFile.FromFile(deskFilePath); err != nil {
            continue
        }

        if isGenerated, _ := m.IsGeneratedDesktop(deskFile); !isGenerated {
            continue
        }

        name, err := deskFile.Category("Desktop Entry").Get("Name")
        if err != nil || name != appName {
            continue
        }

        execPath, _ := deskFile.Category("Desktop Entry").Get("Exec")
        
        if err := os.Remove(execPath); err != nil {
            return err
        }

        if err := os.Remove(deskFilePath); err != nil {
            return err
        }

        return nil
    }

    return fmt.Errorf("application not found")
} 