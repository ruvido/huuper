<script>
	import Cropper from 'svelte-easy-crop';
	import { decode as decodeJpeg, encode as encodeJpeg } from '@jsquash/jpeg';
	import { decode as decodePng } from '@jsquash/png';
	import { decode as decodeWebp } from '@jsquash/webp';
	import resize from '@jsquash/resize';

	export let show = false;
	export let image = '';
	export let onConfirm;
	export let onCancel;

	const MAX_AVATAR_SIZE = 400;
	const JPEG_QUALITY = 85;

	let crop = { x: 0, y: 0 };
	let zoom = 1;
	let croppedAreaPixels = null;
	let loading = false;
	let error = '';

	// Reset state when modal is shown
	$: if (show) {
		crop = { x: 0, y: 0 };
		zoom = 1;
		croppedAreaPixels = null;
		error = '';
		loading = false;
	}

	function handleCropComplete(e) {
		croppedAreaPixels = e.pixels;
	}

	function clampCropArea(imageData, area) {
		const startX = Math.max(0, Math.min(Math.floor(area.x), imageData.width - 1));
		const startY = Math.max(0, Math.min(Math.floor(area.y), imageData.height - 1));
		const maxWidth = imageData.width - startX;
		const maxHeight = imageData.height - startY;

		const width = Math.max(1, Math.min(maxWidth, Math.round(area.width)));
		const height = Math.max(1, Math.min(maxHeight, Math.round(area.height)));

		return { x: startX, y: startY, width, height };
	}

	function cropImageData(imageData, area) {
		const { x, y, width, height } = clampCropArea(imageData, area);
		const croppedData = new Uint8ClampedArray(width * height * 4);

		for (let row = 0; row < height; row++) {
			const srcOffset = ((y + row) * imageData.width + x) * 4;
			const dstOffset = row * width * 4;
			croppedData.set(
				imageData.data.subarray(srcOffset, srcOffset + width * 4),
				dstOffset
			);
		}

		return new ImageData(croppedData, width, height);
	}

	function getResizeDimensions(width, height) {
		if (width <= 0 || height <= 0) {
			return { width: MAX_AVATAR_SIZE, height: MAX_AVATAR_SIZE };
		}

		if (width >= height) {
			return {
				width: MAX_AVATAR_SIZE,
				height: Math.max(1, Math.round((height / width) * MAX_AVATAR_SIZE)),
			};
		}

		return {
			width: Math.max(1, Math.round((width / height) * MAX_AVATAR_SIZE)),
			height: MAX_AVATAR_SIZE,
		};
	}

	async function decodeImageFile(file) {
		const buffer = await file.arrayBuffer();
		const mime = (file.type || '').toLowerCase();
		const name = (file.name || '').toLowerCase();
		const hints = [mime, name];
		const attempts = [];

		if (hints.some(value => value.includes('png'))) {
			attempts.push(() => decodePng(buffer));
		}

		if (hints.some(value => value.includes('webp'))) {
			attempts.push(() => decodeWebp(buffer));
		}

		if (hints.some(value => value.includes('jpg') || value.includes('jpeg'))) {
			attempts.push(() => decodeJpeg(buffer, { preserveOrientation: true }));
		}

		attempts.push(() => decodeJpeg(buffer, { preserveOrientation: true }));

		let lastError;
		for (const decodeFn of attempts) {
			try {
				return await decodeFn();
			} catch (err) {
				lastError = err;
			}
		}

		throw lastError || new Error('Formato immagine non supportato');
	}

	// Export a function that parent can call with the original file
	export async function processCrop(originalFile) {
		if (!croppedAreaPixels || !originalFile) return null;

		loading = true;
		error = '';

		try {
			const imageData = await decodeImageFile(originalFile);
			const cropped = cropImageData(imageData, croppedAreaPixels);
			const { width, height } = getResizeDimensions(cropped.width, cropped.height);
			const resized = await resize(cropped, { width, height });
			const jpegBuffer = await encodeJpeg(resized, { quality: JPEG_QUALITY });
			const blob = new Blob([jpegBuffer], { type: 'image/jpeg' });
			const fileName = originalFile.name.replace(/\.[^/.]+$/, '.jpg');
			const file = new File([blob], fileName, { type: 'image/jpeg' });

			loading = false;
			return file;
		} catch (err) {
			error = 'Errore nel processare l\'immagine';
			loading = false;
			throw err;
		}
	}

	async function handleConfirm() {
		if (!croppedAreaPixels) return;

		loading = true;
		error = '';

		try {
			// Signal to parent that we're ready to process
			// Parent must call processCrop with the original file
			if (onConfirm) {
				await onConfirm();
			}
		} catch (err) {
			error = 'Errore nel processare l\'immagine';
			loading = false;
		}
	}

	function handleCancel() {
		if (onCancel) {
			onCancel();
		}
	}
</script>

{#if show}
	<div class="crop-overlay">
		<div class="crop-modal">
			<h2>Ritaglia foto</h2>
			<div class="crop-area">
				<Cropper
					{image}
					bind:crop
					bind:zoom
					aspect={1}
					oncropcomplete={handleCropComplete}
				/>
			</div>
			{#if error}
				<p class="error">{error}</p>
			{/if}
			<div class="crop-buttons">
				<button type="button" class="crop-btn cancel" on:click={handleCancel} disabled={loading}>
					Annulla
				</button>
				<button type="button" class="crop-btn confirm" on:click={handleConfirm} disabled={loading || !croppedAreaPixels}>
					{loading ? 'Elaborazione...' : 'Conferma'}
				</button>
			</div>
		</div>
	</div>
{/if}

<style>
	.crop-overlay {
		position: fixed;
		top: 0;
		left: 0;
		width: 100%;
		height: 100%;
		background: rgba(0, 0, 0, 0.9);
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 2000;
	}

	.crop-modal {
		background: #fff;
		padding: 1.5rem;
		border-radius: 8px;
		max-width: 450px;
		width: 90%;
	}

	.crop-modal h2 {
		margin: 0 0 1rem 0;
		font-size: 1.25rem;
		text-align: center;
	}

	.crop-area {
		position: relative;
		width: 100%;
		height: 300px;
		background: #000;
		margin-bottom: 1rem;
	}

	.crop-buttons {
		display: flex;
		gap: 1rem;
		justify-content: flex-end;
	}

	.crop-btn {
		padding: 0.75rem 1.5rem;
		font-size: 1rem;
		font-family: inherit;
		font-weight: 600;
		border: 2px solid #000;
		cursor: pointer;
		transition: all 0.2s ease;
	}

	.crop-btn.cancel {
		background: #fff;
		color: #000;
	}

	.crop-btn.cancel:hover:not(:disabled) {
		background: #f0f0f0;
	}

	.crop-btn.confirm {
		background: #000;
		color: #fff;
	}

	.crop-btn.confirm:hover:not(:disabled) {
		background: #333;
	}

	.crop-btn:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.error {
		color: #d32f2f;
		font-size: 0.875rem;
		margin: 0.5rem 0;
		text-align: center;
	}
</style>
