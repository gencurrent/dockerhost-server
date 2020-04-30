// Just some useful utils for this project

package utils

// Difference between 2 string arrays
func Difference(a, b []string) []string {

	target := map[string]bool{}
	for _, x := range b {
		target[x] = true
	}

	result := []string{}
	for _, x := range a {
		if _, ok := target[x]; !ok {
			result = append(result, x)
		}
	}

	return result
}
