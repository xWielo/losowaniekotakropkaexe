import os, json

EXTENSIONS = {'.jpg', '.jpeg', '.png', '.gif', '.webp', '.avif'}
folder = 'koty'

pliki = [
    f for f in os.listdir(folder)
    if os.path.splitext(f)[1].lower() in EXTENSIONS
]

with open(os.path.join(folder, 'lista.json'), 'w') as fp:
    json.dump(pliki, fp, ensure_ascii=False, indent=2)

print(f"Zapisano {len(pliki)} zdjęć do koty/lista.json")
