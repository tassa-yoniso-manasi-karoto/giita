// ☸

package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"html"
	"io/fs"
	"math"
	"os"
	"path"
	"regexp"
	"runtime/pprof"
	"sort"
	"strings"
	"time"
	"unicode/utf8"
	"github.com/gookit/color"
)

const version = "v1.2.9"

// const reference string = "a-ra-haṁ, abhi-vā-de-mi, su-pa-ṭi-pan-no, sam-bud-dho, svāk-khā-to, tas-sa, met-ta, a-haṁ, ho-mi, a-ve-ro, dham-mo, sam-mā, a-haṁ, kho, khan-dho, Ṭhā-nis-sa-ro, ya-thā, sey-yo, ho-ti, hon-ti, sot-thi, phoṭ-ṭhab-ba, khet-te, ya-thāj-ja, cī-va-raṁ, pa-ri-bhut-taṁ, sa-ra-naṁ, ma-kasa, pa-ṭha-mā-nus-sa-ti, Bha-ga-vā, sam-bud-dhas-sa, kit-ti-sad-do, a-ha-mā-da-re-na, khet-te, A-haṁ bhan-te sam-ba-hu-lā nā-nā-vat-thu-kā-ya pā-cit-ti-yā-yo ā-pat-ti-yo ā-pan-no tā pa-ṭi-de-se-mi. Pas-sa-si ā-vu-so? Ā-ma bhan-te pas-sā-mi. Ā-ya-tiṁ ā-vu-so saṁ-va-rey-yā-si. Sā-dhu suṭ-ṭhu bhan-te saṁ-va-ris-sā-mi."

// NOTE: makasa → "ma-kasa" = presumed to be an exception
// var test string = "arahaṁ, abhivādemi, supaṭipanno, sambuddho, svākkhāto, tassa, metta, ahaṁ, homi, avero, dhammo, sammā, ahaṁ, kho, khandho, Ṭhānissaro, yathā, seyyo, hoti, honti, sotthi, phoṭṭhabba, khette, yathājja, cīvaraṁ, paribhuttaṁ, saranaṁ, makasa, paṭhamānussati, Bhagavā, sambuddhassa, kittisaddo, ahamādarena, khette, Ahaṁ bhante sambahulā nānāvatthukāya pācittiyāyo āpattiyo āpanno tā paṭidesemi. Passasi āvuso? Āma bhante passāmi. Āyatiṁ āvuso saṁvareyyāsi. Sādhu suṭṭhu bhante saṁvarissāmi."

/*
TODO
	make test files
COULD
	pretty print? (i.e. wrap quotes nicely)
	diff against ↓ to find exceptions (all long falling tones?): METta, viMOKkha, sometime also pāṭiMOKkhe
	https://www.dhammatalks.org/books/ChantingGuide/Section0000.html,
	use something like /digitalpalireader/_dprhtml/js/analysis_function.js to
		(1) improve accuracy of syllable splitting;
		(2) be able to prefer splitting really long compound words at word boundaries
		(would require a Go rewrite so that's not going to happen)
*/

const (
	LongVwl = iota
	ShortVwl
	Cons
	ElisionMark
	Punct
	Space
	Other
)

