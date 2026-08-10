export type CameraGridLayout = {
	columns: number;
	rows: number;
};

const CAMERA_ASPECT_RATIO = 16 / 9;
const GRID_GAP_PX = 1;

/**
 * Fit camera tiles to the available rectangle while keeping each cell as close
 * as possible to the camera's native aspect ratio.
 */
export function calculateCameraGridLayout(
	cameraCount: number,
	containerWidth: number,
	containerHeight: number,
): CameraGridLayout {
	const count = Math.max(1, Math.floor(cameraCount));
	const width = Math.max(0, containerWidth);
	const height = Math.max(0, containerHeight);

	if (width === 0 || height === 0) {
		const columns = Math.ceil(Math.sqrt(count));
		return { columns, rows: Math.ceil(count / columns) };
	}

	let bestLayout: CameraGridLayout = { columns: 1, rows: count };
	let bestVisibleArea = -1;
	let bestEmptyCells = Number.POSITIVE_INFINITY;

	for (let columns = 1; columns <= count; columns += 1) {
		const rows = Math.ceil(count / columns);
		const cellWidth = Math.max(0, width - GRID_GAP_PX * (columns - 1)) / columns;
		const cellHeight = Math.max(0, height - GRID_GAP_PX * (rows - 1)) / rows;
		const fittedWidth = Math.min(cellWidth, cellHeight * CAMERA_ASPECT_RATIO);
		const fittedHeight = fittedWidth / CAMERA_ASPECT_RATIO;
		const visibleArea = fittedWidth * fittedHeight * count;
		const emptyCells = columns * rows - count;

		if (
			visibleArea > bestVisibleArea ||
			(visibleArea === bestVisibleArea && emptyCells < bestEmptyCells) ||
			(visibleArea === bestVisibleArea &&
				emptyCells === bestEmptyCells &&
				columns > bestLayout.columns)
		) {
			bestLayout = { columns, rows };
			bestVisibleArea = visibleArea;
			bestEmptyCells = emptyCells;
		}
	}

	return bestLayout;
}
