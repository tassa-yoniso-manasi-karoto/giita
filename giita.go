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
)
// const reference string = "a-ra-haṁ, abhi-vā-de-mi, su-pa-ṭi-pan-no, sam-bud-dho, svāk-khā-to, tas-sa, met-ta, a-haṁ, ho-mi, a-ve-ro, dham-mo, sam-mā, a-haṁ, kho, khan-dho, Ṭhā-nis-sa-ro, ya-thā, sey-yo, ho-ti, hon-ti, sot-thi, phoṭ-ṭhab-ba, khet-te, ya-thāj-ja, cī-va-raṁ, pa-ri-bhut-taṁ, sa-ra-naṁ, ma-kasa, pa-ṭha-mā-nus-sa-ti, Bha-ga-vā, sam-bud-dhas-sa, kit-ti-sad-do, a-ha-mā-da-re-na, khet-te, A-haṁ bhan-te sam-ba-hu-lā nā-nā-vat-thu-kā-ya pā-cit-ti-yā-yo ā-pat-ti-yo ā-pan-no tā pa-ṭi-de-se-mi. Pas-sa-si ā-vu-so? Ā-ma bhan-te pas-sā-mi. Ā-ya-tiṁ ā-vu-so saṁ-va-rey-yā-si. Sā-dhu suṭ-ṭhu bhan-te saṁ-va-ris-sā-mi."

// NOTE: makasa → "ma-kasa" = presumed to be an exception
// var test string = "arahaṁ, abhivādemi, supaṭipanno, sambuddho, svākkhāto, tassa, metta, ahaṁ, homi, avero, dhammo, sammā, ahaṁ, kho, khandho, Ṭhānissaro, yathā, seyyo, hoti, honti, sotthi, phoṭṭhabba, khette, yathājja, cīvaraṁ, paribhuttaṁ, saranaṁ, makasa, paṭhamānussati, Bhagavā, sambuddhassa, kittisaddo, ahamādarena, khette, Ahaṁ bhante sambahulā nānāvatthukāya pācittiyāyo āpattiyo āpanno tā paṭidesemi. Passasi āvuso? Āma bhante passāmi. Āyatiṁ āvuso saṁvareyyāsi. Sādhu suṭṭhu bhante saṁvarissāmi."

// TODO flag to load the CSS from a file

const (
	LongVowel = iota
	ShortVowel
	Consonant
	Punct
	Space
	Other
)

var (
	source string
	rePunc = regexp.MustCompile(`^\pP+`)
	reIsNotExeptPunct = regexp.MustCompile(`^[^-“’„	"\(\)\[\]«'‘‚-]+`)
	reSpace = regexp.MustCompile(`(?s)^\s+`)
	reComment = regexp.MustCompile(`(?s)\[.*?\]`)
	newline = "<br>"

	Vowels = []string{"ā", "e", "ī", "o", "ū", "ay", "a", "i", "u"}
	LongVowels = []string{"ā", "e", "ī", "o", "ū", "ay"}
	ShortVowels = []string{"a", "i", "u"}
	reLongVowels []*regexp.Regexp
	reShortVowels []*regexp.Regexp

	Consonants = []string{"bh", "dh", "ḍh", "gh", "jh", "kh", "ph", "th", "ṭh", "sm", "ch", "c", "g", "h", "s", "j", "r", "p", "b", "d", "k", "t", "ṭ", "m", "ṁ", "ṃ", "n", "ñ", "ṅ", "ṇ", "y", "l", "ḷ", "ḍ", "v"}
	reConsonants []*regexp.Regexp
	
	AspiratedConsonants = []string{"bh", "dh", "ḍh", "gh", "jh", "kh", "ph", "th", "ṭh"}
	UnstoppingChar = []string{"n", "ñ", "ṅ", "ṇ", "m", "ṁ", "ṃ", "l", "ḷ", "r", "y"}
	// EXCEPTION: "mok" in Pāṭimokkha takes a high tone: not supported.
	HighToneFirstChar = []string{"ch", "th", "ṭh", "kh", "ph", "sm", "s", "h"}
	OptionalHighToneFirstChar = []string{"v", "bh", "r", "n", "ṇ", "m", "y"}
	VowelTypes = []int{LongVowel, ShortVowel}
	IrrelevantTypes = []int{Punct, Space, Other}

	debug bool
	CurrentDir, refCmt string
	in, out *string
	wantNewlineNum, wantFontSize *int
	wantTxt, wantOptionalHigh, wantDark *bool
	wantHtml = true
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
	Units                               		[]UnitType
	isLong, NotStopped, hasHighToneFirstChar	bool
	Irrelevant					bool
	TrueHigh, OptionalHigh				bool
}

type HtmlClass string

