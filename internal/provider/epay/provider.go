package epay

import (
	"encoding/json"
	"errors"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

func (c *Client) InitTokenRefresher() (err error) {
	if err = c.updateToken(); err != nil {
		return
	}

	ticker := time.Duration(c.credential.ExpiresIn - 60)
	timer := time.NewTicker(ticker * time.Second)

	go func() {
		for {
			<-timer.C

			err = c.updateToken()
			if err != nil {
				return
			}
		}
	}()

	return
}

func (c *Client) PayByTemplate(w http.ResponseWriter, requestSrc Request) (err error) {
	rootDir, err := os.Getwd()
	if err != nil {
		return
	}

	if requestSrc.DueDate != nil && requestSrc.Status.Transaction.StatusName == "" {
		localtime := time.Now().Add(6 * time.Hour)
		dueTime := requestSrc.DueDate.Add(-1500 * time.Second)

		if localtime.Unix() > dueTime.Unix() {
			requestSrc.Status.Transaction.StatusName = "EXPIRED"
			requestSrc.Status.Transaction.StatusTitle = "Истек срок оплаты"

			requestSrc.Status.Transaction.Amount, err = decimal.NewFromString(requestSrc.Amount)
			if err != nil {
				return
			}
		}
	}

	filenames := ""
	switch requestSrc.Status.Transaction.StatusName {
	case "NEW", "AUTH", "EXPIRED":
		requestSrc.Status.Transaction.Status = "pending"
		filenames = filepath.Join(rootDir, "templates", "status.html")
	case "CHARGE":
		requestSrc.Status.Transaction.Status = "success"
		filenames = filepath.Join(rootDir, "templates", "status.html")

		if strings.Contains(requestSrc.Description, "15 ₸") {
			filenames = filepath.Join(rootDir, "templates", "redirect.html")
		}
	case "CANCEL", "CANCEL_OLD", "REFUND":
		requestSrc.Status.Transaction.Status = "cancel"
		filenames = filepath.Join(rootDir, "templates", "status.html")
	case "REJECT", "FAILED", "3D":
		requestSrc.Status.Transaction.Status = "failed"
		filenames = filepath.Join(rootDir, "templates", "status.html")
	case "":
		filenames = filepath.Join(rootDir, "templates", "payment.html")

		token, err := c.getToken(requestSrc)
		if err != nil {
			return err
		}

		requestSrc.Token = token
		requestSrc.PaymentJsLink = c.credential.JS

		requestSrc.BackLink = c.configs.APP.Host + "/invoices/" + requestSrc.ID + "/pay"

	default:
		requestSrc.Status.Transaction.Status = "failed"
		filenames = filepath.Join(rootDir, "templates", "status.html")
	}

	requestSrc.Status.Transaction.Date = requestSrc.Status.Transaction.CreatedDate.Format("02.01.2006 15:04")
	if requestSrc.Status.Transaction.StatusName == "EXPIRED" {
		requestSrc.Status.Transaction.Date = requestSrc.DueDate.Format("02.01.2006 15:04")
	}

	tmpl, err := template.ParseFiles(filenames)
	if err != nil {
		return
	}

	return tmpl.Execute(w, requestSrc)
}

func (c *Client) PayByCard(requestSrc Request) (responseSrc *ResponseCardID, err error) {
	token, err := c.getToken(requestSrc)
	if err != nil {
		return
	}

	req, err := parseToRequestByCardID(requestSrc)
	if err != nil {
		return
	}

	path := c.credential.Endpoint + "/payments/cards/auth"
	resBytes, status, err := c.handler("POST", path, "", token.AccessToken, req, true)
	if err != nil {
		return
	}

	switch status {
	case http.StatusOK:
		err = json.Unmarshal(resBytes, &responseSrc)
	default:
		if ok := json.Unmarshal(resBytes, &responseSrc); ok != nil {
			err = errors.New(string(resBytes))
			return
		}

		err = ErrFromEpay
		return
	}

	return
}

func (c *Client) Charge(transactionID, amount string) (err error) {
	path := c.credential.Endpoint + "/operation/" + transactionID + "/charge"

	resBytes, status, err := c.handler("POST", path, "", "", map[string]string{
		"amount": amount,
	}, true)
	if err != nil {
		return
	}

	if status != http.StatusOK {
		err = errors.New(string(resBytes))
	}

	return
}

func (c *Client) Cancel(transactionID string) (err error) {
	path := c.credential.Endpoint + "/operation/" + transactionID + "/cancel"

	resBytes, status, err := c.handler("POST", path, "", "", nil, true)
	if err != nil {
		return
	}

	if status != http.StatusOK {
		err = errors.New(string(resBytes))
	}

	return
}

func (c *Client) CheckStatus(invoiceID, terminalID string) (res StatusResponse, err error) {
	token, err := c.getToken(Request{TerminalID: terminalID})
	if err != nil {
		return
	}

	path := c.credential.Endpoint + "/check-status/payment/transaction/" + invoiceID
	resBytes, status, err := c.handler("GET", path, "", token.AccessToken, nil, true)
	if err != nil {
		return
	}

	switch status {
	case http.StatusOK, http.StatusBadRequest:
		if err = json.Unmarshal(resBytes, &res); err != nil {
			return
		}

		res.Transaction.StatusDescription = GetDescriptionTitle(res.Transaction.ReasonCode, res.Transaction.Language)
		res.Transaction.StatusTitle = GetStatusTitle(res.Transaction.StatusName, res.Transaction.Language)

		if res.Transaction.StatusName == "AUTH" || res.Transaction.StatusName == "CHARGE" {
			res.Transaction.StatusDescription = ""
		}

	default:
		err = errors.New(string(resBytes))
	}

	return
}
