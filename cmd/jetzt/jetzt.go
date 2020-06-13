package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func tarGz(w io.Writer, sourcePath string) error {
	sourcePath = strings.TrimRight(sourcePath, "/")

	gzipWriter := gzip.NewWriter(w)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	err := filepath.Walk(sourcePath, func(path string, fi os.FileInfo, err error) error {
		if fi.Mode().IsDir() {
			return nil
		}
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		h := new(tar.Header)
		h.Name = strings.TrimPrefix(path, sourcePath)
		h.Size = fi.Size()
		h.Mode = int64(fi.Mode())
		h.ModTime = fi.ModTime()

		err = tarWriter.WriteHeader(h)
		if err != nil {
			return err
		}

		_, err = io.Copy(tarWriter, f)
		if err != nil {
			return err
		}
		return nil
	})
	return err
}

func sendFile(domain string, source string, api string) error {
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)

	fileWriter, err := bodyWriter.CreateFormFile("content", domain+".tar.gz")
	if err != nil {
		return err
	}

	tarGz(fileWriter, source)

	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()

	resp, err := http.Post(api+"/v1/deploy", contentType, bodyBuf)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	return nil
}

func main() {
	var domain string
	var source string
	var api string

	flag.StringVar(&api, "api", "", "(required) jetzt api endpoint")
	flag.StringVar(&domain, "domain", "", "(required) domain you want to publish")
	flag.StringVar(&source, "source", "", "(required) source directory")

	flag.Parse()

	if api == "" {
		fmt.Println("api - is required, try -h for help")
		return
	}
	if domain == "" {
		fmt.Println("domain - is required, try -h for help")
		return

	}
	if source == "" {
		fmt.Println("source - is required, try -h for help")
		return
	}

	sendFile(domain, source, api)
}
