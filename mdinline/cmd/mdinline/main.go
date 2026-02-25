package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/inforix/wxcli/mdinline"
)

func main() {
	mdPath := flag.String("md", "", "Path to markdown file")
	cssPath := flag.String("css", "", "Path to CSS file")
	media := flag.String("media", "screen", "Media type: screen|print")
	outPath := flag.String("out", "", "Output HTML file (default: stdout)")
	flag.Parse()

	if *mdPath == "" || *cssPath == "" {
		fmt.Fprintln(os.Stderr, "-md and -css are required")
		os.Exit(2)
	}

	mdBytes, err := os.ReadFile(*mdPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read markdown: %v\n", err)
		os.Exit(1)
	}

	cssBytes, err := os.ReadFile(*cssPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read css: %v\n", err)
		os.Exit(1)
	}

	html, err := mdinline.RenderMarkdownWithInlineCSS(string(mdBytes), string(cssBytes), *media)
	if err != nil {
		fmt.Fprintf(os.Stderr, "render: %v\n", err)
		os.Exit(1)
	}

	if *outPath == "" {
		fmt.Fprint(os.Stdout, html)
		return
	}

	if err := os.WriteFile(*outPath, []byte(html), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "write output: %v\n", err)
		os.Exit(1)
	}
}
