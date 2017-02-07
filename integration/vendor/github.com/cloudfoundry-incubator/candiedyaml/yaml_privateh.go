/*
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package candiedyaml

const (
	INPUT_RAW_BUFFER_SIZE = 1024

	/*
	 * The size of the input buffer.
	 *
	 * It should be possible to decode the whole raw buffer.
	 */
	INPUT_BUFFER_SIZE = (INPUT_RAW_BUFFER_SIZE * 3)

	/*
	 * The size of the output buffer.
	 */

	OUTPUT_BUFFER_SIZE = 512

	/*
	 * The size of the output raw buffer.
	 *
	 * It should be possible to encode the whole output buffer.
	 */

	OUTPUT_RAW_BUFFER_SIZE = (OUTPUT_BUFFER_SIZE*2 + 2)

	INITIAL_STACK_SIZE = 16
	INITIAL_QUEUE_SIZE = 16
)

func width(b byte) int {
	if b&0x80 == 0 {
		return 1
	}

	if b&0xE0 == 0xC0 {
		return 2
	}

	if b&0xF0 == 0xE0 {
		return 3
	}

	if b&0xF8 == 0xF0 {
		return 4
	}

	return 0
}

func copy_bytes(dest []byte, dest_pos *int, src []byte, src_pos *int) {
	w := width(src[*src_pos])
	switch w {
	case 4:
		dest[*dest_pos+3] = src[*src_pos+3]
		fallthrough
	case 3:
		dest[*dest_pos+2] = src[*src_pos+2]
		fallthrough
	case 2:
		dest[*dest_pos+1] = src[*src_pos+1]
		fallthrough
	case 1:
		dest[*dest_pos] = src[*src_pos]
	default:
		panic("invalid width")
	}
	*dest_pos += w
	*src_pos += w
}

// /*
//  * Check if the character at the specified position is an alphabetical
//  * character, a digit, '_', or '-'.
//  */

func is_alpha(b byte) bool {
	return (b >= '0' && b <= '9') ||
		(b >= 'A' && b <= 'Z') ||
		(b >= 'a' && b <= 'z') ||
		b == '_' || b == '-'
}

// /*
//  * Check if the character at the specified position is a digit.
//  */
//
func is_digit(b byte) bool {
	return b >= '0' && b <= '9'
}

// /*
//  * Get the value of a digit.
//  */
//
func as_digit(b byte) int {
	return int(b) - '0'
}

// /*
//  * Check if the character at the specified position is a hex-digit.
//  */
//
func is_hex(b byte) bool {
	return (b >= '0' && b <= '9') ||
		(b >= 'A' && b <= 'F') ||
		(b >= 'a' && b <= 'f')
}

//
// /*
//  * Get the value of a hex-digit.
//  */
//
func as_hex(b byte) int {
	if b >= 'A' && b <= 'F' {
		return int(b) - 'A' + 10
	} else if b >= 'a' && b <= 'f' {
		return int(b) - 'a' + 10
	}
	return int(b) - '0'
}

// #define AS_HEX_AT(string,offset)                                                \
//       (((string).pointer[offset] >= (yaml_char_t) 'A' &&                        \
//         (string).pointer[offset] <= (yaml_char_t) 'F') ?                        \
//        ((string).pointer[offset] - (yaml_char_t) 'A' + 10) :                    \
//        ((string).pointer[offset] >= (yaml_char_t) 'a' &&                        \
//         (string).pointer[offset] <= (yaml_char_t) 'f') ?                        \
//        ((string).pointer[offset] - (yaml_char_t) 'a' + 10) :                    \
//        ((string).pointer[offset] - (yaml_char_t) '0'))

// /*
//  * Check if the character is a line break, space, tab, or NUL.
//  */
func is_blankz_at(b []byte, i int) bool {
	return is_blank(b[i]) || is_breakz_at(b, i)
}

// /*
//  * Check if the character at the specified position is a line break.
//  */
func is_break_at(b []byte, i int) bool {
	return b[i] == '\r' || /* CR (#xD)*/
		b[i] == '\n' || /* LF (#xA) */
		(b[i] == 0xC2 && b[i+1] == 0x85) || /* NEL (#x85) */
		(b[i] == 0xE2 && b[i+1] == 0x80 && b[i+2] == 0xA8) || /* LS (#x2028) */
		(b[i] == 0xE2 && b[i+1] == 0x80 && b[i+2] == 0xA9) /* PS (#x2029) */
}

func is_breakz_at(b []byte, i int) bool {
	return is_break_at(b, i) || is_z(b[i])
}

