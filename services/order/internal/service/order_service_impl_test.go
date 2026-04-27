package service_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/lppduy/ecom-poc/services/order/internal/client"
	"github.com/lppduy/ecom-poc/services/order/internal/domain"
	"github.com/lppduy/ecom-poc/services/order/internal/repository"
	"github.com/lppduy/ecom-poc/services/order/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ── Manual mocks ─────────────────────────────────────────────────────────────

type mockOrderRepo struct {
	orders map[string]domain.Order // keyed by idempotency key
	byID   map[string]domain.Order // keyed by string id
	nextID int64

	updateStatusFn func(id int64, status string) error
}

func newMockRepo() *mockOrderRepo {
	return &mockOrderRepo{
		orders: make(map[string]domain.Order),
		byID:   make(map[string]domain.Order),
		nextID: 1,
	}
}

func (m *mockOrderRepo) FindByID(id string) (domain.Order, bool, error) {
	o, ok := m.byID[id]
	return o, ok, nil
}

func (m *mockOrderRepo) FindByIdempotencyKey(key string) (domain.Order, bool, error) {
	o, ok := m.orders[key]
	return o, ok, nil
}

func (m *mockOrderRepo) CreateWithItems(_ context.Context, userID, key string, items []domain.OrderItem) (domain.Order, error) {
	if len(items) == 0 {
		return domain.Order{}, errors.New("no items")
	}
	o := domain.Order{ID: m.nextID, UserID: userID, Status: domain.StatusPending, IdempotencyKey: key}
	m.nextID++
	m.orders[key] = o
	m.byID[fmt.Sprint(o.ID)] = o
	return o, nil
}

func (m *mockOrderRepo) UpdateStatus(id int64, newStatus string) error {
	if m.updateStatusFn != nil {
		return m.updateStatusFn(id, newStatus)
	}
	key := fmt.Sprint(id)
	o, ok := m.byID[key]
	if !ok {
		return errors.New("order not found")
	}
	o.Status = newStatus
	m.byID[key] = o
	// update idempotency map too
	if o.IdempotencyKey != "" {
		m.orders[o.IdempotencyKey] = o
	}
	return nil
}

type mockCartClient struct {
	items    []domain.OrderItem
	fetchErr error
	clearErr error
	cleared  []string
}

func (m *mockCartClient) FetchCartItems(_ string) ([]domain.OrderItem, error) {
	return m.items, m.fetchErr
}

func (m *mockCartClient) ClearCart(userID string) error {
	m.cleared = append(m.cleared, userID)
	return m.clearErr
}

// ensure interfaces are satisfied
var _ repository.OrderRepository = (*mockOrderRepo)(nil)
var _ client.CartClient = (*mockCartClient)(nil)

// ── Helpers ───────────────────────────────────────────────────────────────────

func makeService(repo repository.OrderRepository, cart client.CartClient) service.OrderService {
	return service.NewOrderService(repo, cart)
}

func cartWithItems(items ...domain.OrderItem) *mockCartClient {
	return &mockCartClient{items: items}
}

func someItems() []domain.OrderItem {
	return []domain.OrderItem{{ProductID: "p1", Quantity: 2}}
}

// ── CreateOrder ───────────────────────────────────────────────────────────────

func TestCreateOrder_Success(t *testing.T) {
	repo := newMockRepo()
	cart := cartWithItems(someItems()...)
	svc := makeService(repo, cart)

	order, existed, err := svc.CreateOrder(context.Background(), "u1", "key-1")

	require.NoError(t, err)
	assert.False(t, existed)
	assert.Equal(t, "u1", order.UserID)
	assert.Equal(t, domain.StatusPending, order.Status)
	assert.Contains(t, cart.cleared, "u1", "cart should be cleared after order")
}

func TestCreateOrder_IdempotentKey_ReturnsExistingOrder(t *testing.T) {
	repo := newMockRepo()
	cart := cartWithItems(someItems()...)
	svc := makeService(repo, cart)

	first, _, _ := svc.CreateOrder(context.Background(), "u1", "key-idem")
	second, existed, err := svc.CreateOrder(context.Background(), "u1", "key-idem")

	require.NoError(t, err)
	assert.True(t, existed)
	assert.Equal(t, first.ID, second.ID)
}

