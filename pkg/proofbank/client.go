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
}

func NewClient(indexUrl string) (*Client, error) {
	parsedUrl, err := url.Parse(indexUrl)
	if err != nil {
		return nil, err
	}

	return &Client{
		indexUrl:   parsedUrl,
		httpClient: &http.Client{},
	}, nil
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
