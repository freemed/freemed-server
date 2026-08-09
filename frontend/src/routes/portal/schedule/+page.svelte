<script lang="ts">
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import ErrorBanner from '$lib/components/ErrorBanner.svelte';
	import EmptyState from '$lib/components/EmptyState.svelte';
	import { z } from 'zod';

	interface Slot {
		date: string;
		start_time: string;
		end_time: string;
		provider_id: number;
		provider_name: string;
	}

	interface Provider {
		id: number;
		phyfname: string;
		phylname: string;
	}

	let slots = $state<Slot[]>([]);
	let providers = $state<Provider[]>([]);
	let loading = $state(false);
	let booking = $state(false);
	let error = $state('');
	let successMsg = $state('');

	let selectedProvider = $state('all');
	let fromDate = $state(todayStr());
	let toDate = $state(weekFromNowStr());

	let showBookingModal = $state(false);
	let selectedSlot = $state<Slot | null>(null);
	let reason = $state('');
	let bookingError = $state('');

	const reasonSchema = z.object({
		reason: z.string().min(1, 'Please enter a reason for your visit'),
	});

	function todayStr(): string {
		const d = new Date();
		return d.toISOString().slice(0, 10);
	}

	function weekFromNowStr(): string {
		const d = new Date();
		d.setDate(d.getDate() + 7);
		return d.toISOString().slice(0, 10);
	}

	async function portalFetch<T>(path: string): Promise<T> {
		const res = await fetch(`/api/portal${path}`);
		if (!res.ok) {
			const body = await res.text();
			throw new Error(`API error ${res.status}: ${body}`);
		}
		return res.json();
	}

	async function loadProviders() {
		try {
			const data = await portalFetch<Slot[]>('/slots?from=' + fromDate + '&to=' + toDate);
			// Extract unique providers from slots since there's no dedicated /providers endpoint
			const seen = new Set<number>();
			const provs: Provider[] = [];
			for (const s of data) {
				if (!seen.has(s.provider_id)) {
					seen.add(s.provider_id);
					provs.push({ id: s.provider_id, phyfname: s.provider_name.split(' ')[0] || '', phylname: s.provider_name.split(' ').slice(1).join(' ') || '' });
				}
			}
			providers = provs;
		} catch {
			// Providers will be extracted from slots
		}
	}

	async function searchSlots() {
		loading = true;
		error = '';
		successMsg = '';
		try {
			let url = `/slots?from=${fromDate}&to=${toDate}`;
			if (selectedProvider !== 'all') {
				url += `&provider=${selectedProvider}`;
			}
			const data = await portalFetch<Slot[]>(url);
			slots = data || [];
		} catch (e: any) {
			error = e.message || 'Failed to load available slots';
		} finally {
			loading = false;
		}
	}

	$effect(() => {
		searchSlots();
	});

	function openBooking(slot: Slot) {
		selectedSlot = slot;
		reason = '';
		bookingError = '';
		showBookingModal = true;
	}

	function closeBooking() {
		showBookingModal = false;
		selectedSlot = null;
		reason = '';
		bookingError = '';
	}

	async function confirmBooking() {
		const result = reasonSchema.safeParse({ reason });
		if (!result.success) {
			bookingError = result.error.issues[0].message;
			return;
		}

		if (!selectedSlot) return;

		booking = true;
		bookingError = '';
		try {
			const res = await fetch('/api/portal/appointments', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					slot_date: selectedSlot.date,
					slot_time: selectedSlot.start_time,
					provider_id: selectedSlot.provider_id,
					reason: reason,
				}),
			});
			if (!res.ok) {
				const body = await res.json().catch(() => ({ message: 'Booking failed' }));
				throw new Error(body.message || 'Booking failed');
			}
			closeBooking();
			successMsg = 'Appointment booked successfully!';
			await searchSlots();
		} catch (e: any) {
			bookingError = e.message || 'Failed to book appointment';
		} finally {
			booking = false;
		}
	}

	function formatDate(dateStr: string): string {
		const d = new Date(dateStr + 'T00:00:00');
		return d.toLocaleDateString('en-US', {
			weekday: 'short',
			month: 'short',
			day: 'numeric',
		});
	}

	// Group slots by date
	let groupedSlots = $derived.by(() => {
		const groups: Record<string, Slot[]> = {};
		for (const slot of slots) {
			if (!groups[slot.date]) groups[slot.date] = [];
			groups[slot.date].push(slot);
		}
		return Object.entries(groups);
	});
</script>

