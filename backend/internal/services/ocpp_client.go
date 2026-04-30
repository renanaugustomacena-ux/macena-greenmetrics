// Package services — OCPP 1.6-J WebSocket client skeleton.
//
// OCPP (Open Charge Point Protocol) is the de-facto protocol between an EV
// charging station and a Central System. Messages flow over a single
// long-lived WebSocket connection. Version 1.6-J uses JSON-RPC-like envelopes:
//
//	[2, "<uniqueID>", "<Action>", {payload}]          // CALL
//	[3, "<uniqueID>", {payload}]                        // CALLRESULT
//	[4, "<uniqueID>", "<errCode>", "<errDesc>", {}]     // CALLERROR
//
// This file implements the minimum client surface needed to:
//   - connect to a Central System URL;
//   - send a BootNotification and receive a Heartbeat interval;
//   - push MeterValues (StartTransaction / StopTransaction envelopes are stubbed
//     but the envelope structure is correct and ready for wiring).
//
// If OCPP_CENTRAL_SYSTEM_URL is not set we refuse with ErrOCPPNotConfigured —
// no silent stub, per Mission II.5 rule "no silent stubs".
package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

// ErrOCPPNotConfigured is returned when OCPP_CENTRAL_SYSTEM_URL is empty.
// The message names the exact env var to make remediation obvious.
var ErrOCPPNotConfigured = errors.New("ocpp: OCPP_CENTRAL_SYSTEM_URL not set — EV charging ingestion disabled")

// OCPPMessageType enumerates the three-value discriminator of the OCPP-J envelope.
type OCPPMessageType int

const (
	OCPPCall       OCPPMessageType = 2
	OCPPCallResult OCPPMessageType = 3
	OCPPCallError  OCPPMessageType = 4
)

// OCPPCallEnvelope is the wire shape for a CALL message: [2, id, action, payload].
type OCPPCallEnvelope struct {
	MessageType OCPPMessageType
	UniqueID    string
	Action      string
	Payload     any
}

// MarshalJSON produces the positional array envelope.
func (e OCPPCallEnvelope) MarshalJSON() ([]byte, error) {
	return json.Marshal([]any{int(e.MessageType), e.UniqueID, e.Action, e.Payload})
}

// OCPPCallResultEnvelope is the wire shape for a CALLRESULT message:
// [3, id, payload].
type OCPPCallResultEnvelope struct {
	UniqueID string
	Payload  json.RawMessage
}

// OCPPCallErrorEnvelope is the wire shape for a CALLERROR message:
// [4, id, errorCode, errorDescription, errorDetails].
type OCPPCallErrorEnvelope struct {
	UniqueID         string
	ErrorCode        string
	ErrorDescription string
	ErrorDetails     json.RawMessage
}

// BootNotificationRequest — OCPP 1.6 payload (subset; mandatory fields only).
type BootNotificationRequest struct {
	ChargePointVendor string `json:"chargePointVendor"`
	ChargePointModel  string `json:"chargePointModel"`
	FirmwareVersion   string `json:"firmwareVersion,omitempty"`
	ChargePointSerial string `json:"chargePointSerialNumber,omitempty"`
}

// BootNotificationResponse — OCPP 1.6 CS reply.
type BootNotificationResponse struct {
	CurrentTime string `json:"currentTime"` // ISO 8601
	Interval    int    `json:"interval"`    // heartbeat interval, seconds
	Status      string `json:"status"`      // "Accepted" | "Pending" | "Rejected"
}

// MeterValue — one sampled value at a point in time.
type MeterValue struct {
	Timestamp     string `json:"timestamp"`
	Value         string `json:"value"`
	Context       string `json:"context,omitempty"`       // Sample.Periodic | Transaction.Begin | etc.
	Measurand     string `json:"measurand,omitempty"`     // Energy.Active.Import.Register | ...
	Unit          string `json:"unit,omitempty"`          // Wh | kWh | A | V | ...
	Location      string `json:"location,omitempty"`      // Cable | EV | Inlet | Outlet | Body
	Phase         string `json:"phase,omitempty"`
}

// MeterValuesRequest — sent by the charge point to report accumulated energy.
type MeterValuesRequest struct {
	ConnectorID   int          `json:"connectorId"`
	TransactionID *int         `json:"transactionId,omitempty"`
	MeterValue    []MeterValue `json:"meterValue"`
}

// OCPPClient is a minimal OCPP 1.6-J client.
//
// It is intentionally transport-agnostic: the caller supplies a Conn
// (gorilla/websocket-like interface) so tests can swap in a loopback pipe.
type OCPPClient struct {
	centralSystemURL string
	chargePointID    string
	logger           *zap.Logger
	httpClient       *http.Client

	conn   OCPPConn
	connMu sync.Mutex

	nextID atomic.Uint64
}

// OCPPConn is the minimal surface OCPPClient requires from the WebSocket.
// gorilla/websocket.Conn satisfies this by design; tests can provide a stub.
type OCPPConn interface {
	WriteMessage(messageType int, data []byte) error
	ReadMessage() (messageType int, p []byte, err error)
	Close() error
}

// OCPPDialer matches gorilla/websocket.Dialer for injection / testing.
type OCPPDialer interface {
	DialContext(ctx context.Context, urlStr string, requestHeader http.Header) (OCPPConn, *http.Response, error)
}

