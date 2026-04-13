-- Seed data: Additional Swedish recipe classics (15 more, bringing total to 20)
-- Uses proper rec_<uuid> format for validated ID access

INSERT OR IGNORE INTO recipes (id, name, servings, emoji, tags, ingredients, instructions) VALUES
('rec_38f34854-b7e6-4950-b84e-d0321403fb05', 'Pannkakor', 4, '🥞', '["klassiker", "barn", "snabb"]',
 '[{"name": "Ägg", "amount": 3, "unit": "st"}, {"name": "Mjölk", "amount": 6, "unit": "dl"}, {"name": "Vetemjöl", "amount": 3, "unit": "dl"}, {"name": "Smör", "amount": 2, "unit": "msk"}, {"name": "Salt", "amount": 0.5, "unit": "tsk"}]',
 '["Vispa ihop ägg och hälften av mjölken", "Tillsätt mjöl och rör till slät smet", "Späd med resten av mjölken och smält smör", "Låt smeten vila 10 minuter", "Stek tunna pannkakor i smör", "Servera med sylt och grädde"]');

INSERT OR IGNORE INTO recipes (id, name, servings, emoji, tags, ingredients, instructions) VALUES
('rec_72f0ca56-5860-4496-9831-af11a83a28f0', 'Kycklinggryta med ris', 4, '🍗', '["kyckling", "vardag", "gryta"]',
 '[{"name": "Kycklingfilé", "amount": 500, "unit": "g"}, {"name": "Kokosmjölk", "amount": 400, "unit": "ml"}, {"name": "Curry", "amount": 2, "unit": "msk"}, {"name": "Lök", "amount": 1, "unit": "st"}, {"name": "Vitlök", "amount": 2, "unit": "klyftor"}, {"name": "Ris", "amount": 4, "unit": "dl"}, {"name": "Paprika", "amount": 1, "unit": "st"}]',
 '["Skär kycklingen i bitar", "Fräs lök och vitlök i en gryta", "Tillsätt kyckling och bryn", "Rör ner curry och kokosmjölk", "Låt sjuda 20 minuter", "Koka riset enligt förpackningen", "Servera grytan med ris"]');

INSERT OR IGNORE INTO recipes (id, name, servings, emoji, tags, ingredients, instructions) VALUES
('rec_4b1ef646-ca02-4cf3-a83f-d966a325db13', 'Ärtsoppa med pannkakor', 4, '🫛', '["klassiker", "soppa", "husmanskost"]',
 '[{"name": "Gula ärtor", "amount": 5, "unit": "dl"}, {"name": "Lök", "amount": 1, "unit": "st"}, {"name": "Morot", "amount": 2, "unit": "st"}, {"name": "Fläskkorv", "amount": 300, "unit": "g"}, {"name": "Senap", "amount": 2, "unit": "msk"}, {"name": "Timjan", "amount": 1, "unit": "tsk"}]',
 '["Skölj och blötlägg ärtorna över natten", "Koka ärtorna i nytt vatten ca 1 timme", "Tillsätt hackad lök och morot", "Låt sjuda tills ärtorna är mjuka", "Skär korven i skivor och lägg i soppan", "Smaka av med senap, salt och peppar", "Servera med pannkakor som tillbehör"]');

INSERT OR IGNORE INTO recipes (id, name, servings, emoji, tags, ingredients, instructions) VALUES
('rec_0daf7b54-169f-4c30-8ffd-5b86c94b454d', 'Falukorv med stuvade makaroner', 4, '🌭', '["husmanskost", "barn", "snabb"]',
 '[{"name": "Falukorv", "amount": 400, "unit": "g"}, {"name": "Makaroner", "amount": 4, "unit": "dl"}, {"name": "Smör", "amount": 2, "unit": "msk"}, {"name": "Vetemjöl", "amount": 2, "unit": "msk"}, {"name": "Mjölk", "amount": 5, "unit": "dl"}]',
 '["Koka makaronerna enligt förpackningen", "Skär falukorven i skivor och stek gyllene", "Smält smör i en kastrull", "Rör ner mjöl och låt fräsa", "Späd med mjölk under omrörning", "Låt stuvningen koka 5 minuter", "Blanda ner makaronerna", "Servera med falukorv och ketchup"]');

