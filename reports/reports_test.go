package reports

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestNewClient(t *testing.T) {
	expectedAPIToken := "api_token"
	client := NewClient(expectedAPIToken)

	if client.apiToken != expectedAPIToken {
		t.Error("client.apiToken = " + client.apiToken + ", [Expected: " + expectedAPIToken + "]")
	}

	if client.url.String() != defaultBaseURL {
		t.Error("client.url.String() = " + client.url.String() + ", [Expected: " + defaultBaseURL + "]")
	}

	expectedContentType := "application/json"
	if client.header.Get("Content-type") != expectedContentType {
		t.Error(`client.header.Get("Content-type") = ` + client.header.Get("Content-type") + ", [Expected: " + expectedContentType + "]")
	}
}

const (
	apiToken    string = "api_token"
	userAgent   string = "user_agent"
	workSpaceId string = "workspace_id"
)

type detailedReport struct {
	Data []struct {
		User        string `json:"user"`
		Project     string `json:"project"`
		Description string `json:"description"`
	} `json:"data"`
}

func setupMockServerWithOk(t *testing.T, testdataFilePath string) (*httptest.Server, []byte) {
	testdata, err := ioutil.ReadFile(testdataFilePath)
	if err != nil {
		t.Error(err.Error())
	}

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, string(testdata))
	}))

	return mockServer, testdata
}

func setupMockServerWithError(t *testing.T) (*httptest.Server, []byte) {
	errorTestData, err := ioutil.ReadFile("testdata/error.json")
	if err != nil {
		t.Error(err.Error())
	}

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, string(errorTestData))
	}))

	return mockServer, errorTestData
}

func TestGetDetailedWithOk(t *testing.T) {
	mockServer, detailedTestData := setupMockServerWithOk(t, "testdata/detailed.json")
	defer mockServer.Close()

	actualDetailedReport := new(detailedReport)
	client := NewClient(apiToken, baseURL(mockServer.URL))
	err := client.GetDetailed(&DetailedRequestParameters{
		StandardRequestParameters: &StandardRequestParameters{
			UserAgent:   userAgent,
			WorkSpaceId: workSpaceId,
		},
	}, actualDetailedReport)
	if err != nil {
		t.Error("GetDetailed returns error though it gets '200 OK'")
	}

	expectedDetailedReport := new(detailedReport)
	if err := json.Unmarshal(detailedTestData, expectedDetailedReport); err != nil {
		t.Error(err.Error())
	}
	if !reflect.DeepEqual(actualDetailedReport, expectedDetailedReport) {
		t.Error("GetDetailed fails to decode detailedReport")
	}
}

func TestGetDetailedWithError(t *testing.T) {
	mockServer, errorTestData := setupMockServerWithError(t)
	defer mockServer.Close()

	client := NewClient(apiToken, baseURL(mockServer.URL))
	actualReportsError := client.GetDetailed(&DetailedRequestParameters{
		StandardRequestParameters: &StandardRequestParameters{
			UserAgent:   userAgent,
			WorkSpaceId: workSpaceId,
		},
	}, new(detailedReport))
	if actualReportsError == nil {
		t.Error("GetDetailed doesn't return error though it gets '401 Unauthorized'")
	}

	expectedReportsError := new(ReportsError)
	if err := json.Unmarshal(errorTestData, expectedReportsError); err != nil {
		t.Error(err.Error())
	}
	if !reflect.DeepEqual(actualReportsError, expectedReportsError) {
		t.Error("GetDetailed fails to decode ReportsError though it returns error as expected")
	}
}