func init() {
	e, err := os.Executable()
	if err != nil {
		fmt.Println(err)
	} else {
		CurrentDir = path.Dir(e)
	}
	// STRING
	in = flag.String("i", CurrentDir + "/input.txt", "path of input UTF-8 encoded text file\n")
	out = flag.String("o", CurrentDir + "/output.htm", "path of output file\n")
	refCmt = *flag.String("c", "[:]", "allow comments in input file and specify which characters marks\nrespectively the beginning and the end of a comment, separated\nby a colon")
	// BOOL
	wantTxt = flag.Bool("t", false , "use raw text instead of HTML for the output file (on with -t=true)")
	wantOptionalHigh = flag.Bool("optionalhigh", false , "requires -t, it formats optional high tones with capital letters\njust like true high tones (on with -optionalhigh=true)")
	wantDark = flag.Bool("d", false , "dark mode, will use a white font on a dark background (on with -d=true)")
	// INT
	wantNewlineNum = flag.Int("l", 1 , "set how many linebreaks will be created from a single linebreak in\nthe input file. Advisable to use 2 for smartphone/tablet/e-reader.\n")
	wantFontSize = flag.Int("f", 34 , "set font size")
	flag.Parse()
	if len(refCmt) != 3 {
		panic("You provided an invalid input of comment marks.")
	}
	page = fmt.Sprintf(page, *wantFontSize, refCmt[0:1], refCmt[2:3])
	if *wantTxt {
		wantHtml = false
		newline = "\n"
		page = ""
		if !isFlagPassed("o") {
			*out = CurrentDir + "/output.txt"
		}
	} else if *wantDark {
		page = strings.Replace(page, "body {", "body {\n  background: black;\n  color: white;", 1)
		page = strings.Replace(page, ".s::before{\n  content: \"⸱\";", ".s::before{\n  content: \"⸱\";\n  color: darkgrey;", 1)
		page = strings.Replace(page, ".comment {\n  background: lightgrey;", ".comment {\n  background: darkgrey;", 1)
	}
	newline = strings.Repeat(newline, *wantNewlineNum)
	fmt.Println("In:", *in)
	fmt.Println("Out:", *out)
	dat, err := os.ReadFile(*in)
	check(err)
	source = string(dat)
	//------
	for _, ShortVowel := range ShortVowels {
		re := regexp.MustCompile("(?i)^" + ShortVowel)
		reShortVowels = append(reShortVowels, re)
	}
	for _, LongVowel := range LongVowels {
		re := regexp.MustCompile("(?i)^" + LongVowel)
		reLongVowels = append(reLongVowels, re)
	}
	for _, Consonant := range Consonants {
		re := regexp.MustCompile("(?i)^" + Consonant)
		reConsonants = append(reConsonants, re)
	}
}