var (
	FrequentElisionMark = []string{"’"} //, "'"}
	
	FrequentPunc       = []string{".",",", "\"", "“", "”", "’", ";", "?"}
	rePunc             = regexp.MustCompile(`^\pP+`)
	reIsNotExceptPunct = regexp.MustCompile(`^[^-“’„"\(\)\[\]«'‘‚-]+`)
	
	// third index, NO-BREAK SPACE [NBSP], isn't part of "\s"
	FrequentSpace      = []string{" ", "\n", " "}
	reSpace            = regexp.MustCompile(`(?s)^\s+`)
	
	FrequentOther      = []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}
	IrrelevantTypes    = []int{Punct, Space, Other}

	LongVwls    = []string{"ā", "e", "ī", "o", "ū"} /*, "ay"} too many false positives */
	ShortVwls   = []string{"a", "i", "u"}
	VowelTypes  = []int{LongVwl, ShortVwl}
	reLongVwls  []*regexp.Regexp
	reShortVwls []*regexp.Regexp

	C = []string{"bh", "dh", "ḍh", "gh", "jh", "kh", "ph", "th", "ṭh", "sm",
		"ch", "c", "g", "h", "s", "j", "r", "p", "b", "d", "k", "t", "ṭ",
		 "m", "ṁ", "ṃ", "n", "ñ", "ṅ", "ṇ", "y", "l", "ḷ", "ḍ", "v"}
	reC []*regexp.Regexp

	NeverLastPos      = []string{"bh", "dh", "ḍh", "gh", "jh", "kh", "ph", "th", "ṭh", "v", "r"}
	UnstopChar        = []string{"n", "ñ", "ṅ", "ṇ", "m", "ṁ", "ṃ", "l", "ḷ", "y"}
	HighToneFirstChar = []string{"ch", "th", "ṭh", "kh", "ph", "sm", "s", "h"}
	OptHighFirstChar  = []string{"v", "bh", "r", "n", "ṇ", "m", "y"}

	wantDebug                                        debugType
	CurrentDir                                       string
	Orange, Green, ANSIReset                         string
	in, out, refCmt, UserCSSPath, UserRe, debugRaw   *string
	wantNewlineNum, wantFontSize                     *int
	wantHint                                         *float64
	wantTxt, wantOptionalHigh, wantDark, wantVersion *bool
	wantSamyok, wantNoto                             *bool
	wantHtml                                         = true

	DefaultTemplate = `<!DOCTYPE html> <html><head>
<meta charset="UTF-8">
<style>
%s
</style></head>
<body>`
	CSS = `
body {
  font-size: %dpx;
  line-height: 1.4em;
  letter-spacing: -0.04em;
  word-spacing: 0.40em;
}

.w {
  white-space: nowrap;
}

.s::before{
  content: "⸱";
  color: #646464;
}

.punct::after{
  content: "█";
  color: orangered; /*#5c5c5c;*/
}

.truehigh{
  font-weight: bold;
  vertical-align: 13%%;
}

.long {
}

.short {
}

.hint {
  text-decoration-line: underline;
  text-decoration-style: wavy;
}
.hint::after{
  content: "|";
  color: orangered; /*#5c5c5c;*/
}

.comment {
  background: lightgrey;
  font-style: italic;
  word-spacing: normal;
}

.optionalhigh{
  /*font-style: italic;*/
}
`
)

type debugType struct {
	Perf, Hint, Rate, Parser, Stats, CSS, List bool
	Time                                       time.Time
}

type UnitType struct {
	Str     string
	Type    int
	Len     string
	Closing bool
}

type SyllableType struct {
	Units                                    []UnitType
	IsLong, NotStopped, HasHighToneFirstChar bool
	Irrelevant, Hint                         bool
	TrueHigh, OptionalHigh                   bool
}

