# DietNow – Product Backlog

---

## 💡 1. Project Overview & Scope

**DietNow** is a personalized nutrition application designed to help users reach their physique goals (specifically bulking/muscle gain, cutting, or maintenance) through accurate tracking and an AI-driven recipe assistant.

### Onboarding Flow (Questionnaire)
Before accessing the main app, the user answers a questionnaire to set up their profile:
- **Gender:** Female or Male
- **Biometrics:** Current weight (kg), Height (cm), Age
- **Activity Level:** Daily physical activity level
- **Goal & Timeline:** Target weight (kg) to reach for bulk/cut, and the desired timeframe (in weeks or months) to reach that number
- **Automated Calculation:** The app calculates the user's daily caloric requirement (including surplus for bulk) and macronutrient targets (protein, carbs, fats).

### Core Features to Implement

#### Feature 1: Weight & Calorie Tracking + Nutrition Plans
- **Daily 4-Meal Log:** Track calories and foods across four meals:
  1. **Breakfast**
  2. **Lunch**
  3. **Dinner**
  4. **Snacks**
- **Daily Progress Dashboard:** Real-time view of calories consumed vs. calories remaining, plus protein, carbohydrate, and fat progress.
- **Weight Tracking:** Log body weight over time to monitor progress against the target bulk timeline.
- **Nutrition Plans:** Select a structured nutrition plan (e.g., High-Protein Bulk, Balanced, Vegetarian, Low-Carb).

#### Feature 2: AI Recipe & Meal Assistant (AI Chat)
- **Conversational Chat:** In-app chat assistant dedicated to nutrition.
- **Preference-Based Recommendations:** Users communicate their food preferences, tastes, allergies, or ingredients they currently have on hand.
- **Targeted Meal Suggestions:** AI suggests specific recipes and foods to eat for breakfast, lunch, dinner, or snacks that match the user's remaining calories and macronutrient goals.
- **Save to Favorites:** Save generated recipes to a personal favorites list for quick access.

---

## 📋 2. Product Backlog Matrix

| Story ID | Feature / Module | User Story Title | Priority |
| :---: | :--- | :--- | :---: |
| **US-01** | Account | User Registration & Login | **High** |
| **US-02** | Onboarding | Biometric Intake (Gender, Height, Current Weight, Activity Level) | **High** |
| **US-03** | Onboarding | Bulk Goal & Timeframe Definition (Target Weight & Weeks to Goal) | **High** |
| **US-04** | Onboarding | Automated Calorie & Macronutrient Calculation Engine | **High** |
| **US-05** | Tracking | 4-Meal Daily Tracker (Breakfast, Lunch, Dinner, Snacks) | **High** |
| **US-06** | Tracking | Food Search & Portion Logging with Caloric Totals | **High** |
| **US-07** | Tracking | Daily Dashboard (Remaining Calories & Macro Progress) | **High** |
| **US-08** | Tracking | Weight Tracking (Periodic Weigh-In Log & Progress toward Bulk Target) | **High** |
| **US-09** | Tracking | Nutrition Plan Selection (High-Protein Bulk, Balanced, Vegetarian, etc.) | **Medium** |
| **US-10** | AI Chat | Conversational AI Chat Interface | **High** |
| **US-11** | AI Chat | User Dietary Preferences & Intolerances Input | **High** |
| **US-12** | AI Chat | AI Recipe & Food Recommendations by Meal Type (Breakfast, Lunch, Dinner, Snacks) | **High** |
| **US-13** | AI Chat | Recipe Suggestions from Available Fridge Ingredients | **High** |
| **US-14** | AI Chat | Save Recommended Recipes to Favorites | **Medium** |

---

## 📝 3. Detailed Backlog Items & Acceptance Criteria

---

### Module 1: Account & Onboarding Questionnaire

#### US-01: User Registration & Login
- **User Story:** As a user, I want to create an account and log in, so that my personal biometrics, daily food logs, and saved recipes are securely stored.
- **Priority:** **High**
- **Acceptance Criteria:**
  - **Scenario 1 (Account Creation):**
    - *Given:* A new user provides an email and password.
    - *When:* They submit the registration form.
    - *Then:* An account is created, the user is authenticated, and they are guided directly to the onboarding questionnaire.
  - **Scenario 2 (User Login):**
    - *Given:* A registered user enters their credentials.
    - *When:* Credentials are valid.
    - *Then:* The user is authenticated and taken to the main dashboard.

---

