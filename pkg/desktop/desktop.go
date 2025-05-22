package desktop

import (
    "bufio"
    "fmt"
    "os"
    "regexp"
    "strings"
    "path/filepath"
)

type DesktopFile struct {
    categories      map[string]map[string]string
    activeCategory string
    sourcePath     string
}

func New() *DesktopFile {
    return &DesktopFile{
        categories:      make(map[string]map[string]string),
        activeCategory: "root",
        sourcePath:     "self-generated",
    }
}

func (d *DesktopFile) Category(name string) *DesktopFile {
    d.activeCategory = name
    return d
}

func (d *DesktopFile) Get(name string) (string, error) {
    category, exists := d.categories[d.activeCategory]
    if !exists {
        return "", fmt.Errorf("category %s not found", d.activeCategory)
    }
    
    value, exists := category[name]
    if !exists {
        return "", fmt.Errorf("parameter %s not found in category %s", name, d.activeCategory)
    }
    
    d.activeCategory = "root"
    return value, nil
}

func (d *DesktopFile) Set(name, value string) error {
    if _, exists := d.categories[d.activeCategory]; !exists {
        d.categories[d.activeCategory] = make(map[string]string)
    }
    
    d.categories[d.activeCategory][name] = value
    d.activeCategory = "root"
    return nil
}

func (d *DesktopFile) HasValues(category string, values []string) bool {
    for _, value := range values {
        if _, err := d.Category(category).Get(value); err != nil {
            return false
        }
    }
    return true
}

func (d *DesktopFile) FromFile(path string) error {
    file, err := os.Open(path)
    if err != nil {
        return fmt.Errorf("error opening file: %w", err)
    }
    defer file.Close()

    d.sourcePath = path
    return d.parseFile(bufio.NewScanner(file))
}

func (d *DesktopFile) ToFile(path string) error {
    file, err := os.Create(path)
    if err != nil {
        return fmt.Errorf("error creating file: %w", err)
    }
    defer file.Close()

    for category, params := range d.categories {
        if _, err := fmt.Fprintf(file, "[%s]\n", category); err != nil {
            return fmt.Errorf("error writing category: %w", err)
        }
        
        for name, value := range params {
            if _, err := fmt.Fprintf(file, "%s=%s\n", name, value); err != nil {
                return fmt.Errorf("error writing parameter: %w", err)
            }
        }
    }
    return nil
}

func (d *DesktopFile) GetSource() string {
    return d.sourcePath
}

func (d *DesktopFile) parseFile(scanner *bufio.Scanner) error {
    category := "root"
    categoryRegex := regexp.MustCompile(`^[[:space:]]*\[(.+)\][[:space:]]*$`)
    parameterRegex := regexp.MustCompile(`^(.+)=(.+)$`)

    for scanner.Scan() {
        line := strings.TrimSpace(scanner.Text())
        if line == "" || strings.HasPrefix(line, "#") {
            continue
        }

        if matches := categoryRegex.FindStringSubmatch(line); matches != nil {
            category = matches[1]
            if _, exists := d.categories[category]; !exists {
                d.categories[category] = make(map[string]string)
            }
            continue
        }

        if matches := parameterRegex.FindStringSubmatch(line); matches != nil {
            name := strings.TrimSpace(matches[1])
            value := strings.TrimSpace(matches[2])
            d.categories[category][name] = value
        }
    }

    if err := scanner.Err(); err != nil {
        return fmt.Errorf("error reading file: %w", err)
    }
    return nil
}

func (d *DesktopFile) CreateAutostart(autostartDir string) error {
    if err := os.MkdirAll(autostartDir, 0755); err != nil {
        return fmt.Errorf("error creating autostart directory: %w", err)
    }

    name, err := d.Category("Desktop Entry").Get("Name")
    if err != nil {
        return fmt.Errorf("error getting application name: %w", err)
    }

    autostartPath := filepath.Join(autostartDir, fmt.Sprintf("%s.desktop", strings.ToLower(strings.ReplaceAll(name, " ", "-"))))
    
    if err := d.ToFile(autostartPath); err != nil {
        return fmt.Errorf("error creating autostart entry: %w", err)
    }

    return nil
} 