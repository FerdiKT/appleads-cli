package appleads

import (
	"context"
	"net/http"
	"net/url"
)

type UserACL struct {
	OrgName      string   `json:"orgName"`
	OrgID        int64    `json:"orgId"`
	ParentOrgID  int64    `json:"parentOrgId"`
	Currency     string   `json:"currency"`
	TimeZone     string   `json:"timeZone"`
	PaymentModel string   `json:"paymentModel"`
	RoleNames    []string `json:"roleNames"`
}

type UserACLListResponse struct {
	Data       []UserACL      `json:"data"`
	Pagination map[string]any `json:"pagination"`
	Error      map[string]any `json:"error"`
}

func (c *Client) ListUserACLs(ctx context.Context) (*UserACLListResponse, error) {
	var resp UserACLListResponse
	if err := c.doJSON(ctx, http.MethodGet, "/acls", url.Values{}, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
