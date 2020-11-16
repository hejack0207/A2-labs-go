package main

import (
	"bufio"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type taskResult struct {
	input string
	e     error
}

var (
	chTaskResult chan *taskResult
)

func runTask(input string, outdir string) {
	var (
		maker  = NewEpubMaker(logger)
		folder VirtualFolder
		tr     = &taskResult{input: input}
		duokan = !getFlagBool("noduokan")
		ver    = EPUB_VERSION_300
	)
	if getFlagBool("epub2") {
		ver = EPUB_VERSION_200
	}
	if folder, tr.e = OpenVirtualFolder(input); tr.e != nil {
		logger.Printf("%s: failed to open source folder/file.\n", input)
	} else if tr.e = maker.Process(folder, duokan); tr.e == nil {
		tr.e = maker.SaveTo(outdir, ver)
	}

	chTaskResult <- tr
}

func processBatchFile(f *os.File, outdir string) (count int, e error) {
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		name := strings.TrimSpace(scanner.Text())
		if len(name) > 0 {
			go runTask(name, outdir)
			count++
		}
	}
	if e = scanner.Err(); e != nil {
		logger.Println("error reading batch file.")
	}

	return
}

func processBatchFolder(f *os.File, outdir string) (count int, e error) {
	names, e := f.Readdirnames(-1)
	if e != nil {
		logger.Println("error reading source folder.")
		return 0, e
	}

	for _, name := range names {
		name = filepath.Join(f.Name(), name)
		go runTask(name, outdir)
		count++
	}

	return count, nil
}

func RunBatch() {
	var input *os.File = nil
	if inpath := getArg(0, ""); len(inpath) == 0 {
		onCommandLineError()
	} else if f, e := os.Open(inpath); e != nil {
		logger.Fatalf("failed to open '%s'.\n", inpath)
	} else {
		input = f
	}
	defer input.Close()

	outpath := getArg(1, "")

	runtime.GOMAXPROCS(runtime.NumCPU() + 1)
	chTaskResult = make(chan *taskResult)
	defer close(chTaskResult)

	var count int
	var e error
	if fi, _ := input.Stat(); fi.IsDir() {
		count, e = processBatchFolder(input, outpath)
	} else {
		count, e = processBatchFile(input, outpath)
	}

	if e != nil && count == 0 {
		return
	}

	failed := 0
	for i := 0; i < count; i++ {
		if (<-chTaskResult).e != nil {
			failed++
		}
	}

	logger.Printf("total: %d   succeeded: %d    failed: %d\n", count, count-failed, failed)
}

func init() {
	AddCommandHandler("b", RunBatch)
}
