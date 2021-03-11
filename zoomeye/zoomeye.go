package zoomeye

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	loginAPI    = "https://api.zoomeye.org/user/login"
	userinfoAPI = "https://api.zoomeye.org/resources-info"
	searchAPI   = "https://api.zoomeye.org/%s/search"
	historyAPI  = "https://api.zoomeye.org/both/search?history=true&ip=%s"
)

var httpCli = &http.Client{
	Timeout: 30 * time.Second,
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	},
}

var defaultFacets = map[string]string{
	"host": "app,device,service,os,port,country,city",
	"web":  "webapp,component,framework,frontend,server,waf,os,country,city",
}

// ZoomEye represents SDK for using
type ZoomEye struct {
	apiKey      string
	accessToken string
}

func (z *ZoomEye) request(method, u string, body io.Reader, result Result) error {
	req, err := http.NewRequest(method, u, body)
	if z.apiKey != "" {
		req.Header.Set("API-KEY", z.apiKey)
	}
	if z.accessToken != "" {
		req.Header.Set("Authorization", "JWT "+z.accessToken)
	}
	resp, err := httpCli.Do(req)
	if err != nil {
		return err
	}
	b, err := ioutil.ReadAll(resp.Body)
	if resp.Body.Close(); err != nil {
		return err
	}
	if resp.StatusCode == 200 {
		if err = json.Unmarshal(b, result); err != nil {
			return err
		}
		result.setRawData(b)
		return nil
	}
	if resp.StatusCode == 403 && bytes.Contains(b, []byte("specified resource")) {
		return nil
	}
	e := &ErrorResult{}
	if err = json.Unmarshal(b, &e); err != nil {
		return err
	}
	return e
}

func (z *ZoomEye) get(u string, params map[string]interface{}, result Result) error {
	if params != nil {
		uu, err := url.Parse(u)
		if err != nil {
			return err
		}
		query := uu.Query()
		for k, v := range params {
			query.Add(k, fmt.Sprintf("%v", v))
		}
		uu.RawQuery = query.Encode()
		u = uu.String()
	}
	return z.request(http.MethodGet, u, nil, result)
}

func (z *ZoomEye) post(u string, headers map[string]string, data map[string]interface{}, result Result) error {
	var body io.Reader
	if data != nil {
		b, err := json.Marshal(data)
		if err != nil {
			return err
		}
		body = bytes.NewBuffer(b)
	}
	return z.request(http.MethodPost, u, body, result)
}

// Login uses username/password for authentication
func (z *ZoomEye) Login(username, password string) (string, error) {
	var (
		data = map[string]interface{}{
			"username": username,
			"password": password,
		}
		result = &LoginResult{}
	)
	if err := z.post(loginAPI, nil, data, result); err != nil {
		return "", err
	}
	z.accessToken = result.AccessToken
	return z.accessToken, nil
}

// ResourcesInfo gets account resource information
func (z *ZoomEye) ResourcesInfo() (*ResourcesInfoResult, error) {
	var (
		result = &ResourcesInfoResult{}
		err    = z.get(userinfoAPI, nil, result)
	)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// DorkSearch searches the data of the specified page according to dork
func (z *ZoomEye) DorkSearch(dork string, page int, resource string, facet string) (*SearchResult, error) {
	if page <= 0 {
		page = 1
	}
	if resource = strings.ToLower(resource); resource != "web" {
		resource = "host"
	}
	if facet == "" {
		facet = defaultFacets[resource]
	}
	var (
		params = map[string]interface{}{
			"query":  dork,
			"page":   page,
			"facets": facet,
		}
		result = &SearchResult{
			Type: resource,
		}
		err = z.get(fmt.Sprintf(searchAPI, resource), params, result)
	)
	if err != nil {
		return nil, err
	}
	if len(result.Matches) == 0 {
		return nil, fmt.Errorf("no any results for the dork")
	}
	return result, nil
}

func (z *ZoomEye) conMPSearch(dork string, maxPage int, resource string, facet string) (map[int]*SearchResult, error) {
	var (
		results     = make(map[int]*SearchResult)
		ch          = make(chan map[string]interface{}, maxPage-1)
		ctx, cancel = context.WithCancel(context.Background())
	)
	defer close(ch)
	defer cancel()
	var (
		wg        sync.WaitGroup
		groupSize = 20
	)
	if maxPage < 21 {
		groupSize = maxPage - 1
	}
	var currPage int32 = 1
	for i := 0; i < groupSize; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				default:
					page := int(atomic.AddInt32(&currPage, 1))
					if page > maxPage {
						return
					}
					m := map[string]interface{}{
						"page": page,
					}
					if res, err := z.DorkSearch(dork, page, resource, facet); err != nil {
						m["result"] = err
					} else {
						m["result"] = res
					}
					ch <- m
				}
			}
		}()
	}
	var (
		err     error
		waiting bool
		done    = make(chan struct{}, 1)
	)
