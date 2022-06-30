// ☸

package main

import (
	"fmt"
	"regexp"
	"strings"
	"os"
	"path"
	"bytes"
	"flag"
	"html"
	"unicode/utf8"
)
// const reference string = "a-ra-haṁ, abhi-vā-de-mi, su-pa-ṭi-pan-no, sam-bud-dho, svāk-khā-to, tas-sa, met-ta, a-haṁ, ho-mi, a-ve-ro, dham-mo, sam-mā, a-haṁ, kho, khan-dho, Ṭhā-nis-sa-ro, ya-thā, sey-yo, ho-ti, hon-ti, sot-thi, phoṭ-ṭhab-ba, khet-te, ya-thāj-ja, cī-va-raṁ, pa-ri-bhut-taṁ, sa-ra-naṁ, ma-kasa, pa-ṭha-mā-nus-sa-ti, Bha-ga-vā, sam-bud-dhas-sa, kit-ti-sad-do, a-ha-mā-da-re-na, khet-te, A-haṁ bhan-te sam-ba-hu-lā nā-nā-vat-thu-kā-ya pā-cit-ti-yā-yo ā-pat-ti-yo ā-pan-no tā pa-ṭi-de-se-mi. Pas-sa-si ā-vu-so? Ā-ma bhan-te pas-sā-mi. Ā-ya-tiṁ ā-vu-so saṁ-va-rey-yā-si. Sā-dhu suṭ-ṭhu bhan-te saṁ-va-ris-sā-mi."

// NOTE: makasa → "ma-kasa" = presumed to be an exception
// var test string = "arahaṁ, abhivādemi, supaṭipanno, sambuddho, svākkhāto, tassa, metta, ahaṁ, homi, avero, dhammo, sammā, ahaṁ, kho, khandho, Ṭhānissaro, yathā, seyyo, hoti, honti, sotthi, phoṭṭhabba, khette, yathājja, cīvaraṁ, paribhuttaṁ, saranaṁ, makasa, paṭhamānussati, Bhagavā, sambuddhassa, kittisaddo, ahamādarena, khette, Ahaṁ bhante sambahulā nānāvatthukāya pācittiyāyo āpattiyo āpanno tā paṭidesemi. Passasi āvuso? Āma bhante passāmi. Āyatiṁ āvuso saṁvareyyāsi. Sādhu suṭṭhu bhante saṁvarissāmi."

/*
TODO
	flag to load the CSS from a file
	make test files
	discard buf for a modifiable string
COULD
	diff against ↓ to find exceptions (all long falling tones?): METta, viMOKkha, sometime also pāṭiMOKkhe
	https://www.dhammatalks.org/books/ChantingGuide/Section0000.html,
*/

const (
	LongVwl = iota
	ShortVwl
	Cons
	Punct
	Space
	Other
)

