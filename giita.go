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
// const reference string = "a-ra-haṁ, abhi-vā-de-mi, su-pa-ṭi-pan-no, sam-mā, a-haṁ, kho, khan-dho, Ṭhā-nis-sa-ro, ya-thā, sey-yo, ho-ti, hon-ti, sot-thi, phoṭ-ṭhab-ba, khet-te, ya-thāj-ja, cī-va-raṁ, pa-ri-bhut-taṁ, sa-ra-naṁ, ma-kasa, pa-ṭha-mā-nus-sa-ti, Bha-ga-vā, sam-bud-dhas-sa, kit-ti-sad-do, a-ha-mā-da-re-na, khet-te, A-haṁ bhan-te sam-ba-hu-lā nā-nā-vat-thu-kā-ya pā-cit-ti-yā-yo ā-pat-ti-yo ā-pan-no tā pa-ṭi-de-se-mi. Pas-sa-si ā-vu-so? Ā-ma bhan-te pas-sā-mi. Ā-ya-tiṁ ā-vu-so saṁ-va-rey-yā-si. Sā-dhu suṭ-ṭhu bhan-te saṁ-va-ris-sā-mi."

// var test string = "arahaṁ, abhivādemi, supaṭipanno, sammā, ahaṁ, kho, khandho, Ṭhānissaro, yathā, seyyo, hoti, honti, sotthi, phoṭṭhabba, khette, yathājja, cīvaraṁ, paribhuttaṁ, saranaṁ, makasa, paṭhamānussati, Bhagavā, sambuddhassa, kittisaddo, ahamādarena, khette, Ahaṁ bhante sambahulā nānāvatthukāya pācittiyāyo āpattiyo āpanno tā paṭidesemi. Passasi āvuso? Āma bhante passāmi. Āyatiṁ āvuso saṁvareyyāsi. Sādhu suṭṭhu bhante saṁvarissāmi."


var (
	source string
	rePunc = regexp.MustCompile(`^\pP+`)
	reIsNotExeptPunc = regexp.MustCompile(`^[^’-“„	"«'‘‚-]+`)
	reSpace = regexp.MustCompile(`(?s)^\s+`)

	Vowels = []string{"ā", "e", "ī", "o", "ū", "ay", "a", "i", "u"}
	LongVowels = []string{"ā", "e", "ī", "o", "ū", "ay"}
	ShortVowels = []string{"a", "i", "u"}
	reLongVowels []*regexp.Regexp
	reShortVowels []*regexp.Regexp

	Consonants = []string{"bh", "dh", "ḍh", "gh", "jh", "kh", "ph", "th", "ṭh", "c", "g", "h", "s", "j", "r", "p", "b", "d", "k", "t", "ṭ", "m", "ṁ", "ṃ", "n", "ñ", "ṅ", "ṇ", "y", "l", "ḷ", "ḍ", "v"}
	AspiratedConsonants = []string{"bh", "dh", "ḍh", "gh", "jh", "kh", "ph", "th", "ṭh"}
	UnstoppingCar = []string{"n", "ñ", "ṅ", "ṇ", "m", "ṁ", "ṃ", "l", "ḷ", "r", "y"}
	HighToneFirstCar = []string{"s", "h", "ch", "th", "ṭh", "kh", "ph"}
	OptionalHighToneFirstCar = []string{"v", "bh", "r", "n", "ṇ", "m", "y"}
	reConsonants []*regexp.Regexp

	CurrentDir string
	in *string
	out *string
	t, wantOptionalHigh *bool
	wantHtml = true
	htmlpage = `<!DOCTYPE html> <html><head><style>
.separator::before{
  content: "⸱";
}

.punct::after{
  content: "█";
  color: grey;
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

:not(.long):not(body,html){
 /*font-weight: 300;*/
}
</style></head>
<body>`
)


