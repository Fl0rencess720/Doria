package utils

func ConvertFloat64ToFloat32(input [][]float64) [][]float32 {
	result := make([][]float32, len(input))

	for i, row := range input {
		result[i] = make([]float32, len(row))
		for j, val := range row {
			result[i][j] = float32(val)
		}
	}

	return result
}

func ConvertUintToInt64(input []uint) []int64 {
	result := make([]int64, len(input))

	for i, val := range input {
		result[i] = int64(val)
	}

	return result
}