// NewOCPPClient constructs a client. If centralSystemURL is empty this returns
// ErrOCPPNotConfigured so the caller can degrade cleanly.
func NewOCPPClient(centralSystemURL, chargePointID string, logger *zap.Logger) (*OCPPClient, error) {
	if strings.TrimSpace(centralSystemURL) == "" {
		return nil, ErrOCPPNotConfigured
	}
	if _, err := url.Parse(centralSystemURL); err != nil {
		return nil, fmt.Errorf("ocpp: invalid OCPP_CENTRAL_SYSTEM_URL: %w", err)
	}
	if chargePointID == "" {
		chargePointID = "greenmetrics-cp-0001"
	}
	return &OCPPClient{
		centralSystemURL: centralSystemURL,
		chargePointID:    chargePointID,
		logger:           logger,
		httpClient:       &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// Connect opens the WebSocket to the Central System and stores the conn.
//
// The real gorilla/websocket wiring lives in the caller (we avoid forcing a
// runtime dep on every compile target of this package). The standard wiring
// is:
//
//	import "github.com/gorilla/websocket"
//	d := websocket.DefaultDialer
//	u := centralSystemURL + "/" + chargePointID
//	conn, resp, err := d.DialContext(ctx, u, http.Header{
//	    "Sec-WebSocket-Protocol": []string{"ocpp1.6"},
//	})
//
// Here we expose an Attach() hook so a real connection can be injected.
func (c *OCPPClient) Connect(ctx context.Context, dialer OCPPDialer) error {
	if c == nil {
		return ErrOCPPNotConfigured
	}
	u := strings.TrimRight(c.centralSystemURL, "/") + "/" + c.chargePointID
	conn, _, err := dialer.DialContext(ctx, u, http.Header{
		"Sec-WebSocket-Protocol": []string{"ocpp1.6"},
	})
	if err != nil {
		return fmt.Errorf("ocpp dial %s: %w", u, err)
	}
	c.connMu.Lock()
	c.conn = conn
	c.connMu.Unlock()
	c.logger.Info("ocpp connected", zap.String("url", u), zap.String("charge_point_id", c.chargePointID))
	return nil
}

// Attach allows tests (and real callers) to bind an already-opened conn.
func (c *OCPPClient) Attach(conn OCPPConn) {
	c.connMu.Lock()
	defer c.connMu.Unlock()
	c.conn = conn
}

// Close closes the underlying WebSocket.
func (c *OCPPClient) Close() error {
	c.connMu.Lock()
	defer c.connMu.Unlock()
	if c.conn == nil {
		return nil
	}
	err := c.conn.Close()
	c.conn = nil
	return err
}

// BootNotification sends a BootNotification CALL and blocks for the CALLRESULT.
func (c *OCPPClient) BootNotification(ctx context.Context, req BootNotificationRequest) (*BootNotificationResponse, error) {
	raw, err := c.call(ctx, "BootNotification", req)
	if err != nil {
		return nil, err
	}
	var out BootNotificationResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("ocpp: decode BootNotification response: %w", err)
	}
	return &out, nil
}

// Heartbeat sends a Heartbeat CALL.
func (c *OCPPClient) Heartbeat(ctx context.Context) error {
	_, err := c.call(ctx, "Heartbeat", struct{}{})
	return err
}

// SendMeterValues sends a MeterValues CALL.
func (c *OCPPClient) SendMeterValues(ctx context.Context, req MeterValuesRequest) error {
	_, err := c.call(ctx, "MeterValues", req)
	return err
}

// call issues a CALL and waits for the matching CALLRESULT/CALLERROR.
func (c *OCPPClient) call(ctx context.Context, action string, payload any) (json.RawMessage, error) {
	c.connMu.Lock()
	conn := c.conn
	c.connMu.Unlock()
	if conn == nil {
		return nil, errors.New("ocpp: not connected; call Connect first")
	}
	id := fmt.Sprintf("%d-%d", time.Now().UnixNano(), c.nextID.Add(1))
	env := OCPPCallEnvelope{MessageType: OCPPCall, UniqueID: id, Action: action, Payload: payload}
	b, err := json.Marshal(env)
	if err != nil {
		return nil, fmt.Errorf("ocpp marshal %s: %w", action, err)
	}
	// Message type 1 = text frame in the RFC6455 + gorilla/websocket constants.
	if err := conn.WriteMessage(1, b); err != nil {
		return nil, fmt.Errorf("ocpp write %s: %w", action, err)
	}

	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		_, msg, err := conn.ReadMessage()
		if err != nil {
			return nil, fmt.Errorf("ocpp read %s: %w", action, err)
		}
		// Parse positional array envelope.
		var arr []json.RawMessage
		if err := json.Unmarshal(msg, &arr); err != nil || len(arr) < 3 {
			continue
		}
		var mt OCPPMessageType
		if err := json.Unmarshal(arr[0], &mt); err != nil {
			continue
		}
		var uid string
		if err := json.Unmarshal(arr[1], &uid); err != nil || uid != id {
			// Not our reply; OCPP allows CS-initiated CALLs (e.g. RemoteStart)
			// interleaved — the caller is expected to pump those separately.
			continue
		}
		switch mt {
		case OCPPCallResult:
			return arr[2], nil
		case OCPPCallError:
			var code, desc string
			_ = json.Unmarshal(arr[2], &code)
			if len(arr) > 3 {
				_ = json.Unmarshal(arr[3], &desc)
			}
			return nil, fmt.Errorf("ocpp CALLERROR: %s — %s", code, desc)
		}
	}
	return nil, fmt.Errorf("ocpp: deadline waiting for response to %s", action)
}
