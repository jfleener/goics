package goics

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"time"
)

// Line endings
const (
	CRLF   = "\r\n"
	CRLFSP = "\r\n "

	ICSPropertyRecurrenceRule     = "RRULE"
)

// NewComponent returns a new Component and setups
// and setups Properties map for the component
// and also allows more Components inside it.
// VCALENDAR is a Component that has VEVENTS,
// VEVENTS can hold VALARMS
func NewComponent() *Component {
	return &Component{
		Elements:   make([]Componenter, 0),
		Properties: make(map[string]string),
	}
}

// Component is the base type for holding a
// ICal datatree before serilizing it
type Component struct {
	Tipo       string
	Elements   []Componenter
	Properties map[string]string
}

// Writes the component to the Writer
func (c *Component) Write(w *ICalEncode) {
	w.WriteLine("BEGIN:" + c.Tipo + CRLF)

	// Iterate over component properites
	var keys []string
	for k := range c.Properties {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		val := c.Properties[key]
		switch key {
		case ICSPropertyRecurrenceRule:
			w.WriteLine(WriteStringField(key, val, false))
		default:
			w.WriteLine(WriteStringField(key, val, true))
		}
	}

	for _, xc := range c.Elements {
		xc.Write(w)
	}

	w.WriteLine("END:" + c.Tipo + CRLF)
}

// SetType of the component, as
// VCALENDAR VEVENT...
func (c *Component) SetType(t string) {
	c.Tipo = t
}

// AddComponent to the base component, just for building
// the component tree
func (c *Component) AddComponent(cc Componenter) {
	c.Elements = append(c.Elements, cc)
}

// AddProperty ads a property to the component
func (c *Component) AddProperty(key string, val string) {
	c.Properties[key] = val
}

// ICalEncode is the real writer, that wraps every line,
// in 75 chars length... Also gets the component from the emmiter
// and starts the iteration.
type ICalEncode struct {
	w io.Writer
}

// NewICalEncode generates a new encoder, and needs a writer
func NewICalEncode(w io.Writer) *ICalEncode {
	return &ICalEncode{
		w: w,
	}
}

// Encode the Component into the ical format
func (enc *ICalEncode) Encode(c ICalEmiter) {
	component := c.EmitICal()
	component.Write(enc)
}

// LineSize of the ics format
var LineSize = 75

// WriteLine in ics format max length = LineSize
// continuation lines start with a space.
func (enc *ICalEncode) WriteLine(s string) {
	if len(s) <= LineSize {
		_, err := io.WriteString(enc.w, s)
		if err != nil {
			// TODO: Handle this
		}
		return
	}

	// The first line does not begin with a space.
	firstLine := truncateString(s, LineSize-len(CRLF))
	_, err := io.WriteString(enc.w, firstLine+CRLF)
	if err != nil {
		// TODO: Handle this
	}

	// Reserve three bytes for space + CRLF.
	lines := splitLength(s[len(firstLine):], LineSize-len(CRLFSP))
	for i, line := range lines {
		if i < len(lines)-1 {
			_, err = io.WriteString(enc.w, " "+line+CRLF)
			if err != nil {
				// TODO: Handle this
			}
		} else {
			// This is the last line, don't append CRLF.
			_, err = io.WriteString(enc.w, " "+line)
			if err != nil {
				// TODO: Handle this
			}
		}
	}
}

// FormatDateField returns a formated date: "DTEND;VALUE=DATE:20140406"
func FormatDateField(key string, val time.Time) (string, string) {
	return key + ";VALUE=DATE", val.Format("20060102")
}

// FormatDateTimeField in the form "X-MYDATETIME;VALUE=DATE-TIME:20120901T130000"
func FormatDateTimeField(key string, val time.Time) (string, string) {
	return key + ";VALUE=DATE-TIME", val.Format("20060102T150405")
}

// FormatDateTime as "DTSTART:19980119T070000Z"
func FormatDateTime(key string, val time.Time) (string, string) {
	return key, val.UTC().Format("20060102T150405Z")
}

// WriteStringField UID:asdfasdfаs@dfasdf.com
func WriteStringField(key string, val string, escapeValChars bool) string {
	if escapeValChars {
		return fmt.Sprintf("%v:%v%v", strings.ToUpper(key), escapeChars(val), CRLF)
	}
	return fmt.Sprintf("%v:%v%v", strings.ToUpper(key), val, CRLF)
}

func escapeChars(s string) string {
	s = strings.Replace(s, "\\", "\\\\", -1)
	s = strings.Replace(s, ";", "\\;", -1)
	s = strings.Replace(s, ",", "\\,", -1)
	s = strings.Replace(s, "\n", "\\n", -1)
	return s
}
