package geo

// Subject represents a teaching subject with its i18n key.
type Subject struct {
	Key   string // i18n key and DB value, e.g. "math"
	En    string
	Fr    string
	Emoji string
}

// AllSubjects is the canonical list of subjects available on the platform.
var AllSubjects = []Subject{
	{Key: "math", En: "Mathematics", Fr: "Mathématiques", Emoji: "📐"},
	{Key: "physics", En: "Physics", Fr: "Physique", Emoji: "⚛️"},
	{Key: "chemistry", En: "Chemistry", Fr: "Chimie", Emoji: "🧪"},
	{Key: "biology", En: "Biology", Fr: "Biologie", Emoji: "🧬"},
	{Key: "french", En: "French", Fr: "Français", Emoji: "🇫🇷"},
	{Key: "english", En: "English", Fr: "Anglais", Emoji: "🇬🇧"},
	{Key: "arabic", En: "Arabic", Fr: "Arabe", Emoji: "🇸🇦"},
	{Key: "history_geo", En: "History-Geography", Fr: "Histoire-Géographie", Emoji: "🗺️"},
	{Key: "philosophy", En: "Philosophy", Fr: "Philosophie", Emoji: "📖"},
	{Key: "islamic_sciences", En: "Islamic Sciences", Fr: "Sciences islamiques", Emoji: "☪️"},
	{Key: "computer_science", En: "Computer Science", Fr: "Informatique", Emoji: "💻"},
	{Key: "economics", En: "Economics", Fr: "Économie", Emoji: "📊"},
	{Key: "spanish", En: "Spanish", Fr: "Espagnol", Emoji: "🇪🇸"},
	{Key: "german", En: "German", Fr: "Allemand", Emoji: "🇩🇪"},
	{Key: "italian", En: "Italian", Fr: "Italien", Emoji: "🇮🇹"},
	{Key: "civil_engineering", En: "Civil Engineering", Fr: "Génie civil", Emoji: "🏗️"},
	{Key: "electrical_engineering", En: "Electrical Engineering", Fr: "Génie électrique", Emoji: "⚡"},
	{Key: "mechanical_engineering", En: "Mechanical Engineering", Fr: "Génie mécanique", Emoji: "⚙️"},
	{Key: "science", En: "Science", Fr: "Sciences", Emoji: "🔬"},
	{Key: "tamazight", En: "Tamazight", Fr: "Tamazight", Emoji: "ⵣ"},
	{Key: "other", En: "Other", Fr: "Autre", Emoji: "📚"},
}

// SubjectMap returns a map[key]Subject for quick lookup.
func SubjectMap() map[string]Subject {
	m := make(map[string]Subject, len(AllSubjects))
	for _, s := range AllSubjects {
		m[s.Key] = s
	}
	return m
}
