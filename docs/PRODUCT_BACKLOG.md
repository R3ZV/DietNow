# DietNow – Product Backlog & Story Refinement (INVEST)

## Master Product Backlog Matrix

| Story ID | Epic | User Story Title | Priority |
| :---: | :--- | :--- | :---: |
| **US-01** | `EPIC 1` Onboarding | Biometric Data Intake (Gender, Height, Weight, Activity) | **High** |
| **US-02** | `EPIC 1` Onboarding | Bulk Goal & Target Weight Definition | **High** |
| **US-03** | `EPIC 1` Onboarding | Timeframe & Weekly Weight Target Configuration | **High** |
| **US-04** | `EPIC 2` Calculations | Automated BMR & TDEE Metabolic Calculation | **High** |
| **US-05** | `EPIC 2` Calculations | Caloric Surplus Adjustment for Bulking Goal | **High** |
| **US-06** | `EPIC 2` Calculations | Macronutrient Split Breakdown (Protein, Carbs, Fats) | **High** |
| **US-07** | `EPIC 2` Calculations | Nutrition Plan Selection (High-Protein, Balanced, Veg, Low-Carb) | **Medium** |
| **US-08** | `EPIC 3` Tracking | 4-Meal Daily Partitioning (Breakfast, Lunch, Dinner, Snacks) | **High** |
| **US-09** | `EPIC 3` Tracking | Food Database Search & Nutritional Information Lookup | **High** |
| **US-10** | `EPIC 3` Tracking | Food Portion Logging (Grams) with Instant Meal Subtotals | **High** |
| **US-11** | `EPIC 3` Tracking | Live Daily Dashboard (Consumed vs. Remaining Calories & Macros) | **High** |
| **US-12** | `EPIC 3` Tracking | Periodic Body Weight Logging | **High** |
| **US-13** | `EPIC 3` Tracking | Visual Weight Trajectory & Milestone Progress Chart | **Medium** |
| **US-14** | `EPIC 4` AI Chat | Conversational In-App AI Nutrition Chat Interface | **High** |
| **US-15** | `EPIC 4` AI Chat | User Dietary Preferences & Intolerances Input | **High** |
| **US-16** | `EPIC 4` AI Chat | Recipe Recommendations by Meal Type (Breakfast, Lunch, Dinner, Snacks) | **High** |
| **US-17** | `EPIC 4` AI Chat | Recipe Generation from Available Fridge / Pantry Ingredients | **High** |
| **US-18** | `EPIC 4` AI Chat | 1-Click Save Recommended Recipe to Favorites | **Medium** |
| **US-19** | `EPIC 5` Account | User Account Registration & Secure Authentication | **High** |
| **US-20** | `EPIC 5` Account | Profile Settings & Goal/Metric Recalibration | **Medium** |
| **US-21** | `EPIC 6` Longevity | Eating Habit Analysis & Missing Longevity Nutrient Detection | **High** |
| **US-22** | `EPIC 6` Longevity | Science-Backed Longevity Recipe Recommendations (Bryan Johnson Research) | **High** |
| **US-23** | `EPIC 6` Longevity | Interactive Highlighted Longevity Tags & Scientific Insight Modals | **Medium** |

## Detailed Refined User Stories & Acceptance Criteria

### EPIC 1: Onboarding & Biometric Profiling

#### US-01: Biometric Data Intake (Gender, Height, Weight, Activity Level)
- **User Story:** As a new user, I want to input my gender (female or male), age,
    height (cm), current weight (kg), and physical activity level, so that the
    app establishes my baseline physiological stats.
- **Priority:** **High**
- **Acceptance Criteria:**
  - **Scenario 1 (Valid Input Submission):**
    - *Given:* The user is on the biometric onboarding screen.
    - *When:* They select gender, input height (e.g., 180 cm), current weight (e.g., 72.5 kg), age (e.g., 22), and select their activity level (e.g., Moderately Active).
    - *Then:* The values are validated as realistic positive numbers and saved to the user profile.
  - **Scenario 2 (Invalid Input Boundary Handling):**
    - *Given:* The user inputs an unrealistic weight (e.g., -10 kg or 600 kg) or leaves a required field blank.
    - *When:* They attempt to proceed.
    - *Then:* The system displays a validation error highlighting the invalid field and prevents progression.

