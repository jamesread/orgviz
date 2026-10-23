package main

import _ "embed"

//go:embed llms.txt
var llmsTxt []byte

//go:embed gen/openapi.json
var openAPISpec []byte
