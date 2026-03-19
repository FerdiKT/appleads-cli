package appleads

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

type Campaign struct {
	ID            int64          `json:"id"`
	Name          string         `json:"name"`
	Status        string         `json:"status"`
	DisplayStatus string         `json:"displayStatus"`
	ServingStatus string         `json:"servingStatus"`
	OrgID         int64          `json:"orgId"`
	AdamID        int64          `json:"adamId"`
	Deleted       bool           `json:"deleted"`
	BudgetAmount  *MoneyAmount   `json:"budgetAmount"`
	DailyBudget   *MoneyAmount   `json:"dailyBudgetAmount"`
	Raw           map[string]any `json:"-"`
}

type MoneyAmount struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
}

type CampaignListResponse struct {
	Data       []Campaign     `json:"data"`
	Pagination map[string]any `json:"pagination"`
	Error      map[string]any `json:"error"`
	Extra      map[string]any `json:"-"`
}

func (c *Client) ListCampaigns(ctx context.Context, offset, limit int) (*CampaignListResponse, error) {
	query := url.Values{}
	query.Set("offset", strconv.Itoa(offset))
	query.Set("limit", strconv.Itoa(limit))

	var resp CampaignListResponse
	if err := c.doJSON(ctx, http.MethodGet, "/campaigns", query, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
