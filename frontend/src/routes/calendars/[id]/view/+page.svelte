<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { Calendar } from '@fullcalendar/core';
	import dayGridPlugin from '@fullcalendar/daygrid';
	import listPlugin from '@fullcalendar/list';
	import iCalendarPlugin from '@fullcalendar/icalendar';
	import bootstrap5Plugin from '@fullcalendar/bootstrap5';
	import tippy from 'tippy.js';
	import 'tippy.js/dist/tippy.css';
	import 'tippy.js/themes/translucent.css';
	import TagPill from '$lib/components/TagPill.svelte';

	interface TagInfo {
		tag: string;
		color: string;
	}

	let calendarEl: HTMLDivElement;
	let calendar: Calendar | null = null;
	let tags = $state<TagInfo[]>([]);
	let loading = $state(true);
	let error = $state('');

	let calendarId = $derived($page.params.id);

	onMount(() => {
		if (!calendarId) {
			error = 'No calendar ID provided';
			loading = false;
			return;
		}

		initCalendar();

		return () => {
			if (calendar) {
				calendar.destroy();
			}
		};
	});

	function getIcsUrl(): string {
		// Remove /view from the current path to get the ICS URL
		return window.location.href.replace(/\/view$/, '');
	}

	function generateColorFromString(str: string): string {
		let hash = 0;
		for (let i = 0; i < str.length; i++) {
			hash = str.charCodeAt(i) + ((hash << 5) - hash);
		}
		const color = Math.abs(hash).toString(16).substring(0, 6);
		return color.padEnd(6, '0');
	}

	function textClassForBackgroundColor(hexColor: string): string {
		// Convert hex to RGB
		const r = parseInt(hexColor.substring(0, 2), 16);
		const g = parseInt(hexColor.substring(2, 4), 16);
		const b = parseInt(hexColor.substring(4, 6), 16);
		// Calculate luminance
		const luminance = (0.299 * r + 0.587 * g + 0.114 * b) / 255;
		return luminance > 0.5 ? 'text-dark' : 'text-light';
	}

	function initCalendar() {
		const icsUrl = getIcsUrl();

		calendar = new Calendar(calendarEl, {
			plugins: [dayGridPlugin, listPlugin, iCalendarPlugin, bootstrap5Plugin],
			themeSystem: 'bootstrap5',
			headerToolbar: {
				left: 'prev,next today',
				center: 'title',
				right: 'dayGridMonth,listWeek,listYear'
			},
			views: {
				listWeek: { buttonText: 'list week' },
				listYear: { buttonText: 'list year' }
			},
			eventDidMount: function(info) {
				let title = info.event.extendedProps.description || '';
				title = title.replaceAll('\n', ' <br/> ');
				title = title.replace(/(https?:\/\/)\S+/g, function(matched: string) {
					return `<a href="${matched}" target="_blank" rel="noopener noreferrer">${matched}</a>`;
				});
				info.event.setProp('url', '');
				tippy(info.el, {
					placement: 'top',
					content: title,
					arrow: true,
					trigger: 'mouseenter',
					allowHTML: true,
					interactive: true,
					theme: 'translucent',
					maxWidth: 600
				});
			},
			eventClick: function(info) {
				info.jsEvent.preventDefault();
			},
			eventDisplay: 'block',
			eventTimeFormat: {
				hour: '2-digit',
				minute: '2-digit',
				second: '2-digit',
				meridiem: false,
				hour12: false
			},
			displayEventEnd: false,
			initialView: 'dayGridMonth',
			navLinks: true,
			editable: false,
			dayMaxEvents: true,
			nowIndicator: true,
			defaultTimedEventDuration: '00:00:01',
			events: {
				url: icsUrl,
				format: 'ics'
			},
			eventSourceSuccess: function(_eventsInput, response) {
				loading = false;
				const tagsHeader = response?.headers.get('X-Calendar-Tags');
				if (tagsHeader) {
					try {
						const parsedTags = JSON.parse(tagsHeader);
						if (parsedTags && parsedTags.length > 0) {
							tags = parsedTags.map((t: { tag: string; color?: string }) => ({
								tag: t.tag,
								color: t.color || generateColorFromString(t.tag)
							}));
						}
					} catch (e) {
						console.error('Failed to parse calendar tags:', e);
					}
				}
			},
			eventSourceFailure: function(err) {
				loading = false;
				error = 'Failed to load calendar events';
				console.error('Calendar load error:', err);
			}
		});

		calendar.render();
	}
</script>

<svelte:head>
	<title>mytoken - Calendar</title>
</svelte:head>

<div class="calendar-view-page">
	<!-- Header with tags -->
	<div class="d-flex align-items-center justify-content-between mb-4">
		<h2 class="mb-0">
			<i class="fas fa-calendar-alt me-2"></i>
			Calendar
			{#if tags.length > 0}
				<span class="ms-2">
					{#each tags as tagInfo}
						<TagPill tag={tagInfo.tag} color={tagInfo.color} />
					{/each}
				</span>
			{/if}
		</h2>
		<div class="btn-group">
			<a href={getIcsUrl()} class="btn btn-outline-primary" download>
				<i class="fas fa-download me-1"></i>
				Download ICS
			</a>
			<button class="btn btn-outline-secondary" onclick={() => window.history.back()}>
				<i class="fas fa-arrow-left me-1"></i>
				Back
			</button>
		</div>
	</div>

	{#if error}
		<div class="alert alert-danger">
			<i class="fas fa-exclamation-triangle me-2"></i>
			{error}
		</div>
	{/if}

	{#if loading}
		<div class="text-center py-4">
			<div class="spinner-border text-primary" role="status">
				<span class="visually-hidden">Loading...</span>
			</div>
			<p class="mt-2 text-muted">Loading calendar...</p>
		</div>
	{/if}

	<div class="card">
		<div class="card-body">
			<div bind:this={calendarEl} id="calendar"></div>
		</div>
	</div>
</div>

<style>
	.card {
		border: none;
		box-shadow: 0 0.125rem 0.25rem rgba(0, 0, 0, 0.075);
	}

	#calendar {
		min-height: 600px;
	}

	:global(.fc-theme-bootstrap5 .fc-button-primary) {
		background-color: #df691a;
		border-color: #df691a;
	}

	:global(.fc-theme-bootstrap5 .fc-button-primary:hover) {
		background-color: #c73500;
		border-color: #c73500;
	}

	:global(.fc-theme-bootstrap5 .fc-button-primary:not(:disabled).fc-button-active) {
		background-color: #c73500;
		border-color: #c73500;
	}

	:global(.fc-event) {
		cursor: pointer;
	}

	/* Tippy tooltip styling */
	:global(.tippy-box[data-theme~='translucent']) {
		background-color: rgba(0, 0, 0, 0.85);
	}

	:global(.tippy-box[data-theme~='translucent'] a) {
		color: #6ea8fe;
	}

	:global(.tippy-box[data-theme~='translucent'] a:hover) {
		color: #9ec5fe;
	}
</style>
