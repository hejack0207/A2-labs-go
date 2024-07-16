package main

import (
    "fmt"
    "context"
    "sync"
    "strings"
    "os"
    "errors"
    "path/filepath"
    "io/ioutil"
    "path"

    log "github.com/sirupsen/logrus"
    "github.com/chyroc/go-aliyundrive"
    "github.com/olekukonko/tablewriter"
)

type Cli struct {
    ali *aliyundrive.AliyunDrive
    setupDriveOnce sync.Once
    driveID string
    currentFileID string
}

func (r *Cli) setupDrive() (finalErr error) {
    r.setupDriveOnce.Do(func() {
        user, err := r.ali.Auth.LoginByQrcode(context.TODO(), &aliyundrive.LoginByQrcodeReq{SmallQrCode: true})
        if err != nil {
            finalErr = err
            return
        }
        r.driveID = user.DefaultDriveID
    })
    return finalErr
}

func (r *Cli) setupFiles() ([]*aliyundrive.File, error) {
    resp, err := r.ali.File.GetFileList(context.Background(), &aliyundrive.GetFileListReq{
        GetAll: true,
        DriveID: r.driveID,
        ParentFileID: r.currentFileID,
        Limit: 100,
    })
    if err != nil {
        return nil, err
    }
    return resp.Items, nil
}

func (r *Cli) getFileId(path string) (string, error) {
    path = strings.Trim(path, "/")
    r.setupDrive()
    dirnames := strings.Split(path, "/")

    log.Debugf("dirnames: %s",dirnames)
    if len(dirnames) == 0 || (len(dirnames) == 1 && dirnames[0] == "") {
        return "root", nil
    }

    r.currentFileID = "root"
    for _, dirname := range dirnames {
        files, e := r.setupFiles()
        if e != nil {
            log.Errorln("list content fo directory "+ dirname + " failed")
            return "", e
        }
        found := false
        for _, v := range files {
            if v.Name == dirname {
                r.currentFileID = v.FileID
                found = true
                break
            }
        }
        if !found {
            return "", errors.New("directory " + dirname + " not found")
        }
    }
    return r.currentFileID, nil
}

func (r *Cli) fileExists(path string) bool {
    _, e := r.getFileId(path)
    if e != nil {
        return false
    } else {
        return true
    }

}

func (r *Cli) rmFile(path string) error{
    log.Debugf("rmFile %s", path)
    if err := r.setupDrive(); err != nil {
        return err
    }

    fileId, err := r.getFileId(path)
    if err != nil {
        return err
    }

    _, err = r.ali.File.DeleteFile(context.Background(), &aliyundrive.DeleteFileReq{
        DriveID: r.driveID,
        FileID:  fileId,
    })
    if err != nil {
        log.Debugf("rm file %d error: e", path, err)
        return err
    }

    return nil
}

func (r *Cli) mkDir(filepath string) (string, error){
    log.Debugf("mkDir %s", filepath)
    if err := r.setupDrive(); err != nil {
        return "", err
    }
    if !isValidDirName(filepath) {
        return "", fmt.Errorf("%q not a legal directory name", filepath)
    }
    dir, name := path.Split(filepath)
    fileId, e := r.getFileId(dir);
    if e!=nil {
        fileId, e = r.mkDir(dir)
        if e!=nil {
            return "", fmt.Errorf("error occurs when create parent directory %s", dir)
        }
    }
    if name == "" {
        return fileId, nil
    }
    resp, err := r.ali.File.CreateFolder(context.Background(), &aliyundrive.CreateFolderReq{
        DriveID:      r.driveID,
        ParentFileID: fileId,
        Name:         name,
    })
    if err != nil {
        return "", err
    }

    return resp.FileID, nil
}

func (r *Cli) printFiles(files []*aliyundrive.File) {
    if len(files) == 0 {
        log.Fatalln("no file found")
        return
    }

    header := []string{
        "name", "type", "size", "last modify time",
    }
    data := [][]string{}
    for _, f := range files {
        data = append(data, []string{
            f.Name, f.Type, fmt.Sprintf("%d",f.Size), fmt.Sprintf("%d",f.UpdatedAt),
        })
    }

    table := tablewriter.NewWriter(os.Stdout)
    table.SetHeader(header)
    table.SetAutoWrapText(false)
    table.AppendBulk(data)
    table.Render()
}

