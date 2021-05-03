package op

// Attempts to hack around with the 1password cli tool and make it a backend
// This API is completely undocumented, including the responses, all bets are off
import (
	//c "github.com/mostfunkyduck/kp/internal/backend/common"
	// t "github.com/mostfunkyduck/kp/internal/backend/types"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"regexp"
	"text/template"
	"time"
)
// JSON structs
type opVault struct {
	UUID string `json:"uuid"`
	Name string `json:"name"`
}

type opGetItem struct {
	Details			opGetItemDetails	`json:"details"`
	Overview		opGetItemOverview	`json:"overview"`
}

type opGetItemOverview struct {
	Ainfo	string	`json:"ainfo"`
	Title	string	`json:"title"`
}

type opGetItemDetails struct {
	Notes				string						`json:"notesPlain"`
	Fields			[]opGetItemField	`json:"fields"`
}

type opGetItemField struct {
	Designation	string	`json:"designation"`
	Name				string	`json:"name"`
	Value				string	`json:"value"`
}

type opListItem struct {
	UUID				string							`json:"uuid"`
	Trashed			string							`json:"trashed"`
	CreatedAt		time.Time						`json:"createdAt"`
	UpdatedAt		time.Time						`json:"UpdatedAt"`
	Overview		opListItemOverview	`json:"overview"`
}

type opListItemOverview struct {
	Ainfo					string		`json:"ainfo"`
	Title					string		`json:"title"`
	URL						string		`json:"url"`
}

type Credentials struct {
	Username			string
	Password			string
	Account				string
}

func (c Credentials) Verify() bool {
	return (c.Username != "" && c.Password != "" && c.Account != "")
}
// Translates `op` calls into useable types
type Driver struct{
	Commander Commander
	Credentials Credentials
}

// NewDriver creates an op driver based on provided credentials and commander
// The commander will make the actual calls, following the Commander interface.
func NewDriver(creds Credentials, c Commander) (Driver, error) {
	if !creds.Verify() {
		return Driver{}, fmt.Errorf("invalid credentials")
	}
	return Driver{
		Commander: c,
		Credentials: creds,
	}, nil
}
// SignIn Signs in to 1password using the credentials provided in NewDriver
// It retrieves and sets the session token
func (d *Driver) SignIn() error {
	// this is a quasi-hack to get go to easily do piped commands, hooray
	output, err := d.Commander.Command("op signin " + d.Credentials.Account, d.Credentials.Password)
	if err != nil {
		return fmt.Errorf("error signing in to 1password: output: %s, err: %s", output, err)
	}
	re := regexp.MustCompile(`export OP_SESSION_` + d.Credentials.Account + `="(.+)"`)
	// this should absolutely never return more than one token
	tokens := re.FindStringSubmatch(string(output))
	if len(tokens) <= 1 {
		// the first retval is just the string, submatches are subsequent and optional
		return fmt.Errorf("could not find token string in response from 1password. received '%s'", string(output))
	}
	d.Commander.SetSessionToken(tokens[1])
	return nil
}

// ListVaults will list all the vaults for this account and translate them into group objects
func (d *Driver) ListVaults() ([]Group, error) {
	output, err := d.Commander.Command("op list vaults", "")
	if err != nil {
		return []Group{}, fmt.Errorf("could not retrieve vaults: output:'%s' error: '%s'", output, err)
	}
	// OH BABY IT'S TIME TO USE JSON IN GO!!!!
	fmt.Println(string(output))
	var vaults []opVault
	json.Unmarshal(output, &vaults)
	groups := []Group{}
	for _, v := range vaults {
		groups = append(groups, Group{
			UUID:	v.UUID,
			Title: v.Name,
		})
	}
	return groups, nil
}

// ListItems will return the items in a provided vault as entry objects
// Note that op differentiates between "listing" - which only returns metadata and "getting"
// which returns full information
func (d *Driver) ListItems(vault string) ([]*Entry, error) {
	output, err := d.Commander.Command(fmt.Sprintf("op list items --vault=%s", vault), "")
	if err != nil {
		return []*Entry{}, fmt.Errorf("could not retrieve items from vault '%s': %s", vault, err)
	}

	var items []opListItem
	json.Unmarshal(output, &items)
	entries := []*Entry{}
	for _, i := range items {
		newguy := &Entry{
			UUID: i.UUID,
			listItem: i,
			opDriver: d,
		}
		newguy.SetDriver(newguy)
		entries = append(entries, newguy)
	}
	return entries, nil
}

// create item --vault=Whatever
// CreateItem creates a 'login' item - the only supported type thus far - based on a provided entry
// In order for the call to be valid, the Entry must minimally have a Title
func (d *Driver) CreateItem(vault string, entry Entry) error {
	if entry.Title() == "" {
		return fmt.Errorf("entry must provide title in order to be created")
	}

	type templateFodder struct {
		Notes			string
		Username	string
		Password	string
	}

	t := templateFodder {
		Notes: string(entry.Get("notes").Value()),
		Username: string(entry.Get("username").Value()),
		Password: string(entry.Get("password").Value()),
	}

	loginTemplate := `{ "notesPlain": "{{.Notes}}", "sections": [], "passwordHistory": [], "fields": [ { "value": "{{.Username}}", "name": "username", "type": "T", "designation": "username" }, { "value": "{{.Password}}", "name": "password", "type": "P", "designation": "password" } ] }`
	template, err := template.New("Item").Parse(loginTemplate)
	if err != nil {
		return fmt.Errorf("could not template new 'login' item, %s", err)
	}

	var jsonBuffer bytes.Buffer
	err = template.Execute(&jsonBuffer, t)
	if err != nil {
		return fmt.Errorf("could not build op 'login' iten from entry: %s", err)
	}
	encodedBuffer := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString([]byte(loginTemplate))

	// Assumption: the output can be safely discarded because:
	// 	1. it only contains details that the caller should be getting fresh from 1pass before using (uuid, created times)
	//	   the entry is being passed by *value* - it is useless after this call, go get another one if you want that info, dammit!
	//  2. there is no useful information in there from the program's point of view either
	//  3. op appears to handle error cases properly and set non-zero exit codes, so the Commander can do the needful wrt error cases
	_, err = d.Commander.Command(fmt.Sprintf(`op create item --vault=%s --title=%s Login "%s"`,
		vault, entry.Title(), encodedBuffer), "")
	if err != nil {
		return fmt.Errorf("could not create item: %s", err)
	}
	return nil
}

// get single item as entry
func (d Driver) GetItem(uuid string) (*Entry, error) {
	output, err := d.Commander.Command(fmt.Sprintf("op get item %s", uuid), "")
	if err != nil {
		return &Entry{}, fmt.Errorf("could not retrieve item '%s': %s", uuid, err)
	}

	var item opGetItem
	json.Unmarshal(output, &item)

	newguy := &Entry{
		getItem: item,
		opDriver:	&d,
	}
	if err := newguy.Init(); err != nil {
		return newguy, fmt.Errorf("could not initialize entry, %s", err)
	}
	return newguy, nil
}
// delete item --vault=Whatever
