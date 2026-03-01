package gui

// LocaleEnUS is the default en-US locale (all defaults).
var LocaleEnUS = localeDefaults()

// LocaleDeDE is the German (Germany) locale.
var LocaleDeDE = Locale{
	ID: "de-DE",

	Number: NumberFormat{
		DecimalSep: ',',
		GroupSep:   '.',
		GroupSizes: []int{3},
		MinusSign:  '-',
		PlusSign:   '+',
	},

	Date: DateFormat{
		ShortDate:      "D.M.YYYY",
		LongDate:       "D. MMMM YYYY",
		MonthYear:      "MMMM YYYY",
		FirstDayOfWeek: 1,
	},

	Currency: CurrencyFormat{
		Symbol:   "\u20AC",
		Code:     "EUR",
		Position: AffixSuffix,
		Spacing:  true,
		Decimals: 2,
	},

	StrOK:     "OK",
	StrYes:    "Ja",
	StrNo:     "Nein",
	StrCancel: "Abbrechen",

	StrSave:   "Speichern",
	StrDelete: "L\u00F6schen",
	StrAdd:    "Hinzuf\u00FCgen",
	StrClear:  "L\u00F6schen",
	StrSearch: "Suche",
	StrFilter: "Filter",
	StrJump:   "Springen",
	StrReset:  "Zur\u00FCcksetzen",
	StrSubmit: "Absenden",

	StrLoading:        "Laden...",
	StrLoadingDiagram: "Diagramm laden...",
	StrSaving:         "Speichern...",
	StrSaveFailed:     "Speichern fehlgeschlagen",
	StrLoadError:      "Ladefehler",
	StrError:          "Fehler",
	StrClean:          "Sauber",

	StrOpenLink:   "Link \u00F6ffnen",
	StrGoToTarget: "Zum Ziel",
	StrCopyLink:   "Link kopieren",
	StrCopied:     "Kopiert \u2713",

	StrHorizontalScrollbar: "Horizontale Bildlaufleiste",
	StrVerticalScrollbar:   "Vertikale Bildlaufleiste",

	StrColumns:  "Spalten",
	StrSelected: "Ausgew\u00E4hlt",
	StrDraft:    "Entwurf",
	StrDirty:    "Ge\u00E4ndert",
	StrMatches:  "Treffer",
	StrPage:     "Seite",
	StrRows:     "Zeilen",

	WeekdaysShort: [7]string{"S", "M", "D", "M", "D", "F", "S"},
	WeekdaysMed:   [7]string{"So", "Mo", "Di", "Mi", "Do", "Fr", "Sa"},
	WeekdaysFull: [7]string{
		"Sonntag", "Montag", "Dienstag", "Mittwoch",
		"Donnerstag", "Freitag", "Samstag",
	},
	MonthsShort: [12]string{
		"Jan", "Feb", "M\u00E4r", "Apr", "Mai", "Jun",
		"Jul", "Aug", "Sep", "Okt", "Nov", "Dez",
	},
	MonthsFull: [12]string{
		"Januar", "Februar", "M\u00E4rz", "April",
		"Mai", "Juni", "Juli", "August",
		"September", "Oktober", "November", "Dezember",
	},
}

