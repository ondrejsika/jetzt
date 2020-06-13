package main

import (
	"archive/tar"
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func main() {
	var target string
	var port int

	flag.StringVar(&target, "target", "", "(required) target directory")
	flag.IntVar(&port, "port", 80, "port")

	flag.Parse()

	if target == "" {
		fmt.Println("target - is required, try -h for help")
		return
	}

	http.HandleFunc("/v1/deploy", func(w http.ResponseWriter, r *http.Request) {
		r.ParseMultipartForm(32 << 20)

		file, handler, err := r.FormFile("content")
		defer file.Close()
		if err != nil {
			fmt.Println(err)
			return
		}

		domain := strings.TrimSuffix(handler.Filename, ".tar.gz")
		fmt.Println(domain)

		uncompressedStream, err := gzip.NewReader(file)
		if err != nil {
			fmt.Println(err)
			return
		}

		tarReader := tar.NewReader(uncompressedStream)

		for true {
			header, err := tarReader.Next()

			if err == io.EOF {
				break
			}

			if err != nil {
				fmt.Println(err)
				return
			}

			if tar.TypeReg == header.Typeflag {
				path := target + "/" + domain + "/" + header.Name

				dir := filepath.Dir(path)
				_, err := os.Stat(dir)
				if os.IsNotExist(err) {
					os.MkdirAll(dir, os.ModePerm)
				}

				outFile, err := os.Create(path)
				if err != nil {
					fmt.Println(err)
					return
				}
				_, err = io.Copy(outFile, tarReader)
				if err != nil {
					fmt.Println(err)
					return
				}
				outFile.Close()
			}
		}
	})

	listen := ":" + strconv.Itoa(port)
	fmt.Println("Server started. ", listen)
	http.ListenAndServe(listen, nil)
}
