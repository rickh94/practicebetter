const { exec } = require("child_process");
const fs = require("fs");
const os = require("os");
const path = require("path");
const { convertParsedSVG, parseSVGContent } = require("@iconify/utils");
// Nodejs script for automatically generating key signature icons using lilypond

function makeLilypondContents(key, mode) {
  return (
    `\\version "2.24.3"\n` +
    `\\language "english"\n` +
    `#(ly:set-option 'crop #t)\n` +
    `#(set-global-staff-size 8)\n` +
    `\\score {\n` +
    `\\new Staff {\n` +
    `\\relative {\n` +
    `\\omit Staff.TimeSignature\n` +
    `\\key ${key} \\${mode}\n` +
    `\\time 1/128\n` +
    `\\override Score.BarLine.transparent = ##t\n` +
    `s128 |\n` +
    `}\n` +
    `}\n` +
    `}\n`
  );
}

function main() {
  const tmpdir = fs.mkdtempSync(path.join(os.tmpdir(), "lilypond-"));
  console.log(tmpdir);
  const names = [];

  // Create Lilypond files
  ["c", "g", "d", "a", "e", "b", "gf", "df", "af", "ef", "bf", "f"].forEach(
    (key) => {
      const name = `${key}-major`;
      const contents = makeLilypondContents(key, "major");
      fs.writeFileSync(path.join(tmpdir, `${name}.ly`), contents);
      names.push(name);
    },
  );

  ["a", "e", "b", "fs", "cs", "gs", "ef", "bf", "f", "c", "g", "d"].forEach(
    (key) => {
      const name = `${key}-minor`;
      const contents = makeLilypondContents(key, "minor");
      fs.writeFileSync(path.join(tmpdir, `${name}.ly`), contents);
      names.push(name);
    },
  );

  console.log(fs.readdirSync(tmpdir));

  // Invoke Lilypond to convert files to SVG
  const outputDir = path.join(tmpdir, "output");
  fs.mkdirSync(outputDir, { recursive: true });

  const processes = names.map((name) => {
    return exec(
      `lilypond -o "${path.join(outputDir, name)}" -dno-point-and-click --svg "${name}.ly"`,
      {
        cwd: tmpdir,
      },
    );
  });

  Promise.all(
    processes.map(
      (p) =>
        new Promise((resolve, reject) => {
          p.on("close", (code) => {
            if (code === 0) {
              resolve();
            } else {
              reject();
            }
          });
        }),
    ),
  )
    .then(() => {
      const icons = {};
      names.forEach((name) => {
        const svgFile = path.join(outputDir, `${name}.cropped.svg`);
        if (fs.existsSync(svgFile)) {
          const svg = fs.readFileSync(svgFile, "utf-8");
          const parsed = parseSVGContent(svg);
          const converted = convertParsedSVG(parsed);
          icons[name] = converted;
        } else {
          console.log(`SVG file ${svgFile} not found.`);
        }
      });
      fs.writeFileSync(
        path.join(__dirname, "internal", "static", "keysignatures.json"),
        JSON.stringify({ prefix: "key", icons }, null, 2),
      );
      for (const name of names) {
        console.log(`icon-[key--${name}]`);
      }
    })
    .catch((error) => {
      console.error("An error occurred during conversion:", error);
    });
}

main();
