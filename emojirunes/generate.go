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
	"strconv"
	"strings"

	"golang.org/x/exp/slices"

	"github.com/timtyndale/go-util/exerrors"
	"github.com/timtyndale/go-util/unicodeurls"
)

func main() {
	allEmojiRelatedCodepoints := make(map[string]struct{}, 10000)
	unicodeurls.ReadDataFile(unicodeurls.EmojiTest, func(line string) {
		parts := strings.Split(line, "; ")
		if len(parts) < 2 {
			return
		}
		for _, codepoint := range strings.Split(strings.TrimSpace(parts[0]), " ") {
			if !strings.HasPrefix(codepoint, "00") {
				allEmojiRelatedCodepoints[codepoint] = struct{}{}
			}
		}
	})
	emojiRunes := make([]rune, 0, len(allEmojiRelatedCodepoints))
	for runeHex := range allEmojiRelatedCodepoints {
		emojiRunes = append(emojiRunes, rune(exerrors.Must(strconv.ParseInt(runeHex, 16, 32))))
	}
	slices.Sort(emojiRunes)
	file := exerrors.Must(os.OpenFile("data.go", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644))
	exerrors.Must(file.WriteString(`// Code generated by go generate; DO NOT EDIT.

package emojirunes

var EmojiRunes = []rune{
`))
	for _, r := range emojiRunes {
		exerrors.Must(fmt.Fprintf(file, "\t0x%x,\n", r))
	}
	exerrors.Must(file.WriteString("}\n"))
}
