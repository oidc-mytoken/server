<script lang="ts">
	import { onMount, onDestroy, createEventDispatcher } from 'svelte';
	import flatpickr from 'flatpickr';
	import type { Instance } from 'flatpickr/dist/types/instance';
	import 'flatpickr/dist/flatpickr.min.css';

	export let value: string = ''; // Format: "YYYY-MM-DDTHH:MM" or empty
	export let id: string = '';
	export let placeholder: string = 'Select date and time';
	export let disabled: boolean = false;
	export let enableTime: boolean = true;
	export let dateFormat: string = 'Y-m-d H:i';
	export let altFormat: string = 'F j, Y H:i'; // Human-readable format
	export let minDate: string | Date | undefined = undefined;
	export let maxDate: string | Date | undefined = undefined;

	const dispatch = createEventDispatcher<{ change: string }>();

	let inputElement: HTMLInputElement;
	let fp: Instance | null = null;

	onMount(() => {
		fp = flatpickr(inputElement, {
			enableTime,
			dateFormat,
			altInput: true,
			altFormat,
			time_24hr: true,
			defaultDate: value || undefined,
			minDate,
			maxDate,
			onChange: (selectedDates, dateStr) => {
				value = dateStr;
				dispatch('change', dateStr);
			}
		});
	});

	onDestroy(() => {
		if (fp) {
			fp.destroy();
		}
	});

	// Update flatpickr when value changes externally
	$: if (fp && value !== undefined) {
		const currentDate = fp.selectedDates[0];
		const currentStr = currentDate ? fp.formatDate(currentDate, dateFormat) : '';
		if (currentStr !== value) {
			fp.setDate(value || '', false);
		}
	}

	// Update options when props change
	$: if (fp) {
		fp.set('minDate', minDate);
		fp.set('maxDate', maxDate);
	}
</script>

<input
	bind:this={inputElement}
	{id}
	type="text"
	class="form-control"
	{placeholder}
	{disabled}
/>

<style>
	/* Ensure the input looks consistent */
	:global(.flatpickr-input) {
		background-color: #fff !important;
	}

	:global(.flatpickr-input[readonly]) {
		cursor: pointer;
	}

	/* Style the calendar to match Bootstrap theme */
	:global(.flatpickr-calendar) {
		font-family: inherit;
		box-shadow: 0 0.5rem 1rem rgba(0, 0, 0, 0.15);
		border-radius: 0.375rem;
		background: var(--bs-body-bg);
		color: var(--bs-body-color);
	}

	:global(.flatpickr-day.selected),
	:global(.flatpickr-day.selected:hover) {
		background: var(--mytoken-primary);
		border-color: var(--mytoken-primary);
	}

	:global(.flatpickr-day:hover) {
		background: var(--bs-tertiary-bg);
		border-color: var(--bs-border-color);
	}

	:global(.flatpickr-time input:hover),
	:global(.flatpickr-time input:focus) {
		background: var(--bs-tertiary-bg);
	}

	/* Dark mode specific styles for flatpickr */
	:global([data-bs-theme='dark'] .flatpickr-calendar) {
		background: var(--bs-body-bg);
		border-color: var(--bs-border-color);
	}

	:global([data-bs-theme='dark'] .flatpickr-months .flatpickr-month),
	:global([data-bs-theme='dark'] .flatpickr-current-month .flatpickr-monthDropdown-months),
	:global([data-bs-theme='dark'] .flatpickr-weekdays),
	:global([data-bs-theme='dark'] span.flatpickr-weekday) {
		background: var(--bs-body-bg);
		color: var(--bs-body-color);
	}

	:global([data-bs-theme='dark'] .flatpickr-day) {
		color: var(--bs-body-color);
	}

	:global([data-bs-theme='dark'] .flatpickr-day.prevMonthDay),
	:global([data-bs-theme='dark'] .flatpickr-day.nextMonthDay) {
		color: var(--bs-secondary-color);
	}

	:global([data-bs-theme='dark'] .flatpickr-day.today) {
		border-color: var(--mytoken-primary);
	}

	:global([data-bs-theme='dark'] .numInputWrapper span) {
		border-color: var(--bs-border-color);
	}

	:global([data-bs-theme='dark'] .numInputWrapper span:hover) {
		background: var(--bs-tertiary-bg);
	}

	:global([data-bs-theme='dark'] .flatpickr-time input) {
		color: var(--bs-body-color);
	}

	:global([data-bs-theme='dark'] .flatpickr-time .flatpickr-time-separator),
	:global([data-bs-theme='dark'] .flatpickr-time .flatpickr-am-pm) {
		color: var(--bs-body-color);
	}
</style>
