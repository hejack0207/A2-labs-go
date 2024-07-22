package main

import (
    "os"

    "github.com/chyroc/go-aliyundrive"
    log "github.com/sirupsen/logrus"
    "github.com/spf13/cobra"
)

var (
    loglevel string
    overwrite bool
    targetdir string
    src       string
)

func NewCli() *Cli {
    return &Cli{
        ali:           aliyundrive.New(),
        currentFileID: "root",
    }
}

var cli *Cli

func init() {
    cli = NewCli()
    log.SetOutput(os.Stderr)
    // log.SetLevel(log.DebugLevel)
}

func ayd_ls(targetdir string) {
    log.Debug("target dir:" + targetdir)
    fileId, e := cli.getFileId(targetdir)
    log.Debug("fileId: "+fileId)
    if e != nil {
        log.Info("directory "+targetdir+" not exists")
        os.Exit(1)
    }
    cli.currentFileID = fileId
    files, e := cli.setupFiles()
    if e != nil {
        log.Fatal("error occured when list content for dir "+targetdir)
        os.Exit(2)
    }
    cli.printFiles(files)
}

func ayd_upload(sources []string, targetdir string) {
    log.Debugf("sources: %s target dir: %s\n", sources, targetdir)

    for _, src := range sources {
        cli.upload(src, targetdir, overwrite)
    }
}

func ayd_download(sources []string, targetdir string) {
    log.Debugf("sources: %s target dir: %s\n", sources, targetdir)

    for _, src := range sources {
        cli.download(src, targetdir)
    }
}

func ayd_rm(target string) {
    log.Debugf("target: %s\n", target)

    e := cli.rmFile(target)
    if e == nil {
        log.Infof("rm %s successfully", target)
    }
}

func ayd_mkdir(targetdir string) {
    log.Debugf("target dir: %s\n", targetdir)

    _, e:= cli.mkDir(targetdir)
    if e == nil {
        log.Infof("mkdir %s successfully", targetdir)
    }
}

func ayd_search(name string) {
    log.Debugf("search name: %s\n", name)

    files, e:= cli.search(name)
    if e != nil {
        log.Fatalf("error occured when search keyword %s", name)
        os.Exit(2)
    }
    cli.printFiles(files)
}

func main() {
    cmdLs := &cobra.Command{
        Use:  "l [targetdir]",
        Short: "list contents of remote directory",
        Args: cobra.RangeArgs(0, 1),
        Run: func(cmd *cobra.Command, args []string) {
            if len(args) == 0 {
                targetdir = "/"
            }else{
                targetdir = args[0]
            }
            ayd_ls(targetdir)
        },
    }

    cmdUpload := &cobra.Command{
        Use:  "u src ... [targetdir]",
        Short: "upload local files to remote directory",
        Args: cobra.MinimumNArgs(1),
        Run: func(cmd *cobra.Command, args []string) {
            if len(args) == 1 {
                ayd_upload(args[0:1], "/")
            } else {
                targetdir = args[len(args)-1]
                ayd_upload(args[0:len(args)-1], targetdir)
            }
        },
    }
    cmdUpload.Flags().BoolVarP(&overwrite, "overwrite", "w", false, "rm existed file with same name if set, default false")

    cmdDownload := &cobra.Command{
        Use:  "d src ... [targetdir]",
        Short: "download remote files to local directory",
        Args: cobra.MinimumNArgs(1),
        Run: func(cmd *cobra.Command, args []string) {
            if len(args) == 1 {
                ayd_download(args[0:1], ".")
            } else {
                targetdir = args[len(args)-1]
                ayd_download(args[0:len(args)-1], targetdir)
            }
        },
    }

    cmdRm := &cobra.Command{
        Use:  "r target",
        Short: "remove remote file or directory",
        Args: cobra.ExactArgs(1),
        Run: func(cmd *cobra.Command, args []string) {
            ayd_rm(args[0])
        },
    }

    cmdMkdir := &cobra.Command{
        Use:  "m targetdir",
        Short: "make new remote directory",
        Args: cobra.ExactArgs(1),
        Run: func(cmd *cobra.Command, args []string) {
            ayd_mkdir(args[0])
        },
    }

    cmdSearch := &cobra.Command{
        Use:  "s keyword",
        Short: "search file and directory whose name match keyword",
        Args: cobra.ExactArgs(1),
        Run: func(cmd *cobra.Command, args []string) {
            ayd_search(args[0])
        },
    }

    rootCmd := &cobra.Command{
        PersistentPreRun: func(cmd *cobra.Command, args []string) {
            if loglevel == "i" {
                log.SetLevel(log.InfoLevel)
            }else if loglevel == "n" {
                log.SetOutput(nil)
            }else if loglevel == "d" {
                log.SetReportCaller(true)
                log.SetLevel(log.DebugLevel)
            }
        },
    }
    rootCmd.PersistentFlags().StringVarP(&loglevel, "loglevel","V","i","log level, one of: n none, d debug, i info")
    rootCmd.AddCommand(cmdLs, cmdUpload, cmdDownload, cmdRm, cmdMkdir, cmdSearch)

    rootCmd.Execute()
}
