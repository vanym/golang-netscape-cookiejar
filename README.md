# netscapecookiejar

netscapecookiejar is a golang package for storing cookies in [netscape cookie jar](https://curl.se/docs/http-cookies.html) file

## Usage

### Format looks like

```
www.example.com	FALSE	/	TRUE	1338534278	cookiename	value
```

### Importing

```go
import netscapecookiejar "github.com/vanym/golang-netscape-cookiejar"
```

### Marshal and Unmarshal

```go
cookie, err := netscapecookiejar.Unmarshal(".example.com\tTRUE\t/\tTRUE\t1338534278\tcookiename\tvalue")
line, err := netscapecookiejar.Marshal(cookie)
```

### Cookie jar

```go
subjar, err := cookiejar.New(&cookiejar.Options{})
jar, err := netscapecookiejar.New(&netscapecookiejar.Options{
    SubJar:        subjar,
    AutoWritePath: "cookies_auto.txt",
    WriteHeader:   true,
})
file, err := os.Open("cookies_read.txt")
_, err = jar.ReadFrom(file)
file.Close()
file2, err := os.Create("cookies_write.txt")
_, err = jar.WriteTo(file2)
file2.Close()
```