RCPT_LOOP:
	for i := 0; i < groupSize; i++ {
		select {
		case <-ctx.Done():
			if !waiting {
				waiting = true
				go func() {
					wg.Wait()
					close(done)
				}()
			}
		case <-done:
			break RCPT_LOOP
		case c := <-ch:
			switch res := c["result"].(type) {
			case error:
				cancel()
				err = res
			case *SearchResult:
				results[c["page"].(int)] = res
			}
		}
	}
	if wg.Wait(); err != nil && len(results) == 0 {
		return nil, err
	}
	return results, nil
}

// MultiPageSearch searches multiple pages of data according to dork
func (z *ZoomEye) MultiPageSearch(dork string, maxPage int, resource string, facet string) (map[int]*SearchResult, error) {
	if maxPage <= 0 {
		maxPage = 1
	}
	info, err := z.ResourcesInfo()
	if err != nil {
		return nil, err
	}
	allowPage := info.Resources.Search / 20
	if info.Resources.Search%20 > 0 {
		allowPage++
	}
	results := make(map[int]*SearchResult)
	if allowPage > 0 {
		res, err := z.DorkSearch(dork, 1, resource, facet)
		if err != nil {
			return nil, err
		}
		results[1] = res
		n := int(res.Total / 20)
		if res.Total%20 > 0 {
			n++
		}
		if n < allowPage {
			allowPage = n
		}
	}
	if maxPage > allowPage {
		maxPage = allowPage
	}
	if maxPage > 5 {
		corResults, err := z.conMPSearch(dork, maxPage, resource, facet)
		if err != nil {
			if len(results) > 0 {
				return results, nil
			}
			return nil, err
		}
		for k, v := range results {
			if _, ok := corResults[k]; !ok {
				corResults[k] = v
			}
		}
		return corResults, nil
	}
	for i := 1; i < maxPage; i++ {
		var (
			page = i + 1
			res  *SearchResult
		)
		if res, err = z.DorkSearch(dork, page, resource, facet); err != nil {
			break
		}
		results[page] = res
	}
	if err != nil && len(results) == 0 {
		return nil, err
	}
	return results, nil
}

// MultiToOneSearch searches multiple pages of data according to dork, and merges all results
func (z *ZoomEye) MultiToOneSearch(dork string, maxPage int, resource string, facet string) (*SearchResult, error) {
	results, err := z.MultiPageSearch(dork, maxPage, resource, facet)
	if err != nil {
		return nil, err
	}
	result := &SearchResult{
		Type: resource,
	}
	for i, n := 0, len(results); i < maxPage && n > 0; i++ {
		if res, ok := results[i+1]; ok {
			result.Extend(res)
			n--
		}
	}
	return result, nil
}

// HistoryIP queries IP history information
func (z *ZoomEye) HistoryIP(ip string) (*HistoryResult, error) {
	var (
		result = &HistoryResult{}
		err    = z.get(fmt.Sprintf(historyAPI, ip), nil, result)
	)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// NewWithKey creates instance of ZoomEye with API-Key and AccessToken
func NewWithKey(apiKey, accessToken string) *ZoomEye {
	return &ZoomEye{
		apiKey:      apiKey,
		accessToken: accessToken,
	}
}

// New creates instance of ZoomEye
func New() *ZoomEye {
	return &ZoomEye{}
}
