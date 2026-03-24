<script lang="ts">
	import { ui } from '$lib/stores/ui';

	$: show = $ui.errorModal.show;
	$: title = $ui.errorModal.title;
	$: message = $ui.errorModal.message;
	$: details = $ui.errorModal.details;

	function handleClose() {
		ui.hideError();
	}

	function handleKeydown(event: KeyboardEvent) {
		if (event.key === 'Escape' && show) {
			handleClose();
		}
	}
</script>

<svelte:window on:keydown={handleKeydown} />

{#if show}
	<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
	<div
		class="modal fade show d-block"
		tabindex="-1"
		role="dialog"
		aria-modal="true"
		on:click|self={handleClose}
		on:keydown={handleKeydown}
	>
		<div class="modal-dialog modal-dialog-centered">
			<div class="modal-content">
				<div class="modal-header bg-danger text-white">
					<h5 class="modal-title">
						<i class="fas fa-exclamation-triangle me-2"></i>
						{title || 'Error'}
					</h5>
					<button
						type="button"
						class="btn-close btn-close-white"
						aria-label="Close"
						on:click={handleClose}
					></button>
				</div>
				<div class="modal-body">
					<p>{message}</p>
					{#if details}
						<details class="mt-3">
							<summary class="text-muted">Technical Details</summary>
							<pre class="mt-2 p-2 bg-light rounded"><code>{details}</code></pre>
						</details>
					{/if}
				</div>
				<div class="modal-footer">
					<button type="button" class="btn btn-secondary" on:click={handleClose}>
						Close
					</button>
				</div>
			</div>
		</div>
	</div>
	<div class="modal-backdrop fade show"></div>
{/if}

<style>
	.modal {
		background-color: rgba(0, 0, 0, 0.5);
	}

	pre {
		white-space: pre-wrap;
		word-break: break-word;
		max-height: 200px;
		overflow-y: auto;
	}
</style>
