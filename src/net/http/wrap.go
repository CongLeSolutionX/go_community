package http

import (
	"fmt"
	"io"

	"golang.org/x/net/http/httpguts"
)

type rawRoundTripper struct {
	t *Transport
	r httpguts.RoundTripper
}

func (r *rawRoundTripper) RoundTrip(req *Request) (*Response, error) {
	resp := &Response{}
	getBody := req.GetBody
	var replacedBody io.ReadCloser
	if getBody != nil {
		getBody = func() (io.ReadCloser, error) {
			b, err := req.GetBody()
			if err == nil {
				replacedBody = b
			}
			return b, err
		}
	}
	rr, err := r.r.RoundTrip(&httpguts.ClientRequest{
		Context:               req.Context(),
		Method:                req.Method,
		URL:                   req.URL,
		Header:                httpguts.Header(req.Header),
		Trailer:               httpguts.Header(req.Trailer),
		ResponseTrailer:       (*httpguts.Header)(&resp.Trailer),
		Body:                  req.Body,
		GetBody:               req.GetBody,
		ContentLength:         req.ContentLength,
		Close:                 req.Close || r.t.DisableKeepAlives,
		Cancel:                req.Cancel,
		Host:                  req.Host,
		DisableCompression:    r.t.DisableCompression,
		ResponseHeaderTimeout: r.t.ResponseHeaderTimeout,
		ExpectContinueTimeout: r.t.ExpectContinueTimeout,
	})
	if err != nil {
		return nil, err
	}
	resp.Status = fmt.Sprintf("%v %v", rr.StatusCode, StatusText(rr.StatusCode))
	resp.StatusCode = rr.StatusCode
	resp.Proto = rr.Proto
	resp.ProtoMajor = rr.ProtoMajor
	resp.ProtoMinor = rr.ProtoMinor
	resp.Header = Header(rr.Header)
	resp.Body = rr.Body
	resp.ContentLength = rr.ContentLength
	resp.Uncompressed = rr.Uncompressed
	resp.TLS = rr.TLS
	if replacedBody == nil {
		resp.Request = req
	} else {
		resp.Request = new(Request)
		*resp.Request = *req
		resp.Request.Body = replacedBody
	}
	return resp, nil
}

type rawHandler struct {
	Handler
}

func (h rawHandler) ServeHTTP(w httpguts.ResponseWriter, req *httpguts.ServerRequest) {
	h.Handler.ServeHTTP(rawResponseWriter{w}, &Request{
		ctx:           req.Context,
		Proto:         req.Proto,
		ProtoMajor:    req.ProtoMajor,
		ProtoMinor:    req.ProtoMinor,
		Method:        req.Method,
		Host:          req.Host,
		URL:           req.URL,
		Header:        Header(req.Header),
		Trailer:       Header(req.Trailer),
		Body:          req.Body,
		ContentLength: req.ContentLength,
		RemoteAddr:    req.RemoteAddr,
		RequestURI:    req.RequestURI,
		TLS:           req.TLS,
	})
}

type rawResponseWriter struct {
	httpguts.ResponseWriter
}

func (w rawResponseWriter) Header() Header {
	return Header(w.ResponseWriter.Header())
}
