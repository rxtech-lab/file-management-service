package services

import (
	"context"
	"sync"
)

// MockInvoiceService is a mock implementation for testing
type MockInvoiceService struct {
	ProcessInvoiceFunc func(ctx context.Context, fileURL, authToken string, eventChan chan<- InvoiceStreamEvent) (*InvoiceResult, error)
	DeleteInvoiceFunc  func(ctx context.Context, invoiceID int64, authToken string) error
	IsEnabledValue     bool
	DeleteInvoiceCalls []DeleteInvoiceCall
	mu                 sync.Mutex
}

// DeleteInvoiceCall records a call to DeleteInvoice for test assertions
type DeleteInvoiceCall struct {
	InvoiceID int64
	AuthToken string
}

// NewMockInvoiceService creates a new mock invoice service
func NewMockInvoiceService(enabled bool) *MockInvoiceService {
	return &MockInvoiceService{
		IsEnabledValue:     enabled,
		DeleteInvoiceCalls: make([]DeleteInvoiceCall, 0),
	}
}

func (m *MockInvoiceService) ProcessInvoice(ctx context.Context, fileURL, authToken string, eventChan chan<- InvoiceStreamEvent) (*InvoiceResult, error) {
	if eventChan != nil {
		close(eventChan)
	}
	if m.ProcessInvoiceFunc != nil {
		return m.ProcessInvoiceFunc(ctx, fileURL, authToken, eventChan)
	}
	return &InvoiceResult{InvoiceID: 12345, Message: "Mock invoice processed"}, nil
}

func (m *MockInvoiceService) DeleteInvoice(ctx context.Context, invoiceID int64, authToken string) error {
	m.mu.Lock()
	m.DeleteInvoiceCalls = append(m.DeleteInvoiceCalls, DeleteInvoiceCall{
		InvoiceID: invoiceID,
		AuthToken: authToken,
	})
	m.mu.Unlock()

	if m.DeleteInvoiceFunc != nil {
		return m.DeleteInvoiceFunc(ctx, invoiceID, authToken)
	}
	return nil
}

func (m *MockInvoiceService) IsEnabled() bool {
	return m.IsEnabledValue
}

// GetDeleteInvoiceCalls returns a copy of the recorded DeleteInvoice calls
func (m *MockInvoiceService) GetDeleteInvoiceCalls() []DeleteInvoiceCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]DeleteInvoiceCall{}, m.DeleteInvoiceCalls...)
}
