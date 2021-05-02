package main

import (
	"crypto/sha256"
	"encoding/xml"
	errors2 "errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	metricPrefix = "experia_v10_"
)

const (
	tokenUrl = "http://%s/function_module/login_module/login_page/logintoken_lua.lua"
	loginUrl = "http://%s"
	logoutUrl = "http://%s"

	ethernetPageUrl = "http://%s/getpage.lua?pid=123&nextpage=Localnet_LAN_LocalnetStatus_t.lp&Menu3Location=0&_=1611056303063"
	ethernetMetricsUrl = "http://%s/common_page/lanStatus_lua.lua"

	dslPageUrl = "http://%s/getpage.lua?pid=123&nextpage=Internet_InternetStatusforRoute_DSL_t.lp&Menu3Location=0"
	dslMetricsUrl = "http://%s/common_page/internet_dsl_interface_lua.lua"
)

var (
	ethernetDesc = prometheus.NewDesc(
		metricPrefix+"ethernet",
		"All ethernet (eth) related metadata.",
		[]string{"value"}, nil)

	dslDesc = prometheus.NewDesc(
		metricPrefix+"dsl",
		"All dsl related metadata.",
		[]string{"value"}, nil)
	
	ifInOctets = prometheus.NewDesc(
			"ifInOctets",
			"The total number of octets received on the interface",
		[]string{"ifName", "ifAlias"}, nil)
	ifOutOctets = prometheus.NewDesc(
			"ifOutOctets",
			"The total number of octets transmitted out of the interface",
		[]string{"ifName", "ifAlias"}, nil)
)

type experiav10Collector struct {
	ip		           net.IP
	username           string
	password           string
	client             *http.Client
	upMetric           prometheus.Gauge
	authErrorsMetric   prometheus.Counter
	scrapeErrorsMetric prometheus.Counter
}

func newCollector(ip net.IP, username, password string, timeout time.Duration) *experiav10Collector {
	cookieJar, _ := cookiejar.New(nil)

	return &experiav10Collector{
		ip:  ip,
		username: username,
		password: password,
		client: &http.Client{
			Timeout: timeout,
			Jar: cookieJar,
		},
		upMetric: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: metricPrefix + "up",
			Help: "Shows if the Experia Box V10 is deemed up by the collector.",
		}),
		authErrorsMetric: prometheus.NewCounter(prometheus.CounterOpts{
			Name: metricPrefix + "auth_errors_total",
			Help: "Counts number of authentication errors encountered by the collector.",
		}),
		scrapeErrorsMetric: prometheus.NewCounter(prometheus.CounterOpts{
			Name: metricPrefix + "scrape_errors_total",
			Help: "Counts the number of scrape errors by this collector.",
		}),
	}
}

func (c *experiav10Collector) Describe(ch chan<- *prometheus.Desc) {
	c.upMetric.Describe(ch)
	c.authErrorsMetric.Describe(ch)
	c.scrapeErrorsMetric.Describe(ch)
	ch <- ifInOctets
	ch <- ifOutOctets
	ch <- dslDesc
}

func (c *experiav10Collector) Collect(ch chan<- prometheus.Metric) {
	if err := c.login(ch); err != nil {
		log.Printf("Error during authentication: %s", err)

		c.authErrorsMetric.Inc()
		c.upMetric.Set(0)
	} else {
		defer c.logout(ch)

		if err := c.scrape(ch); err != nil {
			log.Printf("Error during scrape: %s", err)

			c.scrapeErrorsMetric.Inc()
			c.upMetric.Set(0)
		} else {
			c.upMetric.Set(1)
		}
	}

	c.upMetric.Collect(ch)
	c.authErrorsMetric.Collect(ch)
	c.scrapeErrorsMetric.Collect(ch)
}

func (c *experiav10Collector) login(ch chan<- prometheus.Metric) error {
	loginPageRequest, err := c.client.Get(fmt.Sprintf(loginUrl, c.ip.String()))
	if err != nil {
		return err
	}

	tokenRequest, err := c.client.Get(fmt.Sprintf(tokenUrl, c.ip.String()))
	if err != nil {
		return err
	}

	type tokenResponseStruct struct {
		Token int `xml:",chardata"`
	}

	tokenData, err := ioutil.ReadAll(tokenRequest.Body)
	if err != nil {
		return err
	}

	var tokenResponse tokenResponseStruct
	err = xml.Unmarshal(tokenData, &tokenResponse)
	if err != nil {
		return err
	}

	loginParams := url.Values{}
	loginParams.Set("Username", c.username)
	loginParams.Set("Password", fmt.Sprintf("%x", sha256.Sum256([]byte(c.password + strconv.Itoa(tokenResponse.Token)))))
	loginParams.Set("action", "login")

	loginRequest, err := c.client.PostForm(fmt.Sprintf(loginUrl, c.ip.String()), loginParams)
	if err != nil {
		return err
	}

	defer loginPageRequest.Body.Close()
	defer tokenRequest.Body.Close()
	defer loginRequest.Body.Close()

	body, _ := ioutil.ReadAll(loginRequest.Body)

	if strings.Contains(string(body), "loginWrapper") {
		return errors2.New("unable to login")
	}

	return nil
}

