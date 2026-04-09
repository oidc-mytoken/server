import {writable} from 'svelte/store';

interface ErrorModal {
	show: boolean;
	title: string;
	message: string;
	details?: string;
}

interface ConfirmModal {
	show: boolean;
	title: string;
	message: string;
	confirmText: string;
	cancelText: string;
	confirmVariant: 'danger' | 'primary' | 'warning';
	resolve: ((value: boolean) => void) | null;
	showCheckbox?: boolean;
	checkboxLabel?: string;
	checkboxChecked?: boolean;
}

interface ToastNotification {
	id: string;
	type: 'success' | 'error' | 'warning' | 'info';
	message: string;
	duration?: number;
}

interface UIState {
	errorModal: ErrorModal;
	confirmModal: ConfirmModal;
	toasts: ToastNotification[];
	loading: boolean;
}

function createUIStore() {
    const {subscribe, update} = writable<UIState>({
		errorModal: {
			show: false,
			title: '',
			message: ''
		},
		confirmModal: {
			show: false,
			title: '',
			message: '',
			confirmText: 'Confirm',
			cancelText: 'Cancel',
			confirmVariant: 'danger',
			resolve: null
		},
		toasts: [],
		loading: false
	});

	return {
		subscribe,

		/**
		 * Show an error modal
		 */
		showError(title: string, message: string, details?: string) {
			update((state) => ({
				...state,
				errorModal: {
					show: true,
					title,
					message,
					details
				}
			}));
		},

		/**
		 * Hide the error modal
		 */
		hideError() {
			update((state) => ({
				...state,
				errorModal: {
					...state.errorModal,
					show: false
				}
			}));
		},

		/**
		 * Show a toast notification
		 */
		showToast(
			type: ToastNotification['type'],
			message: string,
			duration = 5000
		) {
			const id = crypto.randomUUID();
			const toast: ToastNotification = { id, type, message, duration };

			update((state) => ({
				...state,
				toasts: [...state.toasts, toast]
			}));

			// Auto-remove after duration
			if (duration > 0) {
				setTimeout(() => {
					this.removeToast(id);
				}, duration);
			}

			return id;
		},

		/**
		 * Remove a toast by ID
		 */
		removeToast(id: string) {
			update((state) => ({
				...state,
				toasts: state.toasts.filter((t) => t.id !== id)
			}));
		},

		/**
		 * Set global loading state
		 */
		setLoading(loading: boolean) {
			update((state) => ({ ...state, loading }));
		},

		/**
		 * Show a confirmation modal and return a promise that resolves to true/false
		 */
		confirm(options: {
			title?: string;
			message: string;
			confirmText?: string;
			cancelText?: string;
			confirmVariant?: 'danger' | 'primary' | 'warning';
			showCheckbox?: boolean;
			checkboxLabel?: string;
			checkboxChecked?: boolean;
		}): Promise<boolean> {
			return new Promise((resolve) => {
				update((state) => ({
					...state,
					confirmModal: {
						show: true,
						title: options.title ?? 'Confirm',
						message: options.message,
						confirmText: options.confirmText ?? 'Confirm',
						cancelText: options.cancelText ?? 'Cancel',
						confirmVariant: options.confirmVariant ?? 'danger',
						showCheckbox: options.showCheckbox ?? false,
						checkboxLabel: options.checkboxLabel ?? '',
						checkboxChecked: options.checkboxChecked ?? false,
						resolve
					}
				}));
			});
		},

		/**
		 * Resolve the confirmation modal (called by ConfirmModal component)
		 */
		resolveConfirm(result: boolean) {
			update((state) => {
				if (state.confirmModal.resolve) {
					state.confirmModal.resolve(result);
				}
				return {
					...state,
					confirmModal: {
						...state.confirmModal,
						show: false,
						resolve: null
					}
				};
			});
		},

		/**
		 * Set checkbox state in confirm modal
		 */
		setCheckboxState(checked: boolean) {
			update((state) => ({
				...state,
				confirmModal: {
					...state.confirmModal,
					checkboxChecked: checked
				}
			}));
		},

		/**
		 * Convenience methods for toasts
		 */
		success(message: string) {
			return this.showToast('success', message);
		},

		error(message: string) {
			return this.showToast('error', message);
		},

		warning(message: string) {
			return this.showToast('warning', message);
		},

		info(message: string) {
			return this.showToast('info', message);
		}
	};
}

export const ui = createUIStore();
