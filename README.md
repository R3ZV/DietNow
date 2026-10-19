# DietNow - Asistent Inteligent de Nutriție și Monitorizare Calorică

Aplicație modernă pentru gestionarea stilului de viață sănătos, monitorizarea caloriilor și a greutății (bulk, cut, menținere) și recomandarea inteligentă de rețete asistată de Inteligență Artificială (AI Chat).

---

## 📌 Documentație Proiect (Săptămânile 3-4)

Documentația completă cerută pentru livrabilele din Săptămânile 3-4 este structurată în directorul `docs/`:

- 📄 [Specificație de Cerințe și Product Backlog (docs/Specificatie_Cerinte_si_Backlog.md)](docs/Specificatie_Cerinte_si_Backlog.md)
  - **Constituirea Echipei**: Roluri Scrum (PO, SM, Devs, QA) și convenții de lucru.
  - **Alegerea Temei**: Descrierea proiectului, justificarea utilității și valoarea adăugată.
  - **Requirements Document (SRS)**: Cerințe funcționale (RF) și non-funcționale (RNF), user personas, arhitectură de sistem.
  - **Product Goal**: Obiectivul de produs clar definit pentru MVP.
  - **User Stories Inițiale**: User stories detaliate cu criterii de acceptanță (Given-When-Then).
  - **Product Backlog Inițial**: Tabel complet cu prioritizare MoSCoW și estimare în Story Points.
  - **Estimare și Prioritizare**: Metodologie Planning Poker (Fibonacci) și planificarea pe Sprint-uri.

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
