package span

import "testing"

func TestDetectLang3(t *testing.T) {
	var cases = []struct {
		in  string
		out string
		err error
	}{
		// Hits.
		{"", "", nil},
		{"indivisible, with liberty and justice for all", "eng", nil},
		{"in Hoffnung den Grund und die rechte Tieffe darinnen zu finden", "deu", nil},
		{"uomo di cultura e appassionato di astronomia", "ita", nil},
		{"С Востока свет, с Востока силы!", "rus", nil},
		{"Tiam drako estis simbolo de la supernatura", "epo", nil},
		{"Le long du vieux faubourg, où pendent aux masures", "fra", nil},
		{"Hello World: Eine Einführung", "deu", nil},
		{"Reflections on Gestalt therapy", "eng", nil},
		// Misses.
		{"Hello World", "nld", nil},
		{"Samedi soir", "nno", nil},
		{"Samedi soir, je viendrai dîner avec mon amie.", "nno", nil},
	}
	for _, c := range cases {
		result, err := DetectLang3(c.in)
		if result != c.out {
			t.Fatalf("got %v, want %v", result, c.out)
		}
		if err != c.err {
			t.Fatalf("got %v, want %v", err, c.err)
		}
	}
}

func TestLanguageIdentifier(t *testing.T) {
	var cases = []struct {
		in  string
		out string
	}{
		// Hits.
		{"German", "deu"},
		{"de", "deu"},
		{"ger", "deu"},
		{"serbo-croatian", "hbs"},
		{"Albanian", "sqi"},
		// Misses.
		{"Deutsch", ""},
		{"de_DE", ""},
	}
	for _, c := range cases {
		result := LanguageIdentifier(c.in)
		if result != c.out {
			t.Fatalf("got %v, want %v", result, c.out)
		}
	}
}
