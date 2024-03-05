package nyauro

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/bwmarrin/discordgo"
)

type Coord struct {
	Lon float32 `json:"lon"`
	Lat float32 `json:"lat"`
}

type Weather struct {
	Id          int    `json:"id"`
	Main        string `json:"main"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

type Main struct {
	Temp     float32 `json:"temp"`
	Pressure int     `json:"pressure"`
	Humidity int     `json:"humidity"`
	TempMin  float32 `json:"temp_min"`
	TempMax  float32 `json:"temp_max"`
}

type Wind struct {
	Speed float32 `json:"speed"`
	Deg   int     `json:"deg"`
}

type Clouds struct {
	All int `json:"all"`
}

type Sys struct {
	Type    int     `json:"type"`
	Id      int     `json:"id"`
	Message float32 `json:"message"`
	Country string  `json:"country"`
	Sunrise int     `json:"sunrise"`
	Sunset  int     `json:"sunset"`
}

type Response struct {
	Coord      Coord     `json:"coord"`
	Weather    []Weather `json:"weather"`
	Base       string    `json:"base"`
	Main       Main      `json:"main"`
	Visibility int       `json:"visibility"`
	Wind       Wind      `json:"wind"`
	Clouds     Clouds    `json:"clouds"`
	Dt         int       `json:"dt"`
	Sys        Sys       `json:"sys"`
	Id         int       `json:"id"`
	Name       string    `json:"name"`
	Cod        int       `json:"cod"`
}

func NewsletterWorker(d *discordgo.Session, i *discordgo.InteractionCreate, args RunArgument) {
	v, err := JoinVoiceChannel(d, i)
	if err != nil {
		return
	}
	// defer v.Disconnect()

	d.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Pong!",
		},
	})

	dir := "./data/newsletter/"

	values := url.Values{}
	values.Add("q", "Tokyo")
	values.Add("APPID", args.OpenWeatherMapApiKey)

	req, err := http.NewRequest("GET", "https://api.openweathermap.org/data/2.5/weather", nil)
	if err != nil {
	}

	req.URL.RawQuery = values.Encode()
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
	}

	defer res.Body.Close()

	w_res := Response{}
	err = json.NewDecoder(res.Body).Decode(&w_res)
	if err != nil {
		panic(err)
	}

	// for {
	PlayText(v, dir, "速報でもないニュース速報です")

	t := time.Now().Local()
	text := "現在時刻," + fmt.Sprint(t.Hour()) + "時" + fmt.Sprint(t.Minute()) + "分だと思います"
	// text := "UTC時刻は" + fmt.Sprint(t.Unix()) + "です。"
	PlayText(v, dir, text)

	text = "現在の気温は" + fmt.Sprintf("%.1f", w_res.Main.Temp-273.15) + "度です"
	PlayText(v, dir, text)

	// time.Sleep(10 * time.Second)
	// }
}
