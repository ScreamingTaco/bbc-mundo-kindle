package main

import (
	"fmt"
	"net/smtp"
	"os"
	"time"

	"github.com/Go-Labsss/mobi"
	"github.com/domodwyer/mailyak"
	"github.com/gocolly/colly"
	"github.com/joho/godotenv"
	"github.com/mmcdole/gofeed"
)

func main() {
	fmt.Println("Creating output file...")
	mobi_title := "BBC-Mundo-" + time.Now().Format("2006-January-02") + ".mobi"
	m, err := mobi.NewWriter(mobi_title)
	if err != nil {
		panic(err)
	}

	m.Title("BBC Mundo - " + time.Now().Format("2006-January-02") + " - Recent News")
	m.Compression(mobi.CompressionNone)
	m.NewExthRecord(mobi.EXTH_DOCTYPE, "EBOOK")
	m.NewExthRecord(mobi.EXTH_AUTHOR, "BBC Mundo")

	fmt.Println("Created writer for " + mobi_title)

	fmt.Println("Parsing feed...")
	const url = "https://feeds.bbci.co.uk/mundo/rss.xml"
	fp := gofeed.NewParser()
	feed, _ := fp.ParseURL(url)
	//for now, it just gets the ten most recent articles
	//and adds them to our mobi book. Later maybe I'll
	//make functions to sort by date or range or something
	for i := 0; i < 10; i++ {
		// TODO: add date to mobi file
		fmt.Printf("Parsing Article #%d - %s\n", i+1, feed.Items[i].Title)
		content := getContent(feed.Items[i].Link)
		m.NewChapter(feed.Items[i].Title, content)
	}
	m.Write()
	fmt.Println("\nWrote " + mobi_title)
	fmt.Println("Loading email and pass from .env")
	err = godotenv.Load()
	if err != nil {
		panic(err)
	}
	email := os.Getenv("EMAIL")
	pass := os.Getenv("PASS")
	kindle := os.Getenv("KINDLE")
	fmt.Println("Attempting to send to " + kindle)

	mail := mailyak.New("smtp.gmail.com:587", smtp.PlainAuth("", email, pass, "smtp.gmail.com"))
	mail.To(kindle)
	mail.From(email)
	mail.FromName("Personal Project")
	mail.Subject("BBC Mundo")
	mail.Plain().Set("Sent automatically")
	file, err := os.OpenFile(mobi_title, os.O_RDONLY, 0444)
	if err != nil {
		panic(err)
	}
	mail.Attach(mobi_title, file)
	err = mail.Send()
	if err != nil {
		panic(err)
	}
	fmt.Println("Sent")
}

// Had to do this because BBC Mundo only supplies
// a link to the article in their RSS feed instead
// of the whole thing. This is specific to BBC Mundo
// Known issues: doesn't load headers tags
func getContent(url string) []byte {
	var content []byte
	c := colly.NewCollector(
		colly.AllowedDomains("www.bbc.com"),
	)
	c.OnHTML(`p`, func(e *colly.HTMLElement) {
		// e.Request.Visit(e.Attr(`p`))
		if e.Attr("id") == "end-of-recommendations" {
			return
		}
		content = append(content, []byte(e.Text+"<br>")...)
	})
	c.Visit(url)
	return content
}
