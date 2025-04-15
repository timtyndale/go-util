// Copyright (c) 2024 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

//go:build ignore

package main

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/exp/slices"

	"github.com/timtyndale/go-util/exerrors"
	"github.com/timtyndale/go-util/unicodeurls"
)

func readEmojiLines(url, filter string) []string {
	return unicodeurls.ReadDataFileList(url, func(line string) (string, bool) {
		parts := strings.Split(line, "; ")
		if len(parts) < 2 || !strings.HasPrefix(parts[1], filter) {
			return "", false
		}
		return strings.TrimSpace(parts[0]), true
	})
}

func main() {
	variationSequences := readEmojiLines(unicodeurls.EmojiVariationSequences, "emoji style")
	fullyQualifiedSequences := readEmojiLines(unicodeurls.EmojiTest, "fully-qualified")
	var extraVariations []string
	for _, seq := range variationSequences {
		if !slices.Contains(fullyQualifiedSequences, seq) && !strings.HasPrefix(seq, "00") {
			unifiedWithoutVS := strings.Split(seq, " ")[0]
			extraVariations = append(extraVariations, fmt.Sprintf(`\x{%s}`, unifiedWithoutVS))
		}
	}
	file := exerrors.Must(os.OpenFile("emojis.go", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644))
	exerrors.Must(file.WriteString(`// Code generated by go generate; DO NOT EDIT.

package variationselector

import (
	"regexp"
	"strings"
)

func doInit() {
`))
	exerrors.Must(file.WriteString("\tvariationRegex = regexp.MustCompile(\n"))
	exerrors.Must(file.WriteString("\t\t`(^|[^\\x{200D}])(` +\n\t\t\t`"))
	exerrors.Must(file.WriteString(strings.Join(extraVariations, "|` +\n\t\t\t`")))
	exerrors.Must(file.WriteString("` +\n\t\t\t`)([^\\x{FE0F}\\x{FE0E}\\x{200D}\\x{1F3FB}\\x{1F3FC}\\x{1F3FD}\\x{1F3FE}\\x{1F3FF}]|$)`,\n\t)\n"))
	exerrors.Must(file.WriteString("\tvar qualifiedEmojis = []string{\n"))
	for _, unified := range fullyQualifiedSequences {
		if !strings.Contains(unified, "FE0F") {
			continue
		}
		unicode := unicodeurls.ParseHex(strings.Split(unified, " "))
		exerrors.Must(fmt.Fprintf(file, "\t\t\"%s\",\n", unicode))
	}
	exerrors.Must(file.WriteString("\t}\n"))
	exerrors.Must(file.WriteString(`
	replacerArgs := make([]string, len(qualifiedEmojis)*2)
	for i, emoji := range qualifiedEmojis {
		replacerArgs[i*2] = strings.ReplaceAll(emoji, "\ufe0f", "")
		replacerArgs[(i*2)+1] = emoji
	}
`))
	exerrors.Must(file.WriteString("\tfullyQualifier = strings.NewReplacer(replacerArgs...)\n"))
	exerrors.Must(file.WriteString("}\n"))
	exerrors.PanicIfNotNil(file.Close())
}
