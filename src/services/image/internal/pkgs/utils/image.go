package utils

import (
	"encoding/base64"
	"fmt"
	"net/http"
)

func GenImageDataURI(data []byte) string {
	mimeType := http.DetectContentType(data)
	base64Data := base64.StdEncoding.EncodeToString(data)
	dataURI := fmt.Sprintf("data:%s;base64,%s", mimeType, base64Data)
	return dataURI
}
