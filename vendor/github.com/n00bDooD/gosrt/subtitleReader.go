package gosrt

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ValidationStrictness - The strictness of the validation performed
type ValidationStrictness int

const (
	// StrictValidation means that all subtitles have to be fault-free
        StrictValidation ValidationStrictness = iota
	// LenientValidation allows optional srt-elements to be
	// skipped if invalid
	// (and treated as if they were not specified)
	LenientValidation
	// SkipInvalid silently skips invalid subtitles
	SkipInvalid
)

// InputValidationStrictness sets the global level of input validation
var InputValidationStrictness = StrictValidation

func dropCR(data []byte) []byte {
	if len(data) > 0 && data[len(data)-1] == '\r' {
		return data[0 : len(data)-1]
	}
	return data
}

// A bufio.Scanner-function to read a string until there is a double-newline (one empty line).
// Supports reading both LF and CRLF
func scanDoubleNewline(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.Index(data, []byte{'\n', '\n'}); i >= 0 {
		// We have a full double newline-terminated line.
		return i + 2, dropCR(data[0:i]), nil
	} else if i := bytes.Index(data, []byte{'\n', '\r', '\n'}); i >= 0 {
		// We have a full double newline-terminated line.
		return i + 3, dropCR(data[0:i]), nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), dropCR(data), nil
	}
	// Request more data.
	return 0, nil, nil
}

// SubtitleScanner contains the state for scanning
// a .srt-file
type SubtitleScanner struct {
	scanner *bufio.Scanner
	nextSub Subtitle
	err     error
}

// NewScanner creates a new SubtitleScanner from the given io.Reader.
func NewScanner(r io.Reader) *SubtitleScanner {
	s := bufio.NewScanner(r)
	s.Split(scanDoubleNewline)
	return &SubtitleScanner{s, Subtitle{}, nil}
}

// Parse a time formatted as hours:minutes:seconds,milliseconds, strictly formatted as 00:00:00,000
func parseTime(input string) (time.Duration, error) {
	regex := regexp.MustCompile(`(\d{2}):(\d{2}):(\d{2}),(\d{3})`)
	matches := regex.FindStringSubmatch(input)

	if len(matches) < 4 {
		return time.Duration(0), errors.New("Invalid time format")
	}

	hour, err := strconv.Atoi(matches[1])
	if err != nil {
		return time.Duration(0), err
	}
	minute, err := strconv.Atoi(matches[2])
	if err != nil {
		return time.Duration(0), err
	}
	second, err := strconv.Atoi(matches[3])
	if err != nil {
		return time.Duration(0), err
	}
	millisecond, err := strconv.Atoi(matches[4])
	if err != nil {
		return time.Duration(0), err
	}

	return time.Duration(time.Duration(hour)*time.Hour + time.Duration(minute)*time.Minute + time.Duration(second)*time.Second + time.Duration(millisecond)*time.Millisecond), nil
}

// Parse a bounding rectangle definition
// (X1:left X2:right Y1:top Y2:bottom)
func parseRect(input string) (result Rectangle, errResult error) {
	regex := regexp.MustCompile(`X1:(\d+) X2:(\d+) Y1:(\d+) Y2:(\d+)`)
	matches := regex.FindStringSubmatch(input)

	// If validation is set to lenient, set this optional
	// element be ignored if the format is invalid
	if InputValidationStrictness == LenientValidation {
		defer func() {
			if errResult != nil {
				result = Rectangle{0,0,0,0}
				errResult = nil
			}
		}()
	}

	if len(matches) < 4 {
		return Rectangle{0, 0, 0, 0}, errors.New("Invalid bounding format")
	}

	left, err := strconv.Atoi(matches[1])
	if err != nil {
		return Rectangle{0, 0, 0, 0}, err
	}
	right, err := strconv.Atoi(matches[2])
	if err != nil {
		return Rectangle{0, 0, 0, 0}, err
	}
	top, err := strconv.Atoi(matches[3])
	if err != nil {
		return Rectangle{0, 0, 0, 0}, err
	}
	bottom, err := strconv.Atoi(matches[4])
	if err != nil {
		return Rectangle{0, 0, 0, 0}, err
	}

	return Rectangle{left, right, top, bottom}, nil
}

// Scan advances the SubtitleScanner-state, reading a new
// Subtitle-object. Returns true if an object was read
// or false if an error ocurred
func (s *SubtitleScanner) Scan() (wasRead bool) {
	if s.scanner.Scan() {
		var (
			nextnum           int
			start             time.Duration
			end               time.Duration
			subtitletext      string
			subtitleRectangle Rectangle
		)

		// If we are reckless, ignore invalid Subtitles and just
		// find the next one
		if InputValidationStrictness == SkipInvalid {
			defer func() {
				s.err = nil
				if !wasRead {
					wasRead = s.Scan()
					// If we dont' return true here, then
					// the underlying scanner returned false.
					// This means that we either had a read error
					// or the reader is empty
				}
			}()
		}

		str := strings.Split(s.scanner.Text(), "\n")

		for i := 0; i < len(str); i++ {
			text := strings.TrimRight(str[i], "\r")
			switch i {
			case 0:
				num, err := strconv.Atoi(text)
				if err != nil {
					s.err = err
					return false
				}
				nextnum = num
			case 1:
				elements := strings.Split(text, " ")
				if len(elements) >= 3 {
					startTime, err := parseTime(elements[0])
					if err != nil {
						s.err = err
						return false
					}
					endTime, err := parseTime(elements[2])
					if err != nil {
						s.err = err
						return false
					}
					start = startTime
					end = endTime

					if len(elements) >= 7 {
						rect, err := parseRect(strings.Join(elements[3:7], " "))
						if err != nil {
							s.err = err
							return false
						}

						subtitleRectangle = rect
					} else {
						subtitleRectangle = Rectangle{0, 0, 0, 0}
					}
				} else {
					s.err = fmt.Errorf("srt: Invalid timestamp on row: %s", text)
					return false
				}
			default:
				if len(subtitletext) > 0 {
					subtitletext += "\n"
				}
				subtitletext += text
			}
		}

		s.nextSub = Subtitle{nextnum, start, end, subtitletext, subtitleRectangle}

		return true
	}

	return false
}

// Err gets the error of the SubtitleScanner.
// Returns nil if the last error was EOF
func (s *SubtitleScanner) Err() error {
	if s.err != nil {
		return s.err
	}
	return s.scanner.Err()
}

// Subtitle returns the last read subtitle-object
func (s *SubtitleScanner) Subtitle() Subtitle {
	return s.nextSub
}
