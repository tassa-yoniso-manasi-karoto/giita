CLI tool to format latin script Pali for chanting in the Thai style.

This program was written with the Makhot style of the Dhammayut order in mind but with the `-samyok` flag its output can also be used for the Saṁyok style of the Maha Nikaya order. You may also use the `-css` flag to pass personalized CSS.

This program implements the tone rules provided at https://www.dhammatalks.org/ebook_index.html#tone_guide (v151003)

<img src="https://github.com/tassa-yoniso-manasi-karoto/giita/blob/main/pic/img.webp">
<p align="center">Try it on CodePen: https://codepen.io/tassa-yoniso-manasi-karoto/pen/poaOedv</p>

## Please note:
- the **input file needs to be UTF-8 encoded**. Windows users, especially prior to windows 10, should be aware of this.
- without arguments, giita will process the "input.txt" file located in the folder of executable and output it there in a "output.htm" with HTML formatting
- in the HTML format, no formatting is hardcoded and **_all_ formatting can be changed through CSS**
- the above mentionned guide does not provide a way to identify syllables which can get the optional low tone therefore this is not implemented
- optional high tones are disabled by default and *will* result in false positives
- this program is provided here "for posterity" and will not be actively maintained
- **To chant in the Saṁyok style,** try passing the `-samyok` flag which will optimize the default CSS for this style
- keep in mind that syllable delimitations and tone rules can be subject to exceptions and the guidance provided by the formatting is not always accurate!

## Known issues
- certain pali grammatical transformations could create unusual syllables, needs testing
- non standard syllables embedded in the middle/end of a word : any \*brāhma, \*nhārū, \*nhāyeyya derivates

## Formatting of short/long syllables
By default there is no formatting to help differentiate short and long syllables.

With the `-samyok` flag the long syllables are in bold and the short are thin. This formatting makes it very easy to tell them apart but it impairs the readability of a word as whole a lot.

<img src="https://github.com/tassa-yoniso-manasi-karoto/giita/blob/main/pic/samyok.webp">

CSS makes it possible to increase the weight *slightly* through the font-weight attribute, however most fonts do not support font-weight other than with bold and thin.
A font that does support all possible variations *and* has full support of IAST characters is noto-fonts. [Download](https://download-directory.github.io/) this [folder](https://github.com/notofonts/noto-fonts/tree/main/hinted/ttf/NotoSans) and install the fonts.

With the `-noto` flag the long syllables can be formatted differently with little disruption. High tones are kept in bold.

<img src="https://github.com/tassa-yoniso-manasi-karoto/giita/blob/main/pic/notomedium.webp">

## Hints

**NOTE: This feature is obsolete now that compound words can be easily be decomposed with by LLM like Claude/ChatGPT/etc**

You may rarely encounter this formatting, a wavy underline with a vertical bar:

<img src="https://github.com/tassa-yoniso-manasi-karoto/giita/blob/main/pic/hints.webp">

This hint is guaranteed to be on a long syllable. It occurs in sentences with a long compound word or in enumerations where punctuation is missing, it is a suggested location to make the syllable extra long in order to have the time to read the rest, or, a short pause to catch one's breath.



## Usage of giita:

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
        -hint float
    	suggests hints on where to catch one's breath in long compound words or
    	list/enumerations missing proper punctuation.
    	Superior values increase sensitivity as to what counts as a list.
    	Reasonable range between 4 and 8, disabled with -hint 0. (default 4.5)
        -i string
    	path of input UTF-8 encoded text file
    	 (default: "input.txt" in directory of executable)
        -l int
    	set how many linebreaks will be created from a single linebreak in
    	the input file. Advisable to use 2 for smartphone/tablet/e-reader.
    	 (default 1)
        -noto
    	use noto-fonts and a slightly greater font weight for long syllables
        -o string
    	path of output file
    	 (default: "output.htm" in directory of executable)
        -optionalhigh
    	requires -t, it formats optional high tones with capital letters
    	just like true high tones
        -re string
    	on the fly regular expression deletion. Uses Golang (Google RE2) format.
    	See https://github.com/google/re2/wiki/Syntax, https://regex101.com/
        -samyok
    	tweak and optimize default CSS for chanting in the Samyok style
        -t	use raw text instead of HTML for the output file
        -th int
    	transliterate from Thai script from:
    	    	1=Pali put down in regular Thai writing
    	    	2=standard Thai Pali as used in Thai Tipitaka
        -version
    	output version information and exit


Download: [Releases](https://github.com/tassa-yoniso-manasi-karoto/giita/releases)
