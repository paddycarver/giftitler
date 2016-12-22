// Package gosrt provides simple access to SubRipText subtitle files (.srt)
package gosrt

import "time"

// Subtitle contains the subtitle data for a .srt file
type Subtitle struct {
	Number int           // A sequential number for this subtitle
	Start  time.Duration // The duration since the start of the file when the subtitle should be shown
	End    time.Duration // The duration since the start of the file when the subtitle should be hidden
	Text   string        // The contents of the subtitle
	Bounds Rectangle     // Optional bounding box of the subtitles. Use .IsEmpty() to determine if it's set
}

/*
Srt file format:

1
00:00:00,000 --> 00:00:00,000
Blah blah


**repeat**

row 1 is a sequential number starting at 1
row 2 is a timscode, formatted as hours:minutes:seconds,milliseconds
row 3 to an empty row is the content

The content may contain formatting>
Bold: <b> </b> or {b} {/b}
Italic: <i> </i> or {i} {/i}
Underline: <u> </u> or {u} {/u}
Font color: <font color="name or #code"> </font> (HTML colors)

Row 2 may contain DVD rectangle positioning and styling, for ex:

00:00:10,500 --> 00:00:13,000  X1:63 X2:223 Y1:43 Y2:58

or

00:00:15,000 --> 00:00:18,000  X1:53 X2:303 Y1:438 Y2:453

(X1:left X2:right Y1:top Y2:bottom)

*/

// Rectangle represents a on-screen axis-aligned bounding box
type Rectangle struct {
	Left   int
	Right  int
	Top    int
	Bottom int
}

// Width calculates and returns the width of a Rectangle
func (r *Rectangle) Width() int {
	return r.Right - r.Left
}


// Height calculates and returns the height of a Rectangle
func (r *Rectangle) Height() int {
	return r.Bottom - r.Top
}

// FromSizes creates a new rectangle with its top-left corner at (x, y)
// and with the specified with and height
func FromSizes(x, y, wid, hgt int) Rectangle {
	return Rectangle{x, y, x + wid, y + hgt}
}

// IsEmpty checks if the Rectangle has the nil value
func (r *Rectangle) IsEmpty() bool {
	return r.Left == r.Right && r.Top == r.Bottom
}
