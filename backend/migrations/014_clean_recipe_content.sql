-- Migration 014: clean scraped artifacts from recipe instructions.
--
-- The seed migrations are clean, but recipes scraped/edited in production
-- contain three types of cosmetic noise:
--
--   1. Inline serving qualifiers like " (för 4 port)" — leftovers from
--      sites where the user picks a portion size before viewing the recipe.
--      In Måltiden the user selects servings during menu generation, so
--      these annotations are confusing dead context.
--
--   2. A dangling reference to non-existent recipes in the
--      "Ugnsstekt kycklingklubba" instructions.
--
--   3. Inconsistent degree symbols (˚C, U+02DA) — normalize to ° (U+00B0).
--
-- All operations are idempotent string REPLACEs and safe to run twice.

-- 1. Strip " (för N port)" qualifiers (1..12 portions covers everything in prod).
UPDATE recipes SET instructions = REPLACE(instructions, ' (för 1 port)', '');
UPDATE recipes SET instructions = REPLACE(instructions, ' (för 2 port)', '');
UPDATE recipes SET instructions = REPLACE(instructions, ' (för 3 port)', '');
UPDATE recipes SET instructions = REPLACE(instructions, ' (för 4 port)', '');
UPDATE recipes SET instructions = REPLACE(instructions, ' (för 5 port)', '');
UPDATE recipes SET instructions = REPLACE(instructions, ' (för 6 port)', '');
UPDATE recipes SET instructions = REPLACE(instructions, ' (för 7 port)', '');
UPDATE recipes SET instructions = REPLACE(instructions, ' (för 8 port)', '');
UPDATE recipes SET instructions = REPLACE(instructions, ' (för 9 port)', '');
UPDATE recipes SET instructions = REPLACE(instructions, ' (för 10 port)', '');
UPDATE recipes SET instructions = REPLACE(instructions, ' (för 11 port)', '');
UPDATE recipes SET instructions = REPLACE(instructions, ' (för 12 port)', '');

-- Special case: fläskpannkaka has a comma-extended parenthetical.
UPDATE recipes
SET instructions = REPLACE(
    instructions,
    ' (för 4 port, bilden visar fläskpannkaka för 2 port)',
    ''
);

-- 2. Remove dangling reference to recipes that don't exist in the app.
UPDATE recipes
SET instructions = REPLACE(
    instructions,
    ', t ex Ugnsrostade rotfrukter och Het yoghurtsås',
    ''
)
WHERE id = 'rec_484a9cf1-4086-43dc-9653-563d067ee7a9';

-- 3. Normalize ˚C (U+02DA RING ABOVE) to °C (U+00B0 DEGREE SIGN).
UPDATE recipes SET instructions = REPLACE(instructions, '˚C', '°C');
UPDATE recipes SET instructions = REPLACE(instructions, '˚c', '°c');
