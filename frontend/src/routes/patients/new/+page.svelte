<script lang="ts">
	import { goto } from '$app/navigation';
	import { api } from '$lib/api';

	interface DuplicateResult {
		id: number;
		last_name: string;
		first_name: string;
		patient_id: string;
		date_of_birth: string;
	}

	interface PatientForm {
		first_name: string;
		last_name: string;
		middle_name: string;
		date_of_birth: string;
		gender: string;
		patient_id: string;
		address_line_1: string;
		address_line_2: string;
		city: string;
		state: string;
		postal: string;
		phone: string;
	}

	let form = $state<PatientForm>({
		first_name: '',
		last_name: '',
		middle_name: '',
		date_of_birth: '',
		gender: '',
		patient_id: '',
		address_line_1: '',
		address_line_2: '',
		city: '',
		state: '',
		postal: '',
		phone: '',
	});

	let submitting = $state(false);
	let error = $state('');
	let success = $state('');
	let duplicates = $state<DuplicateResult[]>([]);
	let checkingDuplicates = $state(false);
	let duplicateChecked = $state(false);

	async function checkDuplicates() {
		if (!form.last_name || !form.first_name) {
			error = 'Last name and first name are required';
			return;
		}
		checkingDuplicates = true;
		error = '';
		duplicates = [];
		duplicateChecked = false;
		try {
			const data = await api.post<DuplicateResult[]>('/patients/searchDuplicates', {
				ptlname: form.last_name,
				ptfname: form.first_name,
				ptmname: form.middle_name,
				ptsuffix: '',
				ptdob: form.date_of_birth,
			});
			duplicates = data || [];
			duplicateChecked = true;
		} catch (e: any) {
			error = e.message || 'Duplicate check failed';
		} finally {
			checkingDuplicates = false;
		}
	}

	async function createPatient() {
		submitting = true;
		error = '';
		success = '';
		try {
			// Placeholder — actual create endpoint TBD
			await api.post('/patients', {
				ptlname: form.last_name,
				ptfname: form.first_name,
				ptmname: form.middle_name,
				ptdob: form.date_of_birth,
				ptsex: form.gender,
				ptid: form.patient_id,
				ptaddr1: form.address_line_1,
				ptaddr2: form.address_line_2,
				ptcity: form.city,
				ptstate: form.state,
				ptzip: form.postal,
				pthphone: form.phone,
			});
			success = 'Patient created successfully';
			setTimeout(() => {
				goto('/patients');
			}, 1500);
		} catch (e: any) {
			error = e.message || 'Failed to create patient';
		} finally {
			submitting = false;
		}
	}

	async function handleSubmit() {
		error = '';
		success = '';

		if (!form.last_name || !form.first_name) {
			error = 'Last name and first name are required';
			return;
		}

		if (!duplicateChecked) {
			await checkDuplicates();
		}

		if (duplicates.length > 0) {
			return;
		}

		await createPatient();
	}
</script>

