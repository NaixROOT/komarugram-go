package model

// GroupAlbums is a presentation projection. Durable messages retain individual
// IDs so edits, deletes and pages can change part of an album independently.
func GroupAlbums(messages []Message) []Message {
	out := make([]Message, 0, len(messages))
	for i := 0; i < len(messages); {
		first := messages[i]
		end := i + 1
		if first.GroupedID != 0 && first.Media != nil {
			for end < len(messages) && messages[end].GroupedID == first.GroupedID && messages[end].Key.ChatID == first.Key.ChatID && messages[end].Media != nil {
				end++
			}
		}
		if end-i > 1 {
			first.Attachments = append([]Message(nil), messages[i:end]...)
			// Include every ID/revision in the layout revision, including caption edits.
			hash := uint64(1469598103934665603)
			for _, m := range first.Attachments {
				hash = (hash ^ uint64(m.Key.MessageID)) * 1099511628211
				hash = (hash ^ m.ContentRevision) * 1099511628211
			}
			first.ContentRevision = hash & ((1 << 63) - 1)
		}
		out = append(out, first)
		i = end
	}
	return out
}
