package lowhttp

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/yaklang/yaklang/common/utils"
)

func UrlToGetRequestPacket(u string, originRequest []byte, originRequestHttps bool, cookies ...*http.Cookie) []byte {
	return UrlToRequestPacket("GET", u, originRequest, originRequestHttps, cookies...)
}

func UrlToRequestPacket(method string, u string, originRequest []byte, originRequestHttps bool, cookies ...*http.Cookie) []byte {
	if originRequest == nil {
		reqIns, err := http.NewRequest(method, u, http.NoBody)
		if err != nil {
			return nil
		}
		reqIns.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; rv:68.0) Gecko/20100101 Firefox/68.0")
		for _, cookie := range cookies {
			reqIns.AddCookie(cookie)
		}

		reqRaw, err := utils.HttpDumpWithBody(reqIns, true)
		if err != nil {
			return nil
		}
		return FixHTTPRequestOut(reqRaw)
	}

	originUrl, err := ExtractURLFromHTTPRequestRaw(originRequest, originRequestHttps)
	if err != nil {
		return nil
	}

	originHost, _, _ := utils.ParseStringToHostPort(originUrl.String())
	currentHost, _, _ := utils.ParseStringToHostPort(u)
	if (strings.Contains(currentHost, originHost) || strings.Contains(originHost, currentHost)) && originHost != "" {
		req, err := ParseBytesToHttpRequest(originRequest)
		if err != nil {
			return nil
		}
		req.URL, _ = url.Parse(u)
		if req.URL == nil {
			return nil
		}
		req.Host = req.URL.Host
		req.RequestURI = req.URL.RequestURI()
		for _, cookie := range cookies {
			req.AddCookie(cookie)
		}
		raw, err := utils.HttpDumpWithBody(req, true)
		if err != nil {
			return nil
		}
		return raw
	} else {
		reqIns, err := http.NewRequest(method, u, http.NoBody)
		if err != nil {
			return nil
		}
		reqIns.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; rv:68.0) Gecko/20100101 Firefox/68.0")
		for _, cookie := range cookies {
			reqIns.AddCookie(cookie)
		}
		reqRaw, err := utils.HttpDumpWithBody(reqIns, true)
		if err != nil {
			return nil
		}
		return FixHTTPRequestOut(reqRaw)
	}
}

func UrlToHTTPRequest(text string) ([]byte, error) {
	var r *http.Request
	if !(strings.HasPrefix(text, "http://") || strings.HasPrefix(text, "https://")) {
		text = "http://" + text
	}
	u, err := url.Parse(text)
	if err != nil {
		return nil, err
	}
	r, err = http.NewRequest("GET", text, http.NoBody)
	if err != nil {
		return nil, err
	}

	if u.RawPath != "" {
		r.RequestURI = u.RawPath
	} else {
		r.RequestURI = u.Path
	}

	if u.RawQuery != "" {
		r.RequestURI += "?" + u.RawQuery
	}

	if u.Fragment != "" {
		if u.RawFragment != "" {
			r.RequestURI += "#" + u.RawFragment
		} else {
			r.RequestURI += "#" + u.Fragment
		}
	}

	if strings.HasSuffix(text, "#") && u.Fragment == "" {
		r.RequestURI += "#"
	}

	raw, err := utils.HttpDumpWithBody(r, true)
	if err != nil {
		return nil, err
	}
	raw = FixHTTPRequestOut(raw)
	return raw, nil
}