func is_crlf_at(b []byte, i int) bool {
	return b[i] == '\r' && b[i+1] == '\n'
}

// /*
//  * Check if the character at the specified position is NUL.
//  */
func is_z(b byte) bool {
	return b == 0x0
}

// /*
//  * Check if the character at the specified position is space.
//  */
func is_space(b byte) bool {
	return b == ' '
}

//
// /*
//  * Check if the character at the specified position is tab.
//  */
func is_tab(b byte) bool {
	return b == '\t'
}

// /*
//  * Check if the character at the specified position is blank (space or tab).
//  */
func is_blank(b byte) bool {
	return is_space(b) || is_tab(b)
}

// /*
//  * Check if the character is ASCII.
//  */
func is_ascii(b byte) bool {
	return b <= '\x7f'
}

// /*
//  * Check if the character can be printed unescaped.
//  */
func is_printable_at(b []byte, i int) bool {
	return ((b[i] == 0x0A) || /* . == #x0A */
		(b[i] >= 0x20 && b[i] <= 0x7E) || /* #x20 <= . <= #x7E */
		(b[i] == 0xC2 && b[i+1] >= 0xA0) || /* #0xA0 <= . <= #xD7FF */
		(b[i] > 0xC2 && b[i] < 0xED) ||
		(b[i] == 0xED && b[i+1] < 0xA0) ||
		(b[i] == 0xEE) ||
		(b[i] == 0xEF && /* && . != #xFEFF */
			!(b[i+1] == 0xBB && b[i+2] == 0xBF) &&
			!(b[i+1] == 0xBF && (b[i+2] == 0xBE || b[i+2] == 0xBF))))
}

func insert_token(parser *yaml_parser_t, pos int, token *yaml_token_t) {
	// collapse the slice
	if parser.tokens_head > 0 && len(parser.tokens) == cap(parser.tokens) {
		if parser.tokens_head != len(parser.tokens) {
			// move the tokens down
			copy(parser.tokens, parser.tokens[parser.tokens_head:])
		}
		// readjust the length
		parser.tokens = parser.tokens[:len(parser.tokens)-parser.tokens_head]
		parser.tokens_head = 0
	}

	parser.tokens = append(parser.tokens, *token)
	if pos < 0 {
		return
	}
	copy(parser.tokens[parser.tokens_head+pos+1:], parser.tokens[parser.tokens_head+pos:])
	parser.tokens[parser.tokens_head+pos] = *token
}

// /*
//  * Check if the character at the specified position is BOM.
//  */
//
func is_bom_at(b []byte, i int) bool {
	return b[i] == 0xEF && b[i+1] == 0xBB && b[i+2] == 0xBF
}

