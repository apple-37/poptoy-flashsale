// pkg/fsm/product_fsm.go
package fsm

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

// NewProductFSM 创建商品状态机
func NewProductFSM(initState ProductState) *ProductFSM {
	engine := NewEngine[ProductState, ProductEvent](initState)
	
	// 注册默认转换规则
	engine.AddTransition(ProductStatePending, ProductEventApprove, ProductStateOnSale, nil)
	engine.AddTransition(ProductStateOnSale, ProductEventTakeOff, ProductStateOffSale, nil)
	engine.AddTransition(ProductStateOnSale, ProductEventSellOut, ProductStateSoldOut, nil)
	engine.AddTransition(ProductStateOffSale, ProductEventPutOn, ProductStateOnSale, nil)
	engine.AddTransition(ProductStateSoldOut, ProductEventRestock, ProductStateOnSale, nil)
	
	return &ProductFSM{
		engine: engine,
	}
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