type SegmentType []SyllableType

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
	if color.DetectColorLevel() == color.LevelRgb {
		Orange, Green, ANSIReset = "\033[38;5;208m", "\033[38;5;2m", "\033[0m"
	}
	// STRING
	in = flag.String("i", CurrentDir+"/input.txt", "path of input UTF-8 encoded text file\n")
	out = flag.String("o", CurrentDir+"/output.htm", "path of output file\n")
	UserCSSPath = flag.String("css", "", "will overwrite all CSS and CSS-related options with the CSS file at\nthis path.")
	UserRe = flag.String("re", "", "on the fly regular expression deletion. Uses Golang (Google RE2) format."+
		"\nSee https://github.com/google/re2/wiki/Syntax, https://regex101.com/")
	refCmt = flag.String("c", "[:]", "allow comments in input file and specify which "+
		"characters marks\nrespectively the beginning and the end of a comment, separated\nby a colon")
	debugRaw = flag.String("debug", "", "select desired modules:\n \"perf:hint:rate:parser:css:stats:list_pprofFileSuffix\"")
	// BOOL
	wantTxt = flag.Bool("t", false, "use raw text instead of HTML for the output file")
	wantOptionalHigh = flag.Bool(
		"optionalhigh", false, "requires -t, it formats optional "+
			"high tones with capital letters\njust like true high tones")
	wantDark = flag.Bool("d", false, "dark mode, will use a white font on a dark background")
	wantSamyok = flag.Bool("samyok", false, "use CSS optimized for chanting in the Samyok style")
	wantNoto = flag.Bool("noto", false, "use noto-fonts and a slightly greater font weight for long syllables")
	wantVersion = flag.Bool("version", false, "output version information and exit")
	// INT
	wantNewlineNum = flag.Int("l", 1, "set how many linebreaks will be created from a single "+
		"linebreak in\nthe input file. Advisable to use 2 for smartphone/tablet/e-reader.\n")
	wantFontSize = flag.Int("f", 34, "set font size")
	// FLOAT
	wantHint = flag.Float64("hint", 4.5, "suggests hints on where to catch one's breath in long compound words or\n"+
		"list/enumerations missing proper punctuation."+
			"\nSuperior values increase sensitivity as to what counts as a list."+
				"\nReasonable range between 4 and 8, disabled with -hint 0.")
	flag.Parse()
	if *wantVersion {
		fmt.Println("giita", version)
		os.Exit(0)
	}
	if len(*refCmt) != 3 {
		panic("You provided an invalid input of comment marks.")
	}
	CSS = fmt.Sprintf(CSS, *wantFontSize)
	if *debugRaw != "" {
		suffix := parseDbg(*debugRaw)
		dat, err := os.ReadFile(CurrentDir + "/debug.css")
		if wantDebug.CSS && !errors.Is(err, fs.ErrNotExist) {
			CSS = string(dat)
		}
		if wantDebug.Perf {
			f2, _ := os.Create("cpu" + suffix + ".prof")
			pprof.StartCPUProfile(f2)
			defer func() {
				f, _ := os.Create("mem" + suffix + ".mprof")
				pprof.WriteHeapProfile(f)
				f.Close()
				pprof.StopCPUProfile()
			}()
		}
	}
	if wantDebug.Perf || wantDebug.Stats {
		wantDebug.Time = time.Now()
		defer func(){fmt.Println(time.Since(wantDebug.Time))}()
	}
	page := fmt.Sprintf(DefaultTemplate, CSS)
	if *wantDark {
		page = strings.Replace(page, "body {", "body {\n  background: black;\n  color: white;", 1)
		page = strings.Replace(page, ".s::before{\n  content: \"⸱\";\n  color: #646464;",
			".s::before{\n  content: \"⸱\";\n  color: #858585;", 1)
		page = strings.Replace(page, ".comment {\n  background: lightgrey;", ".comment {\n  background: darkgrey;", 1)
	}
	if *wantSamyok {
		page = strings.Replace(page, ".long {", ".long {\n font-weight: bold;", 1)
		page = strings.Replace(page, ".short {", ".short {\n font-weight: 300;", 1)
		page = regexp.MustCompile(`\n\.punct::after[^}]+}\n`).ReplaceAllString(page, "")
	} else if *wantNoto {
		page = strings.Replace(page, "body {", "body {\n  font-family: \"Noto Sans\";", 1)
		page = strings.Replace(page, ".long {", ".long {\n font-family: \"Noto Sans Medium\" !important;", 1)
	}
	if *UserCSSPath != "" {
		dat, err := os.ReadFile(*UserCSSPath)
		check(err)
		page = fmt.Sprintf(DefaultTemplate, string(dat))
	}
	// the \n makes the html source somewhat readable
	newline := "<br>\n"
	separator := "<span class=s></span>"
	if *wantTxt {
		wantHtml = false
		separator = "⸱"
		newline = "\n"
		page = ""
		if !isFlagPassed("o") {
			*out = CurrentDir + "/output.txt"
		}
	}
	newline = strings.Repeat(newline, *wantNewlineNum)
	fmt.Println("In:", *in)
	fmt.Println("Out:", *out)
	dat, err := os.ReadFile(*in)
	if errors.Is(err, fs.ErrNotExist) {
		fmt.Println(Orange + "Input \"giita -h\" to display the command line usage.\n" + ANSIReset)
	}
	check(err)
	src := string(dat)
	if *UserRe != "" {
		re := regexp.MustCompile(*UserRe)
		src = re.ReplaceAllString(src, "")
	}
	var cmts []string
	if isFlagPassed("c") {
		reCmt := regexp.MustCompile(fmt.Sprintf(`(?s)%s.*?%s`, regexp.QuoteMeta((*refCmt)[0:1]), regexp.QuoteMeta((*refCmt)[2:3])))
		cmts = reCmt.FindAllString(src, -1)
		src = reCmt.ReplaceAllString(src, "𓃰")
	}
	src = strings.ReplaceAll(src, "ṃ", "ṁ")
	src = strings.ReplaceAll(src, "Ṃ", "Ṁ")
	// chunks from long compound words need to be reunited or will be treated as separate
	src = strings.ReplaceAll(src, "-", "")
	if strings.Contains(src, "...") || strings.Contains(src, "…") {
		fmt.Printf("%sThe input contains %d occurence(s) of '...' or "+
			"'…' which usually indicates an ellipsis of a repeated formula. "+
			"This could result in an incomplete chanting text.%s\n",
			Orange, strings.Count(src, "...")+strings.Count(src, "…"), ANSIReset)
	}
	RawUnits := Parser(src)
	Syllables := SyllableBuilder(RawUnits)
	Syllables = SetTones(Syllables)
	if *wantHint != 0 {
		Segments := SegmentBuilder(Syllables)
		SegmentProcessed := 0
		for i, Segment := range Segments {
			Segments[i] = Segment.MakeHint(i, &SegmentProcessed)
		}
		if wantDebug.Hint || wantDebug.Stats {
			fmt.Printf("[hint] added hint(s) in %.1f%% of all segments (%d/%d)\n",
				float64(SegmentProcessed)/float64(len(Segments))*100, int(SegmentProcessed), len(Segments))
		}
		Syllables = []SyllableType{}
		for _, Segment := range Segments {
			Syllables = append(Syllables, []SyllableType(Segment)...)
		}
	}
	// TODO Rewrite this writer with Segment instead of Syllables, <div> instead of <br>
	buf := bytes.NewBufferString(page)
	span := "<span class=\"%s\">"
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
			if Syllable.Hint {
				class = appendClass(class, "hint")
			}
			if class != "" {
				fmt.Fprintf(buf, span, class)
			}
		}
		/*if Syllable.Hint {
			buf.WriteString(html.EscapeString("@"))
		}*/
		for _, unit := range Syllable.Units {
			if strings.Contains(unit.Str, "\n") {
				// FIXME one empty newline = two \n, so -l 2 is a factor 2 operation, need a smaller step
				buf.WriteString(strings.ReplaceAll(unit.Str, "\n", newline))
			} else if reSpace.MatchString(unit.Str) {
				buf.WriteString(" ")
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
		if h+1 < len(Syllables) {
			lastUnit := Syllable.Units[len(Syllable.Units)-1]
			NextSylFirstUnit := Syllables[h+1].Units[0]
			if !contains(IrrelevantTypes, lastUnit.Type) && !contains(IrrelevantTypes, NextSylFirstUnit.Type) {
				buf.WriteString(separator)
			}
		}
	}
	if wantHtml {
		buf.WriteString("</body></html>")
	}
	outstr := buf.String()
	if isFlagPassed("c") {
		for _, cmt := range cmts {
			if wantHtml {
				cmt = html.EscapeString(cmt)
				cmt = "<span class=comment>" + cmt + "</span>"
			}
			outstr = strings.Replace(outstr, "𓃰", cmt, 1)
		}
	}
	err = os.WriteFile(*out, []byte(outstr), 0644)
	check(err)
	fmt.Println("Done")
}

