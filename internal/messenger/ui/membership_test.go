// SPDX-License-Identifier: Unlicense OR MIT

package ui

import (
	"context"
	"slices"
	"strings"
	"testing"
	"time"

	"komarugram/internal/messenger/localization"
	"komarugram/internal/messenger/model"
)

type membershipTestSource struct {
	model.ConversationStore
	state  model.ChatMembership
	plan   model.LeavePlan
	leaves chan model.LeavePlan
}

func (s *membershipTestSource) Membership(int64) model.ChatMembership { return s.state }
func (s *membershipTestSource) PrepareLeave(context.Context, int64) (model.LeavePlan, error) {
	return s.plan, nil
}
func (s *membershipTestSource) JoinChat(context.Context, int64) error { return nil }
func (s *membershipTestSource) LeaveChat(_ context.Context, _ int64, p model.LeavePlan) error {
	s.leaves <- p
	return nil
}
func TestMembershipMenuVisibility(t *testing.T) {
	s := &membershipTestSource{state: model.ChatMembership{Known: true}}
	p := &chatPage{source: s, chat: 8, kind: model.KindChannel}
	if !slices.Contains(p.chatMenuActions(), chatMenuLeave) {
		t.Fatal("member cannot leave")
	}
	s.state.Left = true
	if slices.Contains(p.chatMenuActions(), chatMenuLeave) {
		t.Fatal("nonmember offered leave")
	}
	s.state.Left = false
	p.threadRoot = 1
	if slices.Contains(p.chatMenuActions(), chatMenuLeave) {
		t.Fatal("thread offered leave")
	}
}
func TestMembershipConfirmationBeforeLeave(t *testing.T) {
	plan := model.LeavePlan{ChatID: 8, Creator: true, SuccessorID: 55, SuccessorName: "Alice"}
	s := &membershipTestSource{state: model.ChatMembership{Known: true, Creator: true}, plan: plan, leaves: make(chan model.LeavePlan, 1)}
	wake := make(chan struct{}, 4)
	p := &chatPage{source: s, chat: 8, kind: model.KindChannel, invalidate: func() { wake <- struct{}{} }}
	p.askLeave()
	select {
	case <-wake:
	case <-time.After(time.Second):
		t.Fatal("prepare timed out")
	}
	p.updateMembership(8, localization.For("en"))
	if !p.membership.ready || p.membership.plan.SuccessorID != 55 {
		t.Fatal("warning not ready")
	}
	select {
	case <-s.leaves:
		t.Fatal("left without confirmation")
	default:
	}
	text := leaveExplanation(p.membership.plan, localization.For("en"))
	if !strings.Contains(text, "Alice") || !strings.Contains(text, "7 days") {
		t.Fatal(text)
	}
	plan.BasicGroup = true
	if !strings.Contains(leaveExplanation(plan, localization.For("en")), "immediately") {
		t.Fatal("basic-group transfer not immediate")
	}
	p.startMembership(8, membershipLeave)
	select {
	case sent := <-s.leaves:
		if sent != s.plan {
			t.Fatal("wrong plan confirmed")
		}
	case <-time.After(time.Second):
		t.Fatal("leave timed out")
	}
}

func TestLeaveNoticeUsesOperationKind(t *testing.T) {
	for _, channel := range []bool{true, false} {
		name := "group"
		want := "Вы вышли из группы"
		if channel {
			name = "channel"
			want = "Вы отписались от канала"
		}
		t.Run(name, func(t *testing.T) {
			s := &membershipTestSource{state: model.ChatMembership{Known: true, Broadcast: channel}, plan: model.LeavePlan{ChatID: 8}, leaves: make(chan model.LeavePlan, 1)}
			wake := make(chan struct{}, 4)
			p := &chatPage{source: s, chat: 8, kind: model.KindUser, invalidate: func() { wake <- struct{}{} }}
			p.membership.plan = s.plan
			p.startMembership(8, membershipLeave)
			select {
			case <-wake:
			case <-time.After(time.Second):
				t.Fatal("leave timed out")
			}
			p.updateMembership(8, localization.For("ru"))
			if got := p.toast.Text(); got != want {
				t.Fatalf("got %q, want %q", got, want)
			}
		})
	}
}

func TestLeaveNoticeOutlivesRemovedDialog(t *testing.T) {
	var notice string
	p := &chatPage{membershipNotice: func(text string) { notice = text }}
	p.membership = membershipControl{chat: 8, broadcast: true, busy: true, results: make(chan membershipResult, 1)}
	p.membership.results <- membershipResult{operation: membershipLeave}
	p.updateMembership(8, localization.For("ru"))
	if notice != "Вы отписались от канала" || p.toast.Text() != "" || p.membership.busy {
		t.Fatalf("leave outcome stuck in removed page: notice=%q toast=%q busy=%v", notice, p.toast.Text(), p.membership.busy)
	}
}