func (c *experiav10Collector) logout(ch chan<- prometheus.Metric) error {
	logoutParams := url.Values{}
	logoutParams.Set("IF_LogOff", "1")
	logoutParams.Set("IF_LanguageSwitch", "")
	logoutParams.Set("IF_ModeSwitch", "")

	logoutRequest, err := c.client.PostForm(fmt.Sprintf(logoutUrl, c.ip.String()), logoutParams)
	if err != nil {
		return err
	}

	logoutRequest.Body.Close()

	c.client.Jar, _ = cookiejar.New(nil)

	return nil
}


func (c *experiav10Collector) scrape(ch chan<- prometheus.Metric) error {
	// For some reason the page containing the actual data will only contain data
	// after this page is loaded first
	dslPageRequest, err := c.client.Get(fmt.Sprintf(dslPageUrl, c.ip.String()))
	if err != nil {
		return err
	}

	dslMetricsRequest, err := c.client.Get(fmt.Sprintf(dslMetricsUrl, c.ip.String()))
	if err != nil {
		return err
	}

	defer dslPageRequest.Body.Close()
	defer dslMetricsRequest.Body.Close()

	dslMetricsData, err := ioutil.ReadAll(dslMetricsRequest.Body)
	if err != nil {
		return err
	}

	type dslMetricsStruct struct {
		Names []string `xml:"OBJ_DSLINTERFACE_ID>Instance>ParaName"`
		Values []string `xml:"OBJ_DSLINTERFACE_ID>Instance>ParaValue"`
	}

	var dslMetricsResponse dslMetricsStruct
	err = xml.Unmarshal(dslMetricsData, &dslMetricsResponse)
	if err != nil {
		return err
	}

	for i := 0; i < len(dslMetricsResponse.Names); i++ {
		value, err := strconv.ParseFloat(dslMetricsResponse.Values[i], 0)
		if err != nil {
			continue
		}

		metric, err := prometheus.NewConstMetric(dslDesc, prometheus.CounterValue, value, dslMetricsResponse.Names[i])
		if err != nil {
			return fmt.Errorf("error creating metric for %s: %s", dslDesc, err)
		}

		ch <- metric
	}

	ethernetPageRequest, err := c.client.Get(fmt.Sprintf(ethernetPageUrl, c.ip.String()))
	if err != nil {
		return err
	}

	ethernetMetricsRequest, err := c.client.Get(fmt.Sprintf(ethernetMetricsUrl, c.ip.String()))
	if err != nil {
		return err
	}

	defer ethernetPageRequest.Body.Close()
	defer ethernetMetricsRequest.Body.Close()

	ethernetMetricsData, err := ioutil.ReadAll(ethernetMetricsRequest.Body)
	if err != nil {
		return err
	}

	type ethernetMetricsStruct struct {
		Names []string `xml:"OBJ_ETH_ID>Instance>ParaName"`
		Values []string `xml:"OBJ_ETH_ID>Instance>ParaValue"`
	}

	var ethernetMetricsResponse ethernetMetricsStruct
	err = xml.Unmarshal(ethernetMetricsData, &ethernetMetricsResponse)
	if err != nil {
		return err
	}

	// Each LAN Instance has 6 fields
	for i := 0; i < len(ethernetMetricsResponse.Names); i += 6 {
		ifName := ethernetMetricsResponse.Values[i]
		ifAlias := ethernetMetricsResponse.Values[i+1]

		inBytes, err := strconv.ParseFloat(ethernetMetricsResponse.Values[i+2], 0)
		if err != nil {
			continue
		}
		outBytes, err := strconv.ParseFloat(ethernetMetricsResponse.Values[i+5], 0)
		if err != nil {
			continue
		}

		ch <- prometheus.MustNewConstMetric(ifInOctets, prometheus.CounterValue, inBytes, ifName, ifAlias)
		ch <- prometheus.MustNewConstMetric(ifOutOctets, prometheus.CounterValue, outBytes, ifName, ifAlias)

	}

	c.client.CloseIdleConnections()

	return nil
}