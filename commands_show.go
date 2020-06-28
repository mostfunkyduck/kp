package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/abiosoft/ishell"
	"zombiezen.com/go/sandpass/pkg/keepass"
)

func formatTime(t time.Time) (formatted string) {
	timeFormat := "Mon Jan 2 15:04:05 MST 2006"
	if (t == time.Time{}) {
		formatted = "unknown"
	} else {
		since := time.Since(t).Round(time.Duration(1) * time.Second)
		sinceString := since.String()

		// greater than or equal to 1 day
		if since.Hours() >= 24 {
			sinceString = fmt.Sprintf("%d days ago", int(since.Hours()/24))
		}

		// greater than or equal to ~1 month
		if since.Hours() >= 720 {
			// rough estimate, not accounting for non-30-day months
			months := int(since.Hours() / 720)
			sinceString = fmt.Sprintf("about %d months ago", months)
		}

		// greater or equal to 1 year
		if since.Hours() >= 8760 {
			// yes yes yes, leap years aren't 365 days long
			years := int(since.Hours() / 8760)
			sinceString = fmt.Sprintf("about %d years ago", years)
		}

		// less than a second
		if since.Seconds() < 1.0 {
			sinceString = "less than a second ago"
		}

		formatted = fmt.Sprintf("%s (%s)", t.Local().Format(timeFormat), sinceString)
	}
	return
}

func outputEntry(e keepass.Entry, s *ishell.Shell, path string, full bool) {
	s.Printf("UUID: %s\n", e.UUID)

	s.Printf("Creation Time: %s\n", formatTime(e.CreationTime))
	s.Printf("Last Modified: %s\n", formatTime(e.LastModificationTime))
	s.Printf("Last Accessed: %s\n", formatTime(e.LastAccessTime))
	s.Printf("Location: %s\n", path)
	s.Printf("Title: %s\n", e.Title)
	s.Printf("URL: %s\n", e.URL)
	s.Printf("Username: %s\n", e.Username)
	password := "[redacted]"
	if full {
		password = e.Password
	}
	s.Printf("Password: %s\n", password)
	s.Printf("Notes: %s\n", e.Notes)
	if e.HasAttachment() {
		s.Printf("Attachment: %s\n", e.Attachment.Name)
	}

}

func Show(shell *ishell.Shell) (f func(c *ishell.Context)) {
	return func(c *ishell.Context) {
		if len(c.Args) < 1 {
			shell.Println("syntax: " + c.Cmd.Help)
			return
		}

		fullMode := false
		path := c.Args[0]
		for _, arg := range c.Args {
			if strings.HasPrefix(arg, "-") {
				if arg == "-f" {
					fullMode = true
				}
				continue
			}
			path = arg
		}

		currentLocation := c.Get("currentLocation").(*keepass.Group)
		location, err := traversePath(currentLocation, path)
		if err != nil {
			shell.Println(fmt.Sprintf("could not find entry named [%s]", path))
			return
		}

		// get the base name of the entry so that we can compare it to the actual
		// entries in this group
		entryNameBits := strings.Split(path, "/")
		entryName := entryNameBits[len(entryNameBits)-1]
		if *debugMode {
			shell.Printf("looking for entry [%s]", entryName)
		}
		for i, entry := range location.Entries() {
			if *debugMode {
				shell.Printf("looking at entry/idx for entry %s/%d\n", entry.Title, i, entryName)
			}
			if intVersion, err := strconv.Atoi(entryName); err == nil && intVersion == i ||
				entryName == entry.Title ||
				entryName == entry.UUID.String() {
				outputEntry(*entry, shell, path, fullMode)
				return
			}
		}
	}
}
