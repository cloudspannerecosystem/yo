// Copyright (c) 2021 Mercari, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package internal

import (
	"testing"

	"github.com/jinzhu/inflection"
)

type inflectionPattern struct {
	single string
	plural string
}

var inflectionPatterns = []inflectionPattern{
	{"alias", "aliases"},
	{"belief", "beliefs"},
	{"bureau", "bureaus"},
	{"bus", "buses"},
	{"cafe", "cafes"},
	{"chef", "chefs"},
	{"chief", "chiefs"},
	{"cookie", "cookies"},
	{"crisis", "crises"},
	{"drive", "drives"},
	{"foe", "foes"},
	{"foot", "feet"},
	{"glove", "gloves"},
	{"goose", "geese"},
	{"halo", "halos"},
	{"hero", "heroes"},
	{"hive", "hives"},
	{"house", "houses"},
	{"index", "indices"},
	{"matrix", "matrices"},
	{"media", "media"},
	{"menu", "menus"},
	{"multimedia", "multimedia"},
	{"news", "news"},
	{"objective", "objectives"},
	{"octopus", "octopuses"},
	{"person", "people"},
	{"photo", "photos"},
	{"piano", "pianos"},
	{"potato", "potatoes"},
	{"powerhouse", "powerhouses"},
	{"quiz", "quizzes"},
	{"roof", "roofs"},
	{"roof", "roofs"},
	{"shoe", "shoes"},
	{"tax", "taxes"},
	{"thief", "thieves"},
	{"tooth", "teeth"},
	{"vertex", "vertices"},
	{"virus", "viri"},
	{"wave", "waves"},
	{"wife", "wives"},
	{"wolf", "wolves"},
	{`atlas`, `atlases`},
	{`beef`, `beefs`},
	{`brother`, `brothers`},
	{`cafe`, `cafes`},
	{`child`, `children`},
	{`cookie`, `cookies`},
	{`corpus`, `corpuses`},
	{`cow`, `cows`},
	{`drive`, `drives`},
	{`ganglion`, `ganglions`},
	{`genie`, `genies`},
	{`genus`, `genera`},
	{`graffito`, `graffiti`},
	{`harddrive`, `harddrives`},
	{`hero`, `heroes`},
	{`hoof`, `hoofs`},
	{`loaf`, `loaves`},
	{`man`, `men`},
	{`money`, `money`},
	{`mongoose`, `mongooses`},
	{`move`, `moves`},
	{`mythos`, `mythoi`},
	{`niche`, `niches`},
	{`numen`, `numina`},
	{`occiput`, `occiputs`},
	{`octopus`, `octopuses`},
	{`opus`, `opuses`},
	{`ox`, `oxen`},
	{`potato`, `potatoes`},
	{`soliloquy`, `soliloquies`},
	{`testis`, `testes`},
	{`trilby`, `trilbys`},
	{`turf`, `turfs`},
}

func init() {
	registerRule(nil)
}

func TestInflection(t *testing.T) {
	for _, tc := range inflectionPatterns {
		s := inflection.Plural(tc.single)
		if s != tc.plural {
			t.Errorf("Pluralize(%s): got %q, expected: %q", tc.single, s, tc.plural)
		}
		s = inflection.Singular(tc.plural)
		if s != tc.single {
			t.Errorf("Singular(%s): got %q, expected: %q", tc.plural, s, tc.single)
		}
	}
}
