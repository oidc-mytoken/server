<script lang="ts">
	import { onMount } from 'svelte';

	let calendarLoaded = false;

	onMount(async () => {
		// FullCalendar would be loaded here
		calendarLoaded = true;
	});
</script>

<svelte:head>
	<title>mytoken - Calendar</title>
</svelte:head>

<div class="calendar-page">
	<h2 class="mb-4">
		<i class="fas fa-calendar-alt me-2"></i>
		Token Calendar
	</h2>

	<div class="card">
		<div class="card-body">
			{#if calendarLoaded}
				<p class="text-muted mb-4">
					View token events and expirations in calendar format.
				</p>
				<!-- FullCalendar component will go here -->
				<div id="calendar" class="bg-light p-4 rounded text-center">
					<p class="text-info mb-0">
						<i class="fas fa-info-circle me-1"></i>
						Calendar component implementation in progress...
					</p>
				</div>
			{:else}
				<div class="text-center py-4">
					<div class="spinner-border text-primary" role="status">
						<span class="visually-hidden">Loading...</span>
					</div>
					<p class="mt-2 text-muted">Loading calendar...</p>
				</div>
			{/if}
		</div>
	</div>
</div>

<style>
	.card {
		border: none;
		box-shadow: 0 0.125rem 0.25rem rgba(0, 0, 0, 0.075);
	}

	#calendar {
		min-height: 400px;
	}
</style>
