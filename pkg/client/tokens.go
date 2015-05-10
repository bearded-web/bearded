package client

import (
	"github.com/bearded-web/bearded/models/token"
	"golang.org/x/net/context"
)

const tokensUrl = "tokens"

type TokensService struct {
	client *Client
}

func (s *TokensService) String() string {
	return Stringify(s)
}

type TokensListOpts struct {
	Name    string `url:"name"`
	Version string `url:"version"`
	Type    string `url:"type"`
}

// List tokens.
//
//
func (s *TokensService) List(ctx context.Context, opt *TokensListOpts) (*token.TokenList, error) {
	tokenList := &token.TokenList{}
	return tokenList, s.client.List(ctx, tokensUrl, opt, tokenList)
}

func (s *TokensService) Get(ctx context.Context, id string) (*token.Token, error) {
	token := &token.Token{}
	return token, s.client.Get(ctx, tokensUrl, id, token)
}

func (s *TokensService) Create(ctx context.Context, src *token.Token) (*token.Token, error) {
	pl := &token.Token{}
	err := s.client.Create(ctx, tokensUrl, src, pl)
	if err != nil {
		return nil, err
	}
	return pl, nil
}

func (s *TokensService) Update(ctx context.Context, src *token.Token) (*token.Token, error) {
	pl := &token.Token{}
	id := FromId(src.Id)
	err := s.client.Update(ctx, tokensUrl, id, src, pl)
	if err != nil {
		return nil, err
	}
	return pl, nil
}

func (s *TokensService) Delete(ctx context.Context, id string) error {
	return s.client.Delete(ctx, tokensUrl, id)
}
