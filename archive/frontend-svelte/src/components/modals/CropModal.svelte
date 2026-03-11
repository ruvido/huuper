<script>
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
	let CropperComponent = null;
	let cropperLoading = false;

	// Reset state when modal is shown
	$: if (show) {
		crop = { x: 0, y: 0 };
		zoom = 1;
		croppedAreaPixels = null;
		error = '';
		loading = false;
	}

	$: if (show && !CropperComponent && !cropperLoading) {
		void loadCropper();
	}

	async function loadCropper() {
		cropperLoading = true;
		try {
			const module = await import('svelte-easy-crop');
			CropperComponent = module.default;
		} catch {
			error = 'Errore nel caricamento del ritaglio';
		} finally {
			cropperLoading = false;
		}
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

	function canvasToBlob(canvas, type, quality) {
		return new Promise((resolve, reject) => {
			canvas.toBlob((blob) => {
				if (blob) {
					resolve(blob);
					return;
				}
				reject(new Error('Impossibile creare il file immagine'));
			}, type, quality);
		});
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
		try {
			if (typeof createImageBitmap === 'function') {
				return await createImageBitmap(file);
			}
		} catch {
			// Fallback to HTMLImageElement for older browsers.
		}

		const objectUrl = URL.createObjectURL(file);
		try {
			const image = await new Promise((resolve, reject) => {
				const img = new Image();
				img.onload = () => resolve(img);
				img.onerror = () => reject(new Error('Formato immagine non supportato'));
				img.src = objectUrl;
			});
			return image;
		} finally {
			URL.revokeObjectURL(objectUrl);
		}
	}

	// Export a function that parent can call with the original file
	export async function processCrop(originalFile) {
		if (!croppedAreaPixels || !originalFile) return null;

		loading = true;
		error = '';

		try {
			const image = await decodeImageFile(originalFile);
			const { x, y, width, height } = clampCropArea(
				{ width: image.width, height: image.height },
				croppedAreaPixels
			);
			const targetSize = getResizeDimensions(width, height);
			const canvas = document.createElement('canvas');
			canvas.width = targetSize.width;
			canvas.height = targetSize.height;
			const ctx = canvas.getContext('2d');
			if (!ctx) {
				throw new Error('Canvas non disponibile');
			}
			ctx.imageSmoothingEnabled = true;
			ctx.imageSmoothingQuality = 'high';
			ctx.drawImage(image, x, y, width, height, 0, 0, targetSize.width, targetSize.height);

			const blob = await canvasToBlob(canvas, 'image/jpeg', JPEG_QUALITY / 100);
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
				{#if CropperComponent}
					<svelte:component
						this={CropperComponent}
						{image}
						bind:crop
						bind:zoom
						aspect={1}
						oncropcomplete={handleCropComplete}
					/>
				{:else}
					<p class="loading">Caricamento editor...</p>
				{/if}
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

	.loading {
		color: #fff;
		font-size: 0.95rem;
		height: 100%;
		display: flex;
		align-items: center;
		justify-content: center;
	}
</style>