#### US-02: Bulk Goal & Target Weight Definition
- **User Story:** As a user aiming to change my physique, I want to specify that my goal is bulking (weight gain) and set my target weight in kilograms, so that the system knows what target weight I need to reach.
- **Priority:** **High**
- **Acceptance Criteria:**
  - **Scenario 1 (Setting Target Bulk Weight):**
    - *Given:* A user with a current weight of 70 kg.
    - *When:* They select "Bulk" and enter a target weight of 75 kg.
    - *Then:* The system confirms the goal and calculates the net weight differential (+5 kg).
  - **Scenario 2 (Conflicting Goal Validation):**
    - *Given:* A user selects "Bulk" but enters a target weight lower than their current weight.
    - *When:* They submit the goal.
    - *Then:* The system alerts the user: *"For a bulk goal, your target weight must be greater than your current weight."*

#### US-03: Timeframe & Weekly Weight Target Configuration
- **User Story:** As a user pursuing a bulk goal, I want to specify the timeframe (number of weeks or months) to reach my target weight, so that the app calculates a healthy and sustainable weekly rate of weight gain.
- **Priority:** **High**
- **Acceptance Criteria:**
  - **Scenario 1 (Healthy Timeframe Acceptance):**
    - *Given:* A target gain of 5 kg.
    - *When:* The user inputs a timeframe of 10 weeks (~0.5 kg/week).
    - *Then:* The timeframe is confirmed, and the system records the weekly target rate.
  - **Scenario 2 (Unrealistic Pace Warning):**
    - *Given:* A user attempts to configure a 10 kg gain within 2 weeks.
    - *When:* They submit the form.
    - *Then:* The app displays an advisory warning suggesting a realistic gain pace (0.25 to 0.5 kg/week) to prioritize lean muscle over excessive fat.

### EPIC 2: Caloric Calculation & Nutrition Plans

#### US-04: Automated BMR & TDEE Metabolic Calculation
- **User Story:** As a user who completed the biometric intake, I want the system to automatically compute my Basal Metabolic Rate (BMR) and Total Daily Energy Expenditure (TDEE), so that I understand my baseline daily energy expenditure.
- **Priority:** **High**
- **Acceptance Criteria:**
  - **Scenario 1 (Mifflin-St Jeor Calculation):**
    - *Given:* A male user, 70 kg, 180 cm, 22 years old, moderately active ($1.55$ multiplier).
    - *When:* The calculation engine runs.
    - *Then:* $BMR$ is calculated as $1720\text{ kcal}$ and $TDEE$ as $\approx 2666\text{ kcal}$.

#### US-05: Caloric Surplus Adjustment for Bulking Goal
- **User Story:** As a bulking user, I want the system to calculate the necessary daily caloric surplus added to my TDEE, so that I have an exact total daily calorie target calibrated to my bulk timeframe.
- **Priority:** **High**
- **Acceptance Criteria:**
  - **Scenario 1 (Daily Surplus Calculation):**
    - *Given:* A TDEE of 2666 kcal and a target gain rate of 0.5 kg/week.
    - *When:* The surplus engine runs.
    - *Then:* The system adds approximately $+400\text{ kcal/day}$, establishing the daily intake goal at $3066\text{ kcal/day}$.

#### US-06: Macronutrient Split Breakdown (Protein, Carbs, Fats)
- **User Story:** As a fitness-oriented user, I want my total daily calories broken down into specific gram targets for Protein, Carbohydrates, and Fats, so that I know how much of each macronutrient to consume.
- **Priority:** **High**
- **Acceptance Criteria:**
  - **Scenario 1 (High-Protein Split for Bulk):**
    - *Given:* A user weighing 70 kg with a 3000 kcal daily goal.
    - *When:* The macronutrient breakdown is generated.
    - *Then:* The system allocates $2.0\text{ g/kg}$ of Protein ($140\text{ g}$), $1.0\text{ g/kg}$ of Fats ($70\text{ g}$), and the remaining calories to Carbohydrates ($\approx 450\text{ g}$).

#### US-07: Nutrition Plan Selection
- **User Story:** As a user seeking dietary structure, I want to choose a nutrition plan (e.g., High-Protein Bulk, Balanced, Vegetarian, Low-Carb), so that my meal suggestions align with my dietary preferences.
- **Priority:** **Medium**
- **Acceptance Criteria:**
  - **Scenario 1 (Plan Selection):**
    - *Given:* The user is on the "Nutrition Plans" screen.
    - *When:* They choose "Vegetarian Bulk".
    - *Then:* The profile updates to prioritize plant-based protein sources in recommendations.

