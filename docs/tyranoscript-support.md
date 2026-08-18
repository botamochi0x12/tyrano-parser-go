# TyranoScript Support

[← README](../README.md)

What the parser understands, and where it deliberately stops.

## Project layout

The loader expects a standard TyranoScript layout:

```text
project/
└── data/
    ├── scenario/
    │   ├── first.ks
    │   └── title.ks
    └── system/
        └── Config.tjs
```

`loader.DiscoverRoot` walks upward from a starting directory until it finds
`data/scenario/` or `data/system/Config.tjs`. `loader.LayoutFrom` uses an
explicit project root and requires `data/scenario/`.

## Scenario files

`.ks` files are parsed into KAG-compatible JSON-shaped structs. References are
extracted from `@call`, `@jump`, and `[link]` tags, and the project scan reports:

- referenced `.ks` files that do not exist
- labels that do not exist in the target file, for static references

Dynamic runtime references such as `storage=&tf.storage` are skipped rather than
reported, because their target is only known at runtime.

## Config.tjs

Real TyranoScript configs use a leading semicolon as the assignment marker:

```tjs
// comment
;System.title = "My Game";
;scWidth = 1280
;userFace = Quicksand, "Yu Gothic", sans-serif; // inline comment
;url = "http://example.com/path";
```

Supported behavior:

- leading `;` assignment prefix
- optional trailing `;`
- dotted keys
- quoted and unquoted values
- quote-aware `//` inline comment stripping, which leaves `http://` inside a
  quoted value intact
- raw string preservation for expressions such as `960-32`

## Out of scope

- full TJS evaluation
- legacy non-UTF-8 encodings
