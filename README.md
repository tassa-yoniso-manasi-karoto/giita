CLI tool to format latin script pali for chanting in the Makhot style of the Dhammayut sect (ธรรมยุติกนิกาย).

This script implements the tone rules provided at https://www.dhammatalks.org/ebook_index.html#tone_guide (v151003)

<img src="https://github.com/tassa-yoniso-manasi-karoto/giita/blob/main/img.webp">

Please note:
- the **input file needs to be UTF-8 encoded**. Windows users, especially prior to windows 10, should be aware of this.
- without arguments, giita will process the "input.txt" file located in the folder of executable and output it there in a "output.htm" with HTML formatting
- in the HTML format, no styling is hardcoded and **all formatting can be changed in the embedded CSS**
- the above mentionned guide does not provide a way to identify syllables which can get the optional low tone therefore this is not implemented
- optional high tones are disabled by default and *will* result in false positives
- this script is provided here "for posterity" and will not be actively maintained
