package sendfiles

import "komarugram/internal/messenger/model"

// Permission follows uploadFile: music remains music even in document mode.
func Permission(f File, documents bool) model.SendKind {
	if f.Kind == KindMusic {
		return model.SendMusic
	}
	if documents {
		return model.SendFile
	}
	switch f.Kind {
	case KindPhoto:
		return model.SendPhoto
	case KindVideo:
		return model.SendVideo
	case KindAnimation:
		return model.SendGIF
	default:
		return model.SendFile
	}
}
