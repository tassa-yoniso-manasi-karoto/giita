package libgiita

import (
	"regexp"
	"strings"
	"unicode/utf8"
	
	//"github.com/gookit/color"
)


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
	RePunc             = regexp.MustCompile(`^\pP+`)
	ReIsExceptPunct    = regexp.MustCompile(`^[-“’„"\(\)\[\]«'‘‚-]+`)
	
	// third index, NO-BREAK SPACE [NBSP], isn't part of "\s"
	FrequentSpace      = []string{" ", "\n", " "}
	ReSpace            = regexp.MustCompile(`(?s)^\s+`) // \p{Z}  ← better?
	
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
	Irrelevant, Relevant, Hint                         bool // FIXME
	TrueHigh, OptionalHigh                   bool
	ClosingPara                              bool
}

type SegmentType []SyllableType

type ParagraphType []SegmentType



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
	reLists := [][]*regexp.Regexp{reLongVwls, reShortVwls, reC, nil, {RePunc}, {ReSpace}}
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
		/*if wantDebug.Parser { // && char != CmtParaMark && char != CmtSpanMark {
			fmt.Printf("'%s': Non-Pali/Unknown Char (%U)\n", char, r)
		}*/
	}
	/*if wantDebug.Parser || wantDebug.Stats {
		fmt.Printf("[parser] String match account for %.1f%% of all matches (%d/%d)\n", StrMatch/Done*100, int(StrMatch), int(Done))
	}*/
	return
}



// some more work to be done see:
// https://github.com/tassa-yoniso-manasi-karoto/pali-transliteration/blob/ad0786180f524dd9d6f23c0b69de1a22847245f8/kanaMappingMissing.go#L9
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
		if !PrevUnit.IsRelevant() && unit.IsRelevant() {
			Syllables = append(Syllables, Syllable)
			Syllable = *new(SyllableType)
		}
		if !Syllable.Relevant && unit.IsRelevant() {
			Syllable.Relevant = true
		}
		Syllable.Units = append(Syllable.Units, unit)
		if unit.Closing && !mustReject || !NextUnit.IsRelevant() && unit.IsRelevant() {
			Syllables = append(Syllables, Syllable)
			Syllable = *new(SyllableType)
		}
	}
	return Syllables
}




func (unit UnitType) IsRelevant() (b bool) {
	return !contains(IrrelevantTypes, unit.Type)
	/*if !b && ReIsExceptPunct.MatchString(unit.Str) {
		b = true
	}*/
	return
}


func SegmentBuilder(Syllables []SyllableType) (Segments []SegmentType) {
	Segment := *new(SegmentType)
	//capital := true
	for i, Syllable := range Syllables {
		stop := false
		for _, unit := range Syllable.Units {
			/*if capital && *wantCapital {				
				r, _ := utf8.DecodeRuneInString(unit.Str)
				Syllable.Units[0].Str = strings.ToUpper(string(r))
				capital = false
			}*/
			if strings.Contains(unit.Str, "\n") ||
				RePunc.MatchString(unit.Str) && !ReIsExceptPunct.MatchString(unit.Str) {
				stop = true
			}
		}
		Segment = append(Segment, Syllable)
		if stop || i == len(Syllables)-1 {
			Segments = append(Segments, Segment)
			Segment = *new(SegmentType)
			//capital = true
		}
	}
	return
}

func (Syllable *SyllableType) String() (s string) {
	for _, Unit := range Syllable.Units {
		s += Unit.Str
	}
	return
}

func (Segment *SegmentType) String() (s string) {
	for _, Syllable := range *Segment {
		for _, Unit := range Syllable.Units {
			s += Unit.Str
		}
	}
	return
}

func (Segment *SegmentType) SyllableString() (s string) {
	Syllables := *Segment
	for h, Syllable := range Syllables {
		for _, Unit := range Syllable.Units {
			s += Unit.Str
		}
		if h < len(*Segment)-1 {
			lastUnit := Syllable.Units[len(Syllable.Units)-1]
			NextSylFirstUnit := Syllables[h+1].Units[0]
			if lastUnit.IsRelevant() && NextSylFirstUnit.IsRelevant() {
				s += "⸱"
			}
		}
	}
	return
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

