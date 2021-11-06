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
	"github.com/h3isenbug/url-shortener/internal/types"
	"github.com/stretchr/testify/suite"
)

type HappyTestSuite struct {
	suite.Suite

	cleanup    func()
	server     *httptest.Server
	serverPort string
	client     *http.Client

	accessToken  string
	refreshToken string

	slug            string
	recommendedSlug string
	originalUrl     string
}

func (s *HappyTestSuite) SetupSuite() {
	app, cleanup, err := di.Inject()
	s.Require().NoError(err)

	s.cleanup = cleanup
	s.server = httptest.NewServer(app.Handler())
	parts := strings.Split(s.server.Listener.Addr().String(), ":")
	s.Require().NoError(err)
	s.serverPort = parts[1]

	s.client = s.server.Client()
	s.client.CheckRedirect = func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse }

	s.originalUrl = "https://google.com/"
	s.recommendedSlug = "goog"
}
func (s *HappyTestSuite) sendRequest(method, path, host, accessToken string, body io.Reader) (*http.Response, error) {
	request, err := http.NewRequest(method, fmt.Sprintf("http://%s:%s%s", host, s.serverPort, path), body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if accessToken != "" {
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	}

	return s.client.Do(request)
}
func (s *HappyTestSuite) Test_00_Register() {
	body, err := json.Marshal(map[string]string{
		"email":    "h.kalantari.1997@gmail.com",
		"password": "what does the fox say?",
	})
	s.Require().NoError(err)

	response, err := s.sendRequest(
		"POST", "/api/auth/register",
		"short.ir", "", bytes.NewBuffer(body),
	)
	s.Require().NoError(err)

	s.Require().Equal(
		http.StatusOK,
		response.StatusCode,
		fmt.Sprintf(
			"request failed with status code: %d",
			response.StatusCode,
		),
	)
	var parsedResponse struct {
		Message string `json:"message"`
	}

	s.Require().NoError(json.NewDecoder(response.Body).Decode(&parsedResponse))
}

func (s *HappyTestSuite) Test_01_Login() {
	body, err := json.Marshal(map[string]string{
		"email":    "h.kalantari.1997@gmail.com",
		"password": "what does the fox say?",
	})
	s.Require().NoError(err)

	response, err := s.sendRequest(
		"POST", "/api/auth/login",
		"short.ir", "", bytes.NewBuffer(body),
	)
	s.Require().NoError(err)

	s.Require().Equal(
		http.StatusOK,
		response.StatusCode,
		fmt.Sprintf(
			"request failed with status code: %d",
			response.StatusCode,
		),
	)

	var parsedResponse struct {
		AccessToken  string `json:"accessToken"`
		RefreshToken string `json:"refreshToken"`
	}

	s.Require().NoError(json.NewDecoder(response.Body).Decode(&parsedResponse))

	s.accessToken = parsedResponse.AccessToken
	s.refreshToken = parsedResponse.RefreshToken
}

func (s *HappyTestSuite) Test_02_RenewAccessToken() {
	body, err := json.Marshal(map[string]string{
		"oldAccessToken": s.accessToken,
		"refreshToken":   s.refreshToken,
	})
	s.Require().NoError(err)

	response, err := s.sendRequest(
		"POST", "/api/auth/renew",
		"short.ir", "", bytes.NewBuffer(body),
	)
	s.Require().NoError(err)

	s.Require().Equal(
		http.StatusOK,
		response.StatusCode,
		fmt.Sprintf(
			"request failed with status code: %d",
			response.StatusCode,
		),
	)

	var parsedResponse struct {
		AccessToken  string `json:"accessToken"`
		RefreshToken string `json:"refreshToken"`
	}

	s.Require().NoError(json.NewDecoder(response.Body).Decode(&parsedResponse))

	s.accessToken = parsedResponse.AccessToken
	s.refreshToken = parsedResponse.RefreshToken
}
func (s *HappyTestSuite) Test_03_CreateShortUrl() {
	body, err := json.Marshal(map[string]string{
		"originalUrl": s.originalUrl,
		"slug":        s.recommendedSlug,
	})
	s.Require().NoError(err)

	response, err := s.sendRequest(
		"POST", "/api/url",
		"short.ir", s.accessToken, bytes.NewBuffer(body),
	)
	s.Require().NoError(err)

	s.Require().Equal(
		http.StatusCreated,
		response.StatusCode,
		fmt.Sprintf(
			"request failed with status code: %d",
			response.StatusCode,
		),
	)

	var parsedResponse struct {
		OriginalUrl string `json:"originalUrl"`
		Slug        string `json:"slug"`
	}

	s.Require().NoError(json.NewDecoder(response.Body).Decode(&parsedResponse))

	s.slug = parsedResponse.Slug
}

func (s *HappyTestSuite) Test_04_VisitShortUrl() {
	response, err := s.sendRequest(
		"GET", "/"+s.slug,
		"s3t.ir", s.accessToken, nil,
	)
	s.Require().NoError(err)

	s.Require().Equal(
		http.StatusFound,
		response.StatusCode,
		fmt.Sprintf(
			"request failed with status code: %d",
			response.StatusCode,
		),
	)

	location, err := response.Location()
	s.Require().NoError(err)
	s.Require().Equal(s.originalUrl, location.String())
}

func (s *HappyTestSuite) Test_05_CheckUrlInDashboard() {
	response, err := s.sendRequest(
		"GET", "/api/url",
		"short.ir", s.accessToken, nil,
	)
	s.Require().NoError(err)

	s.Require().Equal(
		http.StatusOK,
		response.StatusCode,
		fmt.Sprintf(
			"request failed with status code: %d",
			response.StatusCode,
		),
	)

	var parsedResponse struct {
		Items      []types.Url `json:"items"`
		NextCursor string      `json:"nextCursor"`
	}

	s.Require().NoError(json.NewDecoder(response.Body).Decode(&parsedResponse))

	s.Require().Len(parsedResponse.Items, 1)

	s.Equal(s.originalUrl, parsedResponse.Items[0].OriginalUrl)
	s.Equal(false, parsedResponse.Items[0].Disabled)
	s.Equal(s.recommendedSlug, parsedResponse.Items[0].Slug)
	s.EqualValues(uint64(1), parsedResponse.Items[0].TotalVisits)
	s.EqualValues(uint64(1), parsedResponse.Items[0].UniqueVisits)
}

func (s *HappyTestSuite) Test_06_DisableUrl() {
	body, err := json.Marshal(map[string]interface{}{
		"disabled": true,
	})
	s.Require().NoError(err)

	response, err := s.sendRequest(
		"PATCH", "/api/url/"+s.slug,
		"short.ir", s.accessToken, bytes.NewBuffer(body),
	)
	s.Require().NoError(err)

	s.Require().Equal(
		http.StatusOK,
		response.StatusCode,
		fmt.Sprintf(
			"request failed with status code: %d",
			response.StatusCode,
		),
	)
}

func (s *HappyTestSuite) Test_07_CheckUrlInDashboardAfterDisabling() {
	response, err := s.sendRequest(
		"GET", "/api/url",
		"short.ir", s.accessToken, nil,
	)
	s.Require().NoError(err)

	s.Require().Equal(
		http.StatusOK,
		response.StatusCode,
		fmt.Sprintf(
			"request failed with status code: %d",
			response.StatusCode,
		),
	)

	var parsedResponse struct {
		Items      []types.Url `json:"items"`
		NextCursor string      `json:"nextCursor"`
	}

	s.Require().NoError(json.NewDecoder(response.Body).Decode(&parsedResponse))

	s.Require().Len(parsedResponse.Items, 1)

	s.Equal(s.originalUrl, parsedResponse.Items[0].OriginalUrl)
	s.Equal(true, parsedResponse.Items[0].Disabled)
	s.Equal(s.recommendedSlug, parsedResponse.Items[0].Slug)
	s.EqualValues(uint64(1), parsedResponse.Items[0].TotalVisits)
	s.EqualValues(uint64(1), parsedResponse.Items[0].UniqueVisits)
}

func (s *HappyTestSuite) Test_08_EnableUrl() {
	body, err := json.Marshal(map[string]interface{}{
		"disabled": false,
	})
	s.Require().NoError(err)

	response, err := s.sendRequest(
		"PATCH", "/api/url/"+s.slug,
		"short.ir", s.accessToken, bytes.NewBuffer(body),
	)
	s.Require().NoError(err)

	s.Require().Equal(
		http.StatusOK,
		response.StatusCode,
		fmt.Sprintf(
			"request failed with status code: %d",
			response.StatusCode,
		),
	)
}

func (s *HappyTestSuite) Test_09_CheckUrlInDashboardAfterReEnabling() {
	response, err := s.sendRequest(
		"GET", "/api/url",
		"short.ir", s.accessToken, nil,
	)
	s.Require().NoError(err)

	s.Require().Equal(
		http.StatusOK,
		response.StatusCode,
		fmt.Sprintf(
			"request failed with status code: %d",
			response.StatusCode,
		),
	)

	var parsedResponse struct {
		Items      []types.Url `json:"items"`
		NextCursor string      `json:"nextCursor"`
	}

	s.Require().NoError(json.NewDecoder(response.Body).Decode(&parsedResponse))

	s.Require().Len(parsedResponse.Items, 1)

	s.Equal(s.originalUrl, parsedResponse.Items[0].OriginalUrl)
	s.Equal(false, parsedResponse.Items[0].Disabled)
	s.Equal(s.recommendedSlug, parsedResponse.Items[0].Slug)
	s.EqualValues(uint64(1), parsedResponse.Items[0].TotalVisits)
	s.EqualValues(uint64(1), parsedResponse.Items[0].UniqueVisits)
}

func (s *HappyTestSuite) Test_10_CreateShortLinkWithoutRecommendedSlug() {

	body, err := json.Marshal(map[string]string{
		"originalUrl": s.originalUrl,
	})
	s.Require().NoError(err)

	response, err := s.sendRequest(
		"POST", "/api/url",
		"short.ir", s.accessToken, bytes.NewBuffer(body),
	)
	s.Require().NoError(err)

	s.Require().Equal(
		http.StatusCreated,
		response.StatusCode,
		fmt.Sprintf(
			"request failed with status code: %d",
			response.StatusCode,
		),
	)

	var parsedResponse struct {
		OriginalUrl string `json:"originalUrl"`
		Slug        string `json:"slug"`
	}

	s.Require().NoError(json.NewDecoder(response.Body).Decode(&parsedResponse))
	s.Greater(len(parsedResponse.Slug), 0, "empty slug generated")
	s.Equal(s.originalUrl, parsedResponse.OriginalUrl)
}

func (s *HappyTestSuite) TearDownSuite() {
	s.cleanup()
	s.server.Close()
}
func TestHappySuite(t *testing.T) {
	suite.Run(t, &HappyTestSuite{})
}
