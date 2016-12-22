# giftitler

`giftitler` is a command line application to write subtitles onto GIFs automatically. Given an SRT file, a TTF, an animated
GIF, and the timestamp (starting from the same 0 point the SRT file starts from), it can write the appropriate subtitles to
each framge of the GIF, effectively "burning in" the subtitles.

## Usage

```
Usage of ./giftitler:
  -font string
    	path to font to use
  -font-dpi float
    	dpi to write text in (default 72)
  -font-size float
    	font size (in points) to write text in
  -font-stroke-weight int
    	weight of the font stroke (outline) in pixels (default 3)
  -gif string
    	path to gif
  -out string
    	path to write result (default "output.gif")
  -subtitles string
    	path subtitles file to pull text from
  -subtitles-offset duration
    	timestamp of the subtitles the gif starts from
```

If `--font-size` is not specified, `giftitler` will automatically detect an appropriate font size for the width of the GIF.

## Current limitations

* If a line of text would be too long to fit on a frame, no text at all is written to that frame, and no error is logged. This
  should only happen if a font size is manually specified.
* Some SRT files that try to fit too much text on the screen at once may need some massaging.
* Only TTF fonts are supported.
* Portrait (taller-than-wide) GIFs are going to confuse the text-sizing algorithm and make it make suboptimal choices.
* No special parsing is done on the SRT file--what you see in it is what you're going to get, HTML tags and all. Massage
  accordingly.
