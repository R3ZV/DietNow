# DietNow – Product Backlog & Project Specification

---

## 💡 1. Project Concept & Idea

**DietNow** is an intelligent nutrition and diet-tracking application designed to remove friction from meal planning and fitness goals. The app focuses on two core pillars: seamless macro/weight tracking and an interactive conversational AI that suggests recipes and meal ideas tailored to what you have and what you like.

### Initial Onboarding Questionnaire
When a user joins the app, they complete a guided profiling flow:
- **Biometrics:** Gender (female / male), age, height (cm), current weight (kg).
- **Activity Level:** Sedentary, lightly active, moderately active, very active.
- **Fitness Goal:** Bulking (muscle/weight gain), cutting (fat loss), or maintenance.
- **Target Metrics:** Desired target weight (kg) and the timeframe/deadline to achieve the goal (e.g., 10 weeks, 3 months).
- **Automated Calculation:** The system automatically calculates BMR (Basal Metabolic Rate), TDEE (Total Daily Energy Expenditure), the daily caloric surplus/deficit needed to reach the goal within the chosen timeframe, and the optimal macronutrient breakdown (protein, carbs, fats).

### The Two Core Pillars

#### Pillar 1: Weight & Calorie Tracking + Nutrition Plans
- **4-Meal Daily Tracker:** Daily meal logging structured into 4 distinct categories:
  1. **Breakfast**
  2. **Lunch**
  3. **Dinner**
  4. **Snacks**
- **Calorie & Macro Progress:** Real-time dashboard showing calories consumed vs. calories remaining, along with progress bars for protein, carbohydrate, and fat targets.
- **Weight Tracking:** Periodic weight log tracking progress against the target weight over the specified timeframe.
- **Nutrition Plans:** Ability to browse and select a structured diet plan (e.g., *High-Protein Bulk*, *Balanced Diet*, *Clean Vegetarian*, *Keto / Low-Carb*).

#### Pillar 2: Conversational AI Nutritionist (AI Chat)
- **Natural Language Assistant:** A dedicated in-app chat where users can talk naturally about food, nutrition, and cravings.
- **Custom Recipe Generator:** The user inputs what ingredients they have on hand (e.g., in their fridge/pantry) and what meal they need (breakfast, lunch, dinner, snack), and the AI produces a ready-to-cook recipe matching their exact remaining macros.
- **Dietary Preferences & Intolerances:** Users can specify allergies and preferences (lactose-free, gluten-free, vegetarian, high-protein, sweet vs. savory cravings), which the AI strictly respects in every recommendation.
- **Save to Favorites:** Instant 1-click option to save any generated recipe into the user's personal recipe book.

---

## 🎯 2. Product Goal

> **Product Goal:**  
> *"Deliver a functional MVP of DietNow that enables users to complete an onboarding questionnaire, calculate their personalized calorie and macro targets for bulking or cutting, track their daily meals across Breakfast, Lunch, Dinner, and Snacks, monitor weight progress, and receive instant, personalized recipe recommendations through an interactive AI chat assistant."*

---

## 📊 3. Master Product Backlog Matrix

