import type { PlaybackSession, RecordingDaySource } from "@nvr/api-client";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useEffect, useMemo, useState } from "react";
import { cameraKeys, getCamera } from "@/features/cameras/api";
import {
	listCameraRecordings,
	listRecordingDays,
	recordingKeys,
	requestRecordingPlayback,
} from "@/features/recordings/api";
import {
	type CalendarMonth,
	calendarGridRange,
	calendarMonthForDate,
	calendarMonthForValue,
	localDateValue,
	nearestCoverageTime,
	rangeForLocalDate,
	resolvedTimeZone,
} from "@/features/recordings/date";
import { resolveMediaUrl } from "@/lib/api/config";

export function useRecordingHistory(cameraId: string, active: boolean) {
	const queryClient = useQueryClient();
	const today = localDateValue(new Date());
	const [selectedDay, setSelectedDay] = useState(today);
	const [month, setMonth] = useState<CalendarMonth>(() => calendarMonthForDate(new Date()));
	const [selectedTime, setSelectedTime] = useState<string | null>(null);
	const timeZone = resolvedTimeZone();
	const dayRange = rangeForLocalDate(selectedDay);
	const gridRange = calendarGridRange(month);

	const cameraQuery = useQuery({
		queryKey: cameraKeys.detail(cameraId),
		queryFn: () => getCamera(cameraId),
		enabled: Boolean(cameraId),
	});
	const daysQuery = useQuery({
		queryKey: recordingKeys.days(
			cameraId,
			gridRange.from.toISOString(),
			gridRange.to.toISOString(),
			timeZone,
		),
		queryFn: () =>
			listRecordingDays(
				cameraId,
				gridRange.from.toISOString(),
				gridRange.to.toISOString(),
				timeZone,
			),
		enabled: Boolean(cameraId),
	});
	const recordingsQuery = useQuery({
		queryKey: recordingKeys.list(
			cameraId,
			dayRange?.from.toISOString() ?? "",
			dayRange?.to.toISOString() ?? "",
		),
		queryFn: () =>
			listCameraRecordings(
				cameraId,
				dayRange?.from.toISOString() ?? "",
				dayRange?.to.toISOString() ?? "",
			),
		enabled: Boolean(cameraId && dayRange),
	});
	const playbackKey = recordingKeys.playback(cameraId, selectedTime ?? "");
	const playbackQuery = useQuery({
		queryKey: playbackKey,
		queryFn: () => requestRecordingPlayback(cameraId, selectedTime ?? ""),
		enabled: active && Boolean(cameraId && selectedTime),
		staleTime: 45_000,
	});
	const continuePlaybackMutation = useMutation({
		mutationFn: ({ cameraId: nextCameraId, nextStart }: { cameraId: string; nextStart: string }) =>
			requestRecordingPlayback(nextCameraId, nextStart),
		onSuccess: (nextSession, variables) => {
			queryClient.setQueryData(
				recordingKeys.playback(variables.cameraId, variables.nextStart),
				nextSession,
			);
			setSelectedTime(variables.nextStart);
		},
	});

	const coverage = recordingsQuery.data?.coverage ?? [];
	useEffect(() => {
		if (selectedTime || !dayRange || coverage.length === 0) {
			return;
		}
		const firstPlayable = nearestCoverageTime(coverage, dayRange.from);
		if (firstPlayable) {
			setSelectedTime(firstPlayable.toISOString());
		}
	}, [coverage, dayRange, selectedTime]);

	const availability = useMemo(
		() =>
			new Map<string, RecordingDaySource>(
				(daysQuery.data ?? []).map((day) => [day.date, day.source]),
			),
		[daysQuery.data],
	);
	const playback = resolvePlayback(playbackQuery.data);

	const selectDay = (value: string) => {
		setSelectedDay(value);
		setSelectedTime(null);
		continuePlaybackMutation.reset();
		const selectedMonth = calendarMonthForValue(value);
		if (selectedMonth) {
			setMonth(selectedMonth);
		}
	};
	const seek = (time: Date) => {
		const playable = nearestCoverageTime(coverage, time);
		if (!playable) {
			return;
		}
		setSelectedTime(playable.toISOString());
		continuePlaybackMutation.reset();
	};
	const continuePlayback = () => {
		const session = playbackQuery.data;
		if (!session?.nextRecordingStart || continuePlaybackMutation.isPending) {
			return;
		}
		continuePlaybackMutation.mutate({
			cameraId: session.cameraId,
			nextStart: session.nextRecordingStart,
		});
	};
	const retryPlayback = () => {
		const failedContinuation = continuePlaybackMutation.variables;
		if (continuePlaybackMutation.isError && failedContinuation) {
			continuePlaybackMutation.reset();
			continuePlaybackMutation.mutate(failedContinuation);
			return;
		}
		continuePlaybackMutation.reset();
		return playbackQuery.refetch();
	};

	return {
		camera: cameraQuery.data,
		cameraError: cameraQuery.error,
		cameraPending: cameraQuery.isPending,
		retryCamera: cameraQuery.refetch,
		selectedDay,
		maxDate: today,
		month,
		setMonth,
		selectDay,
		availability,
		daysLoading: daysQuery.isPending,
		daysError: daysQuery.isError,
		dayRange,
		coverage,
		recordingsPending: recordingsQuery.isPending,
		recordingsError: recordingsQuery.error,
		retryRecordings: recordingsQuery.refetch,
		selectedTime: selectedTime ? new Date(selectedTime) : null,
		seek,
		playbackSession: playbackQuery.data,
		playbackUrl: playback.url,
		playbackUrlError: playback.error,
		playbackPending: playbackQuery.isPending || continuePlaybackMutation.isPending,
		playbackError: continuePlaybackMutation.error ?? playbackQuery.error,
		retryPlayback,
		continuePlayback,
	};
}

function resolvePlayback(session: PlaybackSession | undefined): { url?: string; error?: Error } {
	if (!session) {
		return {};
	}
	try {
		return { url: resolveMediaUrl(session.playbackUrl, session.token) };
	} catch (cause) {
		return { error: cause instanceof Error ? cause : new Error("Playback URL is invalid") };
	}
}
