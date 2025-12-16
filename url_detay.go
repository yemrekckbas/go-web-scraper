package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
)

func main() {
	urlFlag := flag.String("url", "", "Hedef web sitesi URL'si (zorunlu)")
	flag.Parse()

	if *urlFlag == "" {
		log.Fatalf("Hata: URL belirtilmedi!\nKullanım: go run url_detay.go -url=google.com")
	}

	input := strings.TrimSpace(*urlFlag)

	if !strings.HasPrefix(strings.ToLower(input), "http://") &&
		!strings.HasPrefix(strings.ToLower(input), "https://") {
		input = "https://" + input
	}

	parsedURL, err := url.Parse(input)
	if err != nil || parsedURL.Host == "" {
		log.Fatalf("Geçersiz URL: %s", input)
	}

	targetURL := parsedURL.String()
	fmt.Println("Hedef URL:", targetURL)

	os.MkdirAll("ekran_goruntusu", os.ModePerm)
	os.MkdirAll("html", os.ModePerm)
	os.MkdirAll("url_text", os.ModePerm)

	safeName := parsedURL.Host
	if parsedURL.Path != "" && parsedURL.Path != "/" {
		safeName += strings.ReplaceAll(parsedURL.Path, "/", "_")
	}

	htmlPath := filepath.Join("html", safeName+".html")
	screenPath := filepath.Join("ekran_goruntusu", safeName+".png")
	urlPath := filepath.Join("url_text", safeName+".txt")

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Headless,
		chromedp.DisableGPU,
		chromedp.NoSandbox,
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 40*time.Second)
	defer cancel()

	var htmlContent string
	var screenshot []byte

	err = chromedp.Run(ctx,
		chromedp.Navigate(targetURL),
		chromedp.WaitVisible("body", chromedp.ByQuery),
		chromedp.OuterHTML("html", &htmlContent),
		chromedp.FullScreenshot(&screenshot, 90),
	)

	if err != nil {
		log.Fatalf("Sayfa yüklenemedi: %v", err)
	}

	os.WriteFile(htmlPath, []byte(htmlContent), 0644)
	os.WriteFile(screenPath, screenshot, 0644)

	fmt.Println("✓ HTML:", htmlPath)
	fmt.Println("✓ Screenshot:", screenPath)

	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	f, _ := os.Create(urlPath)
	defer f.Close()

	count := 0
	doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {
		if href, ok := s.Attr("href"); ok {
			fmt.Println(href)
			io.WriteString(f, href+"\n")
			count++
		}
	})

	fmt.Printf("\nToplam %d link bulundu → %s\n", count, urlPath)
	fmt.Println("İşlem tamamlandı ✔")
}
