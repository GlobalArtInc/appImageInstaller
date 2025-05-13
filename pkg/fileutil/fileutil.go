package fileutil

import (
    "fmt"
    "io"
    "os"
    "path/filepath"
    "strings"
    "syscall"
)

func FindFile(path string, patterns []string) (string, error) {
    foundPath := ""
    err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return nil
        }
        if !info.IsDir() {
            for _, pattern := range patterns {
                if strings.Contains(filepath.Base(path), pattern) {
                    foundPath = path
                    return io.EOF
                }
            }
        }
        return nil
    })
    
    if err == io.EOF {
        err = nil
    }
    return foundPath, err
}

func FindFiles(path string, extensions []string) ([]string, error) {
    var foundFiles []string
    err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
        if err != nil {
            return nil
        }
        if !info.IsDir() {
            ext := strings.ToLower(filepath.Ext(filePath))
            for _, extension := range extensions {
                extToCheck := extension
                if !strings.HasPrefix(extToCheck, ".") {
                    extToCheck = "." + extToCheck
                }
                
                if strings.EqualFold(ext, extToCheck) {
                    foundFiles = append(foundFiles, filePath)
                    break
                }
            }
        }
        return nil
    })
    
    return foundFiles, err
}

func GetOwner(file string) (int, int, error) {
    info, err := os.Stat(file)
    if err != nil {
        return 0, 0, err
    }

    stat, ok := info.Sys().(*syscall.Stat_t)
    if !ok {
        return 0, 0, fmt.Errorf("unable to get file system info")
    }

    return int(stat.Uid), int(stat.Gid), nil
}

func Copy(src, dst string) error {
    srcStat, err := os.Stat(src)
    if err != nil {
        return fmt.Errorf("error getting file info %s: %w", src, err)
    }

    if !srcStat.Mode().IsRegular() {
        return fmt.Errorf("%s is not a regular file", src)
    }

    dstStat, err := os.Stat(dst)
    if err == nil && dstStat.IsDir() {
        dst = filepath.Join(dst, filepath.Base(src))
    }

    source, err := os.Open(src)
    if err != nil {
        return err
    }
    defer source.Close()

    destination, err := os.Create(dst)
    if err != nil {
        return err
    }
    defer destination.Close()

    if _, err = io.Copy(destination, source); err != nil {
        return err
    }

    uid, gid, err := GetOwner(src)
    if err != nil {
        return err
    }

    if err = os.Chmod(dst, srcStat.Mode()); err != nil {
        return err
    }

    return os.Chown(dst, uid, gid)
} 