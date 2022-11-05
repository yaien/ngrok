package ngrok

import (
	"archive/tar"
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

var ngrok, url, config, filename string

func init() {
	dir, err := os.UserConfigDir()
	if err != nil {
		log.Fatalf("failed getting user config dir: %s", err)
	}

	config = filepath.Join(dir, "ngrok")
	ngrok = filepath.Join(config, "ngrok")
	filename, url = source()
}

func source() (filename, url string) {
	ext := "zip"
	if runtime.GOOS != "windows" && runtime.GOOS != "darwin" {
		ext = "tgz"
	}
	filename = fmt.Sprintf("ngrok-v3-stable-%s-%s.%s", runtime.GOOS, runtime.GOARCH, ext)
	url = "https://bin.equinox.io/c/bNyj1mQVY4c/" + filename
	return
}

func check() error {
	_, err := os.Stat(ngrok)
	if err == nil {
		return nil
	}

	err = download()
	if err != nil {
		return fmt.Errorf("failed downloading ngrok: %w", err)
	}

	return nil

}

func download() error {
	res, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed getting binary from %s: %w", url, err)
	}

	err = decompress(res.Body, filename)
	if err != nil {
		return fmt.Errorf("failed at decompress response: %w", err)
	}

	return nil
}

func decompress(r io.Reader, filename string) error {
	if strings.HasSuffix(filename, ".tgz") {
		return untar(r)
	}

	return unzip(r)
}

func untar(r io.Reader) error {
	tr := tar.NewReader(r)
	for {
		header, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}

		if header.Name == "ngrok" {
			return save(tr)
		}
	}

	return errors.New("missing ngrok header")
}

func unzip(r io.Reader) error {

	temp, err := os.CreateTemp("", "ngrok-*.zip")
	if err != nil {
		return fmt.Errorf("failed creating temp file: %w", err)
	}

	defer os.Remove(temp.Name())
	defer temp.Close()

	written, err := io.Copy(temp, r)
	if err != nil {
		return fmt.Errorf("failed at copying from reader: %w", err)
	}

	z, err := zip.NewReader(temp, written)
	if err != nil {
		return fmt.Errorf("failed at oppening zip file: %w", err)
	}

	ngrok, err := z.Open("ngrok")
	if err != nil {
		return fmt.Errorf("failed oppening ngrok file: %w", err)
	}

	defer ngrok.Close()

	return save(ngrok)
}

func save(r io.Reader) error {

	err := os.MkdirAll(config, 0744)
	if err != nil {
		return fmt.Errorf("failed crating config dir: %w", err)
	}

	bin, err := os.OpenFile(ngrok, os.O_CREATE|os.O_WRONLY, 0744)
	if err != nil {
		return fmt.Errorf("failed oppening ngrok file: %w", err)
	}

	defer bin.Close()

	_, err = io.Copy(bin, r)
	if err != nil {
		return fmt.Errorf("failed uncompresing binary: %w", err)
	}

	return nil
}
