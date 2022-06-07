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
	"bufio"
)
// const reference string = "a-ra-haṁ, abhi-vā-de-mi, su-pa-ṭi-pan-no, sam-bud-dho, svāk-khā-to, tas-sa, met-ta, a-haṁ, ho-mi, a-ve-ro, dham-mo, sam-mā, a-haṁ, kho, khan-dho, Ṭhā-nis-sa-ro, ya-thā, sey-yo, ho-ti, hon-ti, sot-thi, phoṭ-ṭhab-ba, khet-te, ya-thāj-ja, cī-va-raṁ, pa-ri-bhut-taṁ, sa-ra-naṁ, ma-kasa, pa-ṭha-mā-nus-sa-ti, Bha-ga-vā, sam-bud-dhas-sa, kit-ti-sad-do, a-ha-mā-da-re-na, khet-te, A-haṁ bhan-te sam-ba-hu-lā nā-nā-vat-thu-kā-ya pā-cit-ti-yā-yo ā-pat-ti-yo ā-pan-no tā pa-ṭi-de-se-mi. Pas-sa-si ā-vu-so? Ā-ma bhan-te pas-sā-mi. Ā-ya-tiṁ ā-vu-so saṁ-va-rey-yā-si. Sā-dhu suṭ-ṭhu bhan-te saṁ-va-ris-sā-mi."

// var test string = "arahaṁ, abhivādemi, supaṭipanno, sambuddho, svākkhāto, tassa, metta, ahaṁ, homi, avero, dhammo, sammā, ahaṁ, kho, khandho, Ṭhānissaro, yathā, seyyo, hoti, honti, sotthi, phoṭṭhabba, khette, yathājja, cīvaraṁ, paribhuttaṁ, saranaṁ, makasa, paṭhamānussati, Bhagavā, sambuddhassa, kittisaddo, ahamādarena, khette, Ahaṁ bhante sambahulā nānāvatthukāya pācittiyāyo āpattiyo āpanno tā paṭidesemi. Passasi āvuso? Āma bhante passāmi. Āyatiṁ āvuso saṁvareyyāsi. Sādhu suṭṭhu bhante saṁvarissāmi."

// TODO flag to load the CSS from a file

var (
	source string
	rePunc = regexp.MustCompile(`^\pP+`)
	reIsNotExeptPunc = regexp.MustCompile(`^[^-“’„	"\(\)\[\]«'‘‚-]+`)
	reSpace = regexp.MustCompile(`(?s)^\s+`)
	newline = "<br>"

	Vowels = []string{"ā", "e", "ī", "o", "ū", "ay", "a", "i", "u"}
	LongVowels = []string{"ā", "e", "ī", "o", "ū", "ay"}
	ShortVowels = []string{"a", "i", "u"}
	reLongVowels []*regexp.Regexp
	reShortVowels []*regexp.Regexp

	Consonants = []string{"bh", "dh", "ḍh", "gh", "jh", "kh", "ph", "th", "ṭh", "sm", "ch", "c", "g", "h", "s", "j", "r", "p", "b", "d", "k", "t", "ṭ", "m", "ṁ", "ṃ", "n", "ñ", "ṅ", "ṇ", "y", "l", "ḷ", "ḍ", "v"}
	reConsonants []*regexp.Regexp
	
	AspiratedConsonants = []string{"bh", "dh", "ḍh", "gh", "jh", "kh", "ph", "th", "ṭh"}
	UnstoppingCar = []string{"n", "ñ", "ṅ", "ṇ", "m", "ṁ", "ṃ", "l", "ḷ", "r", "y"}
	// EXCEPTION: "mok" in Pāṭimokkha takes a high tone: not supported atm.
	HighToneFirstCar = []string{"ch", "th", "ṭh", "kh", "ph", "sm", "s", "h"}
	OptionalHighToneFirstCar = []string{"v", "bh", "r", "n", "ṇ", "m", "y"}

	CurrentDir string
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
  vertical-align: 13%;
}

