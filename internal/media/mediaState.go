package media

type mediaState struct {
	lastPlayer string
	text       []rune
	head       int
	tail       int
}

func (state *mediaState) scroll() string {
	if len(state.text) > 25 {
		if state.head == state.tail {
			state.tail = 21
		}

		if state.tail+1 == len(state.text) {
			state.tail = -1
		}

		if state.head+1 == len(state.text) {
			state.head = -1
		}

		state.tail++
		state.head++

		if state.tail < state.head {
			return string(state.text[state.head:]) + "   " + string(state.text[:state.tail])
		}

		return string(state.text[state.head:state.tail]) + "..."
	}
	state.head = -1
	state.tail = -1
	return string(state.text)
}