| Story ID | Epic | User Story Title | MoSCoW Priority | Story Points | Target Milestone | Dependencies |
| :---: | :--- | :--- | :---: | :---: | :---: | :--- |
| **US-01** | Account | User Registration & Secure Authentication | **MUST** | **5 SP** | MVP | - |
| **US-02** | Onboarding | Biometric Intake (Gender, Age, Height, Weight, Activity) | **MUST** | **5 SP** | MVP | US-01 |
| **US-03** | Onboarding | Goal & Timeframe Definition (Target Weight & Weeks to Bulk) | **MUST** | **3 SP** | MVP | US-02 |
| **US-04** | Calculation | Automated BMR, TDEE & Daily Calorie Target Engine | **MUST** | **3 SP** | MVP | US-02, US-03 |
| **US-05** | Calculation | Macronutrient Split Breakdown (Protein, Carbs, Fats) | **MUST** | **3 SP** | MVP | US-04 |
| **US-06** | Tracker | 4-Meal Daily Log (Breakfast, Lunch, Dinner, Snacks) | **MUST** | **5 SP** | MVP | US-01, US-05 |
| **US-07** | Tracker | Food Search & Verified Nutrition Database | **MUST** | **5 SP** | MVP | US-06 |
| **US-08** | Tracker | Portion Logging (Grams/Servings) & Meal Totals | **MUST** | **5 SP** | MVP | US-07 |
| **US-09** | Tracker | Daily Dashboard (Remaining Calories & Macro Bars) | **MUST** | **5 SP** | MVP | US-06, US-08 |
| **US-10** | Weight | Periodic Weight Log Entry | **MUST** | **3 SP** | MVP | US-01 |
| **US-11** | Weight | Visual Weight Progress Chart vs. Bulk Target | **SHOULD** | **5 SP** | MVP | US-10, US-03 |
| **US-12** | Plans | Nutrition Plan Selection (High-Protein Bulk, Balanced, etc.) | **SHOULD** | **3 SP** | MVP | US-05 |
| **US-13** | AI Chat | Conversational AI Chat Interface & Integration | **MUST** | **5 SP** | MVP | US-01 |
| **US-14** | AI Chat | Context-Aware AI (Aware of User Bulk Goal & Remaining Calories) | **MUST** | **5 SP** | MVP | US-13, US-04, US-09 |
| **US-15** | AI Chat | Recipe Recommendation Based on Available Fridge Ingredients | **MUST** | **8 SP** | MVP | US-14 |
| **US-16** | AI Chat | Save Generated AI Recipes to Favorites Collection | **SHOULD** | **3 SP** | MVP | US-15 |
| **US-17** | AI Chat | Dietary Preferences & Intolerance Filters (Lactose, Gluten, Vegan) | **SHOULD** | **3 SP** | Post-MVP | US-15 |
| **US-18** | Account | Profile Settings & Goal/Metric Recalibration | **SHOULD** | **3 SP** | Post-MVP | US-02, US-04 |
| **US-19** | Tracker | Custom Food / Homemade Meal Creation | **SHOULD** | **5 SP** | Post-MVP | US-07 |
| **US-20** | Tracker | Barcode Scanner for Fast Food Lookup | **COULD** | **8 SP** | Future | US-07 |
| **US-21** | Reports | Weekly Progress PDF Export | **COULD** | **5 SP** | Future | US-09, US-11 |

---

## 📝 4. Detailed Backlog Items & Acceptance Criteria

---

### EPIC 1: Onboarding & Biometric Profiling

#### US-01: User Registration & Secure Authentication
- **User Story:** As a new user, I want to create an account with email and password and securely log in, so that my personal metrics, meal history, and favorite recipes remain private and persistent.
- **Priority:** **MUST HAVE** | **Points:** **5 SP**
- **Acceptance Criteria:**
  - **Scenario 1 (Successful Registration):**
    - *Given:* A new visitor provides a valid email address and a password of at least 8 characters.
    - *When:* They submit the registration form.
    - *Then:* An encrypted user record is created, a secure session/JWT token is issued, and the user is redirected to the onboarding wizard.
  - **Scenario 2 (Duplicate Account):**
    - *Given:* A user attempts to sign up with an existing email address.
    - *When:* They submit the form.
    - *Then:* The system rejects the request and displays: *"An account with this email already exists."*

---

