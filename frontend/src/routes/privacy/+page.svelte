<script lang="ts">
</script>

<svelte:head>
	<title>mytoken - Privacy Policy</title>
</svelte:head>

<div class="privacy-page">
	<h1 class="mb-4">Privacy Policy</h1>

	<div class="card">
		<div class="card-body">
			<h2>Data Collection</h2>
			<p>
				The mytoken service collects and processes the following data:
			</p>
			<ul>
				<li>OIDC subject identifier from your identity provider</li>
				<li>Token metadata (creation time, capabilities, restrictions)</li>
				<li>IP addresses for security logging</li>
				<li>Email address (if you enable notifications)</li>
			</ul>

			<h2>Data Usage</h2>
			<p>
				Your data is used exclusively for:
			</p>
			<ul>
				<li>Providing the mytoken service functionality</li>
				<li>Security monitoring and abuse prevention</li>
				<li>Sending notifications you have subscribed to</li>
			</ul>

			<h2>Data Retention</h2>
			<p>
				Token data is retained until:
			</p>
			<ul>
				<li>The token expires</li>
				<li>You revoke the token</li>
				<li>You delete your account</li>
			</ul>

			<h2>Contact</h2>
			<p>
				For privacy-related inquiries, please contact the service administrator.
			</p>
		</div>
	</div>
</div>

<style>
	.card {
		border: none;
		box-shadow: 0 0.125rem 0.25rem rgba(0, 0, 0, 0.075);
	}

	h2 {
		font-size: 1.5rem;
		margin-top: 1.5rem;
		margin-bottom: 1rem;
	}

	h2:first-child {
		margin-top: 0;
	}
</style>
