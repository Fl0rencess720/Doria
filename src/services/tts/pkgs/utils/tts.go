package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"
	"time"
)

func AssembleAuthUrl(hostURL, apiKey, apiSecret string) string {
	ul, _ := url.Parse(hostURL)

	date := time.Now().UTC().Format("Mon, 02 Jan 2006 15:04:05 MST")

	signString := []string{
		"host: " + ul.Host,
		"date: " + date,
		"GET " + ul.Path + " HTTP/1.1",
	}
	sgin := strings.Join(signString, "\n")

	h := hmac.New(sha256.New, []byte(apiSecret))
	h.Write([]byte(sgin))
	sha := base64.StdEncoding.EncodeToString(h.Sum(nil))

	authOrigin := fmt.Sprintf(
		`api_key="%s", algorithm="%s", headers="%s", signature="%s"`,
		apiKey,
		"hmac-sha256",
		"host date request-line",
		sha,
	)

	authorization := base64.StdEncoding.EncodeToString([]byte(authOrigin))

	v := url.Values{}
	v.Add("host", ul.Host)
	v.Add("date", date)
	v.Add("authorization", authorization)

	return hostURL + "?" + v.Encode()
}
