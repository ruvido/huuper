#!/usr/bin/env node

import { readFile } from 'node:fs/promises';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import sharp from 'sharp';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const repoRoot = path.resolve(__dirname, '..');
const imageDir = path.join(repoRoot, 'test', 'images');
const baseImage = path.join(imageDir, 'avatar.webp');

const targets = [
	{ file: 'avatar.jpg', format: 'jpeg', options: { quality: 90 } },
	{ file: 'avatar.png', format: 'png', options: {} },
	{ file: 'avatar.gif', format: 'gif', options: {} }
];

async function convert() {
	const buffer = await readFile(baseImage);
	await Promise.all(
		targets.map(async ({ file, format, options }) => {
			const output = path.join(imageDir, file);
			await sharp(buffer)
				.toFormat(format, options)
				.toFile(output);
			console.log(`Generated ${file} from avatar.webp`);
		})
	);
}

convert().catch(err => {
	console.error('Failed to generate fixtures:', err);
	process.exit(1);
});
