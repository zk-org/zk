# Styling

`zk` supports a `{{style}}` [template helper](template.md) to format its output with colors and font decorations.

Usage: `{{style "<rules>" "<text>"}}`

Multiple rules can be provided, separated by spaces.

Examples:
```
Inline: {{style "red bold" "One is never alone with a rubber duck."}}

Block:
{{#style "underline"}}
For a moment, nothing happened. Then, after a second
or so, nothing continued to happen.
{{/style}}
```

## Styling rules

* Decorations: `bold`, `italic`, `faint`, `underline`, `strikethrough`, `blink`, `reverse`, `hidden`
* Text color: `black`, `red`, `green`, `yellow`, `blue`, `magenta`, `cyan`, `white`
* Text color (bright): `bright-black`, `bright-red`, `bright-green`, `bright-yellow`, `bright-blue`, `bright-magenta`, `bright-cyan`, `bright-white`
* Background color: `black-bg`, `red-bg`, `green-bg`, `yellow-bg`, `blue-bg`, `magenta-bg`, `cyan-bg`, `white-bg`
* Background color (bright): `bright-black-bg`, `bright-red-bg`, `bright-green-bg`, `bright-yellow-bg`, `bright-blue-bg`, `bright-magenta-bg`, `bright-cyan-bg`, `bright-white-bg`