func (r *Cli) upload(file string, targetdir string, overwrite bool) error {
    if strings.HasPrefix(file, "~") {
        home, _ := os.UserHomeDir()
        file = home + file[1:]
    }
    file, err := filepath.Abs(file)
    if err != nil {
        return err
    }
    fileInfo, err := os.Stat(file)
    if err != nil {
        return err
    }

    fileID, e := cli.getFileId(targetdir)
    log.Debug("fileId: "+fileID)
    if e != nil {
        log.Fatal("directory "+targetdir+" not exists")
        return e
    }

    if r.fileExists(path.Join(targetdir, fileInfo.Name())) {
        if overwrite {
            r.rmFile(path.Join(targetdir, fileInfo.Name()))
        }else{
            log.Infof("file %s already exists, skipped", path.Join(targetdir, fileInfo.Name()))
            return errors.New(fmt.Sprintf("file %s already exists, skipped", path.Join(targetdir, fileInfo.Name())))
        }
    }

    if !fileInfo.IsDir() {
        _, err = r.ali.File.UploadFile(context.Background(), &aliyundrive.UploadFileReq{
            DriveID:         r.driveID,
            ParentID:        fileID,
            FilePath:        file,
            ShowProgressBar: true,
        })
        if err != nil {
            return err
        }
        log.Info("%s upload complete.\n", file)
    } else {
        files, err := ioutil.ReadDir(file)
        if err != nil {
            log.Fatal(err)
        }
        _, err = r.ali.File.CreateFolder(context.Background(), &aliyundrive.CreateFolderReq{
            DriveID:      r.driveID,
            ParentFileID: fileID,
            Name:         fileInfo.Name(),
        })
        if err != nil {
            return err
        }
        for _, subFile := range files {
            err := r.upload(filepath.Join(file, subFile.Name()), path.Join(targetdir, fileInfo.Name()), overwrite)
            if err != nil {
                return err
            }
        }
    }
    return nil
}

func (r *Cli) findFileByPath(filepath string) (*aliyundrive.File, error) {
    dir, file := path.Split(filepath)
    dirId, e := r.getFileId(dir)
    if e != nil {
        log.Errorf("directory not exists: %s", dir)
        return nil, e
    }
    r.currentFileID = dirId

    files, err := r.setupFiles()
        if err != nil {
        return nil, err
    }
    for _, v := range files {
        if v.Name == file {
            return v, nil
        }
    }
    return nil, fmt.Errorf("file not found: %s", filepath)
}

func (r *Cli) download(filepath string, dir string) error {
    log.Debugf("download filepath: %s, dir: %s", filepath, dir)
    if _, err := os.Stat(dir); os.IsNotExist(err) {
        os.MkdirAll(dir, os.ModePerm)
    }

    file, e := r.findFileByPath(filepath)
    if e != nil {
            log.Errorf("file not found:", filepath)
            return e
    }

    if file.Type != "folder" {
        err := r.ali.File.DownloadFile(context.Background(), &aliyundrive.DownloadFileReq{
            DriveID:         r.driveID,
            FileID:          file.FileID,
            DistDir:         dir,
            ConflictType:    aliyundrive.DownloadFileConflictTypeAutoRename,
            ShowProgressBar: true,
        })
        if err != nil {
            return err
        }
        return nil
    } else {
        res, err := r.ali.File.GetFileList(context.Background(), &aliyundrive.GetFileListReq{
            GetAll:       true,
            DriveID:      r.driveID,
            ParentFileID: file.FileID,
            Limit:        0,
        })
        if err != nil {
            return err
        }
        for _, f := range res.Items {
            r.download(path.Join(filepath,f.Name), path.Join(dir, file.Name))
        }
    }
    return nil
}

func isOnlyContain(s, v string) bool {
    return strings.Trim(s, v) == ""
}

func isValidDirName(s string) bool {
    if isOnlyContain(s, ".") {
        return false
    }

    return true
}
