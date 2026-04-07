package geo

// Level represents an education level with its i18n labels.
type Level struct {
	Key   string // DB value, e.g. "3as"
	En    string
	Fr    string
	Group string // grouping label: "primary", "middle", "secondary", "higher"
}

// ─── Algeria (DZ) ──────────────────────────────────────

var LevelsDZ = []Level{
	// Primary
	{Key: "1ap", En: "1st Year Primary (1AP)", Fr: "1ère année primaire (1AP)", Group: "primary"},
	{Key: "2ap", En: "2nd Year Primary (2AP)", Fr: "2ème année primaire (2AP)", Group: "primary"},
	{Key: "3ap", En: "3rd Year Primary (3AP)", Fr: "3ème année primaire (3AP)", Group: "primary"},
	{Key: "4ap", En: "4th Year Primary (4AP)", Fr: "4ème année primaire (4AP)", Group: "primary"},
	{Key: "5ap", En: "5th Year Primary (5AP)", Fr: "5ème année primaire (5AP)", Group: "primary"},
	// Middle
	{Key: "1am", En: "1st Year Middle (1AM)", Fr: "1ère année moyenne (1AM)", Group: "middle"},
	{Key: "2am", En: "2nd Year Middle (2AM)", Fr: "2ème année moyenne (2AM)", Group: "middle"},
	{Key: "3am", En: "3rd Year Middle (3AM)", Fr: "3ème année moyenne (3AM)", Group: "middle"},
	{Key: "4am", En: "4th Year Middle — BEM (4AM)", Fr: "4ème année moyenne — BEM (4AM)", Group: "middle"},
	// Secondary
	{Key: "1as", En: "1st Year Secondary (1AS)", Fr: "1ère année secondaire (1AS)", Group: "secondary"},
	{Key: "2as", En: "2nd Year Secondary (2AS)", Fr: "2ème année secondaire (2AS)", Group: "secondary"},
	{Key: "3as", En: "3rd Year Secondary — BAC (3AS)", Fr: "3ème année secondaire — BAC (3AS)", Group: "secondary"},
	// Higher
	{Key: "licence", En: "Licence (Bachelor)", Fr: "Licence", Group: "higher"},
	{Key: "master", En: "Master", Fr: "Master", Group: "higher"},
	{Key: "doctorat", En: "Doctorate", Fr: "Doctorat", Group: "higher"},
}

// ─── France (FR) ────────────────────────────────────────

var LevelsFR = []Level{
	// Primary
	{Key: "cp", En: "CP", Fr: "CP", Group: "primary"},
	{Key: "ce1", En: "CE1", Fr: "CE1", Group: "primary"},
	{Key: "ce2", En: "CE2", Fr: "CE2", Group: "primary"},
	{Key: "cm1", En: "CM1", Fr: "CM1", Group: "primary"},
	{Key: "cm2", En: "CM2", Fr: "CM2", Group: "primary"},
	// Collège
	{Key: "6eme", En: "6th Grade (6ème)", Fr: "6ème", Group: "middle"},
	{Key: "5eme", En: "5th Grade (5ème)", Fr: "5ème", Group: "middle"},
	{Key: "4eme", En: "4th Grade (4ème)", Fr: "4ème", Group: "middle"},
	{Key: "3eme", En: "3rd Grade — Brevet (3ème)", Fr: "3ème — Brevet", Group: "middle"},
	// Lycée
	{Key: "seconde", En: "Seconde (10th)", Fr: "Seconde", Group: "secondary"},
	{Key: "premiere", En: "Première (11th)", Fr: "Première", Group: "secondary"},
	{Key: "terminale", En: "Terminale — BAC (12th)", Fr: "Terminale — BAC", Group: "secondary"},
	// Higher
	{Key: "licence_fr", En: "Licence (Bachelor)", Fr: "Licence", Group: "higher"},
	{Key: "master_fr", En: "Master", Fr: "Master", Group: "higher"},
	{Key: "prepa", En: "Prépa (CPGE)", Fr: "Prépa (CPGE)", Group: "higher"},
	{Key: "bts", En: "BTS", Fr: "BTS", Group: "higher"},
	{Key: "dut", En: "DUT / BUT", Fr: "DUT / BUT", Group: "higher"},
}