func main() {
	source = strings.ReplaceAll(source, "ṇ", "ṅ")
	source = strings.ReplaceAll(source, "ṃ", "ṁ")
	// chunks from long compound words need to be reunited or will be 
	// treated as separate
	source = strings.ReplaceAll(source, "-", "")
	comments := reComment.FindAllString(source, -1)
	source = reComment.ReplaceAllString(source, "𓃰")
	
	var UnitStack []UnitType
	for source != "" {
		notFound := true
		if rePunc.MatchString(source) {
			found := rePunc.FindString(source)
			UnitStack = append(UnitStack, UnitType{Str: found, Type: Punct})
			source = strings.TrimPrefix(source, found)
			notFound = false
		} else if reSpace.MatchString(source) {
			found := reSpace.FindString(source)
			UnitStack = append(UnitStack, UnitType{Str: found, Type: Space})
			source = strings.TrimPrefix(source, found)
			notFound = false
		}
		for i, list := range [][]*regexp.Regexp{reLongVowels, reShortVowels, reConsonants} {
			for _, re := range list {
				if re.MatchString(source) {
					found := re.FindString(source)
					UnitStack = append(UnitStack, UnitType{Str: found, Type: i})
					source = strings.TrimPrefix(source, found)
					notFound = false
					break
				}
			}
			if !notFound {
				break
			}
		}
		if notFound {
			char := strings.Split(source, "")[0]
			UnitStack = append(UnitStack, UnitType{Str: char, Type: Other})
			source = strings.TrimPrefix(source, char)
			if debug {
				fmt.Printf("'%s' : Character unknown\n", char)
			}
		}
	}
	var (
		Syllables []SyllableType
		Syllable SyllableType
	)
	for i, unit := range UnitStack {
		var PrevUnit, NextUnit, NextNextUnit UnitType
		if len(UnitStack) > i+2 {
			NextNextUnit = UnitStack[i+2]
			NextUnit = UnitStack[i+1]
		}
		if i-1 >= 0 {
			PrevUnit = UnitStack[i-1]
		}
		//assume true, overwrite everything after setting exceptions
		unit.Closing = true		
			// case SU-PA-ṬI-pan-no
		if unit.Type == ShortVowel &&
		!(NextUnit.Type == Consonant && NextNextUnit.Type == Consonant) &&
		!(strings.ToLower(NextUnit.Str) == "ṁ") &&
		!(contains(AspiratedConsonants, NextUnit.Str) && PrevUnit.Type != Consonant) {
			// case HO-mi
		} else if unit.Type == LongVowel &&
		!(NextUnit.Type == Consonant && NextNextUnit.Type == Consonant) &&
		!(strings.ToLower(NextUnit.Str) == "ṁ") {
			// case sag-GAṀ and also "2 consonants in a row" case
		} else if unit.Type == Consonant &&
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
	for h, Syllable := range Syllables {
		for i, unit := range Syllable.Units {
			var NextUnit UnitType
			if len(Syllable.Units) > i+1 {
				NextUnit = Syllable.Units[i+1]
			}
			if contains(IrrelevantTypes, unit.Type) {
				Syllable.Irrelevant = true
			}
			if (unit.Type == ShortVowel &&
			strings.ToLower(NextUnit.Str) == "ṁ") ||
			(unit.Type == ShortVowel && NextUnit.Type == Consonant && NextUnit.Closing) ||
			(unit.Type == LongVowel) {
				Syllable.isLong = true
			}
			if contains(UnstoppingChar, strings.ToLower(unit.Str)) &&
			unit.Closing ||
			(unit.Type == LongVowel && unit.Closing) {
				Syllable.NotStopped = true
			}
			if contains(HighToneFirstChar, strings.ToLower(Syllable.Units[0].Str)) {
				Syllable.hasHighToneFirstChar = true
			}
			if Syllable.hasHighToneFirstChar &&
			Syllable.NotStopped &&
			Syllable.isLong {
				Syllable.TrueHigh = true
				if Syllable.TrueHigh && !wantHtml {
					for k, unit := range Syllable.Units {
						Syllable.Units[k].Str = strings.ToUpper(unit.Str)
					}
				}
			}
			//---
			if !Syllable.TrueHigh && unit.Type == ShortVowel &&
			contains(OptionalHighToneFirstChar, strings.ToLower(Syllable.Units[0].Str)) {
				if unit.Closing ||
				!contains(UnstoppingChar, strings.ToLower(NextUnit.Str)) {
					Syllable.OptionalHigh = true
				}
				if Syllable.OptionalHigh && !wantHtml && *wantOptionalHigh {
					for k, unit := range Syllable.Units {
						Syllable.Units[k].Str = strings.ToUpper(unit.Str)
					}
					
				}
			}
			Syllables[h] = Syllable
		}
	}
	buf := bytes.NewBufferString(page)
	separator := "⸱"
	span := "<span class=\"%s\">"
	if wantHtml {
		separator = "<span class=s></span>"
	}
	openword := false
	for h, Syllable := range Syllables {
		if wantHtml && !Syllable.Irrelevant && !openword {
			fmt.Fprintf(buf, span, "w")
			openword = true
		} else if wantHtml && Syllable.Irrelevant && openword {
			buf.WriteString("</span>")
			openword = false
		}
		
		class := ""
		if t := Syllable.whichTone(); t != "none" {
			class += t
		}
		if Syllable.isLong {
			class = appendClass(class, "long")
		} else if !Syllable.Irrelevant {
			class = appendClass(class, "short")
		}	
		if class != "" && wantHtml {
			fmt.Fprintf(buf, span, class)
		}
		for _, unit := range Syllable.Units {
			if strings.Contains(unit.Str, "\n") {
				unit.Str = strings.ReplaceAll(unit.Str, "\n", newline)
				buf.WriteString(unit.Str)
			} else if reSpace.MatchString(unit.Str) {
				buf.WriteString(" ")
				if wantHtml {
					buf.WriteString("&nbsp;")
				}
			} else if rePunc.MatchString(unit.Str) && reIsNotExeptPunct.MatchString(unit.Str) {
				if wantHtml {
					buf.WriteString(html.EscapeString(unit.Str) + "<span class=\"punct\"></span>")
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
		if len(Syllables) > h+1 &&
		!contains(IrrelevantTypes, Syllable.Units[len(Syllable.Units)-1].Type) &&
		!contains(IrrelevantTypes, Syllables[h+1].Units[0].Type) {
			buf.WriteString(separator) 
		}
	}
	if wantHtml {
		buf.WriteString("</body></html>")
	}
	// maybe just discard buf for a modifiable string
	outstr := buf.String()
	for _, comment := range comments {
		if wantHtml {
			comment = html.EscapeString(comment[1:len(comment)-1])
			comment = "<span class=comment>" + comment + "</span>"
		}
		outstr = strings.Replace(outstr, "𓃰", comment, 1)
	}
	err := os.WriteFile(*out, []byte(outstr), 0644)
	check(err)
	fmt.Println("Done")
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
	return "none"
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