var (
	src                string
	rePunc             = regexp.MustCompile(`^\pP+`)
	reIsNotExceptPunct = regexp.MustCompile(`^[^-“’„"\(\)\[\]«'‘‚-]+`)
	reSpace            = regexp.MustCompile(`(?s)^\s+`)
	reCmt              = regexp.MustCompile(`(?s)\[.*?\]`)
	IrrelevantTypes    = []int{Punct, Space, Other}
	// the \n makes the html source somewhat readable
	newline = "<br>\n"

	//Vowels = []string{"ā", "e", "ī", "o", "ū", "ay", "a", "i", "u"}
	LongVwls    = []string{"ā", "e", "ī", "o", "ū", "ay"}
	ShortVwls   = []string{"a", "i", "u"}
	VowelTypes  = []int{LongVwl, ShortVwl}
	reLongVwls  []*regexp.Regexp
	reShortVwls []*regexp.Regexp

	C   = []string{"bh", "dh", "ḍh", "gh", "jh", "kh", "ph", "th", "ṭh", "sm", "ch", "c", "g", "h", "s", "j", "r", "p", "b", "d", "k", "t", "ṭ", "m", "ṁ", "ṃ", "n", "ñ", "ṅ", "ṇ", "y", "l", "ḷ", "ḍ", "v"}
	reC []*regexp.Regexp

	NeverLastPos      = []string{"bh", "dh", "ḍh", "gh", "jh", "kh", "ph", "th", "ṭh", "v", "r"}
	UnstopChar        = []string{"n", "ñ", "ṅ", "ṇ", "m", "ṁ", "ṃ", "l", "ḷ", "y"}
	HighToneFirstChar = []string{"ch", "th", "ṭh", "kh", "ph", "sm", "s", "h"}
	OptHighFirstChar  = []string{"v", "bh", "r", "n", "ṇ", "m", "y"}

	debug                               *bool
	CurrentDir                          string
	in, out, refCmt                     *string
	wantNewlineNum, wantFontSize        *int
	wantTxt, wantOptionalHigh, wantDark *bool
	wantHtml                            = true
	page = `<!DOCTYPE html> <html><head>
<meta charset="UTF-8">
<style>
body {
  font-size: %dpx;
  line-height: 1.4em;
  letter-spacing: -0.03em;
}

.w {
  white-space: nowrap;
}

.s::before{
  content: "⸱";
}

.punct::after{
  content: "█";
  color: orangered; /*#5c5c5c;*/
}

.truehigh{
  font-weight: bold;
  vertical-align: 13%%;
}

.comment {
  background: lightgrey;
  font-style: italic;
}
.comment::before {
  content: "%s";
}
.comment::after {
  content: "%s";
}

.optionalhigh{
  /*font-style: italic;*/
}

.low {
  /*Low tones: Not implemented!*/
  vertical-align: -13%%;
}

.short {
 /*font-weight: 300;*/
}
</style></head>
<body>`
)

type UnitType struct {
	Str     string
	Type    int
	Len     string
	Closing bool
}
type SyllableType struct {
	Units                                    []UnitType
	IsLong, NotStopped, HasHighToneFirstChar bool
	Irrelevant                               bool
	TrueHigh, OptionalHigh                   bool
}

func init() {
	for _, ShortVwl := range ShortVwls {
		re := regexp.MustCompile("(?i)^" + ShortVwl)
		reShortVwls = append(reShortVwls, re)
	}
	for _, LongVwl := range LongVwls {
		re := regexp.MustCompile("(?i)^" + LongVwl)
		reLongVwls = append(reLongVwls, re)
	}
	for _, Cons := range C {
		re := regexp.MustCompile("(?i)^" + Cons)
		reC = append(reC, re)
	}
}


