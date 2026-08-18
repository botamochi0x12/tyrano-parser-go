; sel is the choice macro every scene calls; ui_btn is a system menu control.
[macro name="sel"]
[link storage=%storage target=%target]%text[endlink]
[endmacro]

[macro name="ui_btn"]
[button graphic="btn.png" storage=&mp.storage target=&mp.target]
[endmacro]
