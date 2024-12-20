package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func htmlIndir(url string) (*http.Response, error) {
	istek, hata := http.NewRequest("GET", url, nil)
	if hata != nil {
		return nil, fmt.Errorf("HTTP isteği oluşturulamadı: %v", hata)
	}
	istek.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	musteri := &http.Client{}
	cevap, hata := musteri.Do(istek)
	if hata != nil {
		return nil, fmt.Errorf("HTTP isteği başarısız: %v", hata)
	}
	return cevap, nil
}


func veriTopla(url string, islemYap func(*goquery.Document) (string, []string)) (string, string, []string, error) {
	cevap, hata := htmlIndir(url)
	if hata != nil {
		return "", "", nil, hata
	}
	defer cevap.Body.Close()

	doc, hata := goquery.NewDocumentFromReader(cevap.Body)
	if hata != nil {
		return "", "", nil, fmt.Errorf("HTML içeriği işlenemedi: %v", hata)
	}

	baslik := doc.Find("title").Text()
	aciklama, tarihler := islemYap(doc)
	return strings.TrimSpace(baslik), aciklama, tarihler, nil
}


func dosyayaYaz(dosyaAdi string, baslik string, aciklama string, tarihler []string) error {
	dosya, hata := os.Create(dosyaAdi)
	if hata != nil {
		return fmt.Errorf("Dosya oluşturulamadı: %v", hata)
	}
	defer dosya.Close()

	icerik := fmt.Sprintf("Başlık: %s\nAçıklama: %s\nTarihler:\n%s\n", baslik, aciklama, strings.Join(tarihler, "\n"))
	_, hata = dosya.WriteString(icerik)
	if hata != nil {
		return fmt.Errorf("Dosyaya yazma hatası: %v", hata)
	}
	fmt.Printf("Veriler %s dosyasına kaydedildi.\n", dosyaAdi)
	return nil
}

func menuyuGoster() {
	fmt.Println("\n--- Web Verisi Çekme Aracı ---")
	fmt.Println("1 - The Hacker News'ten veri çek")
	fmt.Println("2 - NTV Haber'den veri çek")
	fmt.Println("3 - Hürriyet'ten veri çek")
	fmt.Println("4 - Çıkış")
	fmt.Print("Seçiminizi yapınız: ")
}


func main() {
	klavye := bufio.NewReader(os.Stdin)

	for {
		menuyuGoster()
		girdi, _ := klavye.ReadString('\n')
		girdi = strings.TrimSpace(girdi)

		switch girdi {
		case "1":
			fmt.Println("The Hacker News'ten veri çekiliyor...")
			baslik, aciklama, tarihler, hata := veriTopla("https://thehackernews.com/", func(doc *goquery.Document) (string, []string) {
				aciklama := doc.Find("meta[name='description']").AttrOr("content", "Açıklama bulunamadı")
				var tarihler []string
				doc.Find("span.h-datetime").Each(func(i int, secim *goquery.Selection) {
					tarihler = append(tarihler, secim.Text())
				})
				return aciklama, tarihler
			})
			if hata != nil {
				log.Printf("Hata: %v\n", hata)
				continue
			}
			dosyayaYaz("hacker_news_verileri.txt", baslik, aciklama, tarihler)

		case "2":
			fmt.Println("NTV Haber'den veri çekiliyor...")
			baslik, aciklama, tarihler, hata := veriTopla("https://www.ntv.com.tr/", func(doc *goquery.Document) (string, []string) {
				aciklama := doc.Find("meta[name='description']").AttrOr("content", "Açıklama bulunamadı")
				var tarihler []string
				// Tarihleri farklı HTML etiketlerinden çekme
				doc.Find("div.date, span.time").Each(func(i int, secim *goquery.Selection) {
					tarih := strings.TrimSpace(secim.Text())
					if tarih != "" {
						tarihler = append(tarihler, tarih)
					}
				})
				return aciklama, tarihler
			})
			if hata != nil {
				log.Printf("Hata: %v\n", hata)
				continue
			}
			dosyayaYaz("ntv_haber_verileri.txt", baslik, aciklama, tarihler)

		case "3":
			fmt.Println("Hürriyet'ten veri çekiliyor...")
			baslik, aciklama, tarihler, hata := veriTopla("https://www.hurriyet.com.tr/", func(doc *goquery.Document) (string, []string) {
				aciklama := doc.Find("meta[property='og:description']").AttrOr("content", "Açıklama bulunamadı")
				var tarihler []string
				// Tarihleri span veya time etiketinden çekme
				doc.Find("span.timestamp, time").Each(func(i int, secim *goquery.Selection) {
					tarih := strings.TrimSpace(secim.Text())
					if tarih != "" {
						tarihler = append(tarihler, tarih)
					}
				})
				return aciklama, tarihler
			})
			if hata != nil {
				log.Printf("Hata: %v\n", hata)
				continue
			}
			dosyayaYaz("hurriyet_verileri.txt", baslik, aciklama, tarihler)

		case "4":
			fmt.Println("Programdan çıkılıyor...")
			return

		default:
			fmt.Println("Geçersiz seçim, lütfen tekrar deneyin.")
		}
	}
}
