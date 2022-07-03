package json

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

func JSON(object interface{}) string {
	json, _ := json.Marshal(object)
	return string(json)
}

func ReadFile(filePath string) string {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return ""
	}
	defer file.Close()

	var builder strings.Builder
	br := bufio.NewReader(file)
	for {
		byteArray, _, err2 := br.ReadLine()
		if err2 == io.EOF {
			break
		}

		builder.WriteString(strings.TrimSpace(string(byteArray)))
	}
	return builder.String()
}
