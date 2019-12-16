package common

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"sort"
	"strings"
)

const (
	AccountHeader = "X-XTAO-Account"
)
const (
	// common parameters
	authorizationHeader = "Authorization"
	apiKeyParam         = "APIKey"
	signatureParam      = "Signature"
	accountHeader       = AccountHeader
	// parsing bits
	empty   = ""
	comma   = ","
	space   = " "
	eqSign  = "="
	newline = "\n"
)

const (
	invalidTimestamp  = "Invalid timestamp. Requires RFC3339 format."
	invalidParameter  = "Invalid parameter in header string"
	missingParameter  = "Missing parameter in header string"
	invalidSignature  = "Invalid Signature"
	invalidAPIKey     = "Invalid APIKey"
	secretKeyRequired = "HMACAuth Secret Key Locator Required"
	repeatedParameter = "Repeated parameter: %q in header string"
	missingHeader     = "Missing required header: %q"
)

type HMACAuthError struct {
	Message string
}

func (e HMACAuthError) Error() string {
	return e.Message
}

type RepeatedParameterError struct {
	ParameterName string
}

func (e RepeatedParameterError) Error() string {
	return fmt.Sprintf(repeatedParameter, e.ParameterName)
}

type HeaderMissingError struct {
	HeaderName string
}

func (e HeaderMissingError) Error() string {
	return fmt.Sprintf(missingHeader, e.HeaderName)
}

type (
	KeyLocator func(string) (string, string)
)

type Options struct {
	SignedHeaders []string
	SecretKey     KeyLocator
}

type authBits struct {
	APIKey    string
	Signature string
}

func (ab *authBits) IsValid() bool {
	return ab.APIKey != empty &&
		ab.Signature != empty
}

func SignString(str string, secret string) string {
	hash := hmac.New(sha256.New, []byte(secret))
	hash.Write([]byte(str))
	return base64.StdEncoding.EncodeToString(hash.Sum(nil))
}

func StringToSign(req *http.Request, options *Options) (string, error) {
	var buffer bytes.Buffer

	// Standard
	buffer.WriteString(req.Method)
	buffer.WriteString(newline)
	buffer.WriteString(req.Host)
	buffer.WriteString(newline)
	buffer.WriteString(req.URL.RequestURI())
	buffer.WriteString(newline)

	// Headers
	sort.Strings(options.SignedHeaders)
	for _, header := range options.SignedHeaders {
		val := req.Header.Get(header)
		if val == empty {
			return empty, HeaderMissingError{header}
		}
		buffer.WriteString(val)
		buffer.WriteString(newline)
	}

	return buffer.String(), nil
}

func parseAuthHeader(header string) (*authBits, error) {
	if header == empty {
		return nil, HeaderMissingError{authorizationHeader}
	}

	ab := new(authBits)
	parts := strings.Split(header, comma)
	for _, part := range parts {
		kv := strings.SplitN(strings.Trim(part, space), eqSign, 2)
		if kv[0] == apiKeyParam {
			if ab.APIKey != empty {
				return nil, RepeatedParameterError{kv[0]}
			}
			ab.APIKey = kv[1]
		} else if kv[0] == signatureParam {
			if ab.Signature != empty {
				return nil, RepeatedParameterError{kv[0]}
			}
			ab.Signature = kv[1]
		} else {
			return nil, HMACAuthError{invalidParameter}
		}
	}

	if !ab.IsValid() {
		return nil, HMACAuthError{missingParameter}
	}

	return ab, nil
}

func HMACAuth(options Options, req *http.Request) (error, string, string) {
	// Validate options
	var err error = nil
	var ab *authBits
	var sk, u string

	if options.SecretKey == nil {
		err = HMACAuthError{secretKeyRequired}
		return err, empty, empty
	}

	if ab, err = parseAuthHeader(req.Header.Get(authorizationHeader)); err == nil {
		var sts string

		if sts, err = StringToSign(req, &options); err == nil {
			if sk, u = options.SecretKey(ab.APIKey); sk != empty {
				if ab.Signature != SignString(sts, sk) {
					err = HMACAuthError{invalidSignature}
				}
			} else {
				err = HMACAuthError{invalidAPIKey}
			}
		}
	}

	return err, u, sk
}
