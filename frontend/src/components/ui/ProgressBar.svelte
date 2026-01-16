<script>
	export let percentage = 0;
</script>

<div class="progress-bar">
	<div class="progress-fill" style="width: {percentage}%"></div>
</div>

<style>
	.progress-bar {
		position: fixed;
		top: 0;
		left: 0;
		right: 0;
		height: 3px;
		background: #f0f0f0;
		z-index: 101;
	}

	.progress-fill {
		height: 100%;
		background: #000;
		transition: width 0.3s ease;
	}
</style>
