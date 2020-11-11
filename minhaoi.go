package minhaoi

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/knq/chromedp/kb"
	"github.com/ozzono/normalize"
)

var (
	configPath string
	user       UserData
)

// Flow contains and the data and methods needed to crawl through the enel webpage
type Flow struct {
	c        context.Context
	User     UserData
	Invoices []Invoice
	cancel   []context.CancelFunc
}

//Invoice has all the invoice data needed for payment
type Invoice struct {
	DueDate string
	Value   string
	BarCode string
	Status  string
}

//UserData has all the needed data to login
type UserData struct {
	Login string `json:"login"`
	Pw    string `json:"pw"`
	Name  string `json:"name"`
}

//InvoiceFlow crawls through the enel page
func (flow *Flow) InvoiceFlow() ([]Invoice, error) {
	for i := range flow.cancel {
		defer flow.cancel[i]()
	}

	err := flow.login()
	if err != nil {
		log.Println(err)
		return []Invoice{}, err
	}

	err = flow.invoiceList()
	if err != nil {
		return []Invoice{}, err
	}
	return flow.Invoices, nil
}

func (flow *Flow) login() error {
	if err := flow.checkUserData(); err != nil {
		return err
	}
	log.Println("starting login flow")
	name := ""
	err := chromedp.Run(flow.c,
		chromedp.Navigate(`https://minha.oi.com.br/minhaoi/home/`),
		chromedp.Sleep(2*time.Second),
		chromedp.WaitVisible(`h3.instrucao-login`),
		chromedp.Click(`#usernameinput`, chromedp.NodeVisible, chromedp.ByID),
		chromedp.SendKeys("#usernameinput", kb.End+flow.User.Login, chromedp.ByID),
		chromedp.Click(`#passwordinput`, chromedp.NodeVisible, chromedp.ByID),
		chromedp.SendKeys("#passwordinput", kb.End+flow.User.Pw, chromedp.ByID),
		chromedp.Click(`#botaoLogin`, chromedp.NodeVisible, chromedp.ByID),
		chromedp.WaitVisible(`.widget-conta`),
		chromedp.Text(
			`document.querySelector("#application > div > div.Container__ContainerStyle-sc-1iqy2ia-0.fkNjxh.coi-header > header > div > button > p")`,
			&name,
			chromedp.ByJSPath,
		),
	)
	if err != nil {
		return err
	}

	if !strings.Contains(strings.ToLower(normalize.Norm(name)), strings.ToLower(normalize.Norm(flow.User.Name))) {
		return fmt.Errorf("failed to login; user name do not match")
	}
	log.Println("successfully logged in")
	return nil
}

func (flow *Flow) invoiceList() error {
	log.Println("starting invoiceList flow")
	nodeCount := 0
	err := chromedp.Run(flow.c,
		chromedp.Evaluate(jsClassNodeCount("bofSkK"), &nodeCount),
	)
	if err != nil {
		return fmt.Errorf("invoiceFlow err: %v", err)
	}

	for i := 0; i < nodeCount; i++ {
		flow.invoiceData(refList(i + 1))
	}

	log.Printf("%#v", flow.Invoices)

	log.Println("Successfully selected the last listed invoice")
	return nil
}

func (flow *Flow) invoiceData(selectors map[string]string) error {
	log.Println("fetching invoice data")
	invoice := Invoice{}
	err := chromedp.Run(flow.c,
		chromedp.Sleep(5*time.Second),
		chromedp.Text(
			selectors["value"],
			&invoice.Value,
			chromedp.ByQuery,
		),
		chromedp.Text(
			selectors["due-date"],
			&invoice.DueDate,
			chromedp.ByQuery,
		),
		chromedp.Click(selectors["open-barcode"], chromedp.NodeVisible, chromedp.ByJSPath),
		chromedp.Sleep(1*time.Second),
		chromedp.Text(
			selectors["barcode"],
			&invoice.BarCode,
			chromedp.ByQuery,
		),
	)
	if err != nil {
		return fmt.Errorf("chromedp.Run err: %v", err)
	}
	invoice.Value = strings.TrimPrefix(invoice.Value, "R$ \\u00a0")
	invoice.BarCode = strings.Replace(invoice.BarCode, " ", "", -1)
	invoice.BarCode = strings.Replace(invoice.BarCode, "-", "", -1)
	invoice.Status = "pending"

	flow.Invoices = append(flow.Invoices, invoice)
	log.Println("Successfully fetched invoice data")
	return nil
}

