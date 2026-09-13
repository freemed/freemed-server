<script lang="ts">
	import { page } from '$app/stores';
	import { auth } from '$lib/stores/auth.svelte';
	import { api } from '$lib/api';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import ErrorBanner from '$lib/components/ErrorBanner.svelte';

	/**
	 * Browser consent screen for the SMART on FHIR authorization endpoint.
	 *
	 * `/oauth2/authorize` deliberately refuses any request that does not already
	 * carry `consent=approve`, and it answers machine callers with JSON — so a
	 * human never sees it. Discovery (/.well-known/smart-configuration)
	 * therefore advertises THIS route as the authorization endpoint: the app
	 * sends the browser here, the user reviews the requested scopes, and on
	 * approval the page forwards the *same* request to /oauth2/authorize with
	 * `consent=approve`.
	 *
	 * Nothing is decided here. Every check the server applies to a direct
	 * authorization request still applies to the one this page issues: the staff
	 * session, the explicit consent, the exact redirect_uri match, the launch
	 * patient and the narrowing of the requested scopes to the client's
	 * registration. This screen only presents what the server would otherwise
	 * enforce invisibly.
	 */

	interface FhirClientPublic {
		client_id: string;
		client_name: string;
		scopes: string;
		is_confidential: boolean;
	}

	// Plain-language labels for the scopes this deployment can grant. A scope
	// that is not listed is rendered verbatim rather than hidden, so consent is
	// never asked for less than the request actually contains.
	const SCOPE_LABELS: Record<string, string> = {
		launch: 'Launch this app in the context of a patient',
		'launch/patient': 'Choose a patient to launch this app in the context of',
		'launch/encounter': 'Launch this app in the context of an encounter',
		openid: 'Verify who you are (OpenID Connect)',
		profile: 'Read your basic profile',
		fhirUser: 'Read the identity of the signed-in user',
		offline_access: 'Stay connected after you sign out',
		online_access: 'Stay connected while you are signed in',
		'patient/*.read': 'Read the medical record of the launched patient',
		'patient/*.write': 'Add to the medical record of the launched patient',
		'user/*.read': 'Read records on behalf of the signed-in user',
		'user/*.write': 'Write records on behalf of the signed-in user',
		'system/*.read': 'Read records for every patient in the system',
		'system/*.write': 'Write records for every patient in the system'
	};

	function scopeLabel(scope: string): string {
		return SCOPE_LABELS[scope] ?? `Access to ${scope}`;
	}

	function splitScopes(raw: string): string[] {
		return raw.split(/\s+/).filter((s) => s.length > 0);
	}

	type AuthorizeRequest = {
		response_type: string;
		client_id: string;
		redirect_uri: string;
		scope: string;
		state: string;
		launch: string;
		patient: string;
	};

	// The only parameters forwarded to /oauth2/authorize. Rebuilding the query
	// from this list (instead of replaying whatever the caller attached) means
	// the request the user approves is exactly the one shown on screen — an
	// extra parameter such as `aud` is dropped rather than carried along.
	const FORWARDED_PARAMS = [
		'response_type',
		'client_id',
		'redirect_uri',
		'scope',
		'state',
		'launch',
		'patient'
	] as const;

	function readRequest(url: URL): AuthorizeRequest {
		const q = url.searchParams;
		return {
			response_type: q.get('response_type') ?? '',
			client_id: q.get('client_id') ?? '',
			redirect_uri: q.get('redirect_uri') ?? '',
			scope: q.get('scope') ?? '',
			state: q.get('state') ?? '',
			launch: q.get('launch') ?? '',
			patient: q.get('patient') ?? ''
		};
	}

	let request = $derived(readRequest($page.url));

	// Anything that makes the request un-authorizable on its face. The server
	// enforces all of this as well; showing it here means the user is never
	// asked to approve something that cannot succeed.
	const requestError = $derived.by(() => {
		const p = request;
		if (!p.client_id) return 'This authorization request is missing its client_id.';
		if (!p.redirect_uri) return 'This authorization request is missing its redirect_uri.';
		if (p.response_type !== 'code') {
			return p.response_type
				? `This server only supports response_type=code, and the request asked for "${p.response_type}".`
				: 'This authorization request is missing its response_type.';
		}
		try {
			new URL(p.redirect_uri);
		} catch {
			return 'This authorization request has a malformed redirect_uri.';
		}
		return '';
	});

	const redirectTarget = $derived.by(() => {
		try {
			return new URL(request.redirect_uri).host || request.redirect_uri;
		} catch {
			return request.redirect_uri;
		}
	});

	let client = $state<FhirClientPublic | null>(null);
	let loadingClient = $state(true);
	let clientError = $state('');
	let reloadNonce = $state(0);

	async function loadClient(clientId: string) {
		loadingClient = true;
		clientError = '';
		try {
			client = await api.get<FhirClientPublic>(
				`/smart/clients/${encodeURIComponent(clientId)}/public`
			);
		} catch (e: unknown) {
			client = null;
			// `api` throws `API error <status>: <body>`; the two statuses this
			// endpoint can answer with are worth saying in plain language.
			const message = e instanceof Error ? e.message : '';
			if (message.startsWith('API error 404')) {
				clientError = 'This application is not registered, or it has been deactivated.';
			} else if (message.startsWith('API error 403')) {
				clientError = 'You do not have permission to view this application.';
			} else {
				clientError = message || 'Could not load the requesting application.';
			}
		} finally {
			loadingClient = false;
		}
	}

	$effect(() => {
		const clientId = request.client_id;
		const blocked = requestError;
		void reloadNonce; // re-run the lookup when the user retries
		if (!clientId || blocked) {
			client = null;
			loadingClient = false;
			return;
		}
		void loadClient(clientId);
	});

	// A request with no scope means "everything the client is registered for"
	// (see OAuth2Authorize); showing the client's registered scopes then is
	// exactly what will be granted.
	const requestedScopes = $derived.by(() => {
		if (request.scope) return splitScopes(request.scope);
		return client ? splitScopes(client.scopes) : [];
	});

	const registeredScopes = $derived(client ? splitScopes(client.scopes) : []);

	const grantedScopes = $derived(
		requestedScopes.filter((scope) => registeredScopes.includes(scope))
	);

	const droppedScopes = $derived(
		requestedScopes.filter((scope) => !registeredScopes.includes(scope))
	);

	const canAuthorize = $derived(
		!requestError && client !== null && grantedScopes.length > 0
	);

	const authorizeUrl = $derived.by(() => {
		const q = new URLSearchParams();
		for (const key of FORWARDED_PARAMS) {
			const value = request[key];
			if (value) q.set(key, value);
		}
		return `/oauth2/authorize?${q.toString()}`;
	});

	function deny() {
		let target: URL;
		try {
			target = new URL(request.redirect_uri);
		} catch {
			return;
		}
		target.searchParams.set('error', 'access_denied');
		target.searchParams.set('error_description', 'The resource owner denied the request');
		if (request.state) target.searchParams.set('state', request.state);
		// Full navigation, so the app receives the denial at its redirect_uri.
		window.location.assign(target.toString());
	}

	function reload() {
		reloadNonce += 1;
	}
