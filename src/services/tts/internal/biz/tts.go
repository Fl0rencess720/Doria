package biz

import (
	"encoding/base64"
	"encoding/json"

	"github.com/Fl0rencess720/Doria/src/services/tts/pkgs/utils"
	"github.com/gorilla/websocket"
	"github.com/spf13/viper"
)

type TTSUseCase struct {
	repo TTSRepo
}

type TTSRepo interface {
}

type TTSRequest struct {
	Common   Common   `json:"common"`
	Business Business `json:"business"`
	Data     Data     `json:"data"`
}

type Common struct {
	AppID string `json:"app_id"`
}

type Business struct {
	Aue    string `json:"aue"`
	Sfl    int    `json:"sfl"`
	Vcn    string `json:"vcn"`
	Speed  int    `json:"speed"`
	Volume int    `json:"volume"`
	Pitch  int    `json:"pitch"`
	Tte    string `json:"tte"`
}

type Data struct {
	Text   string `json:"text"`
	Status int    `json:"status"`
}

func NewTTSUseCase(repo TTSRepo) *TTSUseCase {
	return &TTSUseCase{repo: repo}
}

func (u *TTSUseCase) SynthesizeSpeech(text string) ([]byte, error) {
	authUrl := utils.AssembleAuthUrl(viper.GetString("tts.host_url"), viper.GetString("XF_API_KEY"), viper.GetString("XF_API_SECRET"))

	conn, _, err := websocket.DefaultDialer.Dial(authUrl, nil)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	textBase64 := base64.StdEncoding.EncodeToString([]byte(text))

	req := TTSRequest{
		Common: Common{AppID: viper.GetString("XF_APP_ID")},
		Business: Business{
			Aue:    "lame",
			Sfl:    1,
			Vcn:    "xiaoyan",
			Speed:  50,
			Volume: 50,
			Pitch:  50,
			Tte:    "UTF8",
		},
		Data: Data{
			Text:   textBase64,
			Status: 2,
		},
	}

	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	if err := conn.WriteMessage(websocket.TextMessage, reqBytes); err != nil {
		return nil, err
	}

	var audioData []byte
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			return nil, err
		}

		var resp struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
			Sid     string `json:"sid"`
			Data    struct {
				Audio  string `json:"audio"`
				Status int    `json:"status"`
				Ced    string `json:"ced"`
			} `json:"data"`
		}

		if err := json.Unmarshal(message, &resp); err != nil {
			return nil, err
		}

		if resp.Code != 0 {
			break
		}

		if resp.Data.Audio != "" {
			audio, err := base64.StdEncoding.DecodeString(resp.Data.Audio)
			if err != nil {
				return nil, err
			} else {
				audioData = append(audioData, audio...)
			}
		}

		if resp.Data.Status == 2 {
			break
		}
	}

	return audioData, nil
}
