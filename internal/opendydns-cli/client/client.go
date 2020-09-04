package client

import (
	"fmt"
	"github.com/creekorful/open-dydns/internal/proto"
	"github.com/go-resty/resty/v2"
)

// TODO make sure everything's right

type Client struct {
	httpClient *resty.Client
}

func NewClient(baseUrl string) *Client {
	httpClient := resty.New()
	httpClient.SetHostURL(baseUrl)
	httpClient.SetAuthScheme("Bearer")

	return &Client{
		httpClient: httpClient,
	}
}

func (c *Client) Authenticate(cred proto.CredentialsDto) (proto.TokenDto, error) {
	var result proto.TokenDto
	var err proto.ErrorDto

	_, _ = c.httpClient.R().SetBody(cred).SetResult(&result).SetError(&err).Post("/sessions")

	return result, nonNilError(err)
}

func (c *Client) GetAliases(token proto.TokenDto) ([]proto.AliasDto, error) {
	var result []proto.AliasDto
	var err proto.ErrorDto

	_, _ = c.httpClient.R().SetAuthToken(token.Token).SetResult(&result).SetError(&err).Get("/aliases")

	return result, nonNilError(err)
}

func (c *Client) RegisterAlias(token proto.TokenDto, alias proto.AliasDto) (proto.AliasDto, error) {
	var result proto.AliasDto
	var err proto.ErrorDto

	_, _ = c.httpClient.R().SetAuthToken(token.Token).SetBody(alias).SetResult(&result).SetError(&err).Post("/aliases")

	return result, nonNilError(err)
}

func (c *Client) DeleteAlias(token proto.TokenDto, name string) error {
	var err proto.ErrorDto

	_, _ = c.httpClient.R().SetAuthToken(token.Token).Delete(fmt.Sprintf("/aliases/%s", name))

	return nonNilError(err)
}

func nonNilError(err proto.ErrorDto) error {
	if err.Message == "" {
		return nil
	}
	return &err
}