package service

import (
	"fmt"

	"poptoy-flashsale/app/user/internal/repository"
	"poptoy-flashsale/pkg/fsm"
)

// InitFSM 在服务启动阶段注册用户状态机动作。
func InitFSM() {
	fsm.InitUserFSMActions(
		func(userID uint64) error {
			return updateUserStatusIfCurrent(userID, fsm.UserStateRegistered, fsm.UserStateActive)
		},
		func(userID uint64) error {
			return updateUserStatusIfCurrent(userID, fsm.UserStateActive, fsm.UserStateDisabled)
		},
		func(userID uint64) error {
			return updateUserStatusIfCurrentIn(userID, []fsm.UserState{fsm.UserStateActive, fsm.UserStateDisabled}, fsm.UserStateDeleted)
		},
		func(userID uint64) error {
			return updateUserStatusIfCurrent(userID, fsm.UserStateDisabled, fsm.UserStateActive)
		},
	)
}

func updateUserStatusIfCurrent(userID uint64, expectedState fsm.UserState, state fsm.UserState) error {
	updated, err := repository.UpdateUserStatusIfCurrent(userID, int8(expectedState), int8(state))
	if err != nil {
		return err
	}
	if !updated {
		return fmt.Errorf("用户状态已变更，流转跳过")
	}
	return nil
}

func updateUserStatusIfCurrentIn(userID uint64, expectedStates []fsm.UserState, state fsm.UserState) error {
	values := make([]int8, 0, len(expectedStates))
	for _, s := range expectedStates {
		values = append(values, int8(s))
	}

	updated, err := repository.UpdateUserStatusIfCurrentIn(userID, values, int8(state))
	if err != nil {
		return err
	}
	if !updated {
		return fmt.Errorf("用户状态已变更，流转跳过")
	}
	return nil
}
