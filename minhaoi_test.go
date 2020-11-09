package minhaoi

import (
	"encoding/json"
	"io/ioutil"
	"testing"
)

func TestFlow(t *testing.T) {
	user, err := readFile("config_test.json")
	if err != nil {
		t.Log(err)
	}
	t.Logf("%#v", user)
	flow := NewFlow(true)
	flow.User = user
	_, err = flow.InvoiceFlow()
	if err != nil {
		t.Log(err)
	}
}

func readFile(filename string) (UserData, error) {
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		return UserData{}, err
	}
	output := UserData{}
	err = json.Unmarshal(file, &output)
	if err != nil {
		return UserData{}, err
	}

	return output, nil
}
