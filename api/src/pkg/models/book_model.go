//go:generate goqueryset -in book_model.go
package models

import (
	"github.com/jinzhu/gorm"
	"time"
)

/*
HTTP Archive (HAR) format
https://w3c.github.io/web-performance/specs/HAR/Overview.html
*/

// gen:qs
type MizuEntry struct {
	gorm.Model
	// The Entry itself
	Entry Entry `json:"entry,omitempty"`
	//TODO: here we will add fields we need to query for
}



type Entry struct {
	Pageref string `json:"pageref,omitempty"`
	// Date and time stamp of the request start
	// (ISO 8601 YYYY-MM-DDThh:mm:ss.sTZD).
	StartedDateTime string `json:"startedDateTime"`
	// Total elapsed time of the request in milliseconds. This is the sum of all
	// timings available in the timings object (i.e. not including -1 values) .
	Time float32 `json:"time"`
	// Detailed info about the request.
	Request Request `json:"request"`
	// Detailed info about the response.
	Response Response `json:"response"`
	// Info about cache usage.
	Cache Cache `json:"cache"`
	// Detailed timing info about request/response round trip.
	PageTimings PageTimings `json:"pageTimings"`
	// optional (new in 1.2) IP address of the server that was connected
	// (result of DNS resolution).
	ServerIPAddress string `json:"serverIPAddress,omitempty"`
	// optional (new in 1.2) Unique ID of the parent TCP/IP connection, can be
	// the client port number. Note that a port number doesn't have to be unique
	// identifier in cases where the port is shared for more connections. If the
	// port isn't available for the application, any other unique connection ID
	// can be used instead (e.g. connection index). Leave out this field if the
	// application doesn't support this info.
	Connection string `json:"connection,omitempty"`
	// (new in 1.2) A comment provided by the user or the application.
	Comment string `json:"comment,omitempty"`
}

// Request contains detailed info about performed request.
type Request struct {
	// Request method (GET, POST, ...).
	Method string `json:"method"`
	// Absolute URL of the request (fragments are not included).
	URL string `json:"url"`
	// Request HTTP Version.
	HTTPVersion string `json:"httpVersion"`
	// List of cookie objects.
	Cookies []Cookie `json:"cookies"`
	// List of header objects.
	Headers []NVP `json:"headers"`
	// List of query parameter objects.
	QueryString []NVP `json:"queryString"`
	// Posted data.
	PostData PostData `json:"postData"`
	// Total number of bytes from the start of the HTTP request message until
	// (and including) the double CRLF before the body. Set to -1 if the info
	// is not available.
	HeaderSize int `json:"headerSize"`
	// Size of the request body (POST data payload) in bytes. Set to -1 if the
	// info is not available.
	BodySize int `json:"bodySize"`
	// (new in 1.2) A comment provided by the user or the application.
	Comment string `json:"comment"`
}

// Response contains detailed info about the response.
type Response struct {
	// Response status.
	Status int `json:"status"`
	// Response status description.
	StatusText string `json:"statusText"`
	// Response HTTP Version.
	HTTPVersion string `json:"httpVersion"`
	// List of cookie objects.
	Cookies []Cookie `json:"cookies"`
	// List of header objects.
	Headers []NVP `json:"headers"`
	// Details about the response body.
	Content Content `json:"content"`
	// Redirection target URL from the Location response header.
	RedirectURL string `json:"redirectURL"`
	// Total number of bytes from the start of the HTTP response message until
	// (and including) the double CRLF before the body. Set to -1 if the info is
	// not available.
	// The size of received response-headers is computed only from headers that
	// are really received from the server. Additional headers appended by the
	// browser are not included in this number, but they appear in the list of
	// header objects.
	HeadersSize int `json:"headersSize"`
	// Size of the received response body in bytes. Set to zero in case of
	// responses coming from the cache (304). Set to -1 if the info is not
	// available.
	BodySize int `json:"bodySize"`
	// optional (new in 1.2) A comment provided by the user or the application.
	Comment string `json:"comment,omitempty"`
}

