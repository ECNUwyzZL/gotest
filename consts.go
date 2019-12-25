package main

var ops = [...]string{
	";",
	",",

	// data types
	"int",
	"float",
	"string",
	"double",
	"long",
	"long long",

	// control flow
	"if",
	"else",
	"switch",
	"case",
	"do",
	"while",
	"for",

	// encapsulation
	"(",
	"{",
	"[",

	// member access
	".",
	"->",

	// arithmetic
	"+",
	"-",
	"*",
	"/",
	"%",
	"=",
	"++",
	"--",

	// logical
	"<",
	">",
	"<=",
	">=",
	"==",

	// keywords
	"break",
	"continue",
	"class",
	"struct",
	"default",
	"goto",
	"operator",
	"return",
}

type void struct{}

var pathPrefixIgnore = []string{
	".",
	"pinpoint_piggy",
}

var cCppExtList = []string {
	".c",
	".cpp",
	".h",
	".hpp",
}
