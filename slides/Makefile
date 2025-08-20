SHELL := /usr/bin/bash

.DEFAULT_GOAL := build

OUT := index.html
SRC := slides.md

REVEAL := https://unpkg.com/reveal.js@5.1.0
THEME := white
WIDTH := 1920
HEIGHT := 1080
MAXSCALE := 1.0

# Tool detection
HAVE_PANDOC := $(shell command -v pandoc >/dev/null 2>&1 && echo yes)
HAVE_MERMAID_FILTER := $(shell command -v mermaid-filter >/dev/null 2>&1 && echo yes)

# Mermaid filter defaults
export MERMAID_FILTER_FORMAT ?= svg
export MERMAID_FILTER_LOC ?= inline

# Help mermaid-cli (mmdc) find a local Chrome/Chromium if available
export PUPPETEER_EXECUTABLE_PATH ?= $(firstword $(wildcard \
	/usr/bin/google-chrome-stable \
	/usr/bin/google-chrome \
	/usr/bin/chromium \
	/usr/bin/chromium-browser))

check-tools:
	$(if $(HAVE_PANDOC),,$(error pandoc not found. Please install pandoc (see https://pandoc.org)))
	$(if $(HAVE_MERMAID_FILTER),,$(error mermaid-filter not found. Install with: npm install -g mermaid-filter))

build: check-tools
	pandoc "$(SRC)" \
		-t revealjs \
		-F mermaid-filter \
		-s \
		-o "$(OUT)" \
		-V revealjs-url="$(REVEAL)" \
		-V theme=$(THEME) \
		-V slideNumber=true \
		-V history=true \
		-V width=$(WIDTH) \
		-V height=$(HEIGHT) \
		-V maxScale=$(MAXSCALE)
	@echo "Built $(OUT)"

clean:
	rm -f "$(OUT)"

.PHONY: build clean check-tools
