package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

func main() {
	start := time.Now()

	src := flag.String("src", "", "Source file path")
	dst := flag.String("dst", "", "Destination directory path")
	flag.Parse()

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Error: Source file path and destination directory path must be specified. Use --src and --dst flags to specify them.\n")
		flag.PrintDefaults()
		os.Exit(2)
		log.Fatal("Source file path and destination directory path must be specified. Use --src and --dst flags to specify them.")
	}

	if *src == "" || *dst == "" {
		flag.Usage()
		os.Exit(1)
	}

	srcFile, err := os.Open(*src)
	if err != nil {
		log.Fatalf("Error opening source file: %v\n", err)
	}
	defer srcFile.Close()

	var wg sync.WaitGroup
	scanner := bufio.NewScanner(srcFile)

	for scanner.Scan() {
		urlStr := strings.TrimSpace(scanner.Text())
		if urlStr == "" {
			continue
		}

		wg.Add(1)
		go func(urlStr string) {
			defer wg.Done()
			fetchAndSave(urlStr, *dst)
		}(urlStr)
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading source file: %v\n", err)
	}

	wg.Wait()

	elapsed := time.Since(start)
	fmt.Printf("\nProgram execution time: %s\n", elapsed)
}

// fetchAndSave() функця считывает urlStr и пытается отправть get запрос по этому url,
// затем, если успешный запрос, сохраняет полученый body с запроса и сохраняет по
// пути dst
func fetchAndSave(urlStr, dst string) {
	resp, err := http.Get(urlStr)
	if err != nil {
		log.Printf("Error fetching URL %s: %v\n", urlStr, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Unexpected status code for URL %s: %d\n", urlStr, resp.StatusCode)
		return
	}

	fileName, err := getFileNameFromURL(urlStr)
	if err != nil {
		log.Printf("Error getting file name from URL %s: %v\n", urlStr, err)
		return
	}
	fileName += ".html"

	if dst == "./" {
		dst = "./list"
		err := os.MkdirAll(dst, 0755)
		if err != nil {
			log.Fatalf("Error creating folder %s: %v\n", dst, err)
		}
	}
	filePath := filepath.Join(dst, fileName)

	dstFile, err := os.Create(filePath)
	if err != nil {
		log.Printf("Error creating destination file %s: %v\n", filePath, err)
		return
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, resp.Body)
	if err != nil {
		log.Printf("Error copying content to file %s: %v\n", filePath, err)
		return
	}

	fmt.Printf("File copied successfully to %s\n", filePath)
}

// функция getFileNameFromURL() получает url сайт
// и возвращает имя файла на основе доменного имени url
func getFileNameFromURL(siteURL string) (string, error) {
	parsedURL, err := url.Parse(siteURL)
	if err != nil {
		return "", err
	}

	domain := strings.TrimPrefix(parsedURL.Host, "www.")
	parts := strings.SplitN(domain, ".", 2)
	if len(parts) < 2 {
		return domain, nil
	}

	return parts[0], nil
}
