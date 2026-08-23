import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { cameraKeys, getCamera, getPlayback } from "@/features/cameras/api";
import { resolvePlayback } from "@/features/cameras/media";
import { acknowledgeEvent, eventKeys, getEvent } from "@/features/events/api";

export function useEventDetail(id: string, active: boolean) {
	const queryClient = useQueryClient();
	const eventQuery = useQuery({
		queryKey: eventKeys.detail(id),
		queryFn: () => getEvent(id),
		enabled: Boolean(id),
	});
	const event = eventQuery.data;
	const cameraId = event?.cameraId ?? undefined;
	const cameraQuery = useQuery({
		queryKey: cameraKeys.detail(cameraId ?? ""),
		queryFn: () => getCamera(cameraId ?? ""),
		enabled: Boolean(cameraId),
	});
	const playbackQuery = useQuery({
		queryKey: cameraKeys.playback(cameraId ?? "", event?.startedAt ?? ""),
		queryFn: () => getPlayback(cameraId ?? "", event?.startedAt ?? ""),
		enabled: active && Boolean(cameraId && event?.startedAt),
		staleTime: 45_000,
	});
	const playbackKey = cameraKeys.playback(cameraId ?? "", event?.startedAt ?? "");
	const continuePlaybackMutation = useMutation({
		mutationFn: ({
			cameraId,
			nextStart,
		}: {
			cameraId: string;
			nextStart: string;
			playbackKey: ReturnType<typeof cameraKeys.playback>;
		}) => getPlayback(cameraId, nextStart),
		onSuccess: (nextSession, variables) => {
			queryClient.setQueryData(variables.playbackKey, nextSession);
		},
	});
	const acknowledgeMutation = useMutation({
		mutationFn: () => acknowledgeEvent(id),
		onSuccess: async (updated) => {
			queryClient.setQueryData(eventKeys.detail(id), updated);
			await queryClient.invalidateQueries({ queryKey: eventKeys.all });
		},
	});
	const playback = resolvePlayback(playbackQuery.data);
	const continuePlayback = () => {
		const session = playbackQuery.data;
		if (!session?.nextRecordingStart || continuePlaybackMutation.isPending) {
			return;
		}
		continuePlaybackMutation.mutate({
			cameraId: session.cameraId,
			nextStart: session.nextRecordingStart,
			playbackKey,
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
		event,
		camera: cameraQuery.data,
		playbackUrl: playback.kind === "ready" ? playback.url : undefined,
		playbackStartOffsetSec: playback.kind === "ready" ? playback.startOffsetSec : 0,
		eventPending: eventQuery.isPending,
		eventError: eventQuery.error,
		playbackPending: playbackQuery.isPending || continuePlaybackMutation.isPending,
		playbackError:
			continuePlaybackMutation.error ??
			playbackQuery.error ??
			(playback.kind === "error" ? playback.error : null),
		acknowledging: acknowledgeMutation.isPending,
		acknowledgeError: acknowledgeMutation.error,
		retryEvent: eventQuery.refetch,
		retryPlayback,
		continuePlayback,
		acknowledge: acknowledgeMutation.mutate,
	};
}