#### US-02: Biometric Onboarding Questionnaire
- **User Story:** As a newly registered user, I want to complete an intuitive onboarding survey entering my gender, age, height, current weight, and activity level, so that the app understands my baseline physical condition.
- **Priority:** **MUST HAVE** | **Points:** **5 SP**
- **Acceptance Criteria:**
  - **Scenario 1 (Valid Input Submission):**
    - *Given:* The user is on the onboarding screen.
    - *When:* They select gender (Male/Female), enter height (e.g., 180 cm), current weight (e.g., 72.0 kg), age (e.g., 22), and select activity level (e.g., Moderately Active).
    - *Then:* The values are validated as within realistic biological ranges and stored in the user profile table.
  - **Scenario 2 (Invalid/Missing Values):**
    - *Given:* The user leaves the weight input empty or enters an impossible value (e.g., -5 or 600 kg).
    - *When:* They attempt to proceed.
    - *Then:* The offending field is highlighted with a descriptive error message and navigation is prevented.

---

#### US-03: Goal Setting & Target Bulk Timeframe
- **User Story:** As a user aiming to change my physique, I want to declare my objective (bulk, cut, or maintain), my target weight, and the timeframe (in weeks or months), so that the application can calculate a safe and realistic rate of weight change.
- **Priority:** **MUST HAVE** | **Points:** **3 SP**
- **Acceptance Criteria:**
  - **Scenario 1 (Healthy Bulk Goal):**
    - *Given:* A user currently at 70 kg sets an objective to bulk to 76 kg over 12 weeks.
    - *When:* They submit the goal.
    - *Then:* The system calculates the weekly surplus (~0.5 kg/week) and confirms the program parameters.
  - **Scenario 2 (Unrealistic Pace Warning):**
    - *Given:* A user attempts to configure a 15 kg weight increase within 2 weeks.
    - *When:* They submit the target.
    - *Then:* The system displays a wellness warning recommending an evidence-based pace of 0.25 to 0.5 kg per week for clean lean mass accrual.

---

### EPIC 2: Caloric Calculation & Nutrition Plans

#### US-04: Automated BMR, TDEE & Daily Calorie Engine
- **User Story:** As a user who completed the questionnaire, I want the system to automatically compute my BMR and TDEE, applying the appropriate caloric surplus for my bulk, so that I have a scientifically calculated daily caloric target.
- **Priority:** **MUST HAVE** | **Points:** **3 SP**
- **Acceptance Criteria:**
  - **Scenario 1 (Mifflin-St Jeor Calculation):**
    - *Given:* A male user, 72 kg, 182 cm, 21 years old, moderately active ($1.55$ multiplier).
    - *When:* The engine processes the profile data.
    - *Then:* $BMR = (10 \times 72) + (6.25 \times 182) - (5 \times 21) + 5 = 1752.5\text{ kcal}$, $TDEE = 1752.5 \times 1.55 \approx 2716\text{ kcal}$. For a moderate bulk, a surplus of $+350\text{ kcal}$ sets the daily target to $\approx 3066\text{ kcal}$.

---

#### US-05: Macronutrient Split Distribution
- **User Story:** As a fitness-oriented user, I want my daily calories translated into specific gram targets for Protein, Carbohydrates, and Fats, so that I can optimize muscle recovery and daily energy.
- **Priority:** **MUST HAVE** | **Points:** **3 SP**
- **Acceptance Criteria:**
  - **Scenario 1 (Bulk Macro Allocation):**
    - *Given:* The user has a bulk target of 3050 kcal and weighs 72 kg.
    - *When:* The macro calculator executes.
    - *Then:* It allocates:
      - Protein: $2.0\text{ g/kg} \rightarrow 144\text{ g}$ ($576\text{ kcal}$);
      - Fats: $1.0\text{ g/kg} \rightarrow 72\text{ g}$ ($648\text{ kcal}$);
      - Carbohydrates: Remainder ($\approx 1826\text{ kcal} \div 4 \approx 456\text{ g}$).

---

