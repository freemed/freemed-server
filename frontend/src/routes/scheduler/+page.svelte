<script lang="ts">
	import { goto } from '$app/navigation';
	import { api } from '$lib/api';
	import { toast } from '$lib/stores/toast.svelte';
	import Calendar from '$lib/components/Calendar.svelte';
	import type { Calendar as FullCalendar } from '@fullcalendar/core';

	// --- Types ---

	interface SchedulerEvent {
		scheduler_id: number;
		date_of: string;
		hour: number;
		minute: number;
		duration: number;
		patient: string;
		provider: string | null;
		note: string;
		template_color: string | null;
		status: string;
		status_color: string | null;
	}

	interface EventDetail {
		date_of: string;
		date_of_mdy: string;
		hour: number;
		minute: number;
		appointment_time: string;
		duration: number;
		provider: string;
		provider_id: number;
		resource_type: string;
		patient: string;
		patient_id: number;
		note: string;
		status: string;
		status_color: string | null;
		scheduler_id: number;
		appointment_template_id: number;
		template_color: string;
	}

	// New appointment form
	interface NewApptForm {
		patient: number;
		provider: number;
		date: string;
		hour: number;
		minute: number;
		duration: number;
		note: string;
	}

	// --- State ---

	let calendarRef = $state<FullCalendar | null>(null);
	let selectedEvent = $state<EventDetail | null>(null);
	let showModal = $state(false);
	let modalLoading = $state(false);

	// New appointment state
	let showNewForm = $state(false);
	let newFormSubmitting = $state(false);
	let newForm = $state<NewApptForm>({
		patient: 0,
		provider: 0,
		date: '',
		hour: 9,
		minute: 0,
		duration: 30,
		note: '',
	});

	// Provider filter
	let selectedProvider = $state(0);

	// Group appointment state
	let showNewGroup = $state(false);
	let groupPatientIds = $state<number[]>([]);
	let groupFormSubmitting = $state(false);
	let groupForm = $state({ date: '', hour: 9, minute: 0, duration: 30, note: '' });

	// Copy appointment state
	let copyDate = $state('');
	let copyHour = $state(9);
	let copyMinute = $state(0);
	let copySubmitting = $state(false);

	// --- Helpers ---

	function dateToJSONLocal(date: Date): string {
		const local = new Date(date);
		local.setMinutes(date.getMinutes() - date.getTimezoneOffset());
		return local.toJSON().slice(0, 10);
	}

	// Initialize form date after helpers are defined
	$effect(() => {
		newForm.date = dateToJSONLocal(new Date());
	});

	function buildTitle(patient: string, provider: string | null, note: string): string {
		let title = patient;
		if (provider) title += ` (${provider})`;
		if (note) title += ` [${note}]`;
		return title;
	}

	function pad(n: number): string {
		return n.toString().padStart(2, '0');
	}

	function refreshCalendar() {
		calendarRef?.refetchEvents();
	}

	// --- New Appointment ---

	async function createAppointment() {
		if (newForm.patient <= 0) {
			toast.error('Patient ID is required.');
			return;
		}
		newFormSubmitting = true;
		try {
			await api.post('/scheduler', {
				date: newForm.date,
				hour: newForm.hour,
				minute: newForm.minute,
				duration: newForm.duration,
				type: 'patient',
				provider: newForm.provider,
				patient: newForm.patient,
				note: newForm.note
			});
			toast.success('Appointment created.');
			showNewForm = false;
			resetNewForm();
			refreshCalendar();
		} catch (e: any) {
			toast.error(e.message || 'Failed to create appointment.');
		} finally {
			newFormSubmitting = false;
		}
	}

	async function createGroupAppointment() {
		if (groupPatientIds.length === 0) {
			toast.error('Add at least one patient.');
			return;
		}
		groupFormSubmitting = true;
		try {
			await api.post('/scheduler/group', {
				patient_ids: groupPatientIds,
				date: groupForm.date,
				hour: groupForm.hour,
				minute: groupForm.minute,
				duration: groupForm.duration,
				note: groupForm.note
			});
			toast.success('Group appointment created.');
			showNewGroup = false;
			groupPatientIds = [];
			groupForm = { date: '', hour: 9, minute: 0, duration: 30, note: '' };
			refreshCalendar();
		} catch (e: any) {
			toast.error(e.message || 'Failed to create group.');
		} finally {
			groupFormSubmitting = false;
		}
	}

	function addGroupPatient() {
		const input = document.getElementById('group-patient-input') as HTMLInputElement;
		const id = parseInt(input?.value || '0');
		if (id > 0 && !groupPatientIds.includes(id)) {
			groupPatientIds = [...groupPatientIds, id];
			input.value = '';
		}
	}

	function removeGroupPatient(id: number) {
		groupPatientIds = groupPatientIds.filter(p => p !== id);
	}

	function resetNewForm() {
		newForm = {
			patient: 0,
			provider: 0,
			date: dateToJSONLocal(new Date()),
			hour: 9,
			minute: 0,
			duration: 30,
			note: ''
		};
	}

	function openNewFormWithDate(dateStr: string) {
		newForm.date = dateStr;
		showNewForm = true;
	}

	// --- Copy Appointment ---

	async function copyAppointment() {
		if (!selectedEvent || !copyDate) return;
		copySubmitting = true;
		try {
			await api.post('/scheduler/' + selectedEvent.scheduler_id + '/copy', {
				date: copyDate,
				hour: copyHour,
				minute: copyMinute
			});
			toast.success('Appointment copied.');
			copyDate = '';
			refreshCalendar();
		} catch (e: any) {
			toast.error(e.message || 'Failed to copy appointment.');
		} finally {
			copySubmitting = false;
		}
	}

	// --- Event loading (FullCalendar events-as-function) ---

	function loadEvents(info: any, successCallback: any, failureCallback: any) {
		const from = dateToJSONLocal(info.start);
		const to = dateToJSONLocal(info.end);

		api
			.get<SchedulerEvent[]>('/scheduler/dailyapptrange/' + from + '/' + to)
			.then((data) => {
				const events = data.map((v) => {
					const startDate = new Date(v.date_of.slice(0, 10));
					startDate.setHours(v.hour);
					startDate.setMinutes(v.minute);

					const s = v.hour * 60 + v.minute;
					const endDate = new Date(v.date_of.slice(0, 10));
					endDate.setHours(Math.floor((s + v.duration) / 60));
					endDate.setMinutes((s + v.duration) % 60);

					return {
						id: v.scheduler_id,
						title: buildTitle(v.patient, v.provider, v.note),
						start: startDate,
						end: endDate,
						color: v.template_color || undefined,
					};
				});
				successCallback(events);
			})
			.catch((err) => {
				console.error('Failed to load events:', err);
				failureCallback(err);
			});
	}

	// --- Event handlers ---

	function onCalendarReady(cal: FullCalendar) {
		calendarRef = cal;
	}

	function onEventClick(info: any) {
		modalLoading = true;
		showModal = true;
		copyDate = '';

		api
			.get<EventDetail>('/scheduler/event/' + info.event.id)
			.then((data) => {
				selectedEvent = data;
			})
			.catch(() => {
				toast.error('Unable to retrieve event.');
				showModal = false;
			})
			.finally(() => {
				modalLoading = false;
			});
	}

	function onEventDrop(info: any) {
		const changes: Record<string, unknown> = {};
		if (info.oldEvent.start?.toString() !== info.event.start?.toString()) {
			changes.date = dateToJSONLocal(info.event.start);
			changes.hour = info.event.start.getHours();
			changes.minute = info.event.start.getMinutes();
		}

		api
			.post('/scheduler/reschedule/' + info.event.id, changes)
			.then(() => {
				toast.info('Appointment rescheduled.');
			})
			.catch(() => {
				toast.error('Unable to reschedule appointment.');
				info.revert();
			});
	}

	function onEventResize(info: any) {
		const duration = Math.floor(
			(info.event.end.getTime() - info.event.start.getTime()) / (1000 * 60)
		);

		api
			.post('/scheduler/reschedule/' + info.event.id, { duration })
			.then(() => {
				toast.info('Appointment duration changed.');
			})
			.catch(() => {
				toast.error('Unable to adjust appointment duration.');
				info.revert();
			});
	}

	async function cancelAppointment() {
		if (!selectedEvent) return;
		if (!confirm('Cancel this appointment?')) return;
		try {
			await api.del('/scheduler/' + selectedEvent.scheduler_id);
			toast.success('Appointment cancelled.');
			closeModal();
			refreshCalendar();
		} catch (e: any) {
			toast.error(e.message || 'Failed to cancel appointment.');
		}
	}

	function viewPatient() {
		if (selectedEvent?.patient_id) {
			closeModal();
			goto(`/patients/${selectedEvent.patient_id}`);
		}
	}

	function closeModal() {
		showModal = false;
		selectedEvent = null;
		copyDate = '';
	}

	function onModalBackdropClick(e: MouseEvent) {
		if (e.target === e.currentTarget) {
			closeModal();
		}
	}

	function onModalKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') {
			closeModal();
		}
	}
