package brignext

type TriggeringEvents struct {
	Source string   `json:"source" bson:"source"`
	Types  []string `json:"types" bson:"types"`
}

func (t *TriggeringEvents) Matches(eventSource, eventType string) bool {
	if t.Source == "" ||
		eventSource == "" ||
		eventType == "" ||
		t.Source != eventSource {
		return false
	}
	if len(t.Types) == 0 {
		return true
	}
	for _, tipe := range t.Types {
		if tipe == eventType {
			return true
		}
	}
	return false
}
