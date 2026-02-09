# Implementation Plan

## Objective
Build a single-file (`index.html`) kid-friendly math practice webpage that shows one addition/subtraction problem at a time, supports difficulty levels 1-5, and validates answers with immediate feedback.

## Constraints From Problem
- Entire app must be in `index.html` (HTML, CSS, JS inline).
- No external frameworks or assets.
- One problem shown at a time in format `A + B = ?` (or subtraction equivalent) with input + submit button.
- Difficulty controlled by slider (`1` to `5`).
- Correct answer: show `Correct`, then load next problem.
- Incorrect answer: show `incorrect`, clear input, let user retry same problem.
- Mix addition and subtraction problems.
- Large, highly legible, kid-friendly design.

## Build Steps
1. Create base HTML structure in `index.html`.
   - App container/card.
   - Title and short instructions.
   - Difficulty control: slider (`min=1`, `max=5`, `step=1`) with visible current value.
   - Problem row: large equation text, answer input, submit button.
   - Feedback/status text area (`Correct` / `incorrect`).

2. Add kid-friendly responsive CSS.
   - Bright, playful color palette and rounded UI elements.
   - Large typography for equation and controls.
   - High contrast for readability.
   - Clear focus styles for keyboard accessibility.
   - Mobile-safe layout with stack/flex behavior.

3. Implement app state in inline JS.
   - `currentDifficulty` (1-5).
   - `currentProblem` object: `{ a, b, op, answer }`.
   - Cached DOM refs for slider, labels, equation, input, button, feedback.

4. Implement difficulty mapping.
   - Level 1: small non-negative integers suitable for early grade school.
   - Level 2: larger two-digit range, mostly non-negative results.
   - Level 3: wider range, regular borrow/carry cases.
   - Level 4: include negatives and larger magnitudes.
   - Level 5: broad integer range with negatives to raise challenge toward high-school readiness.

5. Implement random problem generation.
   - Randomly choose operator (`+` or `-`) each round for a true mix.
   - Generate operands according to difficulty range.
   - For lower levels, bias toward non-negative subtraction results.
   - Compute and store exact numeric answer in `currentProblem.answer`.

6. Render one problem at a time.
   - Equation display as `A op B =`.
   - Keep input field as the `?` response area.
   - Clear input and focus whenever a new round starts.

7. Wire interactions.
   - Slider `input/change`: update label, set difficulty, generate new problem.
   - Form submit/button click:
     - Parse numeric input.
     - If correct: show `Correct`, style success, generate next problem.
     - If incorrect: show `incorrect`, style error, clear input, keep same problem.

8. Add polish and guardrails.
   - Prevent blank/non-numeric submissions from being treated as correct.
   - Ensure feedback text updates consistently each attempt.
   - Keep transitions simple and fast (no heavy animations).

## Verification Checklist
- `index.html` contains all HTML/CSS/JS with no external dependencies.
- Only one problem is visible at all times.
- Slider values 1-5 clearly change problem difficulty.
- Both addition and subtraction appear over multiple rounds.
- Correct answer advances to a new problem with `Correct` message.
- Incorrect answer shows `incorrect`, clears input, and does not change problem.
- Equation and controls are large and easy for kids to read/use on desktop and mobile.