### EPIC 3: Daily Tracking (4 Meals & Weight Progress)

#### US-08: 4-Meal Daily Partitioning (Breakfast, Lunch, Dinner, Snacks)
- **User Story:** As a user tracking my food, I want my daily log partitioned into 4 distinct meals: Breakfast, Lunch, Dinner, and Snacks, so that I can organize and view what I eat throughout the day.
- **Priority:** **High**
- **Acceptance Criteria:**
  - **Scenario 1 (Displaying 4 Meal Cards):**
    - *Given:* The user opens the tracker for the current date.
    - *When:* The page loads.
    - *Then:* Four separate sections are displayed (**Breakfast**, **Lunch**, **Dinner**, **Snacks**), each showing a subtotal of calories and an "Add Food" button.
  - **Scenario 2 (Date Navigation):**
    - *Given:* The user switches to a previous date using the date selector.
    - *When:* The date changes.
    - *Then:* The log updates to show the meals recorded for that specific date.

#### US-09: Food Database Search & Nutritional Information Lookup
- **User Story:** As a user logging food, I want to search for food items by keyword, so that I can view verified calorie and macronutrient information per 100g.
- **Priority:** **High**
- **Acceptance Criteria:**
  - **Scenario 1 (Keyword Search):**
    - *Given:* The user taps "Add Food" under Lunch and types "Chicken breast".
    - *When:* The query is executed.
    - *Then:* A list of matching items appears displaying calories, protein, carbs, and fats per 100g.

#### US-10: Food Portion Logging with Instant Meal Subtotals
- **User Story:** As a user, I want to input the portion size (in grams or servings) for an item, so that the app automatically calculates the exact nutrients and adds them to the meal.
- **Priority:** **High**
- **Acceptance Criteria:**
  - **Scenario 1 (Adding 80g portion):**
    - *Given:* The user selects "Oatmeal" (375 kcal / 100g).
    - *When:* They input "80g" and confirm.
    - *Then:* The Breakfast section adds the entry with $300\text{ kcal}$, updating the Breakfast subtotal and the daily total.
  - **Scenario 2 (Removing a Logged Item):**
    - *Given:* An entry was logged by mistake.
    - *When:* The user clicks delete.
    - *Then:* The entry is removed and totals update immediately.

#### US-11: Live Daily Dashboard (Consumed vs. Remaining Calories & Macros)
- **User Story:** As a user, I want to view a dashboard showing consumed vs. remaining calories and progress bars for protein, carbs, and fats, so that I know what I have left to eat for the day.
- **Priority:** **High**
- **Acceptance Criteria:**
  - **Scenario 1 (Live Remaining Budget):**
    - *Given:* A daily goal of 3000 kcal and 1800 kcal logged so far.
    - *When:* The dashboard is viewed.
    - *Then:* It clearly displays **"1200 kcal Remaining"** and visual progress bars for protein, carbohydrate, and fat intake.

#### US-12: Periodic Body Weight Logging
- **User Story:** As a user tracking my bulk progress, I want to log my weigh-in values (in kg, to one decimal place), so that I maintain a historical record of my body weight.
- **Priority:** **High**
- **Acceptance Criteria:**
  - **Scenario 1 (Logging Weight Entry):**
    - *Given:* The user opens the weight logging modal.
    - *When:* They enter "71.4 kg" and confirm.
    - *Then:* The weight entry is saved with today's date in their weight history.

#### US-13: Visual Weight Trajectory & Milestone Progress Chart
- **User Story:** As a user monitoring my bulk timeline, I want to see a chart plotting my logged weights alongside my target weight goal, so that I can visually verify if my progress matches my planned timeframe.
- **Priority:** **Medium**
- **Acceptance Criteria:**
  - **Scenario 1 (Chart Rendering):**
    - *Given:* The user has multiple logged weigh-ins.
    - *When:* They open the progress screen.
    - *Then:* A line chart renders all weigh-ins with a reference line for the target bulk weight.

### EPIC 4: Conversational AI Nutritionist (AI Chat)

