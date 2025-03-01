package currency

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

type Response struct {
	XMLName     xml.Name `xml:"rates"`
	Text        string   `xml:",chardata"`
	Generator   string   `xml:"generator"`
	Title       string   `xml:"title"`
	Link        string   `xml:"link"`
	Description string   `xml:"description"`
	Copyright   string   `xml:"copyright"`
	Date        string   `xml:"date"`
	Rates       []Rate   `xml:"item"`
}

type Rate struct {
	Text     string          `xml:",chardata" json:"-"`
	Title    string          `xml:"title" json:"id"`
	Fullname string          `xml:"fullname" json:"name"`
	Rate     decimal.Decimal `xml:"description" json:"value"`
	Quant    string          `xml:"quant" json:"quant"`
	Index    string          `xml:"index" json:"index"`
	Change   string          `xml:"change" json:"change"`
}

func (c *Client) GetRateByID(ctx context.Context, id string, datetime time.Time) (dest Rate, err error) {
	if id == "" {
		return dest, errors.New("id: cannot be blank")
	}

	rates, err := c.GetRatesByDate(ctx, datetime)
	if err != nil {
		return
	}

	isNotFound := true
	for i := 0; i < len(rates); i++ {
		if strings.EqualFold(id, rates[i].Title) {
			if rates[i].Rate.IsZero() {
				return dest, errors.New("rate: cannot be blank")
			}
			dest = rates[i]

			isNotFound = false

			break
		}
	}

	if isNotFound {
		return dest, fmt.Errorf("id: %s is not found", id)
	}

	return
}

func (c *Client) GetRatesByDate(ctx context.Context, datetime time.Time) ([]Rate, error) {
	if datetime.IsZero() {
		return nil, errors.New("datetime: cannot be blank")
	}

	rates, err := c.getRatesByDate(ctx, datetime)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch rates: %w", err)
	}

	return rates, nil
}

func (c *Client) getRatesByDate(ctx context.Context, datetime time.Time) (dest []Rate, err error) {
	path, err := url.Parse(c.Credentials.URL)
	if err != nil {
		return
	}
	path = path.JoinPath("/rss/get_rates.cfm")

	params := url.Values{
		"fdate": []string{datetime.Format("02.01.2006")},
	}
	path.RawQuery = params.Encode()

	currency := Response{}
	if err = c.request(ctx, "GET", path.String(), &currency); err != nil {
		return
	}
	dest = currency.Rates

	return
}
