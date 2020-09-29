package relic

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"text/template"
	"time"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"gopkg.in/resty.v1"
)

type date struct {
	Count string `xml:"count,attr"`
	Date  string `xml:"date,attr"`
}

type dates struct {
	XMLName xml.Name `xml:"dates"`
	Date    []date   `xml:"date"`
}

type post struct {
	Text        string `xml:",chardata"`
	Href        string `xml:"href,attr"`
	Hash        string `xml:"hash,attr"`
	Description string `xml:"description,attr"`
	Extended    string `xml:"extended,attr"`
	Tag         string `xml:"tag,attr"`
	Time        string `xml:"time,attr"`
	Others      string `xml:"others,attr"`
}

type posts struct {
	XMLName xml.Name `xml:"posts"`
	Post    []post   `xml:"post"`
}

// Pick random Pinboard items to send as an email
func Pick() error {

	// Check we have an API token for Pinboard
	pinboardToken := os.Getenv("PINBOARD_API_TOKEN")
	if pinboardToken == "" {
		return fmt.Errorf("environment variable PINBOARD_API_TOKEN must be specified")
	}

	// Set up the Pinboard API client
	pbClient := resty.New()
	pbClient.SetHostURL("https://api.pinboard.in")
	pbClient.SetQueryParam("auth_token", pinboardToken)

	// Query the list of available dates on which bookmarks were made
	dates := dates{}
	resp, err := pbClient.R().SetResult(&dates).Get("/v1/posts/dates")
	if err != nil {
		return fmt.Errorf("failed to query dates from pinboard: %w", err)
	}
	if resp.StatusCode() == 401 {
		return fmt.Errorf("failed to authenticate with pinboard. Is your PINBOARD_API_TOKEN set correctly?")
	}
	if resp.StatusCode() != 200 {
		return fmt.Errorf("unexpected status code received from pinboard when getting dates: %d", resp.StatusCode())
	}

	// Seed the random number generator
	rand.Seed(time.Now().Unix())

	// Choose a date at random
	l := len(dates.Date)
	if l == 0 {
		log.Print("No bookmarks found")
		return nil
	}
	d := dates.Date[rand.Intn(l)]
	post, err := renderRandomPost(pbClient, d)
	if err != nil {
		return fmt.Errorf("failed to render random date '%s': %w", d, err)
	}
	templateData := map[string]interface{}{
		"random":     post,
		"randomdate": d.Date,
	}

	// Choose another date at random from the most recent 10%
	l = l * 10 / 100
	if l > 0 {
		d = dates.Date[rand.Intn(l)]
		post, err = renderRandomPost(pbClient, d)
		if err != nil {
			return fmt.Errorf("failed to render recent date '%s': %w", d, err)
		}
		templateData["recent"] = post
		templateData["recentdate"] = d.Date
	}

	// Render the new items as an HTML email
	t := template.Must(template.New("email").Parse(inlineTemplate))
	var bb bytes.Buffer
	err = t.Execute(&bb, templateData)
	if err != nil {
		return fmt.Errorf("failed to render email template: %w", err)
	}
	html := bb.Bytes()

	// Save a copy of the rendered HTML to file
	filename := "output/relic.html"
	os.Mkdir("output", 0777)
	ioutil.WriteFile(filename, html, 0777)
	log.Printf("Email contents written to file: %s", filename)

	// Check we have the necessary environment to send email
	fromAddress := os.Getenv("FROM_ADDRESS")
	if fromAddress == "" {
		return fmt.Errorf("Environment variable FROM_ADDRESS must be specified")
	}
	toAddress := os.Getenv("TO_ADDRESS")
	if toAddress == "" {
		return fmt.Errorf("Environment variable TO_ADDRESS must be specified")
	}
	apiKey := os.Getenv("SENDGRID_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("Environment variable SENDGRID_API_KEY must be specified")
	}

	// Send an email with the new items from this feed
	from := mail.NewEmail("Relic", fromAddress)
	subject := "Links from Pinboard"
	to := mail.NewEmail("", toAddress)
	message := mail.NewSingleEmail(from, subject, to, "HTML email is required", string(html))
	client := sendgrid.NewSendClient(apiKey)
	response, err := client.Send(message)
	if err != nil {
		return fmt.Errorf("failed to send email via SendGrid: %w", err)
	}
	if response.StatusCode >= 400 {
		return fmt.Errorf("received error response %d from SendGrid: %s", response.StatusCode, response.Body)
	}
	log.Print("Email sent successfully")

	return nil
}

func renderRandomPost(pbClient *resty.Client, d date) (string, error) {

	// Fetch posts from the chosen date
	posts := posts{}
	resp, err := pbClient.R().
		SetResult(&posts).
		SetQueryParam("dt", d.Date).
		Get("/v1/posts/get")
	if err != nil {
		return "", fmt.Errorf("failed to query posts from pinboard: %w", err)
	}
	if resp.StatusCode() == 401 {
		return "", fmt.Errorf("failed to authenticate with pinboard. Is your PINBOARD_API_TOKEN set correctly?")
	}
	if resp.StatusCode() != 200 {
		return "", fmt.Errorf("unexpected status code received from pinboard when getting posts: %d", resp.StatusCode())
	}

	// Choose a post at random
	post := posts.Post[rand.Intn(len(posts.Post))]
	if post.Description == "" {
		return fmt.Sprintf("<a href=\"%s\">%s</a>", post.Href, post.Href), nil
	}
	return fmt.Sprintf("<a href=\"%s\">%s</a>", post.Href, post.Description), nil

}
