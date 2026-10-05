package cmd

// FormatFlags holds the output formatting options shared by commands that print lists of items.
type FormatFlags struct {
	Format     string `group:format short:f placeholder:TEMPLATE help:"Pretty print the list using a custom template or one of the predefined formats."`
	Header     string `group:format help:"Arbitrary text printed at the start of the list."`
	Footer     string "group:format default:\n help:\"Arbitrary text printed at the end of the list.\""
	Delimiter  string "group:format short:d default:\n help:\"Print items delimited by the given separator.\""
	Delimiter0 bool   "group:format short:0 name:delimiter0 help:\"Print items delimited by ASCII NUL characters. This is useful when used in conjunction with `xargs -0`.\""
	NoPager    bool   `group:format short:P help:"Do not pipe output into a pager."`
	Quiet      bool   `group:format short:q help:"Do not print the total number of items found."`
}