INSERT OR IGNORE INTO recipes (id, name, servings, emoji, tags, ingredients, instructions) VALUES
('rec_c70fc46a-3169-4a7f-ad83-ff643b6cb433', 'Fiskpinnar med potatismos', 4, '🐠', '["fisk", "barn", "snabb"]',
 '[{"name": "Fiskpinnar", "amount": 12, "unit": "st"}, {"name": "Potatis", "amount": 800, "unit": "g"}, {"name": "Mjölk", "amount": 1, "unit": "dl"}, {"name": "Smör", "amount": 2, "unit": "msk"}, {"name": "Ärtor", "amount": 3, "unit": "dl"}]',
 '["Sätt ugnen på 225°C", "Lägg fiskpinnarna på plåt med bakplåtspapper", "Grädda 15-20 minuter", "Skala och koka potatisen mjuk", "Mosa potatisen med mjölk och smör", "Koka ärtorna", "Servera fiskpinnar med mos, ärtor och ketchup"]');

INSERT OR IGNORE INTO recipes (id, name, servings, emoji, tags, ingredients, instructions) VALUES
('rec_6908af90-a306-4889-b501-c93c724d3cab', 'Korvstroganoff med ris', 4, '🍛', '["husmanskost", "barn", "snabb"]',
 '[{"name": "Falukorv", "amount": 400, "unit": "g"}, {"name": "Lök", "amount": 1, "unit": "st"}, {"name": "Tomatpuré", "amount": 2, "unit": "msk"}, {"name": "Grädde", "amount": 2, "unit": "dl"}, {"name": "Ris", "amount": 4, "unit": "dl"}, {"name": "Smör", "amount": 1, "unit": "msk"}]',
 '["Skär falukorven i stavar", "Hacka löken", "Fräs lök och korv i smör", "Rör ner tomatpuré", "Tillsätt grädde och låt sjuda 10 minuter", "Koka riset enligt förpackningen", "Servera stroganoff med ris"]');

INSERT OR IGNORE INTO recipes (id, name, servings, emoji, tags, ingredients, instructions) VALUES
('rec_5b4e5678-3f3f-4cd3-925b-4a2519730cfa', 'Janssons frestelse', 4, '🥔', '["klassiker", "husmanskost", "högtid"]',
 '[{"name": "Potatis", "amount": 800, "unit": "g"}, {"name": "Ansjovis", "amount": 125, "unit": "g"}, {"name": "Lök", "amount": 2, "unit": "st"}, {"name": "Grädde", "amount": 3, "unit": "dl"}, {"name": "Ströbröd", "amount": 2, "unit": "msk"}, {"name": "Smör", "amount": 2, "unit": "msk"}]',
 '["Sätt ugnen på 225°C", "Skala och skär potatisen i tunna stavar", "Skiva löken tunt", "Varva potatis, lök och ansjovis i smord ugnsform", "Häll över grädde och ansjovisspåd", "Strö över ströbröd och klicka smör", "Gratinera ca 45 minuter tills gyllene"]');

INSERT OR IGNORE INTO recipes (id, name, servings, emoji, tags, ingredients, instructions) VALUES
('rec_49d18675-48ce-4921-a0f0-d16faf26f3af', 'Pytt i panna', 4, '🍳', '["husmanskost", "rester", "snabb"]',
 '[{"name": "Potatis", "amount": 600, "unit": "g"}, {"name": "Falukorv", "amount": 300, "unit": "g"}, {"name": "Lök", "amount": 1, "unit": "st"}, {"name": "Smör", "amount": 2, "unit": "msk"}, {"name": "Ägg", "amount": 4, "unit": "st"}, {"name": "Rödbetor", "amount": 200, "unit": "g"}]',
 '["Skala och tärna potatisen", "Tärna falukorven", "Hacka löken", "Stek potatisen gyllene i smör", "Tillsätt korv och lök, stek vidare", "Stek ägg i separat panna", "Servera pytten med stekt ägg och rödbetor"]');

