<script>
	import { slide } from 'svelte/transition';
	import Button from '../Button.svelte';
	import ErrorMessage from '../ErrorMessage.svelte';
	import GridSelector from '../ui/GridSelector.svelte';
	import { renderContent } from '../../lib/markdown';
	import { X, CheckCircle } from 'lucide-svelte';

	export let step;
	export let formData = {};
	export let loading = false;
	export let error = '';
	export let canProceed = false;
	export let isLastStep = false;
	export let onNext;
	export let onComplete;
	export let onClose;
	export let onToggleOption;
	export let onFileSelect;

	function handleEnterKey(e) {
		if (e.key === 'Enter' && canProceed) {
			e.preventDefault();
			if (isLastStep) {
				onComplete();
			} else {
				onNext();
			}
		}
	}
</script>

<div class="step-content" class:is-start={step.type === 'start'}>
	{#if step.type === 'start'}
		<button class="close-btn-start" on:click={onClose} disabled={loading}>
			<X size={24} />
		</button>
		<h1>{step.title}</h1>
		<div class="start-text">{@html renderContent(step.text)}</div>
		<Button variant="submit" on:click={onNext}>
			{step.button}
		</Button>
	{:else}
		<h1>{step.title}</h1>
		{#if step.type === 'text'}
			<label for={step.id}>{step.label}</label>
			<input
				id={step.id}
				type="text"
				bind:value={formData[step.field]}
				on:keydown={handleEnterKey}
				disabled={loading}
				placeholder="Scrivi qui..."
			/>
		{:else if step.type === 'textarea'}
			<label for={step.id}>{step.label}</label>
			<textarea
				id={step.id}
				bind:value={formData[step.field]}
				disabled={loading}
				rows="8"
				placeholder="Scrivi qui..."
			></textarea>
		{:else if step.type === 'file'}
			{#if formData[step.field]}
				<div class="success-check">
					<CheckCircle size={64} strokeWidth={2} />
				</div>
			{/if}
			<h2 class="file-label">{step.label}</h2>
			<input
				id={step.id}
				type="file"
				accept="image/*"
				on:change={(e) => {
					const file = e.target.files[0];
					if (file) {
						onFileSelect(file, step.field);
					}
				}}
				disabled={loading}
				style="display: none;"
			/>
			<button
				type="button"
				class="file-button"
				on:click={() => document.getElementById(step.id).click()}
				disabled={loading}
			>
				{formData[step.field] ? 'Cambia foto' : 'Carica foto'}
			</button>
			{#if formData[step.field]}
				<p class="file-name">{formData[step.field].name}</p>
			{/if}
		{:else if step.type === 'select'}
			{@const isMultiple = !!step.min}
			{@const selectedCount = formData[step.field]?.length || 0}
			{@const remaining = step.min ? Math.max(0, step.min - selectedCount) : 0}
			{@const showCounter = step.min && step.min > 1 && remaining > 0}
			{@const hideDefaultLabel = !step.label && remaining === 0}
			<p class="field-label" class:invisible={hideDefaultLabel}>
				{#if step.label && step.label !== ''}
					{step.label}{#if showCounter} • {remaining}{/if}
				{:else if step.min}
					Seleziona almeno {remaining || step.min}
				{/if}
			</p>
			<!-- svelte-ignore a11y_no_static_element_interactions -->
			<div on:keydown={handleEnterKey}>
				<GridSelector
					options={step.options}
					selected={isMultiple ? formData[step.field] : formData[step.field]}
					{isMultiple}
					disabled={loading}
					onToggle={(option) => onToggleOption(step.field, option, isMultiple)}
				/>
			</div>
			{#key step.field}
				{@const hasInputOption = isMultiple
					? formData[step.field]?.some(v => v.includes(':input'))
					: formData[step.field]?.includes(':input')}
				{#if hasInputOption}
					<div class="custom-input-container" transition:slide={{ duration: 300 }}>
						<input
							type="text"
							class="custom-input"
							bind:value={formData[step.field + '_custom']}
							placeholder="Specifica..."
							on:keydown={handleEnterKey}
							disabled={loading}
						/>
					</div>
				{/if}
			{/key}
		{/if}
	{/if}

	{#if step.type !== 'start'}
		<div class="error-slot">
			<ErrorMessage {error} />
		</div>
	{/if}
</div>

<style>
	.step-content {
		width: 100%;
		max-width: 32rem;
		background: #fff;
		padding: clamp(1.5rem, 4vw, 2.5rem);
		position: relative;
	}

	.step-content.is-start {
		text-align: center;
		max-width: 28rem;
	}

	.error-slot {
		height: 5.25rem;
		box-sizing: border-box;
		overflow: hidden;
		padding-top: 1rem;
	}

	.error-slot :global(.error) {
		margin: 0;
		display: -webkit-box;
		-webkit-line-clamp: 2;
		-webkit-box-orient: vertical;
		overflow: hidden;
		line-height: 1.3;
	}

	h1 {
		margin: 0 0 clamp(1.5rem, 4vw, 2rem) 0;
		font-size: clamp(1.5rem, 5vw, 2rem);
		text-align: center;
		font-weight: bold;
		color: #000;
	}

	label,
	.field-label {
		display: block;
		margin-bottom: 0.75rem;
		font-weight: 600;
		color: #000;
		font-size: clamp(0.875rem, 2.5vw, 1rem);
	}

	.field-label.invisible {
		visibility: hidden;
	}

	textarea {
		width: 100%;
		padding: clamp(0.75rem, 2vw, 1rem);
		border: 2px solid #000;
		font-size: max(16px, 1rem);
		font-family: inherit;
		resize: vertical;
		min-height: 12rem;
	}

	textarea:focus {
		outline: 2px solid #000;
		outline-offset: -2px;
	}

	.success-check {
		display: flex;
		justify-content: center;
		margin-bottom: 1.5rem;
		color: #22c55e;
		animation: scaleIn 0.3s ease;
	}

	@keyframes scaleIn {
		from {
			transform: scale(0);
		}
		to {
			transform: scale(1);
		}
	}

	.file-label {
		display: block;
		margin-bottom: 1.5rem;
		font-weight: 600;
		color: #000;
		font-size: clamp(0.875rem, 2.5vw, 1rem);
	}

	.file-button {
		width: 100%;
		padding: clamp(1rem, 3vw, 1.25rem);
		background: #000;
		color: #fff;
		border: 2px solid #000;
		font-size: clamp(0.875rem, 2.5vw, 1rem);
		font-weight: 600;
		font-family: inherit;
		cursor: pointer;
		transition: all 0.2s ease;
	}

	.file-button:hover:not(:disabled) {
		background: #333;
	}

	.file-button:disabled {
		background: #ccc;
		border-color: #ccc;
		color: #666;
		cursor: not-allowed;
	}

	.file-name {
		font-size: 0.875rem;
		color: #666;
		margin-top: 0.75rem;
		text-align: center;
	}

	input[type="text"] {
		width: 100%;
		padding: clamp(0.75rem, 2vw, 1rem);
		border: 2px solid #000;
		font-size: max(16px, 1rem);
		font-family: inherit;
	}

	input[type="text"]:focus {
		outline: 2px solid #000;
		outline-offset: -2px;
	}

	.custom-input-container {
		margin-top: clamp(1rem, 3vw, 1.5rem);
	}

	.custom-input {
		width: 100%;
		padding: clamp(0.75rem, 2vw, 1rem);
		border: 2px solid #000;
		font-size: max(16px, 1rem);
		font-family: inherit;
	}

	.custom-input:focus {
		outline: 2px solid #000;
		outline-offset: -2px;
	}

	.close-btn-start {
		position: fixed;
		top: clamp(1rem, 3vw, 1.5rem);
		left: clamp(1rem, 3vw, 1.5rem);
		background: transparent;
		border: none;
		padding: clamp(0.25rem, 2vw, 0.5rem);
		line-height: 1;
		cursor: pointer;
		color: #000;
		transition: color 0.2s ease;
		z-index: 1000;
	}

	.close-btn-start:hover:not(:disabled) {
		color: #666;
	}

	.close-btn-start:disabled {
		opacity: 0.3;
		cursor: not-allowed;
	}

	.start-text {
		text-align: center;
		font-size: clamp(1rem, 3vw, 1.25rem);
		color: #333;
		line-height: 1.6;
		margin: 0 0 2rem 0;
		white-space: pre-line;
	}

	.step-content.is-start h1 {
		margin-top: 0;
		font-size: clamp(1.75rem, 5vw, 2.5rem);
	}
</style>
