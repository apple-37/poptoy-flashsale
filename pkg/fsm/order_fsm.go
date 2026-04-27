// pkg/fsm/order_fsm.go
package fsm

import "fmt"

// 保持原有状态和事件定义
type OrderState int8
type OrderEvent string

const (
    StateInit      OrderState = 0
    StatePending   OrderState = 1 // 待支付
    StatePaid      OrderState = 2 // 已支付
    StateCancelled OrderState = 3 // 已取消
)

const (
    EventCreate  OrderEvent = "EventCreate"
    EventPay     OrderEvent = "EventPay"
    EventTimeout OrderEvent = "EventTimeout"
)

// OrderFSM 订单状态机（使用通用引擎）
type OrderFSM struct {
    engine *Engine[OrderState, OrderEvent]
}

var (
    orderCreateAction  ActionFunc
    orderTimeoutAction ActionFunc
)

type orderTransitionDef struct {
    next   OrderState
    action func() ActionFunc
}

var orderTransitionTable = map[OrderState]map[OrderEvent]orderTransitionDef{
    StateInit: {
        EventCreate: {next: StatePending, action: func() ActionFunc { return orderCreateAction }},
    },
    StatePending: {
        EventPay:     {next: StatePaid, action: func() ActionFunc { return nil }},
        EventTimeout: {next: StateCancelled, action: func() ActionFunc { return orderTimeoutAction }},
    },
}

// InitOrderFSMActions 在应用启动阶段注册订单 FSM 动作。
func InitOrderFSMActions(createAction ActionFunc, timeoutAction ActionFunc) {
    orderCreateAction = createAction
    orderTimeoutAction = timeoutAction
}

// TriggerOrderEvent 触发订单状态流转（无实例模式）。
// 返回 nextState，调用方可用于后续处理。
func TriggerOrderEvent(currentState OrderState, event OrderEvent, ctx Context) (OrderState, error) {
    stateMap, ok := orderTransitionTable[currentState]
    if !ok {
        return currentState, fmt.Errorf("当前状态下没有可执行的事件")
    }

    def, ok := stateMap[event]
    if !ok {
        return currentState, fmt.Errorf("非法流转: 无法从状态 %v 执行事件 %v", currentState, event)
    }

    if ctx == nil {
        ctx = Context{}
    }
    if _, ok := ctx["key"]; !ok {
        if orderNo, ok := ctx["orderNo"].(string); ok {
            ctx["key"] = orderNo
        }
    }

    action := def.action()
    if action != nil {
        if err := action(ctx); err != nil {
            return currentState, fmt.Errorf("执行事件 %v 动作失败: %w", event, err)
        }
    }

    return def.next, nil
}

// NewOrderFSM 创建订单状态机
func NewOrderFSM(initState OrderState) *OrderFSM {
    engine := NewEngine[OrderState, OrderEvent](initState)
    
    // 注册默认转换规则
    engine.AddTransition(StateInit, EventCreate, StatePending, orderCreateAction)
    engine.AddTransition(StatePending, EventPay, StatePaid, nil)
    engine.AddTransition(StatePending, EventTimeout, StateCancelled, orderTimeoutAction)
    
    return &OrderFSM{
        engine: engine,
    }
}

// TriggerWithContext 触发事件并传入完整上下文。
func (f *OrderFSM) TriggerWithContext(event OrderEvent, ctx Context) error {
    if ctx == nil {
        ctx = Context{}
    }
    if _, ok := ctx["key"]; !ok {
        if orderNo, ok := ctx["orderNo"].(string); ok {
            ctx["key"] = orderNo
        }
    }
    return f.engine.Trigger(event, ctx)
}

// WithPersister 添加持久化
func (f *OrderFSM) WithPersister(persister StatePersister[OrderState]) *OrderFSM {
    f.engine.WithPersister(persister)
    return f
}

// AddTransition 注册自定义转换规则
func (f *OrderFSM) AddTransition(from OrderState, event OrderEvent, to OrderState, action func(orderNo string) error) {
    // 适配旧的动作函数签名到新的 Context 格式
    adaptedAction := func(ctx Context) error {
        if action == nil {
            return nil
        }
        orderNo, ok := ctx["orderNo"].(string)
        if !ok {
            return nil
        }
        return action(orderNo)
    }
    
    f.engine.AddTransition(from, event, to, adaptedAction)
}

// Trigger 触发事件
func (f *OrderFSM) Trigger(event OrderEvent, orderNo string) error {
    ctx := Context{
        "orderNo": orderNo,
        "key":     orderNo, // 用于持久化的键
    }
    return f.TriggerWithContext(event, ctx)
}

// GetCurrentState 获取当前状态
func (f *OrderFSM) GetCurrentState() OrderState {
    return f.engine.GetCurrentState()
}

// GetHistory 获取状态变更历史
func (f *OrderFSM) GetHistory() []StateHistory[OrderState, OrderEvent] {
    return f.engine.GetHistory()
}