// first try w/ the list of short, long vwl and cons in lower case,
// frequently encountered punctuation, space/newline/nbsp, digits then try with regex.
// On average 97% of the parsing can be accomplised without using regex.
func Parser(src string) (RawUnits []UnitType) {
	f := func(m string, src *string, i int) (u UnitType) {
		u = UnitType{Str: m, Type: i}
		*src = strings.TrimPrefix(*src, m)
		return
	}
	var StrMatch, Done float64
	Lists := [][]string{LongVwls, ShortVwls, C, FrequentElisionMark, FrequentPunc, FrequentSpace, FrequentOther}
	reLists := [][]*regexp.Regexp{reLongVwls, reShortVwls, reC, nil, []*regexp.Regexp{rePunc}, []*regexp.Regexp{reSpace}}
	// Note: rewrite with generics = func 13% more CPU intensive + more than twice the length
	// archived in commit bc90408901aed35032ced3ca31e3ea7a8ad2cf2e
Outerloop:
	for src != "" {
		found := false
		Done += 1
		for i, list := range Lists {
			for _, s := range list {
				if found = strings.HasPrefix(src, s); found {
					RawUnits = append(RawUnits, f(s, &src, i))
					StrMatch += 1
					continue Outerloop
				}
			}
		}
		for i, list := range reLists {
			for _, re := range list {
				if found = re.MatchString(src); found {
					m := re.FindString(src)
					RawUnits = append(RawUnits, f(m, &src, i))
					continue Outerloop
				}
			}
		}
		r, _ := utf8.DecodeRuneInString(src)
		char := string(r)
		RawUnits = append(RawUnits, f(char, &src, Other))
		if wantDebug.Parser && char != "𓃰" {
			fmt.Printf("'%s': Non-Pali/Unknown Char (%U)\n", char, r)
		}
	}
	if wantDebug.Parser || wantDebug.Stats {
		fmt.Printf("[parser] String match account for %.1f%% of all matches (%d/%d)\n", StrMatch/Done*100, int(StrMatch), int(Done))
	}
	return
}

