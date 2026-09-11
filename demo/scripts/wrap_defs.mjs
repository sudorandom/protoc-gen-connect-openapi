// Wraps a $defs-only JSON Schema bundle (the format=jsonschema output) in a
// root schema referencing every definition. Type generators need a root type
// to start from; a bare $defs bundle gives them nothing to generate.
import { readFileSync, writeFileSync } from "node:fs";

const [input, output] = process.argv.slice(2);
if (input === undefined || output === undefined) {
  console.error("usage: wrap_defs.mjs <input.json> <output.json>");
  process.exit(1);
}

const doc = JSON.parse(readFileSync(input, "utf8"));
const defs = doc.$defs ?? {};
const names = Object.keys(defs);
if (names.length === 0) {
  console.error(`no $defs found in ${input}`);
  process.exit(1);
}

const wrapped = {
  $schema: doc.$schema ?? "https://json-schema.org/draft/2020-12/schema",
  title: "AllTypes",
  anyOf: names.map((name) => ({ $ref: `#/$defs/${name}` })),
  $defs: defs,
};
writeFileSync(output, JSON.stringify(wrapped, null, 2) + "\n");
