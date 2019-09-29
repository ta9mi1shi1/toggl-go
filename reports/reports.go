/*
Package reports is a library of Toggl Reports API v2 for Go programming language.

This package deals with 3 types of reports, detailed report, summary report, and weekly report.
Though each report has their own data structure of successful response, they're not defined in this package.
Users must define a structure corresponding responses of each report to use before sending request.

See API documentation for more details.
https://github.com/toggl/toggl_api_docs/blob/master/reports.md
*/
package reports

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const (
	basicAuthPassword string = "api_token" // Defined in Toggl Reports API
	defaultBaseURL    string = "https://toggl.com"
)

// Client implements a basic request handling used by all of the reports.
type Client struct {
	httpClient *http.Client
	apiToken   string
	header     http.Header
	url        *url.URL
}

// StandardRequestParameters represents request parameters used in all of the reports.
type StandardRequestParameters struct {
	UserAgent           string
	WorkSpaceId         string
	Since               time.Time
	Until               time.Time
	Billable            string
	ClientIds           string
	ProjectIds          string
	UserIds             string
	MembersOfGroupIds   string
	OrMembersOfGroupIds string
	TagIds              string
	TaskIds             string
	TimeEntryIds        string
	Description         string
	WithoutDescription  bool
	OrderField          string
	OrderDesc           bool
	DistinctRates       bool
	Rounding            bool
	DisplayHours        string
}

func (params *StandardRequestParameters) values() url.Values {
	values := url.Values{}

	// user_agent and workspace_id are required.
	values.Add("user_agent", params.UserAgent)
	values.Add("workspace_id", params.WorkSpaceId)
	// since and until must be ISO 8601 date (YYYY-MM-DD) format
	if !params.Since.IsZero() {
		values.Add("since", params.Since.Format("2006-01-02"))
	}
	if !params.Until.IsZero() {
		values.Add("until", params.Until.Format("2006-01-02"))
	}
	if params.Billable != "" {
		values.Add("billable", params.Billable)
	}
	if params.ClientIds != "" {
		values.Add("client_ids", params.ClientIds)
	}
	if params.ProjectIds != "" {
		values.Add("project_ids", params.ProjectIds)
	}
	if params.UserIds != "" {
		values.Add("user_ids", params.UserIds)
	}
	if params.MembersOfGroupIds != "" {
		values.Add("members_of_group_ids", params.MembersOfGroupIds)
	}
	if params.OrMembersOfGroupIds != "" {
		values.Add("or_members_of_group_ids", params.OrMembersOfGroupIds)
	}
	if params.TagIds != "" {
		values.Add("tag_ids", params.TagIds)
	}
	if params.TaskIds != "" {
		values.Add("task_ids", params.TaskIds)
	}
	if params.TimeEntryIds != "" {
		values.Add("time_entry_ids", params.TimeEntryIds)
	}
	if params.Description != "" {
		values.Add("description", params.Description)
	}
	if params.WithoutDescription == true {
		values.Add("without_description", "true")
	}
	if params.OrderField != "" {
		values.Add("order_field", params.OrderField)
	}
	if params.OrderDesc == true {
		values.Add("order_desc", "on")
	}
	if params.DistinctRates == true {
		values.Add("distinct_rates", "on")
	}
	if params.Rounding == true {
		values.Add("rounding", "on")
	}
	if params.DisplayHours != "" {
		values.Add("display_hours", params.DisplayHours)
	}

	return values
}

type urlEncoder interface {
	urlEncode() string
}

// ReportsError represents a response of unsuccessful request.
type ReportsError struct {
	Err struct {
		Message    string `json:"message"`
		Tip        string `json:"tip"`
		StatusCode int    `json:"code"`
	} `json:"error"`
}

func (e ReportsError) Error() string {
	return fmt.Sprintf(
		"HTTP Status: %d\n%s\n\n%s\n",
		e.Err.StatusCode,
		e.Err.Message,
		e.Err.Tip,
	)
}

// Option represents optional parameters of NewClient.
type Option func(c *Client)

// baseURL makes client testable by configurable URL.
func baseURL(rawurl string) Option {
	return func(c *Client) {
		url, _ := url.Parse(rawurl)
		c.url = url
	}
}

// HTTPClient sets an HTTP client to use when sending requests.
// By default, http.DefaultClient is used.
func HTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// NewClient returns a pointer to a new initialized client.
func NewClient(apiToken string, options ...Option) *Client {
	url, _ := url.Parse(defaultBaseURL)
	newClient := &Client{
		httpClient: http.DefaultClient,
		apiToken:   apiToken,
		header:     make(http.Header),
		url:        url,
	}
	newClient.header.Set("Content-type", "application/json")
	for _, option := range options {
		option(newClient)
	}
	return newClient
}

func (c *Client) buildURL(endpoint string, params urlEncoder) string {
	c.url.Path = endpoint
	return c.url.String() + "?" + params.urlEncode()
}

func (c *Client) get(ctx context.Context, url string, report interface{}) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth(c.apiToken, basicAuthPassword)

	if ctx == nil {
		return fmt.Errorf("The provided ctx must be non-nil")
	}
	req = req.WithContext(ctx)

	resp, err := checkResponse(c.httpClient.Do(req))
	if err != nil {
		return err
	}
	if err = decodeJSON(resp, report); err != nil {
		return err
	}
	return nil
}

func checkResponse(resp *http.Response, err error) (*http.Response, error) {
	if err != nil {
		return nil, err
	}
	if resp.StatusCode <= 199 || 300 <= resp.StatusCode {
		var reportsError = new(ReportsError)
		if err := decodeJSON(resp, reportsError); err != nil {
			return nil, err
		}
		return nil, reportsError
	}
	return resp, nil
}

func decodeJSON(resp *http.Response, out interface{}) error {
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	return decoder.Decode(out)
}
