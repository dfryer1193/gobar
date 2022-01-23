package media

type mediaState struct {
	lastPlayer string
	text       string
	head       int
	tail       int
}

func (state *mediaState) scroll() string {
	if len(state.text) > 50 {
		if state.head == state.tail {
			state.tail = 46
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
			return state.text[state.head:] + "   " + state.text[:state.tail]
		}

		return state.text[state.head:state.tail] + "..."
	}
	state.head = -1
	state.tail = -1
	return state.text
}
