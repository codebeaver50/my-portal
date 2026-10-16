package permission

import "testing"

func TestParseLevel(t *testing.T) {
	cases := map[string]Level{"read": Read, "write": Write, "": None, "none": None, "WRITE": None}
	for value, want := range cases {
		if got := ParseLevel(value); got != want {
			t.Errorf("ParseLevel(%q) = %v, want %v", value, got, want)
		}
	}
}

func TestLookupLevel(t *testing.T) {
	for value, want := range map[string]Level{"none": None, "read": Read, "write": Write} {
		if got, ok := LookupLevel(value); !ok || got != want {
			t.Errorf("LookupLevel(%q) = %v, %v, want %v, true", value, got, ok, want)
		}
	}
	for _, value := range []string{"", "WRITE", "admin"} {
		if _, ok := LookupLevel(value); ok {
			t.Errorf("LookupLevel(%q) should not be ok", value)
		}
	}
}

func TestLevel_Allows(t *testing.T) {
	if !Write.Allows(Read) || !Write.Allows(Write) || !Read.Allows(Read) {
		t.Error("stronger or equal level should allow")
	}
	if Read.Allows(Write) || None.Allows(Read) {
		t.Error("weaker level should not allow")
	}
}

func TestSet_Grant(t *testing.T) {
	t.Run("keeps the strongest level across grants", func(t *testing.T) {
		s := Set{}
		s.Grant(BookDatabase, Write)
		s.Grant(BookDatabase, Read)
		if s.Level(BookDatabase) != Write {
			t.Errorf("level = %v, want write", s.Level(BookDatabase))
		}
	})

	t.Run("caps the level at the resource's max level", func(t *testing.T) {
		s := Set{}
		s.Grant(ShopCSS, Write)
		if s.Level(ShopCSS) != Read {
			t.Errorf("level = %v, want read", s.Level(ShopCSS))
		}
	})

	t.Run("ignores unknown resources", func(t *testing.T) {
		s := Set{}
		s.Grant("unknown", Write)
		if len(s) != 0 {
			t.Errorf("set = %v, want empty", s)
		}
	})
}
