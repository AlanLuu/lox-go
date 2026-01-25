package ast

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/AlanLuu/lox/list"
	"github.com/AlanLuu/lox/loxerror"
	"github.com/AlanLuu/lox/token"
)

type LoxHTTPResponse struct {
	res        *http.Response
	elapsed    float64
	url        string
	isClosed   bool
	properties map[string]any
}

func NewLoxHTTPResponse(res *http.Response) *LoxHTTPResponse {
	return &LoxHTTPResponse{
		res:        res,
		isClosed:   false,
		properties: make(map[string]any),
	}
}

func LoxHTTPResHelper(url string, reqFunc func() (*http.Response, error)) (*LoxHTTPResponse, error) {
	startTime := float64(time.Now().UnixMilli()) / 1000
	res, err := reqFunc()
	endTime := float64(time.Now().UnixMilli()) / 1000
	if err != nil {
		return nil, err
	}
	httpRes := NewLoxHTTPResponse(res)
	httpRes.elapsed = endTime - startTime
	httpRes.url = url
	return httpRes, nil
}

func LoxHTTPGetUrl(url string) (*LoxHTTPResponse, error) {
	return LoxHTTPResHelper(url, func() (*http.Response, error) {
		return http.Get(url)
	})
}

func LoxHTTPHeadUrl(url string) (*LoxHTTPResponse, error) {
	return LoxHTTPResHelper(url, func() (*http.Response, error) {
		return http.Head(url)
	})
}

func LoxHTTPPostUrl(url string) (*LoxHTTPResponse, error) {
	return LoxHTTPResHelper(url, func() (*http.Response, error) {
		return http.Post(url, "", nil)
	})
}

func LoxHTTPPostForm(urlStr string, form url.Values) (*LoxHTTPResponse, error) {
	return LoxHTTPResHelper(urlStr, func() (*http.Response, error) {
		if len(form) == 0 {
			return http.Post(urlStr, "", nil)
		}
		return http.PostForm(urlStr, form)
	})
}

func LoxHTTPPostText(url string, body string) (*LoxHTTPResponse, error) {
	return LoxHTTPResHelper(url, func() (*http.Response, error) {
		if len(body) == 0 {
			return http.Post(url, "", nil)
		}
		return http.Post(url, "text/plain", strings.NewReader(body))
	})
}

func LoxHTTPPostJSONText(url string, body string) (*LoxHTTPResponse, error) {
	return LoxHTTPResHelper(url, func() (*http.Response, error) {
		if len(body) == 0 {
			return http.Post(url, "", nil)
		}
		return http.Post(url, "application/json", strings.NewReader(body))
	})
}

func LoxHTTPSendRequest(req *http.Request) (*LoxHTTPResponse, error) {
	return LoxHTTPResHelper(req.URL.String(), func() (*http.Response, error) {
		return http.DefaultClient.Do(req)
	})
}

func LoxHTTPSendRequestClient(client *http.Client, req *http.Request) (*LoxHTTPResponse, error) {
	return LoxHTTPResHelper(req.URL.String(), func() (*http.Response, error) {
		return client.Do(req)
	})
}

func (l *LoxHTTPResponse) close() {
	if !l.isClosed {
		l.res.Body.Close()
		l.isClosed = true
	}
}

func (l *LoxHTTPResponse) Get(name *token.Token) (any, error) {
	lexemeName := name.Lexeme
	if field, ok := l.properties[lexemeName]; ok {
		return field, nil
	}
	responseField := func(field any) (any, error) {
		if _, ok := l.properties[lexemeName]; !ok {
			l.properties[lexemeName] = field
		}
		return field, nil
	}
	responseFunc := func(arity int, method func(*Interpreter, list.List[any]) (any, error)) (*struct{ ProtoLoxCallable }, error) {
		s := &struct{ ProtoLoxCallable }{}
		s.arityMethod = func() int { return arity }
		s.callMethod = method
		s.stringMethod = func() string {
			return fmt.Sprintf("<native http response fn %v at %p>", lexemeName, s)
		}
		if _, ok := l.properties[lexemeName]; !ok {
			l.properties[lexemeName] = s
		}
		return s, nil
	}
	switch lexemeName {
	case "close":
		return responseFunc(0, func(_ *Interpreter, _ list.List[any]) (any, error) {
			l.close()
			return nil, nil
		})
	case "cookies":
		cookies := l.res.Cookies()
		cookiesList := list.NewListCap[any](int64(len(cookies)))
		for _, cookie := range cookies {
			if cookie == nil {
				continue
			}
			cookiesList.Add(NewLoxHTTPCookie(cookie))
		}
		return responseField(NewLoxList(cookiesList))
	case "elapsed":
		return responseField(l.elapsed)
	case "headers":
		dict := EmptyLoxDict()
		for key, value := range l.res.Header {
			valuesList := list.NewListCap[any](int64(len(value)))
			for _, str := range value {
				valuesList.Add(NewLoxStringQuote(str))
			}
			dict.setKeyValue(NewLoxString(key, '\''), NewLoxList(valuesList))
		}
		return responseField(dict)
	case "raw":
		bytes, err := io.ReadAll(l.res.Body)
		if err != nil {
			return nil, loxerror.RuntimeError(name, err.Error())
		}
		if _, ok := l.properties["text"]; !ok {
			l.properties["text"] = NewLoxStringQuote(string(bytes))
		}
		buffer := EmptyLoxBufferCap(int64(len(bytes)))
		for _, element := range bytes {
			addErr := buffer.add(int64(element))
			if addErr != nil {
				return nil, loxerror.RuntimeError(name, addErr.Error())
			}
		}
		return responseField(buffer)
	case "status":
		return responseField(int64(l.res.StatusCode))
	case "text":
		bytes, err := io.ReadAll(l.res.Body)
		if err != nil {
			return nil, loxerror.RuntimeError(name, err.Error())
		}
		if _, ok := l.properties["raw"]; !ok {
			buffer := EmptyLoxBufferCap(int64(len(bytes)))
			for _, element := range bytes {
				addErr := buffer.add(int64(element))
				if addErr != nil {
					return nil, loxerror.RuntimeError(name, addErr.Error())
				}
			}
			l.properties["raw"] = buffer
		}
		return responseField(NewLoxStringQuote(string(bytes)))
	case "url":
		return responseField(NewLoxStringQuote(l.url))
	}
	return nil, loxerror.RuntimeError(name, "HTTP responses have no property called '"+lexemeName+"'.")
}

func (l *LoxHTTPResponse) String() string {
	return fmt.Sprintf("<http response [%v] at %p>", l.res.StatusCode, l)
}

func (l *LoxHTTPResponse) Type() string {
	return "http response"
}
