package logic

func ExactlyOne(terms ...bool) bool {
	if len(terms) == 0 {
		return false
	}
	if len(terms) == 1 {
		return terms[0]
	}
	if terms[0] && terms[1] {
		return false
	}
	terms[1] = terms[0] || terms[1]
	return ExactlyOne(terms[1:]...)
}

func AtMostOne(terms ...bool) bool {
	if len(terms) <= 1 {
		return true
	}
	if terms[0] && terms[1] {
		return false
	}
	terms[1] = terms[0] || terms[1]
	return AtMostOne(terms[1:]...)
}
