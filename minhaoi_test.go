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
	flow := NewFlow(false)
	flow.User = user
	invoices, err := flow.InvoiceFlow()
	if err != nil {
		t.Log(err)
	}
	if len(invoices) == 0 {
		t.Fatalf("not invoices returned")
	}
	for i := range invoices {
		if len(invoices[i].BarCode) == 0 {
			t.Fatalf("invalid barcode; cannot be empty")
		}
		if len(invoices[i].DueDate) == 0 {
			t.Fatalf("invalid DueDate; cannot be empty")
		}
		if len(invoices[i].Value) == 0 {
			t.Fatalf("invalid Value; cannot be empty")
		}
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
