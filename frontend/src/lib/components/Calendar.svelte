<script lang="ts">
	import { onMount } from 'svelte';
	import { Calendar } from '@fullcalendar/core';
	import dayGridPlugin from '@fullcalendar/daygrid';
	import timeGridPlugin from '@fullcalendar/timegrid';
	import interactionPlugin from '@fullcalendar/interaction';

		interface Props {
			events: any[] | ((info: any, success: any, failure: any) => void);
			onEventClick?: (info: any) => void;
			onEventDrop?: (info: any) => void;
			onEventResize?: (info: any) => void;
			onDateClick?: (info: any) => void;
			onReady?: (cal: Calendar) => void;
			editable?: boolean;
			height?: string;
		}

		let {
			events = [],
			onEventClick,
			onEventDrop,
			onEventResize,
			onDateClick,
			onReady,
			editable = true,
			height = 'auto',
		}: Props = $props();

	let calendarEl = $state<HTMLElement | null>(null);
	let calendar: Calendar | null = null;

	function buildOptions() {
		return {
			plugins: [dayGridPlugin, timeGridPlugin, interactionPlugin],
			initialView: 'timeGridWeek',
			weekends: false,
			businessHours: {
				daysOfWeek: [1, 2, 3, 4, 5],
				startTime: '08:00',
				endTime: '18:00',
			},
			eventLimit: 4,
			nowIndicator: true,
			navLinks: true,
			editable,
			height,
			headerToolbar: {
				left: 'prev,next today',
				center: 'title',
				right: 'dayGridMonth,timeGridWeek,timeGridDay',
			},
			events,
			eventClick: onEventClick,
			eventDrop: onEventDrop,
			eventResize: onEventResize,
			dateClick: onDateClick,
		};
	}

		onMount(() => {
			if (calendarEl) {
				calendar = new Calendar(calendarEl, buildOptions());
				calendar.render();
				onReady?.(calendar);
			}
			return () => {
				calendar?.destroy();
			};
		});
</script>

<div bind:this={calendarEl} class="fc-calendar"></div>

<style>
	:global(.fc) {
		font-size: 0.875rem;
		line-height: 1.25rem;
	}
	:global(.fc-toolbar-title) {
		font-size: 1.125rem;
		font-weight: 600;
	}
</style>