// Cookie contains list of all cookies (used in <request> and <response> objects).
type Cookie struct {
	// The name of the cookie.
	Name string `json:"name"`
	// The cookie value.
	Value string `json:"value"`
	// optional The path pertaining to the cookie.
	Path string `json:"path,omitempty"`
	// optional The host of the cookie.
	Domain string `json:"domain,omitempty"`
	// optional Cookie expiration time.
	// (ISO 8601 YYYY-MM-DDThh:mm:ss.sTZD, e.g. 2009-07-24T19:20:30.123+02:00).
	Expires string `json:"expires,omitempty"`
	// optional Set to true if the cookie is HTTP only, false otherwise.
	HTTPOnly bool `json:"httpOnly,omitempty"`
	// optional (new in 1.2) True if the cookie was transmitted over ssl, false
	// otherwise.
	Secure bool `json:"secure,omitempty"`
	// optional (new in 1.2) A comment provided by the user or the application.
	Comment bool `json:"comment,omitempty"`
}

// NVP is simply a name/value pair with a comment
type NVP struct {
	Name    string `json:"name"`
	Value   string `json:"value"`
	Comment string `json:"comment,omitempty"`
}

// PostData describes posted data, if any (embedded in <request> object).
type PostData struct {
	//  Mime type of posted data.
	MimeType string `json:"mimeType"`
	//  List of posted parameters (in case of URL encoded parameters).
	Params []PostParam `json:"params"`
	//  Plain text posted data
	Text string `json:"text"`
	// optional (new in 1.2) A comment provided by the user or the
	// application.
	Comment string `json:"comment,omitempty"`
}

// PostParam is a list of posted parameters, if any (embedded in <postData> object).
type PostParam struct {
	// name of a posted parameter.
	Name string `json:"name"`
	// optional value of a posted parameter or content of a posted file.
	Value string `json:"value,omitempty"`
	// optional name of a posted file.
	FileName string `json:"fileName,omitempty"`
	// optional content type of a posted file.
	ContentType string `json:"contentType,omitempty"`
	// optional (new in 1.2) A comment provided by the user or the application.
	Comment string `json:"comment,omitempty"`
}

// Content describes details about response content (embedded in <response> object).
type Content struct {
	// Length of the returned content in bytes. Should be equal to
	// response.bodySize if there is no compression and bigger when the content
	// has been compressed.
	Size int `json:"size"`
	// optional Number of bytes saved. Leave out this field if the information
	// is not available.
	Compression int `json:"compression,omitempty"`
	// MIME type of the response text (value of the Content-Type response
	// header). The charset attribute of the MIME type is included (if
	// available).
	MimeType string `json:"mimeType"`
	// optional Response body sent from the server or loaded from the browser
	// cache. This field is populated with textual content only. The text field
	// is either HTTP decoded text or a encoded (e.g. "base64") representation of
	// the response body. Leave out this field if the information is not
	// available.
	Text string `json:"text,omitempty"`
	// optional (new in 1.2) Encoding used for response text field e.g
	// "base64". Leave out this field if the text field is HTTP decoded
	// (decompressed & unchunked), than trans-coded from its original character
	// set into UTF-8.
	Encoding string `json:"encoding,omitempty"`
	// optional (new in 1.2) A comment provided by the user or the application.
	Comment string `json:"comment,omitempty"`
}

// Cache contains info about a request coming from browser cache.
type Cache struct {
	// optional State of a cache entry before the request. Leave out this field
	// if the information is not available.
	BeforeRequest CacheObject `json:"beforeRequest,omitempty"`
	// optional State of a cache entry after the request. Leave out this field if
	// the information is not available.
	AfterRequest CacheObject `json:"afterRequest,omitempty"`
	// optional (new in 1.2) A comment provided by the user or the application.
	Comment string `json:"comment,omitempty"`
}

