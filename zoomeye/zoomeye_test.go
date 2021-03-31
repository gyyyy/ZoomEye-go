package zoomeye

import (
	"testing"
)

const (
	tUsername = "username@zoomeye.org"
	tPassword = "password"
	tAPIKey   = "XXXXXXXX-XXXX-XXXXX-XXXX-XXXXXXXXXXX"
)

var defaultZoom = NewWithKey(tAPIKey, "")

func TestLogin(t *testing.T) {
	zoom := New()
	tok, err := zoom.Login(tUsername, tPassword)
	if err != nil || (tok != zoom.accessToken) {
		t.Fail()
	}
	if _, err = New().Login("test", "123456"); err == nil {
		t.Fail()
	}
}

func TestResourcesInfo(t *testing.T) {
	result, err := defaultZoom.ResourcesInfo()
	if err != nil || (result.Plan == "") {
		t.Fail()
	}
	if _, err = NewWithKey("00000000-0000-00000-0000-00000000000", "").ResourcesInfo(); err == nil {
		t.Fail()
	}
}

func TestDorkSearch(t *testing.T) {
	result, err := defaultZoom.DorkSearch("solr", 0, "", "")
	if err != nil {
		t.FailNow()
	}
	t.Log(result.Hosts())
	if _, err = defaultZoom.DorkSearch("country:cn", 0, "", ""); err != nil {
		t.Fail()
	}
	if _, err = defaultZoom.DorkSearch("solr country:cn", 0, "", "os,country"); err != nil {
		t.Fail()
	}
	if result, err = defaultZoom.DorkSearch("solr country:cn", 0, "web", ""); err != nil {
		t.FailNow()
	}
	t.Log(result.Sites())
	if n := len(result.Matches); n > 0 {
		t.Log(result.Matches[n-1].FindString("geoinfo.country.names.zh-CN"))
	}
}

func TestMultiPageSearch(t *testing.T) {
	var (
		maxPage      = 2
		results, err = defaultZoom.MultiPageSearch("dedecms country:cn", maxPage, "web", "")
	)
	if err != nil || (len(results) == 0) {
		t.FailNow()
	}
	t.Log(results[1].Total, results[1].Type, len(results))
}

func TestMultiToOneSearch(t *testing.T) {
	var (
		maxPage     = 2
		result, err = defaultZoom.MultiToOneSearch("dedecms country:cn", maxPage, "web", "")
	)
	if err != nil || (len(result.Matches) != maxPage*20) {
		t.FailNow()
	}
	t.Log(result.Total, result.Type, len(result.Matches))
}

func TestFilter(t *testing.T) {
	result, err := defaultZoom.DorkSearch("port:21", 0, "host", "")
	if err != nil {
		t.Fail()
	} else {
		t.Log(result.Filter("app"))
	}
	if result, err = defaultZoom.DorkSearch("dedecms", 0, "web", ""); err != nil {
		t.FailNow()
	}
	t.Log(result.Filter("site", "ip", "country"))
}

func TestHistoryIP(t *testing.T) {
	result, err := defaultZoom.HistoryIP("1.2.3.4")
	if err != nil {
		t.FailNow()
	}
	t.Log(result)
	t.Log(result.Filter("time=^2016", "app"))
}
