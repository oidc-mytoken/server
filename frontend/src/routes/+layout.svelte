<script lang="ts">
	import { onMount } from 'svelte';
	import { get } from 'svelte/store';
	import { discovery } from '$lib/stores/discovery';
	import { auth, isLoggedIn } from '$lib/stores/auth';
	import { tags } from '$lib/stores/tags';
	import { theme } from '$lib/stores/theme';
	import Navbar from '$lib/components/Navbar.svelte';
	import Footer from '$lib/components/Footer.svelte';
	import ErrorModal from '$lib/components/ErrorModal.svelte';
	import ConfirmModal from '$lib/components/ConfirmModal.svelte';
	import ToastContainer from '$lib/components/ToastContainer.svelte';

	// Import Bootstrap CSS
	import 'bootstrap/dist/css/bootstrap.min.css';

	let bootstrapLoaded = false;

	onMount(async () => {
		// Initialize theme (ensures correct theme is applied)
		theme.initialize();

		// Load Bootstrap JS
		await import('bootstrap');
		bootstrapLoaded = true;

		// Initialize discovery
		const cachedLoaded = discovery.loadFromCache();
		if (!cachedLoaded) {
			await discovery.fetch();
		}

		// Load cached auth state from storage first (for quick UI)
		auth.loadFromStorage();

		// Then check if actually logged in via cookie (like Mustache frontend does)
		const discoveryData = get(discovery);
		if (discoveryData.data?.tokeninfo_endpoint) {
			await auth.checkLogin(discoveryData.data.tokeninfo_endpoint);
		}
		
		// Note: Tags are loaded lazily when needed (e.g., in CreateMytoken)
		// to avoid potential token rotation race conditions
	});
</script>

<svelte:head>
	<title>mytoken</title>
	<link
		rel="stylesheet"
		href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/5.15.4/css/all.min.css"
		integrity="sha512-1ycn6IcaQQ40/MKBW2W4Rhis/DbILU74C1vSrLJxCq57o941Ym01SwNsOMqvEBFlcgUa6xLiPY/NS5R+E6ztJQ=="
		crossorigin="anonymous"
		referrerpolicy="no-referrer"
	/>
</svelte:head>

<div class="app d-flex flex-column min-vh-100">
	<Navbar />

	<main class="container py-4 flex-grow-1">
		<slot />
	</main>

	<Footer />
</div>

<ErrorModal />
<ConfirmModal />
<ToastContainer />

<style>
	/* Define mytoken brand colors as CSS custom properties */
	:global(:root),
	:global([data-bs-theme='light']) {
		--mytoken-primary: #df691a;
		--mytoken-primary-hover: #c73500;
		--mytoken-primary-rgb: 223, 105, 26;
	}

	:global([data-bs-theme='dark']) {
		--mytoken-primary: #df691a;
		--mytoken-primary-hover: #e8823a;
		--mytoken-primary-rgb: 223, 105, 26;
	}

	:global(body) {
		background-color: var(--bs-body-bg);
	}

	.app {
		min-height: 100vh;
	}

	/* Primary color: mytoken orange from the original theme */
	:global(.btn-primary) {
		background-color: var(--mytoken-primary);
		border-color: var(--mytoken-primary);
	}

	:global(.btn-primary:hover),
	:global(.btn-primary:focus),
	:global(.btn-primary:active) {
		background-color: var(--mytoken-primary-hover);
		border-color: var(--mytoken-primary-hover);
	}

	:global(.btn-primary:disabled),
	:global(.btn-primary.disabled) {
		background-color: var(--mytoken-primary);
		border-color: var(--mytoken-primary);
		opacity: 0.65;
	}

	:global(.btn-outline-primary) {
		color: var(--mytoken-primary);
		border-color: var(--mytoken-primary);
	}

	:global(.btn-outline-primary:hover),
	:global(.btn-outline-primary:focus),
	:global(.btn-outline-primary:active) {
		background-color: var(--mytoken-primary);
		border-color: var(--mytoken-primary);
		color: #fff;
	}

	:global(.btn-outline-primary:disabled),
	:global(.btn-outline-primary.disabled) {
		color: var(--mytoken-primary);
		border-color: var(--mytoken-primary);
		opacity: 0.65;
	}

	/* Radio button group with btn-check (used for toggle buttons) */
	:global(.btn-check:checked + .btn-outline-primary) {
		background-color: var(--mytoken-primary);
		border-color: var(--mytoken-primary);
		color: #fff;
	}

	:global(.btn-check:focus + .btn-outline-primary),
	:global(.btn-check:active + .btn-outline-primary) {
		box-shadow: 0 0 0 0.25rem rgba(var(--mytoken-primary-rgb), 0.5);
	}

	:global(.bg-primary) {
		background-color: var(--mytoken-primary) !important;
	}

	:global(.text-primary) {
		color: var(--mytoken-primary) !important;
	}

	:global(.navbar-brand img) {
		filter: brightness(0) invert(1);
	}

	/* Form control focus states with brand color */
	:global(.form-control:focus),
	:global(.form-select:focus) {
		border-color: var(--mytoken-primary);
		box-shadow: 0 0 0 0.25rem rgba(var(--mytoken-primary-rgb), 0.25);
	}

	/* Links with brand color */
	:global(a) {
		color: var(--mytoken-primary);
	}

	:global(a:hover) {
		color: var(--mytoken-primary-hover);
	}

	/* Ensure nav links in navbar stay white */
	:global(.navbar-dark .nav-link) {
		color: rgba(255, 255, 255, 0.85);
	}

	:global(.navbar-dark .nav-link:hover),
	:global(.navbar-dark .nav-link:focus) {
		color: #fff;
	}
</style>
