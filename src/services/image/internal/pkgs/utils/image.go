package utils

import (
	"encoding/base64"
	"fmt"
)

func GenImageDataURI(data []byte) string {
	base64Data := base64.StdEncoding.EncodeToString(data)
	dataURI := fmt.Sprintf("data:image/jpeg;base64,%s", base64Data)
	return dataURI
}