.optionalhigh{
  /*font-style: italic;*/
}

.low {
  /*Low tones: Not implemented!*/
  vertical-align: -13%;
}

.short {
 /*font-weight: 300;*/
}
</style></head>
<body>`
)

type UnitType struct {
	Str     string
	Type    string
	Len     string
	Closing bool
}
type SyllableType struct {
	Units                               	[]UnitType
	isLong, NotStopped, hasHighToneFirstCar	bool
	Irrelevant				bool
	TrueHigh, OptionalHigh			bool
}

func init() {
	e, err := os.Executable()
	if err != nil {
		fmt.Println(err)
	} else {
		CurrentDir = path.Dir(e)
	}
	in = flag.String("i", CurrentDir + "/input.txt", "path of input UTF-8 encoded text file")
	out = flag.String("o", CurrentDir + "/output.htm", "path of output file")
	wantTxt = flag.Bool("t", false , "use raw text format instead of HTML for the output file (turn on with -t=true)")
	wantOptionalHigh = flag.Bool("optionalhigh", false , "requires -t, it formats optional high tones with capital letters just like true high tones (turn on with -optionalhigh=true)")
	wantDark = flag.Bool("d", false , "dark mode, will use a white font on a dark background")
	wantNewlineNum = flag.Int("l", 1 , "set how many linebreaks will be created from a single linebreak in the input file. Advisable to use 2 or 3 for smartphone/tablet/e-reader.")
	wantFontSize = flag.Int("f", 34 , "set font size")
	flag.Parse()
	page = fmt.Sprintf(page, *wantFontSize)
	if *wantTxt {
		wantHtml = false
		newline = "\n"
		page = ""
		if !isFlagPassed("o") {
			*out = CurrentDir + "/output.txt"
		}
	} else if *wantDark {
		page = strings.Replace(page, "body {", "body {\nbackground: black;\n  color: white;", 1)
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
	var UnitStack []UnitType
	for source != "" {
		notFound := true
		if rePunc.MatchString(source) {
			found := rePunc.FindString(source)
			UnitStack = append(UnitStack, UnitType{Str: found, Type: "Punctuation"})
			source = strings.TrimPrefix(source, found)
			notFound = false
		} else if reSpace.MatchString(source) {
			found := reSpace.FindString(source)
			UnitStack = append(UnitStack, UnitType{Str: found, Type: "Space"})
			source = strings.TrimPrefix(source, found)
			notFound = false
		}
		for i, list := range [][]*regexp.Regexp{reLongVowels, reShortVowels, reConsonants} {
			for _, re := range list {
				if re.MatchString(source) {
					found := re.FindString(source)
					UnitStack = append(UnitStack, UnitType{Str: found, Type: whichList(i)})
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
			car := strings.Split(source, "")[0]
			source = strings.TrimPrefix(source, car)
			fmt.Printf("'%s' : Unknown\n", car)
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
		if unit.Type == "ShortVowel" &&
		!(NextUnit.Type == "Consonant" && NextNextUnit.Type == "Consonant") &&
		!(NextUnit.Str == "ṁ" || NextUnit.Str == "Ṁ") &&
		!(contains(AspiratedConsonants, NextUnit.Str) &&
		PrevUnit.Type != "Consonant") {
			// case HO-mi
		} else if unit.Type == "LongVowel" &&
		!(NextUnit.Type == "Consonant" && NextNextUnit.Type == "Consonant") &&
		!(NextUnit.Str == "ṁ" || NextUnit.Str == "Ṁ") {
			// case SAM-mā
		} else if unit.Type == "Consonant" &&
		NextUnit.Type == "Consonant" &&
		(PrevUnit.Type == "LongVowel" || PrevUnit.Type == "ShortVowel") {
			// case DHAM-mo
		} else if contains(UnstoppingCar, strings.ToLower(unit.Str)) &&
		!(NextUnit.Type == "LongVowel" || NextUnit.Type == "ShortVowel") &&
		(PrevUnit.Type == "LongVowel" || PrevUnit.Type == "ShortVowel") {
		} else {
			unit.Closing = false
		}
		//----
		if (PrevUnit.Type == "Punctuation" || PrevUnit.Type == "Space") &&
		!(unit.Type == "Punctuation" || unit.Type == "Space") {
			Syllables = append(Syllables, Syllable)
			Syllable = *new(SyllableType)
		}
		Syllable.Units = append(Syllable.Units, unit)
		if unit.Closing ||
		((NextUnit.Type == "Punctuation" || NextUnit.Type == "Space") &&
		!(unit.Type == "Punctuation" || unit.Type == "Space")) {
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
			if (unit.Type == "Punctuation" || unit.Type == "Space") {
				Syllable.Irrelevant = true
			}
			if (unit.Type == "ShortVowel" &&
			(NextUnit.Str == "ṁ" || NextUnit.Str == "Ṁ")) ||
			(unit.Type == "ShortVowel" && NextUnit.Type == "Consonant" && NextUnit.Closing) ||
			(unit.Type == "LongVowel") {
				Syllable.isLong = true
			}
			if contains(UnstoppingCar, strings.ToLower(unit.Str)) &&
			unit.Closing ||
			(unit.Type == "LongVowel" && unit.Closing) {
				Syllable.NotStopped = true
			}
			if contains(HighToneFirstCar, strings.ToLower(Syllable.Units[0].Str)) {
				Syllable.hasHighToneFirstCar = true
			}
			if Syllable.hasHighToneFirstCar &&
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
			if !Syllable.TrueHigh && unit.Type == "ShortVowel" &&
			contains(OptionalHighToneFirstCar, strings.ToLower(Syllable.Units[0].Str)) {
				if unit.Closing ||
				!contains(UnstoppingCar, strings.ToLower(NextUnit.Str)) {
					Syllable.OptionalHigh = true
				}
				if Syllable.OptionalHigh && !wantHtml && *wantOptionalHigh {
					for k, unit := range Syllable.Units {
						Syllable.Units[k].Str = strings.ToUpper(unit.Str)
					}
					
				}
			}
			//---
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
		if Syllable.whichTone() != "none" {
			class += Syllable.whichTone()
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
			} else if rePunc.MatchString(unit.Str) && reIsNotExeptPunc.MatchString(unit.Str) {
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
		// check: unless it appear as the last char in a row visually,
		// if char is letter and  next char too, add separator
		// FIXME <br> actually only occurs after a point/comma, fix or rm
		if len(Syllables) > h+1 &&
		Syllables[h+1].Units[0].Str != "<br>" &&
		isLetterChar(Syllable.Units[len(Syllable.Units)-1].Type) &&
		isLetterChar(Syllables[h+1].Units[0].Type) {
			buf.WriteString(separator) 
		}
	}
	if wantHtml {
		buf.WriteString("</body></html>")
	}
	fo, err := os.Create(*out)
	check(err)
	defer fo.Close()
	w := bufio.NewWriter(fo)
	_, err = buf.WriteTo(w)
	check(err)
	w.Flush()
	fmt.Println("Done")
}


func appendClass(class, s string) string {
	if class != "" {
		class += " "
	}
	class += s
	return class
}

func isLetterChar(s string) bool {
	b := true
	if s == "Punctuation" {
		b = false
	} else if s == "Space" {
		b = false
	}
	return b
}

func whichList(i int) string {
	switch i {
	case 0:
		return "LongVowel"
	case 1:
		return "ShortVowel"
	case 2:
		return "Consonant"
	}
	return ""
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

func contains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
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

func isFlagPassed(name string) bool {
    found := false
    flag.Visit(func(f *flag.Flag) {
        if f.Name == name {
            found = true
        }
    })
    return found
}