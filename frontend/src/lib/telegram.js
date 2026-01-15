import { pb } from './pocketbase';

export async function generateTelegramDeepLink(botName) {
	const response = await fetch('/api/telegram/generate-token', {
		method: 'POST',
		headers: {
			Authorization: pb.authStore.token,
		},
	});

	if (!response.ok) {
		throw new Error('Failed to generate connection token');
	}

	const data = await response.json();
	const token = data.token;
	const cleanBotName = (botName || '').replace('@', '');

	return `https://t.me/${cleanBotName}?start=${token}`;
}
