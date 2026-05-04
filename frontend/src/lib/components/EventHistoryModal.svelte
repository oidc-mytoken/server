<script lang="ts">
	import type { EventHistoryEntry } from '$lib/types';
	import { api, ApiClientError } from '$lib/api/client';
	import { formatDateTime } from '$lib/utils/format';
	import LoadingSpinner from './LoadingSpinner.svelte';

	let show = $state(false);
	let loading = $state(false);
	let error = $state('');
	let events: EventHistoryEntry[] = $state([]);
	let tokenName = $state('');

	export async function open(momId: string, name?: string) {
		tokenName = name || 'Token';
		show = true;
		loading = true;
		error = '';
		events = [];

		try {
			events = await api.getEventHistoryByMomId(momId);
		} catch (err) {
			if (err instanceof ApiClientError) {
				error = err.description ?? err.code;
			} else {
				error = (err as Error).message;
			}
		} finally {
			loading = false;
		}
	}

	function close() {
		show = false;
	}

	function handleKeydown(event: KeyboardEvent) {
		if (event.key === 'Escape' && show) {
			close();
		}
	}

	function getEventIcon(event: string): string {
		const eventLower = event.toLowerCase();
		if (eventLower.includes('created')) return 'fa-plus-circle text-success';
		if (eventLower.includes('revoked')) return 'fa-ban text-danger';
		if (eventLower.includes('at_created') || eventLower.includes('access_token')) return 'fa-key text-primary';
		if (eventLower.includes('rotated')) return 'fa-sync text-info';
		if (eventLower.includes('transferred')) return 'fa-exchange-alt text-warning';
		if (eventLower.includes('used')) return 'fa-check text-success';
		if (eventLower.includes('blocked') || eventLower.includes('denied')) return 'fa-times-circle text-danger';
		return 'fa-circle text-secondary';
	}

	function formatEventName(event: string): string {
		// Convert snake_case to Title Case with spaces
		return event
			.split('_')
			.map(word => word.charAt(0).toUpperCase() + word.slice(1).toLowerCase())
			.join(' ');
	}

	function parseUserAgent(userAgent: string | undefined): { icon: string; title: string } {
		if (!userAgent) return { icon: 'fas fa-question', title: 'Unknown' };
		
		const ua = userAgent.toLowerCase();
		
		// Check for known clients
		if (ua.includes('oidc-agent')) {
			return { icon: 'fas fa-terminal', title: 'oidc-agent' };
		}
		if (ua.includes('mytoken')) {
			return { icon: 'fas fa-key', title: 'mytoken client' };
		}
		
		// Browser detection
		if (ua.includes('firefox')) {
			return { icon: 'fab fa-firefox', title: 'Firefox' };
		}
		if (ua.includes('chrome') && !ua.includes('edg')) {
			return { icon: 'fab fa-chrome', title: 'Chrome' };
		}
		if (ua.includes('safari') && !ua.includes('chrome')) {
			return { icon: 'fab fa-safari', title: 'Safari' };
		}
		if (ua.includes('edg')) {
			return { icon: 'fab fa-edge', title: 'Edge' };
		}
		
		// OS detection fallback
		if (ua.includes('linux')) {
			return { icon: 'fab fa-linux', title: 'Linux Client' };
		}
		if (ua.includes('windows')) {
			return { icon: 'fab fa-windows', title: 'Windows Client' };
		}
		if (ua.includes('mac')) {
			return { icon: 'fab fa-apple', title: 'macOS Client' };
		}
		
		return { icon: 'fas fa-globe', title: userAgent };
	}
</script>

<svelte:window onkeydown={handleKeydown} />

{#if show}
	<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
	<div
		class="modal fade show d-block"
		tabindex="-1"
		role="dialog"
		aria-modal="true"
		onclick={(e) => e.target === e.currentTarget && close()}
		onkeydown={handleKeydown}
	>
		<div class="modal-dialog modal-dialog-centered modal-xl">
			<div class="modal-content">
				<div class="modal-header">
					<h5 class="modal-title">
						<i class="fas fa-history me-2"></i>
						Event History: {tokenName}
					</h5>
					<button
						type="button"
						class="btn-close"
						aria-label="Close"
						onclick={close}
					></button>
				</div>
				<div class="modal-body">
					{#if loading}
						<LoadingSpinner message="Loading event history..." />
					{:else if error}
						<div class="alert alert-danger">
							<i class="fas fa-exclamation-triangle me-2"></i>
							{error}
						</div>
					{:else if events.length === 0}
						<div class="text-center py-4 text-muted">
							<i class="fas fa-inbox fa-3x mb-3"></i>
							<p>No events recorded for this token</p>
						</div>
					{:else}
						<div class="table-responsive">
							<table class="table table-hover align-middle">
								<thead>
									<tr>
										<th>Event</th>
										<th>Comment</th>
										<th>Time</th>
										<th>IP</th>
										<th class="text-center">Client</th>
									</tr>
								</thead>
								<tbody>
									{#each events as event}
										{@const uaInfo = parseUserAgent(event.user_agent)}
										<tr>
											<td>
												<i class="fas {getEventIcon(event.event)} me-2"></i>
												{formatEventName(event.event)}
											</td>
											<td class="text-break">
												{event.comment || '-'}
											</td>
											<td class="text-nowrap">
												{formatDateTime(event.time)}
											</td>
											<td>
												<code>{event.ip || '-'}</code>
											</td>
											<td class="text-center">
												<i 
													class="{uaInfo.icon}" 
													title={event.user_agent || 'Unknown'}
												></i>
											</td>
										</tr>
									{/each}
								</tbody>
							</table>
						</div>
					{/if}
				</div>
				<div class="modal-footer">
					<button type="button" class="btn btn-secondary" onclick={close}>
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

	.table th {
		border-top: none;
		font-weight: 500;
	}

	.text-break {
		word-break: break-word;
	}

	td code {
		font-size: 0.85em;
	}
</style>
