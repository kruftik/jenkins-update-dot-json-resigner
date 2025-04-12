# json canonical form

All JSON is stored in UTF-8.

## Value

A value is the top-level element type.

```
element = string | number | object | array | "true" | "false" | "null"
```

## Strings

Strings contain unicode characters as they are written, with an exception for
`\` and `"` and the control characters U+0000 to U+001F.

The control characters are written as follows:

- As `\b` for U+0008, `\t` for U+0009, `\n` for U+000A, `\f` for U+000C, and
  `\r` for U+000D
- If not listed above, the character is written `\u00xx`, where `xx` is the hex
  value of the control character in lowercase hex values

```
string-element = unicode value except \ and " above U+001F
               | below-u0020
               | "\\", "\\"
               | "\\", "\""

below-u0020 = "\\", ( "b" | "t" | "n" | "f" | "r" )
            | "\\u00", ("0" | "1"), hex
hex = "0" | "1" | "2" | "3" | "4" | "5" | "6" | "7" | "8" | "9"
    | "a" | "b" | "c" | "d" | "e" | "f"

string = "\"" { string-element } "\""
```

## Numbers

Numbers correspond to the JSON definition of numbers, but are implicitly defined
into two separate categories, described below.

```
number = integer | floating-point
```

### Integers

Integers are numbers inside the range `[-(2**53)+1, (2**53)-1]` that have a zero
fractional part when written without an exponent.

An integer is emitted as a number with no fractional part and no exponent,
without leading zeroes (with the exception of zero):

```
integer = 0
        | [ "-" ], nonzero-digit, { digit }
```

### Floating point numbers

A floating point value is any number outside of the range `[-(2**53)+1, (2**53)-1]`,
or any value with a nonzero fractional part when written without an exponent.

A floating point number is emitted as a single nonzero digit, followed by an
optional fractional part, followed by an exponent. The fractional part must end
with a nonzero digit:

```
trailing-nonzero = nonzero-digit
                 | digit, trailing-nonzero

floating-point = [ "-" ] nonzero-digit, [ ".", trailing-nonzero ], "E", integer
```

Note that, although the EBNF definition of `floating-point` allows numbers such
as `1E0`, they are not valid due to the constraint described at the top of this
section.

## Objects

Objects are printed as JSON objects, where the key-value pairs are ordered
lexicographically on keys.

```
pair = string, ":", element
pairs = pair
      | pair, ",", pairs
object = "{", [ pairs ], "}"
```

## Arrays

Arrays consist of zero or more elements, separated by a comma.

```
elements = element
         | element, ",", elements
array = "[", [ elements ], "]"
```

## Value Stream

A value stream is a sequence of JSON values not grouped by an array or object,
but are still stored in the same sequence of bytes. In some cases, those
elements must be separated by a significant whitespace.

JSON values can be split into two categories: delimiter based and non-delimiter
based. The delimiter based ones are the ones that start and end with delimiters:
Strings, objects and arrays.

If a non-delimiter based value is immediately followed by another non-delimiter
based value, then a single space shall separate them:

```
needs-space = number | "true" | "false" | "null"
delimiter-based = string | object | array

element-stream = ε
               | space-element-stream
               | delimiter-based, element-stream

space-element-stream = needs-space
                     | needs-space, " ", space-element-stream
                     | needs-space, delimiter-based, element-stream
```

(The symbol ε stands for the empty sequence)