<div class="max-w-7xl mx-auto">
	<div class="mb-6">
		<h1 class="text-2xl font-bold text-gray-900">Schedule Appointment</h1>
		<p class="text-sm text-gray-500 mt-1">Select an available time slot to book your appointment</p>
	</div>

	<!-- Success banner -->
	{#if successMsg}
		<div class="mb-4 bg-green-50 border border-green-200 rounded-lg p-4 flex items-center justify-between">
			<div class="flex items-center gap-2">
				<svg class="w-5 h-5 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
				</svg>
				<span class="text-sm text-green-700">{successMsg}</span>
			</div>
			<button onclick={() => successMsg = ''} class="text-green-500 hover:text-green-700 text-sm">&times;</button>
		</div>
	{/if}

	<!-- Filters -->
	<div class="bg-white rounded-lg shadow-sm border border-gray-200 p-4 mb-6">
		<div class="flex flex-wrap gap-4 items-end">
			<div>
				<label for="fromDate" class="block text-xs font-medium text-gray-500 mb-1">From</label>
				<input
					id="fromDate"
					type="date"
					bind:value={fromDate}
					onchange={searchSlots}
					class="px-3 py-2 text-sm border border-gray-300 rounded-md focus:outline-none focus:ring-1 focus:ring-blue-500"
				/>
			</div>
			<div>
				<label for="toDate" class="block text-xs font-medium text-gray-500 mb-1">To</label>
				<input
					id="toDate"
					type="date"
					bind:value={toDate}
					onchange={searchSlots}
					class="px-3 py-2 text-sm border border-gray-300 rounded-md focus:outline-none focus:ring-1 focus:ring-blue-500"
				/>
			</div>
			<div>
				<label for="provider" class="block text-xs font-medium text-gray-500 mb-1">Provider</label>
				<select
					id="provider"
					bind:value={selectedProvider}
					onchange={searchSlots}
					class="px-3 py-2 text-sm border border-gray-300 rounded-md focus:outline-none focus:ring-1 focus:ring-blue-500"
				>
					<option value="all">All Providers</option>
					{#each providers as p}
						<option value={p.id}>{p.phyfname} {p.phylname}</option>
					{/each}
				</select>
			</div>
			<button
				onclick={searchSlots}
				class="px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-md hover:bg-blue-700 transition-colors"
			>
				Search
			</button>
		</div>
	</div>

	{#if loading}
		<LoadingSpinner message="Loading available slots…" />
	{:else if error}
		<ErrorBanner message={error} onRetry={searchSlots} />
	{:else if slots.length === 0}
		<EmptyState
			title="No available slots"
			message="Try adjusting your date range or provider filter."
		/>
	{:else}
		<!-- Slots grouped by date -->
		{#each groupedSlots as [date, daySlots]}
			<div class="mb-6">
				<h3 class="text-sm font-semibold text-gray-700 mb-3">{formatDate(date)}</h3>
				<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
					{#each daySlots as slot (slot.date + slot.start_time + slot.provider_id)}
						<button
							onclick={() => openBooking(slot)}
							class="bg-white rounded-lg shadow-sm border border-gray-200 p-4 text-left hover:border-blue-300 hover:shadow-md transition-all"
						>
							<div class="flex justify-between items-start mb-2">
								<span class="text-lg font-semibold text-gray-900">{slot.start_time}</span>
								<span class="text-xs text-gray-400">&rarr; {slot.end_time}</span>
							</div>
							<p class="text-sm text-gray-600">{slot.provider_name}</p>
						</button>
					{/each}
				</div>
			</div>
		{/each}
	{/if}

	<!-- Booking confirmation modal -->
	{#if showBookingModal && selectedSlot}
		<div class="fixed inset-0 z-50 flex items-center justify-center bg-black/40" role="dialog" aria-modal="true">
			<div class="bg-white rounded-lg shadow-xl w-full max-w-md mx-4" onclick={(e: MouseEvent) => e.stopPropagation()} onkeydown={() => {}}>
				<div class="px-6 py-4 border-b border-gray-100 flex justify-between items-center">
					<h3 class="text-lg font-semibold text-gray-900">Confirm Appointment</h3>
					<button
						onclick={closeBooking}
						class="p-1 text-gray-400 hover:text-gray-600 rounded-md hover:bg-gray-100"
						aria-label="Close"
					>
						<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
						</svg>
					</button>
				</div>

				<div class="p-6 space-y-4">
					<div class="bg-gray-50 rounded-lg p-4 space-y-2">
						<div class="flex justify-between">
							<span class="text-sm text-gray-500">Date</span>
							<span class="text-sm font-medium text-gray-900">{formatDate(selectedSlot.date)}</span>
						</div>
						<div class="flex justify-between">
							<span class="text-sm text-gray-500">Time</span>
							<span class="text-sm font-medium text-gray-900">{selectedSlot.start_time} &ndash; {selectedSlot.end_time}</span>
						</div>
						<div class="flex justify-between">
							<span class="text-sm text-gray-500">Provider</span>
							<span class="text-sm font-medium text-gray-900">{selectedSlot.provider_name}</span>
						</div>
					</div>

					<div>
						<label for="reason" class="block text-sm font-medium text-gray-700 mb-1">Reason for Visit</label>
						<textarea
							id="reason"
							bind:value={reason}
							rows={3}
							class="w-full px-3 py-2 text-sm border border-gray-300 rounded-md focus:outline-none focus:ring-1 focus:ring-blue-500"
							placeholder="Briefly describe the reason for your appointment"
						></textarea>
					</div>

					{#if bookingError}
						<div class="bg-red-50 border border-red-200 rounded-md p-3 text-sm text-red-700">{bookingError}</div>
					{/if}

					<div class="flex justify-end gap-3 pt-2">
						<button
							onclick={closeBooking}
							class="px-4 py-2 text-sm text-gray-600 hover:bg-gray-100 rounded-md transition-colors"
						>
							Cancel
						</button>
						<button
							onclick={confirmBooking}
							disabled={booking}
							class="px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-md hover:bg-blue-700 transition-colors disabled:opacity-50"
						>
							{booking ? 'Booking…' : 'Confirm Booking'}
						</button>
					</div>
				</div>
			</div>
		</div>
	{/if}
</div>
