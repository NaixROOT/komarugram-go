// SPDX-License-Identifier: Unlicense OR MIT

package model

import (
	"context"
	"errors"
)

// ChatMembership is separate from posting rights: an administrator is also a member.
type ChatMembership struct{ Known, Left, Creator, Broadcast bool }

// LeavePlan describes the consequence that must be confirmed before leaving.
// A zero SuccessorID means Telegram did not nominate a successor.
type LeavePlan struct {
	ChatID              int64
	Creator, BasicGroup bool
	SuccessorID         int64
	SuccessorName       string
}

var (
	ErrJoinVerification = errors.New("joining requires bot verification in Telegram Desktop")
	ErrJoinRequested    = errors.New("join request sent")
	ErrLeavePlanChanged = errors.New("ownership changed; reopen the leave confirmation")
	ErrOwnerCannotLeave = errors.New("Telegram requires transferring ownership before leaving")
)

type MembershipStore interface {
	Membership(int64) ChatMembership
	JoinChat(context.Context, int64) error
	PrepareLeave(context.Context, int64) (LeavePlan, error)
	LeaveChat(context.Context, int64, LeavePlan) error
}