func TestCreateOrder_EmptyCart_ReturnsErrEmptyCart(t *testing.T) {
	repo := newMockRepo()
	cart := &mockCartClient{items: []domain.OrderItem{}} // empty
	svc := makeService(repo, cart)

	_, _, err := svc.CreateOrder(context.Background(), "u1", "key-empty")

	assert.ErrorIs(t, err, domain.ErrEmptyCart)
}

func TestCreateOrder_CartFetchError_Propagates(t *testing.T) {
	repo := newMockRepo()
	cart := &mockCartClient{fetchErr: errors.New("network error")}
	svc := makeService(repo, cart)

	_, _, err := svc.CreateOrder(context.Background(), "u1", "key-err")

	assert.Error(t, err)
}

// ── GetOrder ──────────────────────────────────────────────────────────────────

func TestGetOrder_Found(t *testing.T) {
	repo := newMockRepo()
	cart := cartWithItems(someItems()...)
	svc := makeService(repo, cart)

	created, _, _ := svc.CreateOrder(context.Background(), "u1", "key-get")

	found, ok, err := svc.GetOrder(fmt.Sprint(created.ID))
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, created.ID, found.ID)
}

func TestGetOrder_NotFound(t *testing.T) {
	repo := newMockRepo()
	svc := makeService(repo, &mockCartClient{})

	_, ok, err := svc.GetOrder("999")
	require.NoError(t, err)
	assert.False(t, ok)
}

// ── ConfirmOrder ──────────────────────────────────────────────────────────────

func TestConfirmOrder_Success(t *testing.T) {
	repo := newMockRepo()
	cart := cartWithItems(someItems()...)
	svc := makeService(repo, cart)

	created, _, _ := svc.CreateOrder(context.Background(), "u1", "key-confirm")
	updated, err := svc.ConfirmOrder(fmt.Sprint(created.ID))

	require.NoError(t, err)
	assert.Equal(t, domain.StatusConfirmed, updated.Status)
}

func TestConfirmOrder_NotFound_ReturnsErrOrderNotFound(t *testing.T) {
	repo := newMockRepo()
	svc := makeService(repo, &mockCartClient{})

	_, err := svc.ConfirmOrder("999")
	assert.ErrorIs(t, err, domain.ErrOrderNotFound)
}

func TestConfirmOrder_AlreadyConfirmed_ReturnsErrInvalidTransition(t *testing.T) {
	repo := newMockRepo()
	cart := cartWithItems(someItems()...)
	svc := makeService(repo, cart)

	created, _, _ := svc.CreateOrder(context.Background(), "u1", "key-double-confirm")
	_, _ = svc.ConfirmOrder(fmt.Sprint(created.ID))

	_, err := svc.ConfirmOrder(fmt.Sprint(created.ID))
	assert.ErrorIs(t, err, domain.ErrInvalidTransition)
}

// ── FailOrder ─────────────────────────────────────────────────────────────────

func TestFailOrder_Success(t *testing.T) {
	repo := newMockRepo()
	cart := cartWithItems(someItems()...)
	svc := makeService(repo, cart)

	created, _, _ := svc.CreateOrder(context.Background(), "u1", "key-fail")
	updated, err := svc.FailOrder(fmt.Sprint(created.ID))

	require.NoError(t, err)
	assert.Equal(t, domain.StatusFailed, updated.Status)
}

func TestFailOrder_AfterConfirm_ReturnsErrInvalidTransition(t *testing.T) {
	repo := newMockRepo()
	cart := cartWithItems(someItems()...)
	svc := makeService(repo, cart)

	created, _, _ := svc.CreateOrder(context.Background(), "u1", "key-confirm-then-fail")
	_, _ = svc.ConfirmOrder(fmt.Sprint(created.ID))

	_, err := svc.FailOrder(fmt.Sprint(created.ID))
	assert.ErrorIs(t, err, domain.ErrInvalidTransition)
}
