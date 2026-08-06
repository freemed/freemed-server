<script lang="ts">
	import { goto } from '$app/navigation';
	import { api } from '$lib/api';
	import Calendar from '$lib/components/Calendar.svelte';

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

	// --- State ---

	let selectedEvent = $state<EventDetail | null>(null);
	let showModal = $state(false);
	let modalLoading = $state(false);

	// --- Helpers ---

	function dateToJSONLocal(date: Date): string {
		const local = new Date(date);
		local.setMinutes(date.getMinutes() - date.getTimezoneOffset());
		return local.toJSON().slice(0, 10);
	}

	function buildTitle(patient: string, provider: string | null, note: string): string {
		let title = patient;
		if (provider) title += ` (${provider})`;
		if (note) title += ` [${note}]`;
		return title;
	}

	function pad(n: number): string {
		return n.toString().padStart(2, '0');
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

	function onEventClick(info: any) {
		modalLoading = true;
		showModal = true;

		api
			.get<EventDetail>('/scheduler/event/' + info.event.id)
			.then((data) => {
				selectedEvent = data;
			})
			.catch(() => {
				(window as any).toast?.error('Unable to retrieve event.');
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
				(window as any).toast?.info('Appointment rescheduled.');
			})
			.catch(() => {
				(window as any).toast?.error('Unable to reschedule appointment.');
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
				(window as any).toast?.info('Appointment duration changed.');
			})
			.catch(() => {
				(window as any).toast?.error('Unable to adjust appointment duration.');
				info.revert();
			});
	}

	async function cancelAppointment() {
		if (!selectedEvent) return;
		if (!confirm('Cancel this appointment?')) return;
		try {
			await api.del('/scheduler/' + selectedEvent.scheduler_id);
			(window as any).toast?.success('Appointment cancelled.');
			closeModal();
		} catch (e: any) {
			(window as any).toast?.error(e.message || 'Failed to cancel appointment.');
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
	<h1 class="text-2xl font-bold text-gray-900 mb-6">Scheduler</h1>

	<Calendar
		events={loadEvents}
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
			<div class="flex justify-between gap-3 px-6 py-4 bg-gray-50 border-t border-gray-200">
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
						Cancel Appointment
					</button>
				</div>
				<button
					onclick={closeModal}
					class="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors"
				>
					Close
				</button>
			</div>
		</div>
	</div>
{/if}
