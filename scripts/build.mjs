#!/usr/bin/env node
// Build the site into dist/.
//
//   1. Resize every JPG in photos/ to each long-edge size in SIZES.
//   2. Write dist/photos.json (the manifest the slideshow reads).
//   3. Copy src/ to dist/, filling {{tokens}} in HTML files from
//      site.config.json and the manifest.
//
// Run with `npm run build`. Safe to re-run: existing outputs are skipped.
// Output file names carry a content hash, so replacing a photo with a
// new file of the same name produces new URLs and busts every cache.

import { readdir, mkdir, readFile, writeFile, copyFile, access, stat } from 'node:fs/promises';
import { createHash } from 'node:crypto';
import path from 'node:path';
import sharp from 'sharp';

const ROOT = path.resolve(import.meta.dirname, '..');
const PHOTOS_DIR = path.join(ROOT, 'photos');
const SRC_DIR = path.join(ROOT, 'src');
const DIST_DIR = path.join(ROOT, 'dist');
const IMG_DIR = path.join(DIST_DIR, 'img');

// Long-edge pixel sizes. Chosen from current screen-resolution share
// (phones, laptops, 1080p and 2x desktops). Order matters: ascending.
const SIZES = [1080, 1600, 2048, 2880];
const QUALITY = 82;       // mozjpeg quality; visually lossless for photos
const CONCURRENCY = 4;    // photos processed in parallel
const WARN_SOURCE_BYTES = 6 * 1024 * 1024; // nudge the photographer to export smaller

const exists = (p) => access(p).then(() => true, () => false);
const isJpg = (n) => /\.jpe?g$/i.test(n);

function slugify(s) {
  return s.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-+|-+$/g, '') || 'photo';
}

// Every JPG directly inside photos/. Subfolders and hidden files are ignored.
async function listPhotos() {
  const entries = await readdir(PHOTOS_DIR, { withFileTypes: true });
  return entries
    .filter((e) => e.isFile() && isJpg(e.name) && !e.name.startsWith('.'))
    .map((e) => e.name)
    .sort((a, b) => a.localeCompare(b, 'en'))
    .map((name) => ({ file: path.join(PHOTOS_DIR, name), name }));
}

async function processPhoto(photo, seenIds, config) {
  const buf = await readFile(photo.file);
  const hash = createHash('sha1').update(buf).digest('hex').slice(0, 8);

  let id = slugify(photo.name.replace(/\.jpe?g$/i, ''));
  if (seenIds.has(id)) id = `${id}-${hash}`;
  seenIds.add(id);

  if (buf.length > WARN_SOURCE_BYTES) {
    console.warn(`  note: ${photo.name} is ${(buf.length / 1048576).toFixed(1)} MB; exporting at 3000 px long edge keeps the repo small`);
  }

  // .rotate() with no argument applies the EXIF orientation, so a phone
  // photo shot in portrait comes out upright. Metadata is pre-rotation,
  // so swap width/height for the rotated orientations (5-8).
  const base = sharp(buf, { failOn: 'none' }).rotate();
  const meta = await base.metadata();
  const rotated = (meta.orientation || 1) >= 5;
  const w = rotated ? meta.height : meta.width;
  const h = rotated ? meta.width : meta.height;
  const longEdge = Math.max(w, h);

  const stem = `img/${id}-${hash}`;
  let written = 0;
  let bytes = 0;

  for (const size of SIZES) {
    // Never upscale. If the source is smaller than a target size we still
    // emit that size at native resolution so every photo has every size.
    const target = Math.min(size, longEdge);
    const out = path.join(DIST_DIR, `${stem}-${size}.jpg`);
    if (await exists(out)) { bytes += (await stat(out)).size; continue; }
    const info = await base
      .clone()
      .resize({ width: target, height: target, fit: 'inside', withoutEnlargement: true })
      .jpeg({ quality: QUALITY, mozjpeg: true, progressive: true })
      // sharp converts to sRGB and drops all source metadata (camera, GPS,
      // date). withExif then writes just the copyright fields back.
      .withExif({ IFD0: { Copyright: `© ${new Date().getFullYear()} ${config.name}`, Artist: config.name } })
      .toFile(out);
    written += 1;
    bytes += info.size;
  }

  return { entry: { f: stem, w, h }, written, bytes };
}

async function pool(items, limit, fn) {
  const results = new Array(items.length);
  let next = 0;
  async function worker() {
    while (next < items.length) {
      const i = next++;
      results[i] = await fn(items[i], i);
    }
  }
  await Promise.all(Array.from({ length: Math.min(limit, items.length) }, worker));
  return results;
}

async function copySrc(tokens) {
  let count = 0;
  async function walk(rel) {
    const entries = await readdir(path.join(SRC_DIR, rel), { withFileTypes: true });
    for (const e of entries) {
      const relPath = path.join(rel, e.name);
      const from = path.join(SRC_DIR, relPath);
      const to = path.join(DIST_DIR, relPath);
      if (e.isDirectory()) { await mkdir(to, { recursive: true }); await walk(relPath); continue; }
      if (/\.html?$/i.test(e.name)) {
        let html = await readFile(from, 'utf8');
        html = html.replace(/\{\{\s*([a-zA-Z0-9_]+)\s*\}\}/g, (m, key) =>
          key in tokens ? tokens[key] : (console.warn(`  note: unknown token ${m} in ${relPath}`), m));
        await writeFile(to, html);
      } else {
        await copyFile(from, to);
      }
      count += 1;
    }
  }
  await walk('');
  return count;
}

async function main() {
  const t0 = Date.now();
  const config = JSON.parse(await readFile(path.join(ROOT, 'site.config.json'), 'utf8'));
  await mkdir(IMG_DIR, { recursive: true });

  const files = await listPhotos();
  if (files.length === 0) console.warn('  note: photos/ has no JPGs');
  console.log(`Resizing ${files.length} photo(s) to ${SIZES.join(', ')} px long edge`);

  const seenIds = new Set();
  const results = await pool(files, CONCURRENCY, (f) => processPhoto(f, seenIds, config));

  const manifest = { sizes: SIZES, photos: results.map((r) => r.entry) };
  await writeFile(path.join(DIST_DIR, 'photos.json'), JSON.stringify(manifest));

  const tokens = {
    ...Object.fromEntries(Object.entries(config).filter(([, v]) => typeof v !== 'object').map(([k, v]) => [k, String(v)])),
    manifest: JSON.stringify(manifest),
    photoCount: String(manifest.photos.length),
    year: String(new Date().getFullYear()),
  };
  const copied = await copySrc(tokens);

  const written = results.reduce((n, r) => n + r.written, 0);
  const bytes = results.reduce((n, r) => n + r.bytes, 0);
  console.log(`Done in ${((Date.now() - t0) / 1000).toFixed(1)}s: ${written} new image file(s), ` +
    `${(bytes / 1048576).toFixed(1)} MB of images total, ${copied} page file(s) copied to dist/`);
}

main().catch((err) => { console.error(err); process.exit(1); });
