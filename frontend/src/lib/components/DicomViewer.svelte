<script lang="ts">
	/**
	 * Minimal single-instance DICOM viewer.
	 *
	 * Pixels come from the patient-scoped WADO-RS instance endpoint
	 * (`GET /api/dicom/patient/:id/studies/:studyUID/series/:seriesUID/instances/:sopUID`,
	 * served behind the staff JWT cookie). The patient context is mandatory: the
	 * server refuses to serve an instance that does not belong to `patientId`,
	 * so the viewer must be given the patient it is being displayed for.
	 * The bytes are fetched once through the shared `$lib/api` helper and handed
	 * to cornerstone's WADO loader in memory, so the instance is transferred
	 * exactly once.
	 *
	 * Deliberately single-instance / single-frame: no series stack, no
	 * measurements, no cine. Window/level, zoom and pan are hand-rolled against
	 * the cornerstone viewport (cornerstone-tools is not a dependency).
	 */
	import { onMount, tick } from 'svelte';
	import type {
		CornerstoneCore,
		CornerstoneImage,
		CornerstoneImageRenderedEventDetail,
		CornerstoneViewport
	} from 'cornerstone-core';
	import type { WadoImageLoader } from 'cornerstone-wado-image-loader';
	import { initCornerstone } from '$lib/dicom-viewer';
	import { api } from '$lib/api';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import ErrorBanner from '$lib/components/ErrorBanner.svelte';

	interface Props {
		/** Owning patient. The WADO-RS route is patient-scoped, so this is required. */
		patientId: string;
		studyUid: string;
		seriesUid: string;
		sopUid: string;
		/** Human label for the header, e.g. the study description. */
		label?: string;
		/** Same-origin path that streams the raw stored bytes, for the fallback link. */
		rawPath?: string;
		onClose?: () => void;
	}

	let { patientId, studyUid, seriesUid, sopUid, label = '', rawPath = '', onClose }: Props =
		$props();

	let element = $state<HTMLDivElement | undefined>(undefined);
	let loading = $state(true);
	let error = $state('');
	let windowWidth = $state(0);
	let windowCenter = $state(0);
	let scale = $state(1);
	let dimensions = $state('');

	const uidsPresent = $derived(!!(patientId && studyUid && seriesUid && sopUid));
	const wadoPath = $derived(
		`/dicom/patient/${encodeURIComponent(patientId)}/studies/${encodeURIComponent(studyUid)}/series/${encodeURIComponent(seriesUid)}/instances/${encodeURIComponent(sopUid)}`
	);

	// Non-reactive handles owned by this instance; never rendered.
	let cornerstone: CornerstoneCore | undefined;
	let loader: WadoImageLoader | undefined;
	let imageId = '';
	let image: CornerstoneImage | undefined;
	let enabled = false;
	let listenersAttached = false;
	// Declared as state because it gates the retry button in the template.
	let disposed = $state(false);
	let dragging = false;
	let lastX = 0;
	let lastY = 0;

	onMount(() => {
		void load();
		return () => {
			disposed = true;
			dispose();
		};
	});

	async function load() {
		if (disposed || !uidsPresent) {
			loading = false;
			return;
		}
		error = '';
		loading = true;

		let fetched = false;
		let dicomPreamble = false;

		try {
			const stack = await initCornerstone();
			cornerstone = stack.cornerstone;
			loader = stack.loader;
			if (disposed) return;

			await tick();
			const host = element;
			if (!host) throw new Error('Viewer area was not mounted.');
			await ensureSized(host);
			if (disposed) return;

			if (!enabled) {
				cornerstone.enable(host);
				enabled = true;
				// cornerstone appends its own <canvas>; normalise its layout box.
				const canvas = host.querySelector('canvas');
				if (canvas) canvas.style.display = 'block';
				attachListeners(host);
			}

			const bytes = await api.getBinary(wadoPath);
			if (disposed) return;
			fetched = true;
			if (bytes.byteLength === 0) {
				throw new Error('The stored object contains no data.');
			}
			dicomPreamble = hasDicmPreamble(bytes);

			// `dicomfile:` image ids make the loader read the instance from memory
			// rather than refetching the URL it just came from.
			const file = new File([bytes], `${sopUid || 'instance'}.dcm`, {
				type: 'application/dicom'
			});
			imageId = loader.wadouri.fileManager.add(file);

			const loaded = await cornerstone.loadAndCacheImage(imageId);
			if (disposed) return;
			cornerstone.displayImage(host, loaded);
			image = loaded;
			dimensions = `${loaded.columns} × ${loaded.rows}`;
			// displayImage only marks the element invalid and lets the animation
			// frame loop paint. Draw synchronously so the first paint is not at the
			// mercy of the frame loop (throttled or suspended in background tabs).
			drawNow();
			syncReadout(loaded, cornerstone.getViewport(host));
		} catch (e) {
			if (disposed) return;
			error = describeFailure(e, fetched, dicomPreamble);
		} finally {
			if (!disposed) loading = false;
		}
	}

	/**
	 * Best-effort human text for an unknown thrown value. cornerstone rejects its
	 * load promises with plain objects rather than Errors, so naively stringifying
	 * the cause rendered "[object Object]" to the user and hid the real reason.
	 */
	function errorText(cause: unknown): string {
		if (cause instanceof Error) return cause.message;
		if (typeof cause === 'string') return cause;
		if (cause && typeof cause === 'object') {
			const o = cause as Record<string, unknown>;
			for (const key of ['message', 'error', 'reason', 'statusText', 'description', 'exception']) {
				const v = o[key];
				if (typeof v === 'string' && v.trim()) return v;
			}
			try {
				const json = JSON.stringify(cause);
				if (json && json !== '{}' && json !== 'null') return json;
			} catch {
				// circular or otherwise unserialisable — fall through
			}
			const keys = Object.keys(o);
			if (keys.length) return `unrecognised failure (${keys.join(', ')})`;
		}
		return String(cause);
	}

	function describeFailure(cause: unknown, fetched: boolean, dicomPreamble: boolean): string {
		const raw = errorText(cause);
		if (!fetched) return raw;
		if (!dicomPreamble) {
			return `${raw} — the retrieved object does not carry a DICOM file preamble (no "DICM" marker at byte 128), so it is not a readable DICOM image. It was stored without usable header metadata.`;
		}
		return `${raw} — the bytes were retrieved but could not be decoded; the instance is either corrupt or uses a transfer syntax this viewer cannot render.`;
	}

	function dispose() {
		const host = element;
		if (host) detachListeners(host);
		if (host && enabled && cornerstone) {
			try {
				cornerstone.disable(host);
			} catch {
				// already disabled / torn down by cornerstone
			}
		}
		enabled = false;
		if (cornerstone && imageId) {
			try {
				cornerstone.removeImageLoadObject(imageId);
			} catch {
				// not cached
			}
		}
		// Purge the image cache. This is required, not tidiness: `fileManager.purge()`
		// below resets the loader's index counter to 0, so the next viewer instance
		// asks for `dicomfile:0` again — the same imageId as the study just viewed.
		// Without this purge cornerstone serves the cached image and the viewer
		// silently displays the PREVIOUS patient's/instance's pixels under the new
		// study's header. Verified live before the fix (a corrupt instance rendered
		// the previous study's 64x64 image). Single-instance viewer, so purging the
		// whole cache costs nothing.
		try {
			cornerstone?.purgeCache();
		} catch {
			// nothing cached
		}
		// Release the in-memory copy of the instance the loader is holding.
		try {
			loader?.wadouri.dataSetCacheManager.purge();
			loader?.wadouri.fileManager.purge();
		} catch {
			// nothing cached
		}
		imageId = '';
		image = undefined;
	}

	/* ------------------------------------------------------------------ */
	/* Viewport controls (window/level, zoom, pan)                         */
	/* ------------------------------------------------------------------ */

	function viewport(): CornerstoneViewport | undefined {
		if (!cornerstone || !element || !enabled) return undefined;
		return cornerstone.getViewport(element);
	}

	/**
	 * Paints immediately instead of waiting for cornerstone's animation frame
	 * loop. Used for the first paint and for discrete toolbar actions; continuous
	 * input (drag, wheel) intentionally keeps frame-loop coalescing.
	 */
	function drawNow() {
		if (!cornerstone || !element || !enabled) return;
		try {
			cornerstone.draw(element);
		} catch {
			// element was torn down between render and draw
		}
	}

	function apply(mutate: (vp: CornerstoneViewport) => void) {
		const host = element;
		if (!cornerstone || !host) return;
		const vp = viewport();
		if (!vp) return;
		mutate(vp);
		cornerstone.setViewport(host, vp);
		drawNow();
		syncReadout(image, vp);
	}

	/**
	 * @param fromControl false for continuous input (wheel), which relies on the
	 * animation frame loop to coalesce repaints.
	 */
	function changeZoom(factor: number, fromControl = true) {
		const host = element;
		if (!host) return;
		const vp = viewport();
		if (!vp) return;
		const next = Math.min(20, Math.max(0.05, vp.scale * factor));
		if (next === vp.scale) return;
		// Keep the centre of the canvas fixed on the same image pixel.
		const ratio = next / vp.scale;
		const cx = host.clientWidth / 2;
		const cy = host.clientHeight / 2;
		vp.translation.x = cx - ratio * (cx - vp.translation.x);
		vp.translation.y = cy - ratio * (cy - vp.translation.y);
		vp.scale = next;
		cornerstone?.setViewport(host, vp);
		if (fromControl) drawNow();
		syncReadout(image, vp);
	}

	function changeWindowWidth(factor: number) {
		apply((vp) => {
			const voi = ensureVoi(vp);
			voi.windowWidth = Math.max(1, voi.windowWidth * factor);
		});
	}

	function changeWindowCenter(deltaFraction: number) {
		apply((vp) => {
			const voi = ensureVoi(vp);
			const delta = Math.max(1, voi.windowWidth * deltaFraction);
			voi.windowCenter += delta;
		});
	}

	function resetView() {
		const host = element;
		if (!cornerstone || !host) return;
		cornerstone.reset(host);
		drawNow();
		scale = 1;
		syncReadout(image, cornerstone.getViewport(host));
	}

	function ensureVoi(vp: CornerstoneViewport) {
		if (!vp.voi) {
			const min = image?.minPixelValue ?? 0;
			const max = image?.maxPixelValue ?? 255;
			const width = Math.max(1, max - min);
			vp.voi = { windowWidth: width, windowCenter: Math.round(min + width / 2) };
		}
		return vp.voi;
	}

	function syncReadout(img: CornerstoneImage | undefined, vp: CornerstoneViewport | undefined) {
		if (!vp) return;
		scale = vp.scale;
		if (vp.voi) {
			windowWidth = Math.round(vp.voi.windowWidth);
			windowCenter = Math.round(vp.voi.windowCenter);
			return;
		}
		if (img) {
			const min = img.minPixelValue ?? 0;
			const max = img.maxPixelValue ?? 255;
			windowWidth = Math.round(firstNumber(img.windowWidth, Math.max(1, max - min)));
			windowCenter = Math.round(firstNumber(img.windowCenter, min + (max - min) / 2));
		}
	}

	function firstNumber(value: number | number[] | undefined, fallback: number): number {
		if (Array.isArray(value)) return value.length > 0 ? Number(value[0]) : fallback;
		if (typeof value === 'number' && Number.isFinite(value)) return value;
		return fallback;
	}

	/* ------------------------------------------------------------------ */
	/* Element listeners                                                   */
	/* ------------------------------------------------------------------ */

	function onImageRendered(event: Event) {
		const detail = (event as unknown as { detail?: CornerstoneImageRenderedEventDetail }).detail;
		if (!detail) return;
		syncReadout(detail.image, detail.viewport);
	}

	function onPointerDown(event: PointerEvent) {
		if (event.button !== 0 || loading || error) return;
		const host = element;
		if (!host || !enabled) return;
		dragging = true;
		lastX = event.clientX;
		lastY = event.clientY;
		try {
			host.setPointerCapture(event.pointerId);
		} catch {
			// capture unsupported
		}
	}

	function onPointerMove(event: PointerEvent) {
		if (!dragging || !cornerstone || !element) return;
		const dx = event.clientX - lastX;
		const dy = event.clientY - lastY;
		lastX = event.clientX;
		lastY = event.clientY;
		const vp = cornerstone.getViewport(element);
		if (!vp) return;
		vp.translation.x += dx;
		vp.translation.y += dy;
		cornerstone.setViewport(element, vp);
	}

	function onPointerUp(event: PointerEvent) {
		if (!dragging) return;
		dragging = false;
		try {
			element?.releasePointerCapture(event.pointerId);
		} catch {
			// capture was never taken
		}
	}

	function onWheel(event: WheelEvent) {
		if (loading || error || !enabled) return;
		event.preventDefault();
		changeZoom(event.deltaY < 0 ? 1.1 : 1 / 1.1, false);
	}

	function attachListeners(host: HTMLElement) {
		if (listenersAttached || !cornerstone) return;
		host.addEventListener(cornerstone.EVENTS.IMAGE_RENDERED, onImageRendered);
		host.addEventListener('pointerdown', onPointerDown);
		host.addEventListener('pointermove', onPointerMove);
		host.addEventListener('pointerup', onPointerUp);
		host.addEventListener('pointercancel', onPointerUp);
		host.addEventListener('wheel', onWheel, { passive: false });
		listenersAttached = true;
	}

	function detachListeners(host: HTMLElement) {
		if (!listenersAttached) return;
		if (cornerstone) host.removeEventListener(cornerstone.EVENTS.IMAGE_RENDERED, onImageRendered);
		host.removeEventListener('pointerdown', onPointerDown);
		host.removeEventListener('pointermove', onPointerMove);
		host.removeEventListener('pointerup', onPointerUp);
		host.removeEventListener('pointercancel', onPointerUp);
		host.removeEventListener('wheel', onWheel);
		listenersAttached = false;
	}

	function onKeydown(event: KeyboardEvent) {
		if (event.key === 'Escape') onClose?.();
	}

	/* ------------------------------------------------------------------ */
	/* Helpers                                                             */
	/* ------------------------------------------------------------------ */

	/** Cornerstone needs a laid-out element with a non-zero box. */
	async function ensureSized(host: HTMLElement) {
		for (let attempt = 0; attempt < 5; attempt++) {
			if (host.clientWidth > 0 && host.clientHeight > 0) return;
			await new Promise<void>((resolve) => requestAnimationFrame(() => resolve()));
		}
		throw new Error('The viewer area has no size, so the image cannot be rendered.');
	}

	/** DICOM part-10 files carry "DICM" at byte 128. */
	function hasDicmPreamble(bytes: ArrayBuffer): boolean {
		if (bytes.byteLength < 132) return false;
		const magic = new Uint8Array(bytes, 128, 4);
		return magic[0] === 0x44 && magic[1] === 0x49 && magic[2] === 0x43 && magic[3] === 0x4d;
	}

	const controlsDisabled = $derived(loading || !!error);
