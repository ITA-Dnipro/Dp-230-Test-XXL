package optimization

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"text/template"

	"test/internal/model"
)

// GenerateNewRequest is make http.Cilent
func GenerateNewRequest(url, body string, options model.Options) *http.Request {
	req, _ := http.NewRequest("GET", url, nil)
	// Add the Accept header like browsers do.
	req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")

	if options.Data != "" {
		d := []byte(body)
		req, _ = http.NewRequest("POST", url, bytes.NewBuffer(d))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}

	if len(options.Header) > 0 {
		for _, v := range options.Header {
			h := strings.Split(v, ": ")
			if len(h) > 1 {
				req.Header.Add(h[0], h[1])
			}
		}
	}
	if options.Cookie != "" {
		req.Header.Add("Cookie", options.Cookie)
	}
	if options.UserAgent != "" {
		req.Header.Add("User-Agent", options.UserAgent)
	} else {
		req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:75.0) Gecko/20100101 Firefox/75.0")
	}
	if options.Method != "" {
		req.Method = options.Method
	}
	return req
}

// MakeRequestQuery is generate http query with custom parameters
func MakeRequestQuery(target, param, payload, ptype string, pAction string, pEncode string, options model.Options) (*http.Request, map[string]string) {

	tempMap := make(map[string]string)
	tempMap["type"] = ptype
	tempMap["action"] = pAction
	tempMap["encode"] = pEncode
	tempMap["payload"] = payload
	tempMap["param"] = param

	u, _ := url.Parse(target)

	var tempParam string
	var tempParamBody string
	if options.Data == "" {
		tempParam = u.RawQuery // ---> GET
	} else {
		tempParam = u.RawQuery       // ---> GET
		tempParamBody = options.Data // ---> POST
	}

	paramList, _ := url.ParseQuery(tempParam)
	paramListBody, _ := url.ParseQuery(tempParamBody)

	//What we should do to the payload?
	switch tempMap["encode"] {
	case "urlEncode":
		payload = UrlEncode(payload)
		break

	case "urlDoubleEncode":
		payload = (UrlEncode(payload))
		break

	case "htmlEncode":
		payload = template.HTMLEscapeString(payload)
		break

	default:
		break
	}

	// We first check if the parameter exist and then "append or replace" the value
	if strings.Contains(ptype, "FORM") {
		if val, ok := paramListBody[tempMap["param"]]; ok {
			if tempMap["action"] == "toAppend" {
				paramListBody[tempMap["param"]][0] = val[0] + payload
			} else { //toReplace lies here
				paramListBody[tempMap["param"]][0] = payload
			}
		} else {
			//if the parameter doesn't exist, is added.
			paramListBody.Add(tempMap["param"], payload)
		}

		var rst *http.Request
		rst = GenerateNewRequest(u.String(), paramListBody.Encode(), options)
		return rst, tempMap
	} else {
		// PA-URL
		if val, ok := paramList[tempMap["param"]]; ok {
			if tempMap["action"] == "toAppend" {
				paramList[tempMap["param"]][0] = val[0] + payload
			} else { //toReplace lies here
				paramList[tempMap["param"]][0] = payload
			}
		} else {
			//if the parameter doesn't exist, is added.
			paramList.Add(tempMap["param"], payload)
		}

		var rst *http.Request
		u.RawQuery = paramList.Encode()
		rst = GenerateNewRequest(u.String(), paramListBody.Encode(), options)
		return rst, tempMap
	}
}

// TODO: find out
// UrlEncode is custom url encoder for double url encoding
func UrlEncode(s string) (result string) {
	for _, c := range s {
		if c <= 0x7f { // single byte
			result += fmt.Sprintf("%%%X", c)
		} else if c > 0x1fffff { // quaternary byte
			result += fmt.Sprintf("%%%X%%%X%%%X%%%X",
				0xf0+((c&0x1c0000)>>18),
				0x80+((c&0x3f000)>>12),
				0x80+((c&0xfc0)>>6),
				0x80+(c&0x3f),
			)
		} else if c > 0x7ff { // triple byte
			result += fmt.Sprintf("%%%X%%%X%%%X",
				0xe0+((c&0xf000)>>12),
				0x80+((c&0xfc0)>>6),
				0x80+(c&0x3f),
			)
		} else { // double byte
			result += fmt.Sprintf("%%%X%%%X",
				0xc0+((c&0x7c0)>>6),
				0x80+(c&0x3f),
			)
		}
	}

	return result
}

// Optimization is remove payload included badchar
func Optimization(payload string, badchars []string) bool {
	for _, v := range badchars {
		if strings.Contains(payload, v) {
			return false
		}
	}
	return true
}
