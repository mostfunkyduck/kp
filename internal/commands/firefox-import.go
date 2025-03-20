package commands

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/mostfunkyduck/ishell"
	c "github.com/mostfunkyduck/kp/internal/backend/common"
	t "github.com/mostfunkyduck/kp/internal/backend/types"
)

const (
	URL                   = "url"
	USERNAME              = "username"
	PASSWORD              = "password"
	HTTP_REALM            = "httpRealm"
	FORM_ACTION_ORIGIN    = "formActionOrigin"
	GUID                  = "guid"
	TIME_CREATED          = "timeCreated"
	TIME_LAST_USED        = "timeLastUsed"
	TIME_PASSWORD_CHANGED = "timePasswordChanged"
)

var allFields = []string{
	URL, USERNAME, PASSWORD, HTTP_REALM, FORM_ACTION_ORIGIN,
	GUID, TIME_CREATED, TIME_LAST_USED, TIME_PASSWORD_CHANGED,
}

func parseTimestamp(input string) (time.Time, error) {
	i, err := strconv.ParseInt(input[:10], 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(i, 0), nil
}

func parseCSV(shell *ishell.Shell, path string, location t.Group) (int, int, error) {
	updated := 0
	broken := 0
	f, err := os.Open(path)
	if err != nil {
		return updated, broken, fmt.Errorf("error opening %s", err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	records, err := r.ReadAll()
	if err != nil {
		return updated, broken, fmt.Errorf("error reading CSV: %s", err)
	}
	// cache the header indices
	headers := make(map[string]int)
	for idx, val := range records[0] {
		headers[val] = idx
	}

	// verifying that the fields i expect to be there are there
	for _, v := range allFields {
		if _, exists := headers[v]; !exists {
			return updated, broken, fmt.Errorf("field '%s' not found in header row", v)
		}
	}

	for _, record := range records[1:] {
		username := record[headers[USERNAME]]
		password := record[headers[PASSWORD]]
		timeCreated := record[headers[TIME_CREATED]]
		timeLastUsed := record[headers[TIME_LAST_USED]]
		timePasswordChanged := record[headers[TIME_PASSWORD_CHANGED]]
		url := record[headers[URL]]
		title := fmt.Sprintf("%s (%s)", strings.TrimPrefix(url, "https://"), username)
		entry, err := location.NewEntry(title)
		if err != nil {
			shell.Println(fmt.Sprintf("error creating entry for '%s': %s", title, err))
			broken++
			continue
		}

		updated++
		entry.SetUsername(username)
		entry.SetPassword(password)
		if uTime, err := parseTimestamp(timeCreated); err == nil {
			entry.SetCreationTime(uTime)
		} else {
			shell.Println(err)
		}
		if uTime, err := parseTimestamp(timeLastUsed); err == nil {
			entry.SetLastAccessTime(uTime)
		} else {
			shell.Println(err)
		}

		if uTime, err := parseTimestamp(timePasswordChanged); err == nil {
			entry.SetLastModificationTime(uTime)
		} else {
			shell.Println(err)
		}

		// FIXME: this is assuming kpv1 format, not a huge deal rn, but not what it should be
		entry.Set(c.NewValue(
			[]byte(url),
			"URL",
			true,
			false,
			false,
			t.STRING,
		))
		record[headers[PASSWORD]] = "REDACTED"
		entry.Set(c.NewValue(
			fmt.Appendf(nil, "original record: %v\n", record),
			"notes",
			true,
			false,
			false,
			t.LONGSTRING,
		))
	}
	return updated, broken, nil
}

func FirefoxImport(shell *ishell.Shell) (f func(c *ishell.Context)) {
	return func(c *ishell.Context) {
		errString, ok := syntaxCheck(c, 2)
		if !ok {
			shell.Println(errString)
			return
		}

		csvPath := c.Args[0]
		dbPath := c.Args[1]
		db := shell.Get("db").(t.Database)

		pathBits := strings.Split(dbPath, "/")
		parentPath := strings.Join(pathBits[0:len(pathBits)-1], "/")
		location, entry, err := TraversePath(db, db.CurrentLocation(), parentPath)
		if err != nil {
			shell.Println("invalid path: " + err.Error())
			return
		}

		if location == nil {
			shell.Println("location does not exist: " + dbPath)
			return
		}

		if entry != nil {
			shell.Println("path points to entry: %s" + dbPath)
			return
		}

		if location.IsRoot() {
			shell.Println("cannot import entries to root node")
			return
		}

		if len(location.Entries()) != 0 {
			shell.Printf("'%s' contains entries, this could cause conflicts, continue? (y/n)\n", dbPath)
			input := shell.ReadLine()
			if input != "y" {
				return
			}
		}

		updated, broken, err := parseCSV(shell, csvPath, location)
		if err != nil {
			shell.Printf("error importing '%s': %s\n", csvPath, err)
			if updated == 0 {
				return
			}
		}
		shell.Printf("%d entries were imported, %d entries were skipped\n", updated, broken)

		if err := PromptAndSave(shell); err != nil {
			shell.Printf("could not save database: %s\n", err)
		}
	}
}
