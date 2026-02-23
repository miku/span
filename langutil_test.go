package span

import "testing"

func TestDetectLang3(t *testing.T) {
	var cases = []struct {
		about string
		in    string
		out   string
		err   error
	}{
		{"empty", "", "", nil},
		{"english", "indivisible, with liberty and justice for all", "eng", nil},
		{"german", "in Hoffnung den Grund und die rechte Tieffe darinnen zu finden", "deu", nil},
		{"italian", "uomo di cultura e appassionato di astronomia", "ita", nil},
		{"russian", "С Востока свет, с Востока силы!", "rus", nil},
		{"esperanto", "Tiam drako estis simbolo de la supernatura", "epo", nil},
		{"french", "Le long du vieux faubourg, où pendent aux masures", "fra", nil},
		{"german mixed", "Hello World: Eine Einführung", "deu", nil},
		{"english phrase", "Reflections on Gestalt therapy", "eng", nil},
		{"miss hello world", "Hello World", "nld", nil},
		{"miss samedi soir", "Samedi soir", "nno", nil},
		{"miss samedi soir long", "Samedi soir, je viendrai dîner avec mon amie.", "nno", nil},
	}
	for _, c := range cases {
		t.Run(c.about, func(t *testing.T) {
			result, err := DetectLang3(c.in)
			if result != c.out {
				t.Errorf("got %v, want %v", result, c.out)
			}
			if err != c.err {
				t.Errorf("got error %v, want %v", err, c.err)
			}
		})
	}
}

func TestLanguageIdentifier(t *testing.T) {
	var cases = []struct {
		in  string
		out string
	}{
		{"German", "deu"},
		{"de", "deu"},
		{"ger", "deu"},
		{"serbo-croatian", "hbs"},
		{"Albanian", "sqi"},
		{"Deutsch", ""},
		{"de_DE", ""},
	}
	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			result := LanguageIdentifier(c.in)
			if result != c.out {
				t.Errorf("got %v, want %v", result, c.out)
			}
		})
	}
}
