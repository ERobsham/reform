package parser

const (
	stdWrapperPrefixes            = "({[< "
	stdWrapperSuffixes            = ")}]> "
	stdWrapperOrSrcEndingSuffixes = ":" + stdWrapperSuffixes
	stdWrappers                   = "(){}[]<> "
)

var (
	stdWrapperPrefixMap = map[byte]struct{}{
		'(': {},
		'{': {},
		'[': {},
		'<': {},
		' ': {},
	}
	stdWrapperSuffixMap = map[byte]struct{}{
		')': {},
		'}': {},
		']': {},
		'>': {},
		' ': {},
	}

	stdWrapperOrSrcEndingSuffixMap = map[byte]struct{}{
		')': {},
		'}': {},
		']': {},
		'>': {},
		':': {},
		' ': {},
	}
)

func consumeNextSpace(line string, endIdx int) int {
	for {
		if len(line) > endIdx+2 && line[endIdx+1] == ' ' {
			endIdx += 1
		} else {
			break
		}
	}
	return endIdx
}

func consumeNextNumber(line string, idx int) int {
	return consumeNext(line, idx, isNumericChar)
}

// march idx forward, consuming closing wrappers and spaces
func consumeNextClosingWrapper(line string, idx int) int {
	var testFn testFunc = func(b byte) bool { return isInSet(b, stdWrapperSuffixMap) }
	return consumeNext(line, idx, testFn)
}

// march idx backwards, consuming closing wrappers and spaces
func consumePrevClosingWrapper(line string, idx int) int {
	var testFn testFunc = func(b byte) bool { return isInSet(b, stdWrapperSuffixMap) }
	return consumePrev(line, idx, testFn)
}

// march idx forwards, consuming opening wrappers and spaces
func consumeNextOpeningWrapper(line string, idx int) int {
	var testFn testFunc = func(b byte) bool { return isInSet(b, stdWrapperPrefixMap) }
	return consumeNext(line, idx, testFn)
}

// march idx backwards, consuming opening wrappers and spaces
func consumePrevOpeningWrapper(line string, idx int) int {
	var testFn testFunc = func(b byte) bool { return isInSet(b, stdWrapperPrefixMap) }
	return consumePrev(line, idx, testFn)
}

func consumeNextPathChars(line string, idx int) int {
	return consumeNext(line, idx, isValidPathCharacter)
}
func consumePrevPathChars(line string, idx int) int {
	return consumePrev(line, idx, isValidPathCharacter)
}

//
// generic
//

func consumePrev(line string, idx int, testFn testFunc) int {
	for {
		if idx == 0 {
			break
		}
		if !testFn(line[idx-1]) {
			break
		}

		idx -= 1
	}
	return idx
}
func consumeNext(line string, idx int, testFn testFunc) int {
	len := len(line)
	for {
		if !(len > idx+1) {
			idx = len
			break
		}
		if !testFn(line[idx]) {
			break
		}

		idx += 1
	}
	return idx
}

//
// test functions
//

type testFunc func(byte) bool

func isNumericChar(b byte) bool {
	const numbrRange = "09"
	return (b >= numbrRange[0] && b <= numbrRange[1])
}

func isAlphaChar(b byte) bool {
	const (
		upperRange = "AZ"
		lowerRange = "az"
	)
	return (b >= upperRange[0] && b <= upperRange[1]) ||
		(b >= lowerRange[0] && b <= lowerRange[1])
}

func isAlphanumeric(b byte) bool {
	return (isAlphaChar(b)) || (isNumericChar(b))
}

func isInSet(b byte, set map[byte]struct{}) bool {
	_, exists := set[b]
	return exists
}

func isValidPathCharacter(b byte) bool {
	nonAlphanumericPathCharacters := map[byte]struct{}{
		'.': {},
		'-': {},
		'_': {},
		'/': {},
	}

	return isAlphanumeric(b) || isInSet(b, nonAlphanumericPathCharacters)
}

func isValidRustCrateChar(b byte) bool {
	nonAlphanumericPathCharacters := map[byte]struct{}{
		'_': {},
	}

	return isAlphanumeric(b) || isInSet(b, nonAlphanumericPathCharacters)
}
