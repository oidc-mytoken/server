<script lang="ts">
	import { onMount } from 'svelte';
	import { get } from 'svelte/store';
	import { discovery } from '$lib/stores/discovery';
	import { auth, isLoggedIn } from '$lib/stores/auth';
	import { tags } from '$lib/stores/tags';
	import Navbar from '$lib/components/Navbar.svelte';
	import Footer from '$lib/components/Footer.svelte';
	import ErrorModal from '$lib/components/ErrorModal.svelte';
	import ConfirmModal from '$lib/components/ConfirmModal.svelte';
	import ToastContainer from '$lib/components/ToastContainer.svelte';

	// Import Bootstrap CSS
	import 'bootstrap/dist/css/bootstrap.min.css';

	let bootstrapLoaded = false;

	onMount(async () => {
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
	:global(body) {
		background-color: #f8f9fa;
	}

	.app {
		min-height: 100vh;
	}

	/* Primary color: mytoken orange from the original theme */
	:global(.btn-primary) {
		background-color: #df691a;
		border-color: #df691a;
	}

	:global(.btn-primary:hover),
	:global(.btn-primary:focus),
	:global(.btn-primary:active) {
		background-color: #c73500;
		border-color: #c73500;
	}

	:global(.btn-outline-primary) {
		color: #df691a;
		border-color: #df691a;
	}

	:global(.btn-outline-primary:hover),
	:global(.btn-outline-primary:focus),
	:global(.btn-outline-primary:active) {
		background-color: #df691a;
		border-color: #df691a;
		color: #fff;
	}

	:global(.bg-primary) {
		background-color: #df691a !important;
	}

	:global(.text-primary) {
		color: #df691a !important;
	}

	:global(.navbar-brand img) {
		filter: brightness(0) invert(1);
	}
</style>