// LocaleArSA is the Arabic (Saudi Arabia) locale.
var LocaleArSA = Locale{
	ID:      "ar-SA",
	TextDir: TextDirRTL,

	Number: NumberFormat{
		DecimalSep: '.',
		GroupSep:   ',',
		GroupSizes: []int{3},
		MinusSign:  '-',
		PlusSign:   '+',
	},

	Date: DateFormat{
		ShortDate:      "D/M/YYYY",
		LongDate:       "D MMMM YYYY",
		MonthYear:      "MMMM YYYY",
		FirstDayOfWeek: 6, // Saturday
	},

	Currency: CurrencyFormat{
		Symbol:   "\u0631.\u0633",
		Code:     "SAR",
		Position: AffixSuffix,
		Spacing:  true,
		Decimals: 2,
	},

	StrOK:     "\u0645\u0648\u0627\u0641\u0642",
	StrYes:    "\u0646\u0639\u0645",
	StrNo:     "\u0644\u0627",
	StrCancel: "\u0625\u0644\u063A\u0627\u0621",

	StrSave:   "\u062D\u0641\u0638",
	StrDelete: "\u062D\u0630\u0641",
	StrAdd:    "\u0625\u0636\u0627\u0641\u0629",
	StrClear:  "\u0645\u0633\u062D",
	StrSearch: "\u0628\u062D\u062B",
	StrFilter: "\u062A\u0635\u0641\u064A\u0629",
	StrJump:   "\u0627\u0646\u062A\u0642\u0627\u0644",
	StrReset:  "\u0625\u0639\u0627\u062F\u0629 \u062A\u0639\u064A\u064A\u0646",
	StrSubmit: "\u0625\u0631\u0633\u0627\u0644",

	StrLoading:        "\u062C\u0627\u0631\u064D \u0627\u0644\u062A\u062D\u0645\u064A\u0644...",
	StrLoadingDiagram: "\u062C\u0627\u0631\u064D \u062A\u062D\u0645\u064A\u0644 \u0627\u0644\u0645\u062E\u0637\u0637...",
	StrSaving:         "\u062C\u0627\u0631\u064D \u0627\u0644\u062D\u0641\u0638...",
	StrSaveFailed:     "\u0641\u0634\u0644 \u0627\u0644\u062D\u0641\u0638",
	StrLoadError:      "\u062E\u0637\u0623 \u0641\u064A \u0627\u0644\u062A\u062D\u0645\u064A\u0644",
	StrError:          "\u062E\u0637\u0623",
	StrClean:          "\u0646\u0638\u064A\u0641",

	StrOpenLink:   "\u0641\u062A\u062D \u0627\u0644\u0631\u0627\u0628\u0637",
	StrGoToTarget: "\u0627\u0644\u0630\u0647\u0627\u0628 \u0625\u0644\u0649 \u0627\u0644\u0647\u062F\u0641",
	StrCopyLink:   "\u0646\u0633\u062E \u0627\u0644\u0631\u0627\u0628\u0637",
	StrCopied:     "\u062A\u0645 \u0627\u0644\u0646\u0633\u062E \u2713",

	StrHorizontalScrollbar: "\u0634\u0631\u064A\u0637 \u0627\u0644\u062A\u0645\u0631\u064A\u0631 \u0627\u0644\u0623\u0641\u0642\u064A",
	StrVerticalScrollbar:   "\u0634\u0631\u064A\u0637 \u0627\u0644\u062A\u0645\u0631\u064A\u0631 \u0627\u0644\u0639\u0645\u0648\u062F\u064A",

	StrColumns:  "\u0627\u0644\u0623\u0639\u0645\u062F\u0629",
	StrSelected: "\u0645\u062D\u062F\u062F",
	StrDraft:    "\u0645\u0633\u0648\u062F\u0629",
	StrDirty:    "\u0645\u0639\u062F\u0651\u0644",
	StrMatches:  "\u062A\u0637\u0627\u0628\u0642",
	StrPage:     "\u0635\u0641\u062D\u0629",
	StrRows:     "\u0635\u0641\u0648\u0641",

	WeekdaysShort: [7]string{
		"\u062D", "\u0646", "\u062B", "\u0631",
		"\u062E", "\u062C", "\u0633",
	},
	WeekdaysMed: [7]string{
		"\u0623\u062D\u062F", "\u0627\u062B\u0646",
		"\u062B\u0644\u0627", "\u0623\u0631\u0628",
		"\u062E\u0645\u064A", "\u062C\u0645\u0639",
		"\u0633\u0628\u062A",
	},
	WeekdaysFull: [7]string{
		"\u0627\u0644\u0623\u062D\u062F",
		"\u0627\u0644\u0627\u062B\u0646\u064A\u0646",
		"\u0627\u0644\u062B\u0644\u0627\u062B\u0627\u0621",
		"\u0627\u0644\u0623\u0631\u0628\u0639\u0627\u0621",
		"\u0627\u0644\u062E\u0645\u064A\u0633",
		"\u0627\u0644\u062C\u0645\u0639\u0629",
		"\u0627\u0644\u0633\u0628\u062A",
	},
	MonthsShort: [12]string{
		"\u064A\u0646\u0627", "\u0641\u0628\u0631",
		"\u0645\u0627\u0631", "\u0623\u0628\u0631",
		"\u0645\u0627\u064A", "\u064A\u0648\u0646",
		"\u064A\u0648\u0644", "\u0623\u063A\u0633",
		"\u0633\u0628\u062A", "\u0623\u0643\u062A",
		"\u0646\u0648\u0641", "\u062F\u064A\u0633",
	},
	MonthsFull: [12]string{
		"\u064A\u0646\u0627\u064A\u0631",
		"\u0641\u0628\u0631\u0627\u064A\u0631",
		"\u0645\u0627\u0631\u0633",
		"\u0623\u0628\u0631\u064A\u0644",
		"\u0645\u0627\u064A\u0648",
		"\u064A\u0648\u0646\u064A\u0648",
		"\u064A\u0648\u0644\u064A\u0648",
		"\u0623\u063A\u0633\u0637\u0633",
		"\u0633\u0628\u062A\u0645\u0628\u0631",
		"\u0623\u0643\u062A\u0648\u0628\u0631",
		"\u0646\u0648\u0641\u0645\u0628\u0631",
		"\u062F\u064A\u0633\u0645\u0628\u0631",
	},
}
