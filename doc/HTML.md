# HTML methods and fields

The following methods and fields are defined in the built-in `HTML` class:
- `HTML.escape(string)`, which returns a new string that is the original string with all HTML special characters from the original string escaped into HTML entities
    - The following special characters are escaped: `<`, `>`, `&`, `'`, `"`
- `HTML.nodeType(integer)`, which returns the string representation of the specified HTML node type field integer, or `"Unknown"` if the specified integer does not correspond to a valid node type field
- `HTML.parse(file/string)`, which parses the HTML content from the specified file or string and returns an HTML node object corresponding to the root node of the parsed HTML content
    - If a file is specified, that file must be open in read mode or else a runtime error is thrown
- `HTML.render(htmlNode)`, which recursively renders the specified HTML node object along with its children and siblings to a string and returns that string
- `HTML.renderToBuf(htmlNode)`, which recursively renders the specified HTML node object along with its children and siblings to a buffer and returns that buffer
- `HTML.renderToFile(htmlNode, file/string)`, which recursively renders the specified HTML node object along with its children and siblings to the specified file in plaintext, which can be specified as a string or a file object
    - If the destination file is specified as a string, it is created if it doesn't already exist and truncated if it already exists
- `HTML.token`, which is a class that defines the following HTML tag type fields, which are all integers:
    - `HTML.token.error`, `HTML.token.text`, `HTML.token.startTag`, `HTML.token.endTag`, `HTML.token.selfClosing`, `HTML.token.comment`, `HTML.token.doctype`
- `HTML.tokenize(file/string)`, which returns an HTML tokenizer object that tokenizes a string of HTML from the specified file or string, with the assumption that the string of HTML is UTF-8 encoded
    - If a file is specified, that file must be open in read mode or else a runtime error is thrown
- `HTML.tokenType(integer)`, which returns the string representation of the specified HTML tag type field integer, or `"Unknown"` if the specified integer does not correspond to a valid tag type field
- `HTML.unescape(string)`, which returns a new string that is the original string with all HTML entities from the original string escaped into HTML special characters

HTML tokenizer objects have the following methods associated with them:
- `HTML tokenizer.err()`, which throws the error obtained from scanning the latest token as a runtime error, or does nothing and returns `nil` if there is no error
- `HTML tokenizer.iterNoNewLines()`, which returns an iterator that iterates over the current HTML tokenizer object that skips over all text tokens that only have newlines in them
- `HTML tokenizer.next()`, which advances the current tokenizer to the next token and returns the token type of that token as an integer corresponding to an HTML tag type field from the `HTML.token` class
    - This method returns the value of `HTML.token.error` if there are no more tokens left in the tokenizer
- `HTML tokenizer.nextNoNewLines()`, which advances the current tokenizer to the next token that is not a text token with only newlines and returns the next token type as an integer corresponding to an HTML tag type field from the `HTML.token` class
- `HTML tokenizer.raw()`, which returns the raw text of the current token as a buffer
- `HTML tokenizer.rawStr()`, which returns the raw text of the current token as a string
- `HTML tokenizer.token()`, which returns the current token as an HTML token object
- `HTML tokenizer.tokenType()`, which returns the token type of the current token as an integer corresponding to an HTML tag type field from the `HTML.token` class
- `HTML tokenizer.tokenTypeStr()`, which returns the token type of the current token as a string
- `HTML tokenizer.toList()`, which returns a list of all tokens from the current tokenizer as HTML token objects
    - Calling this method will exhaust the underlying stream of tokens in the current tokenizer, meaning subsequent calls to this method will return an empty list instead
- `HTML tokenizer.toListNoNewLines()`, which returns a list of all tokens from the current tokenizer as HTML token objects, excluding any text tokens with only newlines
    - Calling this method will exhaust the underlying stream of tokens in the current tokenizer, meaning subsequent calls to this method will return an empty list instead

HTML token objects have the following fields associated with them:
- `HTML token.attributes`, which is a list of all HTML attributes associated with the current HTML token object as HTML attribute objects
- `HTML token.data`, which is the token content associated with the current HTML token object as a string
    - For HTML tag tokens, this field is the tag name as a string
    - For HTML text tokens, this field is the content of that text token
- `HTML token.tag`, which is the tag name associated with the current HTML token object as a string
    - For HTML text tokens, this field is an empty string
- `HTML token.type`, which is the token type of the current HTML token object as an integer corresponding to an HTML tag type field from the `HTML.token` class
- `HTML token.typeStr`, which is the token type of the current HTML token object as a string

HTML attribute objects have the following fields associated with them:
- `HTML attribute.key`, which is the key associated with the current attribute as a string
- `HTML attribute.value`, which is the value associated with the current attribute as a string

HTML node objects have the following methods and fields associated with them:
- `HTML node.ancestors()`
- `HTML node.ancestorsIter()`
- `HTML node.attributes`
- `HTML node.children()`
- `HTML node.childrenIter()`
- `HTML node.commentNodesByContent(contentStr)`
- `HTML node.commentNodesByContentIter(contentStr)`
- `HTML node.data`
- `HTML node.descendents()`
- `HTML node.descendentsIter()`
- `HTML node.dfsIter()`, which is an alias for `HTML node.descendentsIter`
- `HTML node.firstChild`
- `HTML node.fc`, which is an alias for `HTML node.firstChild`
- `HTML node.isRootNode`
- `HTML node.lastChild`
- `HTML node.lc`, which is an alias for `HTML node.lastChild`
- `HTML node.nextSibling`
- `HTML node.ns`, which is an alias for `HTML node.nextSibling`
- `HTML node.nodesByType(typeInt)`
- `HTML node.nodesByTypeIter(typeInt)`
- `HTML node.nodesByTypeStr(typeStr)`
- `HTML node.nodesByTypeStrIter(typeStr)`
- `HTML node.parent`
- `HTML node.p`, which is an alias for `HTML node.parent`
- `HTML node.prevSibling`
- `HTML node.ps`, which is an alias for `HTML node.prevSibling`
- `HTML node.render()`
- `HTML node.renderToBuf()`
- `HTML node.renderToFile(file)`
- `HTML node.tag`
- `HTML node.tagNodesByAttrKey(keyStr)`
- `HTML node.tagNodesByAttrKeyIter(keyStr)`
- `HTML node.tagNodesByAttrKeysAll(attrKeyStrs...)`
- `HTML node.tagNodesByAttrKeysAny(attrKeyStrs...)`
- `HTML node.tagNodesByAttrKeysNotAll(attrKeyStrs...)`
- `HTML node.tagNodesByAttrKeysNotAny(attrKeyStrs...)`
- `HTML node.tagNodesByAttrKeyVal(keyStr, valueStr)`
- `HTML node.tagNodesByAttrKeyValIter(keyStr, valueStr)`
- `HTML node.tagNodesByName(nameStr)`
- `HTML node.tagNodesByNameIter(nameStr)`
- `HTML node.tagNodesNoAttrs()`
- `HTML node.textNodes()`
- `HTML node.textNodesByContent(contentStr)`
- `HTML node.textNodesByContentIter(contentStr)`
- `HTML node.textNodesIter()`
- `HTML node.type`
- `HTML node.typeStr`
