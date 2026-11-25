package twitch

import "strings"

var languages = map[string]bool{
	"AR": true, "ASL": true, "BG": true, "CA": true, "CS": true, "DA": true, "DE": true, "EL": true, "EN": true,
	"ES": true, "FI": true, "FR": true, "HI": true, "HU": true, "ID": true, "IT": true, "JA": true, "KO": true,
	"MS": true, "NL": true, "NO": true, "PL": true, "PT": true, "RO": true, "RU": true, "SK": true, "SV": true,
	"TH": true, "TL": true, "TR": true, "UK": true, "VI": true, "ZH": true, "ZH_HK": true, "OTHER": true,
}

// ParseLanguages refuses the whole list for one unknown value, because Twitch rejects the whole query then.
func ParseLanguages(raw string) ([]string, bool) {
	var out []string
	seen := map[string]bool{}
	for _, part := range strings.Split(raw, ",") {
		code := strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(part), "-", "_"))
		if code == "" {
			continue
		}
		if !languages[code] {
			return nil, false
		}
		if !seen[code] {
			seen[code] = true
			out = append(out, code)
		}
	}
	return out, true
}

func languageVars(vars map[string]any, languages []string) map[string]any {
	if len(languages) > 0 {
		vars["languages"] = languages
	}
	return vars
}
