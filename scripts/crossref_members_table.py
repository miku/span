import fileinput
import json

for line in fileinput.input():
    doc = json.loads(line)
    id = doc["id"]
    for prefix in doc.get("prefix", []):
        values = (str(id), prefix["value"], str(doc["counts"]["total-dois"]), prefix["name"])
        print("\t".join(values))
