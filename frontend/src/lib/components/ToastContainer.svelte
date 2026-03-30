<script lang="ts">
	import { ui } from '$lib/stores/ui';

	$: toasts = $ui.toasts;

	function getIcon(type: string): string {
		switch (type) {
			case 'success':
				return 'fa-check-circle';
			case 'error':
				return 'fa-times-circle';
			case 'warning':
				return 'fa-exclamation-triangle';
			case 'info':
			default:
				return 'fa-info-circle';
		}
	}

	function getBgClass(type: string): string {
		switch (type) {
			case 'success':
				return 'bg-success';
			case 'error':
				return 'bg-danger';
			case 'warning':
				return 'bg-warning text-dark';
			case 'info':
			default:
				return 'bg-info';
		}
	}
</script>

<div class="toast-container position-fixed top-0 end-0 p-3" style="z-index: 1100;">
	{#each toasts as toast (toast.id)}
		<div
			class="toast show align-items-center text-white border-0 {getBgClass(toast.type)}"
			role="alert"
			aria-live="assertive"
			aria-atomic="true"
		>
			<div class="d-flex">
				<div class="toast-body">
					<i class="fas {getIcon(toast.type)} me-2"></i>
					{toast.message}
				</div>
				<button
					type="button"
					class="btn-close btn-close-white me-2 m-auto"
					aria-label="Close"
					on:click={() => ui.removeToast(toast.id)}
				></button>
			</div>
		</div>
	{/each}
</div>

<style>
	.toast {
		min-width: 250px;
		margin-bottom: 0.5rem;
	}
</style>
