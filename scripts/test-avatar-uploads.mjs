#!/usr/bin/env node

import { readFile } from 'node:fs/promises';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const repoRoot = path.resolve(__dirname, '..');
const imageDir = path.join(repoRoot, 'test', 'images');

const fixtures = [
	{ slug: 'jpg', file: 'avatar.jpg', mime: 'image/jpeg' },
	{ slug: 'png', file: 'avatar.png', mime: 'image/png' },
	{ slug: 'webp', file: 'avatar.webp', mime: 'image/webp' },
	{ slug: 'gif', file: 'avatar.gif', mime: 'image/gif' }
];

const baseUrl = process.env.POCKETBASE_URL || 'http://127.0.0.1:8090';
const password = process.env.TEST_PASSWORD || 'Test1234!';
const timestamp = Date.now();

async function postUserWithAvatar({ slug, file, mime }) {
	const filePath = path.join(imageDir, file);
	const buffer = await readFile(filePath);
	const formData = new FormData();
	const email = `${slug}-avatar+${timestamp}@test.local`;
	const username = `${slug}-avatar-${timestamp}`;

	formData.append('email', email);
	formData.append('username', username);
	formData.append('password', password);
	formData.append('passwordConfirm', password);
	formData.append('status', 'pending');
	formData.append('avatar', new File([buffer], file, { type: mime }));

	const response = await fetch(`${baseUrl}/api/collections/users/records`, {
		method: 'POST',
		body: formData
	});

	let payload;
	try {
		payload = await response.json();
	} catch {
		payload = null;
	}

	if (!response.ok) {
		const details = payload ? JSON.stringify(payload, null, 2) : response.statusText;
		throw new Error(`Upload failed for ${file}: ${details}`);
	}

	return { email, id: payload.id };
}

async function run() {
	console.log(`Posting ${fixtures.length} avatar samples to ${baseUrl}`);
	const results = [];
	for (const fixture of fixtures) {
		try {
			const result = await postUserWithAvatar(fixture);
			results.push({ format: fixture.slug.toUpperCase(), email: result.email, id: result.id });
			console.log(`✅ ${fixture.slug.toUpperCase()} uploaded (${result.email})`);
		} catch (err) {
			console.error(`❌ ${fixture.slug.toUpperCase()} failed: ${err.message}`);
			process.exitCode = 1;
		}
	}

	if (results.length) {
		console.log('\nCreated records:');
		results.forEach(r => console.log(`- ${r.format}: ${r.email} (id: ${r.id})`));
	}
}

run().catch(err => {
	console.error(err);
	process.exitCode = 1;
});
