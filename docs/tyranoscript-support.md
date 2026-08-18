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
extracted from `@call`, `@jump`, `[link]`, `[glink]`, `[button]`, `[s_button]`
and `[showbutton]`, and the project scan reports:

- referenced `.ks` files that do not exist
- labels that do not exist in the target file, for static references

Dynamic runtime references such as `storage=&tf.storage` are skipped rather than
reported, because their target is only known at runtime. They still appear in
the scan output, flagged through the `dynamic` and `resolved` fields described
in the [CLI reference](cli.md#reference-records).

A branch also carries the wording a player reads: the `text` parameter of a
button-style tag, or the body of a `[link] … [endlink]` pair.

## Macros

`[macro name="..."] … [endmacro]` definitions are collected across the whole
project, and a call to one is expanded before references are collected. A choice
authored as:

```text
[macro name="sel"]
[link storage=%storage target=%target]%text[endlink]
[endmacro]

[sel storage="route_a.ks" target="*good_end" text="森へ行く"]
```

reports `storage=route_a.ks`, `target=*good_end` and `text=森へ行く` against the
call site, rather than the unbound `%storage` placeholder or the equivalent
`&mp.storage` reference the macro body may use instead.

Supported binding forms inside a macro body:

- `%name` — the argument the call site passes
- `%name|default` — that argument, or the default when the call site omits it
- `&mp.name` — the same argument written as a runtime reference
- `%name` inside body text, so `%nameに会う` becomes `アリスに会う`

An argument the call site never binds keeps its placeholder and is listed in the
reference's `dynamic` field. Nested macros expand up to 8 levels deep, and a
macro that calls itself stops rather than looping.

References written inside a macro body are not reported at the definition site:
they are templates, and only become branches once called.

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
