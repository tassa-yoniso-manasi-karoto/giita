CLI tool to format latin script Pali for chanting in the Thai style.

This script was written with the Makhot style of the Dhammayut order (ธรรมยุติกนิกาย) in mind but its output can also be used for the Saṁyok style of the Maha Nikaya order.
This script implements the tone rules provided at https://www.dhammatalks.org/ebook_index.html#tone_guide (v151003)

<img src="https://github.com/tassa-yoniso-manasi-karoto/giita/blob/main/img.webp">
<p align="center">Try it on CodePen: https://codepen.io/tassa-yoniso-manasi-karoto/pen/poaOedv</p>

### Please note:
- the **input file needs to be UTF-8 encoded**. Windows users, especially prior to windows 10, should be aware of this.
- without arguments, giita will process the "input.txt" file located in the folder of executable and output it there in a "output.htm" with HTML formatting
- in the HTML format, no styling is hardcoded and **all formatting can be changed in the embedded CSS**
- the above mentionned guide does not provide a way to identify syllables which can get the optional low tone therefore this is not implemented
- optional high tones are disabled by default and *will* result in false positives
- this script is provided here "for posterity" and will not be actively maintained
- To chant in the Saṁyok style, uncomment  "/\*font-weight: 300;\*/" in the embedded CSS to create a visual difference between long and short syllables.
- keep in mind that syllable delimitations and tone rules can be subject to exceptions and the guidance provided by the formatting may not be accurate!

### Usage of giita:
    -i string
      path of input UTF-8 encoded text file (default: directory of executable + "input.txt")
    -o string
      path of output file (default: directory of executable + "output.htm")
    -optionalhigh
      requires -t, it formats optional high tones with capital letters just like true high tones (turn on with -optionalhigh=true)
    -t
      use raw text format instead of HTML for the output file (turn on with -t=true)
      
Download: [Releases](https://github.com/tassa-yoniso-manasi-karoto/giita/releases)