#### US-12: Nutrition Plan Selection
- **User Story:** As a user seeking dietary structure, I want to pick from curated nutrition plans (e.g., High-Protein Bulk, Balanced Whole Foods, Vegetarian / Vegan, Low-Carb), so that my meal suggestions align with my dietary lifestyle.
- **Priority:** **SHOULD HAVE** | **Points:** **3 SP**
- **Acceptance Criteria:**
  - **Scenario 1 (Selecting a High-Protein Plan):**
    - *Given:* The user visits the "Nutrition Plans" screen.
    - *When:* They select "High-Protein Bulk".
    - *Then:* Their target macro balance prioritizes protein sources, and the AI chat assistant defaults to high-protein meal templates.

---

### EPIC 3: Daily Tracking (4 Meals & Weight Progress)

#### US-06: 4-Meal Daily Log (Breakfast, Lunch, Dinner, Snacks)
- **User Story:** As an active user, I want my daily food log partitioned into four separate meal sections (Breakfast, Lunch, Dinner, Snacks), so that I can organize and track my food intake throughout the day.
- **Priority:** **MUST HAVE** | **Points:** **5 SP**
- **Acceptance Criteria:**
  - **Scenario 1 (Displaying 4 Distinct Meal Sections):**
    - *Given:* The user navigates to the tracker screen for the current date.
    - *When:* The page loads.
    - *Then:* Four separate sections are rendered (*Breakfast*, *Lunch*, *Dinner*, *Snacks*), each displaying its individual subtotal of calories and an *"Add Food"* button.
  - **Scenario 2 (Date Navigation):**
    - *Given:* The user selects a past date from the calendar/navigation bar.
    - *When:* The date changes.
    - *Then:* The log reflects the meals and totals recorded for that specific day.

---

#### US-07: Food Database Search
- **User Story:** As a user adding food to a meal, I want to search an extensive nutritional database by keyword, so that I can quickly find ingredients and items with verified caloric and macro data.
- **Priority:** **MUST HAVE** | **Points:** **5 SP**
- **Acceptance Criteria:**
  - **Scenario 1 (Keyword Search Query):**
    - *Given:* The user clicks *"Add Food"* under "Lunch" and types "chicken breast".
    - *When:* At least 2 characters are entered.
    - *Then:* A list of matching food items is returned displaying name, calories, and protein/carb/fat content per 100g.

---

#### US-08: Portion Logging & Dynamic Totaling
- **User Story:** As a user, I want to enter the specific quantity (in grams or standard servings) for an item, so that the app calculates and adds the exact nutritional values to my meal and daily total.
- **Priority:** **MUST HAVE** | **Points:** **5 SP**
- **Acceptance Criteria:**
  - **Scenario 1 (Adding 180g of an item):**
    - *Given:* The user selects "Oatmeal" (375 kcal / 100g).
    - *When:* They input "80" grams and confirm.
    - *Then:* The Breakfast section adds the entry with $300\text{ kcal}$, updating the Breakfast subtotal and the daily consumed counter immediately.
  - **Scenario 2 (Deleting an Entry):**
    - *Given:* A previously logged food entry.
    - *When:* The user clicks the delete button.
    - *Then:* The item is removed and the meal and day totals decrease instantly.

---

#### US-09: Live Dashboard (Remaining Calories & Macro Progress)
- **User Story:** As a user, I want to view a visual summary on my home dashboard displaying calories consumed vs. calories remaining and progress bars for each macronutrient, so that I know exactly how much room I have left to eat.
- **Priority:** **MUST HAVE** | **Points:** **5 SP**
- **Acceptance Criteria:**
  - **Scenario 1 (Remaining Budget Display):**
    - *Given:* A daily allowance of 3000 kcal with 1900 kcal logged so far.
    - *When:* The dashboard is viewed.
    - *Then:* The main indicator highlights **"1100 kcal Remaining"**, and individual progress bars show current percentage towards protein, carb, and fat goals.

---

#### US-10: Periodic Weight Log Entry
- **User Story:** As a user monitoring my body composition, I want to log my weigh-in values (in kg, to one decimal place), so that I maintain a historical record of my weight trajectory.
- **Priority:** **MUST HAVE** | **Points:** **3 SP**
- **Acceptance Criteria:**
  - **Scenario 1 (Logging Weight):**
    - *Given:* The user opens the weight tracker dialog.
    - *When:* They enter "72.4" kg and save.
    - *Then:* The weight log is updated for the day and saved to their profile history.

