package helper

func Head(msg []byte) []byte {
	if len(msg) >= 10 {
		return msg[:10]
	}
	return msg
}
