// pkg/fsm/order_fsm.go
package fsm

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

// NewOrderFSM 创建订单状态机
func NewOrderFSM(initState OrderState) *OrderFSM {
    engine := NewEngine[OrderState, OrderEvent](initState)
    
    // 注册默认转换规则
    engine.AddTransition(StateInit, EventCreate, StatePending, nil)
    engine.AddTransition(StatePending, EventPay, StatePaid, nil)
    engine.AddTransition(StatePending, EventTimeout, StateCancelled, nil)
    
    return &OrderFSM{
        engine: engine,
    }
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
    return f.engine.Trigger(event, ctx)
}

// GetCurrentState 获取当前状态
func (f *OrderFSM) GetCurrentState() OrderState {
    return f.engine.GetCurrentState()
}

// GetHistory 获取状态变更历史
func (f *OrderFSM) GetHistory() []StateHistory[OrderState, OrderEvent] {
    return f.engine.GetHistory()
}