#### US-02: Biometric Onboarding Questionnaire
- **User Story:** As a new user, I want to answer initial questions about my gender (female or male), height, current weight (kg), age, and activity level, so that the app understands my baseline physical stats.
- **Priority:** **High**
- **Acceptance Criteria:**
  - **Scenario 1 (Valid Biometric Submission):**
    - *Given:* The user is on the onboarding questionnaire.
    - *When:* They select gender, enter height (cm), current weight (kg), and select their activity level (e.g., sedentary, moderate, active).
    - *Then:* The values are validated and stored in their profile.
  - **Scenario 2 (Validation of Required Fields):**
    - *Given:* Any required question is left unanswered or contains invalid numbers.
    - *When:* The user tries to proceed.
    - *Then:* The app displays an error message highlighting the missing or invalid field.

---

#### US-03: Goal Setting & Bulk Timeframe
- **User Story:** As a user, I want to specify what I want to achieve (e.g., bulk), my target weight (kg), and the timeframe (in weeks or months) to reach that weight, so that the app can plan my progress.
- **Priority:** **High**
- **Acceptance Criteria:**
  - **Scenario 1 (Setting Bulk Target):**
    - *Given:* The user is currently at 70 kg.
    - *When:* They choose "Bulk", enter a target weight of 75 kg, and choose a timeline of 10 weeks.
    - *Then:* The app saves the goal and calculates the target weight gain rate (~0.5 kg/week).

---

#### US-04: Automated Daily Calorie & Macronutrient Engine
- **User Story:** As a user who completed the questionnaire, I want the app to automatically calculate my daily calorie target (with the appropriate surplus for bulking) and my macronutrient split (proteins, carbs, fats), so that I know exactly what numbers to hit every day.
- **Priority:** **High**
- **Acceptance Criteria:**
  - **Scenario 1 (Calculation of Daily Targets):**
    - *Given:* The user's biometric stats and bulk timeframe.
    - *When:* The calculation runs.
    - *Then:* The system calculates BMR, TDEE, adds the required surplus for the bulk goal, and outputs daily calories, protein (g), carbohydrates (g), and fats (g).

---

### Module 2: Tracking (Calories, Weight & Nutrition Plans)

#### US-05: 4-Meal Daily Tracker (Breakfast, Lunch, Dinner, Snacks)
- **User Story:** As a user, I want my daily tracker organized into 4 distinct categories: Breakfast, Lunch, Dinner, and Snacks, so that I can log what I eat at each meal.
- **Priority:** **High**
- **Acceptance Criteria:**
  - **Scenario 1 (4 Meal Display):**
    - *Given:* The user is on the tracking page.
    - *When:* The tracker renders.
    - *Then:* Four separate sections are displayed: **Breakfast**, **Lunch**, **Dinner**, and **Snacks**, each showing its own calorie subtotal.
  - **Scenario 2 (Date Selector):**
    - *Given:* The user wants to view or log meals for a different day.
    - *When:* They change the date in the tracker.
    - *Then:* The food log for the selected date is displayed.

---

#### US-06: Food Search & Portion Logging
- **User Story:** As a user, I want to search for food items and enter the quantity (in grams or portions) eaten, so that the app adds the calories and macros to that specific meal.
- **Priority:** **High**
- **Acceptance Criteria:**
  - **Scenario 1 (Adding Food to a Meal):**
    - *Given:* The user taps "Add Food" under Breakfast.
    - *When:* They search for "Oats", enter "80g", and confirm.
    - *Then:* The item is added to Breakfast, and the calories and macros are added to the Breakfast subtotal and the daily total.
  - **Scenario 2 (Deleting or Editing a Food Item):**
    - *Given:* An item was logged by mistake.
    - *When:* The user removes or edits the item.
    - *Then:* The meal subtotal and daily total update immediately.

---

#### US-07: Daily Dashboard (Remaining Calories & Macro Progress)
- **User Story:** As a user, I want to see a live dashboard showing calories consumed versus calories remaining, along with progress bars for protein, carbs, and fats, so that I can see how close I am to my daily bulk goal.
- **Priority:** **High**
- **Acceptance Criteria:**
  - **Scenario 1 (Live Calorie & Macro Budget):**
    - *Given:* A daily goal of 3,000 kcal and 1,800 kcal logged so far.
    - *When:* The user looks at the dashboard.
    - *Then:* The dashboard clearly shows "1,200 kcal remaining" and visual progress bars for protein, carbohydrate, and fat intake.

---