INSERT OR IGNORE INTO recipes (id, name, servings, emoji, tags, ingredients, instructions) VALUES
('rec_48b585bf-8a4e-4476-a1ed-86a0878288b0', 'Vegetarisk pasta med pesto', 4, '🌿', '["vegetariskt", "pasta", "snabb"]',
 '[{"name": "Pasta", "amount": 400, "unit": "g"}, {"name": "Pesto", "amount": 1, "unit": "dl"}, {"name": "Körsbärstomater", "amount": 250, "unit": "g"}, {"name": "Mozzarella", "amount": 125, "unit": "g"}, {"name": "Parmesanost", "amount": 50, "unit": "g"}, {"name": "Ruccola", "amount": 50, "unit": "g"}]',
 '["Koka pastan enligt förpackningen", "Dela tomaterna på mitten", "Riv eller tärna mozzarellan", "Blanda varm pasta med pesto", "Vänd ner tomater och mozzarella", "Toppa med ruccola och riven parmesan"]');

INSERT OR IGNORE INTO recipes (id, name, servings, emoji, tags, ingredients, instructions) VALUES
('rec_152d961c-c8fb-4176-b472-28595c67387d', 'Stekt fläsk med löksås', 4, '🥓', '["husmanskost", "klassiker"]',
 '[{"name": "Sidfläsk", "amount": 400, "unit": "g"}, {"name": "Lök", "amount": 3, "unit": "st"}, {"name": "Smör", "amount": 2, "unit": "msk"}, {"name": "Vetemjöl", "amount": 2, "unit": "msk"}, {"name": "Mjölk", "amount": 5, "unit": "dl"}, {"name": "Potatis", "amount": 800, "unit": "g"}, {"name": "Lingonsylt", "amount": 1, "unit": "burk"}]',
 '["Skär fläsket i skivor", "Stek fläsket krispigt i panna", "Skiva löken och fräs i smör", "Strö över mjöl och rör", "Späd med mjölk under omrörning", "Låt löksåsen koka 10 minuter", "Koka potatisen", "Servera med lingonsylt"]');

INSERT OR IGNORE INTO recipes (id, name, servings, emoji, tags, ingredients, instructions) VALUES
('rec_c7b2cc91-3b7d-49c4-86a8-9850d2fa9645', 'Köttbullar med gräddsås och potatis', 4, '🧆', '["klassiker", "husmanskost", "barn"]',
 '[{"name": "Köttfärs", "amount": 500, "unit": "g"}, {"name": "Lök", "amount": 1, "unit": "st"}, {"name": "Ströbröd", "amount": 0.5, "unit": "dl"}, {"name": "Ägg", "amount": 1, "unit": "st"}, {"name": "Mjölk", "amount": 1, "unit": "dl"}, {"name": "Grädde", "amount": 2, "unit": "dl"}, {"name": "Soja", "amount": 1, "unit": "msk"}, {"name": "Potatis", "amount": 800, "unit": "g"}, {"name": "Lingonsylt", "amount": 1, "unit": "burk"}]',
 '["Blötlägg ströbröd i mjölk", "Blanda köttfärs, ägg, finriven lök och ströbröd", "Krydda med salt och peppar", "Rulla till köttbullar", "Stek köttbullarna runtom i smör", "Gör gräddsås i stekpannan med grädde och soja", "Koka potatisen", "Servera med lingonsylt"]');

INSERT OR IGNORE INTO recipes (id, name, servings, emoji, tags, ingredients, instructions) VALUES
('rec_90106211-1c40-473c-b322-f77e0cc2f1d2', 'Ugnsbakad torsk med rotfrukter', 4, '🐟', '["fisk", "nyttigt", "vardag"]',
 '[{"name": "Torskfilé", "amount": 600, "unit": "g"}, {"name": "Morot", "amount": 3, "unit": "st"}, {"name": "Palsternacka", "amount": 2, "unit": "st"}, {"name": "Potatis", "amount": 4, "unit": "st"}, {"name": "Olivolja", "amount": 2, "unit": "msk"}, {"name": "Citron", "amount": 1, "unit": "st"}, {"name": "Timjan", "amount": 1, "unit": "tsk"}]',
 '["Sätt ugnen på 200°C", "Skala och skär rotfrukterna i bitar", "Lägg rotfrukterna på plåt med olivolja, salt och timjan", "Rosta i ugnen 20 minuter", "Lägg torskfiléerna ovanpå rotfrukterna", "Ringla citronjuice och olivolja över fisken", "Baka ytterligare 15 minuter", "Servera direkt från plåten"]');