// CacheObject is used by both beforeRequest and afterRequest
type CacheObject struct {
	// optional - Expiration time of the cache entry.
	Expires string `json:"expires,omitempty"`
	// The last time the cache entry was opened.
	LastAccess string `json:"lastAccess"`
	// Etag
	ETag string `json:"eTag"`
	// The number of times the cache entry has been opened.
	HitCount int `json:"hitCount"`
	// optional (new in 1.2) A comment provided by the user or the application.
	Comment string `json:"comment,omitempty"`
}

// PageTimings describes various phases within request-response round trip.
// All times are specified in milliseconds.
type PageTimings struct {
	Blocked int `json:"blocked,omitempty"`
	// optional - Time spent in a queue waiting for a network connection. Use -1
	// if the timing does not apply to the current request.
	DNS int `json:"dns,omitempty"`
	// optional - DNS resolution time. The time required to resolve a host name.
	// Use -1 if the timing does not apply to the current request.
	Connect int `json:"connect,omitempty"`
	// optional - Time required to create TCP connection. Use -1 if the timing
	// does not apply to the current request.
	Send int `json:"send"`
	// Time required to send HTTP request to the server.
	Wait int `json:"wait"`
	// Waiting for a response from the server.
	Receive int `json:"receive"`
	// Time required to read entire response from the server (or cache).
	Ssl int `json:"ssl,omitempty"`
	// optional (new in 1.2) - Time required for SSL/TLS negotiation. If this
	// field is defined then the time is also included in the connect field (to
	// ensure backward compatibility with HAR 1.1). Use -1 if the timing does not
	// apply to the current request.
	Comment string `json:"comment,omitempty"`
	// optional (new in 1.2) - A comment provided by the user or the application.
}

// TestResult contains results for an individual HTTP request
type TestResult struct {
	URL       string    `json:"url"`
	Status    int       `json:"status"` // 200, 500, etc.
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
	Latency   int       `json:"latency"` // milliseconds
	Method    string    `json:"method"`
	HarFile   string    `json:"harfile"`
}
//package models
//
//import (
//	"database/sql/driver"
//	"encoding/json"
//	"errors"
//	"time"
//
//	"github.com/google/uuid"
//)
//
//// Book struct to describe book object.
//type Book struct {
//	ID         uuid.UUID `db:"id" json:"id" validate:"required,uuid"`
//	CreatedAt  time.Time `db:"created_at" json:"created_at"`
//	UpdatedAt  time.Time `db:"updated_at" json:"updated_at"`
//	UserID     uuid.UUID `db:"user_id" json:"user_id" validate:"required,uuid"`
//	Title      string    `db:"title" json:"title" validate:"required,lte=255"`
//	Author     string    `db:"author" json:"author" validate:"required,lte=255"`
//	BookStatus int       `db:"book_status" json:"book_status" validate:"required,len=1"`
//	BookAttrs  BookAttrs `db:"book_attrs" json:"book_attrs" validate:"required,dive"`
//}
//
//// BookAttrs struct to describe book attributes.
//type BookAttrs struct {
//	Picture     string `json:"picture"`
//	Description string `json:"description"`
//	Rating      int    `json:"rating" validate:"min=1,max=10"`
//}
//
//// Value make the BookAttrs struct implement the driver.Valuer interface.
//// This method simply returns the JSON-encoded representation of the struct.
//func (b BookAttrs) Value() (driver.Value, error) {
//	return json.Marshal(b)
//}
//
//// Scan make the BookAttrs struct implement the sql.Scanner interface.
//// This method simply decodes a JSON-encoded value into the struct fields.
//func (b *BookAttrs) Scan(value interface{}) error {
//	j, ok := value.([]byte)
//	if !ok {
//		return errors.New("type assertion to []byte failed")
//	}
//
//	return json.Unmarshal(j, &b)
//}
