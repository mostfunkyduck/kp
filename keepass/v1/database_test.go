package keepassv1_test

import (
	v1 "github.com/mostfunkyduck/kp/keepass/v1"
	"testing"
)

func TestSavePath(t *testing.T) {
	sp := "adsfasdfjkalskdfj"
	db := &v1.Database{}

	db.SetSavePath(sp)
	dbSp := db.SavePath()
	if sp != dbSp {
		t.Errorf("%s != %s", sp, dbSp)
	}
}
