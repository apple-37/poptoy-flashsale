// pkg/fsm/product_fsm.go
package fsm

import "fmt"

// 商品状态定义
type ProductState int8

const (
	ProductStatePending   ProductState = 0 // 待审核
	ProductStateOnSale    ProductState = 1 // 上架
	ProductStateOffSale   ProductState = 2 // 下架
	ProductStateSoldOut   ProductState = 3 // 售罄
	ProductStatePreSale   ProductState = 4 // 预售
)

// 商品事件定义
type ProductEvent string

const (
	ProductEventSubmit   ProductEvent = "Submit"
	ProductEventApprove  ProductEvent = "Approve"
	ProductEventPutOn    ProductEvent = "PutOn"
	ProductEventTakeOff  ProductEvent = "TakeOff"
	ProductEventSellOut  ProductEvent = "SellOut"
	ProductEventRestock  ProductEvent = "Restock"
)

// ProductFSM 商品状态机
type ProductFSM struct {
	engine *Engine[ProductState, ProductEvent]
}

var (
	productApproveAction func(productID uint64) error
	productTakeOffAction func(productID uint64) error
	productSellOutAction func(productID uint64) error
	productPutOnAction   func(productID uint64) error
	productRestockAction func(productID uint64) error
)

type productTransitionDef struct {
	next   ProductState
	action func() func(productID uint64) error
}

var productTransitionTable = map[ProductState]map[ProductEvent]productTransitionDef{
	ProductStatePending: {
		ProductEventApprove: {next: ProductStateOnSale, action: func() func(productID uint64) error { return productApproveAction }},
	},
	ProductStateOnSale: {
		ProductEventTakeOff: {next: ProductStateOffSale, action: func() func(productID uint64) error { return productTakeOffAction }},
		ProductEventSellOut: {next: ProductStateSoldOut, action: func() func(productID uint64) error { return productSellOutAction }},
	},
	ProductStateOffSale: {
		ProductEventPutOn: {next: ProductStateOnSale, action: func() func(productID uint64) error { return productPutOnAction }},
	},
	ProductStateSoldOut: {
		ProductEventRestock: {next: ProductStateOnSale, action: func() func(productID uint64) error { return productRestockAction }},
	},
}

// InitProductFSMActions 在应用启动阶段注册商品 FSM 动作。
func InitProductFSMActions(
	approveAction func(productID uint64) error,
	takeOffAction func(productID uint64) error,
	sellOutAction func(productID uint64) error,
	putOnAction func(productID uint64) error,
	restockAction func(productID uint64) error,
) {
	productApproveAction = approveAction
	productTakeOffAction = takeOffAction
	productSellOutAction = sellOutAction
	productPutOnAction = putOnAction
	productRestockAction = restockAction
}

// TriggerProductEvent 触发商品状态流转（无实例模式）。
func TriggerProductEvent(currentState ProductState, event ProductEvent, productID uint64) (ProductState, error) {
	stateMap, ok := productTransitionTable[currentState]
	if !ok {
		return currentState, fmt.Errorf("当前状态下没有可执行的事件")
	}

	def, ok := stateMap[event]
	if !ok {
		return currentState, fmt.Errorf("非法流转: 无法从状态 %v 执行事件 %v", currentState, event)
	}

	action := def.action()
	if action != nil {
		if err := action(productID); err != nil {
			return currentState, fmt.Errorf("执行事件 %v 动作失败: %w", event, err)
		}
	}

	return def.next, nil
}

// NewProductFSM 创建商品状态机
func NewProductFSM(initState ProductState) *ProductFSM {
	engine := NewEngine[ProductState, ProductEvent](initState)
	
	// 注册默认转换规则
	f := &ProductFSM{engine: engine}
	
	f.AddTransition(ProductStatePending, ProductEventApprove, ProductStateOnSale, productApproveAction)
	f.AddTransition(ProductStateOnSale, ProductEventTakeOff, ProductStateOffSale, productTakeOffAction)
	f.AddTransition(ProductStateOnSale, ProductEventSellOut, ProductStateSoldOut, productSellOutAction)
	f.AddTransition(ProductStateOffSale, ProductEventPutOn, ProductStateOnSale, productPutOnAction)
	f.AddTransition(ProductStateSoldOut, ProductEventRestock, ProductStateOnSale, productRestockAction)
	
	return f
}

// AddTransition 注册状态转换规则
func (f *ProductFSM) AddTransition(from ProductState, event ProductEvent, to ProductState, action func(productID uint64) error) {
	// 适配动作函数签名
	adaptedAction := func(ctx Context) error {
		if action == nil {
			return nil
		}
		productID, ok := ctx["productID"].(uint64)
		if !ok {
			return nil
		}
		return action(productID)
	}
	
	f.engine.AddTransition(from, event, to, adaptedAction)
}

// Trigger 触发事件
func (f *ProductFSM) Trigger(event ProductEvent, productID uint64) error {
	ctx := Context{
		"productID": productID,
		"key":       productID, // 用于持久化的键
	}
	return f.engine.Trigger(event, ctx)
}

// GetCurrentState 获取当前状态
func (f *ProductFSM) GetCurrentState() ProductState {
	return f.engine.GetCurrentState()
}