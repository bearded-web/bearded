package utils

import (
	"bytes"
	"fmt"
	"github.com/sirupsen/logrus"
	"sort"
	"strings"
	"time"
)

const (
	nocolor = 0
	white   = 37
	red     = 31
	green   = 32
	yellow  = 33
	blue    = 34
	gray    = 90
)

var baseTimestamp time.Time

func init() {
	baseTimestamp = time.Now()
}

func miniTS() int {
	return int(time.Since(baseTimestamp) / time.Second)
}

func miniTSS() string {
	return fmt.Sprintf("%04d", miniTS())
}

func toColor(s string, color int) string {
	return fmt.Sprintf("\x1b[%dm%s\x1b[0m", color, s)
}

func toGray(s string) string {
	return toColor(s, gray)
}

type ExtendedFormatter struct {
	// Set to true to bypass checking for a TTY before outputting colors.
	ForceColors   bool
	DisableColors bool
}

func (f *ExtendedFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	b := &bytes.Buffer{}

	prefixFieldClashes(entry)

	if (f.ForceColors || logrus.IsTerminal()) && !f.DisableColors {
		levelText := strings.ToUpper(entry.Data["level"].(string))[0:1]

		levelColor := blue

		switch entry.Data["level"] {
		case "debug":
			levelColor = white
		case "info":
			levelColor = blue
		case "warning":
			levelColor = yellow
		case "error":
			levelColor = red
		case "fatal":
			levelColor = red
		case "panic":
			levelColor = red
		default:
			levelColor = nocolor
		}
		from := ""
		if source, ok := entry.Data["source"]; ok {
			id := entry.Data["id"].(string)
			if len(id) > 5 {
				id = id[len(id)-5:]
			}
			from = fmt.Sprintf("%s[%s]", source, id)
		}
		fmt.Fprintf(b, "\x1b[%dm%s%04d\x1b[0m \x1b[%dm%-14s\x1b[0m: %-44s ", levelColor, levelText, miniTS(),
			levelColor, from, entry.Data["msg"])

		keys := make([]string, 0)
		for k, _ := range entry.Data {
			if k != "level" && k != "time" && k != "msg" && k != "id" && k != "source" {
				keys = append(keys, k)
			}
		}
		sort.Strings(keys)
		for _, k := range keys {
			v := entry.Data[k]
			fmt.Fprintf(b, " \x1b[%dm%s\x1b[0m=%v", levelColor, k, v)
		}
	} else {
		f.AppendKeyValue(b, "time", entry.Data["time"].(string))
		f.AppendKeyValue(b, "level", entry.Data["level"].(string))
		f.AppendKeyValue(b, "msg", entry.Data["msg"].(string))

		for key, value := range entry.Data {
			if key != "time" && key != "level" && key != "msg" && key != "id" && key != "source" {
				f.AppendKeyValue(b, key, value)
			}
		}
	}

	b.WriteByte('\n')
	return b.Bytes(), nil
}

func (f *ExtendedFormatter) AppendKeyValue(b *bytes.Buffer, key, value interface{}) {
	if _, ok := value.(string); ok {
		fmt.Fprintf(b, "%v=%q ", key, value)
	} else {
		fmt.Fprintf(b, "%v=%v ", key, value)
	}
}

// This is to not silently overwrite `time`, `msg` and `level` fields when
// dumping it. If this code wasn't there doing:
//
//  logrus.WithField("level", 1).Info("hello")
//
// Would just silently drop the user provided level. Instead with this code
// it'll logged as:
//
//  {"level": "info", "fields.level": 1, "msg": "hello", "time": "..."}
//
// It's not exported because it's still using Data in an opionated way. It's to
// avoid code duplication between the two default formatters.
func prefixFieldClashes(entry *logrus.Entry) {
	_, ok := entry.Data["time"]
	if ok {
		entry.Data["fields.time"] = entry.Data["time"]
	}

	entry.Data["time"] = entry.Time.Format(time.RFC3339)

	_, ok = entry.Data["msg"]
	if ok {
		entry.Data["fields.msg"] = entry.Data["msg"]
	}

	entry.Data["msg"] = entry.Message

	_, ok = entry.Data["level"]
	if ok {
		entry.Data["fields.level"] = entry.Data["level"]
	}

	entry.Data["level"] = entry.Level.String()
}