#### US-14: Conversational In-App AI Nutrition Chat Interface
- **User Story:** As a user, I want to interact with a conversational AI chat assistant in the app, so that I can ask questions and discuss meal ideas in natural language.
- **Priority:** **High**
- **Acceptance Criteria:**
  - **Scenario 1 (Interactive Chat Flow):**
    - *Given:* The user opens the AI Chat screen.
    - *When:* They send a message (e.g., *"What is a good high-protein snack for bulk?"*).
    - *Then:* The AI generates a contextual, helpful response within 3 seconds.

#### US-15: User Dietary Preferences & Intolerances Input
- **User Story:** As a user with specific food preferences or intolerances (e.g., lactose-free, gluten-free, vegan), I want to communicate these to the AI assistant, so that it never suggests foods that violate my dietary restrictions.
- **Priority:** **High**
- **Acceptance Criteria:**
  - **Scenario 1 (Enforcing Dietary Exclusions):**
    - *Given:* The user specifies that they are lactose intolerant.
    - *When:* They ask the AI for a snack or meal idea.
    - *Then:* The AI strictly avoids dairy, recommending certified lactose-free or plant-based alternatives.

#### US-16: Recipe Recommendations by Meal Type (Breakfast, Lunch, Dinner, Snacks)
- **User Story:** As a user looking for meals, I want the AI to generate complete recipes categorized by meal type (Breakfast, Lunch, Dinner, Snacks), so that I have varied, healthy options for any meal of the day.
- **Priority:** **High**
- **Acceptance Criteria:**
  - **Scenario 1 (Structured Recipe Card):**
    - *Given:* The user requests: *"Give me a high-protein dinner recipe for my bulk."*
    - *When:* The AI generates the output.
    - *Then:* The recipe includes: Recipe Name, Cook Time, Ingredient Quantities, Step-by-Step Instructions, and Estimated Calories/Macros.

#### US-17: Recipe Generation from Available Fridge / Pantry Ingredients
- **User Story:** As a user with groceries at home, I want to list the ingredients I have in my fridge or pantry, so that the AI invents a recipe tailored to what I already have on hand.
- **Priority:** **High**
- **Acceptance Criteria:**
  - **Scenario 1 (Fridge-to-Plate Recipe):**
    - *Given:* The user writes: *"I have eggs, cottage cheese, spinach, and whole-wheat toast. Give me a breakfast recipe."*
    - *When:* The AI processes the ingredients.
    - *Then:* It returns a recipe using those exact items, with cooking instructions and estimated nutritional breakdown.

#### US-18: 1-Click Save Recommended Recipe to Favorites
- **User Story:** As a user who received a recipe recommendation from the AI chat, I want to save it to my Favorites with one click, so that I can easily revisit and cook it later.
- **Priority:** **Medium**
- **Acceptance Criteria:**
  - **Scenario 1 (Saving a Recipe):**
    - *Given:* A recipe generated by the AI is displayed in chat.
    - *When:* The user clicks the "⭐ Save to Favorites" button.
    - *Then:* The recipe is saved to their personal collection and accessible under "My Favorite Recipes".

### EPIC 5: Account Management & Enhancements

#### US-19: User Account Registration & Secure Authentication
- **User Story:** As a new user, I want to register and log in securely, so that my personal biometrics, daily tracking history, and favorite recipes are safely stored.
- **Priority:** **High**
- **Acceptance Criteria:**
  - **Scenario 1 (Registration & Session Creation):**
    - *Given:* A new user provides an email and a password of at least 8 characters.
    - *When:* They submit the form.
    - *Then:* The account is created, encrypted, and an authenticated session is established.
  - **Scenario 2 (User Login):**
    - *Given:* A registered user enters their valid credentials.
    - *When:* They submit the login form.
    - *Then:* The user is authenticated and redirected to their dashboard.

#### US-20: Profile Settings & Goal/Metric Recalibration
- **User Story:** As a user whose weight, activity, or target has changed, I want to update my profile settings, so that the app recalibrates my daily calorie and macro goals.
- **Priority:** **Medium**
- **Acceptance Criteria:**
  - **Scenario 1 (Target Reached & Goal Recalibration):**
    - *Given:* The user reaches their target bulk weight and switches their goal to "Maintenance".
    - *When:* They save the new objective in Settings.
    - *Then:* The caloric surplus is removed, and the daily intake goal adjusts to match their updated TDEE.

