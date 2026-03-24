<script lang="ts">
	import { auth, isLoggedIn } from '$lib/stores/auth';
	import { providers } from '$lib/stores/discovery';
	import { api, ApiClientError } from '$lib/api/client';
	import { ui } from '$lib/stores/ui';

	export let instanceUrl: string = '';
	export let empty: boolean = false;

	let loggingIn = false;

	async function handleLogin(issuer: string) {
		if (loggingIn) return;
		loggingIn = true;

		try {
			const response = await api.login(issuer);

			// Get the redirect URL (consent_uri or authorization_uri)
			const redirectUrl = response.consent_uri ?? response.authorization_uri;

			if (redirectUrl) {
				// Redirect to the authorization URL (same window, like Mustache version)
				window.location.href = redirectUrl;
			} else {
				ui.showError('Login Error', 'No authorization URL returned');
				loggingIn = false;
			}
		} catch (error) {
			if (error instanceof ApiClientError) {
				ui.showError('Login Failed', error.description ?? error.code);
			} else {
				ui.showError('Login Failed', (error as Error).message);
			}
			loggingIn = false;
		}
	}

	async function handleLogout() {
		try {
			// Revoke the session token on the server before clearing local state
			await api.revokeSession(true);
		} catch (error) {
			// Log but don't block logout - we still want to clear local state
			console.warn('Failed to revoke session:', error);
		}
		auth.logout();
		window.location.href = '/';
	}
</script>

<nav class="navbar navbar-expand-lg navbar-dark bg-primary">
	<div class="container-fluid">
		<a class="navbar-brand" href="/">
			<img
				src="{instanceUrl}/static/img/mytoken.png"
				alt="mytoken"
				height="30"
				class="d-inline-block align-top me-2"
			/>
			mytoken
		</a>

		{#if !empty}
			<button
				class="navbar-toggler"
				type="button"
				data-bs-toggle="collapse"
				data-bs-target="#navbarNav"
				aria-controls="navbarNav"
				aria-expanded="false"
				aria-label="Toggle navigation"
			>
				<span class="navbar-toggler-icon"></span>
			</button>

			<div class="collapse navbar-collapse" id="navbarNav">
				<ul class="navbar-nav me-auto">
					<li class="nav-item">
						<a class="nav-link" href="/">Home</a>
					</li>
					{#if $isLoggedIn}
						<li class="nav-item">
							<a class="nav-link" href="/settings">Settings</a>
						</li>
					{/if}
					<li class="nav-item">
						<a class="nav-link" href="/privacy">Privacy</a>
					</li>
				</ul>

				<ul class="navbar-nav">
					{#if $isLoggedIn}
						<li class="nav-item">
							<button class="btn btn-outline-light" on:click={handleLogout}>
								<i class="fas fa-sign-out-alt me-1"></i>
								Logout
							</button>
						</li>
					{:else}
						<li class="nav-item dropdown">
							<button
								class="btn btn-outline-light dropdown-toggle"
								type="button"
								id="loginDropdown"
								data-bs-toggle="dropdown"
								aria-expanded="false"
								disabled={loggingIn}
							>
								{#if loggingIn}
									<span class="spinner-border spinner-border-sm me-1"></span>
								{:else}
									<i class="fas fa-sign-in-alt me-1"></i>
								{/if}
								Login
							</button>
							<ul class="dropdown-menu dropdown-menu-end" aria-labelledby="loginDropdown">
								{#each $providers as provider}
									<li>
										<button
											class="dropdown-item"
											on:click={() => handleLogin(provider.issuer)}
											disabled={loggingIn}
										>
											{provider.name ?? provider.issuer}
										</button>
									</li>
								{:else}
									<li>
										<span class="dropdown-item text-muted">No providers available</span>
									</li>
								{/each}
							</ul>
						</li>
					{/if}
				</ul>
			</div>
		{/if}
	</div>
</nav>
