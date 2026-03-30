<script lang="ts">
	import { ui } from '$lib/stores/ui';

	$: show = $ui.confirmModal.show;
	$: title = $ui.confirmModal.title;
	$: message = $ui.confirmModal.message;
	$: confirmText = $ui.confirmModal.confirmText;
	$: cancelText = $ui.confirmModal.cancelText;
	$: confirmVariant = $ui.confirmModal.confirmVariant;

	function handleConfirm() {
		ui.resolveConfirm(true);
	}

	function handleCancel() {
		ui.resolveConfirm(false);
	}

	function handleKeydown(event: KeyboardEvent) {
		if (!show) return;
		if (event.key === 'Escape') {
			handleCancel();
		} else if (event.key === 'Enter') {
			handleConfirm();
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
		on:click|self={handleCancel}
		on:keydown={handleKeydown}
	>
		<div class="modal-dialog modal-dialog-centered">
			<div class="modal-content">
				<div class="modal-header">
					<h5 class="modal-title">
						{#if confirmVariant === 'danger'}
							<i class="fas fa-exclamation-triangle text-danger me-2"></i>
						{:else if confirmVariant === 'warning'}
							<i class="fas fa-exclamation-circle text-warning me-2"></i>
						{:else}
							<i class="fas fa-question-circle text-primary me-2"></i>
						{/if}
						{title}
					</h5>
					<button
						type="button"
						class="btn-close"
						aria-label="Close"
						on:click={handleCancel}
					></button>
				</div>
				<div class="modal-body">
					<p class="mb-0">{message}</p>
				</div>
				<div class="modal-footer">
					<button type="button" class="btn btn-secondary" on:click={handleCancel}>
						{cancelText}
					</button>
					<button 
						type="button" 
						class="btn btn-{confirmVariant}" 
						on:click={handleConfirm}
					>
						{confirmText}
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
</style>
