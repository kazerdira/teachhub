```markdown
# Design System Strategy: The Academic Atelier

## 1. Overview & Creative North Star
This design system is built upon the Creative North Star: **"The Academic Atelier."** 

Unlike traditional education platforms that feel like rigid spreadsheets or chaotic playgrounds, this system treats digital learning as a curated, high-end editorial experience. We are bridging the gap between the structured utility of Notion and the delightful, gamified rhythm of Duolingo. 

To break the "SaaS template" look, we employ **Intentional Asymmetry** and **Tonal Depth**. We favor generous, "breathable" white space over information density. The layout should feel like a premium physical journal—tactile, layered, and deeply intentional. We do not use boxes to contain content; we use space and light to invite focus.

---

## 2. Colors & Surface Logic

The palette is anchored in warm, organic neutrals to reduce cognitive load and eye strain, while using a sophisticated Emerald Green to signal growth and achievement.

### The "No-Line" Rule
**Strict Mandate:** Designers are prohibited from using 1px solid borders to define sections or containers. Visual boundaries must be achieved through:
1.  **Background Shifts:** Placing a `surface-container-low` card on a `surface` background.
2.  **Tonal Transitions:** Using subtle variations in the warm neutral scale to signify a change in context.
3.  **Negative Space:** Using the Spacing Scale to create "invisible" containment.

### Surface Hierarchy & Nesting
Treat the interface as a series of stacked, fine-paper sheets.
*   **Base Layer (`surface` / `#faf9f5`):** The canvas.
*   **In-Page Sections (`surface-container-low`):** Sub-grouping of content.
*   **Primary Cards (`surface-container-lowest` / `#ffffff`):** High-priority interactive elements.
*   **Hover/Active States (`surface-container-high`):** Deepening the "indent" into the page.

### The Glass & Gradient Rule
To move beyond a flat "Web 2.0" look, use **Glassmorphism** for floating elements (like persistent navigation bars or mobile overlays). Use the `surface-tint` at 40% opacity with a `24px` backdrop blur. 
*   **Signature Textures:** For primary CTAs, use a subtle linear gradient from `primary` (#006a44) to `primary-container` (#008558) at a 135-degree angle. This adds a "jewel-tone" depth that feels premium and tactile.

---

## 3. Typography: The Editorial Voice

We use a high-contrast typographic scale to create an authoritative yet approachable voice.

*   **Display & Headlines (Plus Jakarta Sans):** These are our "Editorial Hooks." Use `display-lg` for hero moments. The generous x-height and geometric curves of Plus Jakarta Sans provide the "Duolingo-esque" friendliness while maintaining a "Premium" weight.
*   **Body & Labels (Manrope):** Manrope is selected for its modern, highly legible structure. Use `body-lg` for lesson content to ensure maximum readability.
*   **The Hierarchy Rule:** Never use more than three levels of typography on a single screen. Focus on the contrast between a large `headline-md` and a clean `body-md` to guide the eye.

---

## 4. Elevation & Depth: Tonal Layering

We reject the standard "drop shadow" approach. Depth in this system is organic and environmental.

*   **The Layering Principle:** Instead of shadows, stack containers. A `surface-container-lowest` card sitting on a `surface-container` background creates a natural "lift" that feels integrated, not pasted on.
*   **Ambient Shadows:** If a floating element (like a modal) requires a shadow, use an **Extended Ambient Shadow**.
    *   *Spec:* `0px 20px 40px rgba(27, 28, 26, 0.06)` (A tinted shadow using the `on-surface` color).
*   **The "Ghost Border" Fallback:** If accessibility requirements demand a border (e.g., in high-contrast scenarios), use a **Ghost Border**.
    *   *Spec:* `outline-variant` (#bccabf) at **15% opacity**. It should be felt, not seen.
*   **Tactile Radius:** Use the `DEFAULT` (1rem / 16px) for most containers. Use `full` (pill) for interactive elements like buttons and status indicators to maintain the "approachable" brand personality.

---

## 5. Components

### Buttons
*   **Primary:** Pill-shaped (`full` radius). Gradient fill (Primary to Primary-Container). White text. Subtle hover lift (2px Y-offset).
*   **Secondary:** Pill-shaped. `surface-container-highest` background. No border.
*   **Tertiary:** Text-only with a heavy `label-md` weight. No container until hover.

### Input Fields
*   **Style:** Minimalist. No bottom line or full border. Use a `surface-container-low` background with a `md` (1.5rem) radius.
*   **Focus State:** A 2px `primary` ghost-border (20% opacity) and the label shifting to `primary` color.

### Cards & Lists (The Academic Feed)
*   **Constraint:** **Strictly forbid horizontal divider lines.** 
*   **Execution:** Separate list items using a 12px vertical gap. For curriculum lists, use a alternating background shift (`surface` to `surface-container-low`) or a simple vertical white-space "breathing room" of 24px.

### Specialized Education Components
*   **Progress Pill:** A `full` radius progress bar using `primary` for the fill and `primary-fixed-dim` for the track.
*   **Curator’s Note:** A callout box using `secondary-container` (#6063ee) at 10% opacity, with a `secondary` accent vertical bar on the left (4px width, `full` radius).

---

## 6. Do’s and Don’ts

### Do:
*   **Do** use asymmetrical layouts. A lesson title might be left-aligned while the primary action is offset to the right with extra margin.
*   **Do** lean into "Warmth." Ensure the `#faf9f5` background is the dominant color to keep the experience feeling "Human."
*   **Do** use `Video Purple` and `PDF Red` as small, high-chroma accents (icons or small badges) to help students scan file types quickly.

### Don’t:
*   **Don't** use pure black (#000000) for text. Always use `on-surface` (#1b1c1a) to maintain the soft editorial feel.
*   **Don't** use 90-degree sharp corners. Everything must have a minimum of a `sm` radius to stay within the "approachable" brand pillar.
*   **Don't** crowd the screen. If you feel like you need a divider line, you actually need more white space. Increase the padding by 1.5x instead.

---
**Director's Note:** This system is not a kit of parts; it is a philosophy of space. Build for the student’s focus and the teacher’s authority. If it looks like a generic dashboard, simplify the color palette and increase the margins.```