</script>

<div class="max-w-7xl mx-auto">
	<div class="flex items-center justify-between mb-6">
		<h1 class="text-2xl font-bold text-gray-900">Scheduler</h1>
		<div class="flex items-center gap-2">
			<a
				href="/scheduler/templates"
				class="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors"
			>
				Templates
			</a>
			<button
				onclick={() => (showNewForm = true)}
				class="px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-lg hover:bg-blue-700 transition-colors"
			>
				+ New Appointment
			</button>
		</div>
	</div>

	<Calendar
		events={loadEvents}
		onReady={onCalendarReady}
		onEventClick={onEventClick}
		onEventDrop={onEventDrop}
		onEventResize={onEventResize}
	/>
</div>

<!-- Event detail modal -->
{#if showModal}
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<div
		class="fixed inset-0 z-50 flex items-center justify-center bg-black/40"
		onclick={onModalBackdropClick}
		onkeydown={onModalKeydown}
		role="dialog"
		aria-modal="true"
		aria-label="Appointment details"
		tabindex="-1"
	>
		<div class="bg-white rounded-xl shadow-xl w-full max-w-lg mx-4 overflow-hidden">
			<!-- Header -->
			<div class="flex items-center justify-between px-6 py-4 border-b border-gray-200">
				<h2 class="text-lg font-semibold text-gray-900">
					Appointment: {modalLoading ? 'Loading...' : selectedEvent?.patient ?? ''}
				</h2>
				<button
					onclick={closeModal}
					class="text-gray-400 hover:text-gray-600 transition-colors p-1 rounded-lg hover:bg-gray-100"
					aria-label="Close"
				>
					<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
						<path
							fill-rule="evenodd"
							d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z"
							clip-rule="evenodd"
						/>
					</svg>
				</button>
			</div>

			<!-- Body -->
			<div class="px-6 py-4">
				{#if modalLoading}
					<div class="flex justify-center py-8">
						<div class="w-6 h-6 border-2 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
					</div>
				{:else if selectedEvent}
					<dl class="space-y-3 text-sm">
						<div class="flex justify-between">
							<dt class="text-gray-500 font-medium">Date</dt>
							<dd class="text-gray-900">{selectedEvent.date_of_mdy}</dd>
						</div>
						<div class="flex justify-between">
							<dt class="text-gray-500 font-medium">Time</dt>
							<dd class="text-gray-900">
								{pad(selectedEvent.hour)}:{pad(selectedEvent.minute)}
								{#if selectedEvent.appointment_time}
									<span class="text-gray-400 ml-1">({selectedEvent.appointment_time})</span>
								{/if}
							</dd>
						</div>
						<div class="flex justify-between">
							<dt class="text-gray-500 font-medium">Duration</dt>
							<dd class="text-gray-900">{selectedEvent.duration} min</dd>
						</div>
						<div class="flex justify-between">
							<dt class="text-gray-500 font-medium">Provider</dt>
							<dd class="text-gray-900">{selectedEvent.provider || '—'}</dd>
						</div>
						<div class="flex justify-between">
							<dt class="text-gray-500 font-medium">Patient</dt>
							<dd class="text-gray-900 font-semibold">{selectedEvent.patient}</dd>
						</div>
						{#if selectedEvent.note}
							<div class="flex justify-between">
								<dt class="text-gray-500 font-medium">Note</dt>
								<dd class="text-gray-900 italic">{selectedEvent.note}</dd>
							</div>
						{/if}
						<div class="flex justify-between items-center">
							<dt class="text-gray-500 font-medium">Status</dt>
							<dd class="flex items-center gap-1.5">
								{#if selectedEvent.status_color}
									<span
										class="inline-block w-2.5 h-2.5 rounded-full"
										style="background-color: {selectedEvent.status_color}"
									></span>
								{/if}
								<span class="text-gray-900">{selectedEvent.status}</span>
							</dd>
						</div>
					</dl>
				{/if}
			</div>

			<!-- Footer -->
			<div class="px-6 py-4 bg-gray-50 border-t border-gray-200">
				<div class="flex justify-between gap-3 mb-3">
					<div class="flex gap-2">
						{#if selectedEvent?.patient_id}
							<button
								onclick={viewPatient}
								class="px-3 py-2 text-sm font-medium text-blue-700 bg-blue-50 border border-blue-200 rounded-lg hover:bg-blue-100 transition-colors"
							>
								View Patient
							</button>
						{/if}
						<button
							onclick={cancelAppointment}
							class="px-3 py-2 text-sm font-medium text-red-700 bg-red-50 border border-red-200 rounded-lg hover:bg-red-100 transition-colors"
						>
							Cancel
						</button>
					</div>
					<button
						onclick={closeModal}
						class="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors"
					>
						Close
					</button>
				</div>

				<!-- Copy section -->
				<div class="border-t border-gray-200 pt-3 flex flex-wrap items-end gap-2">
					<span class="text-sm font-medium text-gray-500 w-full">Copy to new date:</span>
					<input
						type="date"
						bind:value={copyDate}
						class="border border-gray-300 rounded-md px-2 py-1 text-sm"
					/>
					<select bind:value={copyHour} class="border border-gray-300 rounded-md px-2 py-1 text-sm">
						{#each Array(24) as _, h}
							<option value={h}>{pad(h)}</option>
						{/each}
					</select>
					<span class="text-gray-400">:</span>
					<select bind:value={copyMinute} class="border border-gray-300 rounded-md px-2 py-1 text-sm">
						{#each [0, 15, 30, 45] as m}
							<option value={m}>{pad(m)}</option>
						{/each}
					</select>
					<button
						onclick={copyAppointment}
						disabled={copySubmitting || !copyDate}
						class="px-3 py-1.5 text-sm font-medium text-white bg-green-600 rounded-lg hover:bg-green-700 transition-colors disabled:opacity-50"
					>
						{copySubmitting ? 'Copying...' : 'Copy'}
					</button>
				</div>
			</div>
		</div>
	</div>
{/if}

<!-- New Appointment Modal -->
{#if showNewForm}
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<div
		class="fixed inset-0 z-50 flex items-center justify-center bg-black/40"
		onclick={(e: MouseEvent) => { if (e.target === e.currentTarget) showNewForm = false; }}
		onkeydown={(e: KeyboardEvent) => { if (e.key === 'Escape') showNewForm = false; }}
		role="dialog"
		aria-modal="true"
		aria-label="New appointment"
		tabindex="-1"
	>
		<div class="bg-white rounded-xl shadow-xl w-full max-w-md mx-4 overflow-hidden">
			<div class="flex items-center justify-between px-6 py-4 border-b border-gray-200">
				<h2 class="text-lg font-semibold text-gray-900">New Appointment</h2>
				<button
					onclick={() => (showNewForm = false)}
					class="text-gray-400 hover:text-gray-600 transition-colors p-1 rounded-lg hover:bg-gray-100"
					aria-label="Close"
				>
					<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
						<path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd"/>
					</svg>
				</button>
			</div>

			<div class="px-6 py-4 space-y-4">
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Patient ID <span class="text-red-500">*</span></label>
					<input
						type="number"
						bind:value={newForm.patient}
						class="w-full border border-gray-300 rounded-md px-3 py-2 text-sm"
						placeholder="e.g. 42"
					/>
				</div>
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Provider ID</label>
					<input
						type="number"
						bind:value={newForm.provider}
						class="w-full border border-gray-300 rounded-md px-3 py-2 text-sm"
						placeholder="e.g. 1"
					/>
				</div>
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Date</label>
					<input
						type="date"
						bind:value={newForm.date}
						class="w-full border border-gray-300 rounded-md px-3 py-2 text-sm"
					/>
				</div>
				<div class="flex gap-2">
					<div class="flex-1">
						<label class="block text-sm font-medium text-gray-700 mb-1">Hour</label>
						<select bind:value={newForm.hour} class="w-full border border-gray-300 rounded-md px-2 py-2 text-sm">
							{#each Array(24) as _, h}
								<option value={h}>{pad(h)}</option>
							{/each}
						</select>
					</div>
					<div class="flex-1">
						<label class="block text-sm font-medium text-gray-700 mb-1">Minute</label>
						<select bind:value={newForm.minute} class="w-full border border-gray-300 rounded-md px-2 py-2 text-sm">
							{#each [0, 15, 30, 45] as m}
								<option value={m}>{pad(m)}</option>
							{/each}
						</select>
					</div>
				</div>
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Duration (min)</label>
					<select bind:value={newForm.duration} class="w-full border border-gray-300 rounded-md px-3 py-2 text-sm">
						<option value={15}>15</option>
						<option value={30}>30</option>
						<option value={45}>45</option>
						<option value={60}>60</option>
						<option value={90}>90</option>
					</select>
				</div>
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Note</label>
					<textarea
						bind:value={newForm.note}
						class="w-full border border-gray-300 rounded-md px-3 py-2 text-sm"
						rows="2"
						placeholder="Reason for visit..."
					></textarea>
				</div>
			</div>

			<div class="flex justify-end gap-3 px-6 py-4 bg-gray-50 border-t border-gray-200">
				<button
					onclick={() => (showNewForm = false)}
					class="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors"
				>
					Cancel
				</button>
				<button
					onclick={createAppointment}
					disabled={newFormSubmitting}
					class="px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-lg hover:bg-blue-700 transition-colors disabled:opacity-50"
				>
					{newFormSubmitting ? 'Creating...' : 'Create Appointment'}
				</button>
			</div>
		</div>
	</div>
{/if}
