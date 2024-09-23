package utho

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const BaseUrl = "https://api.utho.com/v2/"

var defaultHTTPClient = &http.Client{Timeout: time.Second * 300}

type Client interface {
	NewRequest(method, url string, body ...interface{}) (*http.Request, error)
	Do(req *http.Request, v interface{}) (*http.Response, error)

	Account() *AccountService
	ApiKey() *ApiKeyService
	Action() *ActionService
	CloudInstances() *CloudInstancesService
	Domain() *DomainService
	Firewall() *FirewallService
	ISO() *ISOService
	Loadbalancers() *LoadbalancersService
	Monitoring() *MonitoringService
	ObjectStorage() *ObjectStorageService
	Sqs() *SqsService
	Ssl() *SslService
	Stacks() *StacksService
	TargetGroup() *TargetGroupService
	Vpc() *VpcService
	AutoScaling() *AutoScalingService
	Kubernetes() *KubernetesService
	Ebs() *EBService
}

type service struct {
	client Client
}

type client struct {
	client  *http.Client
	baseURL *url.URL
	token   string

	account        *AccountService
	apiKey         *ApiKeyService
	action         *ActionService
	cloudInstances *CloudInstancesService
	domain         *DomainService
	firewall       *FirewallService
	iso            *ISOService
	loadbalancers  *LoadbalancersService
	monitoring     *MonitoringService
	objectStorage  *ObjectStorageService
	sqs            *SqsService
	ssl            *SslService
	stacks         *StacksService
	targetgroup    *TargetGroupService
	vpc            *VpcService
	autoscaling    *AutoScalingService
	kubernetes     *KubernetesService
	ebs            *EBService
}

// NewClient creates a new Utho client.
// Because the token supplied will be used for all authenticated requests,
// the created client should not be used across different users
func NewClient(token string, options ...UthoOption) (Client, error) {
	if token == "" {
		return nil, errors.New("you must provide an API token")
	}

	defaultBaseURL, err := toURLWithEndingSlash(BaseUrl)
	if err != nil {
		return nil, err
	}

	client := &client{
		client:  defaultHTTPClient,
		baseURL: defaultBaseURL,
		token:   token,
	}

	for _, option := range options {
		if err = option(client); err != nil {
			return nil, err
		}
	}

	commonService := &service{client: client}
	client.account = (*AccountService)(commonService)
	client.apiKey = (*ApiKeyService)(commonService)
	client.action = (*ActionService)(commonService)
	client.cloudInstances = (*CloudInstancesService)(commonService)
	client.domain = (*DomainService)(commonService)
	client.firewall = (*FirewallService)(commonService)
	client.iso = (*ISOService)(commonService)
	client.loadbalancers = (*LoadbalancersService)(commonService)
	client.monitoring = (*MonitoringService)(commonService)
	client.objectStorage = (*ObjectStorageService)(commonService)
	client.sqs = (*SqsService)(commonService)
	client.ssl = (*SslService)(commonService)
	client.stacks = (*StacksService)(commonService)
	client.targetgroup = (*TargetGroupService)(commonService)
	client.vpc = (*VpcService)(commonService)
	client.autoscaling = (*AutoScalingService)(commonService)
	client.kubernetes = (*KubernetesService)(commonService)
	client.ebs = (*EBService)(commonService)

	return client, nil
}

func toURLWithEndingSlash(u string) (*url.URL, error) {
	baseURL, err := url.Parse(u)
	if err != nil {
		return nil, err
	}

	if !strings.HasSuffix(baseURL.Path, "/") {
		baseURL.Path += "/"
	}

	return baseURL, err
}

// NewRequest creates an API request.
// A relative URL `url` can be specified which is resolved relative to the baseURL of the client.
// Relative URLs should be specified without a preceding slash.
// The `body` parameter can be used to pass a body to the request. If no body is required, the parameter can be omitted.
func (c *client) NewRequest(method, url string, body ...interface{}) (*http.Request, error) {
	fullUrl, err := c.baseURL.Parse(url)
	if err != nil {
		return nil, err
	}

	var buf io.ReadWriter
	if len(body) > 0 && body[0] != nil {
		buf = &bytes.Buffer{}
		enc := json.NewEncoder(buf)
		enc.SetEscapeHTML(false)
		err := enc.Encode(body[0])
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, fullUrl.String(), buf)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept-Encoding", "application/json")

	return req, nil
}

// Do will send the given request using the client `c` on which it is called.
// If the response contains a body, it will be unmarshalled in `v`.
func (c *client) Do(req *http.Request, v interface{}) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = checkForErrors(resp)
	if err != nil {
		return resp, err
	}

	if resp.Body != nil && v != nil {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return resp, err
		}

		err = json.Unmarshal(body, &v)
		if err != nil {
			return resp, err
		}
	}

	return resp, nil
}

func checkForErrors(resp *http.Response) error {
	if c := resp.StatusCode; c >= 200 && c < 400 {
		return nil
	}

	errorResponse := &ErrorResponse{Response: resp}

	data, err := io.ReadAll(resp.Body)
	if err == nil && data != nil {
		// it's ok if we cannot unmarshal to Utho's error response
		_ = json.Unmarshal(data, errorResponse)
	}

	return errorResponse
}

func (c *client) Account() *AccountService {
	return c.account
}

func (c *client) ApiKey() *ApiKeyService {
	return c.apiKey
}

func (c *client) Action() *ActionService {
	return c.action
}

func (c *client) CloudInstances() *CloudInstancesService {
	return c.cloudInstances
}

func (c *client) Domain() *DomainService {
	return c.domain
}

func (c *client) Firewall() *FirewallService {
	return c.firewall
}

func (c *client) ISO() *ISOService {
	return c.iso
}

func (c *client) Loadbalancers() *LoadbalancersService {
	return c.loadbalancers
}

func (c *client) Monitoring() *MonitoringService {
	return c.monitoring
}

func (c *client) ObjectStorage() *ObjectStorageService {
	return c.objectStorage
}

func (c *client) Sqs() *SqsService {
	return c.sqs
}

func (c *client) Ssl() *SslService {
	return c.ssl
}

func (c *client) Stacks() *StacksService {
	return c.stacks
}

func (c *client) TargetGroup() *TargetGroupService {
	return c.targetgroup
}

func (c *client) Vpc() *VpcService {
	return c.vpc
}

func (c *client) AutoScaling() *AutoScalingService {
	return c.autoscaling
}

func (c *client) Kubernetes() *KubernetesService {
	return c.kubernetes
}

func (c *client) Ebs() *EBService {
	return c.ebs
}