<div class="max-w-3xl mx-auto">
	<h1 class="text-2xl font-bold text-gray-900 mb-6">New Patient</h1>

	{#if error}
		<div class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg text-sm mb-6">
			{error}
		</div>
	{/if}

	{#if success}
		<div class="bg-green-50 border border-green-200 text-green-700 px-4 py-3 rounded-lg text-sm mb-6">
			{success}
		</div>
	{/if}

	<!-- Duplicate results -->
	{#if duplicates.length > 0}
		<div class="bg-yellow-50 border border-yellow-200 rounded-lg p-4 mb-6">
			<h2 class="text-sm font-semibold text-yellow-800 mb-2">
				Possible duplicates found ({duplicates.length})
			</h2>
			<p class="text-sm text-yellow-700 mb-3">
				Existing patients match the name{form.date_of_birth ? ' and date of birth' : ''} you entered. Please review before creating.
			</p>
			<div class="space-y-2">
				{#each duplicates as dup}
					<a
						href="/patients/{dup.id}"
						class="block bg-white border border-yellow-200 rounded px-3 py-2 text-sm hover:bg-yellow-50 transition-colors"
					>
						<span class="font-medium text-gray-900">{dup.last_name}, {dup.first_name}</span>
						<span class="text-gray-500 ml-2">ID: {dup.patient_id}</span>
						{#if dup.date_of_birth}
							<span class="text-gray-400 ml-2">DOB: {new Date(dup.date_of_birth).toLocaleDateString('en-US')}</span>
						{/if}
					</a>
				{/each}
			</div>
			<div class="mt-3 flex gap-2">
				<button
					onclick={createPatient}
					disabled={submitting}
					class="px-4 py-2 bg-yellow-600 text-white text-sm font-medium rounded-lg hover:bg-yellow-700 disabled:opacity-50 transition-colors"
				>
					Create Anyway
				</button>
				<a
					href="/patients"
					class="px-4 py-2 bg-gray-100 text-gray-700 text-sm font-medium rounded-lg hover:bg-gray-200 transition-colors"
				>
					Cancel
				</a>
			</div>
		</div>
	{/if}

	<form onsubmit={(e) => { e.preventDefault(); handleSubmit(); }} class="bg-white rounded-lg shadow-sm border border-gray-200 p-6 space-y-5">
		<!-- Name fields -->
		<fieldset>
			<legend class="text-sm font-semibold text-gray-700 mb-3">Name</legend>
			<div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
				<div>
					<label for="last_name" class="block text-sm font-medium text-gray-600 mb-1">
						Last Name <span class="text-red-500">*</span>
					</label>
					<input
						id="last_name"
						type="text"
						bind:value={form.last_name}
						required
						class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none transition-shadow"
					/>
				</div>
				<div>
					<label for="first_name" class="block text-sm font-medium text-gray-600 mb-1">
						First Name <span class="text-red-500">*</span>
					</label>
					<input
						id="first_name"
						type="text"
						bind:value={form.first_name}
						required
						class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none transition-shadow"
					/>
				</div>
				<div>
					<label for="middle_name" class="block text-sm font-medium text-gray-600 mb-1">
						Middle Name
					</label>
					<input
						id="middle_name"
						type="text"
						bind:value={form.middle_name}
						class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none transition-shadow"
					/>
				</div>
			</div>
		</fieldset>

		<!-- Demographics -->
		<fieldset>
			<legend class="text-sm font-semibold text-gray-700 mb-3">Demographics</legend>
			<div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
				<div>
					<label for="date_of_birth" class="block text-sm font-medium text-gray-600 mb-1">
						Date of Birth
					</label>
					<input
						id="date_of_birth"
						type="date"
						bind:value={form.date_of_birth}
						class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none transition-shadow"
					/>
				</div>
				<div>
					<label for="gender" class="block text-sm font-medium text-gray-600 mb-1">
						Gender
					</label>
					<select
						id="gender"
						bind:value={form.gender}
						class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none transition-shadow bg-white"
					>
						<option value="">Select...</option>
						<option value="M">Male</option>
						<option value="F">Female</option>
						<option value="O">Other</option>
						<option value="U">Unknown</option>
					</select>
				</div>
				<div>
					<label for="patient_id" class="block text-sm font-medium text-gray-600 mb-1">
						Patient ID
					</label>
					<input
						id="patient_id"
						type="text"
						bind:value={form.patient_id}
						placeholder="Auto-generated if blank"
						class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none transition-shadow"
					/>
				</div>
			</div>
		</fieldset>

		<!-- Contact -->
		<fieldset>
			<legend class="text-sm font-semibold text-gray-700 mb-3">Address &amp; Contact</legend>
			<div class="space-y-4">
				<div>
					<label for="address_line_1" class="block text-sm font-medium text-gray-600 mb-1">
						Address Line 1
					</label>
					<input
						id="address_line_1"
						type="text"
						bind:value={form.address_line_1}
						class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none transition-shadow"
					/>
				</div>
				<div>
					<label for="address_line_2" class="block text-sm font-medium text-gray-600 mb-1">
						Address Line 2
					</label>
					<input
						id="address_line_2"
						type="text"
						bind:value={form.address_line_2}
						class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none transition-shadow"
					/>
				</div>
				<div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
					<div>
						<label for="city" class="block text-sm font-medium text-gray-600 mb-1">
							City
						</label>
						<input
							id="city"
							type="text"
							bind:value={form.city}
							class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none transition-shadow"
						/>
					</div>
					<div>
						<label for="state" class="block text-sm font-medium text-gray-600 mb-1">
							State
						</label>
						<input
							id="state"
							type="text"
							bind:value={form.state}
							maxlength="2"
							placeholder="e.g. CA"
							class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none transition-shadow"
						/>
					</div>
					<div>
						<label for="postal" class="block text-sm font-medium text-gray-600 mb-1">
							Postal Code
						</label>
						<input
							id="postal"
							type="text"
							bind:value={form.postal}
							class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none transition-shadow"
						/>
					</div>
				</div>
				<div>
					<label for="phone" class="block text-sm font-medium text-gray-600 mb-1">
						Phone
					</label>
					<input
						id="phone"
						type="tel"
						bind:value={form.phone}
						class="w-full max-w-xs px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none transition-shadow"
					/>
				</div>
			</div>
		</fieldset>

		<!-- Actions -->
		<div class="flex items-center gap-3 pt-2">
			<button
				type="button"
				onclick={checkDuplicates}
				disabled={checkingDuplicates || submitting}
				class="px-4 py-2.5 bg-gray-100 text-gray-700 text-sm font-medium rounded-lg hover:bg-gray-200 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
			>
				{#if checkingDuplicates}
					<span class="inline-block w-4 h-4 border-2 border-gray-500 border-t-transparent rounded-full animate-spin align-middle mr-1"></span>
				{/if}
				Check Duplicates
			</button>
			<button
				type="button"
				onclick={handleSubmit}
				disabled={submitting || checkingDuplicates}
				class="px-6 py-2.5 bg-blue-600 text-white text-sm font-medium rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
			>
				{#if submitting}
					<span class="inline-block w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin align-middle mr-1"></span>
				{/if}
				Create Patient
			</button>
			<a
				href="/patients"
				class="px-4 py-2.5 text-gray-500 text-sm font-medium hover:text-gray-700 transition-colors"
			>
				Cancel
			</a>
		</div>
	</form>
</div>
