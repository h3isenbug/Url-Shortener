package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/h3isenbug/url-shortener/cmd/url-shortener/di"
	presentation "github.com/h3isenbug/url-shortener/internal/presentation/http"
	"github.com/stretchr/testify/suite"
)

type UnauthorizedAccessTestSuite struct {
	suite.Suite

	cleanup    func()
	server     *httptest.Server
	serverPort string
	client     *http.Client
}

func (s *UnauthorizedAccessTestSuite) SetupSuite() {
	app, cleanup, err := di.Inject()
	s.Require().NoError(err)

	s.cleanup = cleanup
	s.server = httptest.NewServer(app.Handler())
	parts := strings.Split(s.server.Listener.Addr().String(), ":")
	s.Require().NoError(err)
	s.serverPort = parts[1]

	s.client = s.server.Client()
	s.client.CheckRedirect = func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse }
}
func (s *UnauthorizedAccessTestSuite) sendRequest(method, path, host string, body io.Reader) (*http.Response, error) {
	request, err := http.NewRequest(method, fmt.Sprintf("http://%s:%s%s", host, s.serverPort, path), body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	return s.client.Do(request)
}

func (s *UnauthorizedAccessTestSuite) assertUnauthorizedResponse(response *http.Response) {
	var parsedResponse presentation.ResponseWithMessage
	s.Equal(
		http.StatusUnauthorized,
		response.StatusCode,
		fmt.Sprintf(
			"request failed with status code: %d",
			response.StatusCode,
		),
	)

	s.NoError(json.NewDecoder(response.Body).Decode(&parsedResponse))
	s.Equal(http.StatusText(http.StatusUnauthorized), parsedResponse.Message)
}

func (s *UnauthorizedAccessTestSuite) Test_00_CreateShortUrl() {
	body, err := json.Marshal(map[string]string{
		"originalUrl": "https://twitter.com",
		"slug":        "twttr",
	})
	s.Require().NoError(err)

	response, err := s.sendRequest(
		"POST", "/api/url",
		"short.ir", bytes.NewBuffer(body),
	)
	s.Require().NoError(err)

	s.assertUnauthorizedResponse(response)
}

func (s *UnauthorizedAccessTestSuite) Test_01_CheckUrlInDashboard() {
	response, err := s.sendRequest(
		"GET", "/api/url",
		"short.ir", nil,
	)
	s.Require().NoError(err)
	s.assertUnauthorizedResponse(response)
}

func (s *UnauthorizedAccessTestSuite) TearDownSuite() {
	s.cleanup()
	s.server.Close()
}
func TestUnauthorizedAccessSuite(t *testing.T) {
	suite.Run(t, &UnauthorizedAccessTestSuite{})
}
