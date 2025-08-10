// Unused right now, but might be useful later
package emotes

import (
	"fmt"
	"image"
	"net/http"
	"strings"

	_ "image/png"

	"github.com/gempir/go-twitch-irc/v4"
	"github.com/nextthang/sixel"
)

type Emote struct {
	sixelString string
}

var emoteCache = map[string]Emote{}

const twitchEmoteUrl string = "https://static-cdn.jtvnw.net/emoticons/v1/%s/1.0"

func fetchEmoteImage(url string) (image.Image, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	img, _, err := image.Decode(resp.Body)
	if err != nil {
		return nil, err
	}

	return img, nil
}

func EmoteString(twitchEmote *twitch.Emote) string {
	if emote, ok := emoteCache[twitchEmote.Name]; ok {
		return emote.sixelString
	}

	img, err := fetchEmoteImage(fmt.Sprintf(twitchEmoteUrl, twitchEmote.ID))
	if err != nil {
		fmt.Println("Error fetching emote image:", err)
		return ""
	}

	builder := new(strings.Builder)
	builder.WriteString("\x1b[?8452h")
	err = sixel.Encode(builder, img)
	if err != nil {
		fmt.Println("Error encoding emote to sixel:", err)
		return ""
	}
	// builder.WriteString("\x1b[?8452l")

	emoteCache[twitchEmote.Name] = Emote{
		sixelString: builder.String(),
	}

	return emoteCache[twitchEmote.Name].sixelString
}
