// pkg/fsm/engine.go
package fsm

import (
    "errors"
    "fmt"
    "sync"
    "time"
)

// State 通用状态类型（使用泛型）
type State interface{}

// Event 通用事件类型（使用泛型）
type Event interface{}

// Context 上下文信息，传递给动作函数
type Context map[string]interface{}

// ActionFunc 通用动作函数类型
type ActionFunc func(ctx Context) error

// Transition 状态转换定义
type Transition[S State, E Event] struct {
    NextState S
    Action    ActionFunc
}

// Engine 通用状态机引擎
type Engine[S comparable, E comparable] struct {
    currentState S
    transitions  map[S]map[E]Transition[S, E]
    mu           sync.RWMutex
    history      []StateHistory[S, E]
    persister    StatePersister[S]
}

// StateHistory 状态变更历史
type StateHistory[S State, E Event] struct {
    State     S
    Event     E
    Timestamp time.Time
}

// StatePersister 状态持久化接口
type StatePersister[S State] interface {
    SaveState(key string, state S) error
    LoadState(key string) (S, error)
}

// NewEngine 创建新的状态机引擎
func NewEngine[S comparable, E comparable](initialState S) *Engine[S, E] {
    return &Engine[S, E]{
        currentState: initialState,
        transitions:  make(map[S]map[E]Transition[S, E]),
        history:      make([]StateHistory[S, E], 0),
    }
}


// WithPersister 添加状态持久化
func (e *Engine[S, E]) WithPersister(persister StatePersister[S]) *Engine[S, E] {
    e.persister = persister
    return e
}

// AddTransition 注册状态转换规则
func (e *Engine[S, E]) AddTransition(from S, event E, to S, action ActionFunc) {
    e.mu.Lock()
    defer e.mu.Unlock()
    
    if e.transitions[from] == nil {
        e.transitions[from] = make(map[E]Transition[S, E])
    }
    e.transitions[from][event] = Transition[S, E]{
        NextState: to,
        Action:    action,
    }
}

// Trigger 触发事件
func (e *Engine[S, E]) Trigger(event E, ctx Context) error {
    e.mu.Lock()
    currentState := e.currentState
    e.mu.Unlock()
    
    // 检查转换规则是否存在
    e.mu.RLock()
    stateMap, ok := e.transitions[currentState]
    e.mu.RUnlock()
    
    if !ok {
        return errors.New("当前状态下没有可执行的事件")
    }
    
    e.mu.RLock()
    trans, ok := stateMap[event]
    e.mu.RUnlock()
    
    if !ok {
        return fmt.Errorf("非法流转: 无法从状态 %v 执行事件 %v", currentState, event)
    }
    
    // 执行动作
    if trans.Action != nil {
        if err := trans.Action(ctx); err != nil {
            return fmt.Errorf("执行事件 %v 动作失败: %w", event, err)
        }
    }
    
    // 更新状态
    e.mu.Lock()
    e.currentState = trans.NextState
    newState := e.currentState
    // 记录历史
    e.history = append(e.history, StateHistory[S, E]{
        State:     newState,
        Event:     event,
        Timestamp: time.Now(),
    })
    e.mu.Unlock()
    
    // 持久化状态
    if e.persister != nil {
        key, ok := ctx["key"].(string)
        if ok {
            if err := e.persister.SaveState(key, newState); err != nil {
                // 记录错误但不影响状态变更
                fmt.Printf("状态持久化失败: %v\n", err)
            }
        }
    }
    
    return nil
}

// GetCurrentState 获取当前状态
func (e *Engine[S, E]) GetCurrentState() S {
    e.mu.RLock()
    defer e.mu.RUnlock()
    return e.currentState
}

// GetHistory 获取状态变更历史
func (e *Engine[S, E]) GetHistory() []StateHistory[S, E] {
    e.mu.RLock()
    defer e.mu.RUnlock()
    
    history := make([]StateHistory[S, E], len(e.history))
    copy(history, e.history)
    return history
}