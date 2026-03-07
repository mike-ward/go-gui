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

// LocaleFrFR is the French (France) locale.
var LocaleFrFR = Locale{
	ID: "fr-FR",

	Number: NumberFormat{
		DecimalSep: ',',
		GroupSep:   '\u00A0', // narrow no-break space
		GroupSizes: []int{3},
		MinusSign:  '-',
		PlusSign:   '+',
	},

	Date: DateFormat{
		ShortDate:      "DD/MM/YYYY",
		LongDate:       "D MMMM YYYY",
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
	StrYes:    "Oui",
	StrNo:     "Non",
	StrCancel: "Annuler",

	StrSave:   "Enregistrer",
	StrDelete: "Supprimer",
	StrAdd:    "Ajouter",
	StrClear:  "Effacer",
	StrSearch: "Rechercher",
	StrFilter: "Filtrer",
	StrJump:   "Aller \u00E0",
	StrReset:  "R\u00E9initialiser",
	StrSubmit: "Envoyer",

	StrLoading:        "Chargement...",
	StrLoadingDiagram: "Chargement du diagramme...",
	StrSaving:         "Enregistrement...",
	StrSaveFailed:     "\u00C9chec de l\u2019enregistrement",
	StrLoadError:      "Erreur de chargement",
	StrError:          "Erreur",
	StrClean:          "Propre",

	StrOpenLink:   "Ouvrir le lien",
	StrGoToTarget: "Aller \u00E0 la cible",
	StrCopyLink:   "Copier le lien",
	StrCopied:     "Copi\u00E9 \u2713",

	StrHorizontalScrollbar: "Barre de d\u00E9filement horizontale",
	StrVerticalScrollbar:   "Barre de d\u00E9filement verticale",

	StrColumns:  "Colonnes",
	StrSelected: "S\u00E9lectionn\u00E9",
	StrDraft:    "Brouillon",
	StrDirty:    "Modifi\u00E9",
	StrMatches:  "Correspondances",
	StrPage:     "Page",
	StrRows:     "Lignes",

	WeekdaysShort: [7]string{"D", "L", "M", "M", "J", "V", "S"},
	WeekdaysMed:   [7]string{"dim", "lun", "mar", "mer", "jeu", "ven", "sam"},
	WeekdaysFull: [7]string{
		"dimanche", "lundi", "mardi", "mercredi",
		"jeudi", "vendredi", "samedi",
	},
	MonthsShort: [12]string{
		"janv", "f\u00E9vr", "mars", "avr", "mai", "juin",
		"juil", "ao\u00FBt", "sept", "oct", "nov", "d\u00E9c",
	},
	MonthsFull: [12]string{
		"janvier", "f\u00E9vrier", "mars", "avril",
		"mai", "juin", "juillet", "ao\u00FBt",
		"septembre", "octobre", "novembre", "d\u00E9cembre",
	},
}

// LocaleEsES is the Spanish (Spain) locale.
var LocaleEsES = Locale{
	ID: "es-ES",

	Number: NumberFormat{
		DecimalSep: ',',
		GroupSep:   '.',
		GroupSizes: []int{3},
		MinusSign:  '-',
		PlusSign:   '+',
	},

	Date: DateFormat{
		ShortDate:      "DD/MM/YYYY",
		LongDate:       "D de MMMM de YYYY",
		MonthYear:      "MMMM de YYYY",
		FirstDayOfWeek: 1,
	},

	Currency: CurrencyFormat{
		Symbol:   "\u20AC",
		Code:     "EUR",
		Position: AffixSuffix,
		Spacing:  true,
		Decimals: 2,
	},

	StrOK:     "Aceptar",
	StrYes:    "S\u00ED",
	StrNo:     "No",
	StrCancel: "Cancelar",

	StrSave:   "Guardar",
	StrDelete: "Eliminar",
	StrAdd:    "A\u00F1adir",
	StrClear:  "Limpiar",
	StrSearch: "Buscar",
	StrFilter: "Filtrar",
	StrJump:   "Ir a",
	StrReset:  "Restablecer",
	StrSubmit: "Enviar",

	StrLoading:        "Cargando...",
	StrLoadingDiagram: "Cargando diagrama...",
	StrSaving:         "Guardando...",
	StrSaveFailed:     "Error al guardar",
	StrLoadError:      "Error de carga",
	StrError:          "Error",
	StrClean:          "Limpio",

	StrOpenLink:   "Abrir enlace",
	StrGoToTarget: "Ir al destino",
	StrCopyLink:   "Copiar enlace",
	StrCopied:     "Copiado \u2713",

	StrHorizontalScrollbar: "Barra de desplazamiento horizontal",
	StrVerticalScrollbar:   "Barra de desplazamiento vertical",

	StrColumns:  "Columnas",
	StrSelected: "Seleccionado",
	StrDraft:    "Borrador",
	StrDirty:    "Modificado",
	StrMatches:  "Coincidencias",
	StrPage:     "P\u00E1gina",
	StrRows:     "Filas",

	WeekdaysShort: [7]string{"D", "L", "M", "X", "J", "V", "S"},
	WeekdaysMed:   [7]string{"dom", "lun", "mar", "mi\u00E9", "jue", "vie", "s\u00E1b"},
	WeekdaysFull: [7]string{
		"domingo", "lunes", "martes", "mi\u00E9rcoles",
		"jueves", "viernes", "s\u00E1bado",
	},
	MonthsShort: [12]string{
		"ene", "feb", "mar", "abr", "may", "jun",
		"jul", "ago", "sep", "oct", "nov", "dic",
	},
	MonthsFull: [12]string{
		"enero", "febrero", "marzo", "abril",
		"mayo", "junio", "julio", "agosto",
		"septiembre", "octubre", "noviembre", "diciembre",
	},
}

// LocalePtBR is the Portuguese (Brazil) locale.
var LocalePtBR = Locale{
	ID: "pt-BR",

	Number: NumberFormat{
		DecimalSep: ',',
		GroupSep:   '.',
		GroupSizes: []int{3},
		MinusSign:  '-',
		PlusSign:   '+',
	},

	Date: DateFormat{
		ShortDate:      "DD/MM/YYYY",
		LongDate:       "D de MMMM de YYYY",
		MonthYear:      "MMMM de YYYY",
		FirstDayOfWeek: 0,
	},

	Currency: CurrencyFormat{
		Symbol:   "R$",
		Code:     "BRL",
		Position: AffixPrefix,
		Spacing:  true,
		Decimals: 2,
	},

	StrOK:     "OK",
	StrYes:    "Sim",
	StrNo:     "N\u00E3o",
	StrCancel: "Cancelar",

	StrSave:   "Salvar",
	StrDelete: "Excluir",
	StrAdd:    "Adicionar",
	StrClear:  "Limpar",
	StrSearch: "Pesquisar",
	StrFilter: "Filtrar",
	StrJump:   "Ir para",
	StrReset:  "Redefinir",
	StrSubmit: "Enviar",

	StrLoading:        "Carregando...",
	StrLoadingDiagram: "Carregando diagrama...",
	StrSaving:         "Salvando...",
	StrSaveFailed:     "Falha ao salvar",
	StrLoadError:      "Erro ao carregar",
	StrError:          "Erro",
	StrClean:          "Limpo",

	StrOpenLink:   "Abrir link",
	StrGoToTarget: "Ir ao destino",
	StrCopyLink:   "Copiar link",
	StrCopied:     "Copiado \u2713",

	StrHorizontalScrollbar: "Barra de rolagem horizontal",
	StrVerticalScrollbar:   "Barra de rolagem vertical",

	StrColumns:  "Colunas",
	StrSelected: "Selecionado",
	StrDraft:    "Rascunho",
	StrDirty:    "Modificado",
	StrMatches:  "Correspond\u00EAncias",
	StrPage:     "P\u00E1gina",
	StrRows:     "Linhas",

	WeekdaysShort: [7]string{"D", "S", "T", "Q", "Q", "S", "S"},
	WeekdaysMed:   [7]string{"dom", "seg", "ter", "qua", "qui", "sex", "s\u00E1b"},
	WeekdaysFull: [7]string{
		"domingo", "segunda-feira", "ter\u00E7a-feira",
		"quarta-feira", "quinta-feira", "sexta-feira",
		"s\u00E1bado",
	},
	MonthsShort: [12]string{
		"jan", "fev", "mar", "abr", "mai", "jun",
		"jul", "ago", "set", "out", "nov", "dez",
	},
	MonthsFull: [12]string{
		"janeiro", "fevereiro", "mar\u00E7o", "abril",
		"maio", "junho", "julho", "agosto",
		"setembro", "outubro", "novembro", "dezembro",
	},
}

// LocaleJaJP is the Japanese (Japan) locale.
var LocaleJaJP = Locale{
	ID: "ja-JP",

	Number: NumberFormat{
		DecimalSep: '.',
		GroupSep:   ',',
		GroupSizes: []int{3},
		MinusSign:  '-',
		PlusSign:   '+',
	},

	Date: DateFormat{
		ShortDate:      "YYYY/M/D",
		LongDate:       "YYYY\u5E74M\u6708D\u65E5",
		MonthYear:      "YYYY\u5E74M\u6708",
		FirstDayOfWeek: 0,
	},

	Currency: CurrencyFormat{
		Symbol:   "\u00A5",
		Code:     "JPY",
		Position: AffixPrefix,
		Decimals: 0,
	},

	StrOK:     "OK",
	StrYes:    "\u306F\u3044",
	StrNo:     "\u3044\u3044\u3048",
	StrCancel: "\u30AD\u30E3\u30F3\u30BB\u30EB",

	StrSave:   "\u4FDD\u5B58",
	StrDelete: "\u524A\u9664",
	StrAdd:    "\u8FFD\u52A0",
	StrClear:  "\u30AF\u30EA\u30A2",
	StrSearch: "\u691C\u7D22",
	StrFilter: "\u30D5\u30A3\u30EB\u30BF\u30FC",
	StrJump:   "\u30B8\u30E3\u30F3\u30D7",
	StrReset:  "\u30EA\u30BB\u30C3\u30C8",
	StrSubmit: "\u9001\u4FE1",

	StrLoading:        "\u8AAD\u307F\u8FBC\u307F\u4E2D...",
	StrLoadingDiagram: "\u56F3\u3092\u8AAD\u307F\u8FBC\u307F\u4E2D...",
	StrSaving:         "\u4FDD\u5B58\u4E2D...",
	StrSaveFailed:     "\u4FDD\u5B58\u5931\u6557",
	StrLoadError:      "\u8AAD\u307F\u8FBC\u307F\u30A8\u30E9\u30FC",
	StrError:          "\u30A8\u30E9\u30FC",
	StrClean:          "\u30AF\u30EA\u30FC\u30F3",

	StrOpenLink:   "\u30EA\u30F3\u30AF\u3092\u958B\u304F",
	StrGoToTarget: "\u30BF\u30FC\u30B2\u30C3\u30C8\u3078",
	StrCopyLink:   "\u30EA\u30F3\u30AF\u3092\u30B3\u30D4\u30FC",
	StrCopied:     "\u30B3\u30D4\u30FC\u6E08\u307F \u2713",

	StrHorizontalScrollbar: "\u6C34\u5E73\u30B9\u30AF\u30ED\u30FC\u30EB\u30D0\u30FC",
	StrVerticalScrollbar:   "\u5782\u76F4\u30B9\u30AF\u30ED\u30FC\u30EB\u30D0\u30FC",

	StrColumns:  "\u5217",
	StrSelected: "\u9078\u629E\u6E08\u307F",
	StrDraft:    "\u4E0B\u66F8\u304D",
	StrDirty:    "\u5909\u66F4\u6709\u308A",
	StrMatches:  "\u4E00\u81F4",
	StrPage:     "\u30DA\u30FC\u30B8",
	StrRows:     "\u884C",

	WeekdaysShort: [7]string{
		"\u65E5", "\u6708", "\u706B", "\u6C34",
		"\u6728", "\u91D1", "\u571F",
	},
	WeekdaysMed: [7]string{
		"\u65E5\u66DC", "\u6708\u66DC", "\u706B\u66DC",
		"\u6C34\u66DC", "\u6728\u66DC", "\u91D1\u66DC",
		"\u571F\u66DC",
	},
	WeekdaysFull: [7]string{
		"\u65E5\u66DC\u65E5", "\u6708\u66DC\u65E5",
		"\u706B\u66DC\u65E5", "\u6C34\u66DC\u65E5",
		"\u6728\u66DC\u65E5", "\u91D1\u66DC\u65E5",
		"\u571F\u66DC\u65E5",
	},
	MonthsShort: [12]string{
		"1\u6708", "2\u6708", "3\u6708", "4\u6708",
		"5\u6708", "6\u6708", "7\u6708", "8\u6708",
		"9\u6708", "10\u6708", "11\u6708", "12\u6708",
	},
	MonthsFull: [12]string{
		"1\u6708", "2\u6708", "3\u6708", "4\u6708",
		"5\u6708", "6\u6708", "7\u6708", "8\u6708",
		"9\u6708", "10\u6708", "11\u6708", "12\u6708",
	},
}

// LocaleZhCN is the Chinese Simplified (China) locale.
var LocaleZhCN = Locale{
	ID: "zh-CN",

	Number: NumberFormat{
		DecimalSep: '.',
		GroupSep:   ',',
		GroupSizes: []int{3},
		MinusSign:  '-',
		PlusSign:   '+',
	},

	Date: DateFormat{
		ShortDate:      "YYYY-M-D",
		LongDate:       "YYYY\u5E74M\u6708D\u65E5",
		MonthYear:      "YYYY\u5E74M\u6708",
		FirstDayOfWeek: 1,
	},

	Currency: CurrencyFormat{
		Symbol:   "\u00A5",
		Code:     "CNY",
		Position: AffixPrefix,
		Decimals: 2,
	},

	StrOK:     "\u786E\u5B9A",
	StrYes:    "\u662F",
	StrNo:     "\u5426",
	StrCancel: "\u53D6\u6D88",

	StrSave:   "\u4FDD\u5B58",
	StrDelete: "\u5220\u9664",
	StrAdd:    "\u6DFB\u52A0",
	StrClear:  "\u6E05\u9664",
	StrSearch: "\u641C\u7D22",
	StrFilter: "\u7B5B\u9009",
	StrJump:   "\u8DF3\u8F6C",
	StrReset:  "\u91CD\u7F6E",
	StrSubmit: "\u63D0\u4EA4",

	StrLoading:        "\u52A0\u8F7D\u4E2D...",
	StrLoadingDiagram: "\u52A0\u8F7D\u56FE\u8868\u4E2D...",
	StrSaving:         "\u4FDD\u5B58\u4E2D...",
	StrSaveFailed:     "\u4FDD\u5B58\u5931\u8D25",
	StrLoadError:      "\u52A0\u8F7D\u9519\u8BEF",
	StrError:          "\u9519\u8BEF",
	StrClean:          "\u65E0\u4FEE\u6539",

	StrOpenLink:   "\u6253\u5F00\u94FE\u63A5",
	StrGoToTarget: "\u8F6C\u5230\u76EE\u6807",
	StrCopyLink:   "\u590D\u5236\u94FE\u63A5",
	StrCopied:     "\u5DF2\u590D\u5236 \u2713",

	StrHorizontalScrollbar: "\u6C34\u5E73\u6EDA\u52A8\u6761",
	StrVerticalScrollbar:   "\u5782\u76F4\u6EDA\u52A8\u6761",

	StrColumns:  "\u5217",
	StrSelected: "\u5DF2\u9009\u62E9",
	StrDraft:    "\u8349\u7A3F",
	StrDirty:    "\u5DF2\u4FEE\u6539",
	StrMatches:  "\u5339\u914D",
	StrPage:     "\u9875",
	StrRows:     "\u884C",

	WeekdaysShort: [7]string{
		"\u65E5", "\u4E00", "\u4E8C", "\u4E09",
		"\u56DB", "\u4E94", "\u516D",
	},
	WeekdaysMed: [7]string{
		"\u5468\u65E5", "\u5468\u4E00", "\u5468\u4E8C",
		"\u5468\u4E09", "\u5468\u56DB", "\u5468\u4E94",
		"\u5468\u516D",
	},
	WeekdaysFull: [7]string{
		"\u661F\u671F\u65E5", "\u661F\u671F\u4E00",
		"\u661F\u671F\u4E8C", "\u661F\u671F\u4E09",
		"\u661F\u671F\u56DB", "\u661F\u671F\u4E94",
		"\u661F\u671F\u516D",
	},
	MonthsShort: [12]string{
		"1\u6708", "2\u6708", "3\u6708", "4\u6708",
		"5\u6708", "6\u6708", "7\u6708", "8\u6708",
		"9\u6708", "10\u6708", "11\u6708", "12\u6708",
	},
	MonthsFull: [12]string{
		"\u4E00\u6708", "\u4E8C\u6708", "\u4E09\u6708",
		"\u56DB\u6708", "\u4E94\u6708", "\u516D\u6708",
		"\u4E03\u6708", "\u516B\u6708", "\u4E5D\u6708",
		"\u5341\u6708", "\u5341\u4E00\u6708", "\u5341\u4E8C\u6708",
	},
}

// LocaleKoKR is the Korean (South Korea) locale.
var LocaleKoKR = Locale{
	ID: "ko-KR",

	Number: NumberFormat{
		DecimalSep: '.',
		GroupSep:   ',',
		GroupSizes: []int{3},
		MinusSign:  '-',
		PlusSign:   '+',
	},

	Date: DateFormat{
		ShortDate:      "YYYY.M.D",
		LongDate:       "YYYY\uB144 M\uC6D4 D\uC77C",
		MonthYear:      "YYYY\uB144 M\uC6D4",
		FirstDayOfWeek: 0,
	},

	Currency: CurrencyFormat{
		Symbol:   "\u20A9",
		Code:     "KRW",
		Position: AffixPrefix,
		Decimals: 0,
	},

	StrOK:     "\uD655\uC778",
	StrYes:    "\uC608",
	StrNo:     "\uC544\uB2C8\uC624",
	StrCancel: "\uCDE8\uC18C",

	StrSave:   "\uC800\uC7A5",
	StrDelete: "\uC0AD\uC81C",
	StrAdd:    "\uCD94\uAC00",
	StrClear:  "\uC9C0\uC6B0\uAE30",
	StrSearch: "\uAC80\uC0C9",
	StrFilter: "\uD544\uD130",
	StrJump:   "\uC774\uB3D9",
	StrReset:  "\uCD08\uAE30\uD654",
	StrSubmit: "\uC81C\uCD9C",

	StrLoading:        "\uB85C\uB529 \uC911...",
	StrLoadingDiagram: "\uB2E4\uC774\uC5B4\uADF8\uB7A8 \uB85C\uB529 \uC911...",
	StrSaving:         "\uC800\uC7A5 \uC911...",
	StrSaveFailed:     "\uC800\uC7A5 \uC2E4\uD328",
	StrLoadError:      "\uB85C\uB4DC \uC624\uB958",
	StrError:          "\uC624\uB958",
	StrClean:          "\uAE68\uB057\uD568",

	StrOpenLink:   "\uB9C1\uD06C \uC5F4\uAE30",
	StrGoToTarget: "\uB300\uC0C1\uC73C\uB85C \uC774\uB3D9",
	StrCopyLink:   "\uB9C1\uD06C \uBCF5\uC0AC",
	StrCopied:     "\uBCF5\uC0AC\uB428 \u2713",

	StrHorizontalScrollbar: "\uAC00\uB85C \uC2A4\uD06C\uB864\uBC14",
	StrVerticalScrollbar:   "\uC138\uB85C \uC2A4\uD06C\uB864\uBC14",

	StrColumns:  "\uC5F4",
	StrSelected: "\uC120\uD0DD\uB428",
	StrDraft:    "\uCD08\uC548",
	StrDirty:    "\uC218\uC815\uB428",
	StrMatches:  "\uC77C\uCE58",
	StrPage:     "\uD398\uC774\uC9C0",
	StrRows:     "\uD589",

	WeekdaysShort: [7]string{
		"\uC77C", "\uC6D4", "\uD654", "\uC218",
		"\uBAA9", "\uAE08", "\uD1A0",
	},
	WeekdaysMed: [7]string{
		"\uC77C\uC694", "\uC6D4\uC694", "\uD654\uC694",
		"\uC218\uC694", "\uBAA9\uC694", "\uAE08\uC694",
		"\uD1A0\uC694",
	},
	WeekdaysFull: [7]string{
		"\uC77C\uC694\uC77C", "\uC6D4\uC694\uC77C",
		"\uD654\uC694\uC77C", "\uC218\uC694\uC77C",
		"\uBAA9\uC694\uC77C", "\uAE08\uC694\uC77C",
		"\uD1A0\uC694\uC77C",
	},
	MonthsShort: [12]string{
		"1\uC6D4", "2\uC6D4", "3\uC6D4", "4\uC6D4",
		"5\uC6D4", "6\uC6D4", "7\uC6D4", "8\uC6D4",
		"9\uC6D4", "10\uC6D4", "11\uC6D4", "12\uC6D4",
	},
	MonthsFull: [12]string{
		"1\uC6D4", "2\uC6D4", "3\uC6D4", "4\uC6D4",
		"5\uC6D4", "6\uC6D4", "7\uC6D4", "8\uC6D4",
		"9\uC6D4", "10\uC6D4", "11\uC6D4", "12\uC6D4",
	},
}

// LocaleHeIL is the Hebrew (Israel) locale.
var LocaleHeIL = Locale{
	ID:      "he-IL",
	TextDir: TextDirRTL,

	Number: NumberFormat{
		DecimalSep: '.',
		GroupSep:   ',',
		GroupSizes: []int{3},
		MinusSign:  '-',
		PlusSign:   '+',
	},

	Date: DateFormat{
		ShortDate:      "D.M.YYYY",
		LongDate:       "D \u05D1MMMM YYYY",
		MonthYear:      "MMMM YYYY",
		FirstDayOfWeek: 0,
	},

	Currency: CurrencyFormat{
		Symbol:   "\u20AA",
		Code:     "ILS",
		Position: AffixPrefix,
		Spacing:  true,
		Decimals: 2,
	},

	StrOK:     "\u05D0\u05D9\u05E9\u05D5\u05E8",
	StrYes:    "\u05DB\u05DF",
	StrNo:     "\u05DC\u05D0",
	StrCancel: "\u05D1\u05D9\u05D8\u05D5\u05DC",

	StrSave:   "\u05E9\u05DE\u05D9\u05E8\u05D4",
	StrDelete: "\u05DE\u05D7\u05D9\u05E7\u05D4",
	StrAdd:    "\u05D4\u05D5\u05E1\u05E4\u05D4",
	StrClear:  "\u05E0\u05D9\u05E7\u05D5\u05D9",
	StrSearch: "\u05D7\u05D9\u05E4\u05D5\u05E9",
	StrFilter: "\u05E1\u05D9\u05E0\u05D5\u05DF",
	StrJump:   "\u05E7\u05E4\u05D9\u05E6\u05D4",
	StrReset:  "\u05D0\u05D9\u05E4\u05D5\u05E1",
	StrSubmit: "\u05E9\u05DC\u05D9\u05D7\u05D4",

	StrLoading:        "\u05D8\u05D5\u05E2\u05DF...",
	StrLoadingDiagram: "\u05D8\u05D5\u05E2\u05DF \u05EA\u05E8\u05E9\u05D9\u05DD...",
	StrSaving:         "\u05E9\u05D5\u05DE\u05E8...",
	StrSaveFailed:     "\u05E9\u05DE\u05D9\u05E8\u05D4 \u05E0\u05DB\u05E9\u05DC\u05D4",
	StrLoadError:      "\u05E9\u05D2\u05D9\u05D0\u05EA \u05D8\u05E2\u05D9\u05E0\u05D4",
	StrError:          "\u05E9\u05D2\u05D9\u05D0\u05D4",
	StrClean:          "\u05E0\u05E7\u05D9",

	StrOpenLink:   "\u05E4\u05EA\u05D7 \u05E7\u05D9\u05E9\u05D5\u05E8",
	StrGoToTarget: "\u05E2\u05D1\u05D5\u05E8 \u05DC\u05D9\u05E2\u05D3",
	StrCopyLink:   "\u05D4\u05E2\u05EA\u05E7 \u05E7\u05D9\u05E9\u05D5\u05E8",
	StrCopied:     "\u05D4\u05D5\u05E2\u05EA\u05E7 \u2713",

	StrHorizontalScrollbar: "\u05E4\u05E1 \u05D2\u05DC\u05D9\u05DC\u05D4 \u05D0\u05D5\u05E4\u05E7\u05D9",
	StrVerticalScrollbar:   "\u05E4\u05E1 \u05D2\u05DC\u05D9\u05DC\u05D4 \u05D0\u05E0\u05DB\u05D9",

	StrColumns:  "\u05E2\u05DE\u05D5\u05D3\u05D5\u05EA",
	StrSelected: "\u05E0\u05D1\u05D7\u05E8",
	StrDraft:    "\u05D8\u05D9\u05D5\u05D8\u05D4",
	StrDirty:    "\u05E9\u05D5\u05E0\u05D4",
	StrMatches:  "\u05D4\u05EA\u05D0\u05DE\u05D5\u05EA",
	StrPage:     "\u05E2\u05DE\u05D5\u05D3",
	StrRows:     "\u05E9\u05D5\u05E8\u05D5\u05EA",

	WeekdaysShort: [7]string{
		"\u05D0", "\u05D1", "\u05D2", "\u05D3",
		"\u05D4", "\u05D5", "\u05E9",
	},
	WeekdaysMed: [7]string{
		"\u05D9\u05D5\u05DD \u05D0",
		"\u05D9\u05D5\u05DD \u05D1",
		"\u05D9\u05D5\u05DD \u05D2",
		"\u05D9\u05D5\u05DD \u05D3",
		"\u05D9\u05D5\u05DD \u05D4",
		"\u05D9\u05D5\u05DD \u05D5",
		"\u05E9\u05D1\u05EA",
	},
	WeekdaysFull: [7]string{
		"\u05D9\u05D5\u05DD \u05E8\u05D0\u05E9\u05D5\u05DF",
		"\u05D9\u05D5\u05DD \u05E9\u05E0\u05D9",
		"\u05D9\u05D5\u05DD \u05E9\u05DC\u05D9\u05E9\u05D9",
		"\u05D9\u05D5\u05DD \u05E8\u05D1\u05D9\u05E2\u05D9",
		"\u05D9\u05D5\u05DD \u05D7\u05DE\u05D9\u05E9\u05D9",
		"\u05D9\u05D5\u05DD \u05E9\u05D9\u05E9\u05D9",
		"\u05E9\u05D1\u05EA",
	},
	MonthsShort: [12]string{
		"\u05D9\u05E0\u05D5",
		"\u05E4\u05D1\u05E8",
		"\u05DE\u05E8\u05E5",
		"\u05D0\u05E4\u05E8",
		"\u05DE\u05D0\u05D9",
		"\u05D9\u05D5\u05E0",
		"\u05D9\u05D5\u05DC",
		"\u05D0\u05D5\u05D2",
		"\u05E1\u05E4\u05D8",
		"\u05D0\u05D5\u05E7",
		"\u05E0\u05D5\u05D1",
		"\u05D3\u05E6\u05DE",
	},
	MonthsFull: [12]string{
		"\u05D9\u05E0\u05D5\u05D0\u05E8",
		"\u05E4\u05D1\u05E8\u05D5\u05D0\u05E8",
		"\u05DE\u05E8\u05E5",
		"\u05D0\u05E4\u05E8\u05D9\u05DC",
		"\u05DE\u05D0\u05D9",
		"\u05D9\u05D5\u05E0\u05D9",
		"\u05D9\u05D5\u05DC\u05D9",
		"\u05D0\u05D5\u05D2\u05D5\u05E1\u05D8",
		"\u05E1\u05E4\u05D8\u05DE\u05D1\u05E8",
		"\u05D0\u05D5\u05E7\u05D8\u05D5\u05D1\u05E8",
		"\u05E0\u05D5\u05D1\u05DE\u05D1\u05E8",
		"\u05D3\u05E6\u05DE\u05D1\u05E8",
	},
}