func SyllableBuilder(Units []UnitType) []SyllableType {
	var (
		Syllable  SyllableType
		Syllables []SyllableType
	)
	for i, unit := range Units {
		var (
			PrevUnit, NextUnit, NextNextUnit UnitType
			NextNextNextUnit                 UnitType
			notBeforeTwoCons, mustReject     bool
		)
		if i+3 < len(Units) {
			NextNextNextUnit = Units[i+3]
		}
		if i+2 < len(Units) {
			NextNextUnit = Units[i+2]
		} else {
			notBeforeTwoCons = true
		}
		if i+1 < len(Units) {
			NextUnit = Units[i+1]
		}
		if i-1 >= 0 {
			PrevUnit = Units[i-1]
		}
		if !(NextUnit.Type == Cons && NextNextUnit.Type == Cons) {
			notBeforeTwoCons = true
		}
		// get a dangling consonant at the end of the word included in
		// the (currently iterated) previous syllable
		// BUT treat an apostrophe as a vowel sandhi marker when when combining
		// the following word it creates a consistant syllable
		if !contains(VowelTypes, NextNextNextUnit.Type) &&
			(NextUnit.Type == Cons && NextNextUnit.Type > 2 ||
				unit.Type == Cons && NextUnit.Type == ElisionMark) {
			mustReject = true
		}
		//assume true, overwrite everything after setting exceptions
		unit.Closing = true
		// case no further input
		if i+1 == len(Units) {
		// case SU-PA-ṬI-pan-no
		} else if unit.Type == ShortVwl && notBeforeTwoCons &&
			!(strings.ToLower(NextUnit.Str) == "ṁ") &&
			!(contains(NeverLastPos, NextUnit.Str) && !contains(VowelTypes, NextNextUnit.Type)) {
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
		if contains(IrrelevantTypes, PrevUnit.Type) && !contains(IrrelevantTypes, unit.Type) {
			Syllables = append(Syllables, Syllable)
			Syllable = *new(SyllableType)
		}
		Syllable.Units = append(Syllable.Units, unit)
		if unit.Closing && !mustReject || contains(IrrelevantTypes, NextUnit.Type) && !contains(IrrelevantTypes, unit.Type) {
			Syllables = append(Syllables, Syllable)
			Syllable = *new(SyllableType)
		}
	}
	return Syllables
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
			if Syllable.HasHighToneFirstChar && Syllable.NotStopped && Syllable.IsLong {
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

func SegmentBuilder(Syllables []SyllableType) (Segments []SegmentType) {
	Segment := *new(SegmentType)
	for _, Syllable := range Syllables {
		stop := false
		for _, unit := range Syllable.Units {
			if strings.Contains(unit.Str, "\n") ||
				rePunc.MatchString(unit.Str) && reIsNotExceptPunct.MatchString(unit.Str) {
				stop = true
			}
		}
		Segment = append(Segment, Syllable)
		if stop {
			Segments = append(Segments, Segment)
			Segment = *new(SegmentType)
		}
	}
	return
}

func (Segment SegmentType) MakeHint(i int, SegmentProcessed *int) SegmentType {
	SubsegmentTotal := 1
	StatsTotal := Segment.DescribeUpTo(-1)
	BeatsTotal := StatsTotal.Long*2 + StatsTotal.Short
	BeatsDone := 0
	// Target is the number of beats around which we search. Beats = long Syllables increment by 2 and spaces by 0.
	// TargetIdx is the corresponding position expressed as a regular array index.
	// MaxSpread is also expressed in array increments, not in beats, and corresponds to the radius, not the diameter.
	Target := 22
	MaxSpread := 4
	//if wantDebug.Hint { fmt.Printf("BeatsTotal=%d\tBeatsDone=%d\tTarget=%d\n", BeatsTotal, BeatsDone, Target)}
	for BeatsTotal-BeatsDone > Target+MaxSpread {
		// +1 because need 1 more slot for the int that is the target (= the starting point)
		vals := make([]int, MaxSpread*2+1)
		indexes := make([]int, MaxSpread*2+1)
		TargetIdx := Segment.FindIdxMatchingBeats(Target + BeatsDone)
		// if wantDebug.Hint {fmt.Println("TargetIdx", TargetIdx, "Target", Target, "SubsegmentTotal", SubsegmentTotal)}
		idx := 0
		for sp := -MaxSpread; sp <= MaxSpread; sp++ {
			if 0 <= TargetIdx+sp && TargetIdx+sp < len(Segment) {
				vals[idx] = StatsTotal.rate(Segment, TargetIdx, Target, MaxSpread, sp)
				indexes[idx] = TargetIdx + sp
				idx += 1
			}
		}
		Rating := RatingType{sort.IntSlice(vals), indexes}
		sort.Stable(sort.Reverse(Rating))
		if wantDebug.Hint { fmt.Printf("%v\n", Rating) }
		if HighestRatedVal := Rating.IntSlice[0]; HighestRatedVal > 0 {
			//fmt.Printf("SubsegmentTotal=%d\tTarget=%d\tMaxSpread=%d\n", SubsegmentTotal, Target, MaxSpread)
			HighestRatedIdx := Rating.indexes[0]
			if wantDebug.Hint {
				fmt.Print("@")
				for j, Syllable := range Segment {
					if j == HighestRatedIdx { fmt.Print(Green)}
					for _, unit := range Syllable.Units {
						fmt.Print(unit.Str)
					}
					if j == HighestRatedIdx { fmt.Print(ANSIReset) }
				}
			}
			Segment[HighestRatedIdx].Hint = true
			StatsAtPos := Segment.DescribeUpTo(HighestRatedIdx)
			BeatsDone = StatsAtPos.Long*2 + StatsAtPos.Short
			Target = 13
			if SubsegmentTotal == 1 {
				*SegmentProcessed += 1
			}
			SubsegmentTotal += 1
		} else {
			break
		}
	}
	return Segment
}

type StatsType struct {
	Long, Short, Space int
}
// TODO Rewrite these two func with one underlying
func (Segment SegmentType) FindIdxMatchingBeats(Target int) (TargetIdx int) {
	var Stats StatsType
	for i, Syllable := range Segment {
		if Syllable.Irrelevant {
			for _, unit := range Syllable.Units {
				// ContainsAny with NBSP??
				if strings.Contains(unit.Str, " ") {
					Stats.Space += 1
				}
			}
		} else if Syllable.IsLong {
			Stats.Long += 1
		} else if !Syllable.IsLong {
			Stats.Short += 1
		}
		beats := Stats.Long*2 + Stats.Short
		if false && wantDebug.Hint { // FIXME
			fmt.Printf("index=%d    beats=%d   Target+BeatsDone=%d\t", i, beats, Target)
			for _, unit := range Syllable.Units {
				fmt.Printf(unit.Str)
			}
			fmt.Print("\n")
		}
		// long cause +2 increment so an equality may not necessarily occur
		if beats >= Target {
			TargetIdx = i
			break
		}
	}
	return
}

// negative maxIndex means no max index
func (Segment SegmentType) DescribeUpTo(maxIndex int) (Stats StatsType) {
	for i, Syllable := range Segment {
		if i > maxIndex && 0 <= maxIndex {
			break
		} else if Syllable.IsLong {
			Stats.Long += 1
		} else if Syllable.Irrelevant {
			for _, unit := range Syllable.Units {
				// ContainsAny with NBSP??
				if strings.Contains(unit.Str, " ") {
					Stats.Space += 1
				}
			}
		} else if !Syllable.IsLong {
			Stats.Short += 1
		}
	}
	return
}

// https://stackoverflow.com/questions/31141202/get-the-indices-of-the-array-after-sorting-in-golang/31141540#31141540
type RatingType struct {
	sort.IntSlice
	indexes []int
}

func (Rating RatingType) Swap(i, j int) {
	Rating.indexes[i], Rating.indexes[j] = Rating.indexes[j], Rating.indexes[i]
	Rating.IntSlice.Swap(i, j)
}

func (StatsTotal StatsType) rate(Segment SegmentType, TargetIdx int, Target int, MaxSpread int, spread int) int {
	Syllable := Segment[TargetIdx+spread]
	if wantDebug.Rate {
		fmt.Print("\"")
		for _, unit := range Syllable.Units {
			fmt.Print(unit.Str)
		}
		fmt.Print("\"\n")
	}
	if !Syllable.IsLong {
		if wantDebug.Rate {
			fmt.Println("score", 0, "\n")
		}
		return 0

	}
	var (
		score = 100
		listMode = false
		StatsAtPos = Segment.DescribeUpTo(TargetIdx + spread)
		beatsTotal = StatsTotal.Long*2 + StatsTotal.Short
		beatsAtPos = StatsAtPos.Long*2 + StatsAtPos.Short
		PrevTargetIdx = Segment.FindIdxMatchingBeats(beatsAtPos-Target)
		NextTargetIdx = Segment.FindIdxMatchingBeats(beatsAtPos+Target)
		StatsAtPrev = Segment.DescribeUpTo(PrevTargetIdx)
		StatsAtNext = Segment.DescribeUpTo(NextTargetIdx)
	)
	//-------------------------------
	// bonus when likely to be going through a list of items that corresponds to
	// small words and where punctuation is likely missing
	// pācittiya 60               = 83/20 = 4.15	DESIRED
	// pācittiya 84               = 68/16 = 4.25	DESIRED	(case of a tiny list embedded in a sentence)
	// pācittiya 36               = 79/18 = 4.39	?	(case of a tiny list embedded in a sentence)
	// ajjhattikā pathavīdhātu    = 84/19 = 4.42	DESIRED
	// ajjhattikā ākāsadhātu pt2  = 80/16 = 5.00	NOT
	// akankheyya std formula     = 79/8  = 9.88	NOT
	ratio := float64(beatsTotal)/float64(StatsTotal.Space)
	if wantDebug.List && spread == 0 {
		i := 0
		for _, syl := range Segment {
			for _, unit := range syl.Units {
				if i += 1; i > 80 { break }
				fmt.Print(unit.Str)
			}
		}
		fmt.Printf("\nbeatsTotal=%d\t3*Target=%d\t%t\t\tTotal.Space=%d\tratio=%.2f\t%t\n",
			beatsTotal, 3*Target, beatsTotal > 3*Target,
				StatsTotal.Space, ratio, ratio < *wantHint)
	}
	if beatsTotal > 3*Target && ratio < *wantHint {
		if wantDebug.List && spread == 0 {
			fmt.Println(Green + "^ IS LIST ^" + ANSIReset)
		}
		if wantDebug.Rate {
			fmt.Println("\t[rate] List override")
		}
		listMode = true
	}
	//-------------------------------
	// Penalty/bonus for surrounding spaces
	//           │ i negative │ i positive
	// ──────────┼────────────┼──────────────
	// w/ space  │    ---     │     +++
	// ──────────┼────────────┼──────────────
	// w/o space │     +      │      0
	penalty := 0
	MaxSpreadSpace := 5
	if wantDebug.Rate {
		fmt.Println("\t[rate] SpaceAROUND SubPenalties ")
	}
	for i := -MaxSpreadSpace; i <= MaxSpreadSpace; i++ {
		j := TargetIdx + spread + i
		if i != 0 && 0 <= j && j < len(Segment) {
			var fullstring string
			for _, unit := range Segment[j].Units {
				fullstring += unit.Str
			}
			var factor float64
			// negative factor = bonus
			switch { // ContainsAny with NBSP??
			case strings.Contains(fullstring, " ") && i < 0:
				factor = 30.0
			case strings.Contains(fullstring, " ") && i > 1:
				// FIXME was superseded by immediately upcomming space Bonus
				factor = -20.0 * float64(MaxSpreadSpace) / float64(i)
			case !strings.Contains(fullstring, " ") && i < 0:
				factor = -3.0
			}
			// in lists words are likely to be short, a pause suggestion in the
			// middle of a word is unwanted
			if listMode && factor < 0 {
				factor = factor*10
			}
			subPenalty := int(float64(MaxSpreadSpace) / -float64(i) * factor)
			penalty += subPenalty
			if wantDebug.Rate && subPenalty != 0 {
				fmt.Printf("\t\t\t%d due to \"%s\" at index %d (factor %d)",
					-subPenalty, fullstring, i, int(factor))
				if listMode && factor < 0 { fmt.Print(" ", listMode) }
				fmt.Print("\n")
			}
			score -= subPenalty
		}
	}
	if wantDebug.Rate {
		fmt.Println("\t       SpaceAROUND TOTAL Penalty of", -penalty)
	}
	//-------------------------------
	// FIXME is this really useful?
	// the last part of the if checks if we're anywhere inside a long compound word
	if !listMode && StatsTotal.Space-StatsAtPos.Space >= 0 &&
		!(StatsAtNext.Space-StatsAtPos.Space == 0 || StatsAtPrev.Space-StatsAtPos.Space == 0) {
		penalty := (StatsTotal.Space - StatsAtPos.Space) * 50
		score -= penalty
		if wantDebug.Rate {
			fmt.Println("\t[rate] SpaceLEFT Penalty of", -penalty)
		}
	}
	//-------------------------------
	/*if Syllable.TrueHigh {
		score -= 20
	}*/
	//-------------------------------
	// penality for a pause close to the end of the segment
	if i := beatsTotal-beatsAtPos; i < Target+MaxSpread {
		// arbitrary func that provides the desired values: 5^(Target/0.42*(i+1))
		// +1 to prevent a zero division panic. 0.42 = finetuned = careful
		penalty := int(math.Pow(5.0, float64(Target)/(float64(i+1.0)*0.42)))
		score -= penalty
		if penalty < 0 {
			if wantDebug.Rate {
				fmt.Println("\t[rate] Border Penalty: Aborting due to overflow")
			}
			return 0
		}
		if wantDebug.Rate {
			fmt.Println("\t[rate] Border Penalty of", -penalty)
		}
	}
	//-------------------------------
	// bonus for the the syl before the space before a long compound word or within a list
	bonus := 0
	var NextFullstring string
	if i:= TargetIdx+spread+1; i < len(Segment) {
		for _, unit := range Segment[i].Units {
			NextFullstring += unit.Str
		}
	}
	if strings.Contains(NextFullstring, " ") && (StatsAtNext.Space-StatsAtPos.Space == 1 || listMode)  {
		bonus = 150
		score += bonus
		if wantDebug.Rate {
			fmt.Println("\t[rate] with one immediately upcomming space Bonus of", bonus)
		}
	}
	//-------------------------------
	if spread != 0 {
		if spread < 0 {
			spread = -spread
		}
		x := float64(spread)
		// arbitrary func that provides the desired values 12x³−66x²+122x−60
		// 		1→8     2→16     3→35     4→140
		// w/ this it's technically possible to further increase MaxSpread (untested)
		penalty = int(12.0*math.Pow(x, 3)-66.0*math.Pow(x, 2)+122.0*x-60)
		score -= penalty
		if wantDebug.Rate {
			fmt.Printf("\t[rate] Spread Penalty of %d (spread=%d)\n", -penalty, spread)
		}
	}
	if wantDebug.Rate {
		fmt.Println("score", score)
	}
	return score
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

func contains[T comparable](arr []T, i T) bool {
	for _, a := range arr {
		if a == i {
			return true
		}
	}
	return false
}

func parseDbg(debugRaw string) (suffix string) {
	if arr := strings.Split(debugRaw, "_"); len(arr) > 1 {
		suffix = "_" + arr[1]
		debugRaw = strings.TrimSuffix(debugRaw, suffix)
	}
	for _, s := range strings.Split(debugRaw, ":") {
		switch s {
		case "perf":
			wantDebug.Perf = true
		case "hint":
			wantDebug.Hint = true
		case "rate":
			wantDebug.Rate = true
		case "parser":
			wantDebug.Parser = true
		case "stats":
			wantDebug.Stats = true
		case "css":
			wantDebug.CSS = true
		case "list":
			wantDebug.List = true
		}
	}
	return
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

