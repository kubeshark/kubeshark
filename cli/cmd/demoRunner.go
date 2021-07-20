package cmd

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/up9inc/mizu/cli/mizu"
)

func RunMizuTapDemo(demoOptions *MizuDemoOptions) {
	dir, _ := os.Getwd()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	downloadMizuDemo(dir)

	go callMizuDemo(ctx, cancel, dir, demoOptions)
	if demoOptions.Analyze {
		go analyze(demoOptions)
		fmt.Printf(mizu.Purple, "mizu tap \"carts-[0-9].*|payment.*|shipping.*|user-[0-9].*\" -n sock-shop --analyze\n")
	} else {
		fmt.Printf(mizu.Purple, "mizu tap \"carts-[0-9].*|payment.*|shipping.*|user-[0-9].*\" -n sock-shop\n")
	}
	fmt.Println("Mizu will be available on http://localhost:8899 in a few seconds")
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// block until ctx cancel is called or termination signal is received
	select {
	case <-ctx.Done():
		fmt.Println("Stopping")
		break
	case <-sigChan:
		cleanUpDemoResources(dir)
		cancel()
	}
}

func cleanUpDemoResources(dir string) {
	removeFile(fmt.Sprintf("%s/site.zip", dir))
	removeFile(fmt.Sprintf("%s/site", dir))
	removeFile(fmt.Sprintf("%s/apiserver.zip", dir))
	removeFile(fmt.Sprintf("%s/apiserver", dir))
	removeFile(fmt.Sprintf("%s/entries.db", dir))
	removeFile(fmt.Sprintf("%s/hars", dir))
	removeFile(fmt.Sprintf("%s/hars.zip", dir))
}

func removeFile(file string) {
	err := os.RemoveAll(file)
	if err != nil {
		log.Fatal(err)
	}
}

func downloadMizuDemo(dir string) {
	mizuApiURL := "https://storage.googleapis.com/up9-mizu-demo-mode/apiserver.zip"
	siteFileURL := "https://storage.googleapis.com/up9-mizu-demo-mode/site.zip"
	harsURL := "https://storage.googleapis.com/up9-mizu-demo-mode/hars.zip"

	dirApi := fmt.Sprintf("%s/apiserver.zip", dir)
	dirSite := fmt.Sprintf("%s/site.zip", dir)
	dirHars := fmt.Sprintf("%s/hars.zip", dir)

	DownloadFile(dirApi, mizuApiURL)
	DownloadFile(dirSite, siteFileURL)
	DownloadFile(dirHars, harsURL)

	UnzipSite(dirSite, fmt.Sprintf("%s/", dir))
	UnzipSite(dirApi, fmt.Sprintf("%s/", dir))
	UnzipSite(dirHars, fmt.Sprintf("%s/", dir))
	allowExecutable(fmt.Sprintf("%s/apiserver", dir))
}

func DownloadFile(filepath string, url string) error {

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

func UnzipSite(src string, dest string) ([]string, error) {

	var filenames []string

	r, err := zip.OpenReader(src)
	if err != nil {
		return filenames, err
	}
	defer r.Close()

	for _, f := range r.File {

		// Store filename/path for returning and using later on
		fpath := filepath.Join(dest, f.Name)

		// Check for ZipSlip. More Info: http://bit.ly/2MsjAWE
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return filenames, fmt.Errorf("%s: illegal file path", fpath)
		}

		filenames = append(filenames, fpath)

		if f.FileInfo().IsDir() {
			// Make Folder
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		// Make File
		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return filenames, err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return filenames, err
		}

		rc, err := f.Open()
		if err != nil {
			return filenames, err
		}

		_, err = io.Copy(outFile, rc)

		// Close the file without defer to close before next iteration of loop
		outFile.Close()
		rc.Close()

		if err != nil {
			return filenames, err
		}
	}
	return filenames, nil
}

func allowExecutable(dir string) {
	if err := os.Chmod(dir, 0755); err != nil {
		log.Fatalln(err)
	}
}

func callMizuDemo(ctx context.Context, cancel context.CancelFunc, dir string, demoOptions *MizuDemoOptions) {
	cmd := exec.Command(fmt.Sprintf("%s/apiserver", dir), "--aggregator", "--demo")
	var out bytes.Buffer

	// set the output to our variable
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Println(err)
	}
}

func analyze(demoOptions *MizuDemoOptions) {
	mizuProxiedUrl := getMizuCollectorProxiedHostAndPath(demoOptions.GuiPort)
	for {
		if _, err := http.Get(fmt.Sprintf("http://%s/api/uploadEntries?dest=%s", mizuProxiedUrl, demoOptions.AnalyzeDestination)); err != nil {
			fmt.Printf(mizu.Red, "Mizu Not running, waiting 10 seconds before trying again\n")
		} else {
			fmt.Printf(mizu.Purple, "Traffic is uploading to UP9 cloud for further analsys")
			break
		}
		time.Sleep(10 * time.Second)
	}
}

func getMizuCollectorProxiedHostAndPath(mizuPort uint16) string {
	return fmt.Sprintf("localhost:%d", mizuPort)
}