//
// #ifdef HAVE_CONFIG_H
// #include <config.h>
// #endif
//
// #include "./yaml.h"
//
// #include <assert.h>
// #include <limits.h>
//
// /*
//  * Memory management.
//  */
//
// yaml_DECLARE(void *)
// yaml_malloc(size_t size);
//
// yaml_DECLARE(void *)
// yaml_realloc(void *ptr, size_t size);
//
// yaml_DECLARE(void)
// yaml_free(void *ptr);
//
// yaml_DECLARE(yaml_char_t *)
// yaml_strdup(const yaml_char_t *);
//
// /*
//  * Reader: Ensure that the buffer contains at least `length` characters.
//  */
//
// yaml_DECLARE(int)
// yaml_parser_update_buffer(yaml_parser_t *parser, size_t length);
//
// /*
//  * Scanner: Ensure that the token stack contains at least one token ready.
//  */
//
// yaml_DECLARE(int)
// yaml_parser_fetch_more_tokens(yaml_parser_t *parser);
//
// /*
//  * The size of the input raw buffer.
//  */
//
// #define INPUT_RAW_BUFFER_SIZE   16384
//
// /*
//  * The size of the input buffer.
//  *
//  * It should be possible to decode the whole raw buffer.
//  */
//
// #define INPUT_BUFFER_SIZE       (INPUT_RAW_BUFFER_SIZE*3)
//
// /*
//  * The size of the output buffer.
//  */
//
// #define OUTPUT_BUFFER_SIZE      16384
//
// /*
//  * The size of the output raw buffer.
//  *
//  * It should be possible to encode the whole output buffer.
//  */
//
// #define OUTPUT_RAW_BUFFER_SIZE  (OUTPUT_BUFFER_SIZE*2+2)
//
// /*
//  * The size of other stacks and queues.
//  */
//
// #define INITIAL_STACK_SIZE  16
// #define INITIAL_QUEUE_SIZE  16
// #define INITIAL_STRING_SIZE 16
//
// /*
//  * Buffer management.
//  */
//
// #define BUFFER_INIT(context,buffer,size)                                        \
//     (((buffer).start = yaml_malloc(size)) ?                                     \
//         ((buffer).last = (buffer).pointer = (buffer).start,                     \
//          (buffer).end = (buffer).start+(size),                                  \
//          1) :                                                                   \
//         ((context)->error = yaml_MEMORY_ERROR,                                  \
//          0))
//
// #define BUFFER_DEL(context,buffer)                                              \
//     (yaml_free((buffer).start),                                                 \
//      (buffer).start = (buffer).pointer = (buffer).end = 0)
//
// /*
//  * String management.
//  */
//
// typedef struct {
//     yaml_char_t *start;
//     yaml_char_t *end;
//     yaml_char_t *pointer;
// } yaml_string_t;
//
// yaml_DECLARE(int)
// yaml_string_extend(yaml_char_t **start,
//         yaml_char_t **pointer, yaml_char_t **end);
//
// yaml_DECLARE(int)
// yaml_string_join(
//         yaml_char_t **a_start, yaml_char_t **a_pointer, yaml_char_t **a_end,
//         yaml_char_t **b_start, yaml_char_t **b_pointer, yaml_char_t **b_end);
//
// #define NULL_STRING { NULL, NULL, NULL }
//
// #define STRING(string,length)   { (string), (string)+(length), (string) }
//
// #define STRING_ASSIGN(value,string,length)                                      \
//     ((value).start = (string),                                                  \
//      (value).end = (string)+(length),                                           \
//      (value).pointer = (string))
//
// #define STRING_INIT(context,string,size)                                        \
//     (((string).start = yaml_malloc(size)) ?                                     \
//         ((string).pointer = (string).start,                                     \
//          (string).end = (string).start+(size),                                  \
//          memset((string).start, 0, (size)),                                     \
//          1) :                                                                   \
//         ((context)->error = yaml_MEMORY_ERROR,                                  \
//          0))
//
// #define STRING_DEL(context,string)                                              \
//     (yaml_free((string).start),                                                 \
//      (string).start = (string).pointer = (string).end = 0)
//
// #define STRING_EXTEND(context,string)                                           \
//     (((string).pointer+5 < (string).end)                                        \
//         || yaml_string_extend(&(string).start,                                  \
//             &(string).pointer, &(string).end))
//
// #define CLEAR(context,string)                                                   \
//     ((string).pointer = (string).start,                                         \
//      memset((string).start, 0, (string).end-(string).start))
//
// #define JOIN(context,string_a,string_b)                                         \
//     ((yaml_string_join(&(string_a).start, &(string_a).pointer,                  \
//                        &(string_a).end, &(string_b).start,                      \
//                        &(string_b).pointer, &(string_b).end)) ?                 \
//         ((string_b).pointer = (string_b).start,                                 \
//          1) :                                                                   \
//         ((context)->error = yaml_MEMORY_ERROR,                                  \
//          0))
//
// /*
//  * String check operations.
//  */
//
// /*
//  * Check the octet at the specified position.
//  */
//
// #define CHECK_AT(string,octet,offset)                                           \
//     ((string).pointer[offset] == (yaml_char_t)(octet))
//
// /*
//  * Check the current octet in the buffer.
//  */
//
// #define CHECK(string,octet) CHECK_AT((string),(octet),0)
//
// /*
//  * Check if the character at the specified position is an alphabetical
//  * character, a digit, '_', or '-'.
//  */
//
// #define IS_ALPHA_AT(string,offset)                                              \
//      (((string).pointer[offset] >= (yaml_char_t) '0' &&                         \
//        (string).pointer[offset] <= (yaml_char_t) '9') ||                        \
//       ((string).pointer[offset] >= (yaml_char_t) 'A' &&                         \
//        (string).pointer[offset] <= (yaml_char_t) 'Z') ||                        \
//       ((string).pointer[offset] >= (yaml_char_t) 'a' &&                         \
//        (string).pointer[offset] <= (yaml_char_t) 'z') ||                        \
//       (string).pointer[offset] == '_' ||                                        \
//       (string).pointer[offset] == '-')
//
// #define IS_ALPHA(string)    IS_ALPHA_AT((string),0)
//
// /*
//  * Check if the character at the specified position is a digit.
//  */
//
// #define IS_DIGIT_AT(string,offset)                                              \
//      (((string).pointer[offset] >= (yaml_char_t) '0' &&                         \
//        (string).pointer[offset] <= (yaml_char_t) '9'))
//
// #define IS_DIGIT(string)    IS_DIGIT_AT((string),0)
//
// /*
//  * Get the value of a digit.
//  */
//
// #define AS_DIGIT_AT(string,offset)                                              \
//      ((string).pointer[offset] - (yaml_char_t) '0')
//
// #define AS_DIGIT(string)    AS_DIGIT_AT((string),0)
//
// /*
//  * Check if the character at the specified position is a hex-digit.
//  */
//
// #define IS_HEX_AT(string,offset)                                                \
//      (((string).pointer[offset] >= (yaml_char_t) '0' &&                         \
//        (string).pointer[offset] <= (yaml_char_t) '9') ||                        \
//       ((string).pointer[offset] >= (yaml_char_t) 'A' &&                         \
//        (string).pointer[offset] <= (yaml_char_t) 'F') ||                        \
//       ((string).pointer[offset] >= (yaml_char_t) 'a' &&                         \
//        (string).pointer[offset] <= (yaml_char_t) 'f'))
//
// #define IS_HEX(string)    IS_HEX_AT((string),0)
//
// /*
//  * Get the value of a hex-digit.
//  */
//
// #define AS_HEX_AT(string,offset)                                                \
//       (((string).pointer[offset] >= (yaml_char_t) 'A' &&                        \
//         (string).pointer[offset] <= (yaml_char_t) 'F') ?                        \
//        ((string).pointer[offset] - (yaml_char_t) 'A' + 10) :                    \
//        ((string).pointer[offset] >= (yaml_char_t) 'a' &&                        \
//         (string).pointer[offset] <= (yaml_char_t) 'f') ?                        \
//        ((string).pointer[offset] - (yaml_char_t) 'a' + 10) :                    \
//        ((string).pointer[offset] - (yaml_char_t) '0'))
//
// #define AS_HEX(string)  AS_HEX_AT((string),0)
//
// /*
//  * Check if the character is ASCII.
//  */
//
// #define IS_ASCII_AT(string,offset)                                              \
//     ((string).pointer[offset] <= (yaml_char_t) '\x7F')
//
// #define IS_ASCII(string)    IS_ASCII_AT((string),0)
//
// /*
//  * Check if the character can be printed unescaped.
//  */
//
// #define IS_PRINTABLE_AT(string,offset)                                          \
//     (((string).pointer[offset] == 0x0A)         /* . == #x0A */                 \
//      || ((string).pointer[offset] >= 0x20       /* #x20 <= . <= #x7E */         \
//          && (string).pointer[offset] <= 0x7E)                                   \
//      || ((string).pointer[offset] == 0xC2       /* #0xA0 <= . <= #xD7FF */      \
//          && (string).pointer[offset+1] >= 0xA0)                                 \
//      || ((string).pointer[offset] > 0xC2                                        \
//          && (string).pointer[offset] < 0xED)                                    \
//      || ((string).pointer[offset] == 0xED                                       \
//          && (string).pointer[offset+1] < 0xA0)                                  \
//      || ((string).pointer[offset] == 0xEE)                                      \
//      || ((string).pointer[offset] == 0xEF      /* #xE000 <= . <= #xFFFD */      \
//          && !((string).pointer[offset+1] == 0xBB        /* && . != #xFEFF */    \
//              && (string).pointer[offset+2] == 0xBF)                             \
//          && !((string).pointer[offset+1] == 0xBF                                \
//              && ((string).pointer[offset+2] == 0xBE                             \
//                  || (string).pointer[offset+2] == 0xBF))))
//
// #define IS_PRINTABLE(string)    IS_PRINTABLE_AT((string),0)
//
// /*
//  * Check if the character at the specified position is NUL.
//  */
//
// #define IS_Z_AT(string,offset)    CHECK_AT((string),'\0',(offset))
//
// #define IS_Z(string)    IS_Z_AT((string),0)
//
// /*
//  * Check if the character at the specified position is BOM.
//  */
//
// #define IS_BOM_AT(string,offset)                                                \
//      (CHECK_AT((string),'\xEF',(offset))                                        \
//       && CHECK_AT((string),'\xBB',(offset)+1)                                   \
//       && CHECK_AT((string),'\xBF',(offset)+2))  /* BOM (#xFEFF) */
//
// #define IS_BOM(string)  IS_BOM_AT(string,0)
//
// /*
//  * Check if the character at the specified position is space.
//  */
//
// #define IS_SPACE_AT(string,offset)  CHECK_AT((string),' ',(offset))
//
// #define IS_SPACE(string)    IS_SPACE_AT((string),0)
//
// /*
//  * Check if the character at the specified position is tab.
//  */
//
// #define IS_TAB_AT(string,offset)    CHECK_AT((string),'\t',(offset))
//
// #define IS_TAB(string)  IS_TAB_AT((string),0)
//
// /*
//  * Check if the character at the specified position is blank (space or tab).
//  */
//
// #define IS_BLANK_AT(string,offset)                                              \
//     (IS_SPACE_AT((string),(offset)) || IS_TAB_AT((string),(offset)))
//
// #define IS_BLANK(string)    IS_BLANK_AT((string),0)
//
// /*
//  * Check if the character at the specified position is a line break.
//  */
//
// #define IS_BREAK_AT(string,offset)                                              \
//     (CHECK_AT((string),'\r',(offset))               /* CR (#xD)*/               \
//      || CHECK_AT((string),'\n',(offset))            /* LF (#xA) */              \
//      || (CHECK_AT((string),'\xC2',(offset))                                     \
//          && CHECK_AT((string),'\x85',(offset)+1))   /* NEL (#x85) */            \
//      || (CHECK_AT((string),'\xE2',(offset))                                     \
//          && CHECK_AT((string),'\x80',(offset)+1)                                \
//          && CHECK_AT((string),'\xA8',(offset)+2))   /* LS (#x2028) */           \
//      || (CHECK_AT((string),'\xE2',(offset))                                     \
//          && CHECK_AT((string),'\x80',(offset)+1)                                \
//          && CHECK_AT((string),'\xA9',(offset)+2)))  /* PS (#x2029) */
//
// #define IS_BREAK(string)    IS_BREAK_AT((string),0)
//
// #define IS_CRLF_AT(string,offset)                                               \
//      (CHECK_AT((string),'\r',(offset)) && CHECK_AT((string),'\n',(offset)+1))
//
// #define IS_CRLF(string) IS_CRLF_AT((string),0)
//
// /*
//  * Check if the character is a line break or NUL.
//  */
//
// #define IS_BREAKZ_AT(string,offset)                                             \
//     (IS_BREAK_AT((string),(offset)) || IS_Z_AT((string),(offset)))
//
// #define IS_BREAKZ(string)   IS_BREAKZ_AT((string),0)
//
// /*
//  * Check if the character is a line break, space, or NUL.
//  */
//
// #define IS_SPACEZ_AT(string,offset)                                             \
//     (IS_SPACE_AT((string),(offset)) || IS_BREAKZ_AT((string),(offset)))
//
// #define IS_SPACEZ(string)   IS_SPACEZ_AT((string),0)
//
// /*
//  * Check if the character is a line break, space, tab, or NUL.
//  */
//
// #define IS_BLANKZ_AT(string,offset)                                             \
//     (IS_BLANK_AT((string),(offset)) || IS_BREAKZ_AT((string),(offset)))
//
// #define IS_BLANKZ(string)   IS_BLANKZ_AT((string),0)
//
// /*
//  * Determine the width of the character.
//  */
//
// #define WIDTH_AT(string,offset)                                                 \
//      (((string).pointer[offset] & 0x80) == 0x00 ? 1 :                           \
//       ((string).pointer[offset] & 0xE0) == 0xC0 ? 2 :                           \
//       ((string).pointer[offset] & 0xF0) == 0xE0 ? 3 :                           \
//       ((string).pointer[offset] & 0xF8) == 0xF0 ? 4 : 0)
//
// #define WIDTH(string)   WIDTH_AT((string),0)
//
// /*
//  * Move the string pointer to the next character.
//  */
//
// #define MOVE(string)    ((string).pointer += WIDTH((string)))
//
// /*
//  * Copy a character and move the pointers of both strings.
//  */
//
// #define COPY(string_a,string_b)                                                 \
//     ((*(string_b).pointer & 0x80) == 0x00 ?                                     \
//      (*((string_a).pointer++) = *((string_b).pointer++)) :                      \
//      (*(string_b).pointer & 0xE0) == 0xC0 ?                                     \
//      (*((string_a).pointer++) = *((string_b).pointer++),                        \
//       *((string_a).pointer++) = *((string_b).pointer++)) :                      \
//      (*(string_b).pointer & 0xF0) == 0xE0 ?                                     \
//      (*((string_a).pointer++) = *((string_b).pointer++),                        \
//       *((string_a).pointer++) = *((string_b).pointer++),                        \
//       *((string_a).pointer++) = *((string_b).pointer++)) :                      \
//      (*(string_b).pointer & 0xF8) == 0xF0 ?                                     \
//      (*((string_a).pointer++) = *((string_b).pointer++),                        \
//       *((string_a).pointer++) = *((string_b).pointer++),                        \
//       *((string_a).pointer++) = *((string_b).pointer++),                        \
//       *((string_a).pointer++) = *((string_b).pointer++)) : 0)
//
// /*
//  * Stack and queue management.
//  */
//
// yaml_DECLARE(int)
// yaml_stack_extend(void **start, void **top, void **end);
//
// yaml_DECLARE(int)
// yaml_queue_extend(void **start, void **head, void **tail, void **end);
//
// #define STACK_INIT(context,stack,size)                                          \
//     (((stack).start = yaml_malloc((size)*sizeof(*(stack).start))) ?             \
//         ((stack).top = (stack).start,                                           \
//          (stack).end = (stack).start+(size),                                    \
//          1) :                                                                   \
//         ((context)->error = yaml_MEMORY_ERROR,                                  \
//          0))
//
// #define STACK_DEL(context,stack)                                                \
//     (yaml_free((stack).start),                                                  \
//      (stack).start = (stack).top = (stack).end = 0)
//
// #define STACK_EMPTY(context,stack)                                              \
//     ((stack).start == (stack).top)
//
// #define PUSH(context,stack,value)                                               \
//     (((stack).top != (stack).end                                                \
//       || yaml_stack_extend((void **)&(stack).start,                             \
//               (void **)&(stack).top, (void **)&(stack).end)) ?                  \
//         (*((stack).top++) = value,                                              \
//          1) :                                                                   \
//         ((context)->error = yaml_MEMORY_ERROR,                                  \
//          0))
//
// #define POP(context,stack)                                                      \
//     (*(--(stack).top))
//
// #define QUEUE_INIT(context,queue,size)                                          \
//     (((queue).start = yaml_malloc((size)*sizeof(*(queue).start))) ?             \
//         ((queue).head = (queue).tail = (queue).start,                           \
//          (queue).end = (queue).start+(size),                                    \
//          1) :                                                                   \
//         ((context)->error = yaml_MEMORY_ERROR,                                  \
//          0))
//
// #define QUEUE_DEL(context,queue)                                                \
//     (yaml_free((queue).start),                                                  \
//      (queue).start = (queue).head = (queue).tail = (queue).end = 0)
//
// #define QUEUE_EMPTY(context,queue)                                              \
//     ((queue).head == (queue).tail)
//
// #define ENQUEUE(context,queue,value)                                            \
//     (((queue).tail != (queue).end                                               \
//       || yaml_queue_extend((void **)&(queue).start, (void **)&(queue).head,     \
//             (void **)&(queue).tail, (void **)&(queue).end)) ?                   \
//         (*((queue).tail++) = value,                                             \
//          1) :                                                                   \
//         ((context)->error = yaml_MEMORY_ERROR,                                  \
//          0))
//
// #define DEQUEUE(context,queue)                                                  \
//     (*((queue).head++))
//
// #define QUEUE_INSERT(context,queue,index,value)                                 \
//     (((queue).tail != (queue).end                                               \
//       || yaml_queue_extend((void **)&(queue).start, (void **)&(queue).head,     \
//             (void **)&(queue).tail, (void **)&(queue).end)) ?                   \
//         (memmove((queue).head+(index)+1,(queue).head+(index),                   \
//             ((queue).tail-(queue).head-(index))*sizeof(*(queue).start)),        \
//          *((queue).head+(index)) = value,                                       \
//          (queue).tail++,                                                        \
//          1) :                                                                   \
//         ((context)->error = yaml_MEMORY_ERROR,                                  \
//          0))
//
// /*
//  * Token initializers.
//  */
//
// #define TOKEN_INIT(token,token_type,token_start_mark,token_end_mark)            \
//     (memset(&(token), 0, sizeof(yaml_token_t)),                                 \
//      (token).type = (token_type),                                               \
//      (token).start_mark = (token_start_mark),                                   \
//      (token).end_mark = (token_end_mark))
//
// #define STREAM_START_TOKEN_INIT(token,token_encoding,start_mark,end_mark)       \
//     (TOKEN_INIT((token),yaml_STREAM_START_TOKEN,(start_mark),(end_mark)),       \
//      (token).data.stream_start.encoding = (token_encoding))
//
// #define STREAM_END_TOKEN_INIT(token,start_mark,end_mark)                        \
//     (TOKEN_INIT((token),yaml_STREAM_END_TOKEN,(start_mark),(end_mark)))
//
// #define ALIAS_TOKEN_INIT(token,token_value,start_mark,end_mark)                 \
//     (TOKEN_INIT((token),yaml_ALIAS_TOKEN,(start_mark),(end_mark)),              \
//      (token).data.alias.value = (token_value))
//
// #define ANCHOR_TOKEN_INIT(token,token_value,start_mark,end_mark)                \
//     (TOKEN_INIT((token),yaml_ANCHOR_TOKEN,(start_mark),(end_mark)),             \
//      (token).data.anchor.value = (token_value))
//
// #define TAG_TOKEN_INIT(token,token_handle,token_suffix,start_mark,end_mark)     \
//     (TOKEN_INIT((token),yaml_TAG_TOKEN,(start_mark),(end_mark)),                \
//      (token).data.tag.handle = (token_handle),                                  \
//      (token).data.tag.suffix = (token_suffix))
//
// #define SCALAR_TOKEN_INIT(token,token_value,token_length,token_style,start_mark,end_mark)   \
//     (TOKEN_INIT((token),yaml_SCALAR_TOKEN,(start_mark),(end_mark)),             \
//      (token).data.scalar.value = (token_value),                                 \
//      (token).data.scalar.length = (token_length),                               \
//      (token).data.scalar.style = (token_style))
//
// #define VERSION_DIRECTIVE_TOKEN_INIT(token,token_major,token_minor,start_mark,end_mark)     \
//     (TOKEN_INIT((token),yaml_VERSION_DIRECTIVE_TOKEN,(start_mark),(end_mark)),  \
//      (token).data.version_directive.major = (token_major),                      \
//      (token).data.version_directive.minor = (token_minor))
//
// #define TAG_DIRECTIVE_TOKEN_INIT(token,token_handle,token_prefix,start_mark,end_mark)       \
//     (TOKEN_INIT((token),yaml_TAG_DIRECTIVE_TOKEN,(start_mark),(end_mark)),      \
//      (token).data.tag_directive.handle = (token_handle),                        \
//      (token).data.tag_directive.prefix = (token_prefix))
//
// /*
//  * Event initializers.
//  */
//
// #define EVENT_INIT(event,event_type,event_start_mark,event_end_mark)            \
//     (memset(&(event), 0, sizeof(yaml_event_t)),                                 \
//      (event).type = (event_type),                                               \
//      (event).start_mark = (event_start_mark),                                   \
//      (event).end_mark = (event_end_mark))
//
// #define STREAM_START_EVENT_INIT(event,event_encoding,start_mark,end_mark)       \
//     (EVENT_INIT((event),yaml_STREAM_START_EVENT,(start_mark),(end_mark)),       \
//      (event).data.stream_start.encoding = (event_encoding))
//
// #define STREAM_END_EVENT_INIT(event,start_mark,end_mark)                        \
//     (EVENT_INIT((event),yaml_STREAM_END_EVENT,(start_mark),(end_mark)))
//
// #define DOCUMENT_START_EVENT_INIT(event,event_version_directive,                \
//         event_tag_directives_start,event_tag_directives_end,event_implicit,start_mark,end_mark) \
//     (EVENT_INIT((event),yaml_DOCUMENT_START_EVENT,(start_mark),(end_mark)),     \
//      (event).data.document_start.version_directive = (event_version_directive), \
//      (event).data.document_start.tag_directives.start = (event_tag_directives_start),   \
//      (event).data.document_start.tag_directives.end = (event_tag_directives_end),   \
//      (event).data.document_start.implicit = (event_implicit))
//
// #define DOCUMENT_END_EVENT_INIT(event,event_implicit,start_mark,end_mark)       \
//     (EVENT_INIT((event),yaml_DOCUMENT_END_EVENT,(start_mark),(end_mark)),       \
//      (event).data.document_end.implicit = (event_implicit))
//
// #define ALIAS_EVENT_INIT(event,event_anchor,start_mark,end_mark)                \
//     (EVENT_INIT((event),yaml_ALIAS_EVENT,(start_mark),(end_mark)),              \
//      (event).data.alias.anchor = (event_anchor))
//
// #define SCALAR_EVENT_INIT(event,event_anchor,event_tag,event_value,event_length,    \
//         event_plain_implicit, event_quoted_implicit,event_style,start_mark,end_mark)    \
//     (EVENT_INIT((event),yaml_SCALAR_EVENT,(start_mark),(end_mark)),             \
//      (event).data.scalar.anchor = (event_anchor),                               \
//      (event).data.scalar.tag = (event_tag),                                     \
//      (event).data.scalar.value = (event_value),                                 \
//      (event).data.scalar.length = (event_length),                               \
//      (event).data.scalar.plain_implicit = (event_plain_implicit),               \
//      (event).data.scalar.quoted_implicit = (event_quoted_implicit),             \
//      (event).data.scalar.style = (event_style))
//
// #define SEQUENCE_START_EVENT_INIT(event,event_anchor,event_tag,                 \
//         event_implicit,event_style,start_mark,end_mark)                         \
//     (EVENT_INIT((event),yaml_SEQUENCE_START_EVENT,(start_mark),(end_mark)),     \
//      (event).data.sequence_start.anchor = (event_anchor),                       \
//      (event).data.sequence_start.tag = (event_tag),                             \
//      (event).data.sequence_start.implicit = (event_implicit),                   \
//      (event).data.sequence_start.style = (event_style))
//
// #define SEQUENCE_END_EVENT_INIT(event,start_mark,end_mark)                      \
//     (EVENT_INIT((event),yaml_SEQUENCE_END_EVENT,(start_mark),(end_mark)))
//
// #define MAPPING_START_EVENT_INIT(event,event_anchor,event_tag,                  \
//         event_implicit,event_style,start_mark,end_mark)                         \
//     (EVENT_INIT((event),yaml_MAPPING_START_EVENT,(start_mark),(end_mark)),      \
//      (event).data.mapping_start.anchor = (event_anchor),                        \
//      (event).data.mapping_start.tag = (event_tag),                              \
//      (event).data.mapping_start.implicit = (event_implicit),                    \
//      (event).data.mapping_start.style = (event_style))
//
// #define MAPPING_END_EVENT_INIT(event,start_mark,end_mark)                       \
//     (EVENT_INIT((event),yaml_MAPPING_END_EVENT,(start_mark),(end_mark)))
//
// /*
//  * Document initializer.
//  */
//
// #define DOCUMENT_INIT(document,document_nodes_start,document_nodes_end,         \
//         document_version_directive,document_tag_directives_start,               \
//         document_tag_directives_end,document_start_implicit,                    \
//         document_end_implicit,document_start_mark,document_end_mark)            \
//     (memset(&(document), 0, sizeof(yaml_document_t)),                           \
//      (document).nodes.start = (document_nodes_start),                           \
//      (document).nodes.end = (document_nodes_end),                               \
//      (document).nodes.top = (document_nodes_start),                             \
//      (document).version_directive = (document_version_directive),               \
//      (document).tag_directives.start = (document_tag_directives_start),         \
//      (document).tag_directives.end = (document_tag_directives_end),             \
//      (document).start_implicit = (document_start_implicit),                     \
//      (document).end_implicit = (document_end_implicit),                         \
//      (document).start_mark = (document_start_mark),                             \
//      (document).end_mark = (document_end_mark))
//
// /*
//  * Node initializers.
//  */
//
// #define NODE_INIT(node,node_type,node_tag,node_start_mark,node_end_mark)        \
//     (memset(&(node), 0, sizeof(yaml_node_t)),                                   \
//      (node).type = (node_type),                                                 \
//      (node).tag = (node_tag),                                                   \
//      (node).start_mark = (node_start_mark),                                     \
//      (node).end_mark = (node_end_mark))
//
// #define SCALAR_NODE_INIT(node,node_tag,node_value,node_length,                  \
//         node_style,start_mark,end_mark)                                         \
//     (NODE_INIT((node),yaml_SCALAR_NODE,(node_tag),(start_mark),(end_mark)),     \
//      (node).data.scalar.value = (node_value),                                   \
//      (node).data.scalar.length = (node_length),                                 \
//      (node).data.scalar.style = (node_style))
//
// #define SEQUENCE_NODE_INIT(node,node_tag,node_items_start,node_items_end,       \
//         node_style,start_mark,end_mark)                                         \
//     (NODE_INIT((node),yaml_SEQUENCE_NODE,(node_tag),(start_mark),(end_mark)),   \
//      (node).data.sequence.items.start = (node_items_start),                     \
//      (node).data.sequence.items.end = (node_items_end),                         \
//      (node).data.sequence.items.top = (node_items_start),                       \
//      (node).data.sequence.style = (node_style))
//
// #define MAPPING_NODE_INIT(node,node_tag,node_pairs_start,node_pairs_end,        \
//         node_style,start_mark,end_mark)                                         \
//     (NODE_INIT((node),yaml_MAPPING_NODE,(node_tag),(start_mark),(end_mark)),    \
//      (node).data.mapping.pairs.start = (node_pairs_start),                      \
//      (node).data.mapping.pairs.end = (node_pairs_end),                          \
//      (node).data.mapping.pairs.top = (node_pairs_start),                        \
//      (node).data.mapping.style = (node_style))
//
