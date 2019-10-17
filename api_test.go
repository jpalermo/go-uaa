package uaa_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/cloudfoundry-community/go-uaa"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/sclevine/spec"
	"golang.org/x/oauth2"
)

func testNew(t *testing.T, when spec.G, it spec.S) {
	it.Before(func() {
		RegisterTestingT(t)
	})

	when("TokenFormat.String()", func() {
		it("prints the string representation appropriately", func() {
			var t uaa.TokenFormat
			Expect(t.String()).To(Equal("opaque"))
			t = 3
			Expect(t.String()).To(Equal(""))
			Expect(uaa.JSONWebToken.String()).To(Equal("jwt"))
			Expect(uaa.OpaqueToken.String()).To(Equal("opaque"))
		})
	})

	when("New()", func() {
		it("does not returns an API if the target is an invalid URL", func() {
			api, err := uaa.New("(*#&^@%$&%)", uaa.WithToken(&oauth2.Token{
				AccessToken: "blergh",
				Expiry:      time.Now().Add(60 * time.Second),
			}))
			Expect(api).To(BeNil())
			Expect(err).NotTo(BeNil())
		})

		it("sets the TargetURL", func() {
			api, err := uaa.New("https://example.net", uaa.WithToken(&oauth2.Token{
				AccessToken: "blergh",
				Expiry:      time.Now().Add(60 * time.Second),
			}), uaa.WithZoneID("zone-1"))
			Expect(err).NotTo(HaveOccurred())
			Expect(api).NotTo(BeNil())
			Expect(api.TargetURL).NotTo(BeNil())
			Expect(api.TargetURL.String()).To(Equal("https://example.net"))
		})

		it("Token() fails because when there is no mechanism to get a token", func() {
			api := uaa.API{}
			t, err := api.Token(context.Background())
			Expect(err).To(HaveOccurred())
			Expect(t).To(BeNil())
		})
	})

	when("New() WithToken()", func() {
		it("fails if the target url is invalid", func() {
			api, err := uaa.New("(*#&^@%$&%)", uaa.WithToken(&oauth2.Token{Expiry: time.Now().Add(20 * time.Second), AccessToken: "test-token"}))
			Expect(err).To(HaveOccurred())
			Expect(api).To(BeNil())
		})

		it("fails if the token is invalid", func() {
			api, err := uaa.New("https://example.net", uaa.WithToken(&oauth2.Token{Expiry: time.Now().Add(20 * time.Second), AccessToken: ""}))
			Expect(err).To(HaveOccurred())
			Expect(api).To(BeNil())
			api, err = uaa.New("https://example.net", uaa.WithToken(&oauth2.Token{Expiry: time.Now().Add(-20 * time.Second), AccessToken: "test-token"}))
			Expect(err).To(HaveOccurred())
			Expect(api).To(BeNil())
		})

		it("returns an API with a TargetURL", func() {
			api, err := uaa.New("https://example.net", uaa.WithToken(&oauth2.Token{Expiry: time.Now().Add(20 * time.Second), AccessToken: "test-token"}))
			Expect(err).NotTo(HaveOccurred())
			Expect(api).NotTo(BeNil())
			Expect(api.TargetURL.String()).To(Equal("https://example.net"))
		})

		it("returns an API with an HTTPClient", func() {
			api, err := uaa.New("https://example.net", uaa.WithToken(&oauth2.Token{Expiry: time.Now().Add(20 * time.Second), AccessToken: "test-token"}))
			Expect(err).NotTo(HaveOccurred())
			Expect(api).NotTo(BeNil())
			Expect(api.Client).NotTo(BeNil())
			Expect(reflect.TypeOf(api.Client.Transport).String()).To(Equal("*uaa.tokenTransport"))
		})

		it("sets the authorization header correctly when round tripping", func() {
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				Expect(req.Header.Get("Authorization")).To(Equal("Bearer test-token"))
				w.WriteHeader(http.StatusOK)
			}))
			api, err := uaa.New("https://example.net", uaa.WithToken(&oauth2.Token{Expiry: time.Now().Add(20 * time.Second), AccessToken: "test-token"}))
			Expect(err).NotTo(HaveOccurred())
			Expect(api).NotTo(BeNil())
			Expect(api.Client).NotTo(BeNil())
			r, err := api.Client.Get(s.URL)
			Expect(err).NotTo(HaveOccurred())
			Expect(r.StatusCode).To(Equal(http.StatusOK))
		})

		it("Token() fails when the mode is token and the token is invalid", func() {
			t := &oauth2.Token{Expiry: time.Now().Add(20 * time.Second), AccessToken: "test-token"}
			api, err := uaa.New("https://example.net", uaa.WithToken(t))
			Expect(api).NotTo(BeNil())
			Expect(err).NotTo(HaveOccurred())
			t.Expiry = time.Now().Add(-20 * time.Second)
			t, err = api.Token(context.Background())
			Expect(err).To(HaveOccurred())
			Expect(t).To(BeNil())
		})

		it("Token() succeeds when the mode is token and the token is valid", func() {
			api, err := uaa.New("https://example.net", uaa.WithToken(&oauth2.Token{Expiry: time.Now().Add(20 * time.Second), AccessToken: "test-token"}))
			Expect(api).NotTo(BeNil())
			Expect(err).NotTo(HaveOccurred())
			t, err := api.Token(context.Background())
			Expect(err).NotTo(HaveOccurred())
			Expect(t).NotTo(BeNil())
			Expect(t.Valid()).To(BeTrue())
		})
	})

	when("WithClientCredentials()", func() {
		it("fails if the target url is invalid", func() {
			api, err := uaa.New("(*#&^@%$&%)", uaa.WithClientCredentials("", "", uaa.OpaqueToken))
			Expect(err).To(HaveOccurred())
			Expect(api).To(BeNil())
		})

		it("returns an API with a TargetURL", func() {
			api, err := uaa.New("https://example.net", uaa.WithClientCredentials("", "", uaa.OpaqueToken))
			Expect(err).NotTo(HaveOccurred())
			Expect(api).NotTo(BeNil())
			Expect(api.TargetURL.String()).To(Equal("https://example.net"))
		})

		it("returns an API with an HTTPClient", func() {
			api, err := uaa.New("https://example.net", uaa.WithClientCredentials("", "", uaa.OpaqueToken))
			Expect(err).NotTo(HaveOccurred())
			Expect(api).NotTo(BeNil())
			Expect(api.Client).NotTo(BeNil())
			Expect(reflect.TypeOf(api.Client.Transport.(*oauth2.Transport).Base).String()).To(Equal("*uaa.UaaTransport"))
		})

		when("the server returns tokens", func() {
			var (
				s *ghttp.Server
			)

			it.Before(func() {
				s = ghttp.NewServer()
				t := &oauth2.Token{
					AccessToken:  "test-access-token",
					RefreshToken: "test-refresh-token",
					TokenType:    "bearer",
					Expiry:       time.Now().Add(60 * time.Second),
				}
				s.AppendHandlers(ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/oauth/token"),
					ghttp.VerifyFormKV("grant_type", "client_credentials"),
					ghttp.VerifyFormKV("token_format", "opaque"),
					ghttp.RespondWithJSONEncoded(http.StatusOK, t),
				))
			})

			it.After(func() {
				if s != nil {
					s.Close()
				}
			})

			it("Token() succeeds when the mode is client credentials and the client credentials are valid", func() {
				api, err := uaa.New(s.URL(), uaa.WithClientCredentials("client-id", "client-secret", uaa.OpaqueToken))
				Expect(api).NotTo(BeNil())
				Expect(err).NotTo(HaveOccurred())
				api.TargetURL = nil
				t, err := api.Token(context.Background())
				Expect(err).NotTo(HaveOccurred())
				Expect(t).NotTo(BeNil())
				Expect(t.Valid()).To(BeTrue())
			})
		})
	})

	when("WithTransport()", func() {
		it("uses the provided transport for the client", func() {
			transport := http.RoundTripper(fakeTransport{})
			api, _ := uaa.New("https://example.net", uaa.WithNoAuthentication(), uaa.WithTransport(&transport))
			Expect(api.Client.Transport.(*uaa.UaaTransport).Transport).To(Equal(&transport))
		})
	})

	when("NewWithPasswordCredentials()", func() {
		it("fails if the target url is invalid", func() {
			api, err := uaa.New("(*#&^@%$&%)", uaa.WithPasswordCredentials("", "", "", "", uaa.OpaqueToken))
			Expect(err).To(HaveOccurred())
			Expect(api).To(BeNil())
		})

		it("returns an API with a TargetURL", func() {
			api, err := uaa.New("https://example.net", uaa.WithPasswordCredentials("", "", "", "", uaa.OpaqueToken))
			Expect(err).NotTo(HaveOccurred())
			Expect(api).NotTo(BeNil())
			Expect(api.TargetURL.String()).To(Equal("https://example.net"))
		})

		it("returns an API with an HTTPClient", func() {
			api, err := uaa.New("https://example.net", uaa.WithPasswordCredentials("", "", "", "", uaa.OpaqueToken))
			Expect(err).NotTo(HaveOccurred())
			Expect(api).NotTo(BeNil())
			Expect(api.Client).NotTo(BeNil())
			Expect(reflect.TypeOf(api.Client.Transport.(*oauth2.Transport).Base).String()).To(Equal("*uaa.UaaTransport"))
		})
	})

	when("NewWithAuthorizationCode", func() {
		var s *ghttp.Server
		redirectUrl, _ := url.ParseRequestURI("https://example.net")

		stubTokenRequest := func(clientId string, clientSecret string, authCode string, tokenFormat uaa.TokenFormat, response http.HandlerFunc) {
			s.AppendHandlers(ghttp.CombineHandlers(
				ghttp.VerifyRequest("POST", "/oauth/token"),
				ghttp.VerifyFormKV("grant_type", "authorization_code"),
				ghttp.VerifyFormKV("code", authCode),
				ghttp.VerifyFormKV("token_format", tokenFormat.String()),
				response,
			))
		}

		stubTokenSuccess := func(clientId string, clientSecret string, authCode string, tokenFormat uaa.TokenFormat) {
			t := &oauth2.Token{
				AccessToken:  "test-access-token",
				RefreshToken: "test-refresh-token",
				TokenType:    "bearer",
				Expiry:       time.Now().Add(60 * time.Second),
			}

			stubTokenRequest(clientId, clientSecret, authCode, tokenFormat, ghttp.RespondWithJSONEncoded(http.StatusOK, t))
		}

		stubMalformedTokenSuccess := func(clientId string, clientSecret string, authCode string, tokenFormat uaa.TokenFormat) {
			stubTokenRequest(clientId, clientSecret, authCode, tokenFormat, ghttp.RespondWithJSONEncoded(http.StatusOK, nil))
		}

		stubTokenFailure := func(clientId string, clientSecret string, authCode string, tokenFormat uaa.TokenFormat) {
			stubTokenRequest(clientId, clientSecret, authCode, tokenFormat, ghttp.RespondWithJSONEncoded(http.StatusBadRequest, nil))
		}

		it.Before(func() {
			s = ghttp.NewServer()
		})

		it.After(func() {
			s.Close()
		})

		when("success", func() {
			it.Before(func() {
				// Token retrieval is done as part of validateAuthorizationCode
				// validateAuthorizationCode is called two times on construction
				// AuthStyle is set to AuthStyleInHeader, failed token requests are not retried
				// Because the first token reqest succeeds, later token attempts are skipped
				// 1 token request, 1 attempt each => 1 request
				stubTokenSuccess("client-id", "client-secret", "auth-code", uaa.OpaqueToken)
			})

			it("returns an API with a TargetURL", func() {
				api, err := uaa.New(s.URL(), uaa.WithAuthorizationCode("client-id", "client-secret", "auth-code", uaa.OpaqueToken, redirectUrl))
				Expect(err).NotTo(HaveOccurred())
				Expect(api).NotTo(BeNil())
				Expect(api.TargetURL.String()).To(Equal(s.URL()))
			})

			it("returns an API with an HTTPClient", func() {
				api, err := uaa.New(s.URL(), uaa.WithAuthorizationCode("client-id", "client-secret", "auth-code", uaa.OpaqueToken, redirectUrl))
				Expect(err).NotTo(HaveOccurred())
				Expect(api).NotTo(BeNil())
				Expect(api.Client).NotTo(BeNil())
				Expect(reflect.TypeOf(api.Client.Transport.(*oauth2.Transport).Base).String()).To(Equal("*uaa.UaaTransport"))
			})
		})

		when("invalid target url", func() {
			it("returns an error", func() {
				api, err := uaa.New("(*#&^@%$&%)", uaa.WithAuthorizationCode("client-id", "client-secret", "auth-code", uaa.OpaqueToken, redirectUrl))
				Expect(err).To(HaveOccurred())
				Expect(api).To(BeNil())
			})
		})

		when("created with an invalid auth code", func() {
			it.Before(func() {
				// Token retrieval is done as part of validateAuthorizationCode
				// validateAuthorizationCode is called two times on construction
				// AuthStyle is set to AuthStyleInHeader, failed token requests are not retried
				// 2 token requests, 1 attempt each => 2 requests
				stubTokenFailure("client-id", "client-secret", "", uaa.JSONWebToken)
				stubTokenFailure("client-id", "client-secret", "", uaa.JSONWebToken)
			})

			it("returns an error", func() {
				api, err := uaa.New(s.URL(), uaa.WithAuthorizationCode("client-id", "client-secret", "", uaa.JSONWebToken, redirectUrl))
				Expect(err).To(HaveOccurred())
				Expect(api).To(BeNil())
			})
		})

		when("the token response is missing a token", func() {
			it.Before(func() {
				// Token retrieval is done as part of validateAuthorizationCode
				// validateAuthorizationCode is called two times on construction
				// AuthStyle is set to AuthStyleInHeader, failed token requests are not retried
				// 2 token requests, 2 attempts each => 4 requests
				stubMalformedTokenSuccess("client-id", "client-secret", "auth-code", uaa.OpaqueToken)
				stubMalformedTokenSuccess("client-id", "client-secret", "auth-code", uaa.OpaqueToken)
			})

			it("returns an error", func() {
				api, err := uaa.New(s.URL(), uaa.WithAuthorizationCode("client-id", "client-secret", "auth-code", uaa.OpaqueToken, redirectUrl))
				Expect(err).To(HaveOccurred())
				Expect(api).To(BeNil())
			})
		})

		when("the unauthenticatedClient is removed", func() {
			it.Before(func() {
				// Token retrieval is done as part of validateAuthorizationCode
				// validateAuthorizationCode is called two times on construction
				// AuthStyle is set to AuthStyleInHeader, failed token requests are not retried
				// Because the first token reqest succeeds, later token attempts are skipped
				// Then another token is explicitly requested
				// 2 token requests, 1 attempt each => 2 requests
				stubTokenSuccess("client-id", "client-secret", "auth-code", uaa.OpaqueToken)
				stubTokenSuccess("client-id", "client-secret", "auth-code", uaa.OpaqueToken)
			})
		})
	})

	when("NewWithRefreshToken", func() {
		var s *ghttp.Server

		stubTokenRequest := func(clientId string, clientSecret string, refreshToken string, tokenFormat uaa.TokenFormat, response http.HandlerFunc) {
			s.AppendHandlers(ghttp.CombineHandlers(
				ghttp.VerifyRequest("POST", "/oauth/token", fmt.Sprintf("token_format=%s", tokenFormat)),
				ghttp.VerifyFormKV("grant_type", "refresh_token"),
				ghttp.VerifyFormKV("refresh_token", refreshToken),
				response,
			))
		}

		stubTokenSuccess := func(clientId string, clientSecret string, refreshToken string, tokenFormat uaa.TokenFormat) {
			token := &oauth2.Token{
				AccessToken:  "test-access-token",
				RefreshToken: "test-refresh-token",
				TokenType:    "bearer",
				Expiry:       time.Now().Add(60 * time.Second),
			}

			stubTokenRequest(clientId, clientSecret, refreshToken, tokenFormat, ghttp.RespondWithJSONEncoded(http.StatusOK, token))
		}

		stubMalformedTokenSuccess := func(clientId string, clientSecret string, refreshToken string, tokenFormat uaa.TokenFormat) {
			stubTokenRequest(clientId, clientSecret, refreshToken, tokenFormat, ghttp.RespondWithJSONEncoded(http.StatusOK, nil))
		}

		it.Before(func() {
			s = ghttp.NewServer()
		})

		it.After(func() {
			s.Close()
		})

		when("success", func() {
			it.Before(func() {
				// Token retrieval is done as part of validateRefreshToken
				// validateRefreshToken is called two times on construction
				// AuthStyle is set to AuthStyleInHeader, failed token requests are not retried
				// Because the first token reqest succeeds, later token attempts are skipped
				// 1 token request, 1 attempt each => 1 request
				stubTokenSuccess("client-id", "client-secret", "refresh-token", uaa.JSONWebToken)
			})

			it("returns an API with a TargetURL", func() {
				api, err := uaa.New(s.URL(), uaa.WithRefreshToken("client-id", "client-secret", "refresh-token", uaa.JSONWebToken))
				Expect(err).NotTo(HaveOccurred())
				Expect(api).NotTo(BeNil())
				Expect(api.TargetURL.String()).To(Equal(s.URL()))
			})

			it("returns an API with an HTTPClient", func() {
				api, err := uaa.New(s.URL(), uaa.WithRefreshToken("client-id", "client-secret", "refresh-token", uaa.JSONWebToken))
				Expect(err).NotTo(HaveOccurred())
				Expect(api).NotTo(BeNil())
				Expect(api.Client).NotTo(BeNil())
				Expect(reflect.TypeOf(api.Client.Transport.(*oauth2.Transport).Base).String()).To(Equal("*uaa.UaaTransport"))
			})
		})

		when("created with an invalid target url", func() {
			it("returns an error", func() {
				api, err := uaa.New("(*#&^@%$&%)", uaa.WithRefreshToken("client-id", "client-secret", "refresh-token", uaa.JSONWebToken))
				Expect(err).To(HaveOccurred())
				Expect(api).To(BeNil())
			})
		})

		when("created with an invalid refresh token", func() {
			it("returns an error", func() {
				api, err := uaa.New(s.URL(), uaa.WithRefreshToken("client-id", "client-secret", "", uaa.JSONWebToken))
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("oauth2: token expired and refresh token is not set"))
				Expect(api).To(BeNil())
			})
		})

		when("the token response is missing a token", func() {
			it.Before(func() {
				// Token retrieval is done as part of validateRefreshToken
				// validateRefreshToken is called two times on construction
				// AuthStyle is set to AuthStyleInHeader, failed token requests are not retried
				// 2 token requests, 1 attempt each => 2 requests
				stubMalformedTokenSuccess("client-id", "client-secret", "refresh-token", uaa.JSONWebToken)
				stubMalformedTokenSuccess("client-id", "client-secret", "refresh-token", uaa.JSONWebToken)
			})

			it("returns an error", func() {
				api, err := uaa.New(s.URL(), uaa.WithRefreshToken("client-id", "client-secret", "refresh-token", uaa.JSONWebToken))
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("oauth2: server response missing access_token"))
				Expect(api).To(BeNil())
			})
		})

		when("the unauthenticatedClient is removed", func() {
			it.Before(func() {
				// Token retrieval is done as part of validateRefreshToken
				// validateRefreshToken is called two times on construction
				// AuthStyle is set to AuthStyleInHeader, failed token requests are not retried
				// Because the first token reqest succeeds, later token attempts are skipped
				// Then another token is explicitly requested
				// 2 token requests, 1 attempt each => 2 requests
				stubTokenSuccess("client-id", "client-secret", "refresh-token", uaa.JSONWebToken)
				stubTokenSuccess("client-id", "client-secret", "refresh-token", uaa.JSONWebToken)
			})
		})
	})
}

type fakeTransport struct {
}

func (transport fakeTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	return nil, nil
}