---

#### US-11: Visual Weight Progress Chart vs. Bulk Target
- **User Story:** As a user tracking a bulk or cut, I want to see an interactive line chart illustrating my recorded weigh-ins alongside a target trendline, so that I can see whether I am on pace to hit my goal within my chosen timeframe.
- **Priority:** **SHOULD HAVE** | **Points:** **5 SP**
- **Acceptance Criteria:**
  - **Scenario 1 (Chart Visualization):**
    - *Given:* The user has multiple logged weigh-ins across several weeks.
    - *When:* They view the "Progress" tab.
    - *Then:* A line chart renders all weigh-in data points with a reference dashed line representing the target weight trajectory.

---

### EPIC 4: Conversational AI Nutritionist (AI Chat)

#### US-13: Conversational AI Chat Interface & Integration
- **User Story:** As a user, I want to access a conversational chat screen powered by an AI assistant, so that I can ask questions about meals, get nutritional tips, and brainstorm ideas in natural language.
- **Priority:** **MUST HAVE** | **Points:** **5 SP**
- **Acceptance Criteria:**
  - **Scenario 1 (Real-Time Conversational Flow):**
    - *Given:* The user is on the AI Chat screen.
    - *When:* They send a query (e.g., *"What is a quick high-protein snack I can eat before my workout?"*).
    - *Then:* The message is sent, a typing/thinking indicator appears, and the AI streams back a coherent, actionable answer within 3 seconds.

---

#### US-14: Context-Aware AI (Profile & Remaining Macro Injection)
- **User Story:** As a user, I want the AI assistant to automatically know my biometrics, that I am currently bulking, and how many calories/protein I have left today, so that I do not need to re-explain my stats every time I chat.
- **Priority:** **MUST HAVE** | **Points:** **5 SP**
- **Acceptance Criteria:**
  - **Scenario 1 (Automatic Context Injection):**
    - *Given:* The user has 650 kcal and 45g of protein remaining today and has an active bulk goal.
    - *When:* The user asks: *"What should I have for dinner tonight?"*
    - *Then:* The AI prompt receives this context invisibly and responds with: *"Since you have 650 kcal and 45g of protein left today for your bulk target, here is a dinner idea that hits those numbers..."*

---

#### US-15: Recipe Generator Based on Available Fridge Ingredients
- **User Story:** As a user with limited groceries at home, I want to tell the AI what ingredients I have in my fridge and which meal I am cooking for (Breakfast, Lunch, Dinner, Snack), so that it invents a tasty recipe with exact steps and estimated macros.
- **Priority:** **MUST HAVE** | **Points:** **8 SP**
- **Acceptance Criteria:**
  - **Scenario 1 (Structured Recipe Output):**
    - *Given:* The user types: *"I have eggs, canned black beans, onions, and salsa. Give me a lunch recipe for bulk."*
    - *When:* The AI generates the output.
    - *Then:* The recipe is returned in a structured format containing:
      1. **Recipe Title**;
      2. **Preparation & Cook Time**;
      3. **Ingredient List with exact quantities**;
      4. **Numbered Step-by-Step Cooking Instructions**;
      5. **Nutrition Estimation Card** (Calories, Protein, Carbs, Fat).

---

#### US-16: 1-Click Save Recipe to Favorites
- **User Story:** As a user pleased with an AI-generated recipe, I want to click a "Save to Favorites" button directly inside the chat, so that I can store the recipe in my permanent collection without copy-pasting.
- **Priority:** **SHOULD HAVE** | **Points:** **3 SP**
- **Acceptance Criteria:**
  - **Scenario 1 (Recipe Persistence):**
    - *Given:* A recipe generated by the AI is displayed in chat.
    - *When:* The user clicks the bookmark / ⭐ "Save Recipe" button.
    - *Then:* The full recipe object is saved to the user's database record and immediately accessible in the "My Favorite Recipes" section.

