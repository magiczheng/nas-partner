package util

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

// ===== Aliyun Signer =====

var signMethodMap = map[string]func() hash.Hash{
	"HMAC-SHA1":   sha1New,
	"HMAC-SHA256": sha256.New,
	"HMAC-MD5":    md5.New,
}

func sha1New() hash.Hash { return sha1.New() }

var encTilde = "%7E"
var encBlank = []byte("%20")
var tilde = []byte("~")

func AliyunSigner(appKey, appKeySecret string, vals *url.Values, httpMethod, apiVersion string) {
	vals.Set("Format", "JSON")
	vals.Set("Version", apiVersion)
	vals.Set("AccessKeyId", appKey)
	vals.Set("SignatureMethod", "HMAC-SHA1")
	vals.Set("Timestamp", time.Now().UTC().Format("2006-01-02T15:04:05Z"))
	vals.Set("SignatureVersion", "1.0")
	vals.Set("SignatureNonce", strconv.FormatInt(time.Now().UnixNano(), 10))

	signature := HmacSignToB64("HMAC-SHA1", httpMethod, appKeySecret, *vals)
	vals.Set("Signature", signature)
}

func HmacSign(signMethod string, httpMethod string, appKeySecret string, vals url.Values) []byte {
	key := []byte(appKeySecret + "&")
	var h hash.Hash
	if method, ok := signMethodMap[signMethod]; ok {
		h = hmac.New(method, key)
	} else {
		h = hmac.New(sha1New, key)
	}
	makeDataToSign(h, httpMethod, vals)
	return h.Sum(nil)
}

func HmacSignToB64(signMethod string, httpMethod string, appKeySecret string, vals url.Values) string {
	return base64.StdEncoding.EncodeToString(HmacSign(signMethod, httpMethod, appKeySecret, vals))
}

type strToEnc struct {
	s string
	e bool
}

func makeDataToSign(w io.Writer, httpMethod string, vals url.Values) {
	in := make(chan *strToEnc)
	go func() {
		in <- &strToEnc{s: httpMethod}
		in <- &strToEnc{s: "&"}
		in <- &strToEnc{s: "/", e: true}
		in <- &strToEnc{s: "&"}
		in <- &strToEnc{s: vals.Encode(), e: true}
		close(in)
	}()
	specialUrlEncode(in, w)
}

func specialUrlEncode(in <-chan *strToEnc, w io.Writer) {
	for s := range in {
		if !s.e {
			io.WriteString(w, s.s)
			continue
		}
		l := len(s.s)
		for i := 0; i < l; {
			ch := s.s[i]
			switch ch {
			case '%':
				if encTilde == s.s[i:i+3] {
					w.Write(tilde)
					i += 3
					continue
				}
				fallthrough
			case '*', '/', '&', '=':
				fmt.Fprintf(w, "%%%02X", ch)
			case '+':
				w.Write(encBlank)
			default:
				fmt.Fprintf(w, "%c", ch)
			}
			i += 1
		}
	}
}

// ===== TencentCloud Signer =====

func sha256hex(s string) string {
	b := sha256.Sum256([]byte(s))
	return hex.EncodeToString(b[:])
}

func tencentCloudHmacsha256(s, key string) string {
	hashed := hmac.New(sha256.New, []byte(key))
	hashed.Write([]byte(s))
	return string(hashed.Sum(nil))
}

const (
	DnsPod  = "dnspod"
	EdgeOne = "teo"
)

func TencentCloudSigner(secretId, secretKey string, r *http.Request, action, payload, service string) {
	algorithm := "TC3-HMAC-SHA256"
	host := service + ".tencentcloudapi.com"
	timestamp := time.Now().Unix()
	timestampStr := strconv.FormatInt(timestamp, 10)

	canonicalHeaders := "content-type:application/json\nhost:" + host + "\nx-tc-action:" + strings.ToLower(action) + "\n"
	signedHeaders := "content-type;host;x-tc-action"
	hashedRequestPayload := sha256hex(payload)
	canonicalRequest := "POST\n/\n\n" + canonicalHeaders + "\n" + signedHeaders + "\n" + hashedRequestPayload

	date := time.Unix(timestamp, 0).UTC().Format("2006-01-02")
	credentialScope := date + "/" + service + "/tc3_request"
	hashedCanonicalRequest := sha256hex(canonicalRequest)
	string2sign := algorithm + "\n" + timestampStr + "\n" + credentialScope + "\n" + hashedCanonicalRequest

	secretDate := tencentCloudHmacsha256(date, "TC3"+secretKey)
	secretService := tencentCloudHmacsha256(service, secretDate)
	secretSigning := tencentCloudHmacsha256("tc3_request", secretService)
	signature := hex.EncodeToString([]byte(tencentCloudHmacsha256(string2sign, secretSigning)))

	authorization := algorithm + " Credential=" + secretId + "/" + credentialScope + ", SignedHeaders=" + signedHeaders + ", Signature=" + signature

	r.Header.Add("Authorization", authorization)
	r.Header.Set("Host", host)
	r.Header.Set("X-TC-Action", action)
	r.Header.Add("X-TC-Timestamp", timestampStr)
}

// ===== HuaweiCloud Signer =====

const (
	BasicDateFormat     = "20060102T150405Z"
	Algorithm           = "SDK-HMAC-SHA256"
	HeaderXDate         = "X-Sdk-Date"
	HeaderHost          = "host"
	HeaderAuthorization = "Authorization"
	HeaderContentSha256 = "X-Sdk-Content-Sha256"
)

