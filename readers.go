package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

//go:embed index.php
var phpContent string


func indexPhpReader() *strings.Reader {
	return strings.NewReader(phpContent)
}

func metadataJSON() *bytes.Reader  {
	now := time.Now()
	metadata := deployMetaData{DeployDate: now.Unix()}
	b, err := json.Marshal(metadata)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return bytes.NewReader(b)
}
