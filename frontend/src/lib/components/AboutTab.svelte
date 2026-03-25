<script lang="ts">
	import { isLoggedIn } from '$lib/stores/auth';
	import { discovery } from '$lib/stores/discovery';
	import { createEventDispatcher } from 'svelte';

	const dispatch = createEventDispatcher<{ navigate: { tab: string } }>();

	// Check if notifications are enabled
	$: notificationsEnabled = !!$discovery.data?.notifications_endpoint;

	function navigateTo(tab: string) {
		dispatch('navigate', { tab });
	}
</script>

<div class="row">
	<div class="col-md-6 mb-4 mb-md-0">
		<div class="card h-100">
			<div class="card-body">
				<h4 class="text-primary mb-3">
					<i class="fas fa-info-circle me-2"></i>
					The Mytoken Service
				</h4>
				<p>
					Mytoken is a service to obtain OpenID Connect Access Tokens in an easy but secure way
					for extended periods of time and across multiple devices.
				</p>
				<p>
					To do so, users can create mytokens with exactly the properties they need for the job.
					These mytokens can easily be used (from multiple devices) to obtain OIDC access tokens.
					Mytokens and access tokens can be obtained from this web interface or the command line.
				</p>
				<p class="mb-0">
					For more details please refer to the 
					<a href="https://mytoken-docs.data.kit.edu" target="_blank" rel="noopener noreferrer">
						<i class="fas fa-external-link-alt me-1"></i>
						full documentation
					</a>.
				</p>
			</div>
		</div>
	</div>
	<div class="col-md-6">
		<div class="card h-100">
			<div class="card-body">
				<h4 class="text-primary mb-3">
					<i class="fas fa-globe me-2"></i>
					Mytoken Web
				</h4>
				<p>On this web interface you can:</p>
				<ul class="feature-list">
					<li>
						<button type="button" class="btn btn-link p-0" on:click={() => navigateTo('mt')}>
							<i class="fas fa-plus-circle me-1"></i>
							Create a mytoken
						</button>
					</li>
					<li>
						<button type="button" class="btn btn-link p-0" on:click={() => navigateTo('info')}>
							<i class="fas fa-info-circle me-1"></i>
							Get information about a mytoken
						</button>
					</li>
					<li>
						<button type="button" class="btn btn-link p-0" on:click={() => navigateTo('transfer')}>
							<i class="fas fa-exchange-alt me-1"></i>
							Exchange a transfer code into a mytoken
						</button>
					</li>
					<li>
						<button type="button" class="btn btn-link p-0" on:click={() => navigateTo('at')}>
							<i class="fas fa-key me-1"></i>
							Obtain access tokens
						</button>
						{#if !$isLoggedIn}
							<small class="text-muted d-block ms-4">requires a mytoken</small>
						{/if}
					</li>
				</ul>

				{#if !$isLoggedIn}
					<p class="mt-3 mb-2">After sign in you additionally can:</p>
					<ul class="feature-list text-muted">
						<li>
							<i class="fas fa-list me-1"></i>
							List your mytokens and revoke them
						</li>
						{#if notificationsEnabled}
							<li>
								<i class="fas fa-bell me-1"></i>
								Manage notifications
							</li>
						{/if}
						<li>
							<i class="fas fa-cog me-1"></i>
							Change settings
						</li>
					</ul>
				{:else}
					<ul class="feature-list">
						<li>
							<button type="button" class="btn btn-link p-0" on:click={() => navigateTo('list')}>
								<i class="fas fa-list me-1"></i>
								List your mytokens and revoke them
							</button>
						</li>
						{#if notificationsEnabled}
							<li>
								<button type="button" class="btn btn-link p-0" on:click={() => navigateTo('notifications')}>
									<i class="fas fa-bell me-1"></i>
									Manage notifications
								</button>
							</li>
						{/if}
						<li>
							<a href="/settings" class="btn btn-link p-0">
								<i class="fas fa-cog me-1"></i>
								Change settings
							</a>
						</li>
					</ul>
				{/if}
			</div>
		</div>
	</div>
</div>

<style>
	.feature-list {
		list-style: none;
		padding-left: 0;
		margin-bottom: 0;
	}

	.feature-list li {
		padding: 0.35rem 0;
	}

	.feature-list .btn-link {
		text-decoration: none;
		color: var(--mytoken-primary);
		text-align: left;
	}

	.feature-list .btn-link:hover {
		text-decoration: underline;
		color: var(--mytoken-primary-hover);
	}

	.card {
		border: 1px solid var(--bs-border-color);
	}
</style>
