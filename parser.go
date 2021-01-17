package netscapecookiejar

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// based on https://curl.se/docs/http-cookies.html Cookie file format.

const httpOnlyPrefix = "#HttpOnly_"

// Marshal function converts *http.Cookie to netscape cookie jar line.
func Marshal(cookie *http.Cookie) (string, error) {
	tokens := make([]string, 7)

	domain := cookie.Domain
	includeSubdomains := strings.HasPrefix(cookie.Domain, ".")
	if cookie.HttpOnly {
		domain = httpOnlyPrefix + domain
	}
	tokens[0] = domain
	tokens[1] = strings.ToUpper(strconv.FormatBool(includeSubdomains))
	path := cookie.Path
	if path == "" {
		path = "/"
	}
	tokens[2] = path
	tokens[3] = strings.ToUpper(strconv.FormatBool(cookie.Secure))
	tokens[4] = strconv.FormatInt(cookie.Expires.Unix(), 10)
	tokens[5] = cookie.Name
	tokens[6] = cookie.Value

	return strings.Join(tokens, "\t"), nil
}

// Unmarshal function read netscape cookie jar line and returns *http.Cookie or nil.
func Unmarshal(line string) (*http.Cookie, error) {
	if line == "" || (strings.HasPrefix(line, "#") && !strings.HasPrefix(line, httpOnlyPrefix)) {
		return nil, nil
	}
	tokens := strings.SplitN(line, "\t", 7)
	if len(tokens) < 7 {
		return nil, fmt.Errorf("not enough tokens %d", len(tokens))
	}
	domain := tokens[0]
	includeSubdomains, err := strconv.ParseBool(tokens[1])
	if err != nil {
		return nil, err
	}
	path := tokens[2]
	secure, err := strconv.ParseBool(tokens[3])
	if err != nil {
		return nil, err
	}
	expires, err := strconv.ParseInt(tokens[4], 10, 64)
	if err != nil {
		return nil, err
	}
	name := tokens[5]
	value := tokens[6]

	var httpOnly bool
	if strings.HasPrefix(domain, httpOnlyPrefix) {
		domain = domain[len(httpOnlyPrefix):]
		httpOnly = true
	} else {
		httpOnly = false
	}

	if includeSubdomains {
		if !strings.HasPrefix(domain, ".") {
			domain = "." + domain
		}
	} else {
		if strings.HasPrefix(domain, ".") {
			domain = domain[1:]
		}
	}

	cookie := new(http.Cookie)
	cookie.Domain = domain
	cookie.HttpOnly = httpOnly
	cookie.Path = path
	cookie.Secure = secure
	cookie.Expires = time.Unix(expires, 0)
	cookie.Name = name
	cookie.Value = value
	return cookie, nil
}