</script>

<svelte:window onkeydown={onKeydown} />

<div
	class="fixed inset-0 z-50 flex items-center justify-center bg-black/70 p-4"
	role="dialog"
	aria-modal="true"
	aria-label="DICOM image viewer"
	tabindex="-1"
>
	<div class="bg-white rounded-lg shadow-xl w-full max-w-5xl flex flex-col max-h-[92vh] overflow-hidden">
		<!-- Header -->
		<div class="px-4 py-3 border-b border-gray-200 flex items-start justify-between gap-4">
			<div class="min-w-0">
				<h2 class="text-sm font-semibold text-gray-900 truncate">{label || 'DICOM instance'}</h2>
				<p class="text-xs text-gray-500 font-mono truncate">
					{sopUid || 'no SOP Instance UID stored'}
				</p>
			</div>
			<div class="flex items-center gap-2 shrink-0">
				{#if rawPath}
					<a
						href={rawPath}
						class="px-3 py-1.5 text-xs font-medium text-gray-700 border border-gray-300 rounded-md hover:bg-gray-50 transition-colors"
					>
						Download
					</a>
				{/if}
				<button
					type="button"
					onclick={() => onClose?.()}
					class="px-3 py-1.5 text-xs font-medium text-white bg-gray-700 rounded-md hover:bg-gray-800 transition-colors"
				>
					Close
				</button>
			</div>
		</div>

		<!-- Toolbar -->
		<div
			class="px-4 py-2 border-b border-gray-200 bg-gray-50 flex flex-wrap items-center gap-x-5 gap-y-2 text-xs text-gray-600"
		>
			<div class="flex items-center gap-1">
				<span class="font-medium text-gray-500 uppercase tracking-wide mr-1">Zoom</span>
				<button
					type="button"
					disabled={controlsDisabled}
					onclick={() => changeZoom(1 / 1.25)}
					class="px-2 py-0.5 rounded border border-gray-300 bg-white hover:bg-gray-100 disabled:opacity-40 transition-colors"
				>
					&minus;
				</button>
				<span class="w-12 text-center tabular-nums">{Math.round(scale * 100)}%</span>
				<button
					type="button"
					disabled={controlsDisabled}
					onclick={() => changeZoom(1.25)}
					class="px-2 py-0.5 rounded border border-gray-300 bg-white hover:bg-gray-100 disabled:opacity-40 transition-colors"
				>
					+
				</button>
			</div>

			<div class="flex items-center gap-1">
				<span class="font-medium text-gray-500 uppercase tracking-wide mr-1">Window</span>
				<button
					type="button"
					disabled={controlsDisabled}
					onclick={() => changeWindowWidth(1 / 1.1)}
					class="px-2 py-0.5 rounded border border-gray-300 bg-white hover:bg-gray-100 disabled:opacity-40 transition-colors"
					title="Decrease window width"
				>
					W&minus;
				</button>
				<button
					type="button"
					disabled={controlsDisabled}
					onclick={() => changeWindowWidth(1.1)}
					class="px-2 py-0.5 rounded border border-gray-300 bg-white hover:bg-gray-100 disabled:opacity-40 transition-colors"
					title="Increase window width"
				>
					W+
				</button>
				<button
					type="button"
					disabled={controlsDisabled}
					onclick={() => changeWindowCenter(-0.05)}
					class="px-2 py-0.5 rounded border border-gray-300 bg-white hover:bg-gray-100 disabled:opacity-40 transition-colors"
					title="Decrease window centre"
				>
					L&minus;
				</button>
				<button
					type="button"
					disabled={controlsDisabled}
					onclick={() => changeWindowCenter(0.05)}
					class="px-2 py-0.5 rounded border border-gray-300 bg-white hover:bg-gray-100 disabled:opacity-40 transition-colors"
					title="Increase window centre"
				>
					L+
				</button>
				<span class="ml-1 tabular-nums">W {windowWidth} / L {windowCenter}</span>
			</div>

			<button
				type="button"
				disabled={controlsDisabled}
				onclick={resetView}
				class="px-2 py-0.5 rounded border border-gray-300 bg-white hover:bg-gray-100 disabled:opacity-40 transition-colors"
			>
				Reset
			</button>

			{#if dimensions}
				<span class="text-gray-400">{dimensions} px</span>
			{/if}
			<span class="text-gray-400 ml-auto hidden sm:inline">Drag to pan · scroll to zoom</span>
		</div>

		<!-- Image area -->
		<div class="relative bg-black flex-1 min-h-[55vh]" style="min-height: 55vh;">
			<!-- The host geometry is set inline (not via utilities) so the viewer still
			     renders if the Tailwind stylesheet is absent from the build. -->
			<div
				bind:this={element}
				class="w-full h-[55vh]"
				style="display: block; width: 100%; height: 55vh; min-height: 240px;"
			></div>

			{#if loading}
				<div class="absolute inset-0 flex items-center justify-center bg-black/70">
					<LoadingSpinner message="Loading DICOM instance…" />
				</div>
			{/if}
		</div>

		<!-- Error / explanation -->
		{#if !uidsPresent}
			<div class="p-4">
				<ErrorBanner
					message="This object cannot be previewed: it has no Study / Series / SOP Instance UID stored, or no patient context was supplied, so it is not addressable through the patient-scoped WADO-RS instance endpoint."
				/>
			</div>
		{:else if error}
			<div class="p-4">
				<ErrorBanner message={error} onRetry={disposed ? undefined : load} />
				{#if rawPath}
					<p class="text-xs text-gray-500 mt-3 text-center">
						The stored bytes can still be retrieved directly:
						<a href={rawPath} class="text-blue-600 hover:text-blue-800 font-medium">download raw file</a>
					</p>
				{/if}
			</div>
		{/if}
	</div>
</div>
