package core

type exclusions []string

// func makeExclusions(in []string) exclusions {
// 	var e exclusions
// 	copy(e, in)
// 	sort.Strings(e)
// 	return e
// }

func (e exclusions) Has(name string) bool {
	// i := sort.SearchStrings(e, name)
	// return i < len(e) && e[i] == name
	for _, x := range e {
		if x == name {
			return true
		}
	}
	return false
}
