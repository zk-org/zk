# Strings are a complicated data structure

Given the Hindi word "नमस्ते":

1.  It can be represented as a byte array of 18 bytes:
    `[224, 164, 168, 224, 164, 174, 224, 164, 184, 224, 165, 141, 224, 164, 164, 224, 165, 135]`
    
2.  If you look at Unicode scalar values, you get an array of 6 characters:
    `['न', 'म', 'स', '्', 'त', 'े']`

3.  But the fourth and sixth letters are diacritics. To get the human-readable letters, you need to look at the strings as an array of *grapheme clusters*:
    `["न", "म", "स्", "ते"]`

:programming:
