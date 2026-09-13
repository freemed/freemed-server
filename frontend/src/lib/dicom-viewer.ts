/**
 * Lazy initialiser for the cornerstone v1 DICOM viewer stack
 * (`cornerstone-core` + `cornerstone-wado-image-loader`).
 *
 * Why dynamic imports: both packages are webpack UMD bundles that touch
 * `window`/`document` at module scope. The SvelteKit app is built with
 * `adapter-static` in SPA mode, which still renders every route in Node during
 * the build, so a top-level import would break the build. They are therefore
 * pulled in on demand, from `onMount` only, and cached in a single promise.
 *
 * Type declarations for both packages live in `$lib/types/cornerstone.d.ts`.
 */
import type { CornerstoneCore } from 'cornerstone-core';
import type { WadoImageLoader } from 'cornerstone-wado-image-loader';

export interface CornerstoneStack {
	cornerstone: CornerstoneCore;
	loader: WadoImageLoader;
}

let stackPromise: Promise<CornerstoneStack> | null = null;
let parkedScript: HTMLScriptElement | null = null;

function lastScriptHasSrc(): boolean {
	const scripts = document.getElementsByTagName('script');
	if (scripts.length === 0) return false;
	const last = scripts[scripts.length - 1] as HTMLScriptElement;
	return !!last.src;
}

/**
 * cornerstone-wado-image-loader is a webpack UMD bundle built with
 * `publicPath: 'auto'`. Its runtime derives the public path from the LAST
 * `<script>` element in document order and throws
 * "Automatic publicPath is not supported in this browser" when that element has
 * no `src`. SvelteKit's built `index.html` boots the SPA from an *inline* script
 * inside `<body>`, so the check fails the first time the loader is imported.
 * The public path is never used here (the loader's codecs are inlined as `data:`
 * URLs and no webpack chunk or worker is ever requested), so the fix is simply
 * to make the check pass: park a src'd no-op script at the end of the document
 * while the loader initialises, then take it back out.
 */
function satisfyWebpackAutoPublicPath(): void {
	if (typeof document === 'undefined') return;
	if (lastScriptHasSrc()) return;

	const script = document.createElement('script');
	// A data: URL executes nothing and costs no network request.
	script.src = 'data:text/javascript,';
	script.async = true;
	// Document order matters: <body> follows <head>, and SvelteKit's inline boot
	// script lives in <body>, so the no-op has to go there (falling back to
	// <html> if body is not present yet).
	(document.body ?? document.documentElement).appendChild(script);
	if (!lastScriptHasSrc()) document.documentElement.appendChild(script);
	parkedScript = script;
}

function releaseParkedScript(): void {
	if (parkedScript?.parentNode) parkedScript.parentNode.removeChild(parkedScript);
	parkedScript = null;
}

/**
 * Loads cornerstone + the WADO-RS loader once per page load and wires the loader
 * to its two required dependencies. Safe to call from any number of viewers; the
 * work happens once.
 */
export function initCornerstone(): Promise<CornerstoneStack> {
	if (stackPromise) return stackPromise;

	stackPromise = (async (): Promise<CornerstoneStack> => {
		satisfyWebpackAutoPublicPath();

		// The default entrypoint is the WEB-WORKER build. Its decoder script is
		// resolved from a webpack public path that does not exist in a Vite build,
		// so `loadAndCacheImage` never settles and the viewer hangs on "Loading
		// DICOM instance…" with no error surfaced. The package's NoWebWorkers
		// build decodes on the main thread, exposes the identical API, and needs no
		// worker asset — so it is imported explicitly instead.
		const [coreModule, loaderModule, parserModule] = await Promise.all([
			import('cornerstone-core'),
			import('cornerstone-wado-image-loader/dist/cornerstoneWADOImageLoaderNoWebWorkers.bundle.min.js'),
			import('dicom-parser')
		]);
		releaseParkedScript();

		// UMD -> ESM interop: Vite hands back the `module.exports` object as
		// `default`. Fall back to the namespace itself in case a bundler exposes
		// the exports as named bindings instead.
		const core = coreModule as unknown as { default?: CornerstoneCore } & Partial<CornerstoneCore>;
		const cornerstone = (core.default ?? core) as CornerstoneCore;
		if (!cornerstone || typeof cornerstone.enable !== 'function') {
			throw new Error('cornerstone-core loaded without its expected API');
		}

		const wado = loaderModule as unknown as { default?: WadoImageLoader };
		const loader = (wado.default ?? (wado as unknown as WadoImageLoader)) as WadoImageLoader;
		if (!loader || !loader.wadouri) {
			throw new Error('cornerstone-wado-image-loader loaded without its expected API');
		}

		const parser = parserModule as unknown as { default?: unknown };
		const dicomParser = parser.default ?? parser;

		// Assigning `external.cornerstone` also registers the wadouri/dicomfile
		// image loaders and the metadata provider against cornerstone.
		loader.external.cornerstone = cornerstone;
		loader.external.dicomParser = dicomParser;

		// Instances are served same-origin behind the httpOnly staff JWT cookie,
		// which the browser attaches automatically; `withCredentials` keeps that
		// true if a deployment ever puts the API on another origin.
		loader.configure({
			beforeSend: (xhr: XMLHttpRequest) => {
				xhr.withCredentials = true;
			}
		});

		return { cornerstone, loader };
	})();

	return stackPromise;
}
