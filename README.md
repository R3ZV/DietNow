# DietNow - Asistent Inteligent de Nutritie si Monitorizare Calorica

Aplicație modernă pentru gestionarea stilului de viață sănătos, monitorizarea caloriilor și a greutății (bulk, cut, menținere) și recomandarea inteligentă de rețete asistată de Inteligență Artificială (AI Chat).

---

## 📌 Documentație Proiect (Săptămânile 3-4)

Documentația completă cerută pentru livrabilele din Săptămânile 3-4 este structurată în directorul `docs/`:

- 📄 [Product Backlog Complet (docs/PRODUCT_BACKLOG.md)](docs/PRODUCT_BACKLOG.md)
  - **Matricea Centralizată (Master Table)**: 22 User Stories organizate pe 6 Epics.
  - **Detaliere User Stories**: Descrieri complete, criterii de acceptanță (Given-When-Then), sarcini tehnice.
  - **MoSCoW & Story Points**: Estimare bazată pe Planning Poker (Fibonacci) și raport Valoare Business vs Complexitate.
  - **Planificare pe Sprint-uri**: Împărțire strategică pe 4 Sprint-uri și cadență de viteză.
  - **DoD & DoR**: Definiția calității pentru terminarea fiecărei funcționalități.
- 📄 [Specificație de Cerințe și Arhitectură (docs/Specificatie_Cerinte_si_Backlog.md)](docs/Specificatie_Cerinte_si_Backlog.md)
  - **Constituirea Echipei**: Roluri Scrum (PO, SM, Devs, QA) și convenții de lucru.
  - **Alegerea Temei**: Descrierea proiectului, justificarea utilității și valoarea adăugată.
  - **Requirements Document (SRS)**: Cerințe funcționale (RF) și non-funcționale (RNF), user personas, arhitectură de sistem.
  - **Product Goal**: Obiectivul de produs clar definit pentru MVP.

---

## 🎯 Funcționalități Cheie

1. **Onboarding & Profilare Biometrică:**
   - Chestionar la înregistrare: sex, vârstă, înălțime, greutate actuală, nivel de activitate, obiectiv (bulk / cut / menținere), greutate țintă și orizont de timp.
   - Calcul automat al BMR (Rata Metabolică Bazală), TDEE (Consumul Total Zilnic de Energie) și repartiția optimă a macronutrienților (Proteine / Carbohidrați / Grăsimi).

2. **Jurnal Nutrițional & Monitorizare Greutate:**
   - Log zilnic structurat pe 4 mese: *Breakfast* (Mic dejun), *Lunch* (Prânz), *Dinner* (Cină), *Snacks* (Gustări).
   - Urmărirea caloriilor și a macronutrienților consumați vs. ținta zilnică.
   - Monitorizarea greutății corporale cu grafice de evoluție în timp.
   - Selectarea unui plan nutrițional prestabilit (High-Protein, Echilibrat, Mediteranean, etc.).

3. **Asistent Nutrițional AI (Conversational Chat):**
   - Chatbot AI dedicat căruia utilizatorul îi poate specifica preferințele culinare, alergiile sau ingredientele disponibile.
   - Generare de rețete personalizate, adaptate exact la macro-urile rămase pentru ziua respectivă.
   - Sugestii inteligente de gustări și substituiri sănătoase de alimente.
