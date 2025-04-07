package hub

func newEvent(p any, topicArgs ...string) *event {
	return &event{
		payload: p,
		topic:   T(topicArgs...),
	}
}