//NewFlow creates a flow with context besides user and invoice data
func NewFlow(headless bool) Flow {
	ctx, cancel := setContext(headless)
	return Flow{c: ctx, cancel: cancel}
}

func setContext(headless bool) (context.Context, []context.CancelFunc) {
	outputFunc := []context.CancelFunc{}
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.DisableGPU,
		// Set the headless flag to false to display the browser window
		chromedp.Flag("headless", headless),
	)
	ctx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	outputFunc = append(outputFunc, cancel)
	ctx, cancel = chromedp.NewContext(ctx)
	outputFunc = append(outputFunc, cancel)
	return ctx, outputFunc
}

func (flow *Flow) textByPath(path string) (string, error) {
	output := ""
	err := chromedp.Run(flow.c,
		chromedp.Text(
			path,
			&output,
			chromedp.ByJSPath,
		),
	)
	if err != nil {
		return "", fmt.Errorf("flow.textByPath err: %v", err)
	}
	return output, nil
}

func (flow *Flow) textByID(id string) (string, error) {
	output := ""
	err := chromedp.Run(flow.c,
		chromedp.Text(
			id,
			&output,
			chromedp.ByID,
		),
	)
	if err != nil {
		return "", fmt.Errorf("flow.textByID err: %v", err)
	}
	return output, nil
}

func (flow *Flow) waitVisible(something string) error {
	log.Printf("waiting for %v", something)
	return chromedp.Run(flow.c,
		chromedp.WaitVisible(something),
	)
}

func (flow *Flow) checkUserData() error {
	if len(flow.User.Login) == 0 {
		return fmt.Errorf("invalid login; cannot be empty")
	}
	if len(flow.User.Name) == 0 {
		return fmt.Errorf("invalid name; cannot be empty")
	}
	if len(flow.User.Pw) == 0 {
		return fmt.Errorf("invalid pw; cannot be empty")
	}
	return nil
}

func jsClassNodeCount(class string) string {
	return fmt.Sprintf(
		`function nodeCount(c){
			return document.getElementsByClassName(c).length/2
		}
		nodeCount("%s");`,
		class,
	)
}

func refList(row int) map[string]string {
	return map[string]string{
		"value": fmt.Sprintf(
			`#application > div > div.Container__ContainerStyle-sc-1iqy2ia-0.fPvmFf > div > div > div.Container__ContainerStyle-sc-1iqy2ia-0.kyItLy > div:nth-child(2) > div > div > div > div:nth-child(3) > div:nth-child(%d) > div > div.Container__ContainerStyle-sc-1iqy2ia-0.dUMzeo > p`,
			row,
		),
		"due-date": fmt.Sprintf(
			`#application > div > div.Container__ContainerStyle-sc-1iqy2ia-0.fPvmFf > div > div > div.Container__ContainerStyle-sc-1iqy2ia-0.kyItLy > div:nth-child(2) > div > div > div > div:nth-child(3) > div:nth-child(%d) > div > div.Container__ContainerStyle-sc-1iqy2ia-0.neYx > div.Container__ContainerStyle-sc-1iqy2ia-0.hjqMVf > div > p`,
			row,
		),
		"barcode": fmt.Sprintf(
			`#application > div > div.Container__ContainerStyle-sc-1iqy2ia-0.fPvmFf > div > div > div.Container__ContainerStyle-sc-1iqy2ia-0.kyItLy > div:nth-child(2) > div > div > div > div:nth-child(3) > div:nth-child(%d) > div > div.Collapsible__CollapsibleContainer-sc-12167wd-0.kaHzyu > div > div > div.Container__ContainerStyle-sc-1iqy2ia-0.fsKtkE > p`,
			row,
		),
		"open-barcode": fmt.Sprintf(
			`document.querySelector("#application > div > div.Container__ContainerStyle-sc-1iqy2ia-0.fPvmFf > div > div > div.Container__ContainerStyle-sc-1iqy2ia-0.kyItLy > div:nth-child(2) > div > div > div > div:nth-child(3) > div:nth-child(%d) > div > div.Container__ContainerStyle-sc-1iqy2ia-0.neYx > div.Container__ContainerStyle-sc-1iqy2ia-0.cRgMTL.desalinhado > div > button")`,
			row,
		),
	}
}
