<script lang="ts">
	import { page } from '$app/stores';

	let application = $derived($page.url.searchParams.get('application') || '');
</script>

<svelte:head>
	<title>mytoken - Aborted</title>
</svelte:head>

<div class="native-page">
	<div class="row justify-content-center">
		<div class="col-md-8 col-lg-6">
			<div class="card shadow text-center">
				<div class="card-body py-5">
					<div class="display-1 text-warning mb-4">
						<i class="fas fa-exclamation-circle"></i>
					</div>
					<h3 class="mb-4">Aborted!</h3>
					{#if application}
						<h4 class="text-muted">Please go back to '{application}'.</h4>
					{:else}
						<h4 class="text-muted">Please go back to the application that started this flow.</h4>
					{/if}
				</div>
			</div>
		</div>
	</div>
</div>

<style>
	.native-page {
		padding-top: 3rem;
	}

	.card {
		border: none;
	}
</style>
