#!/usr/bin/env python3
"""One-off generator: fetch Livsmedelsverket Livsmedelsdatabasen and emit seed SQL.
Source: https://dataportal.livsmedelsverket.se/livsmedel (CC-BY 4.0).
Run once; output committed as backend/migrations/021_seed_livsmedel.sql."""
import json, urllib.request, concurrent.futures, sys, time

BASE = "https://dataportal.livsmedelsverket.se/livsmedel/api/v1"
MACRO_NAMES = {
    "Energi (kcal)": "kcal",
    "Protein": "protein",
    "Kolhydrater, tillgängliga": "carbs",
    "Fett, totalt": "fat",
}

def get(url, retries=4):
    for i in range(retries):
        try:
            req = urllib.request.Request(url, headers={"Accept": "application/json"})
            with urllib.request.urlopen(req, timeout=30) as r:
                return json.load(r)
        except Exception as e:
            if i == retries - 1:
                raise
            time.sleep(1 + i)

def fetch_list():
    foods = []
    offset, limit = 0, 500
    while True:
        d = get(f"{BASE}/livsmedel?offset={offset}&limit={limit}&sprak=1")
        foods.extend((f["nummer"], f["namn"]) for f in d["livsmedel"])
        total = d["_meta"]["totalRecords"]
        offset += limit
        if offset >= total:
            break
    return foods

def fetch_macros(item):
    nummer, namn = item
    macros = {"kcal": 0.0, "protein": 0.0, "carbs": 0.0, "fat": 0.0}
    try:
        d = get(f"{BASE}/livsmedel/{nummer}/naringsvarden?sprak=1")
        for n in d:
            key = MACRO_NAMES.get(n.get("namn", ""))
            if key is not None:
                macros[key] = float(n.get("varde") or 0)
    except Exception as e:
        print(f"WARN {nummer} {namn}: {e}", file=sys.stderr)
    return (nummer, namn, macros)

def esc(s):
    return s.replace("'", "''")

def main():
    print("Fetching food list...", file=sys.stderr)
    foods = fetch_list()
    print(f"{len(foods)} foods. Fetching nutrients...", file=sys.stderr)
    rows = []
    done = 0
    with concurrent.futures.ThreadPoolExecutor(max_workers=12) as ex:
        for r in ex.map(fetch_macros, foods):
            rows.append(r)
            done += 1
            if done % 200 == 0:
                print(f"  {done}/{len(foods)}", file=sys.stderr)
    rows.sort(key=lambda x: x[0])
    with open("/tmp/021_seed_livsmedel.sql", "w") as f:
        f.write("-- Livsmedelsverkets Livsmedelsdatabasen (food composition data).\n")
        f.write("-- Source: https://dataportal.livsmedelsverket.se/livsmedel\n")
        f.write("-- Licensed under Creative Commons CC-BY 4.0. Attribution: Livsmedelsverket (Swedish Food Agency).\n")
        f.write("-- Generated once via scripts/gen_livsmedel.py; values are per 100 g edible portion.\n\n")
        # chunked multi-row inserts
        CHUNK = 100
        for i in range(0, len(rows), CHUNK):
            chunk = rows[i:i + CHUNK]
            f.write("INSERT INTO livsmedel (livsmedelsnummer, namn, kcal_per_100g, protein_per_100g, carbs_per_100g, fat_per_100g) VALUES\n")
            vals = []
            for nummer, namn, m in chunk:
                vals.append(f"  ({nummer}, '{esc(namn)}', {m['kcal']:g}, {m['protein']:g}, {m['carbs']:g}, {m['fat']:g})")
            f.write(",\n".join(vals) + ";\n\n")
    print(f"WROTE /tmp/021_seed_livsmedel.sql ({len(rows)} rows)", file=sys.stderr)

if __name__ == "__main__":
    main()
