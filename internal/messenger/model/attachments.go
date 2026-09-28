package model

// WebPreview is supplied by Telegram. Clients never fetch its URL to build it.
type WebPreview struct {
	URL, DisplayURL, Site, Title, Description string
	Photo                                     *MessageMedia
}

// Gift is presentation metadata for a saved gift, not a chat-history message.
type Gift struct {
	ID, Title, Slug                                 string
	Number                                          int
	Unique, SenderHidden                            bool
	SenderID                                        int64
	SenderName                                      string
	Stars                                           int64
	Issued, Total                                   int
	Model, Symbol, Backdrop                         string
	ModelRarity, SymbolRarity, BackdropRarity       int
	CenterColor, EdgeColor, PatternColor, TextColor uint32
	HasBackdrop                                     bool
	Pattern                                         *MessageMedia
}