### EPIC 6: Longevity Recommendation Engine & Scientific Insights (Bryan Johnson Research)

#### US-21: Eating Habit Analysis & Missing Longevity Nutrient Detection
- **User Story:** As a health-conscious user, I want the system to analyze my logged meals over time against Bryan Johnson's longevity research standards, so that I can discover what key longevity foods and micronutrients are currently missing from my diet.
- **Priority:** **High**
- **INVEST Evaluation:**
  - *Valuable:* Delivers immediate personalized insight into chronic nutrient deficiencies that impact lifespan.
  - *Testable:* Verifiable by feeding sample meal logs and asserting that detected deficiencies (e.g., lack of polyphenols or healthy fats) are properly flagged.
- **Acceptance Criteria:**
  - **Scenario 1 (Nutritional Gap Detection):**
    - *Given:* A user whose recent 3-day meal log consists primarily of refined carbohydrates and chicken breast, lacking dietary polyphenols, cruciferous vegetables, and omega-3s.
    - *When:* The habit analysis module evaluates the food logs.
    - *Then:* The system generates an intake summary highlighting missing longevity compounds: *"Your diet is low in polyphenols, sulforaphane, and omega-3 fatty acids."*
  - **Scenario 2 (Balanced Longevity Intake):**
    - *Given:* A user regularly logging berries, broccoli sprouts, lentils, and extra virgin olive oil.
    - *When:* The habit analysis runs.
    - *Then:* The system confirms high compliance with core longevity biomarkers and displays a positive status.

#### US-22: Science-Backed Longevity Recipe Recommendations
- **User Story:** As a user with detected nutritional gaps, I want the recommendation system to suggest recipes engineered to supply my missing longevity nutrients based on Bryan Johnson's research, so that I can improve my cellular health and lifespan without falling for unverified influencer trends.
- **Priority:** **High**
- **INVEST Evaluation:**
  - *Independent:* Generates standalone recipe recommendations based on the gap analysis.
  - *Small:* Scoped specifically to querying and formatting longevity recipes that target identified gaps.
  - *Testable:* Verifiable by checking that suggested recipes include the specific missing longevity ingredients.
- **Acceptance Criteria:**
  - **Scenario 1 (Targeted Longevity Meal Suggestion):**
    - *Given:* The system identified that the user is missing polyphenols and healthy fats.
    - *When:* The user accesses the recommendation section.
    - *Then:* The system suggests recipes rich in extra virgin olive oil (high-polyphenol EVOO), dark cocoa/berries, or steamed cruciferous vegetables, specifying how each meal closes their identified nutrient gap.
  - **Scenario 2 (Integration with Daily Macro Target):**
    - *Given:* The user has 600 kcal remaining for their bulk target.
    - *When:* The longevity recipe is proposed.
    - *Then:* The suggested recipe fits within their remaining caloric budget while maximizing nutrient density.

#### US-23: Interactive Highlighted Longevity Tags & Scientific Insight Modals
- **User Story:** As an informed user, I want recommended recipes to display highlighted longevity tags (e.g., `#Polyphenols`, `#Sulforaphane`, `#CellularRepair`, `#EVOO-Oleocanthal`) that I can click to read scientific summaries from Bryan Johnson's research, so that I understand the biological evidence and healthspan benefits behind my food.
- **Priority:** **Medium**
- **INVEST Evaluation:**
  - *Negotiable:* The modal design and depth of research citations can be refined collaboratively.
  - *Testable:* Verifiable by triggering click events on tags and verifying the presence of scientific research explanations.
- **Acceptance Criteria:**
  - **Scenario 1 (Clicking a Highlighted Tag):**
    - *Given:* A recipe card is displayed with the highlighted tag `#EVOO-Oleocanthal`.
    - *When:* The user clicks the tag.
    - *Then:* A modal window opens displaying:
      1. Scientific summary of oleocanthal and its anti-inflammatory / autophagy-promoting properties;
      2. Citation/reference to Bryan Johnson's longevity protocol;
      3. Impact on biological markers (e.g., arterial elasticity, hs-CRP reduction).
  - **Scenario 2 (Tag Dismissal):**
    - *Given:* An open research insight modal.
    - *When:* The user clicks the close icon or outside the modal backdrop.
    - *Then:* The modal closes smoothly, returning the user to the recipe view.