</script>

<svelte:head>
	<title>Authorize application — FreeMED EMR</title>
</svelte:head>

<div class="flex items-center justify-center min-h-[calc(100vh-8rem)]">
	<div class="w-full max-w-xl">
		<div class="bg-white rounded-xl shadow-lg border border-gray-200">
			<div class="px-6 py-5 border-b border-gray-100">
				<h1 class="text-lg font-semibold text-gray-900">Authorize application</h1>
				<p class="text-xs text-gray-500 mt-0.5">
					A third-party application is asking to access FreeMED on your behalf
				</p>
			</div>

			<div class="p-6 space-y-5">
				{#if requestError}
					<div
						class="bg-red-50 border border-red-200 rounded-md p-3 text-sm text-red-700"
						role="alert"
					>
						{requestError}
					</div>
					<p class="text-xs text-gray-500">
						Nothing was authorized. The application has to correct its request.
					</p>
				{:else if loadingClient}
					<LoadingSpinner message="Loading application details…" />
				{:else if clientError || !client}
					<ErrorBanner
						message={clientError || 'This application is not registered.'}
						onRetry={reload}
					/>
					<p class="text-xs text-gray-500">
						An application that is not registered, or has been deactivated, cannot be authorized.
					</p>
				{:else}
					<div class="flex items-start gap-3">
						<div
							class="w-10 h-10 shrink-0 rounded-lg bg-blue-50 text-blue-700 flex items-center justify-center text-base font-semibold"
							aria-hidden="true"
						>
							{(client.client_name || '?').charAt(0).toUpperCase()}
						</div>
						<div class="min-w-0">
							<p class="text-sm font-semibold text-gray-900 break-words">{client.client_name}</p>
							<p class="text-xs text-gray-500">
								Signed in as {auth.username ?? 'your staff account'}
							</p>
						</div>
					</div>

					<div>
						<h2 class="text-sm font-medium text-gray-900">This application wants to</h2>
						{#if requestedScopes.length === 0}
							<p class="text-sm text-gray-500 mt-1">
								The application is registered for no scopes, so nothing can be granted.
							</p>
						{:else}
							<ul class="mt-2 space-y-2">
								{#each requestedScopes as scope (scope)}
									{@const dropped = droppedScopes.includes(scope)}
									<li class="flex items-start gap-2">
										<span class={dropped ? 'text-gray-300' : 'text-green-600'} aria-hidden="true">●</span>
										<div class="min-w-0">
											<p class={dropped ? 'text-sm text-gray-400 line-through' : 'text-sm text-gray-700'}>
												{scopeLabel(scope)}
											</p>
											<p class="text-xs text-gray-400 font-mono break-all">{scope}</p>
											{#if dropped}
												<p class="text-xs text-amber-600">
													Not registered to this application — it will not be granted.
												</p>
											{/if}
										</div>
									</li>
								{/each}
							</ul>
							{#if droppedScopes.length > 0}
								<p class="text-xs text-gray-500 mt-2">
									{grantedScopes.length} of {requestedScopes.length} requested permissions will be granted.
								</p>
							{/if}
						{/if}
					</div>

					<dl class="text-xs text-gray-500 space-y-1 border-t border-gray-100 pt-4">
						<div class="flex gap-2">
							<dt class="w-36 shrink-0 text-gray-400">Application ID</dt>
							<dd class="font-mono break-all">{client.client_id}</dd>
						</div>
						<div class="flex gap-2">
							<dt class="w-36 shrink-0 text-gray-400">Sends results to</dt>
							<dd class="break-all">{redirectTarget}</dd>
						</div>
						{#if request.launch || request.patient}
							<div class="flex gap-2">
								<dt class="w-36 shrink-0 text-gray-400">Patient context</dt>
								<dd>Patient #{request.launch || request.patient}</dd>
							</div>
						{/if}
					</dl>

					<p class="text-xs text-gray-500">
						Approving returns you to the application with a one-time code. You can revoke access
						later by deactivating the application.
					</p>

					<!-- A real form POST, not fetch(): the server answers with a 302 to the
					     application's redirect_uri, which the browser must follow itself. -->
					<form method="POST" action={authorizeUrl} class="flex justify-end gap-3 pt-1">
						<input type="hidden" name="consent" value="approve" />
						<button
							type="button"
							onclick={deny}
							class="px-4 py-2 text-sm text-gray-600 hover:bg-gray-100 rounded-md transition-colors"
						>
							Deny
						</button>
						<button
							type="submit"
							disabled={!canAuthorize}
							class="px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-md hover:bg-blue-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
						>
							Authorize
						</button>
					</form>
					{#if !canAuthorize}
						<p class="text-xs text-amber-600 text-right">
							Nothing can be granted to this application, so authorization is disabled.
						</p>
					{/if}
				{/if}
			</div>
		</div>
	</div>
</div>
