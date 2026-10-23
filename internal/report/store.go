// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

package report

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Keep is how many finished reports are kept in the folder; the oldest are removed beyond that.
const Keep = 30

// Store keeps the reports in a folder, one JSON file for each run.
type Store struct{ Dir string }

var unsafe = regexp.MustCompile(`[^A-Za-z0-9_-]+`)

// Name is the file name (without the folder) a report is saved under: when it started, then the show.
func Name(r Report) string {
	return r.Started.Local().Format("20060102-150405.000000000") + "-" + safeShowID(r.ShowID) + ".json"
}

// DownloadName is readable in a browser's downloads list; the stored JSON name has finer precision
// because two runs of the same show can start within one second.
func DownloadName(r Report) string {
	return "audience-report-" + r.Started.Local().Format("20060102-150405") + "-" + safeShowID(r.ShowID) + ".csv"
}

func safeShowID(id string) string { return strings.Trim(unsafe.ReplaceAllString(id, "-"), "-") }

// Save writes the report (replacing one with the same name, so a run saved while going on is replaced by the final
// one), and removes the oldest beyond Keep.
func (s Store) Save(r Report) error {
	if s.Dir == "" {
		return errors.New("no folder for the reports")
	}
	if err := os.MkdirAll(s.Dir, 0o700); err != nil {
		return err
	}
	b, err := json.MarshalIndent(r, "", " ")
	if err != nil {
		return err
	}
	f, err := os.CreateTemp(s.Dir, ".report-*")
	if err != nil {
		return err
	}
	tmp := f.Name()
	if _, err = f.Write(b); err == nil {
		err = f.Close()
	} else {
		f.Close()
	}
	if err == nil {
		err = os.Chmod(tmp, 0o600)
	}
	if err == nil {
		err = os.Rename(tmp, filepath.Join(s.Dir, Name(r)))
	}
	if err != nil {
		os.Remove(tmp)
		return err
	}
	s.prune()
	return nil
}

func (s Store) prune() {
	names := s.Names()
	for i := Keep; i < len(names); i++ {
		os.Remove(filepath.Join(s.Dir, names[i]))
	}
}

// Names lists the saved reports, the newest first.
func (s Store) Names() []string {
	ents, _ := os.ReadDir(s.Dir)
	var out []string
	for _, e := range ents {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") && !strings.HasPrefix(e.Name(), ".") {
			out = append(out, e.Name())
		}
	}
	sort.Sort(sort.Reverse(sort.StringSlice(out)))
	return out
}

// Load reads one saved report by its name.
func (s Store) Load(name string) (Report, error) {
	if name != filepath.Base(name) || !strings.HasSuffix(name, ".json") || strings.HasPrefix(name, ".") {
		return Report{}, errors.New("no such report")
	}
	b, err := os.ReadFile(filepath.Join(s.Dir, name))
	if err != nil {
		return Report{}, errors.New("no such report")
	}
	var r Report
	if err := json.Unmarshal(b, &r); err != nil {
		return Report{}, errors.New("the report cannot be read")
	}
	return r, nil
}

// Delete removes one saved report by its name.
func (s Store) Delete(name string) error {
	if name != filepath.Base(name) || !strings.HasSuffix(name, ".json") || strings.HasPrefix(name, ".") {
		return errors.New("no such report")
	}
	if err := os.Remove(filepath.Join(s.Dir, name)); err != nil {
		return errors.New("no such report")
	}
	return nil
}

// Latest is the newest saved report, or false when there is none.
func (s Store) Latest() (Report, string, bool) {
	for _, n := range s.Names() {
		if r, err := s.Load(n); err == nil {
			return r, n, true
		}
	}
	return Report{}, "", false
}

// Duration formats a length of time as h:mm:ss.
func Duration(d Span) string {
	s := int(time.Duration(d).Round(time.Second).Seconds())
	return fmt.Sprintf("%d:%02d:%02d", s/3600, s/60%60, s%60)
}

// WriteCSV writes one row for each device, for a spreadsheet. Times are in the computer's own time.
func WriteCSV(w io.Writer, r Report) error {
	cw := csv.NewWriter(w)
	cw.Write([]string{"Device", "Address", "Type", "Screen", "Languages", "First seen", "Last seen", "Connections", "Time connected", "Longest stay", "Connected at the end"})
	at := func(t time.Time) string { return t.Local().Format("2006-01-02 15:04:05") }
	for _, d := range r.Devices {
		screen := "device"
		if d.Projector {
			screen = "projector"
		}
		now := "no"
		if d.Now {
			now = "yes"
		}
		cw.Write([]string{d.ID, d.Address, d.Kind, screen, strings.Join(d.Languages, " "), at(d.First), at(d.Last), strconv.Itoa(d.Connections), Duration(d.Connected), Duration(d.Longest), now})
	}
	cw.Flush()
	return cw.Error()
}