func init() {
	e, err := os.Executable()
	if err != nil {
		fmt.Println(err)
	} else {
		CurrentDir = path.Dir(e)
	}
	in = flag.String("i", CurrentDir + "/input.txt", "path of input UTF-8 encoded text file")
	out = flag.String("o", CurrentDir + "/output.htm", "path of output file")
	t = flag.Bool("t", false , "use raw text format instead of HTML for the output file (turn on with -t=true)")
	wantOptionalHigh = flag.Bool("optionalhigh", false , "used with -t it formats optional high tones with capital letters just like true high tones (turn on with -optionalhigh=true)")
	flag.Parse()
	if *t {
		wantHtml = false
		htmlpage = ""
		if !isFlagPassed("o") {
			*out = CurrentDir + "/output.txt"
		}
	}
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

type UnitType struct {
	Str     string
	Type    string
	Len     string
	Closing bool
}
type SyllableType struct {
	Units                                 []UnitType
	isLong, NotStopped, hasHighToneFirstCar bool
	TrueHigh, OptionalHigh bool
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
		for i, list := range [][]*regexp.Regexp{reLongVowels, reShortVowels, reConsonants} {
			for _, re := range list {
				if re.MatchString(source) {
					found := re.FindString(source)
					UnitStack = append(UnitStack, UnitType{Str: found, Type: whichList(i)})
					source = strings.TrimPrefix(source, found)
					notFound = false
					break
				} else if rePunc.MatchString(source) {
					found := rePunc.FindString(source)
					UnitStack = append(UnitStack, UnitType{Str: found, Type: "Punctuation"})
					source = strings.TrimPrefix(source, found)
					notFound = false
					break
				} else if reSpace.MatchString(source) {
					found := reSpace.FindString(source)
					UnitStack = append(UnitStack, UnitType{Str: found, Type: "Space"})
					source = strings.TrimPrefix(source, found)
					notFound = false
					break
				}
			}
		}
		if notFound {
			found := strings.Split(source, "")[0]
			source = strings.TrimPrefix(source, found)
			fmt.Printf("'%s' : Unknown\n", found)
		}
	}
	var (
		Syllables []SyllableType
		Syllable SyllableType
	)
	for i, unit := range UnitStack {
		var (
			isShortVowel bool
			PrevUnit UnitType
			NextNextUnit UnitType
			NextUnit UnitType
		)
		for _, re := range reShortVowels {
			if re.MatchString(unit.Str) {
				isShortVowel = true
			}
		}
		if len(UnitStack) > i+2 {
			NextNextUnit = UnitStack[i+2]
			NextUnit = UnitStack[i+1]
		}
		if i-1 >= 0 {
			PrevUnit = UnitStack[i-1]
		}
		//assume true, overwrite everything after setting exceptions
		unit.Closing = true
		UnitStack[i] = unit
		if isShortVowel &&
		!(NextUnit.Type == "Consonant" && NextNextUnit.Type == "Consonant") &&
		!(NextUnit.Str == "ṁ" || NextUnit.Str == "Ṁ") &&
		!(contains(AspiratedConsonants, NextUnit.Str) &&
		PrevUnit.Type != "Consonant") {
		} else if unit.Type == "LongVowel" &&
		!(NextUnit.Type == "Consonant" && NextNextUnit.Type == "Consonant") &&
		!(NextUnit.Str == "ṁ" || NextUnit.Str == "Ṁ") {
			// case sam-mā
		} else if unit.Type == "Consonant" &&
		NextUnit.Type == "Consonant" &&
		(PrevUnit.Type == "LongVowel" || PrevUnit.Type == "ShortVowel") {
		} else if contains(UnstoppingCar, strings.ToLower(unit.Str)) &&
		!(NextUnit.Type == "LongVowel" || NextUnit.Type == "ShortVowel") &&
		(PrevUnit.Type == "LongVowel" || PrevUnit.Type == "ShortVowel") {
		} else {
			unit.Closing = false
			UnitStack[i] = unit
		}
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
			if (unit.Type == "ShortVowel" && (NextUnit.Str == "ṁ" || NextUnit.Str == "Ṁ")) ||
				(unit.Type == "ShortVowel" && NextUnit.Type == "Consonant" && NextUnit.Closing) ||
				(unit.Type == "LongVowel") {
				Syllable.isLong = true
			}
			if contains(UnstoppingCar, strings.ToLower(unit.Str)) && unit.Closing ||
				(unit.Type == "LongVowel" && unit.Closing) {
				Syllable.NotStopped = true
			}
			if contains(HighToneFirstCar, strings.ToLower(Syllable.Units[0].Str)) {
				Syllable.hasHighToneFirstCar = true
			}
			if Syllable.hasHighToneFirstCar && Syllable.NotStopped && Syllable.isLong {
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
	buf := bytes.NewBufferString(htmlpage)
	separator := "⸱"
	span := "<span class=\"%s\">"
	if wantHtml {
		separator = "<span class=\"separator\"></span>"
	}
	for h, Syllable := range Syllables {
		class := ""
		if Syllable.whichTone() != "none" {
			class += Syllable.whichTone()
		}
		if Syllable.isLong {
			if class != "" {
				class += " "
			}
			class += "long"
		}		
		if class != "" && wantHtml {
			fmt.Fprintf(buf, span, class)
		}
		for _, unit := range Syllable.Units {
			if strings.Contains(unit.Str, "\n") {
				if wantHtml {
					unit.Str = strings.ReplaceAll(unit.Str, "\n", "<br>")
				}
				buf.WriteString(unit.Str)
			} else if reSpace.MatchString(unit.Str) {
				buf.WriteString(" ")
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


func isLetterChar(s string) (bool) {
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