func main() {
	e, err := os.Executable()
	if err != nil {
		fmt.Println(err)
	} else {
		CurrentDir = path.Dir(e)
	}
	// STRING
	in = flag.String("i", CurrentDir + "/input.txt", "path of input UTF-8 encoded text file\n")
	out = flag.String("o", CurrentDir + "/output.htm", "path of output file\n")
	refCmt = flag.String("c", "[:]", "allow comments in input file and specify which characters marks\nrespectively the beginning and the end of a comment, separated\nby a colon")
	// BOOL
	wantTxt = flag.Bool("t", false , "use raw text instead of HTML for the output file (on with -t=true)")
	wantOptionalHigh = flag.Bool("optionalhigh", false , "requires -t, it formats optional high tones with capital letters\njust like true high tones (on with -optionalhigh=true)")
	wantDark = flag.Bool("d", false , "dark mode, will use a white font on a dark background (on with -d=true)")
	debug = flag.Bool("debug", false , "")
	// INT
	wantNewlineNum = flag.Int("l", 1 , "set how many linebreaks will be created from a single linebreak in\nthe input file. Advisable to use 2 for smartphone/tablet/e-reader.\n")
	wantFontSize = flag.Int("f", 34 , "set font size")
	flag.Parse()
	if len(*refCmt) != 3 {
		panic("You provided an invalid input of comment marks.")
	}
	page = fmt.Sprintf(page, *wantFontSize, (*refCmt)[0:1], (*refCmt)[2:3])
	if *wantTxt {
		wantHtml = false
		newline = "\n"
		page = ""
		if !isFlagPassed("o") {
			*out = CurrentDir + "/output.txt"
		}
	} else if *wantDark {
		page = strings.Replace(page, "body {", "body {\n  background: black;\n  color: white;", 1)
		page = strings.Replace(page, ".s::before{\n  content: \"⸱\";", ".s::before{\n  content: \"⸱\";\n  color: #858585;", 1)
		page = strings.Replace(page, ".comment {\n  background: lightgrey;", ".comment {\n  background: darkgrey;", 1)
	}
	newline = strings.Repeat(newline, *wantNewlineNum)
	fmt.Println("In:", *in)
	fmt.Println("Out:", *out)
	dat, err := os.ReadFile(*in)
	check(err)
	src = string(dat)
	src = strings.ReplaceAll(src, "ṃ", "ṁ")
	src = strings.ReplaceAll(src, "Ṃ", "Ṁ")
	// chunks from long compound words need to be reunited or will be treated as separate
	src = strings.ReplaceAll(src, "-", "")
	cmts := reCmt.FindAllString(src, -1)
	src = reCmt.ReplaceAllString(src, "𓃰")
	// As a consequence of putting the 2 characters consonants/vowels at
	// the beginning of the the reference consonants/vowels arrays, this 
	// parser performs a blind greedy matching. 
	// There is however one exception where this isn't desired:
	// the "ay" long vowel versus a "a" short vowel followed by a "y" consonant. 
	// This script tries to distinguish the two by assessing if a "ay" would 
	// result in a syllable imbalance in the next syllable.
	RawUnits := Parser(src)
	Syllables := SyllableBuilder(RawUnits)
	RawUnits = []UnitType{}
	SkipNext := false
	for h, Syllable := range Syllables {
		if SkipNext {
			SkipNext = false
			continue
		}
		for _, unit := range Syllable.Units {
			var (
				ok bool
				NextSyl SyllableType
			)
			if strings.ToLower(unit.Str) != "ay" {
				ok = true
			} else if h+1 > len(Syllables) {
				ok = true
			} else {
				NextSyl = Syllables[h+1]				
				VwlNum, ConsNum := NextSyl.Describe()
				if VwlNum == 1 && ConsNum == 0 {						
				} else if VwlNum == 2 {
				} else {
					ok = true
				}
			}
			if !ok {
				// preserves the capital letter if there is one
				y := UnitType{Str: unit.Str[1:2], Type: Cons}
				unit = UnitType{Str: unit.Str[:1], Type: ShortVwl}
				nxtRpl := append([]UnitType{y}, NextSyl.Units...)
				rpl := append([]UnitType{unit}, nxtRpl...)
				RawUnits = append(RawUnits, rpl...)
				SkipNext = true
			} else {
				RawUnits = append(RawUnits, unit)
			}
		}
	}
	// units have been corrected, just rebuild from scratch
	Syllables = SyllableBuilder(RawUnits)
	//---------
	Syllables = SetTones(Syllables)
	buf := bytes.NewBufferString(page)
	separator := "⸱"
	span := "<span class=\"%s\">"
	if wantHtml {
		separator = "<span class=s></span>"
	}
	openword := false
	for h, Syllable := range Syllables {
		class := ""
		if wantHtml {
			if !Syllable.Irrelevant && !openword {
				fmt.Fprintf(buf, span, "w")
				openword = true
			} else if Syllable.Irrelevant && openword {
				buf.WriteString("</span>")
				openword = false
			}
			class += Syllable.whichTone()
			if Syllable.IsLong {
				class = appendClass(class, "long")
			} else if !Syllable.Irrelevant {
				class = appendClass(class, "short")
			}	
			if class != "" {
				fmt.Fprintf(buf, span, class)
			}
		}
		for _, unit := range Syllable.Units {
			if strings.Contains(unit.Str, "\n") {
				buf.WriteString(strings.ReplaceAll(unit.Str, "\n", newline))
			} else if reSpace.MatchString(unit.Str) {
				buf.WriteString(" ")
				if wantHtml {
					buf.WriteString("&nbsp;")
				}
			} else if rePunc.MatchString(unit.Str) && reIsNotExceptPunct.MatchString(unit.Str) {
				if wantHtml {
					buf.WriteString(html.EscapeString(unit.Str) + "<span class=punct></span>")
				} else {
					buf.WriteString(unit.Str + "█")
				}
			} else {
				if wantHtml {
					buf.WriteString(html.EscapeString(unit.Str))
				} else {
					buf.WriteString(unit.Str)
				}
			}
		}
		if class != "" && wantHtml {
			buf.WriteString("</span>")
		}
		//-----
		if len(Syllables) > h+1 {
			lastUnit := Syllable.Units[len(Syllable.Units)-1]
			NextSylFirstUnit := Syllables[h+1].Units[0]
			if !contains(IrrelevantTypes, lastUnit.Type) &&
			!contains(IrrelevantTypes, NextSylFirstUnit.Type) {
				buf.WriteString(separator) 
			}
		}
	}
	if wantHtml {
		buf.WriteString("</body></html>")
	}
	outstr := buf.String()
	for _, cmt := range cmts {
		if wantHtml {
			cmt = html.EscapeString(cmt[1:len(cmt)-1])
			cmt = "<span class=comment>" + cmt + "</span>"
		}
		outstr = strings.Replace(outstr, "𓃰", cmt, 1)
	}
	err = os.WriteFile(*out, []byte(outstr), 0644)
	check(err)
	fmt.Println("Done")
}


