package reports_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"

	"github.com/it-akumi/toggl-go/reports"
)

type detailedReport struct {
	TotalCount int `json:"total_count"`
	PerPage    int `json:"per_page"`
	Data       []struct {
		User        string `json:"user"`
		Project     string `json:"project"`
		Description string `json:"description"`
	} `json:"data"`
}

func TestGetDetailed(t *testing.T) {
	cases := []struct {
		name             string
		httpStatus       int
		testdataFilePath string
		in               context.Context
		out              error
	}{
		{
			name:             "200 OK",
			httpStatus:       http.StatusOK,
			testdataFilePath: "testdata/detailed.json",
			in:               context.Background(),
			out:              nil,
		},
		{
			name:             "401 Unauthorized",
			httpStatus:       http.StatusUnauthorized,
			testdataFilePath: "testdata/401_unauthorized.json",
			in:               context.Background(),
			out: &reports.ReportsError{
				Err: struct {
					Message string `json:"message"`
					Tip     string `json:"tip"`
					Code    int    `json:"code"`
				}{
					Message: "api token missing",
					Tip:     "You can find your API Token in your profile at https://www.toggl.com",
					Code:    http.StatusUnauthorized,
				},
			},
		},
		{
			name:             "429 Too Many Requests",
			httpStatus:       http.StatusTooManyRequests,
			testdataFilePath: "testdata/429_too_many_requests.html",
			in:               context.Background(),
			out: &reports.ReportsError{
				Err: struct {
					Message string `json:"message"`
					Tip     string `json:"tip"`
					Code    int    `json:"code"`
				}{
					Message: "Too Many Requests",
					Tip:     "Add delay between requests",
					Code:    http.StatusTooManyRequests,
				},
			},
		},
		{
			name:             "Without context",
			httpStatus:       http.StatusOK,
			testdataFilePath: "testdata/detailed.json",
			in:               nil,
			out:              reports.ErrContextNotFound,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			mockServer, testdata := setupMockServer(t, c.httpStatus, c.testdataFilePath)
			defer mockServer.Close()

			actualDetailedReport := new(detailedReport)
			client := reports.NewClient(apiToken, baseURL(mockServer.URL))
			err := client.GetDetailed(
				c.in,
				&reports.DetailedRequestParameters{
					StandardRequestParameters: &reports.StandardRequestParameters{
						UserAgent:   userAgent,
						WorkSpaceId: workSpaceId,
					},
				},
				actualDetailedReport,
			)

			if err == nil {
				expectedDetailedReport := new(detailedReport)
				if err := json.Unmarshal(testdata, expectedDetailedReport); err != nil {
					t.Error(err.Error())
				}
				if !reflect.DeepEqual(actualDetailedReport, expectedDetailedReport) {
					t.Errorf("\ngot : %+v\nwant: %+v\n", actualDetailedReport, expectedDetailedReport)
				}
			} else {
				if !reflect.DeepEqual(actualDetailedReport, &detailedReport{}) {
					t.Errorf("\ngot : %+v\nwant: %+v\n", actualDetailedReport, &detailedReport{})
				}
			}

			var reportsError reports.Error
			if errors.As(err, &reportsError) {
				if !reflect.DeepEqual(reportsError, c.out) {
					t.Errorf("\ngot : %#+v\nwant: %#+v\n", reportsError, c.out)
				}
			} else {
				if !errors.Is(err, c.out) {
					t.Errorf("\ngot : %#+v\nwant: %#+v\n", err, c.out)
				}
			}
		})
	}
}

func TestGetDetailedEncodeRequestParameters(t *testing.T) {
	expectedQueryString := url.Values{
		"user_agent":   []string{userAgent},
		"workspace_id": []string{workSpaceId},
		"page":         []string{"10"},
	}

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actualQueryString := r.URL.Query()
		if !reflect.DeepEqual(actualQueryString, expectedQueryString) {
			t.Error("Actual query string (" + actualQueryString.Encode() + ") is not as expected.")
		}
	}))

	client := reports.NewClient(apiToken, baseURL(mockServer.URL))
	_ = client.GetDetailed(
		context.Background(),
		&reports.DetailedRequestParameters{
			StandardRequestParameters: &reports.StandardRequestParameters{
				UserAgent:   userAgent,
				WorkSpaceId: workSpaceId,
			},
			Page: 10,
		},
		new(detailedReport),
	)
}
