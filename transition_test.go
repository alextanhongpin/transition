package transition_test

import (
	"context"
	"errors"
	"testing"

	"github.com/alextanhongpin/transition"
)

type Order struct {
	Id      int
	Address string

	transition.Transition
}

func getStateMachine() *transition.StateMachine {
	var orderStateMachine = transition.New()

	orderStateMachine.Initial("draft")
	orderStateMachine.State("checkout")
	orderStateMachine.State("paid")
	orderStateMachine.State("processed")
	orderStateMachine.State("delivered")
	orderStateMachine.State("cancelled")
	orderStateMachine.State("paid_cancelled")

	orderStateMachine.Event("checkout").To("checkout").From("draft")
	orderStateMachine.Event("pay").To("paid").From("checkout")

	return orderStateMachine
}

func TestStateTransition(t *testing.T) {
	order := &Order{}
	ctx := context.TODO()

	if err := getStateMachine().Trigger(ctx, "checkout", order); err != nil {
		t.Errorf("should not raise any error when trigger event checkout")
	}

	if order.GetState() != "checkout" {
		t.Errorf("state doesn't changed to checkout")
	}
}

func TestGetLastStateChange(t *testing.T) {
	order := &Order{}
	ctx := context.TODO()

	if err := getStateMachine().Trigger(ctx, "checkout", order); err != nil {
		t.Errorf("should not raise any error when trigger event checkout")
	}

	if err := getStateMachine().Trigger(ctx, "pay", order); err != nil {
		t.Errorf("should not raise any error when trigger event checkout")
	}

	if order.GetState() != "paid" {
		t.Errorf("state doesn't changed to paid")
	}
}

func TestMultipleTransitionWithOneEvent(t *testing.T) {
	orderStateMachine := getStateMachine()
	cancellEvent := orderStateMachine.Event("cancel")
	cancellEvent.To("cancelled").From("draft", "checkout")
	cancellEvent.To("paid_cancelled").From("paid", "processed")
	ctx := context.TODO()

	unpaidOrder1 := &Order{}
	if err := orderStateMachine.Trigger(ctx, "cancel", unpaidOrder1); err != nil {
		t.Errorf("should not raise any error when trigger event cancel")
	}

	if unpaidOrder1.State != "cancelled" {
		t.Errorf("order status doesn't transitioned correctly")
	}

	unpaidOrder2 := &Order{}
	unpaidOrder2.State = "draft"
	if err := orderStateMachine.Trigger(ctx, "cancel", unpaidOrder2); err != nil {
		t.Errorf("should not raise any error when trigger event cancel")
	}

	if unpaidOrder2.State != "cancelled" {
		t.Errorf("order status doesn't transitioned correctly")
	}

	paidOrder := &Order{}
	paidOrder.State = "paid"
	if err := orderStateMachine.Trigger(ctx, "cancel", paidOrder); err != nil {
		t.Errorf("should not raise any error when trigger event cancel")
	}

	if paidOrder.State != "paid_cancelled" {
		t.Errorf("order status doesn't transitioned correctly")
	}
}

func TestStateCallbacks(t *testing.T) {
	orderStateMachine := getStateMachine()
	order := &Order{}
	ctx := context.TODO()

	address1 := "I'm an address should be set when enter checkout"
	address2 := "I'm an address should be set when exit checkout"
	orderStateMachine.State("checkout").Enter(func(ctx context.Context, order interface{}) error {
		order.(*Order).Address = address1
		return nil
	}).Exit(func(ctx context.Context, order interface{}) error {
		order.(*Order).Address = address2
		return nil
	})

	if err := orderStateMachine.Trigger(ctx, "checkout", order); err != nil {
		t.Errorf("should not raise any error when trigger event checkout")
	}

	if order.Address != address1 {
		t.Errorf("enter callback not triggered")
	}

	if err := orderStateMachine.Trigger(ctx, "pay", order); err != nil {
		t.Errorf("should not raise any error when trigger event pay")
	}

	if order.Address != address2 {
		t.Errorf("exit callback not triggered")
	}
}

func TestEventCallbacks(t *testing.T) {
	var (
		order                 = &Order{}
		orderStateMachine     = getStateMachine()
		prevState, afterState string
		ctx                   = context.TODO()
	)

	orderStateMachine.Event("checkout").To("checkout").From("draft").Before(func(ctx context.Context, order interface{}) error {
		prevState = order.(*Order).State
		return nil
	}).After(func(ctx context.Context, order interface{}) error {
		afterState = order.(*Order).State
		return nil
	})

	order.State = "draft"
	if err := orderStateMachine.Trigger(ctx, "checkout", order); err != nil {
		t.Errorf("should not raise any error when trigger event checkout")
	}

	if prevState != "draft" {
		t.Errorf("Before callback triggered after state change")
	}

	if afterState != "checkout" {
		t.Errorf("After callback triggered after state change")
	}
}

func TestTransitionOnEnterCallbackError(t *testing.T) {
	var (
		order             = &Order{}
		orderStateMachine = getStateMachine()
		ctx               = context.TODO()
	)

	orderStateMachine.State("checkout").Enter(func(ctx context.Context, order interface{}) (err error) {
		return errors.New("intentional error")
	})

	if err := orderStateMachine.Trigger(ctx, "checkout", order); err == nil {
		t.Errorf("should raise an intentional error")
	}

	if order.State != "draft" {
		t.Errorf("state transitioned on Enter callback error")
	}
}

func TestTransitionOnExitCallbackError(t *testing.T) {
	var (
		order             = &Order{}
		orderStateMachine = getStateMachine()
		ctx               = context.TODO()
	)

	orderStateMachine.State("checkout").Exit(func(ctx context.Context, order interface{}) (err error) {
		return errors.New("intentional error")
	})

	if err := orderStateMachine.Trigger(ctx, "checkout", order); err != nil {
		t.Errorf("should not raise error when checkout")
	}

	if err := orderStateMachine.Trigger(ctx, "pay", order); err == nil {
		t.Errorf("should raise an intentional error")
	}

	if order.State != "checkout" {
		t.Errorf("state transitioned on Enter callback error")
	}
}

func TestEventOnBeforeCallbackError(t *testing.T) {
	var (
		order             = &Order{}
		orderStateMachine = getStateMachine()
		ctx               = context.TODO()
	)

	orderStateMachine.Event("checkout").To("checkout").From("draft").Before(func(ctx context.Context, order interface{}) error {
		return errors.New("intentional error")
	})

	if err := orderStateMachine.Trigger(ctx, "checkout", order); err == nil {
		t.Errorf("should raise an intentional error")
	}

	if order.State != "draft" {
		t.Errorf("state transitioned on Enter callback error")
	}
}

func TestEventOnAfterCallbackError(t *testing.T) {
	var (
		order             = &Order{}
		orderStateMachine = getStateMachine()
		ctx               = context.TODO()
	)

	orderStateMachine.Event("checkout").To("checkout").From("draft").After(func(ctx context.Context, order interface{}) error {
		return errors.New("intentional error")
	})

	if err := orderStateMachine.Trigger(ctx, "checkout", order); err == nil {
		t.Errorf("should raise an intentional error")
	}

	if order.State != "draft" {
		t.Errorf("state transitioned on Enter callback error")
	}
}
