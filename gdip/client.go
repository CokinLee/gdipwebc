package gdip

import (
	"crypto/md5"
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
)

// RequestCode represents registration mode.  See also REGISTER.
type RequestCode int

const (
	// Active register mode.  IP address must be specified.
	REGISTER RequestCode = iota

	// Go-offline mode.  (XXX: borken or some server does'nt implement it)
	OFFLINE

	// Passive mode.  The address seen by the server will be registered.
	REGISTER_PASSIVE
)

var (
	debug              = false
	logger *log.Logger = nil
)

func dolog(fmt string, args []interface{}) {
	if logger != nil {
		logger.Printf(fmt, args...)
	}
}

func logerr(fmt string, args ...interface{}) {
	dolog("E: "+fmt, args)
}

func loginfo(fmt string, args ...interface{}) {
	dolog("I: "+fmt, args)
}

func logdebug(fmt string, args ...interface{}) {
	if debug {
		dolog("D: "+fmt, args)
	}
}

// Set logger for logging.  default: nil (logging is disabled)
func SetLogger(l *log.Logger) {
	logger = l
}

// Turn on debug logging.  Note that logging is disabled by default.
// See also SetLogger.
func DebugOn() {
	debug = true
}

// Turn off debug logging.  Note that logging is disabled by default.
// See also SetLogger.
func DebugOff() {
	debug = false
}

// Client represents GnuDIP client
type Client struct {
	// URL specifies the URL of the GDIP http service
	URL string

	// User specifies the user name of the GDIP service
	User string

	// Password specifies the password for the User
	Password string

	// DomainName specifies the DNS name for update
	DomainName string

	// RequestCode
	requestCode RequestCode

	// Address specifies the IP Address for update
	// This is required when the register mode is REGISTER
	Address string

	// http.Client for update requests
	http *http.Client
}

type response struct {
	XMLName xml.Name `xml:"html"`
	Meta    []meta   `xml:"head>meta"`
}

type meta struct {
	Name    string `xml:"name,attr"`
	Content string `xml:"content,attr"`
}

func md5_hex(s string) string {
	hs := md5.Sum([]byte(s))
	xs := ""
	for _, b := range hs {
		xs = xs + fmt.Sprintf("%02x", b)
	}
	return xs
}

// New returns new Client.  if reqCode is REGISTER,  addr must be specified.
// if reqCode is REGISTER_PASSIVE, addr is ignored.
func New(url, user, passwd, domain string, reqCode RequestCode, addr string) (*Client, error) {
	if url == "" {
		return nil, fmt.Errorf("URL ot the service must be specified")
	}

	if user == "" {
		return nil, fmt.Errorf("User name must be specified")
	}

	// Empty password should be allowed?
	if passwd == "" {
		return nil, fmt.Errorf("Password must be specified")
	}

	if reqCode == REGISTER && addr == "" {
		return nil, fmt.Errorf("IP Address must be specified when reqCode is REGISTER")
	}

	c := Client{
		url,
		user,
		md5_hex(passwd),
		domain,
		reqCode,
		addr,
		&http.Client{},
	}

	logdebug("gdip.New(): %s@%p", c, &c)
	return &c, nil
}

// Implements Stringer interface
func (c Client) String() string {
	return fmt.Sprintf("gdip.Client{URL: %s, User: %s, Password: XXX, DomainName: %s, reqCode: %d, Address: %s}",
		c.URL,
		c.User,
		c.DomainName,
		c.requestCode,
		c.Address)
}

func (c *Client) get(params string) (*response, error) {
	url := c.URL
	if params != "" {
		url = url + "?" + params
	}

	logdebug("gdip.Client@%p.get(%s)", c, url)
	resp, err := c.http.Get(url)
	if err != nil {
		logerr("gdip.Client@%p.get(): %v", c, err)
		return nil, err
	}
	defer resp.Body.Close()

	d := xml.NewDecoder(resp.Body)
	d.Strict = false
	d.AutoClose = xml.HTMLAutoClose
	d.Entity = xml.HTMLEntity

	var r response
	err = d.Decode(&r)
	if err != nil {
		logerr("gdip.Client@%p.get(%s): xml.Decoder.Decode() failed: %v", c, url, err)
		return nil, err
	}
	logdebug("gdip.Client@%p.get(): %v", c, r)
	return &r, nil
}

func (c *Client) getSalt() (salt, time, sign string, err error) {
	r, err := c.get("")
	if err != nil {
		return salt, time, sign, err
	}

	for _, m := range r.Meta {
		var s *string
		switch m.Name {
		case "salt":
			s = &salt
		case "time":
			s = &time
		case "sign":
			s = &sign
		}
		*s = m.Content
	}
	return
}

func (c *Client) requestUpdate(salt, time, sign string) (retc, addr string, err error) {
	params := ""
	params = params + fmt.Sprintf("salt=%s", salt)
	params = params + "&" + fmt.Sprintf("time=%s", time)
	params = params + "&" + fmt.Sprintf("sign=%s", sign)
	params = params + "&" + fmt.Sprintf("user=%s", c.User)
	params = params + "&" + fmt.Sprintf("pass=%s", md5_hex(c.Password+"."+salt))
	params = params + "&" + fmt.Sprintf("domn=%s", c.DomainName)
	params = params + "&" + fmt.Sprintf("reqc=%d", int(c.requestCode))
	if c.requestCode == REGISTER {
		params = params + "&" + fmt.Sprintf("addr=%s", c.Address)
	}
	r, err := c.get(params)
	if err != nil {
		return
	}

	for _, m := range r.Meta {
		var s *string
		switch m.Name {
		case "retc":
			s = &retc
		case "addr":
			s = &addr
		}
		*s = m.Content
	}
	return
}

// Updates the Client's domain name.  It returns update address if succeeded.
func (c *Client) Update() (addr string, err error) {
	salt, time, sign, err := c.getSalt()
	if err != nil {
		return
	}

	retc, addr, err := c.requestUpdate(salt, time, sign)
	if retc == "0" {
		if c.requestCode == 0 && addr == "" {
			addr = c.Address
		}
		loginfo("gdip.Client@%p.Update(): Successfully updated %s as %s", c, c.DomainName, addr)
	} else if retc == "1" {
		err = fmt.Errorf("Invalid login (or other problem)")
		logerr("gdip.Client@%p.Update(): %v", c, err)
	} else if retc == "2" {
		loginfo("gdip.Client@%p.Update(): Successfully went offline %s", c, c.DomainName)
	} else {
		err = fmt.Errorf("Unknown error")
		logerr("gdip.Client@%p.Update(): %v", c, err)
	}
	return
}
