package api

// copied from martian/har and adapted

import (
	"github.com/google/martian/messageview"
	"github.com/google/martian/proxyutil"
	har "github.com/mrichman/hargo"
	"github.com/up9inc/mizu/shared/logger"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"time"
)

// type aliases to minimize copied code changes

type Header = har.NVP
type QueryString = har.NVP
type Response = har.Response
type Request = har.Request
type Content = har.Content
type Cookie = har.Cookie
type PostData = har.PostData
type Param = har.PostParam

// NewRequest constructs and returns a Request from req. If withBody is true,
// req.Body is read to EOF and replaced with a copy in a bytes.Buffer. An error
// is returned (and req.Body may be in an intermediate state) if an error is
// returned from req.Body.Read.
func NewRequest(req *http.Request, withBody bool) (*Request, error) {
	r := &Request{
		Method:      req.Method,
		URL:         req.URL.String(),
		HTTPVersion: req.Proto,
		HeaderSize:  -1,
		BodySize:    int(req.ContentLength),
		QueryString: []QueryString{},
		Headers:     headers(proxyutil.RequestHeader(req).Map()),
		Cookies:     cookies(req.Cookies()),
	}

	for n, vs := range req.URL.Query() {
		for _, v := range vs {
			r.QueryString = append(r.QueryString, QueryString{
				Name:  n,
				Value: v,
			})
		}
	}

	pd, err := postData(req, withBody)
	if err != nil {
		return nil, err
	}
	r.PostData = *pd

	return r, nil
}

// NewResponse constructs and returns a Response from resp. If withBody is true,
// resp.Body is read to EOF and replaced with a copy in a bytes.Buffer. An error
// is returned (and resp.Body may be in an intermediate state) if an error is
// returned from resp.Body.Read.
func NewResponse(res *http.Response, withBody bool) (*Response, error) {
	r := &Response{
		HTTPVersion: res.Proto,
		Status:      res.StatusCode,
		StatusText:  http.StatusText(res.StatusCode),
		HeadersSize: -1,
		BodySize:    int(res.ContentLength),
		Headers:     headers(proxyutil.ResponseHeader(res).Map()),
		Cookies:     cookies(res.Cookies()),
	}

	if res.StatusCode >= 300 && res.StatusCode < 400 {
		r.RedirectURL = res.Header.Get("Location")
	}

	r.Content = Content{
		Encoding: "base64",
		MimeType: res.Header.Get("Content-Type"),
	}

	if withBody {
		mv := messageview.New()
		if err := mv.SnapshotResponse(res); err != nil {
			return nil, err
		}

		br, err := mv.BodyReader(messageview.Decode())
		if err != nil {
			return nil, err
		}

		body, err := ioutil.ReadAll(br)
		if err != nil {
			return nil, err
		}

		r.Content.Text = string(body)
		r.Content.Size = len(body)
	}
	return r, nil
}

func cookies(cs []*http.Cookie) []Cookie {
	hcs := make([]Cookie, 0, len(cs))

	for _, c := range cs {
		var expires string
		if !c.Expires.IsZero() {
			expires = c.Expires.Format(time.RFC3339)
		}

		hcs = append(hcs, Cookie{
			Name:     c.Name,
			Value:    c.Value,
			Path:     c.Path,
			Domain:   c.Domain,
			HTTPOnly: c.HttpOnly,
			Secure:   c.Secure,
			Expires:  expires,
		})
	}

	return hcs
}

func headers(hs http.Header) []Header {
	hhs := make([]Header, 0, len(hs))

	for n, vs := range hs {
		for _, v := range vs {
			hhs = append(hhs, Header{
				Name:  n,
				Value: v,
			})
		}
	}

	return hhs
}

func postData(req *http.Request, logBody bool) (*PostData, error) {
	// If the request has no body (no Content-Length and Transfer-Encoding isn't
	// chunked), skip the post data.
	if req.ContentLength <= 0 && len(req.TransferEncoding) == 0 {
		return nil, nil
	}

	ct := req.Header.Get("Content-Type")
	mt, ps, err := mime.ParseMediaType(ct)
	if err != nil {
		logger.Log.Errorf("har: cannot parse Content-Type header %q: %v", ct, err)
		mt = ct
	}

	pd := &PostData{
		MimeType: mt,
		Params:   []Param{},
	}

	if !logBody {
		return pd, nil
	}

	mv := messageview.New()
	if err := mv.SnapshotRequest(req); err != nil {
		return nil, err
	}

	br, err := mv.BodyReader()
	if err != nil {
		return nil, err
	}

	switch mt {
	case "multipart/form-data":
		mpr := multipart.NewReader(br, ps["boundary"])

		for {
			p, err := mpr.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, err
			}
			defer p.Close()

			body, err := ioutil.ReadAll(p)
			if err != nil {
				return nil, err
			}

			pd.Params = append(pd.Params, Param{
				Name:        p.FormName(),
				FileName:    p.FileName(),
				ContentType: p.Header.Get("Content-Type"),
				Value:       string(body),
			})
		}
	case "application/x-www-form-urlencoded":
		body, err := ioutil.ReadAll(br)
		if err != nil {
			return nil, err
		}

		vs, err := url.ParseQuery(string(body))
		if err != nil {
			return nil, err
		}

		for n, vs := range vs {
			for _, v := range vs {
				pd.Params = append(pd.Params, Param{
					Name:  n,
					Value: v,
				})
			}
		}
	default:
		body, err := ioutil.ReadAll(br)
		if err != nil {
			return nil, err
		}

		pd.Text = string(body)
	}

	return pd, nil
}
