package proofbank

import (
	"bytes"
	"errors"
	genproto "github.com/proofman-dev/commons/protos/generated"
	"google.golang.org/protobuf/encoding/protojson"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

type Client struct {
	indexUrl   *url.URL
	httpClient *http.Client
	apiToken   string
}

func NewAuthenticatedClient(indexUrl, token string) (*Client, error) {
	parsedUrl, err := url.Parse(indexUrl)
	if err != nil {
		return nil, err
	}

	return &Client{
		indexUrl:   parsedUrl,
		httpClient: &http.Client{},
		apiToken:   token,
	}, nil
}

func NewUnauthenticatedClient(indexUrl string) (*Client, error) {
	return NewAuthenticatedClient(indexUrl, "")
}

func (c *Client) GetPackageInfo(packageName string) (*genproto.PackageInfoResponse, error) {
	response, err := c.httpClient.Get(c.indexUrl.JoinPath("api", "packages", packageName).String())
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, errors.New("unexpected status code: " + strconv.Itoa(response.StatusCode))
	}

	unmarshalled := &genproto.PackageInfoResponse{}
	err = protojson.Unmarshal(body, unmarshalled)

	return unmarshalled, err
}

func (c *Client) GetPackageVersionInfo(packageName, version string) (*genproto.PackageVersionInfoResponse, error) {
	response, err := c.httpClient.Get(c.indexUrl.JoinPath("api", "packages", packageName, version).String())
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, errors.New("unexpected status code: " + strconv.Itoa(response.StatusCode))
	}

	unmarshalled := &genproto.PackageVersionInfoResponse{}
	err = protojson.Unmarshal(body, unmarshalled)

	return unmarshalled, err
}

func (c *Client) QueryPackagesForSessions(sessions []string) (*genproto.PackageQueryResponse, error) {
	body, err := protojson.Marshal(&genproto.PackageQueryRequest{
		Sessions: sessions,
	})
	if err != nil {
		return nil, err
	}

	response, err := c.httpClient.Post(
		c.indexUrl.JoinPath("api", "packages", "query").String(),
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, err
	}

	body, err = io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, errors.New("unexpected status code: " + strconv.Itoa(response.StatusCode))
	}

	unmarshalled := &genproto.PackageQueryResponse{}
	err = protojson.Unmarshal(body, unmarshalled)

	return unmarshalled, err
}

func (c *Client) UploadPackage(archive io.Reader) error {
	req, err := http.NewRequest(
		"POST",
		c.indexUrl.JoinPath("api", "packages").String(),
		archive,
	)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("Authorization", "Bearer "+c.apiToken)

	response, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	// TODO - StatusCreated? - waiting on proofbank impl
	if response.StatusCode != http.StatusOK {
		return errors.New("unexpected status code: " + strconv.Itoa(response.StatusCode))
	}
	return nil
}
