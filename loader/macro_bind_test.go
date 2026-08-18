package loader

import "testing"

func TestBindMacroValue(t *testing.T) {
	args := map[string]string{"storage": "route_a.ks", "target": "*good_end"}
	cases := []struct {
		name     string
		value    string
		args     map[string]string
		want     string
		resolved bool
	}{
		{"percent placeholder bound", "%storage", args, "route_a.ks", true},
		{"mp reference bound", "&mp.target", args, "*good_end", true},
		{"percent placeholder unbound", "%missing", args, "%missing", false},
		{"mp reference unbound", "&mp.missing", args, "&mp.missing", false},
		{"percent default used when unbound", "%missing|fallback.ks", args, "fallback.ks", true},
		{"percent default ignored when bound", "%storage|fallback.ks", args, "route_a.ks", true},
		{"runtime expression stays dynamic", "&f.next_storage", args, "&f.next_storage", false},
		{"literal passes through", "title.ks", args, "title.ks", true},
		{"empty passes through", "", args, "", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, resolved := bindMacroValue(tc.value, tc.args)
			if got != tc.want || resolved != tc.resolved {
				t.Errorf("bindMacroValue(%q) = (%q, %v), want (%q, %v)",
					tc.value, got, resolved, tc.want, tc.resolved)
			}
		})
	}
}

func TestBindMacroText(t *testing.T) {
	args := map[string]string{"name": "アリス", "text": "森へ行く"}
	cases := []struct {
		name     string
		value    string
		want     string
		resolved bool
	}{
		{"whole value bound", "%text", "森へ行く", true},
		{"placeholder embedded in a sentence", "%nameに会う", "アリスに会う", true},
		{"unbound placeholder kept verbatim", "%unknownに会う", "%unknownに会う", false},
		{"plain text passes through", "森へ行く", "森へ行く", true},
		{"bare percent is not a placeholder", "100% 達成", "100% 達成", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, resolved := bindMacroText(tc.value, args)
			if got != tc.want || resolved != tc.resolved {
				t.Errorf("bindMacroText(%q) = (%q, %v), want (%q, %v)",
					tc.value, got, resolved, tc.want, tc.resolved)
			}
		})
	}
}

func TestBindMacroValue_NilArgsLeavesPlaceholderUnresolved(t *testing.T) {
	got, resolved := bindMacroValue("%text", nil)
	if got != "%text" || resolved {
		t.Errorf("bindMacroValue(%q, nil) = (%q, %v), want (%q, false)", "%text", got, resolved, "%text")
	}
}
