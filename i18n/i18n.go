package i18n

// Translations maps: key → translated string
// "en" is the default / fallback language.
var Translations = map[string]map[string]string{
	// ─────────────────────────────────────────────────
	// LAYOUTS
	// ─────────────────────────────────────────────────
	"nav_brand":          {"en": "TeachHub", "fr": "TeachHub"},
	"nav_admin":          {"en": "admin", "fr": "admin"},
	"nav_logout":         {"en": "Logout", "fr": "Déconnexion"},
	"nav_brand_student":  {"en": "📚 TeachHub", "fr": "📚 TeachHub"},
	"page_title":         {"en": "TeachHub Admin", "fr": "TeachHub Admin"},
	"page_title_student": {"en": "TeachHub", "fr": "TeachHub"},
	"page_title_live":    {"en": "TeachHub Live", "fr": "TeachHub Live"},

	// ─────────────────────────────────────────────────
	// ADMIN LOGIN
	// ─────────────────────────────────────────────────
	"login_heading":    {"en": "Teacher Portal", "fr": "Espace Enseignant"},
	"login_subheading": {"en": "Sign in to manage your classrooms", "fr": "Connectez-vous pour gérer vos classes"},
	"login_error":      {"en": "Invalid credentials. Please try again.", "fr": "Identifiants invalides. Veuillez réessayer."},
	"login_username":   {"en": "Username", "fr": "Nom d'utilisateur"},
	"login_password":   {"en": "Password", "fr": "Mot de passe"},
	"login_submit":     {"en": "Sign In", "fr": "Se connecter"},

	// ─────────────────────────────────────────────────
	// ADMIN DASHBOARD
	// ─────────────────────────────────────────────────
	"dash_heading":     {"en": "Your Classrooms", "fr": "Vos classes"},
	"dash_subheading":  {"en": "Manage your teaching spaces", "fr": "Gérez vos espaces d'enseignement"},
	"dash_placeholder": {"en": "New classroom name...", "fr": "Nom de la nouvelle classe..."},
	"dash_create":      {"en": "+ Create", "fr": "+ Créer"},
	"dash_students":    {"en": "students", "fr": "étudiants"},
	"dash_pending":     {"en": "pending", "fr": "en attente"},
	"dash_resources":   {"en": "resources", "fr": "ressources"},
	"dash_quizzes_n":   {"en": "quizzes", "fr": "quiz"},
	"dash_empty":       {"en": "No classrooms yet", "fr": "Aucune classe pour le moment"},
	"dash_empty_sub":   {"en": "Create one above to get started", "fr": "Créez-en une ci-dessus pour commencer"},

	// ─────────────────────────────────────────────────
	// ADMIN CLASSROOM
	// ─────────────────────────────────────────────────
	"back":              {"en": "← Back", "fr": "← Retour"},
	"back_to_classroom": {"en": "← Back to Classroom", "fr": "← Retour à la classe"},
	"delete":            {"en": "Delete", "fr": "Supprimer"},
	"confirm_delete":    {"en": "Delete this classroom?", "fr": "Supprimer cette classe ?"},
	"join_link":         {"en": "Join link:", "fr": "Lien d'accès :"},
	"regenerate_code":   {"en": "↻ Regenerate code", "fr": "↻ Régénérer le code"},

	// Live class banner
	"live_active":      {"en": "Live class is active", "fr": "Le cours en direct est actif"},
	"live_started":     {"en": "Started", "fr": "Commencé"},
	"live_enter":       {"en": "📹 Enter", "fr": "📹 Entrer"},
	"live_end":         {"en": "End", "fr": "Terminer"},
	"live_end_confirm": {"en": "End the live class?", "fr": "Terminer le cours en direct ?"},
	"live_start":       {"en": "Start Live Class", "fr": "Démarrer un cours en direct"},

	// Tabs
	"tab_resources":   {"en": "Resources", "fr": "Ressources"},
	"tab_assignments": {"en": "Assignments", "fr": "Devoirs"},
	"tab_quizzes":     {"en": "Quizzes", "fr": "Quiz"},
	"tab_students":    {"en": "Students", "fr": "Étudiants"},
	"tab_analytics":   {"en": "📊 Analytics", "fr": "📊 Statistiques"},

	// Resources section
	"res_categories":      {"en": "Categories", "fr": "Catégories"},
	"res_new_cat":         {"en": "New category...", "fr": "Nouvelle catégorie..."},
	"add":                 {"en": "Add", "fr": "Ajouter"},
	"res_upload":          {"en": "📁 Upload Resource", "fr": "📁 Téléverser une ressource"},
	"title":               {"en": "Title", "fr": "Titre"},
	"desc_optional":       {"en": "Description (optional)", "fr": "Description (optionnel)"},
	"no_category":         {"en": "No category", "fr": "Sans catégorie"},
	"res_external_url":    {"en": "Or paste external URL (YouTube, etc.)", "fr": "Ou collez un lien externe (YouTube, etc.)"},
	"upload":              {"en": "Upload", "fr": "Téléverser"},
	"download":            {"en": "Download", "fr": "Télécharger"},
	"confirm_delete_item": {"en": "Delete?", "fr": "Supprimer ?"},
	"res_empty":           {"en": "No resources yet", "fr": "Aucune ressource pour le moment"},

	// Assignments section
	"assign_create":        {"en": "✏️ Create Assignment", "fr": "✏️ Créer un devoir"},
	"assign_title_ph":      {"en": "Assignment title", "fr": "Titre du devoir"},
	"assign_instructions":  {"en": "Instructions for students...", "fr": "Instructions pour les étudiants..."},
	"assign_response_type": {"en": "Response Type", "fr": "Type de réponse"},
	"assign_file":          {"en": "📎 File upload", "fr": "📎 Fichier"},
	"assign_text":          {"en": "📝 Text answer", "fr": "📝 Texte"},
	"assign_both":          {"en": "📎📝 File + Text", "fr": "📎📝 Fichier + Texte"},
	"deadline_optional":    {"en": "Deadline (optional)", "fr": "Date limite (optionnel)"},
	"max_file_size":        {"en": "Max file size (MB)", "fr": "Taille max du fichier (Mo)"},
	"max_chars":            {"en": "Max characters (0 = unlimited)", "fr": "Caractères max (0 = illimité)"},
	"max_grade":            {"en": "Max Grade", "fr": "Note maximale"},
	"create_assignment":    {"en": "Create Assignment", "fr": "Créer le devoir"},
	"submissions":          {"en": "submissions", "fr": "soumissions"},
	"graded":               {"en": "graded", "fr": "notées"},
	"due":                  {"en": "Due:", "fr": "Échéance :"},
	"view":                 {"en": "View", "fr": "Voir"},
	"profile":              {"en": "Profile", "fr": "Profil"},
	"view_profile_notes":   {"en": "View profile & add notes", "fr": "Voir profil & ajouter des notes"},
	"assign_empty":         {"en": "No assignments yet", "fr": "Aucun devoir pour le moment"},

	// Quizzes section
	"quiz_create":       {"en": "❓ Create Quiz", "fr": "❓ Créer un quiz"},
	"quiz_title_ph":     {"en": "Quiz title", "fr": "Titre du quiz"},
	"quiz_time_limit":   {"en": "Time Limit (minutes, 0=none)", "fr": "Durée limite (minutes, 0=aucune)"},
	"quiz_max_attempts": {"en": "Max Attempts (0=unlimited)", "fr": "Tentatives max (0=illimité)"},
	"create":            {"en": "Create", "fr": "Créer"},
	"published":         {"en": "Published", "fr": "Publié"},
	"draft":             {"en": "Draft", "fr": "Brouillon"},
	"questions":         {"en": "questions", "fr": "questions"},
	"attempts":          {"en": "attempts", "fr": "tentatives"},
	"attempt":           {"en": "attempt", "fr": "tentative"},
	"edit":              {"en": "Edit", "fr": "Modifier"},
	"cancel":            {"en": "Cancel", "fr": "Annuler"},
	"unpublish":         {"en": "Unpublish", "fr": "Dépublier"},
	"publish":           {"en": "Publish", "fr": "Publier"},
	"quiz_empty":        {"en": "No quizzes yet", "fr": "Aucun quiz pour le moment"},

	// Students section
	"students_pending":       {"en": "⏳ Pending Approval", "fr": "⏳ En attente d'approbation"},
	"approve":                {"en": "Approve", "fr": "Approuver"},
	"reject":                 {"en": "Reject", "fr": "Rejeter"},
	"reject_confirm":         {"en": "Reject this student?", "fr": "Rejeter cet étudiant ?"},
	"students_approved":      {"en": "Approved Students", "fr": "Étudiants approuvés"},
	"remove":                 {"en": "Remove", "fr": "Retirer"},
	"remove_confirm":         {"en": "Remove student?", "fr": "Retirer l'étudiant ?"},
	"optional":               {"en": "optional", "fr": "optionnel"},
	"replace":                {"en": "replace", "fr": "remplacer"},
	"teacher_attachment":     {"en": "Teacher attachment — download to complete this assignment", "fr": "Pièce jointe du professeur — téléchargez pour compléter ce devoir"},
	"students_empty":         {"en": "No approved students yet", "fr": "Aucun étudiant approuvé"},
	"students_preregistered": {"en": "📋 Pre-registered Students", "fr": "📋 Étudiants pré-inscrits"},
	"students_prereg_help":   {"en": "Students with these emails will be auto-approved when they join. Others go to pending.", "fr": "Les étudiants avec ces emails seront approuvés automatiquement. Les autres seront mis en attente."},
	"student_email_ph":       {"en": "student@email.com", "fr": "etudiant@email.com"},
	"name_optional":          {"en": "Name (optional)", "fr": "Nom (optionnel)"},
	"add_multiple":           {"en": "Add multiple emails at once", "fr": "Ajouter plusieurs emails à la fois"},
	"add_multiple_help":      {"en": "One per line. Format: email or email,name", "fr": "Un par ligne. Format : email ou email,nom"},
	"add_all":                {"en": "Add All", "fr": "Tout ajouter"},
	"prereg_empty":           {"en": "No pre-registered emails. All join requests will require manual approval.", "fr": "Aucun email pré-inscrit. Toutes les demandes nécessiteront une approbation manuelle."},

	// ─────────────────────────────────────────────────
	// QUIZ EDIT PAGE
	// ─────────────────────────────────────────────────
	"back_to":            {"en": "← Back to", "fr": "← Retour à"},
	"status":             {"en": "Status:", "fr": "Statut :"},
	"deadline":           {"en": "Deadline:", "fr": "Échéance :"},
	"time_limit":         {"en": "time limit", "fr": "durée limite"},
	"unlimited_attempts": {"en": "Unlimited attempts", "fr": "Tentatives illimitées"},

	// Quiz settings
	"quiz_settings":       {"en": "⚙️ Quiz Settings", "fr": "⚙️ Paramètres du quiz"},
	"quiz_desc_label":     {"en": "Description", "fr": "Description"},
	"quiz_deadline_label": {"en": "Deadline (optional)", "fr": "Échéance (optionnel)"},
	"quiz_time_label":     {"en": "Time limit (min, 0 = none)", "fr": "Durée limite (min, 0 = aucune)"},
	"quiz_attempts_label": {"en": "Max attempts (0 = unlimited)", "fr": "Tentatives max (0 = illimité)"},
	"save_settings":       {"en": "Save Settings", "fr": "Enregistrer"},

	// Questions
	"questions_heading":  {"en": "Questions", "fr": "Questions"},
	"delete_question":    {"en": "Delete question?", "fr": "Supprimer la question ?"},
	"true":               {"en": "True", "fr": "Vrai"},
	"false":              {"en": "False", "fr": "Faux"},
	"answer":             {"en": "Answer:", "fr": "Réponse :"},
	"open_ended_review":  {"en": "Open-ended — requires manual review", "fr": "Réponse libre — correction manuelle requise"},
	"sample":             {"en": "Sample:", "fr": "Exemple :"},
	"file_upload_review": {"en": "📎 File upload — requires manual review", "fr": "📎 Fichier — correction manuelle requise"},
	"no_questions":       {"en": "No questions added yet.", "fr": "Aucune question ajoutée."},

	// Add question form
	"add_question":     {"en": "Add New Question", "fr": "Ajouter une question"},
	"question_text":    {"en": "Question Text", "fr": "Texte de la question"},
	"question_text_ph": {"en": "Enter your question...", "fr": "Entrez votre question..."},
	"type":             {"en": "Type", "fr": "Type"},
	"type_mcq":         {"en": "Multiple Choice", "fr": "Choix multiple"},
	"type_tf":          {"en": "True / False", "fr": "Vrai / Faux"},
	"type_fill":        {"en": "Fill in the Blank", "fr": "Texte à trous"},
	"type_open":        {"en": "Open Ended (text)", "fr": "Réponse libre (texte)"},
	"type_file":        {"en": "File Upload", "fr": "Téléversement de fichier"},
	"points":           {"en": "Points", "fr": "Points"},
	"sort_order":       {"en": "Sort Order", "fr": "Ordre"},
	"options":          {"en": "Options", "fr": "Options"},
	"option_a":         {"en": "Option A", "fr": "Option A"},
	"option_b":         {"en": "Option B", "fr": "Option B"},
	"option_c":         {"en": "Option C", "fr": "Option C"},
	"option_d":         {"en": "Option D", "fr": "Option D"},
	"correct_answer":   {"en": "✓ Correct Answer", "fr": "✓ Bonne réponse"},
	"correct_match":    {"en": "Must match one of the options exactly", "fr": "Doit correspondre exactement à une des options"},
	"correct_tf":       {"en": "True or False", "fr": "Vrai ou Faux"},
	"correct_text":     {"en": "The correct answer text", "fr": "Le texte de la bonne réponse"},
	"sample_answer":    {"en": "Sample Answer (for reference)", "fr": "Réponse exemple (pour référence)"},
	"sample_answer_ph": {"en": "A sample correct answer...", "fr": "Un exemple de réponse correcte..."},
	"add_question_btn": {"en": "Add Question", "fr": "Ajouter la question"},

	// AI Generation
	"ai_generate":      {"en": "✨ Generate Questions with AI", "fr": "✨ Générer des questions avec l'IA"},
	"ai_topic_ph":      {"en": "Describe the topic...", "fr": "Décrivez le sujet..."},
	"ai_num_questions": {"en": "# Questions", "fr": "Nb questions"},
	"ai_difficulty":    {"en": "Difficulty", "fr": "Difficulté"},
	"ai_easy":          {"en": "Easy", "fr": "Facile"},
	"ai_intermediate":  {"en": "Intermediate", "fr": "Intermédiaire"},
	"ai_hard":          {"en": "Hard", "fr": "Difficile"},
	"ai_types":         {"en": "Types (comma-separated)", "fr": "Types (séparés par des virgules)"},
	"ai_generate_btn":  {"en": "✨ Generate", "fr": "✨ Générer"},

	// Math editor
	"insert_math":    {"en": "Insert Math", "fr": "Insérer une formule"},
	"pick_formula":   {"en": "Pick a formula or symbol to start", "fr": "Choisissez une formule ou un symbole"},
	"formulas":       {"en": "Formulas", "fr": "Formules"},
	"symbols_tab":    {"en": "Symbols", "fr": "Symboles"},
	"add_to_formula": {"en": "Add to formula", "fr": "Ajouter à la formule"},
	"insert":         {"en": "Insert", "fr": "Insérer"},

	// Student attempts
	"student_attempts": {"en": "Student Attempts", "fr": "Tentatives des étudiants"},
	"reviewed":         {"en": "Reviewed", "fr": "Corrigé"},
	"save":             {"en": "Save", "fr": "Enregistrer"},
	"view_answers":     {"en": "View answers", "fr": "Voir les réponses"},

	// ─────────────────────────────────────────────────
	// ADMIN SUBMISSIONS
	// ─────────────────────────────────────────────────
	"submissions_heading": {"en": "Submissions", "fr": "Soumissions"},
	"text_response":       {"en": "Text Response", "fr": "Réponse texte"},
	"attached_file":       {"en": "Attached File", "fr": "Fichier joint"},
	"teacher_feedback":    {"en": "Teacher Feedback", "fr": "Commentaire de l'enseignant"},
	"feedback_ph":         {"en": "Write feedback here...", "fr": "Écrivez votre commentaire ici..."},
	"grade_label":         {"en": "Grade", "fr": "Note"},
	"status_label":        {"en": "Status", "fr": "Statut"},
	"status_reviewed":     {"en": "✓ Reviewed", "fr": "✓ Corrigé"},
	"status_revision":     {"en": "⚠ Needs Revision", "fr": "⚠ À réviser"},
	"status_submitted":    {"en": "Submitted", "fr": "Soumis"},
	"save_review":         {"en": "Save Review", "fr": "Enregistrer la correction"},
	"no_submissions":      {"en": "No submissions yet for this assignment.", "fr": "Aucune soumission pour ce devoir."},

	// ─────────────────────────────────────────────────
	// ANALYTICS
	// ─────────────────────────────────────────────────
	"analytics_heading":  {"en": "📊 Analytics", "fr": "📊 Statistiques"},
	"print_report":       {"en": "🖨️ Print Report", "fr": "🖨️ Imprimer le rapport"},
	"export_roster":      {"en": "� Roster", "fr": "👥 Liste"},
	"export_quizzes":     {"en": "📝 Quiz Results", "fr": "📝 Résultats Quiz"},
	"export_assignments": {"en": "📋 Assignment Grades", "fr": "📋 Notes Devoirs"},
	"export_attendance":  {"en": "📅 Attendance", "fr": "📅 Présences"},

	// Sub-tabs
	"sub_quizzes":     {"en": "Quizzes", "fr": "Quiz"},
	"sub_assignments": {"en": "Assignments", "fr": "Devoirs"},
	"sub_roster":      {"en": "Roster", "fr": "Liste"},
	"sub_live":        {"en": "Live Classes", "fr": "Cours en direct"},
	"sub_trends":      {"en": "📈 Trends", "fr": "📈 Tendances"},
	"sub_risk":        {"en": "⚠️ At Risk", "fr": "⚠️ En difficulté"},
	"sub_resources":   {"en": "📦 Resources", "fr": "📦 Ressources"},

	// Quiz analytics
	"avg_score_by_quiz":   {"en": "Average Score by Quiz", "fr": "Score moyen par quiz"},
	"quiz":                {"en": "Quiz", "fr": "Quiz"},
	"n_students":          {"en": "students", "fr": "étudiants"},
	"highest":             {"en": "Highest", "fr": "Max"},
	"lowest":              {"en": "Lowest", "fr": "Min"},
	"average":             {"en": "Average", "fr": "Moyenne"},
	"details":             {"en": "Details", "fr": "Détails"},
	"view_arrow":          {"en": "View →", "fr": "Voir →"},
	"question_difficulty": {"en": "Question Difficulty", "fr": "Difficulté des questions"},
	"manual_review":       {"en": "Manual review required", "fr": "Correction manuelle requise"},
	"common_wrong":        {"en": "Most common wrong answer:", "fr": "Erreur la plus fréquente :"},
	"student_results":     {"en": "Student Results", "fr": "Résultats des étudiants"},
	"student":             {"en": "Student", "fr": "Étudiant"},
	"score":               {"en": "Score", "fr": "Score"},
	"percentage":          {"en": "Percentage", "fr": "Pourcentage"},
	"time_taken":          {"en": "Time Taken", "fr": "Durée"},
	"date":                {"en": "Date", "fr": "Date"},
	"no_quizzes":          {"en": "No quizzes in this classroom yet", "fr": "Aucun quiz dans cette classe"},

	// Assignment analytics
	"assignment":          {"en": "Assignment", "fr": "Devoir"},
	"max":                 {"en": "Max:", "fr": "Max :"},
	"not_graded":          {"en": "Not graded", "fr": "Non noté"},
	"missing":             {"en": "missing", "fr": "manquant(s)"},
	"all_submitted":       {"en": "✓ All submitted", "fr": "✓ Tous soumis"},
	"missing_submissions": {"en": "⚠️ Missing Submissions", "fr": "⚠️ Soumissions manquantes"},
	"all_submitted_msg":   {"en": "🎉 All students have submitted!", "fr": "🎉 Tous les étudiants ont soumis !"},
	"no_assignments":      {"en": "No assignments in this classroom yet", "fr": "Aucun devoir dans cette classe"},
	"distribution":        {"en": "Distribution", "fr": "Distribution"},

	// Roster analytics
	"avg_quiz_score":       {"en": "Avg Quiz Score", "fr": "Moy. Quiz"},
	"avg_assign_grade":     {"en": "Avg Assignment Grade", "fr": "Moy. Devoirs"},
	"avg_engagement":       {"en": "Avg Engagement", "fr": "Engagement moy."},
	"quiz_avg":             {"en": "Quiz Avg", "fr": "Moy. Quiz"},
	"quizzes_done":         {"en": "Quizzes Done", "fr": "Quiz faits"},
	"assign_avg":           {"en": "Assignment Avg", "fr": "Moy. Devoirs"},
	"submitted":            {"en": "Submitted", "fr": "Soumis"},
	"engagement":           {"en": "Engagement", "fr": "Engagement"},
	"no_approved_students": {"en": "No approved students in this classroom yet", "fr": "Aucun étudiant approuvé dans cette classe"},

	// Live classes analytics
	"live_session_history": {"en": "📹 Live Session History", "fr": "📹 Historique des cours en direct"},
	"started":              {"en": "Started", "fr": "Début"},
	"ended":                {"en": "Ended", "fr": "Fin"},
	"duration":             {"en": "Duration", "fr": "Durée"},
	"attendees":            {"en": "Attendees", "fr": "Participants"},
	"live_badge":           {"en": "● Live", "fr": "● En direct"},
	"ended_badge":          {"en": "Ended", "fr": "Terminé"},
	"active_badge":         {"en": "Active", "fr": "Actif"},
	"session_attendance":   {"en": "👥 Session Attendance", "fr": "👥 Présences de la session"},
	"joined_at":            {"en": "Joined At", "fr": "Arrivée"},
	"left_at":              {"en": "Left At", "fr": "Départ"},
	"time_spent":           {"en": "Time Spent", "fr": "Temps passé"},
	"still_in":             {"en": "Still in", "fr": "Toujours présent"},
	"no_attendance":        {"en": "No attendance records for this session", "fr": "Aucune donnée de présence pour cette session"},
	"attendance_rates":     {"en": "📊 Student Attendance Rates", "fr": "📊 Taux de présence des étudiants"},
	"sessions_attended":    {"en": "Sessions Attended", "fr": "Sessions présentes"},
	"attendance_rate":      {"en": "Attendance Rate", "fr": "Taux de présence"},
	"total_time":           {"en": "Total Time", "fr": "Temps total"},
	"no_live_sessions":     {"en": "No live sessions have been held yet", "fr": "Aucun cours en direct n'a été donné"},

	// Trends analytics
	"quiz_trend":        {"en": "📈 Quiz Score Trend (Class Average Over Time)", "fr": "📈 Tendance des scores (Moyenne de classe)"},
	"no_quiz_trends":    {"en": "No quiz data to show trends", "fr": "Pas assez de données pour les tendances quiz"},
	"assign_trend":      {"en": "📊 Assignment Grade Trend (Class Average Over Time)", "fr": "📊 Tendance des notes (Moyenne de classe)"},
	"no_assign_trends":  {"en": "No assignment data to show trends", "fr": "Pas assez de données pour les tendances devoirs"},
	"submission_timing": {"en": "⏰ Submission Timing Patterns", "fr": "⏰ Délais de soumission"},
	"early":             {"en": "Early (>24h)", "fr": "En avance (>24h)"},
	"on_time":           {"en": "On Time (<24h)", "fr": "À l'heure (<24h)"},
	"late":              {"en": "Late", "fr": "En retard"},
	"early_legend":      {"en": "Early (>24h before)", "fr": "En avance (>24h avant)"},
	"on_time_legend":    {"en": "On time (<24h before)", "fr": "À l'heure (<24h avant)"},
	"late_legend":       {"en": "Late (after deadline)", "fr": "En retard (après l'échéance)"},
	"no_deadline_subs":  {"en": "submissions (no deadline)", "fr": "soumissions (sans échéance)"},

	// At risk analytics
	"risk_desc":      {"en": "Students flagged for: missing 2+ assignments, scoring <50% on last 2 quizzes, or attendance below 50%.", "fr": "Étudiants signalés : 2+ devoirs manqués, <50% aux 2 derniers quiz, ou présence <50%."},
	"risk_score":     {"en": "⚠ Risk Score:", "fr": "⚠ Score de risque :"},
	"view_detail":    {"en": "View Detail →", "fr": "Voir le détail →"},
	"no_risk":        {"en": "No at-risk students detected!", "fr": "Aucun étudiant en difficulté détecté !"},
	"all_performing": {"en": "All students are performing well.", "fr": "Tous les étudiants se portent bien."},

	// Resources analytics
	"resource":              {"en": "Resource", "fr": "Ressource"},
	"total_views":           {"en": "Total Views", "fr": "Vues totales"},
	"unique_students":       {"en": "Unique Students", "fr": "Étudiants uniques"},
	"popularity":            {"en": "Popularity", "fr": "Popularité"},
	"no_views_yet":          {"en": "No views yet", "fr": "Aucune vue"},
	"views":                 {"en": "views", "fr": "vues"},
	"views_tracked":         {"en": "Views are tracked when students download or open resources.", "fr": "Les vues sont comptées lorsque les étudiants téléchargent ou ouvrent des ressources."},
	"no_resources_uploaded": {"en": "No resources uploaded yet", "fr": "Aucune ressource téléversée"},

	// ─────────────────────────────────────────────────
	// ADMIN REPORT
	// ─────────────────────────────────────────────────
	"report_heading":         {"en": "📚 TeachHub — Classroom Report", "fr": "📚 TeachHub — Rapport de classe"},
	"report_generated":       {"en": "Generated on", "fr": "Généré le"},
	"report_print":           {"en": "🖨️ Print / Save as PDF", "fr": "🖨️ Imprimer / Enregistrer en PDF"},
	"back_to_analytics":      {"en": "← Back to Analytics", "fr": "← Retour aux statistiques"},
	"at_risk_students":       {"en": "At-Risk Students", "fr": "Étudiants en difficulté"},
	"quiz_perf_report":       {"en": "📝 Quiz Performance", "fr": "📝 Performance aux quiz"},
	"assignment_overview":    {"en": "📋 Assignment Overview", "fr": "📋 Vue d'ensemble des devoirs"},
	"student_roster":         {"en": "👥 Student Roster Summary", "fr": "👥 Résumé de la liste des étudiants"},
	"at_risk_section":        {"en": "⚠️ At-Risk Students", "fr": "⚠️ Étudiants en difficulté"},
	"report_footer":          {"en": "TeachHub Classroom Report", "fr": "Rapport de classe TeachHub"},
	"rpt_attendance":         {"en": "Attendance", "fr": "Présence"},
	"rpt_quiz_results":       {"en": "📊 Quiz Results", "fr": "📊 Résultats des quiz"},
	"rpt_assign_results":     {"en": "📊 Assignment Results", "fr": "📊 Résultats des devoirs"},
	"rpt_student_cards":      {"en": "👤 Student Report Cards", "fr": "👤 Bulletins des étudiants"},
	"rpt_quizzes":            {"en": "Quizzes", "fr": "Quiz"},
	"rpt_assignments":        {"en": "Assignments", "fr": "Devoirs"},
	"rpt_overall":            {"en": "Overall", "fr": "Global"},
	"rpt_excellent":          {"en": "Excellent", "fr": "Excellent"},
	"rpt_good":               {"en": "Good", "fr": "Bien"},
	"rpt_average":            {"en": "Average", "fr": "Moyen"},
	"rpt_struggling":         {"en": "Needs Help", "fr": "En difficulté"},
	"rpt_submitted":          {"en": "submitted", "fr": "soumis"},
	"rpt_students":           {"en": "students", "fr": "étudiants"},
	"rpt_grade_dist":         {"en": "Grades", "fr": "Notes"},
	"rpt_needs_attention":    {"en": "⚠️ Students Needing Attention", "fr": "⚠️ Étudiants nécessitant une attention"},
	"rpt_class_avg":          {"en": "Class Average", "fr": "Moyenne de classe"},
	"rpt_sessions":           {"en": "sessions", "fr": "séances"},
	"rpt_attendance_summary": {"en": "📊 Attendance Summary", "fr": "📊 Résumé de présence"},
	"rpt_sessions_attended":  {"en": "Sessions Attended", "fr": "Séances présentes"},
	"rpt_total_minutes":      {"en": "Total Minutes", "fr": "Minutes totales"},

	// ─────────────────────────────────────────────────
	// ADMIN STUDENT DETAIL
	// ─────────────────────────────────────────────────
	"back_to_roster":        {"en": "← Back to Student Roster", "fr": "← Retour à la liste des étudiants"},
	"quiz_attempts_heading": {"en": "📝 Quiz Attempts", "fr": "📝 Tentatives de quiz"},
	"total":                 {"en": "total", "fr": "total"},
	"no_quiz_attempts":      {"en": "No quiz attempts yet", "fr": "Aucune tentative de quiz"},
	"quiz_over_time":        {"en": "Quiz Performance Over Time", "fr": "Performance aux quiz dans le temps"},
	"assign_submissions":    {"en": "📋 Assignment Submissions", "fr": "📋 Soumissions de devoirs"},
	"no_assign_subs":        {"en": "No assignment submissions yet", "fr": "Aucune soumission de devoir"},
	"teacher_notes":         {"en": "💬 Teacher Notes", "fr": "💬 Notes de l'enseignant"},
	"add_note_ph":           {"en": "Add a note for this student...", "fr": "Ajouter une note pour cet étudiant..."},
	"add_note":              {"en": "Add Note", "fr": "Ajouter"},
	"notes_visible":         {"en": "Notes are visible to the student in their dashboard.", "fr": "Les notes sont visibles par l'étudiant dans son tableau de bord."},
	"no_notes":              {"en": "No notes yet. Add one above.", "fr": "Aucune note. Ajoutez-en une ci-dessus."},
	"delete_note_confirm":   {"en": "Delete this note?", "fr": "Supprimer cette note ?"},

	// ─────────────────────────────────────────────────
	// STUDENT HOME
	// ─────────────────────────────────────────────────
	"my_classrooms":        {"en": "My Classrooms", "fr": "Mes classes"},
	"go_to_classroom":      {"en": "Go to classroom →", "fr": "Aller à la classe →"},
	"waiting_approval":     {"en": "Waiting for teacher approval", "fr": "En attente d'approbation"},
	"waiting_approval_sub": {"en": "You'll get access once the teacher accepts you", "fr": "Vous aurez accès dès que l'enseignant vous accepte"},
	"no_classrooms_joined": {"en": "You haven't joined any classrooms yet", "fr": "Vous n'avez rejoint aucune classe"},
	"use_join_code":        {"en": "Use the form below to enter a join code.", "fr": "Utilisez le formulaire ci-dessous pour entrer un code."},
	"join_classroom":       {"en": "Join a new classroom", "fr": "Rejoindre une nouvelle classe"},
	"join_code_desc":       {"en": "Enter the code provided by your teacher", "fr": "Entrez le code fourni par votre enseignant"},
	"join_code_ph":         {"en": "e.g. JOIN-123", "fr": "ex. JOIN-123"},
	"join":                 {"en": "Join", "fr": "Rejoindre"},

	// Cross-portal navigation
	"nav_or":                  {"en": "Or sign in as", "fr": "Ou connectez-vous en tant que"},
	"nav_student_portal":      {"en": "Student", "fr": "Élève"},
	"nav_teacher_login":       {"en": "Teacher Login", "fr": "Espace Enseignant"},
	"nav_teacher_login_desc":  {"en": "Sign in to manage your classrooms", "fr": "Connectez-vous pour gérer vos classes"},
	"nav_become_teacher":      {"en": "Become a Teacher", "fr": "Devenir Enseignant"},
	"nav_become_teacher_desc": {"en": "Apply to create your own classrooms", "fr": "Postulez pour créer vos propres classes"},
	"nav_platform_admin":      {"en": "Platform Admin", "fr": "Admin Plateforme"},
	"nav_platform_admin_desc": {"en": "Platform owner access", "fr": "Accès propriétaire de la plateforme"},

	// ─────────────────────────────────────────────────
	// LANDING PAGE
	// ─────────────────────────────────────────────────
	"land_tagline":        {"en": "The classroom platform for teachers", "fr": "La plateforme de classe pour les enseignants"},
	"land_badge":          {"en": "Built for the classroom", "fr": "Conçu pour la classe"},
	"land_hero_1":         {"en": "Your classroom, ", "fr": "Votre classe, "},
	"land_hero_highlight": {"en": "simplified.", "fr": "simplifiée."},
	"land_hero_2":         {"en": "", "fr": ""},
	"land_hero_desc":      {"en": "Create quizzes, share resources, run live sessions, and track student progress — all in one lightweight platform built for teachers.", "fr": "Créez des quiz, partagez des ressources, lancez des sessions en direct et suivez la progression des élèves — le tout dans une plateforme légère conçue pour les enseignants."},
	"land_cta_teacher":    {"en": "Apply as Teacher", "fr": "Postuler comme Enseignant"},
	"land_cta_student":    {"en": "I'm a Student", "fr": "Je suis Élève"},
	"land_cta_free":       {"en": "Free to apply · No credit card required", "fr": "Inscription gratuite · Aucune carte bancaire requise"},
	"land_teacher_signin": {"en": "Teacher Sign In", "fr": "Connexion Enseignant"},
	"land_apply_btn":      {"en": "Apply as Teacher", "fr": "Postuler"},

	// Student quick-join on landing
	"land_student_title": {"en": "Already have a classroom code?", "fr": "Vous avez un code de classe ?"},
	"land_student_desc":  {"en": "Enter the code your teacher gave you to join your classroom instantly.", "fr": "Entrez le code donné par votre enseignant pour rejoindre votre classe instantanément."},
	"land_student_go":    {"en": "Join", "fr": "Rejoindre"},

	// How it works
	"land_how_title":   {"en": "How it works", "fr": "Comment ça marche"},
	"land_how_desc":    {"en": "Get started in three simple steps.", "fr": "Commencez en trois étapes simples."},
	"land_step1_title": {"en": "Apply", "fr": "Postulez"},
	"land_step1_desc":  {"en": "Fill out a quick application with your school info. It takes less than 2 minutes.", "fr": "Remplissez une candidature rapide avec les infos de votre établissement. Ça prend moins de 2 minutes."},
	"land_step2_title": {"en": "Get Approved", "fr": "Soyez Approuvé"},
	"land_step2_desc":  {"en": "We review your application and create your teacher account with login credentials.", "fr": "Nous examinons votre candidature et créons votre compte enseignant avec vos identifiants."},
	"land_step3_title": {"en": "Start Teaching", "fr": "Enseignez"},
	"land_step3_desc":  {"en": "Create classrooms, invite students with a code, and start teaching right away.", "fr": "Créez des classes, invitez vos élèves avec un code, et commencez à enseigner immédiatement."},

	// Features
	"land_feat_title":    {"en": "Everything you need to teach", "fr": "Tout ce qu'il faut pour enseigner"},
	"land_feat_desc":     {"en": "Powerful tools designed to be simple and fast.", "fr": "Des outils puissants conçus pour être simples et rapides."},
	"land_feat_quiz_t":   {"en": "Quizzes & Exams", "fr": "Quiz & Examens"},
	"land_feat_quiz_d":   {"en": "MCQ, true/false, open-ended, and file upload questions with automatic grading.", "fr": "QCM, vrai/faux, questions ouvertes et téléversement de fichiers avec correction automatique."},
	"land_feat_assign_t": {"en": "Assignments", "fr": "Devoirs"},
	"land_feat_assign_d": {"en": "Collect student work with file uploads and text responses. Set deadlines and grade easily.", "fr": "Récupérez le travail des élèves par fichiers ou texte. Fixez des délais et notez facilement."},
	"land_feat_live_t":   {"en": "Live Sessions", "fr": "Sessions en Direct"},
	"land_feat_live_d":   {"en": "Real-time video classes with screen sharing. Students join with one click.", "fr": "Cours vidéo en temps réel avec partage d'écran. Les élèves rejoignent en un clic."},
	"land_feat_res_t":    {"en": "Resources", "fr": "Ressources"},
	"land_feat_res_d":    {"en": "Share PDFs, documents, videos, and links organized by category.", "fr": "Partagez PDF, documents, vidéos et liens organisés par catégorie."},
	"land_feat_ai_t":     {"en": "AI-Assisted Grading", "fr": "Correction Assistée par IA"},
	"land_feat_ai_d":     {"en": "Let AI help grade open-ended responses and provide feedback suggestions.", "fr": "Laissez l'IA aider à corriger les réponses ouvertes et suggérer des retours."},
	"land_feat_stats_t":  {"en": "Analytics", "fr": "Statistiques"},
	"land_feat_stats_d":  {"en": "Track student progress, quiz scores, and classroom performance at a glance.", "fr": "Suivez la progression des élèves, les notes de quiz et la performance de la classe en un coup d'œil."},

	// Final CTA & Footer
	"land_final_title":    {"en": "Ready to transform your classroom?", "fr": "Prêt à transformer votre classe ?"},
	"land_final_desc":     {"en": "Join teachers who are already using TeachHub to make teaching simpler.", "fr": "Rejoignez les enseignants qui utilisent déjà TeachHub pour simplifier l'enseignement."},
	"land_footer_tagline": {"en": "Made for teachers", "fr": "Fait pour les enseignants"},

	// Student home (logged out) — kept for student_home.html
	"home_heading":     {"en": "TeachHub", "fr": "TeachHub"},
	"home_subheading":  {"en": "Your portal for learning and assignments.", "fr": "Votre portail pour l'apprentissage et les devoirs."},
	"home_access":      {"en": "Access your classroom", "fr": "Accédez à votre classe"},
	"home_access_desc": {"en": "Enter your classroom join code to get started.", "fr": "Entrez le code de votre classe pour commencer."},
	"home_code_ph":     {"en": "JOIN CODE", "fr": "CODE D'ACCÈS"},
	"home_enter":       {"en": "Enter Classroom", "fr": "Accéder à la classe"},

	// ─────────────────────────────────────────────────
	// STUDENT JOIN
	// ─────────────────────────────────────────────────
	"join_heading":     {"en": "Join Classroom", "fr": "Rejoindre la classe"},
	"join_subheading":  {"en": "You are requesting to join", "fr": "Vous demandez à rejoindre"},
	"back_to_home":     {"en": "← Back to home", "fr": "← Retour à l'accueil"},
	"pending_heading":  {"en": "Request Pending", "fr": "Demande en attente"},
	"pending_text":     {"en": "Your join request has been sent to the teacher. You will be able to access the classroom once approved.", "fr": "Votre demande a été envoyée à l'enseignant. Vous pourrez accéder à la classe une fois approuvé."},
	"go_to_dashboard":  {"en": "Go to Dashboard", "fr": "Aller au tableau de bord"},
	"rejected_heading": {"en": "Request Rejected", "fr": "Demande rejetée"},
	"rejected_text":    {"en": "Your join request was declined by the teacher.", "fr": "Votre demande a été refusée par l'enseignant."},
	"full_name":        {"en": "Your Full Name", "fr": "Votre nom complet"},
	"full_name_ph":     {"en": "John Doe", "fr": "Mohamed Amine"},
	"email_label":      {"en": "Email Address", "fr": "Adresse email"},
	"email_ph":         {"en": "student@example.com", "fr": "etudiant@exemple.com"},
	"email_help":       {"en": "If your teacher pre-registered you, use that exact email to be approved instantly.", "fr": "Si votre enseignant vous a pré-inscrit, utilisez cet email exact pour être approuvé automatiquement."},
	"phone_label":      {"en": "Phone Number (parent)", "fr": "Numéro de téléphone (parent)"},
	"phone_ph":         {"en": "0612345678", "fr": "0612345678"},
	"phone_help":       {"en": "10-digit phone number so the teacher can contact you.", "fr": "Numéro à 10 chiffres pour que l'enseignant puisse vous contacter."},
	"request_join":     {"en": "Request to Join", "fr": "Demander à rejoindre"},

	// ─────────────────────────────────────────────────
	// STUDENT CLASSROOM
	// ─────────────────────────────────────────────────
	"live_in_progress":     {"en": "Live class in progress!", "fr": "Cours en direct en cours !"},
	"join_now":             {"en": "📹 Join Now", "fr": "📹 Rejoindre"},
	"tab_my_progress":      {"en": "📊 My Progress", "fr": "📊 Ma progression"},
	"stab_resources":       {"en": "📁 Resources", "fr": "📁 Ressources"},
	"stab_assignments":     {"en": "📝 Assignments", "fr": "📝 Devoirs"},
	"stab_quizzes":         {"en": "❓ Quizzes", "fr": "❓ Quiz"},
	"no_deadline":          {"en": "No deadline", "fr": "Sans échéance"},
	"take_quiz":            {"en": "Take quiz →", "fr": "Commencer le quiz →"},
	"no_resources_avail":   {"en": "No resources available yet.", "fr": "Aucune ressource disponible."},
	"no_assignments_avail": {"en": "No assignments yet.", "fr": "Aucun devoir pour le moment."},
	"no_quizzes_avail":     {"en": "No quizzes available.", "fr": "Aucun quiz disponible."},

	// ─────────────────────────────────────────────────
	// STUDENT ASSIGNMENT
	// ─────────────────────────────────────────────────
	"deadline_passed":        {"en": "⏰ The deadline has passed. Submissions are no longer accepted.", "fr": "⏰ La date limite est passée. Les soumissions ne sont plus acceptées."},
	"submit_work":            {"en": "Submit your work", "fr": "Soumettre votre travail"},
	"your_answer":            {"en": "Your Answer", "fr": "Votre réponse"},
	"type_answer_ph":         {"en": "Type your answer here...", "fr": "Tapez votre réponse ici..."},
	"attach_file":            {"en": "Attach File", "fr": "Joindre un fichier"},
	"max_file_help":          {"en": "Max file size:", "fr": "Taille max :"},
	"submit_assignment":      {"en": "Submit Assignment", "fr": "Soumettre le devoir"},
	"your_submissions":       {"en": "Your submissions", "fr": "Vos soumissions"},
	"s_reviewed":             {"en": "✓ Reviewed", "fr": "✓ Corrigé"},
	"s_revision":             {"en": "⚠️ Needs Revision", "fr": "⚠️ À réviser"},
	"s_submitted":            {"en": "⏳ Submitted", "fr": "⏳ Soumis"},
	"teacher_feedback_label": {"en": "Teacher Feedback:", "fr": "Commentaire de l'enseignant :"},

	// ─────────────────────────────────────────────────
	// STUDENT QUIZ
	// ─────────────────────────────────────────────────
	"attempts_used":         {"en": "attempts used", "fr": "tentatives utilisées"},
	"previous_attempts":     {"en": "Your Previous Attempts", "fr": "Vos tentatives précédentes"},
	"attempt_n":             {"en": "Attempt", "fr": "Tentative"},
	"pending_review":        {"en": "Pending review", "fr": "En attente de correction"},
	"no_file_uploaded":      {"en": "No file uploaded", "fr": "Aucun fichier envoyé"},
	"your_answer_label":     {"en": "Your answer:", "fr": "Votre réponse :"},
	"correct_label":         {"en": "Correct:", "fr": "Correct :"},
	"time_remaining":        {"en": "⏱ Time remaining:", "fr": "⏱ Temps restant :"},
	"questions_to_complete": {"en": "questions to complete", "fr": "questions à compléter"},
	"submit_quiz":           {"en": "Submit Quiz", "fr": "Soumettre le quiz"},
	"times_up":              {"en": "Time's up! Submitting...", "fr": "Temps écoulé ! Soumission en cours..."},
	"quiz_locked":           {"en": "This quiz is not available right now.", "fr": "Ce quiz n'est pas disponible actuellement."},
	"true_label":            {"en": "✓ True", "fr": "✓ Vrai"},
	"false_label":           {"en": "✗ False", "fr": "✗ Faux"},
	"fill_blank_ph":         {"en": "Type your answer...", "fr": "Tapez votre réponse..."},
	"open_ended_ph":         {"en": "Write your detailed answer here...", "fr": "Rédigez votre réponse détaillée ici..."},
	"file_upload_label":     {"en": "Upload your file (PDF, image, document)", "fr": "Téléversez votre fichier (PDF, image, document)"},

	// ─────────────────────────────────────────────────
	// STUDENT DASHBOARD
	// ─────────────────────────────────────────────────
	"my_progress":        {"en": "📊 My Progress", "fr": "📊 Ma progression"},
	"join_live":          {"en": "📹 Join", "fr": "📹 Rejoindre"},
	"quiz_avg_card":      {"en": "Quiz Avg", "fr": "Moy. Quiz"},
	"assign_avg_card":    {"en": "Assignment Avg", "fr": "Moy. Devoirs"},
	"above_avg":          {"en": "↑ Above", "fr": "↑ Au-dessus"},
	"below_avg":          {"en": "↓ Below", "fr": "↓ En-dessous"},
	"class_avg":          {"en": "class avg", "fr": "moy. classe"},
	"attendance_card":    {"en": "Attendance", "fr": "Présence"},
	"sessions":           {"en": "sessions", "fr": "sessions"},
	"teacher_notes_card": {"en": "Teacher Notes", "fr": "Notes enseignant"},

	// Dashboard sub-tabs
	"dtab_overview":    {"en": "Overview", "fr": "Aperçu"},
	"dtab_quizzes":     {"en": "Quiz Scores", "fr": "Scores quiz"},
	"dtab_assignments": {"en": "Assignments", "fr": "Devoirs"},
	"dtab_attendance":  {"en": "Attendance", "fr": "Présence"},
	"dtab_notes":       {"en": "Teacher Notes", "fr": "Notes enseignant"},

	// Overview sub-tab
	"quiz_performance":     {"en": "Quiz Performance", "fr": "Performance aux quiz"},
	"recent_assign_grades": {"en": "Recent Assignment Grades", "fr": "Notes récentes des devoirs"},
	"pending_status":       {"en": "Pending", "fr": "En attente"},
	"needs_revision":       {"en": "Needs revision", "fr": "À réviser"},
	"latest_teacher_notes": {"en": "Latest Teacher Notes", "fr": "Dernières notes de l'enseignant"},
	"no_data_yet":          {"en": "No performance data yet. Start taking quizzes and submitting assignments!", "fr": "Pas encore de données. Commencez les quiz et soumettez vos devoirs !"},

	// Quiz scores sub-tab
	"your_average":       {"en": "Your average:", "fr": "Votre moyenne :"},
	"class_average":      {"en": "Class average:", "fr": "Moyenne de classe :"},
	"above_average":      {"en": "↑ Above average", "fr": "↑ Au-dessus de la moyenne"},
	"below_average":      {"en": "↓ Below average", "fr": "↓ En-dessous de la moyenne"},
	"no_quiz_attempts_s": {"en": "No quiz attempts yet.", "fr": "Aucune tentative de quiz."},

	// Assignments sub-tab
	"no_assign_subs_s": {"en": "No assignment submissions yet.", "fr": "Aucune soumission de devoir."},

	// Attendance sub-tab
	"attendance_rate_card":   {"en": "Attendance Rate", "fr": "Taux de présence"},
	"sessions_attended_card": {"en": "Sessions Attended", "fr": "Sessions présentes"},
	"total_sessions_card":    {"en": "Total Sessions", "fr": "Sessions totales"},
	"attendance_progress":    {"en": "Attendance Progress", "fr": "Progression de présence"},
	"present":                {"en": "✓ Present", "fr": "✓ Présent"},
	"absent":                 {"en": "✗ Absent", "fr": "✗ Absent"},
	"min":                    {"en": "min", "fr": "min"},
	"no_live_sessions_s":     {"en": "No live sessions have been held yet.", "fr": "Aucun cours en direct n'a été donné."},

	// Teacher notes sub-tab
	"no_teacher_notes": {"en": "No teacher notes yet.", "fr": "Aucune note de l'enseignant."},
	"no_notes_desc":    {"en": "Your teacher can leave feedback and notes here.", "fr": "Votre enseignant peut laisser des commentaires et notes ici."},

	// ─────────────────────────────────────────────────
	// LIVE VIDEO (teacher & student)
	// ─────────────────────────────────────────────────
	"end_class":              {"en": "End Class", "fr": "Terminer le cours"},
	"cam_requests":           {"en": "📷 Requests", "fr": "📷 Demandes"},
	"participants":           {"en": "👥 Participants", "fr": "👥 Participants"},
	"mute_all":               {"en": "🔇 Mute All", "fr": "🔇 Tout couper"},
	"you_teacher":            {"en": "You (Teacher)", "fr": "Vous (Enseignant)"},
	"chat":                   {"en": "💬 Chat", "fr": "💬 Chat"},
	"type_message":           {"en": "Type a message...", "fr": "Écrivez un message..."},
	"send":                   {"en": "Send", "fr": "Envoyer"},
	"connecting":             {"en": "Connecting...", "fr": "Connexion..."},
	"connected_waiting":      {"en": "Connected ✓", "fr": "Connecté ✓"},
	"connected_waiting_hint": {"en": "Waiting for the teacher to share their screen or camera", "fr": "En attente que l'enseignant partage son écran ou sa caméra"},
	"connection_failed":      {"en": "Connection Failed", "fr": "Connexion échouée"},
	"end_class_confirm":      {"en": "End the live class for everyone?", "fr": "Terminer le cours en direct pour tout le monde ?"},
	"teacher_label":          {"en": "Teacher", "fr": "Enseignant"},
	"teacher_screen":         {"en": "Teacher's Screen", "fr": "Écran de l'enseignant"},
	"connected_ready":        {"en": "Connected ✓ — Ready", "fr": "Connecté ✓ — Prêt"},
	"connected_ready_hint":   {"en": "Your mic is live — turn on camera when ready", "fr": "Votre micro est actif — activez la caméra quand vous êtes prêt"},
	"camera_off_label":       {"en": "Camera is off", "fr": "Caméra désactivée"},
	"upload_profile_pic":     {"en": "Upload profile picture", "fr": "Télécharger photo de profil"},
	"change_profile_pic":     {"en": "Change", "fr": "Changer"},
	"requests_camera":        {"en": "requests camera access", "fr": "demande l'accès caméra"},

	// Live toasts
	"toast_joined":       {"en": "joined", "fr": "a rejoint"},
	"toast_left":         {"en": "left", "fr": "a quitté"},
	"toast_muted":        {"en": "Muted", "fr": "Micro coupé pour"},
	"toast_unmuted":      {"en": "Unmuted", "fr": "Micro réactivé pour"},
	"toast_all_muted":    {"en": "All students muted", "fr": "Tous les étudiants en sourdine"},
	"toast_cam_approved": {"en": "Camera approved for", "fr": "Caméra approuvée pour"},
	"toast_cam_denied":   {"en": "Camera denied for", "fr": "Caméra refusée pour"},
	"toast_cam_off":      {"en": "Turned off camera for", "fr": "Caméra désactivée pour"},

	// Student live
	"microphone":             {"en": "Microphone", "fr": "Microphone"},
	"camera":                 {"en": "Camera", "fr": "Caméra"},
	"camera_ask":             {"en": "Camera (ask teacher)", "fr": "Caméra (demander)"},
	"request_camera":         {"en": "Request Camera", "fr": "Demander la caméra"},
	"request_sent":           {"en": "Sent...", "fr": "Envoyé..."},
	"leave":                  {"en": "Leave", "fr": "Quitter"},
	"cam_approved_msg":       {"en": "Camera approved! Tap to turn on.", "fr": "Caméra approuvée ! Appuyez pour activer."},
	"cam_denied_msg":         {"en": "Camera request denied", "fr": "Demande de caméra refusée"},
	"cam_off_teacher":        {"en": "Camera turned off by teacher", "fr": "Caméra désactivée par l'enseignant"},
	"cam_waiting":            {"en": "Waiting for teacher approval...", "fr": "En attente d'approbation..."},
	"toast_cam_sent":         {"en": "Camera request sent", "fr": "Demande de caméra envoyée"},
	"toast_cam_approved_s":   {"en": "Camera approved!", "fr": "Caméra approuvée !"},
	"toast_cam_denied_s":     {"en": "Camera denied", "fr": "Caméra refusée"},
	"toast_mic_muted":        {"en": "Teacher muted your mic", "fr": "L'enseignant a coupé votre micro"},
	"toast_mic_unmuted":      {"en": "Teacher unmuted your mic", "fr": "L'enseignant a réactivé votre micro"},
	"toast_all_muted_s":      {"en": "Teacher muted all students", "fr": "L'enseignant a coupé tous les micros"},
	"unmute_request_title":   {"en": "Teacher is asking you to unmute", "fr": "L'enseignant vous demande de réactiver votre micro"},
	"unmute_accept":          {"en": "Unmute", "fr": "Réactiver"},
	"unmute_decline":         {"en": "Stay muted", "fr": "Rester muet"},
	"toast_unmute_declined":  {"en": "declined unmute request", "fr": "a refusé la demande de micro"},
	"toast_unmute_requested": {"en": "Unmute request sent to", "fr": "Demande de micro envoyée à"},
	"toast_cam_off_s":        {"en": "Teacher turned off your camera", "fr": "L'enseignant a désactivé votre caméra"},
	"toast_class_ended":      {"en": "Teacher ended the class", "fr": "L'enseignant a terminé le cours"},
	"toast_ask_teacher":      {"en": "Ask teacher for permission first", "fr": "Demandez d'abord l'autorisation à l'enseignant"},

	// Raise hand
	"raise_hand":           {"en": "Raise Hand", "fr": "Lever la main"},
	"lower_hand":           {"en": "Lower Hand", "fr": "Baisser la main"},
	"hand_raised":          {"en": "✋ Raised Hands", "fr": "✋ Mains levées"},
	"toast_hand_raised":    {"en": "raised their hand", "fr": "a levé la main"},
	"toast_hand_lowered":   {"en": "lowered their hand", "fr": "a baissé la main"},
	"toast_hand_lowered_t": {"en": "Teacher lowered your hand", "fr": "L'enseignant a baissé votre main"},
	"lower_all_hands":      {"en": "Lower All", "fr": "Tout baisser"},

	// Screen share mobile
	"screen_share_pc_only": {"en": "Screen sharing requires a computer", "fr": "Le partage d'écran nécessite un ordinateur"},

	// Live Poll
	"poll_create":     {"en": "📊 Create Poll", "fr": "📊 Créer un sondage"},
	"poll_question":   {"en": "Question", "fr": "Question"},
	"poll_option":     {"en": "Option", "fr": "Option"},
	"poll_add_option": {"en": "+ Add Option", "fr": "+ Ajouter une option"},
	"poll_timer":      {"en": "Time limit (seconds)", "fr": "Durée (secondes)"},
	"poll_no_timer":   {"en": "No limit", "fr": "Sans limite"},
	"poll_launch":     {"en": "Launch Poll", "fr": "Lancer le sondage"},
	"poll_close":      {"en": "Close Poll", "fr": "Fermer le sondage"},
	"poll_results":    {"en": "Poll Results", "fr": "Résultats du sondage"},
	"poll_votes":      {"en": "votes", "fr": "votes"},
	"poll_vote":       {"en": "vote", "fr": "vote"},
	"poll_voted":      {"en": "Vote submitted!", "fr": "Vote envoyé !"},
	"poll_ended":      {"en": "Poll ended", "fr": "Sondage terminé"},
	"poll_active":     {"en": "📊 Active Poll", "fr": "📊 Sondage actif"},
	"poll_time_left":  {"en": "Time left", "fr": "Temps restant"},
	"poll_total":      {"en": "Total", "fr": "Total"},

	// Pinned message
	"pin_message":    {"en": "Pin", "fr": "Épingler"},
	"unpin_message":  {"en": "Unpin", "fr": "Désépingler"},
	"pinned":         {"en": "📌 Pinned", "fr": "📌 Épinglé"},
	"delete_message": {"en": "Delete message", "fr": "Supprimer le message"},

	// Image presenter
	"share_image":        {"en": "Share Image", "fr": "Partager une image"},
	"stop_sharing_image": {"en": "Stop Sharing", "fr": "Arrêter le partage"},
	"image_shared":       {"en": "shared an image", "fr": "a partagé une image"},
	"image_stopped":      {"en": "stopped sharing image", "fr": "a arrêté le partage d'image"},
	"uploading":          {"en": "Uploading...", "fr": "Envoi..."},

	// Whiteboard
	"whiteboard":    {"en": "Whiteboard", "fr": "Tableau blanc"},
	"wb_pen":        {"en": "Pen", "fr": "Stylo"},
	"wb_marker":     {"en": "Marker", "fr": "Marqueur"},
	"wb_eraser":     {"en": "Eraser", "fr": "Gomme"},
	"wb_undo":       {"en": "Undo", "fr": "Annuler"},
	"wb_redo":       {"en": "Redo", "fr": "Rétablir"},
	"wb_clear":      {"en": "Clear", "fr": "Effacer"},
	"wb_select":     {"en": "Select", "fr": "Sélection"},
	"wb_delete_img": {"en": "Delete Image", "fr": "Supprimer l'image"},
	"wb_bg_image":   {"en": "Background Image", "fr": "Image de fond"},
	"wb_close":      {"en": "Close Whiteboard", "fr": "Fermer le tableau"},
	"wb_zoom_in":    {"en": "Zoom In", "fr": "Zoom avant"},
	"wb_zoom_out":   {"en": "Zoom Out", "fr": "Zoom arrière"},
	"wb_zoom_reset": {"en": "Reset Zoom", "fr": "Réinitialiser le zoom"},
	"wb_opened":     {"en": "Whiteboard opened", "fr": "Tableau blanc ouvert"},
	"wb_closed":     {"en": "Whiteboard closed", "fr": "Tableau blanc fermé"},
	"wb_cleared":    {"en": "Whiteboard cleared", "fr": "Tableau effacé"},
	"wb_load_pdf":   {"en": "Load PDF", "fr": "Charger PDF"},
	"wb_pdf_page":   {"en": "Page", "fr": "Page"},
	"wb_pdf_of":     {"en": "of", "fr": "sur"},
	"wb_pdf_prev":   {"en": "Previous Page", "fr": "Page précédente"},
	"wb_pdf_next":   {"en": "Next Page", "fr": "Page suivante"},
	"wb_pdf_loaded": {"en": "PDF loaded", "fr": "PDF chargé"},
	"wb_pdf_close":  {"en": "Close PDF", "fr": "Fermer PDF"},

	// ─────────────────────────────────────────────────
	// LANGUAGE TOGGLE
	// ─────────────────────────────────────────────────
	"switch_to_fr": {"en": "🇫🇷 Français", "fr": "🇫🇷 Français"},
	"switch_to_en": {"en": "🇬🇧 English", "fr": "🇬🇧 English"},

	// ─────────────────────────────────────────────────
	// MISC / SHARED
	// ─────────────────────────────────────────────────
	"chars": {"en": "chars", "fr": "car."},
	"file":  {"en": "file", "fr": "fichier"},
	"text":  {"en": "text", "fr": "texte"},
	"both":  {"en": "both", "fr": "les deux"},
	"of":    {"en": "of", "fr": "sur"},
	"no":    {"en": "No", "fr": "Non"},
	"yes":   {"en": "Yes", "fr": "Oui"},

	// ─────────────────────────────────────────────────
	// PLATFORM: TEACHER APPLICATION
	// ─────────────────────────────────────────────────
	"apply_title":           {"en": "Become a Teacher", "fr": "Devenir enseignant"},
	"apply_heading":         {"en": "Join TeachHub as a Teacher", "fr": "Rejoignez TeachHub en tant qu'enseignant"},
	"apply_subheading":      {"en": "Apply for a teacher account and start managing your classrooms", "fr": "Demandez un compte enseignant et commencez à gérer vos classes"},
	"apply_already_teacher": {"en": "Already a teacher? Sign in →", "fr": "Déjà enseignant ? Connexion →"},
	"apply_full_name":       {"en": "Full Name", "fr": "Nom complet"},
	"apply_name_ph":         {"en": "Mohamed Amine Bourega", "fr": "Mohamed Amine Bourega"},
	"apply_email":           {"en": "Email Address", "fr": "Adresse email"},
	"apply_email_ph":        {"en": "teacher@school.dz", "fr": "enseignant@ecole.dz"},
	"apply_phone":           {"en": "Phone Number", "fr": "Numéro de téléphone"},
	"apply_phone_ph":        {"en": "0555 12 34 56", "fr": "0555 12 34 56"},
	"apply_school":          {"en": "School / University", "fr": "École / Université"},
	"apply_school_ph":       {"en": "Your school or university", "fr": "Votre école ou université"},
	"apply_wilaya":          {"en": "Wilaya", "fr": "Wilaya"},
	"apply_select_wilaya":   {"en": "Select your wilaya...", "fr": "Sélectionnez votre wilaya..."},
	"apply_message":         {"en": "Why do you want to use TeachHub?", "fr": "Pourquoi souhaitez-vous utiliser TeachHub ?"},
	"apply_message_ph":      {"en": "Tell us about your teaching needs...", "fr": "Parlez-nous de vos besoins pédagogiques..."},
	"apply_submit":          {"en": "Submit Application", "fr": "Envoyer la demande"},
	"apply_footer":          {"en": "We'll review your application and contact you within 48 hours.", "fr": "Nous examinerons votre demande et vous contacterons sous 48 heures."},
	"apply_error_required":  {"en": "Full name and email are required.", "fr": "Le nom complet et l'email sont obligatoires."},
	"apply_error_failed":    {"en": "Something went wrong. Please try again.", "fr": "Une erreur est survenue. Veuillez réessayer."},

	// Features
	"apply_feat1_title": {"en": "Classrooms", "fr": "Classes"},
	"apply_feat1_desc":  {"en": "Create and manage classrooms with quizzes, assignments, and resources", "fr": "Créez et gérez des classes avec quiz, devoirs et ressources"},
	"apply_feat2_title": {"en": "Analytics", "fr": "Statistiques"},
	"apply_feat2_desc":  {"en": "Track student performance with detailed charts and reports", "fr": "Suivez les performances avec des graphiques et rapports détaillés"},
	"apply_feat3_title": {"en": "Live Classes", "fr": "Cours en direct"},
	"apply_feat3_desc":  {"en": "Host live video sessions with real-time interaction", "fr": "Animez des sessions vidéo en direct avec interaction temps réel"},

	// Success page
	"apply_success_title":   {"en": "Application Sent", "fr": "Demande envoyée"},
	"apply_success_heading": {"en": "Application Submitted!", "fr": "Demande envoyée !"},
	"apply_success_text1":   {"en": "Thank you for your interest in TeachHub.", "fr": "Merci pour votre intérêt pour TeachHub."},
	"apply_success_text2":   {"en": "Our team will review your application and get back to you soon.", "fr": "Notre équipe examinera votre demande et vous recontactera bientôt."},
	"apply_success_next":    {"en": "What happens next?", "fr": "Et maintenant ?"},
	"apply_success_step1":   {"en": "We review your application (24-48h)", "fr": "Nous examinons votre demande (24-48h)"},
	"apply_success_step2":   {"en": "We contact you to discuss your needs and payment", "fr": "Nous vous contactons pour discuter de vos besoins et du paiement"},
	"apply_success_step3":   {"en": "You receive your teacher credentials and start teaching!", "fr": "Vous recevez vos identifiants et commencez à enseigner !"},
	"apply_success_home":    {"en": "← Back to Home", "fr": "← Retour à l'accueil"},
	"apply_success_another": {"en": "Submit Another", "fr": "Autre demande"},

	// ─────────────────────────────────────────────────
	// PLATFORM: ADMIN DASHBOARD
	// ─────────────────────────────────────────────────
	"plat_nav_dashboard":        {"en": "Dashboard", "fr": "Tableau de bord"},
	"plat_nav_applications":     {"en": "Applications", "fr": "Candidatures"},
	"plat_login_title":          {"en": "Sign In", "fr": "Connexion"},
	"plat_login_subtitle":       {"en": "Platform administration area", "fr": "Espace d'administration de la plateforme"},
	"plat_login_footer":         {"en": "Platform owner access only", "fr": "Accès réservé au propriétaire"},
	"plat_dash_heading":         {"en": "Platform Dashboard", "fr": "Tableau de bord"},
	"plat_dash_subheading":      {"en": "Overview of teacher applications and platform activity", "fr": "Aperçu des candidatures et de l'activité de la plateforme"},
	"plat_pending":              {"en": "Pending", "fr": "En attente"},
	"plat_contacted":            {"en": "Contacted", "fr": "Contacté"},
	"plat_approved":             {"en": "Approved", "fr": "Approuvé"},
	"plat_rejected":             {"en": "Rejected", "fr": "Rejeté"},
	"plat_total":                {"en": "Total", "fr": "Total"},
	"plat_action_review":        {"en": "Review Pending Applications", "fr": "Examiner les candidatures"},
	"plat_action_review_desc":   {"en": "applications waiting for review", "fr": "candidatures en attente"},
	"plat_action_all_apps":      {"en": "All Applications", "fr": "Toutes les candidatures"},
	"plat_action_all_apps_desc": {"en": "total applications", "fr": "candidatures au total"},

	// Applications list
	"plat_apps_heading":    {"en": "Teacher Applications", "fr": "Candidatures enseignants"},
	"plat_apps_subheading": {"en": "Review and manage teacher registration requests", "fr": "Examinez et gérez les demandes d'inscription enseignant"},
	"plat_filter_all":      {"en": "All", "fr": "Toutes"},
	"plat_apps_empty":      {"en": "No applications matching this filter", "fr": "Aucune candidature correspondant à ce filtre"},

	// Application detail
	"plat_detail_applied":  {"en": "Applied on", "fr": "Candidature le"},
	"plat_detail_saved":    {"en": "Changes saved successfully!", "fr": "Modifications enregistrées !"},
	"plat_detail_update":   {"en": "Update Application Status", "fr": "Mettre à jour le statut"},
	"plat_detail_notes":    {"en": "Internal Notes", "fr": "Notes internes"},
	"plat_detail_notes_ph": {"en": "Payment received, credentials sent via WhatsApp...", "fr": "Paiement reçu, identifiants envoyés par WhatsApp..."},
	"plat_detail_save":     {"en": "Save Changes", "fr": "Enregistrer"},

	// ─────────────────────────────────────────────────
	// PLATFORM: TEACHERS & CREDENTIALS
	// ─────────────────────────────────────────────────
	"plat_nav_teachers":         {"en": "Teachers", "fr": "Enseignants"},
	"plat_action_teachers_desc": {"en": "Manage active teacher accounts", "fr": "Gérer les comptes enseignants actifs"},

	// Credentials page
	"plat_cred_title":       {"en": "Teacher Credentials", "fr": "Identifiants enseignant"},
	"plat_cred_heading":     {"en": "Account Created!", "fr": "Compte créé !"},
	"plat_cred_subheading":  {"en": "Share these credentials with the teacher securely", "fr": "Partagez ces identifiants avec l'enseignant de manière sécurisée"},
	"plat_cred_reset_title": {"en": "Password Reset", "fr": "Réinitialisation du mot de passe"},
	"plat_cred_reset_sub":   {"en": "New credentials generated — share with the teacher", "fr": "Nouveaux identifiants générés — à partager avec l'enseignant"},
	"plat_cred_credentials": {"en": "Login Credentials", "fr": "Identifiants de connexion"},
	"plat_cred_warning":     {"en": "Save these credentials now! The password cannot be retrieved later.", "fr": "Sauvegardez ces identifiants maintenant ! Le mot de passe ne pourra pas être récupéré plus tard."},
	"plat_cred_login_url":   {"en": "Login URL", "fr": "URL de connexion"},
	"plat_cred_back_app":    {"en": "Back to Application", "fr": "Retour à la candidature"},

	// Teachers list
	"plat_teachers_heading":    {"en": "Teachers", "fr": "Enseignants"},
	"plat_teachers_subheading": {"en": "Manage teacher accounts and subscriptions", "fr": "Gérez les comptes enseignants et abonnements"},
	"plat_teachers_classrooms": {"en": "classrooms", "fr": "classes"},
	"plat_teachers_students":   {"en": "students", "fr": "étudiants"},
	"plat_teachers_quizzes":    {"en": "quizzes", "fr": "quiz"},
	"plat_teachers_resources":  {"en": "resources", "fr": "ressources"},
	"plat_teachers_since":      {"en": "Since", "fr": "Depuis"},
	"plat_teachers_empty":      {"en": "No teachers yet", "fr": "Aucun enseignant pour le moment"},
	"plat_teachers_empty_sub":  {"en": "Approve an application to create a teacher account", "fr": "Approuvez une candidature pour créer un compte enseignant"},
	"plat_never_logged_in":     {"en": "Never logged in", "fr": "Jamais connecté"},

	// Teachers list page — stats & filters
	"plat_stat_total_teachers": {"en": "Total Teachers", "fr": "Total enseignants"},
	"plat_stat_active":         {"en": "Active Subscriptions", "fr": "Abonnements actifs"},
	"plat_stat_logged_30d":     {"en": "Logged in (30 days)", "fr": "Connectés (30 jours)"},
	"plat_stat_never_logged":   {"en": "Never Logged In", "fr": "Jamais connectés"},
	"plat_search_teachers":     {"en": "Search by name, email, school or IP", "fr": "Rechercher par nom, email, école ou IP"},
	"plat_filter_never_logged": {"en": "Never Logged", "fr": "Jamais connecté"},
	"plat_filter_recent":       {"en": "Recent (30d)", "fr": "Récents (30j)"},
	"plat_col_teacher":         {"en": "Teacher", "fr": "Enseignant"},
	"plat_col_email":           {"en": "Email", "fr": "Email"},
	"plat_col_students":        {"en": "Students", "fr": "Étudiants"},
	"plat_col_classes":         {"en": "Classes", "fr": "Classes"},
	"plat_col_last_login":      {"en": "Last Login", "fr": "Dernière connexion"},
	"plat_col_status":          {"en": "Status", "fr": "Statut"},
	"plat_no_results":          {"en": "No teachers match your search", "fr": "Aucun enseignant ne correspond à votre recherche"},
	"plat_last_login_label":    {"en": "Last Login", "fr": "Dernière connexion"},
	"plat_login_ip_label":      {"en": "IP Address", "fr": "Adresse IP"},

	// Subscription status
	"plat_sub_active":    {"en": "Active", "fr": "Actif"},
	"plat_sub_suspended": {"en": "Suspended", "fr": "Suspendu"},
	"plat_sub_expired":   {"en": "Expired", "fr": "Expiré"},

	// Teacher detail
	"plat_teacher_start":           {"en": "Since", "fr": "Depuis"},
	"plat_teacher_end":             {"en": "Expires", "fr": "Expire"},
	"plat_teacher_actions":         {"en": "Actions", "fr": "Actions"},
	"plat_teacher_suspend":         {"en": "Suspend Account", "fr": "Suspendre le compte"},
	"plat_teacher_activate":        {"en": "Activate Account", "fr": "Activer le compte"},
	"plat_teacher_reset_pw":        {"en": "Reset Password", "fr": "Réinitialiser le mot de passe"},
	"plat_teacher_confirm_suspend": {"en": "Are you sure you want to suspend this teacher?", "fr": "Êtes-vous sûr de vouloir suspendre cet enseignant ?"},
	"plat_teacher_confirm_reset":   {"en": "Generate a new password for this teacher?", "fr": "Générer un nouveau mot de passe pour cet enseignant ?"},

	// Login error (subscription)
	"login_error_suspended": {"en": "Your account has been suspended. Please contact the platform administrator.", "fr": "Votre compte a été suspendu. Veuillez contacter l'administrateur de la plateforme."},
	"login_error_expired":   {"en": "Your subscription has expired. Please contact the platform administrator to renew.", "fr": "Votre abonnement a expiré. Veuillez contacter l'administrateur pour le renouveler."},

	// ─────────────────────────────────────────────────
	// PLATFORM: DASHBOARD STATS (Phase 4)
	// ─────────────────────────────────────────────────
	"plat_dash_active":         {"en": "Active Teachers", "fr": "Enseignants actifs"},
	"plat_dash_suspended":      {"en": "Suspended", "fr": "Suspendus"},
	"plat_dash_revenue":        {"en": "Total Revenue", "fr": "Revenu total"},
	"plat_dash_monthly":        {"en": "This Month", "fr": "Ce mois"},
	"plat_dash_expiring_title": {"en": "Subscriptions Expiring Soon", "fr": "Abonnements expirant bientôt"},
	"plat_dash_expiring_desc":  {"en": "teacher(s) expiring within 7 days", "fr": "enseignant(s) expirant dans 7 jours"},
	"plat_dash_view_teachers":  {"en": "View Teachers", "fr": "Voir les enseignants"},

	// ─────────────────────────────────────────────────
	// PLATFORM: SUBSCRIPTION MANAGEMENT (Phase 4)
	// ─────────────────────────────────────────────────
	"plat_sub_manage":       {"en": "Subscription Management", "fr": "Gestion de l'abonnement"},
	"plat_sub_extend_quick": {"en": "Quick Extend", "fr": "Extension rapide"},
	"plat_sub_month":        {"en": "month", "fr": "mois"},
	"plat_sub_months":       {"en": "months", "fr": "mois"},
	"plat_sub_year":         {"en": "year", "fr": "an"},
	"plat_sub_set_end":      {"en": "Set Custom End Date", "fr": "Définir une date de fin"},
	"plat_sub_set":          {"en": "Set", "fr": "Définir"},

	// ─────────────────────────────────────────────────
	// PLATFORM: PAYMENT TRACKING (Phase 4)
	// ─────────────────────────────────────────────────
	"plat_pay_record":         {"en": "Record Payment", "fr": "Enregistrer un paiement"},
	"plat_pay_amount":         {"en": "Amount", "fr": "Montant"},
	"plat_pay_method":         {"en": "Payment Method", "fr": "Méthode de paiement"},
	"plat_pay_cash":           {"en": "Cash", "fr": "Espèces"},
	"plat_pay_other":          {"en": "Other", "fr": "Autre"},
	"plat_pay_reference":      {"en": "Reference", "fr": "Référence"},
	"plat_pay_ref_ph":         {"en": "Transaction ID, receipt number...", "fr": "N° de transaction, reçu..."},
	"plat_pay_notes_ph":       {"en": "Optional notes...", "fr": "Notes optionnelles..."},
	"plat_pay_submit":         {"en": "Record Payment", "fr": "Enregistrer le paiement"},
	"plat_pay_history":        {"en": "Payment History", "fr": "Historique des paiements"},
	"plat_pay_date":           {"en": "Date", "fr": "Date"},
	"plat_pay_total_paid":     {"en": "Total Paid", "fr": "Total payé"},
	"plat_pay_count":          {"en": "Payments", "fr": "Paiements"},
	"plat_pay_total_label":    {"en": "Total", "fr": "Total"},
	"plat_pay_empty":          {"en": "No payments recorded yet", "fr": "Aucun paiement enregistré"},
	"plat_pay_confirm_delete": {"en": "Delete this payment record?", "fr": "Supprimer cet enregistrement de paiement ?"},

	// ─────────────────────────────────────────────────
	// PLATFORM: ANALYTICS (Phase 5)
	// ─────────────────────────────────────────────────
	"plat_nav_analytics":          {"en": "Analytics", "fr": "Statistiques"},
	"plat_analytics_title":        {"en": "Platform Analytics", "fr": "Statistiques de la plateforme"},
	"plat_analytics_subtitle":     {"en": "Business intelligence and platform growth", "fr": "Intelligence d'affaires et croissance de la plateforme"},
	"plat_analytics_teachers":     {"en": "Total Teachers", "fr": "Enseignants"},
	"plat_analytics_students":     {"en": "Total Students", "fr": "Étudiants"},
	"plat_analytics_classrooms":   {"en": "Classrooms", "fr": "Classes"},
	"plat_analytics_quizzes":      {"en": "Quizzes", "fr": "Quiz"},
	"plat_analytics_revenue":      {"en": "Revenue", "fr": "Revenus"},
	"plat_analytics_total":        {"en": "total", "fr": "total"},
	"plat_analytics_app_trend":    {"en": "Applications Trend (6 months)", "fr": "Tendance des candidatures (6 mois)"},
	"plat_analytics_top_teachers": {"en": "Teachers", "fr": "Enseignants"},
	"plat_analytics_teacher":      {"en": "Teacher", "fr": "Enseignant"},
	"plat_analytics_school":       {"en": "School", "fr": "École"},
	"plat_analytics_status":       {"en": "Status", "fr": "Statut"},

	// ─────────────────────────────────────────────────
	// PLATFORM: CSV EXPORTS (Phase 5)
	// ─────────────────────────────────────────────────
	"plat_export_teachers": {"en": "Export Teachers CSV", "fr": "Exporter enseignants CSV"},
	"plat_export_payments": {"en": "Export Payments CSV", "fr": "Exporter paiements CSV"},

	// ─────────────────────────────────────────────────
	// PLATFORM: PASSWORD CHANGE (Phase 5)
	// ─────────────────────────────────────────────────
	"plat_change_password":   {"en": "Change Password", "fr": "Changer le mot de passe"},
	"plat_pw_current":        {"en": "Current Password", "fr": "Mot de passe actuel"},
	"plat_pw_new":            {"en": "New Password", "fr": "Nouveau mot de passe"},
	"plat_pw_confirm":        {"en": "Confirm New Password", "fr": "Confirmer le nouveau mot de passe"},
	"plat_pw_submit":         {"en": "Update Password", "fr": "Mettre à jour"},
	"plat_pw_success":        {"en": "Password updated successfully!", "fr": "Mot de passe mis à jour !"},
	"plat_pw_error_wrong":    {"en": "Current password is incorrect.", "fr": "Le mot de passe actuel est incorrect."},
	"plat_pw_error_mismatch": {"en": "New passwords do not match.", "fr": "Les nouveaux mots de passe ne correspondent pas."},
	"plat_pw_error_short":    {"en": "Password must be at least 6 characters.", "fr": "Le mot de passe doit contenir au moins 6 caractères."},

	// ─────────────────────────────────────────────────
	// PARENT REPORT & SHARING
	// ─────────────────────────────────────────────────
	"parent_report":      {"en": "Parent Progress Report", "fr": "Suivi pour les parents"},
	"parent_report_desc": {"en": "Share this link with the parent to give them read-only access to their child's progress.", "fr": "Partagez ce lien avec le parent pour lui donner accès en lecture au suivi de son enfant."},
	"share_whatsapp":     {"en": "Share via WhatsApp", "fr": "Partager via WhatsApp"},
	"copy_link":          {"en": "Copy Link", "fr": "Copier le lien"},
	"copy_parent_link":   {"en": "Copy parent report link", "fr": "Copier le lien du suivi parent"},
	"copied":             {"en": "Copied!", "fr": "Copié !"},
	"regen_code":         {"en": "Regenerate link", "fr": "Régénérer le lien"},
	"regen_confirm":      {"en": "Regenerate link? The old link will stop working.", "fr": "Régénérer le lien ? L'ancien lien ne fonctionnera plus."},

	// ─── CGU / Privacy ──────────────────────────────────
	"cgu_title":        {"en": "Terms of Use & Privacy Policy", "fr": "Conditions d'Utilisation & Politique de Confidentialité"},
	"cgu_link":         {"en": "Terms & Privacy", "fr": "CGU & Confidentialité"},
	"cgu_back_home":    {"en": "Home", "fr": "Accueil"},
	"cgu_last_updated": {"en": "Last updated", "fr": "Dernière mise à jour"},
	"cgu_april":        {"en": "April", "fr": "Avril"},
	"cgu_tab_terms":    {"en": "Terms of Use", "fr": "Conditions d'Utilisation"},
	"cgu_tab_privacy":  {"en": "Privacy Policy", "fr": "Politique de Confidentialité"},

	// Terms sections
	"cgu_s1_title": {"en": "Acceptance of Terms", "fr": "Acceptation des conditions"},
	"cgu_s1_text":  {"en": "By accessing or using TeachHub, you agree to be bound by these Terms of Use. If you do not agree, please do not use the platform. These terms apply to all users: teachers, students, and parents.", "fr": "En accédant ou en utilisant TeachHub, vous acceptez d'être lié par ces Conditions d'Utilisation. Si vous n'êtes pas d'accord, veuillez ne pas utiliser la plateforme. Ces conditions s'appliquent à tous les utilisateurs : enseignants, élèves et parents."},

	"cgu_s2_title": {"en": "Description of Service", "fr": "Description du service"},
	"cgu_s2_text":  {"en": "TeachHub is an educational platform that allows teachers to create virtual classrooms, share resources, assign work, create quizzes, run live sessions, and track student progress. Students can join classrooms, submit work, and take quizzes. Parents can view their child's progress reports via a shared link.", "fr": "TeachHub est une plateforme éducative qui permet aux enseignants de créer des classes virtuelles, partager des ressources, donner des devoirs, créer des quiz, animer des sessions en direct et suivre la progression des élèves. Les élèves peuvent rejoindre des classes, soumettre des travaux et passer des quiz. Les parents peuvent consulter les rapports de progression de leur enfant via un lien partagé."},

	"cgu_s3_title": {"en": "User Accounts", "fr": "Comptes utilisateurs"},
	"cgu_s3_text":  {"en": "Different users access TeachHub in different ways:", "fr": "Les différents utilisateurs accèdent à TeachHub de différentes manières :"},
	"cgu_s3_li1":   {"en": "Teachers: must apply and be approved by the platform administrator. Login credentials are provided upon approval.", "fr": "Enseignants : doivent postuler et être approuvés par l'administrateur de la plateforme. Les identifiants de connexion sont fournis lors de l'approbation."},
	"cgu_s3_li2":   {"en": "Students: join classrooms using a code shared by their teacher. A name and email are required to create a student session.", "fr": "Élèves : rejoignent les classes à l'aide d'un code partagé par leur enseignant. Un nom et un e-mail sont requis pour créer une session élève."},
	"cgu_s3_li3":   {"en": "Parents: access progress reports via a unique link shared by the teacher. No account is required.", "fr": "Parents : accèdent aux rapports de progression via un lien unique partagé par l'enseignant. Aucun compte n'est requis."},
	"cgu_s3_li4":   {"en": "You are responsible for keeping your login credentials confidential.", "fr": "Vous êtes responsable de la confidentialité de vos identifiants de connexion."},

	"cgu_s4_title": {"en": "Acceptable Use", "fr": "Utilisation acceptable"},
	"cgu_s4_text":  {"en": "You agree not to:", "fr": "Vous vous engagez à ne pas :"},
	"cgu_s4_li1":   {"en": "Use the platform for any illegal or unauthorized purpose.", "fr": "Utiliser la plateforme à des fins illégales ou non autorisées."},
	"cgu_s4_li2":   {"en": "Upload content that is offensive, harmful, or violates the rights of others.", "fr": "Téléverser du contenu offensant, nuisible ou portant atteinte aux droits d'autrui."},
	"cgu_s4_li3":   {"en": "Attempt to access accounts or data belonging to other users.", "fr": "Tenter d'accéder aux comptes ou aux données d'autres utilisateurs."},

	"cgu_s5_title": {"en": "Intellectual Property", "fr": "Propriété intellectuelle"},
	"cgu_s5_text":  {"en": "Teachers retain ownership of the content they create and upload (quizzes, resources, assignments). TeachHub does not claim ownership of user-generated content. The TeachHub platform, its design, and code are the property of TeachHub.", "fr": "Les enseignants conservent la propriété du contenu qu'ils créent et téléversent (quiz, ressources, devoirs). TeachHub ne revendique pas la propriété du contenu généré par les utilisateurs. La plateforme TeachHub, son design et son code sont la propriété de TeachHub."},

	"cgu_s6_title": {"en": "Limitation of Liability", "fr": "Limitation de responsabilité"},
	"cgu_s6_text":  {"en": "TeachHub is provided \"as is\" without warranty of any kind. We strive to keep the platform available and reliable, but we cannot guarantee uninterrupted access. TeachHub shall not be liable for any loss of data or any indirect damages arising from the use of the platform.", "fr": "TeachHub est fourni « en l'état » sans aucune garantie. Nous nous efforçons de maintenir la plateforme disponible et fiable, mais nous ne pouvons pas garantir un accès ininterrompu. TeachHub ne saurait être tenu responsable de toute perte de données ou de tout dommage indirect résultant de l'utilisation de la plateforme."},

	"cgu_s7_title": {"en": "Changes to Terms", "fr": "Modifications des conditions"},
	"cgu_s7_text":  {"en": "We may update these terms from time to time. Continued use of the platform after changes constitutes acceptance of the new terms. We will indicate the date of the last update at the top of this page.", "fr": "Nous pouvons mettre à jour ces conditions de temps à autre. L'utilisation continue de la plateforme après les modifications constitue l'acceptation des nouvelles conditions. Nous indiquerons la date de la dernière mise à jour en haut de cette page."},

	// Privacy sections — RGPD compliant
	"priv_s1_title": {"en": "Data Controller", "fr": "Responsable du traitement"},
	"priv_s1_text":  {"en": "TeachHub is operated as an independent educational platform. For any questions regarding your personal data, you can contact us at:", "fr": "TeachHub est exploité en tant que plateforme éducative indépendante. Pour toute question relative à vos données personnelles, vous pouvez nous contacter à :"},
	"priv_s1_li1":   {"en": "Email: contact@teachhub.chickenkiller.com", "fr": "E-mail : contact@teachhub.chickenkiller.com"},

	"priv_s2_title": {"en": "Data We Collect & Legal Basis", "fr": "Données collectées & base légale"},
	"priv_s2_text":  {"en": "We collect only the minimum data necessary. For each type of data, we indicate the legal basis under RGPD/GDPR (Article 6):", "fr": "Nous collectons uniquement les données minimales nécessaires. Pour chaque type de données, nous indiquons la base légale selon le RGPD (Article 6) :"},
	"priv_s2_li1":   {"en": "Teacher applications: name, email, phone, school — Legal basis: contract (Art. 6(1)(b)). Retained until account deletion or application rejection + 6 months.", "fr": "Candidatures enseignant : nom, e-mail, téléphone, établissement — Base légale : contrat (Art. 6(1)(b)). Conservées jusqu'à la suppression du compte ou le rejet de la candidature + 6 mois."},
	"priv_s2_li2":   {"en": "Teacher accounts: username, hashed password, email — Legal basis: contract (Art. 6(1)(b)). Retained while the account is active.", "fr": "Comptes enseignant : identifiant, mot de passe haché, e-mail — Base légale : contrat (Art. 6(1)(b)). Conservés tant que le compte est actif."},
	"priv_s2_li3":   {"en": "Students: name and email (provided when joining a classroom) — Legal basis: legitimate interest in providing educational services (Art. 6(1)(f)). Retained while the classroom exists.", "fr": "Élèves : nom et e-mail (fournis lors de l'inscription) — Base légale : intérêt légitime pour la fourniture de services éducatifs (Art. 6(1)(f)). Conservés tant que la classe existe."},
	"priv_s2_li4":   {"en": "Academic data: quiz responses, assignment submissions, grades, attendance, resource views — Legal basis: legitimate interest (Art. 6(1)(f)). Deleted when the classroom is deleted.", "fr": "Données académiques : réponses aux quiz, soumissions de devoirs, notes, assiduité, consultations de ressources — Base légale : intérêt légitime (Art. 6(1)(f)). Supprimées avec la classe."},
	"priv_s2_li5":   {"en": "Uploaded files: resources, assignment files, student submissions — Deleted when the classroom is deleted.", "fr": "Fichiers : ressources, fichiers de devoirs, soumissions — Supprimés avec la classe."},
	"priv_s2_li6":   {"en": "Teacher notes about students (remarks visible on progress reports) — Legal basis: legitimate interest (Art. 6(1)(f)).", "fr": "Notes de l'enseignant sur les élèves (remarques visibles sur les rapports) — Base légale : intérêt légitime (Art. 6(1)(f))."},

	"priv_s3_title": {"en": "Minors' Data", "fr": "Données des mineurs"},
	"priv_s3_text":  {"en": "TeachHub is designed for educational use and students may be minors. Students join classrooms via a code provided by their teacher within a school context. The teacher, acting within their professional duties, is responsible for ensuring appropriate use. We collect only the minimum data necessary (name, email) and do not profile students for commercial purposes. Parents can view their child's data via a progress report link provided by the teacher, and may contact us to exercise rights on behalf of their child.", "fr": "TeachHub est conçu pour un usage éducatif et les élèves peuvent être mineurs. Les élèves rejoignent les classes via un code fourni par leur enseignant dans un cadre scolaire. L'enseignant, agissant dans le cadre de ses fonctions professionnelles, est responsable de l'utilisation appropriée. Nous collectons uniquement les données minimales nécessaires (nom, e-mail) et ne profilons pas les élèves à des fins commerciales. Les parents peuvent consulter les données de leur enfant via un lien de rapport fourni par l'enseignant, et peuvent nous contacter pour exercer des droits au nom de leur enfant."},

	"priv_s4_title": {"en": "Cookies", "fr": "Cookies"},
	"priv_s4_text":  {"en": "TeachHub uses the following cookies:", "fr": "TeachHub utilise les cookies suivants :"},
	"priv_s4_li1":   {"en": "Session cookies (teachhub-session, teachhub-platform): strictly necessary for authentication. Duration: 7 days. No consent required.", "fr": "Cookies de session (teachhub-session, teachhub-platform) : strictement nécessaires à l'authentification. Durée : 7 jours. Pas de consentement requis."},
	"priv_s4_li2":   {"en": "CSRF cookie (teachhub-csrf): strictly necessary for security (prevents cross-site request forgery). Duration: 7 days. No consent required.", "fr": "Cookie CSRF (teachhub-csrf) : strictement nécessaire à la sécurité (protection contre les attaques CSRF). Durée : 7 jours. Pas de consentement requis."},
	"priv_s4_li3":   {"en": "Language preference cookie (lang): stores your language choice (fr/en). Duration: 1 year. Set only with your consent.", "fr": "Cookie de préférence linguistique (lang) : enregistre votre choix de langue (fr/en). Durée : 1 an. Déposé uniquement avec votre consentement."},
	"priv_s4_li4":   {"en": "We do not use any analytics, advertising, or tracking cookies.", "fr": "Nous n'utilisons aucun cookie d'analyse, de publicité ou de traçage."},
	"priv_s4_li5":   {"en": "We record your last login IP address and timestamp for security purposes (fraud prevention). This data is overwritten on each login and is not shared with third parties.", "fr": "Nous enregistrons votre dernière adresse IP de connexion et l'horodatage à des fins de sécurité (prévention de la fraude). Cette donnée est écrasée à chaque connexion et n'est pas partagée avec des tiers."},

	"priv_s5_title": {"en": "Third-Party Services", "fr": "Services tiers"},
	"priv_s5_text":  {"en": "To deliver the platform, we use the following third-party services that may receive your IP address when loading resources:", "fr": "Pour fournir la plateforme, nous utilisons les services tiers suivants qui peuvent recevoir votre adresse IP lors du chargement des ressources :"},
	"priv_s5_li1":   {"en": "jsDelivr CDN (Prospect One, EU/Global): delivers the KaTeX math rendering library and MathLive editor. Privacy policy: jsdelivr.com/terms/privacy-policy-jsdelivr-net", "fr": "jsDelivr CDN (Prospect One, UE/Global) : distribue la bibliothèque de rendu mathématique KaTeX et l'éditeur MathLive. Politique de confidentialité : jsdelivr.com/terms/privacy-policy-jsdelivr-net"},
	"priv_s5_li2":   {"en": "Quill CDN (cdn.quilljs.com via jsDelivr): delivers the Quill rich text editor stylesheet.", "fr": "Quill CDN (cdn.quilljs.com via jsDelivr) : distribue la feuille de style de l'éditeur de texte Quill."},
	"priv_s5_li3":   {"en": "Fonts are self-hosted — no data is sent to Google or other font providers.", "fr": "Les polices sont hébergées localement — aucune donnée n'est envoyée à Google ou à d'autres fournisseurs de polices."},
	"priv_s5_li4":   {"en": "We do not sell, rent, or share your personal data with third parties for marketing purposes.", "fr": "Nous ne vendons, louons ni ne partageons vos données personnelles avec des tiers à des fins marketing."},

	"priv_s6_title": {"en": "Data Security", "fr": "Sécurité des données"},
	"priv_s6_text":  {"en": "We implement appropriate technical measures to protect your data:", "fr": "Nous mettons en œuvre des mesures techniques appropriées pour protéger vos données :"},
	"priv_s6_li1":   {"en": "All connections are encrypted via HTTPS (TLS).", "fr": "Toutes les connexions sont chiffrées via HTTPS (TLS)."},
	"priv_s6_li2":   {"en": "Passwords are irreversibly hashed using bcrypt.", "fr": "Les mots de passe sont hachés de manière irréversible avec bcrypt."},
	"priv_s6_li3":   {"en": "Access to student data is restricted to their classroom teacher.", "fr": "L'accès aux données des élèves est limité à l'enseignant de leur classe."},
	"priv_s6_li4":   {"en": "Parent progress reports are accessible only via a unique private link.", "fr": "Les rapports de progression des parents ne sont accessibles que via un lien privé unique."},

	"priv_s7_title": {"en": "Your Rights (RGPD Articles 15–22)", "fr": "Vos droits (RGPD Articles 15–22)"},
	"priv_s7_text":  {"en": "Under the RGPD/GDPR, you have the following rights regarding your personal data:", "fr": "En vertu du RGPD, vous disposez des droits suivants concernant vos données personnelles :"},
	"priv_s7_li1":   {"en": "Right of access (Art. 15): obtain a copy of your personal data.", "fr": "Droit d'accès (Art. 15) : obtenir une copie de vos données personnelles."},
	"priv_s7_li2":   {"en": "Right to rectification (Art. 16): correct inaccurate data.", "fr": "Droit de rectification (Art. 16) : corriger des données inexactes."},
	"priv_s7_li3":   {"en": "Right to erasure (Art. 17): request deletion of your data.", "fr": "Droit à l'effacement (Art. 17) : demander la suppression de vos données."},
	"priv_s7_li4":   {"en": "Right to data portability (Art. 20): receive your data in a structured format.", "fr": "Droit à la portabilité (Art. 20) : recevoir vos données dans un format structuré."},
	"priv_s7_li5":   {"en": "Right to object (Art. 21): object to data processing based on legitimate interest.", "fr": "Droit d'opposition (Art. 21) : s'opposer au traitement basé sur l'intérêt légitime."},
	"priv_s7_li6":   {"en": "Right to restrict processing (Art. 18): request limitation of processing.", "fr": "Droit à la limitation du traitement (Art. 18) : demander la limitation du traitement."},
	"priv_s7_li7":   {"en": "To exercise any of these rights, contact us at: contact@teachhub.chickenkiller.com", "fr": "Pour exercer l'un de ces droits, contactez-nous à : contact@teachhub.chickenkiller.com"},
	"priv_s7_li8":   {"en": "You also have the right to lodge a complaint with your local data protection supervisory authority.", "fr": "Vous disposez également du droit d'introduire une réclamation auprès de votre autorité locale de protection des données."},

	"priv_s8_title": {"en": "Data Retention", "fr": "Conservation des données"},
	"priv_s8_text":  {"en": "Data is retained only as long as necessary:", "fr": "Les données ne sont conservées que le temps nécessaire :"},
	"priv_s8_li1":   {"en": "Student data and academic records: deleted when the classroom is deleted by the teacher.", "fr": "Données des élèves et dossiers académiques : supprimés lorsque la classe est supprimée par l'enseignant."},
	"priv_s8_li2":   {"en": "Teacher accounts: retained while active. Teachers can request full account deletion.", "fr": "Comptes enseignant : conservés tant qu'ils sont actifs. Les enseignants peuvent demander la suppression complète."},
	"priv_s8_li3":   {"en": "Rejected applications: deleted after 6 months.", "fr": "Candidatures rejetées : supprimées après 6 mois."},
	"priv_s8_li4":   {"en": "Payment records: retained for legal accounting obligations.", "fr": "Données de paiement : conservées pour les obligations comptables légales."},

	"priv_s9_title": {"en": "Contact & Complaints", "fr": "Contact & réclamations"},
	"priv_s9_text":  {"en": "For any questions about this policy or to exercise your rights, contact us at contact@teachhub.chickenkiller.com. If you believe your rights have not been respected, you may lodge a complaint with the relevant data protection authority.", "fr": "Pour toute question sur cette politique ou pour exercer vos droits, contactez-nous à contact@teachhub.chickenkiller.com. Si vous estimez que vos droits n'ont pas été respectés, vous pouvez déposer une réclamation auprès de l'autorité de protection des données compétente."},

	// Cookie consent banner
	"cookie_msg":     {"en": "This site uses a cookie to save your language preference.", "fr": "Ce site utilise un cookie pour enregistrer votre préférence de langue."},
	"cookie_accept":  {"en": "Accept", "fr": "Accepter"},
	"cookie_decline": {"en": "Decline", "fr": "Refuser"},

	// ─────────────────────────────────────────────────
	// EXPLORE (public teacher directory)
	// ─────────────────────────────────────────────────
	"explore_title":          {"en": "Find a Teacher", "fr": "Trouver un enseignant"},
	"explore_nav_badge":      {"en": "Explore", "fr": "Explorer"},
	"explore_nav_student":    {"en": "Student Portal", "fr": "Espace élève"},
	"explore_heading":        {"en": "Find Your Teacher", "fr": "Trouvez votre enseignant"},
	"explore_subheading":     {"en": "Browse verified teachers and request to join their classrooms", "fr": "Parcourez les enseignants vérifiés et demandez à rejoindre leurs classes"},
	"explore_country":        {"en": "Country", "fr": "Pays"},
	"explore_subject":        {"en": "Subject", "fr": "Matière"},
	"explore_level":          {"en": "Level", "fr": "Niveau"},
	"explore_all":            {"en": "All", "fr": "Tous"},
	"explore_search":         {"en": "Search", "fr": "Rechercher"},
	"explore_results":        {"en": "teachers found", "fr": "enseignants trouvés"},
	"explore_classrooms":     {"en": "classrooms", "fr": "classes"},
	"explore_empty":          {"en": "No teachers found", "fr": "Aucun enseignant trouvé"},
	"explore_empty_sub":      {"en": "Try adjusting your filters", "fr": "Essayez de modifier vos filtres"},
	"explore_back":           {"en": "← Back to Explore", "fr": "← Retour à l'exploration"},
	"explore_algeria":        {"en": "Algeria", "fr": "Algérie"},
	"explore_france":         {"en": "France", "fr": "France"},
	"explore_subjects_taught": {"en": "Subjects Taught", "fr": "Matières enseignées"},
	"explore_levels_taught":  {"en": "Levels Taught", "fr": "Niveaux enseignés"},
	"explore_classrooms_title": {"en": "Classrooms", "fr": "Classes"},
	"explore_join_cta":       {"en": "Request to Join", "fr": "Demander à rejoindre"},
	"explore_join_heading":   {"en": "Join Request", "fr": "Demande d'inscription"},
	"explore_join_name":      {"en": "Full Name", "fr": "Nom complet"},
	"explore_join_name_ph":   {"en": "Your full name", "fr": "Votre nom complet"},
	"explore_join_email":     {"en": "Email", "fr": "E-mail"},
	"explore_join_email_ph":  {"en": "your@email.com", "fr": "votre@email.com"},
	"explore_join_phone":     {"en": "Phone (optional)", "fr": "Téléphone (optionnel)"},
	"explore_join_phone_ph":  {"en": "+213...", "fr": "+213..."},
	"explore_join_classroom": {"en": "Preferred Classroom", "fr": "Classe souhaitée"},
	"explore_join_classroom_none": {"en": "No preference — teacher will assign", "fr": "Pas de préférence — l'enseignant assignera"},
	"explore_join_level":     {"en": "Your Level", "fr": "Votre niveau"},
	"explore_join_message":   {"en": "Message (optional)", "fr": "Message (optionnel)"},
	"explore_join_message_ph": {"en": "Introduce yourself...", "fr": "Présentez-vous..."},
	"explore_join_submit":    {"en": "Send Request", "fr": "Envoyer la demande"},
	"explore_join_success_title": {"en": "Request Sent!", "fr": "Demande envoyée !"},
	"explore_join_success_text":  {"en": "Your join request has been sent to the teacher. You will receive an email when they review your request.", "fr": "Votre demande a été envoyée à l'enseignant. Vous recevrez un e-mail lorsqu'il examinera votre demande."},
	"explore_join_success_back":  {"en": "← Back to teacher", "fr": "← Retour à l'enseignant"},

	// ─────────────────────────────────────────────────
	// TEACHER PROFILE SETTINGS
	// ─────────────────────────────────────────────────
	"nav_profile":         {"en": "Profile", "fr": "Profil"},
	"nav_requests":        {"en": "Requests", "fr": "Demandes"},
	"profile_heading":     {"en": "Public Profile", "fr": "Profil public"},
	"profile_subheading":  {"en": "Control how students find you on Explore", "fr": "Contrôlez comment les élèves vous trouvent sur Explorer"},
	"profile_public_toggle": {"en": "Visible on Explore", "fr": "Visible sur Explorer"},
	"profile_public_desc": {"en": "When enabled, students can find you and request to join your classrooms", "fr": "Lorsque activé, les élèves peuvent vous trouver et demander à rejoindre vos classes"},
	"profile_bio":         {"en": "Bio", "fr": "Biographie"},
	"profile_bio_ph":      {"en": "Tell students about yourself, your teaching style, experience...", "fr": "Parlez de vous, votre style d'enseignement, votre expérience..."},
	"profile_subjects":    {"en": "Subjects You Teach", "fr": "Matières que vous enseignez"},
	"profile_levels":      {"en": "Levels You Teach", "fr": "Niveaux que vous enseignez"},
	"profile_save":        {"en": "Save Profile", "fr": "Enregistrer le profil"},
	"profile_saved":       {"en": "Profile saved successfully!", "fr": "Profil enregistré avec succès !"},
	"profile_preview":     {"en": "Preview your public profile", "fr": "Aperçu de votre profil public"},

	// ─────────────────────────────────────────────────
	// JOIN REQUESTS (teacher side)
	// ─────────────────────────────────────────────────
	"requests_heading":        {"en": "Join Requests", "fr": "Demandes d'inscription"},
	"requests_subheading":     {"en": "Students who want to join your classrooms", "fr": "Élèves souhaitant rejoindre vos classes"},
	"requests_select_class":   {"en": "Select classroom...", "fr": "Sélectionnez une classe..."},
	"requests_approve":        {"en": "Approve", "fr": "Accepter"},
	"requests_reject":         {"en": "Reject", "fr": "Refuser"},
	"requests_reject_confirm": {"en": "Reject this request?", "fr": "Refuser cette demande ?"},
	"requests_approved":       {"en": "Approved", "fr": "Acceptée"},
	"requests_rejected":       {"en": "Rejected", "fr": "Refusée"},
	"requests_empty":          {"en": "No join requests yet", "fr": "Aucune demande pour le moment"},
	"requests_empty_sub":      {"en": "When students find you on Explore, their requests will appear here", "fr": "Lorsque des élèves vous trouveront sur Explorer, leurs demandes apparaîtront ici"},
	"footer_cgu":              {"en": "Terms & Privacy", "fr": "CGU & Confidentialité"},
}

// T returns a translated string for the given key and language.
// Falls back to English, then to the key itself if not found.
func T(lang, key string) string {
	if tr, ok := Translations[key]; ok {
		if val, ok := tr[lang]; ok {
			return val
		}
		if val, ok := tr["en"]; ok {
			return val
		}
	}
	return key
}
