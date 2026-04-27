// pkg/fsm/user_fsm.go
package fsm

import "fmt"

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

var (
	userActivateAction   func(userID uint64) error
	userDisableAction    func(userID uint64) error
	userDeleteAction     func(userID uint64) error
	userReactivateAction func(userID uint64) error
)

type userTransitionDef struct {
	next   UserState
	action func() func(userID uint64) error
}

var userTransitionTable = map[UserState]map[UserEvent]userTransitionDef{
	UserStateRegistered: {
		UserEventActivate: {next: UserStateActive, action: func() func(userID uint64) error { return userActivateAction }},
	},
	UserStateActive: {
		UserEventDisable: {next: UserStateDisabled, action: func() func(userID uint64) error { return userDisableAction }},
		UserEventDelete:  {next: UserStateDeleted, action: func() func(userID uint64) error { return userDeleteAction }},
	},
	UserStateDisabled: {
		UserEventReactivate: {next: UserStateActive, action: func() func(userID uint64) error { return userReactivateAction }},
		UserEventDelete:     {next: UserStateDeleted, action: func() func(userID uint64) error { return userDeleteAction }},
	},
}

// InitUserFSMActions 在应用启动阶段注册用户 FSM 动作。
func InitUserFSMActions(
	activateAction func(userID uint64) error,
	disableAction func(userID uint64) error,
	deleteAction func(userID uint64) error,
	reactivateAction func(userID uint64) error,
) {
	userActivateAction = activateAction
	userDisableAction = disableAction
	userDeleteAction = deleteAction
	userReactivateAction = reactivateAction
}

// TriggerUserEvent 触发用户状态流转（无实例模式）。
func TriggerUserEvent(currentState UserState, event UserEvent, userID uint64) (UserState, error) {
	stateMap, ok := userTransitionTable[currentState]
	if !ok {
		return currentState, fmt.Errorf("当前状态下没有可执行的事件")
	}

	def, ok := stateMap[event]
	if !ok {
		return currentState, fmt.Errorf("非法流转: 无法从状态 %v 执行事件 %v", currentState, event)
	}

	action := def.action()
	if action != nil {
		if err := action(userID); err != nil {
			return currentState, fmt.Errorf("执行事件 %v 动作失败: %w", event, err)
		}
	}

	return def.next, nil
}

// NewUserFSM 创建用户状态机
func NewUserFSM(initState UserState) *UserFSM {
	engine := NewEngine[UserState, UserEvent](initState)
	
	// 注册默认转换规则
	f := &UserFSM{engine: engine}
	
	f.AddTransition(UserStateRegistered, UserEventActivate, UserStateActive, userActivateAction)
	f.AddTransition(UserStateActive, UserEventDisable, UserStateDisabled, userDisableAction)
	f.AddTransition(UserStateActive, UserEventDelete, UserStateDeleted, userDeleteAction)
	f.AddTransition(UserStateDisabled, UserEventReactivate, UserStateActive, userReactivateAction)
	f.AddTransition(UserStateDisabled, UserEventDelete, UserStateDeleted, userDeleteAction)
	
	return f
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