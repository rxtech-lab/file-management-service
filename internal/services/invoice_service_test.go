package services

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDeleteInvoice_Success(t *testing.T) {
	var requestCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/api/invoices/12345", r.URL.Path)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	service := NewInvoiceService(InvoiceConfig{ServerURL: "dummy", McpServerURL: server.URL})
	err := service.DeleteInvoice(context.Background(), 12345, "test-token")

	assert.NoError(t, err)
	assert.Equal(t, int32(1), atomic.LoadInt32(&requestCount))
}

func TestDeleteInvoice_SuccessWithOK(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	service := NewInvoiceService(InvoiceConfig{ServerURL: "dummy", McpServerURL: server.URL})
	err := service.DeleteInvoice(context.Background(), 12345, "test-token")

	assert.NoError(t, err)
}

func TestDeleteInvoice_NotFoundIsSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	service := NewInvoiceService(InvoiceConfig{ServerURL: "dummy", McpServerURL: server.URL})
	err := service.DeleteInvoice(context.Background(), 12345, "test-token")

	assert.NoError(t, err) // 404 should not be an error (invoice already deleted)
}

func TestDeleteInvoice_RetryOnFailure(t *testing.T) {
	var requestCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&requestCount, 1)
		if count < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	service := NewInvoiceService(InvoiceConfig{ServerURL: "dummy", McpServerURL: server.URL})
	err := service.DeleteInvoice(context.Background(), 12345, "test-token")

	assert.NoError(t, err)
	assert.Equal(t, int32(3), atomic.LoadInt32(&requestCount)) // Failed twice, succeeded on third
}

func TestDeleteInvoice_AllRetriesFail(t *testing.T) {
	var requestCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	service := NewInvoiceService(InvoiceConfig{ServerURL: "dummy", McpServerURL: server.URL})
	err := service.DeleteInvoice(context.Background(), 12345, "test-token")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected status: 500")
	assert.Equal(t, int32(4), atomic.LoadInt32(&requestCount)) // 1 initial + 3 retries
}

func TestDeleteInvoice_NoAuthToken(t *testing.T) {
	// Server should never be called
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Server should not be called when no auth token")
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	service := NewInvoiceService(InvoiceConfig{ServerURL: "dummy", McpServerURL: server.URL})
	err := service.DeleteInvoice(context.Background(), 12345, "")

	assert.NoError(t, err) // Should skip without error
}

func TestDeleteInvoice_ServiceNotEnabled(t *testing.T) {
	service := NewInvoiceService(InvoiceConfig{ServerURL: ""})
	err := service.DeleteInvoice(context.Background(), 12345, "test-token")

	assert.NoError(t, err) // Should skip without error
}

func TestDeleteInvoice_ContextCancelled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	service := NewInvoiceService(InvoiceConfig{ServerURL: "dummy", McpServerURL: server.URL})
	err := service.DeleteInvoice(ctx, 12345, "test-token")

	assert.Error(t, err) // Should fail due to context timeout
}

func TestDeleteInvoice_AuthorizationHeader(t *testing.T) {
	var receivedAuthHeader string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuthHeader = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	service := NewInvoiceService(InvoiceConfig{ServerURL: "dummy", McpServerURL: server.URL})
	err := service.DeleteInvoice(context.Background(), 99999, "my-oauth-token-123")

	assert.NoError(t, err)
	assert.Equal(t, "Bearer my-oauth-token-123", receivedAuthHeader)
}

func TestDeleteInvoice_CorrectEndpoint(t *testing.T) {
	var receivedPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	service := NewInvoiceService(InvoiceConfig{ServerURL: "dummy", McpServerURL: server.URL})
	err := service.DeleteInvoice(context.Background(), 67890, "test-token")

	assert.NoError(t, err)
	assert.Equal(t, "/api/invoices/67890", receivedPath)
}

// Test MockInvoiceService
func TestMockInvoiceService_TracksCalls(t *testing.T) {
	mock := NewMockInvoiceService(true)

	err := mock.DeleteInvoice(context.Background(), 111, "token-aaa")
	assert.NoError(t, err)

	err = mock.DeleteInvoice(context.Background(), 222, "token-bbb")
	assert.NoError(t, err)

	calls := mock.GetDeleteInvoiceCalls()
	assert.Len(t, calls, 2)
	assert.Equal(t, int64(111), calls[0].InvoiceID)
	assert.Equal(t, "token-aaa", calls[0].AuthToken)
	assert.Equal(t, int64(222), calls[1].InvoiceID)
	assert.Equal(t, "token-bbb", calls[1].AuthToken)
}

func TestMockInvoiceService_CustomDeleteFunc(t *testing.T) {
	mock := NewMockInvoiceService(true)
	mock.DeleteInvoiceFunc = func(ctx context.Context, invoiceID int64, authToken string) error {
		return assert.AnError
	}

	err := mock.DeleteInvoice(context.Background(), 123, "token")
	assert.Error(t, err)

	// Should still track the call
	calls := mock.GetDeleteInvoiceCalls()
	assert.Len(t, calls, 1)
}

func TestMockInvoiceService_IsEnabled(t *testing.T) {
	mockEnabled := NewMockInvoiceService(true)
	assert.True(t, mockEnabled.IsEnabled())

	mockDisabled := NewMockInvoiceService(false)
	assert.False(t, mockDisabled.IsEnabled())
}
