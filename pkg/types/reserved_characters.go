package types

// ReservedCharacters contains the mapping of the percent-encoded form to the ASCII
// form of the reserved characters according to RFC 3986 §2.2 (plus ‘%’ itself).
var ReservedCharacters = map[string]rune{
	"%3A": ':',
	"%2F": '/',
	"%3F": '?',
	"%23": '#',
	"%5B": '[',
	"%5D": ']',
	"%40": '@',
	"%21": '!',
	"%24": '$',
	"%26": '&',
	"%27": '\'',
	"%28": '(',
	"%29": ')',
	"%2A": '*',
	"%2B": '+',
	"%2C": ',',
	"%3B": ';',
	"%3D": '=',
	"%25": '%',
}
