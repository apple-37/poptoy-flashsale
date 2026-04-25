// pkg/fsm/user_fsm.go
package fsm

// 用户状态定义
type UserState int8

const (
	UserStateRegistered UserState = 0 // 已注册
	UserStateActive     UserState = 1 // 活跃
	UserStateDisabled   UserState = 2 // 禁用
	UserStateDeleted    UserState = 3 // 已删除
)

// 用户事件定义
type UserEvent string

const (
	UserEventRegister   UserEvent = "Register"
	UserEventActivate   UserEvent = "Activate"
	UserEventDisable    UserEvent = "Disable"
	UserEventDelete     UserEvent = "Delete"
	UserEventReactivate UserEvent = "Reactivate"
)

// UserFSM 用户状态机
type UserFSM struct {
	engine *Engine[UserState, UserEvent]
}

// NewUserFSM 创建用户状态机
func NewUserFSM(initState UserState) *UserFSM {
	engine := NewEngine[UserState, UserEvent](initState)
	
	// 注册默认转换规则
	engine.AddTransition(UserStateRegistered, UserEventActivate, UserStateActive, nil)
	engine.AddTransition(UserStateActive, UserEventDisable, UserStateDisabled, nil)
	engine.AddTransition(UserStateActive, UserEventDelete, UserStateDeleted, nil)
	engine.AddTransition(UserStateDisabled, UserEventReactivate, UserStateActive, nil)
	engine.AddTransition(UserStateDisabled, UserEventDelete, UserStateDeleted, nil)
	
	return &UserFSM{
		engine: engine,
	}
}

// AddTransition 注册状态转换规则
func (f *UserFSM) AddTransition(from UserState, event UserEvent, to UserState, action func(userID uint64) error) {
	// 适配动作函数签名
	adaptedAction := func(ctx Context) error {
		if action == nil {
			return nil
		}
		userID, ok := ctx["userID"].(uint64)
		if !ok {
			return nil
		}
		return action(userID)
	}
	
	f.engine.AddTransition(from, event, to, adaptedAction)
}

// Trigger 触发事件
func (f *UserFSM) Trigger(event UserEvent, userID uint64) error {
	ctx := Context{
		"userID": userID,
		"key":    userID, // 用于持久化的键
	}
	return f.engine.Trigger(event, ctx)
}

// GetCurrentState 获取当前状态
func (f *UserFSM) GetCurrentState() UserState {
	return f.engine.GetCurrentState()
}