func Parser(src string) (RawUnits []UnitType) {
	f := func(m string, src *string, i int) (u UnitType) {
		u = UnitType{Str: m, Type: i}
		*src = strings.TrimPrefix(*src, m)
		return
	}
	lists := [][]*regexp.Regexp{reLongVwls, reShortVwls, reC, []*regexp.Regexp{rePunc}, []*regexp.Regexp{reSpace}}
	for src != "" {
		found := false
		for i, list := range lists {
			for _, re := range list {
				if found = re.MatchString(src); found {	
					m := re.FindString(src)
					RawUnits = append(RawUnits, f(m, &src, i))
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			char := strings.Split(src, "")[0]
			RawUnits = append(RawUnits, f(char, &src, Other))
			if *debug && char != "𓃰" {
				r, _ := utf8.DecodeRuneInString(char)
				fmt.Printf("'%s': Char unknown (%U)\n",	char, r)
			}
		}
	}
	return
}

func SyllableBuilder(Units []UnitType) (Syllables []SyllableType) {
	var Syllable SyllableType
	for i, unit := range Units {
		var (
			PrevUnit, NextUnit, NextNextUnit UnitType
			notBeforeTwoCons bool
		)
		if i+2 < len(Units){
			NextNextUnit = Units[i+2]
		} else {
			notBeforeTwoCons = true
		}
		if i+1 < len(Units){
			NextUnit = Units[i+1]
		}
		if i-1 >= 0 {
			PrevUnit = Units[i-1]
		}
		if !(NextUnit.Type == Cons && NextNextUnit.Type == Cons) {
			notBeforeTwoCons = true
		}
		//assume true, overwrite everything after setting exceptions
		unit.Closing = true		
			// case no further input
		if i+1 == len(Units) {
			// case SU-PA-ṬI-pan-no
		} else if unit.Type == ShortVwl && notBeforeTwoCons &&
		!(strings.ToLower(NextUnit.Str) == "ṁ") &&
		!(contains(NeverLastPos,NextUnit.Str) && !contains(VowelTypes, NextNextUnit.Type)) {
			// case HO-mi
		} else if unit.Type == LongVwl && notBeforeTwoCons &&
		!(strings.ToLower(NextUnit.Str) == "ṁ") {
			// case sag-GAṀ and also "2 consonants in a row" case
		} else if unit.Type == Cons &&
		!contains(VowelTypes, NextUnit.Type) &&
		contains(VowelTypes, PrevUnit.Type) {
		} else {
			unit.Closing = false
		}		
		//----
		if contains(IrrelevantTypes, PrevUnit.Type) &&
		!contains(IrrelevantTypes, unit.Type) {
			Syllables = append(Syllables, Syllable)
			Syllable = *new(SyllableType)
		}
		Syllable.Units = append(Syllable.Units, unit)
		if unit.Closing ||
		contains(IrrelevantTypes, NextUnit.Type) && 
		!contains(IrrelevantTypes, unit.Type) {
			Syllables = append(Syllables, Syllable)
			Syllable = *new(SyllableType)
		}
	}
	if len(Syllable.Units) != 0 {
		if *debug {
			fmt.Println("len(Syllable.Units) != 0, APPENDING....")
		}
		Syllables = append(Syllables, Syllable)
	}
	return
}

func SetTones(Syllables []SyllableType) []SyllableType {
	for h, Syllable := range Syllables {
		for i, unit := range Syllable.Units {
			var NextUnit UnitType
			firstUnit := strings.ToLower(Syllable.Units[0].Str)
			if len(Syllable.Units) > i+1 {
				NextUnit = Syllable.Units[i+1]
			}
			if contains(IrrelevantTypes, unit.Type) {
				Syllable.Irrelevant = true
			}
			if (unit.Type == ShortVwl && strings.ToLower(NextUnit.Str) == "ṁ") ||
			(unit.Type == ShortVwl && NextUnit.Type == Cons && NextUnit.Closing) ||
			(unit.Type == LongVwl) {
				Syllable.IsLong = true
			}
			if contains(UnstopChar, strings.ToLower(unit.Str)) && unit.Closing ||
			(unit.Type == LongVwl && unit.Closing) {
				Syllable.NotStopped = true
			}
			if contains(HighToneFirstChar, firstUnit) {
				Syllable.HasHighToneFirstChar = true
			}
			if Syllable.HasHighToneFirstChar &&  Syllable.NotStopped && Syllable.IsLong {
				Syllable.TrueHigh = true
				if Syllable.TrueHigh && !wantHtml {
					for k, unit := range Syllable.Units {
						s := strings.ToUpper(unit.Str)
						Syllable.Units[k].Str = s
					}
				}
			}
			//---
			if !Syllable.TrueHigh && unit.Type == ShortVwl && contains(OptHighFirstChar, firstUnit) {
				if unit.Closing || !contains(UnstopChar, strings.ToLower(NextUnit.Str)) {
					Syllable.OptionalHigh = true
				}
				if Syllable.OptionalHigh && !wantHtml && *wantOptionalHigh {
					for k, unit := range Syllable.Units {
						s := strings.ToUpper(unit.Str)
						Syllable.Units[k].Str = s
					}
					
				}
			}
			Syllables[h] = Syllable
		}
	}
	return Syllables
}


func appendClass(class, s string) string {
	if class != "" {
		class += " "
	}
	class += s
	return class
}

func (Syllable *SyllableType) whichTone() string {
	switch {
	case Syllable.TrueHigh:
		return "truehigh"
	case Syllable.OptionalHigh:
		return "optionalhigh"
	default:
	}
	return ""
}


func (Syllable *SyllableType) Describe() (VwlNum int, ConsNum int) {
	for _, unit := range Syllable.Units {
		switch {
		case contains(VowelTypes, unit.Type):
			VwlNum += 1
		case unit.Type == Cons:
			ConsNum += 1
		}				
	}
	return 
}

func contains[T comparable] (arr []T, i T) bool {
	for _, a := range arr {
		if a == i {
			return true
		}
	}
	return false
}

func check(e error) {
    if e != nil {
        panic(e)
    }
}

func isFlagPassed(name string) (found bool) {
    flag.Visit(func(f *flag.Flag) {
        if f.Name == name {
            found = true
        }
    })
    return
}