---

#### US-17: Dietary Preferences & Intolerance Filters
- **User Story:** As a user with food intolerances or specific eating preferences (lactose intolerance, gluten-free, vegan, nut allergies), I want the AI to strictly enforce these exclusions, so that every meal and snack recommended is safe for me to eat.
- **Priority:** **SHOULD HAVE** | **Points:** **3 SP**
- **Acceptance Criteria:**
  - **Scenario 1 (Strict Allergen Compliance):**
    - *Given:* The user has flagged "Lactose Intolerant" in their preferences.
    - *When:* They ask the AI for a dessert or snack recommendation.
    - *Then:* The AI never proposes dairy products, recommending safe alternatives (e.g., coconut yogurt, plant-based milk).

---

### EPIC 5: Account Management & Enhancements

#### US-18: Profile Settings & Recalibration
- **User Story:** As a user whose weight, activity, or target changes over time, I want to update my profile settings, so that the application recalibrates my daily calorie and macro goals accordingly.
- **Priority:** **SHOULD HAVE** | **Points:** **3 SP**
- **Acceptance Criteria:**
  - **Scenario 1 (Target Reached / Recalibration):**
    - *Given:* A user finishes their bulk and switches their goal to "Maintenance".
    - *When:* They save the new objective in Settings.
    - *Then:* The surplus calories are removed and the daily calorie allowance is reset to match their current TDEE.

---

#### US-19: Custom Food Item Creation
- **User Story:** As a user consuming local or specialty foods not present in the default database, I want to manually create custom food items with their nutritional label data, so that I can accurately log them into my meals.
- **Priority:** **SHOULD HAVE** | **Points:** **5 SP**
- **Acceptance Criteria:**
  - **Scenario 1 (Custom Food Creation):**
    - *Given:* The user cannot find a specific snack in search.
    - *When:* They input the Name, Brand, Calories, and Macros per 100g in the "Create Custom Food" form.
    - *Then:* The item is saved to their personal library and selectable in any meal.

---

#### US-20: Barcode Scanner for Packaging
- **User Story:** As a mobile user, I want to scan product barcodes using my device camera, so that I can log packaged foods instantly without typing.
- **Priority:** **COULD HAVE** | **Points:** **8 SP**
- **Acceptance Criteria:**
  - **Scenario 1 (UPC / EAN Lookup):**
    - *Given:* A user scans a food barcode.
    - *When:* The barcode is recognized.
    - *Then:* The nutrition data is fetched automatically from an open database (e.g., OpenFoodFacts) and pre-filled in the meal log.

---

#### US-21: Weekly Nutrition & Weight Progress PDF Export
- **User Story:** As a dedicated user or trainee working with a coach, I want to export my weekly nutrition and weigh-in log as a clean PDF summary, so that I can review long-term consistency.
- **Priority:** **COULD HAVE** | **Points:** **5 SP**
- **Acceptance Criteria:**
  - **Scenario 1 (Export Generation):**
    - *Given:* A week with logged data.
    - *When:* The user clicks "Export Weekly Report".
    - *Then:* A formatted PDF is downloaded displaying daily average calories, macro distribution, and weight change.

---

## 📈 5. Backlog Summary & Estimation Totals

- **Total Backlog Items:** 21 User Stories
- **MVP Must-Have Scope:** **50 Story Points** (US-01 through US-10, US-13, US-14, US-15)
  - Covers complete onboarding, metabolic math, 4-meal tracking, daily dashboard, weight log, and interactive AI recipe generator.
- **Should-Have Scope:** **25 Story Points** (Weight chart, Nutrition plans, Favorite recipes, Allergy filters, Profile recalibration, Custom foods)
- **Could-Have Scope (Future Iterations):** **13 Story Points** (Barcode scanner, PDF reports)