type Signer struct {
	Key    string
	Secret string
}

func (s *Signer) Sign(r *http.Request) error {
	var t time.Time
	var err error
	var dt string
	if dt = r.Header.Get(HeaderXDate); dt != "" {
		t, err = time.Parse(BasicDateFormat, dt)
	}
	if err != nil || dt == "" {
		t = time.Now()
		r.Header.Set(HeaderXDate, t.UTC().Format(BasicDateFormat))
	}
	signedHeaders := signedHeaders(r)
	canonicalRequest, err := canonicalRequest(r, signedHeaders)
	if err != nil {
		return err
	}
	stringToSign, err := stringToSign(canonicalRequest, t)
	if err != nil {
		return err
	}
	signature, err := signStringToSign(stringToSign, []byte(s.Secret))
	if err != nil {
		return err
	}
	authValue := authHeaderValue(signature, s.Key, signedHeaders)
	r.Header.Set(HeaderAuthorization, authValue)
	return nil
}

func hmacsha256(key []byte, data string) ([]byte, error) {
	h := hmac.New(sha256.New, key)
	if _, err := h.Write([]byte(data)); err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}

func canonicalRequest(r *http.Request, signedHeaders []string) (string, error) {
	var hexencode string
	if hex := r.Header.Get(HeaderContentSha256); hex != "" {
		hexencode = hex
	} else {
		data, err := requestPayload(r)
		if err != nil {
			return "", err
		}
		hexencode, err = hexEncodeSHA256Hash(data)
		if err != nil {
			return "", err
		}
	}
	return fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s", r.Method, canonicalURI(r), canonicalQueryString(r), canonicalHeaders(r, signedHeaders), strings.Join(signedHeaders, ";"), hexencode), nil
}

func canonicalURI(r *http.Request) string {
	patterns := strings.Split(r.URL.Path, "/")
	var uri []string
	for _, v := range patterns {
		uri = append(uri, percentEncode(v))
	}
	urlpath := strings.Join(uri, "/")
	if len(urlpath) == 0 || urlpath[len(urlpath)-1] != '/' {
		urlpath = urlpath + "/"
	}
	return urlpath
}

func canonicalQueryString(r *http.Request) string {
	var keys []string
	query := r.URL.Query()
	for key := range query {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var a []string
	for _, key := range keys {
		k := percentEncode(key)
		sort.Strings(query[key])
		for _, v := range query[key] {
			kv := fmt.Sprintf("%s=%s", k, percentEncode(v))
			a = append(a, kv)
		}
	}
	queryStr := strings.Join(a, "&")
	r.URL.RawQuery = queryStr
	return queryStr
}

func canonicalHeaders(r *http.Request, signerHeaders []string) string {
	var a []string
	header := make(map[string][]string)
	for k, v := range r.Header {
		header[strings.ToLower(k)] = v
	}
	for _, key := range signerHeaders {
		value := header[key]
		if strings.EqualFold(key, HeaderHost) {
			value = []string{r.Host}
		}
		sort.Strings(value)
		for _, v := range value {
			a = append(a, key+":"+strings.TrimSpace(v))
		}
	}
	return fmt.Sprintf("%s\n", strings.Join(a, "\n"))
}

func signedHeaders(r *http.Request) []string {
	var a []string
	for key := range r.Header {
		a = append(a, strings.ToLower(key))
	}
	sort.Strings(a)
	return a
}

func requestPayload(r *http.Request) ([]byte, error) {
	if r.Body == nil {
		return []byte(""), nil
	}
	b, err := io.ReadAll(r.Body)
	if err != nil {
		return []byte(""), err
	}
	r.Body = io.NopCloser(bytes.NewBuffer(b))
	return b, err
}

func stringToSign(canonicalRequest string, t time.Time) (string, error) {
	hash := sha256.New()
	_, err := hash.Write([]byte(canonicalRequest))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s\n%s\n%x", Algorithm, t.UTC().Format(BasicDateFormat), hash.Sum(nil)), nil
}

func signStringToSign(stringToSign string, signingKey []byte) (string, error) {
	hm, err := hmacsha256(signingKey, stringToSign)
	return fmt.Sprintf("%x", hm), err
}

func hexEncodeSHA256Hash(body []byte) (string, error) {
	hash := sha256.New()
	if body == nil {
		body = []byte("")
	}
	_, err := hash.Write(body)
	return fmt.Sprintf("%x", hash.Sum(nil)), err
}

func authHeaderValue(signature, accessKey string, signedHeaders []string) string {
	return fmt.Sprintf("%s Access=%s, SignedHeaders=%s, Signature=%s", Algorithm, accessKey, strings.Join(signedHeaders, ";"), signature)
}

func percentEncode(s string) string {
	hexCount := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		if shouldEscape(c) {
			hexCount++
		}
	}
	if hexCount == 0 {
		return s
	}
	t := make([]byte, len(s)+2*hexCount)
	j := 0
	for i := 0; i < len(s); i++ {
		switch c := s[i]; {
		case shouldEscape(c):
			t[j] = '%'
			t[j+1] = "0123456789ABCDEF"[c>>4]
			t[j+2] = "0123456789ABCDEF"[c&15]
			j += 3
		default:
			t[j] = s[i]
			j++
		}
	}
	return string(t)
}

func shouldEscape(c byte) bool {
	if 'A' <= c && c <= 'Z' || 'a' <= c && c <= 'z' || '0' <= c && c <= '9' || c == '_' || c == '-' || c == '~' || c == '.' {
		return false
	}
	return true
}
