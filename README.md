CLI tool to format latin script Pali for chanting in the Thai style.

This script was written with the Makhot style of the Dhammayut order in mind but its output can also be used for the Saṁyok style of the Maha Nikaya order.
This script implements the tone rules provided at https://www.dhammatalks.org/ebook_index.html#tone_guide (v151003)

<img src="https://github.com/tassa-yoniso-manasi-karoto/giita/blob/main/img.webp">
<p align="center">Try it on CodePen: https://codepen.io/tassa-yoniso-manasi-karoto/pen/poaOedv</p>

### Please note:
- the **input file needs to be UTF-8 encoded**. Windows users, especially prior to windows 10, should be aware of this.
- without arguments, giita will process the "input.txt" file located in the folder of executable and output it there in a "output.htm" with HTML formatting
- in the HTML format, no formatting is hardcoded and **all formatting can be changed in the embedded CSS**
- the above mentionned guide does not provide a way to identify syllables which can get the optional low tone therefore this is not implemented
- optional high tones are disabled by default and *will* result in false positives
- this script is provided here "for posterity" and will not be actively maintained
- To chant in the Saṁyok style, uncomment  "/\*font-weight: 300;\*/" in the embedded CSS to create a visual difference between long and short syllables.
- keep in mind that syllable delimitations and tone rules can be subject to exceptions and the guidance provided by the formatting is not always accurate!

### Known issues
- ambiguous cases "ay" cases like "viheṭhayanto" where it could referer either to the "ay" long vowel or an "a" followed by a "y"
- \*brāhma\* words.
- nh* exceptions: nhārū, nhāyeyya

### Usage of giita:
        -c string
    	allow comments in input file and specify which characters marks
    	respectively the beginning and the end of a comment, separated
    	by a colon
        -css string
    	will overwrite all CSS and CSS-related options with the CSS file at
    	this path.
        -d	dark mode, will use a white font on a dark background
        -f int
    	set font size (default 34)
        -hint
    	suggests hints on where to catch one's breath in long compound words.
    	(disable with -hint=false) (default true)
        -i string
    	path of input UTF-8 encoded text file
    	 (default: "input.txt" in directory of executable)
        -l int
    	set how many linebreaks will be created from a single linebreak in
    	the input file. Advisable to use 2 for smartphone/tablet/e-reader.
    	 (default 1)
        -o string
    	path of output file
    	 (default: "output.htm" in directory of executable)
        -optionalhigh
    	requires -t, it formats optional high tones with capital letters
    	just like true high tones
        -re string
    	on the fly regular expression deletion. Uses Golang (Google RE2) format.
    	See https://github.com/google/re2/wiki/Syntax, https://regex101.com/
        -t	use raw text instead of HTML for the output file

Download: [Releases](https://github.com/tassa-yoniso-manasi-karoto/giita/releases)
