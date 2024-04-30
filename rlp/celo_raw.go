package rlp

// ReadKind exposes readKind so to enable introspection of rlp encoded data.
func ReadKind(buf []byte) (k Kind, tagsize, contentsize uint64, err error) {
	return readKind(buf)
}

// ReadNext reads the next RLP item from the given buffer and returns the remaining buffer,
// the kind of the item, the size of the tag, the size of the content, and any error encountered.
func ReadNext(buf []byte) (remaining []byte, kind Kind, tagsize, contentsize uint64, err error) {
	remaining = buf
	kind, tagsize, contentsize, err = ReadKind(remaining)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	remaining = remaining[tagsize:]
	if kind != List {
		remaining = remaining[contentsize:]
	}
	return
}
