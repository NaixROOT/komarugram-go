package sendfiles

import (
	"komarugram/internal/messenger/model"
	"testing"
)

func TestSendingPermissionFollowsUploadMode(t *testing.T) {
	for _, tc := range []struct {
		kind      Kind
		documents bool
		want      model.SendKind
	}{
		{KindPhoto, false, model.SendPhoto}, {KindPhoto, true, model.SendFile},
		{KindVideo, false, model.SendVideo}, {KindVideo, true, model.SendFile},
		{KindAnimation, false, model.SendGIF}, {KindAnimation, true, model.SendFile},
		{KindMusic, false, model.SendMusic}, {KindMusic, true, model.SendMusic},
		{KindFile, false, model.SendFile},
	} {
		if got := Permission(File{Kind: tc.kind}, tc.documents); got != tc.want {
			t.Errorf("%v documents=%v: %v want %v", tc.kind, tc.documents, got, tc.want)
		}
	}
}
