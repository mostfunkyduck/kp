package op_test

import (
	"fmt"
	"testing"
	"github.com/mostfunkyduck/kp/internal/backend/op"
)

var defaultCredentials = op.Credentials {
	Username: "frank",
	Password: "paddyspub",
	Account:  "wolfCola",
}

type TestCommander struct{
	MockOutput		[]byte
	MockError			error
	sessionToken	string
}

func (t TestCommander) Command(cmd string, stdin string) ([]byte, error){
	return t.MockOutput, t.MockError
}

func (t *TestCommander) SetSessionToken(token string) {
	t.sessionToken = token
}

func (t TestCommander) SessionToken() string {
	return t.sessionToken
}

func TestSignIn(t *testing.T) {
	token := "SEKRITS0UCE"
	d, err := op.NewDriver(
		defaultCredentials,
		&TestCommander{
			MockOutput: []byte(`
# BLAH BLAH BLAH BLAH BLAH
export OP_SESSION_` + defaultCredentials.Account + `="` + token + `"
#BLAH BLAH BLAH BLAH BLAH
`),
		},
	)
	err = d.SignIn()
	if err != nil {
		t.Fatal(err)
	}

	if d.Commander.SessionToken() != token {
		t.Fatalf("did retrieve expected token: %s != %s", d.Commander.SessionToken(), token)
	}
}

func TestSignInCommandError(t *testing.T) {
	d, err := op.NewDriver(
		op.Credentials{
			Username: "frank",
			Password: "paddyspub",
			Account:  "wolfCola",
		},
		&TestCommander{
			MockError: fmt.Errorf("BLAH BLAH BLAH BLAH BLAH"),
		},
	)
	err = d.SignIn()
	if err == nil {
		t.Fatal("error from the commander did not get returned by SignIn()")
	}
}

func TestSignInParseError(t *testing.T) {
	d, err := op.NewDriver(
		defaultCredentials,
		&TestCommander{
			MockOutput: []byte("BLAH BLAH BLAH BLAH BLAH"),
		},
	)
	err = d.SignIn()
	if err == nil {
		t.Fatal("parse error from the commander did not get handled properly by SignIn()")
	}
}

func TestListVaults(t *testing.T) {
	d, err := op.NewDriver(
		defaultCredentials,
		&TestCommander{
			// OH JOY MORE JSON
			MockOutput: []byte(`[
{
"uuid": "uuid0",
"name": "name0"
},
{
"uuid": "uuid1",
"name": "name1",
"junk": "junk"
}
]`),
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	groups, err := d.ListVaults()
	if err != nil {
		t.Fatal(err)
	}
	if len(groups) != 2 {
		t.Fatalf("json was not translated properly to groups: %d != %d. [%v]", len(groups), 2, groups)
	}

	// array order is probably deterministic, based on light googling (https://github.com/golang/go/issues/27179#issuecomment-740859594)
	for i, g := range groups {
		if g.UUID != fmt.Sprintf("uuid%d",i) {
			t.Fatalf("group %d had incorrect uuid: %s", i, g.UUID)
		}

		if g.Title != fmt.Sprintf("name%d",i) {
			t.Fatalf("group %d had incorrect name: %s", i, g.Title)
		}
	}
}

func TestListVaultsEmpty(t *testing.T) {
	d, err := op.NewDriver(
		defaultCredentials,
		&TestCommander{
			MockOutput: []byte("[]"),
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	groups, err := d.ListVaults()
	if err != nil {
		t.Fatal(err)
	}
	if len(groups) != 0 {
		t.Fatalf("an empty array caused the driver to return a populated one, wtf? %v", groups)
	}
}

func TestListVaultsGarbage(t *testing.T) {
	d, err := op.NewDriver(
		defaultCredentials,
		&TestCommander{
			MockOutput: []byte("RANDOM CHARACTERS ARE BEST CHARACTERS"),
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	groups, err := d.ListVaults()
	if err != nil {
		t.Fatalf("garbage data should trigger an empty response, got an error instead: %s", err)
	}
	if len(groups) != 0 {

	}
}

func TestListItems(t *testing.T) {
	d, err := op.NewDriver(
		defaultCredentials,
		&TestCommander{
			// OH JOY MORE JSON
			MockOutput: []byte(`[
	{
		"uuid": "uuid0",
		"templateUuid": "005",
		"trashed": "N",
		"createdAt": "2018-12-07T19:55:21Z",
		"updatedAt": "2018-12-07T19:55:21Z",
		"overview": {
			"ainfo": "ainfo0",
			"title": "title0"
		}
	},
	{
		"uuid": "uuid1",
		"templateUuid": "005",
		"trashed": "N",
		"createdAt": "2018-12-07T19:55:21Z",
		"updatedAt": "2018-12-07T19:55:21Z",
		"overview": {
			"ainfo": "ainfo1",
			"title": "title1"
		}
	}
]`),
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	entries, err := d.ListItems("vaulty the vault")
	if err != nil {
		t.Fatal(err)
	}

	for i, e := range entries {
		if e.UUID != fmt.Sprintf("uuid%d",i) {
			t.Fatalf("entry %d had incorrect uuid: %s", i, e.UUID)
		}

		if e.Title() != fmt.Sprintf("title%d",i) {
			t.Fatalf("entry %d had incorrect name: %s", i, e.Title())
		}
	}
}

func TestListItemsNoOverview(t *testing.T) {
	d, err := op.NewDriver(
		defaultCredentials,
		&TestCommander{
			// OH JOY MORE JSON
			MockOutput: []byte(`[
	{
		"uuid": "uuid1",
		"templateUuid": "005",
		"trashed": "N",
		"createdAt": "2018-12-07T19:55:21Z",
		"updatedAt": "2018-12-07T19:55:21Z",
	}
]`),
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	entries, err := d.ListItems("vaulty the vault")
	if err != nil {
		t.Fatal(err)
	}

	for _, e := range entries {
		if e.Title() != "" {
			t.Fatalf("entry with no source title... has a title!: %v", e)
		}
	}
}

/** not running create tests until create is worked out
func TestCreateItem(t *testing.T) {
	d, err := op.NewDriver(
		defaultCredentials,
		&TestCommander{
			MockOutput: []byte(`[]`),
		},
	)

	if err != nil {
		t.Fatal(err)
	}

	entry := op.Entry{
		title: "TITLE",
		username: "USNAME",
		notes: "MOTES",
		password: "SEKRIT",
		opDriver: &d,
	}
	if err := d.CreateItem("vaulty the vault", entry); err != nil {
		t.Fatal(err)
	}
}

func TestCreateItemIncompleteEntry(t *testing.T) {
	d, err := op.NewDriver(
		defaultCredentials,
		&TestCommander{
			MockOutput: []byte(`[]`),
		},
	)

	if err != nil {
		t.Fatal(err)
	}

	entry := op.Entry{}
	if err := d.CreateItem("vaulty the vault", entry); err == nil {
		t.Fatal("creating an entry without a title worked, that's dumb")
	}
}
**/
