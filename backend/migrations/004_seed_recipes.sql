-- Seed data: Swedish recipe classics

INSERT OR IGNORE INTO recipes (id, name, servings, emoji, tags, ingredients, instructions) VALUES
('rec_1', 'Pasta Carbonara', 4, '🍝', '["pasta", "vardag", "snabb"]',
 '[{"name": "Spaghetti", "amount": 400, "unit": "g"}, {"name": "Bacon", "amount": 200, "unit": "g"}, {"name": "Ägg", "amount": 4, "unit": "st"}, {"name": "Parmesan", "amount": 100, "unit": "g"}, {"name": "Svartpeppar", "amount": 1, "unit": "tsk"}]',
 '["Koka pastan enligt förpackningen", "Stek baconet krispigt", "Vispa ihop ägg och parmesan", "Blanda pasta med bacon", "Rör ner äggblandningen", "Servera med extra parmesan"]');

INSERT OR IGNORE INTO recipes (id, name, servings, emoji, tags, ingredients, instructions) VALUES
('rec_2', 'Kycklingwok', 4, '🥘', '["kyckling", "asiatiskt", "vardag"]',
 '[{"name": "Kycklingfilé", "amount": 500, "unit": "g"}, {"name": "Wokgrönsaker", "amount": 400, "unit": "g"}, {"name": "Sojasås", "amount": 3, "unit": "msk"}, {"name": "Ris", "amount": 4, "unit": "dl"}]',
 '["Skär kycklingen i strimlor", "Stek kycklingen i het wok", "Tillsätt grönsaker", "Krydda med sojasås", "Servera med ris"]');

INSERT OR IGNORE INTO recipes (id, name, servings, emoji, tags, ingredients, instructions) VALUES
('rec_3', 'Tacos', 4, '🌮', '["mexikanskt", "fredagsmys", "barn"]',
 '[{"name": "Köttfärs", "amount": 500, "unit": "g"}, {"name": "Tacokrydda", "amount": 1, "unit": "påse"}, {"name": "Tacoskal", "amount": 12, "unit": "st"}, {"name": "Sallad", "amount": 1, "unit": "st"}, {"name": "Tomat", "amount": 3, "unit": "st"}, {"name": "Riven ost", "amount": 200, "unit": "g"}]',
 '["Stek köttfärsen", "Tillsätt krydda och vatten", "Låt sjuda 5 min", "Skär grönsaker", "Servera med tillbehör"]');

INSERT OR IGNORE INTO recipes (id, name, servings, emoji, tags, ingredients, instructions) VALUES
('rec_4', 'Laxfilé med potatis', 4, '🐟', '["fisk", "nyttigt", "vardag"]',
 '[{"name": "Laxfilé", "amount": 600, "unit": "g"}, {"name": "Potatis", "amount": 800, "unit": "g"}, {"name": "Citron", "amount": 1, "unit": "st"}, {"name": "Dill", "amount": 1, "unit": "knippe"}]',
 '["Sätt ugnen på 200°C", "Koka potatisen", "Lägg laxen i ugnsform", "Salta, peppra och lägg på citron", "Grädda 15-20 min", "Servera med dill"]');

INSERT OR IGNORE INTO recipes (id, name, servings, emoji, tags, ingredients, instructions) VALUES
('rec_5', 'Köttfärssås', 4, '🍖', '["pasta", "klassiker", "barn"]',
 '[{"name": "Köttfärs", "amount": 400, "unit": "g"}, {"name": "Krossade tomater", "amount": 400, "unit": "g"}, {"name": "Lök", "amount": 1, "unit": "st"}, {"name": "Vitlök", "amount": 2, "unit": "klyftor"}, {"name": "Pasta", "amount": 400, "unit": "g"}]',
 '["Hacka lök och vitlök", "Bryn köttfärsen", "Tillsätt lök och vitlök", "Häll i krossade tomater", "Låt sjuda 20 min", "Servera med pasta"]');
