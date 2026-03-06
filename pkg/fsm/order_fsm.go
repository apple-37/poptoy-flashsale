package fsm

import (
	"errors"
	"fmt"
)

type OrderState int8
type OrderEvent string

// 状态定义 (与 MySQL TINYINT 对应)
const (
	StateInit      OrderState = 0
	StatePending   OrderState = 1 // 待支付
	StatePaid      OrderState = 2 // 已支付
	StateCancelled OrderState = 3 // 已取消
)

// 事件定义
const (
	EventCreate  OrderEvent = "EventCreate"
	EventPay     OrderEvent = "EventPay"
	EventTimeout OrderEvent = "EventTimeout"
)

type ActionFunc func(orderNo string) error

type transition struct {
	NextState OrderState
	Action    ActionFunc
}

type OrderFSM struct {
	CurrentState OrderState
	transitions  map[OrderState]map[OrderEvent]transition
}

// NewOrderFSM 实例化并注册路由表
func NewOrderFSM(initState OrderState) *OrderFSM {
	fsm := &OrderFSM{
		CurrentState: initState,
		transitions:  make(map[OrderState]map[OrderEvent]transition),
	}
	return fsm
}

// AddTransition 注册状态流转规则
func (m *OrderFSM) AddTransition(from OrderState, event OrderEvent, to OrderState, action ActionFunc) {
	if m.transitions[from] == nil {
		m.transitions[from] = make(map[OrderEvent]transition)
	}
	m.transitions[from][event] = transition{NextState: to, Action: action}
}

// Trigger 触发事件
func (m *OrderFSM) Trigger(event OrderEvent, orderNo string) error {
	stateMap, ok := m.transitions[m.CurrentState]
	if !ok {
		return errors.New("当前状态下没有任何可执行的事件")
	}

	trans, ok := stateMap[event]
	if !ok {
		return fmt.Errorf("非法流转: 无法从状态 %d 执行事件 %s", m.CurrentState, event)
	}

	// 只有 Action 成功，状态才会流转
	if err := trans.Action(orderNo); err != nil {
		return fmt.Errorf("执行事件 %s 动作失败: %w", event, err)
	}

	m.CurrentState = trans.NextState
	return nil
}