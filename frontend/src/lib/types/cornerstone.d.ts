/**
 * Ambient type declarations for the cornerstone v1 ("legacy") DICOM viewer stack.
 *
 * Neither `cornerstone-core@2.6.1` nor `cornerstone-wado-image-loader@4.13.2`
 * ships TypeScript definitions — both publish webpack UMD bundles and nothing
 * else. Declaring the modules here (rather than letting the compiler infer types
 * from the minified bundles with `allowJs`) keeps the 1.4 MB / 2.6 MB bundles out
 * of the type-check program entirely and pins the surface we actually use.
 *
 * Both packages are reached only through dynamic `import()` inside `onMount`, so
 * they never execute during the adapter-static SSR/prerender build.
 *
 * This file must stay a *global* script (no top-level `import`/`export`) for the
 * `declare module` blocks below to be ambient module declarations.
 */

declare module 'cornerstone-core' {
	export interface CornerstoneImage {
		imageId: string;
		rows: number;
		columns: number;
		height: number;
		width: number;
		color: boolean;
		invert: boolean;
		minPixelValue: number;
		maxPixelValue: number;
		slope: number;
		intercept: number;
		windowCenter: number | number[];
		windowWidth: number | number[];
		columnPixelSpacing: number | null;
		rowPixelSpacing: number | null;
		sizeInBytes: number;
		getPixelData(): ArrayBufferView;
	}

	export interface CornerstoneViewport {
		scale: number;
		translation: { x: number; y: number };
		voi?: {
			windowWidth: number;
			windowCenter: number;
			windowWidthDelta?: number;
			windowCenterDelta?: number;
		};
		invert: boolean;
		pixelReplication: boolean;
		rotation: number;
		hflip: boolean;
		vflip: boolean;
		modalityLUT: unknown;
		voiLUT: unknown;
		colormap: string | undefined;
		labelmap: boolean;
		displayedArea: {
			tlhc: { x: number; y: number };
			brhc: { x: number; y: number };
			rowPixelSpacing: number;
			columnPixelSpacing: number;
			direction: string;
		};
	}

	export interface CornerstoneEnabledElement {
		element: HTMLElement;
		canvas: HTMLCanvasElement;
		image?: CornerstoneImage;
		viewport?: CornerstoneViewport;
	}

	export interface CornerstoneImageRenderedEventDetail {
		element: HTMLElement;
		viewport: CornerstoneViewport;
		image: CornerstoneImage;
		enabledElement: CornerstoneEnabledElement;
		canvasContext: CanvasRenderingContext2D;
		renderTimeInMs?: number;
	}

	export interface CornerstoneCore {
		EVENTS: Record<string, string>;
		enable(element: HTMLElement, options?: { renderer?: 'webgl' | 'canvas' }): void;
		disable(element: HTMLElement): void;
		displayImage(element: HTMLElement, image: CornerstoneImage): void;
		getImage(element: HTMLElement): CornerstoneImage | undefined;
		getViewport(element: HTMLElement): CornerstoneViewport | undefined;
		setViewport(element: HTMLElement, viewport: CornerstoneViewport): void;
		updateImage(element: HTMLElement, invalidated?: boolean): void;
		/** Draws the enabled element synchronously, without waiting for the frame loop. */
		draw(element: HTMLElement): void;
		reset(element: HTMLElement): void;
		invalidate(element: HTMLElement, ...rest: unknown[]): void;
		purgeCache(): void;
		removeImageLoadObject(imageId: string): void;
		triggerEvent(element: HTMLElement, type: string, detail?: unknown): boolean;
		loadImage(imageId: string, options?: unknown): Promise<CornerstoneImage>;
		loadAndCacheImage(imageId: string, options?: unknown): Promise<CornerstoneImage>;
		getEnabledElement(element: HTMLElement): CornerstoneEnabledElement | undefined;
	}

	const cornerstone: CornerstoneCore;
	export default cornerstone;
}

declare module 'cornerstone-wado-image-loader' {
	export interface WadoImageLoaderConfig {
		beforeSend?: (
			xhr: XMLHttpRequest,
			imageId?: string,
			headers?: Record<string, string>,
			params?: unknown
		) => void;
		beforeProcessing?: (xhr: XMLHttpRequest) => Promise<ArrayBuffer>;
		errorInterceptor?: (error: {
			status?: number;
			request?: unknown;
			response?: unknown;
		}) => void;
		strict?: boolean;
		[key: string]: unknown;
	}

	export interface WadoFileManager {
		add(file: File | Blob): string;
		get(index: number): File | Blob | undefined;
		remove(index: number): void;
		purge(): void;
	}

	export interface WadoDataSetCacheManager {
		unload(uri: string): void;
		purge(): void;
	}

	export interface WadoImageLoader {
		external: {
			cornerstone: unknown;
			dicomParser: unknown;
		};
		configure(options: WadoImageLoaderConfig): void;
		webWorkerManager: { initialize(options?: unknown): void };
		version: string;
		wadouri: {
			fileManager: WadoFileManager;
			dataSetCacheManager: WadoDataSetCacheManager;
			loadImage(imageId: string): Promise<unknown>;
			parseImageId(imageId: string): { scheme: string; url: string; frame?: number };
		};
	}

	const cornerstoneWADOImageLoader: WadoImageLoader;
	export default cornerstoneWADOImageLoader;
}

/**
 * The package also ships a build with web workers compiled out. It is preferred
 * here: the worker variant resolves its decoder script from a webpack public path
 * that does not exist in a Vite/SvelteKit build, so the decode promise never
 * settles and the viewer hangs on "Loading DICOM instance…" forever. The
 * no-worker build decodes on the main thread, exposes the same API, and needs no
 * worker asset to be copied or served.
 */
declare module 'cornerstone-wado-image-loader/dist/cornerstoneWADOImageLoaderNoWebWorkers.bundle.min.js' {
	import type { WadoImageLoader } from 'cornerstone-wado-image-loader';
	const loader: WadoImageLoader;
	export default loader;
}
