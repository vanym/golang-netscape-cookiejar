package netscapecookiejar

import (
	"bufio"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
	"sync"
)

// Options are the options for creating a new Jar.
// If AutoWritePath is zero, auto writing will not be done on every cookies change.
type Options struct {
	SubJar        http.CookieJar
	AutoWritePath string
	WriteHeader   bool
}

// Jar implements the http.CookieJar interface from the net/http package.
type Jar struct {
	subJar http.CookieJar

	autoWritePath string
	writeHeader   bool

	mutex   sync.Mutex
	entries map[entryKey]*http.Cookie
}

// New returns a new netscape cookie jar.
// A nil *Options is equivalent to a zero Options.
func New(options *Options) (*Jar, error) {
	jar := &Jar{
		entries: make(map[entryKey]*http.Cookie),
	}
	if options != nil {
		jar.subJar = options.SubJar
		jar.autoWritePath = options.AutoWritePath
		jar.writeHeader = options.WriteHeader
	}
	if jar.subJar == nil {
		var err error
		jar.subJar, err = cookiejar.New(nil)
		if err != nil {
			return nil, err
		}
	}
	return jar, nil
}

// Cookies implements the Cookies method of the http.CookieJar interface.
func (j *Jar) Cookies(u *url.URL) (cookies []*http.Cookie) {
	return j.subJar.Cookies(u)
}

// SetCookies implements the SetCookies method of the http.CookieJar interface.
func (j *Jar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	j.subJar.SetCookies(u, cookies)
	j.mutex.Lock()
	defer j.mutex.Unlock()
	var modified bool
	for _, cookie := range cookies {
		modified = j.putCookie(cookie) || modified
	}
	if modified && j.autoWritePath != "" {
		file, err := os.Create(j.autoWritePath)
		if err != nil {
			panic(err)
		}
		defer func() {
			err := file.Close()
			if err != nil {
				panic(err)
			}
		}()
		_, err = j.unlockedWriteTo(file)
		if err != nil {
			panic(err)
		}
	}
}

type entryKey struct {
	Domain, Path, Name string
}

func (j *Jar) putCookie(cookie *http.Cookie) bool {
	key := entryKey{
		Domain: cookie.Domain,
		Path:   cookie.Path,
		Name:   cookie.Name,
	}
	e := j.entries[key]
	j.entries[key] = cookie
	if e != nil && e.Secure == cookie.Secure && e.Expires == cookie.Expires && e.Value == cookie.Value {
		return false
	}
	return true
}

type counter int64

func (c *counter) Write(p []byte) (n int, err error) {
	n = len(p)
	*c += counter(n)
	return
}

// ReadFrom methed reads cookies from netscape cookie jar.
func (j *Jar) ReadFrom(r io.Reader) (int64, error) {
	var c counter
	scanner := bufio.NewScanner(io.TeeReader(r, &c))
	scanner.Split(bufio.ScanLines)
	var cookies []*http.Cookie
	for scanner.Scan() {
		line := scanner.Text()
		cookie, err := Unmarshal(line)
		if err != nil {
			return int64(c), err
		}
		if cookie == nil {
			continue
		}
		cookies = append(cookies, cookie)
	}
	type key struct {
		Secure       bool
		Domain, Path string
	}
	j.mutex.Lock()
	defer j.mutex.Unlock()
	var modified bool
	cookiesMap := make(map[key][]*http.Cookie)
	for _, cookie := range cookies {
		domain := cookie.Domain
		if strings.HasPrefix(domain, ".") {
			domain = domain[1:]
		}
		k := key{
			Secure: cookie.Secure,
			Domain: domain,
			Path:   cookie.Path,
		}
		cookiesMap[k] = append(cookiesMap[k], cookie)
		modified = j.putCookie(cookie) || modified
	}
	for k, v := range cookiesMap {
		u := new(url.URL)
		if k.Secure {
			u.Scheme = "https"
		} else {
			u.Scheme = "http"
		}
		u.Host = k.Domain
		u.Path = k.Path
		j.subJar.SetCookies(u, v)
	}
	return int64(c), nil
}

// WriteTo method writes netscape cookie jar.
func (j *Jar) WriteTo(w io.Writer) (n int64, err error) {
	j.mutex.Lock()
	defer j.mutex.Unlock()
	return j.unlockedWriteTo(w)
}

func (j *Jar) unlockedWriteTo(w io.Writer) (n int64, err error) {
	if j.writeHeader {
		var sn int
		sn, err = w.Write([]byte("# Netscape HTTP Cookie File\n# https://curl.se/docs/http-cookies.html\n\n"))
		n += int64(sn)
		if err != nil {
			return
		}
	}
	for _, cookie := range j.entries {
		var line string
		line, err = Marshal(cookie)
		if err != nil {
			return
		}
		var sn int
		sn, err = w.Write([]byte(line + "\n"))
		n += int64(sn)
		if err != nil {
			return
		}
	}
	return
}