// LevelsForCountry returns the level list for a given ISO country code.
// Defaults to DZ if unknown.
func LevelsForCountry(country string) []Level {
	switch country {
	case "FR":
		return LevelsFR
	default:
		return LevelsDZ
	}
}

// ─── Wilayas (Algeria) ─────────────────────────────────

var WilayasDZ = []string{
	"Adrar", "Chlef", "Laghouat", "Oum El Bouaghi", "Batna",
	"Béjaïa", "Biskra", "Béchar", "Blida", "Bouira",
	"Tamanrasset", "Tébessa", "Tlemcen", "Tiaret", "Tizi Ouzou",
	"Alger", "Djelfa", "Jijel", "Sétif", "Saïda",
	"Skikda", "Sidi Bel Abbès", "Annaba", "Guelma", "Constantine",
	"Médéa", "Mostaganem", "M'Sila", "Mascara", "Ouargla",
	"Oran", "El Bayadh", "Illizi", "Bordj Bou Arréridj", "Boumerdès",
	"El Tarf", "Tindouf", "Tissemsilt", "El Oued", "Khenchela",
	"Souk Ahras", "Tipaza", "Mila", "Aïn Defla", "Naâma",
	"Aïn Témouchent", "Ghardaïa", "Relizane",
	"El M'Ghair", "El Meniaa", "Ouled Djellal", "Bordj Badji Mokhtar",
	"Béni Abbès", "Timimoun", "Touggourt", "Djanet", "In Salah", "In Guezzam",
}

// ─── Départements (France) ─────────────────────────────

var DepartementsFR = []string{
	"Ain", "Aisne", "Allier", "Alpes-de-Haute-Provence", "Hautes-Alpes",
	"Alpes-Maritimes", "Ardèche", "Ardennes", "Ariège", "Aube",
	"Aude", "Aveyron", "Bouches-du-Rhône", "Calvados", "Cantal",
	"Charente", "Charente-Maritime", "Cher", "Corrèze", "Corse-du-Sud",
	"Haute-Corse", "Côte-d'Or", "Côtes-d'Armor", "Creuse", "Dordogne",
	"Doubs", "Drôme", "Eure", "Eure-et-Loir", "Finistère",
	"Gard", "Haute-Garonne", "Gers", "Gironde", "Hérault",
	"Ille-et-Vilaine", "Indre", "Indre-et-Loire", "Isère", "Jura",
	"Landes", "Loir-et-Cher", "Loire", "Haute-Loire", "Loire-Atlantique",
	"Loiret", "Lot", "Lot-et-Garonne", "Lozère", "Maine-et-Loire",
	"Manche", "Marne", "Haute-Marne", "Mayenne", "Meurthe-et-Moselle",
	"Meuse", "Morbihan", "Moselle", "Nièvre", "Nord",
	"Oise", "Orne", "Pas-de-Calais", "Puy-de-Dôme", "Pyrénées-Atlantiques",
	"Hautes-Pyrénées", "Pyrénées-Orientales", "Bas-Rhin", "Haut-Rhin", "Rhône",
	"Haute-Saône", "Saône-et-Loire", "Sarthe", "Savoie", "Haute-Savoie",
	"Paris", "Seine-Maritime", "Seine-et-Marne", "Yvelines", "Deux-Sèvres",
	"Somme", "Tarn", "Tarn-et-Garonne", "Var", "Vaucluse",
	"Vendée", "Vienne", "Haute-Vienne", "Vosges", "Yonne",
	"Territoire de Belfort", "Essonne", "Hauts-de-Seine", "Seine-Saint-Denis", "Val-de-Marne",
	"Val-d'Oise",
	// DOM-TOM
	"Guadeloupe", "Martinique", "Guyane", "La Réunion", "Mayotte",
}

// RegionsForCountry returns the region list for a given ISO country code.
func RegionsForCountry(country string) []string {
	switch country {
	case "FR":
		return DepartementsFR
	default:
		return WilayasDZ
	}
}

// RegionLabel returns the i18n label for the region field based on country.
func RegionLabel(country, lang string) string {
	switch country {
	case "FR":
		if lang == "fr" {
			return "Département"
		}
		return "Department"
	default:
		return "Wilaya"
	}
}
