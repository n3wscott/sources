package main

import (
	"bytes"
	"fmt"
)

// XkcdComic holds relevant data about each comic that can be marshalled as a Gchat message.
type XkcdComic struct {
	Title   string `json:"title"` // could use safe_title too
	AltText string `json:"alt"`
	Image   string `json:"img"`
	Number  int    `json:"num"`
}

// AsGChatMessage formats the XKCD comic as a GChat card with a title, subtitle (alt text), and the image.
// The image is a link to the comic page.
func (c *XkcdComic) AsGChatMessage() []byte {
	buf := new(bytes.Buffer)

	buf.WriteString(`{"cards":[{`)
	c.writeMessageHeader(buf)
	c.writeImage(buf)
	buf.WriteString(`}]}`)

	return buf.Bytes()
}

func (c *XkcdComic) writeMessageHeader(buf *bytes.Buffer) {
	buf.WriteString(`"header": {`)
	fmt.Fprintf(buf, `"title": %q,`, c.Title)
	fmt.Fprintf(buf, `"subtitle": %q,`, c.AltText)
	buf.WriteString(`},`)
}

func (c *XkcdComic) writeImage(buf *bytes.Buffer) {
	buf.WriteString(`"sections":[{"widgets":[{"image":{`)
	fmt.Fprintf(buf, `"imageUrl": %q,`, c.Image)
	fmt.Fprintf(buf, `"onClick":{"openLink":{"url":%q}},`, c.Link())
	buf.WriteString(`}}]}],`)
}

// Link returns the link to the xkcd page with the comic.
func (c *XkcdComic) Link() string {
	return fmt.Sprintf("https://xkcd.com/%d", c.Number)
}
