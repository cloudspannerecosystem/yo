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

type inflectionRule struct {
	find    string
	replace string
}

var defaultSingularInflections = []inflectionRule{
	{`(slave)s$`, `$1`},
	{`(drive)s$`, `$1`},
}

var defaultPluralInflections = []inflectionRule{
	{"^people$", "people"},
}

type irregularRule struct {
	singlar string
	plural  string
}

var defaultIrregularRules = []irregularRule{
	{"foot", "feet"},
	{"tooth", "teeth"},
	{`mythos`, `mythoi`},
	{`genie`, `genies`},
	{`genus`, `genera`},
	{`graffito`, `graffiti`},
	{`mongoose`, `mongooses`},
	{"goose", "geese"},
	{`niche`, `niches`},
	{`numen`, `numina`},
	{`occiput`, `occiputs`},
	{`trilby`, `trilbys`},
	{`testis`, `testes`},

	{`corpus`, `corpuses`},
	{`octopus`, `octopuses`},
	{`opus`, `opuses`},
	{`atlas`, `atlases`},

	{`move`, `moves`},
	{`wave`, `waves`},
	{`curve`, `curves`},
	{"glove", "gloves"},
	{`loaf`, `loaves`},
	{"thief", "thieves"},
	{"belief", "beliefs"},
	{"chief", "chiefs"},
	{"chef", "chefs"},
	{`turf`, `turfs`},
	{`beef`, `beefs`},
	{`hoof`, `hoofs`},
	{`cafe`, `cafes`},

	{"photo", "photos"},
	{"piano", "pianos"},
	{"halo", "halos"},
	{`hero`, `heroes`},
	{`potato`, `potatoes`},
	{`foe`, `foes`},

	{`media`, `media`},
}
