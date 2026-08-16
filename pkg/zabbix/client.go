package zabbix

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/ErKatta/zbxctl/pkg/config"
	"github.com/ErKatta/zbxctl/pkg/prompt"
)

type Request struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	Auth    string      `json:"auth,omitempty"`
	ID      int64       `json:"id"`
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"`
}

func (e *RPCError) Error() string {
	if e.Data != "" {
		return fmt.Sprintf("Zabbix API Error [%d]: %s (%s)", e.Code, e.Message, e.Data)
	}
	return fmt.Sprintf("Zabbix API Error [%d]: %s", e.Code, e.Message)
}

type Response struct {
	JSONRPC string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *RPCError       `json:"error,omitempty"`
	ID      int64           `json:"id"`
}

type Client struct {
	url                string
	apiToken           string
	username           string
	password           string
	httpUser           string
	httpPassword       string
	httpHeaders        map[string]string
	tlsCertFile        string
	tlsKeyFile         string
	tlsCAFile          string
	insecureSkipVerify bool
	httpClient         *http.Client
	sessionID          string
	mu                 sync.Mutex
	requestID          int64
}

func NewClient(ctx *config.Context) *Client {
	timeout := time.Duration(ctx.Timeout) * time.Second
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	tlsCfg := &tls.Config{
		InsecureSkipVerify: ctx.InsecureSkipVerify,
	}

	if ctx.TLSCAFile != "" {
		caCert, err := os.ReadFile(ctx.TLSCAFile)
		if err == nil {
			caPool, err := x509.SystemCertPool()
			if err != nil || caPool == nil {
				caPool = x509.NewCertPool()
			}
			caPool.AppendCertsFromPEM(caCert)
			tlsCfg.RootCAs = caPool
		}
	}

	if ctx.TLSCertFile != "" && ctx.TLSKeyFile != "" {
		cert, err := tls.LoadX509KeyPair(ctx.TLSCertFile, ctx.TLSKeyFile)
		if err == nil {
			tlsCfg.Certificates = []tls.Certificate{cert}
		}
	}

	tr := &http.Transport{
		TLSClientConfig: tlsCfg,
	}

	return &Client{
		url:                config.NormalizeURL(ctx.URL),
		apiToken:           ctx.APIToken,
		username:           ctx.Username,
		password:           ctx.Password,
		httpUser:           ctx.HTTPUser,
		httpPassword:       ctx.HTTPPassword,
		httpHeaders:        ctx.HTTPHeaders,
		tlsCertFile:        ctx.TLSCertFile,
		tlsKeyFile:         ctx.TLSKeyFile,
		tlsCAFile:          ctx.TLSCAFile,
		insecureSkipVerify: ctx.InsecureSkipVerify,
		httpClient: &http.Client{
			Timeout:   timeout,
			Transport: tr,
		},
	}
}

func (c *Client) Call(ctx context.Context, method string, params interface{}) (json.RawMessage, error) {
	c.mu.Lock()
	c.requestID++
	id := c.requestID
	c.mu.Unlock()

	authToken := c.apiToken
	if authToken == "" && c.username != "" && method != "user.login" && method != "apiinfo.version" {
		var err error
		authToken, err = c.ensureSession(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to authenticate session: %w", err)
		}
	}

	// Default params to empty map if nil
	if params == nil {
		params = map[string]interface{}{}
	}

	reqBody := Request{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
		Auth:    authToken,
		ID:      id,
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal json-rpc request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json-rpc")

	if c.httpUser != "" {
		if c.httpPassword == "" {
			pass, err := prompt.PromptPassword(fmt.Sprintf("Enter HTTP Basic Auth password for %q: ", c.httpUser))
			if err != nil {
				return nil, fmt.Errorf("http-password required for user %q: %w", c.httpUser, err)
			}
			c.httpPassword = pass
		}
		req.SetBasicAuth(c.httpUser, c.httpPassword)
	}

	if authToken != "" && req.Header.Get("Authorization") == "" {
		req.Header.Set("Authorization", "Bearer "+authToken)
	}

	for k, v := range c.httpHeaders {
		req.Header.Set(k, v)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http status %d: %s", resp.StatusCode, string(respBytes))
	}

	var rpcResp Response
	if err := json.Unmarshal(respBytes, &rpcResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal json-rpc response: %w", err)
	}

	if rpcResp.Error != nil {
		return nil, rpcResp.Error
	}

	return rpcResp.Result, nil
}

func (c *Client) ensureSession(ctx context.Context) (string, error) {
	c.mu.Lock()
	if c.sessionID != "" {
		token := c.sessionID
		c.mu.Unlock()
		return token, nil
	}
	c.mu.Unlock()

	if c.password == "" {
		pass, err := prompt.PromptPassword(fmt.Sprintf("Enter password for Zabbix user %q: ", c.username))
		if err != nil {
			return "", fmt.Errorf("password required for user %q: %w", c.username, err)
		}
		c.password = pass
	}

	params := map[string]interface{}{
		"username": c.username,
		"password": c.password,
	}

	res, err := c.Call(ctx, "user.login", params)
	if err != nil {
		// Fallback for legacy Zabbix API (< 5.4) which used "user" instead of "username"
		if strings.Contains(err.Error(), "username") || strings.Contains(err.Error(), "-32602") {
			paramsLegacy := map[string]interface{}{
				"user":     c.username,
				"password": c.password,
			}
			res, err = c.Call(ctx, "user.login", paramsLegacy)
		}
		if err != nil {
			return "", err
		}
	}

	var token string
	if err := json.Unmarshal(res, &token); err != nil {
		return "", fmt.Errorf("invalid token response from user.login: %w", err)
	}

	c.mu.Lock()
	c.sessionID = token
	c.mu.Unlock()

	return token, nil
}

func (c *Client) GetVersion(ctx context.Context) (string, error) {
	res, err := c.Call(ctx, "apiinfo.version", nil)
	if err != nil {
		return "", err
	}
	var ver string
	if err := json.Unmarshal(res, &ver); err != nil {
		return "", err
	}
	return ver, nil
}