#### US-08: Body Weight Tracking
- **User Story:** As a user on a bulk/cut plan, I want to log my weight (kg) regularly, so that I can track my progress over time toward my target bulk weight.
- **Priority:** **High**
- **Acceptance Criteria:**
  - **Scenario 1 (Logging Weight):**
    - *Given:* The user opens the weight tracking section.
    - *When:* They enter their current weight (e.g., 71.2 kg) and save.
    - *Then:* The entry is saved with the date and added to their weight history.
  - **Scenario 2 (Progress toward Target):**
    - *Given:* Logged weight entries over time.
    - *When:* The user views their weight progress.
    - *Then:* The app displays their starting weight, current weight, target bulk weight, and progress made.

---

#### US-09: Nutrition Plan Selection
- **User Story:** As a user, I want to select a nutrition plan (e.g., High-Protein Bulk, Balanced, Vegetarian, Low-Carb), so that my diet follows a clear nutritional framework.
- **Priority:** **Medium**
- **Acceptance Criteria:**
  - **Scenario 1 (Choosing a Plan):**
    - *Given:* The user accesses the "Nutrition Plans" screen.
    - *When:* They select a plan (e.g., "High-Protein Bulk").
    - *Then:* The plan is assigned to their profile, adjusting default macro recommendations accordingly.

---

### Module 3: AI Recipe & Meal Assistant (AI Chat)

#### US-10: Conversational AI Chat Interface
- **User Story:** As a user, I want to interact with a conversational AI chat assistant in the app, so that I can ask for food suggestions and nutrition advice in natural language.
- **Priority:** **High**
- **Acceptance Criteria:**
  - **Scenario 1 (Sending and Receiving Messages):**
    - *Given:* The user opens the AI Chat.
    - *When:* They type a message and send it.
    - *Then:* The AI responds with helpful, clear advice and meal ideas within a few seconds.

---

#### US-11: Dietary Preferences & Intolerances Input
- **User Story:** As a user, I want to tell the AI chat about my food preferences, dislikes, or dietary restrictions (e.g., lactose-free, gluten-free, vegetarian, high-protein), so that it only recommends foods I actually want or can eat.
- **Priority:** **High**
- **Acceptance Criteria:**
  - **Scenario 1 (Enforcing Preferences & Allergies):**
    - *Given:* The user specifies that they are lactose intolerant or prefer vegetarian meals.
    - *When:* The AI makes meal or recipe recommendations.
    - *Then:* The AI strictly excludes dairy or meat, suggesting appropriate alternatives.

---

#### US-12: AI Recipe & Meal Recommendations by Meal Type
- **User Story:** As a user, I want the AI to recommend specific recipes and things to eat for breakfast, lunch, dinner, or snacks, so that I never run out of meal ideas for my bulk.
- **Priority:** **High**
- **Acceptance Criteria:**
  - **Scenario 1 (Meal-Specific Recommendation):**
    - *Given:* The user asks: *"Give me a high-protein breakfast recipe for my bulk."*
    - *When:* The AI generates the suggestion.
    - *Then:* The AI returns a complete recipe with ingredients, preparation steps, and estimated calories/protein.
  - **Scenario 2 (Snack Recommendation):**
    - *Given:* The user asks: *"What is a quick snack I can eat between meals that has around 300 kcal?"*
    - *When:* The AI responds.
    - *Then:* The AI provides quick snack options matching the requested calorie target.

---

#### US-13: Recipe Suggestions from Available Fridge Ingredients
- **User Story:** As a user with groceries at home, I want to list the ingredients I have in my fridge/pantry, so that the AI creates a custom recipe using what I already have.
- **Priority:** **High**
- **Acceptance Criteria:**
  - **Scenario 1 (Fridge Ingredient Cooking):**
    - *Given:* The user says: *"I have eggs, cottage cheese, spinach, and toast. What can I make for breakfast?"*
    - *When:* The AI processes the request.
    - *Then:* It provides a recipe utilizing those ingredients, complete with portion recommendations and cooking instructions.

---

#### US-14: Save Recommended Recipes to Favorites
- **User Story:** As a user who received a recipe recommendation from the AI chat, I want to save it to my Favorites with one click, so that I can easily find and cook it again later.
- **Priority:** **Medium**
- **Acceptance Criteria:**
  - **Scenario 1 (Saving a Recipe):**
    - *Given:* A recipe generated by the AI is displayed in the chat.
    - *When:* The user clicks "Save to Favorites".
    - *Then:* The recipe is added to the user's saved recipes list and is accessible from their profile or tracking view.
