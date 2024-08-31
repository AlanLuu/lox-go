package ast

func defineStringFields(stringClass *LoxClass) {
	digits := "0123456789"
	stringClass.classProperties["digits"] = NewLoxString(digits, '\'')

	hexDigits := "0123456789abcdefABCDEF"
	stringClass.classProperties["hexDigits"] = NewLoxString(hexDigits, '\'')

	hexDigitsLower := "0123456789abcdef"
	stringClass.classProperties["hexDigitsLower"] = NewLoxString(hexDigitsLower, '\'')

	hexDigitsUpper := "0123456789ABCDEF"
	stringClass.classProperties["hexDigitsUpper"] = NewLoxString(hexDigitsUpper, '\'')

	letters := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	stringClass.classProperties["letters"] = NewLoxString(letters, '\'')

	lowercase := "abcdefghijklmnopqrstuvwxyz"
	stringClass.classProperties["lowercase"] = NewLoxString(lowercase, '\'')

	octDigits := "01234567"
	stringClass.classProperties["octDigits"] = NewLoxString(octDigits, '\'')

	punctuation := "!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~"
	stringClass.classProperties["punctuation"] = NewLoxString(punctuation, '\'')

	qwertyLower := "qwertyuiopasdfghjklzxcvbnm"
	stringClass.classProperties["qwertyLower"] = NewLoxString(qwertyLower, '\'')

	qwertyUpper := "QWERTYUIOPASDFGHJKLZXCVBNM"
	stringClass.classProperties["qwertyUpper"] = NewLoxString(qwertyUpper, '\'')

	uppercase := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	stringClass.classProperties["uppercase"] = NewLoxString(uppercase, '\'')
}

func (i *Interpreter) defineStringFuncs() {
	className := "String"
	stringClass := NewLoxClass(className, nil, false)

	defineStringFields(stringClass)

	i.globals.Define(className, stringClass)
}