INSERT OR IGNORE INTO recipes (id, name, servings, emoji, tags, ingredients, instructions) VALUES
('rec_73a7ce9a-f2b6-4932-b6c2-7dc92f8bb3e2', 'Chili con carne', 4, '🌶️', '["mexikanskt", "gryta", "vardag"]',
 '[{"name": "Köttfärs", "amount": 500, "unit": "g"}, {"name": "Kidneybönor", "amount": 400, "unit": "g"}, {"name": "Krossade tomater", "amount": 400, "unit": "g"}, {"name": "Lök", "amount": 1, "unit": "st"}, {"name": "Vitlök", "amount": 2, "unit": "klyftor"}, {"name": "Chilipulver", "amount": 1, "unit": "msk"}, {"name": "Spiskummin", "amount": 1, "unit": "tsk"}, {"name": "Ris", "amount": 4, "unit": "dl"}]',
 '["Hacka lök och vitlök", "Bryn köttfärsen i en gryta", "Fräs lök och vitlök", "Tillsätt krossade tomater", "Rör ner kryddor och bönor", "Låt sjuda 30 minuter", "Koka riset", "Servera med gräddfil och ris"]');

INSERT OR IGNORE INTO recipes (id, name, servings, emoji, tags, ingredients, instructions) VALUES
('rec_7de43fb9-6c51-455f-bc90-380f2244e59c', 'Tomatsoppa med ostmacka', 4, '🍅', '["soppa", "snabb", "vegetariskt"]',
 '[{"name": "Krossade tomater", "amount": 800, "unit": "g"}, {"name": "Lök", "amount": 1, "unit": "st"}, {"name": "Vitlök", "amount": 2, "unit": "klyftor"}, {"name": "Grönsaksbuljong", "amount": 5, "unit": "dl"}, {"name": "Basilika", "amount": 1, "unit": "msk"}, {"name": "Grädde", "amount": 1, "unit": "dl"}, {"name": "Bröd", "amount": 8, "unit": "skivor"}, {"name": "Ost", "amount": 200, "unit": "g"}]',
 '["Fräs hackad lök och vitlök i smör", "Tillsätt krossade tomater och buljong", "Krydda med basilika, salt och peppar", "Låt koka 15 minuter", "Mixa soppan slät med stavmixer", "Rör ner grädde", "Lägg ost mellan brödskivorna", "Grilla ostmackorna i stekpanna", "Servera soppan med grillad ostmacka"]');

INSERT OR IGNORE INTO recipes (id, name, servings, emoji, tags, ingredients, instructions) VALUES
('rec_d7eed46b-b5a2-4a92-9ea9-de1b513254fd', 'Kyckling med currysås och ris', 4, '🍛', '["kyckling", "vardag", "snabb"]',
 '[{"name": "Kycklingfilé", "amount": 500, "unit": "g"}, {"name": "Curry", "amount": 2, "unit": "msk"}, {"name": "Grädde", "amount": 2, "unit": "dl"}, {"name": "Lök", "amount": 1, "unit": "st"}, {"name": "Banan", "amount": 1, "unit": "st"}, {"name": "Jordnötter", "amount": 0.5, "unit": "dl"}, {"name": "Ris", "amount": 4, "unit": "dl"}, {"name": "Mango chutney", "amount": 2, "unit": "msk"}]',
 '["Skär kycklingen i bitar", "Stek kycklingen gyllene i smör", "Fräs hackad lök", "Strö över curry och rör", "Tillsätt grädde och mango chutney", "Låt sjuda 15 minuter", "Koka riset", "Servera med skivad banan och jordnötter"]');
