package prediction

import (
	"fmt"
	"strconv"
	"strings"

	"gonum.org/v1/gonum/stat"
	"gonum.org/v1/gonum/stat/distuv"
)

// TimeToMinutes "HH:MM"形式の時刻を分に変換する
func TimeToMinutes(timeStr string) (int, error) {
	parts := strings.Split(timeStr, ":")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid time format: %s", timeStr)
	}

	hours, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, fmt.Errorf("invalid hours: %s", parts[0])
	}

	minutes, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, fmt.Errorf("invalid minutes: %s", parts[1])
	}

	return hours*60 + minutes, nil
}

// MinutesToTime 分を"HH:MM"形式に変換する
func MinutesToTime(minutes int) string {
	hours := minutes / 60
	mins := minutes % 60
	return fmt.Sprintf("%02d:%02d", hours, mins)
}

// GetProbability 来訪確率を計算する
// Python版と同等のロジック:
// 1. GMMを使用してデータをクラスタリング
// 2. 各クラスタの確率を計算
//   - クラスタ中心を平均として使用
//   - クラスタの標準偏差を計算
//   - 正規分布のCDFを使用して確率を計算
//   - 確率を（クラスタサイズ / 週数）で重み付け
// 3. 確率を合計して返す
func GetProbability(data []string, time string, weeks int) (float64, error) {
	// 時刻文字列を分に変換
	dataMinutes := make([]int, 0, len(data))
	for _, d := range data {
		minutes, err := TimeToMinutes(d)
		if err != nil {
			return 0, err
		}
		dataMinutes = append(dataMinutes, minutes)
	}

	timeMinutes, err := TimeToMinutes(time)
	if err != nil {
		return 0, err
	}

	// データポイントが1つの場合の特別処理
	if len(dataMinutes) == 1 {
		if timeMinutes >= dataMinutes[0] {
			return 1.0 / float64(weeks), nil
		}
		return 0, nil
	}

	// 1. GMMを使用してデータをクラスタリング
	clusters := Clustering(dataMinutes)

	// 2. 各クラスタの確率を計算
	var totalProbability float64

	for _, c := range clusters {
		// クラスタ内のデータポイントが1つの場合
		if len(c.Data) == 1 {
			if float64(timeMinutes) >= c.Data[0] {
				totalProbability += 1.0 / float64(weeks)
			}
			continue
		}

		// 2-1. クラスタ中心（平均）
		loc := c.Center

		// 2-2. クラスタの標準偏差を計算
		scale := stat.StdDev(c.Data, nil)

		// scale = 0（クラスタ内のすべてのデータが同じ）
		if scale == 0 {
			if c.Data[0] == loc && float64(timeMinutes) >= loc {
				totalProbability += 1.0 * (float64(len(c.Data)) / float64(weeks))
			}
			continue
		}

		// 2-3. 正規分布のCDFを使用して確率を計算
		normDist := distuv.Normal{
			Mu:    loc,
			Sigma: scale,
		}
		cdf := normDist.CDF(float64(timeMinutes))

		// 2-4. 重み付けされた確率
		weightedProb := cdf * (float64(len(c.Data)) / float64(weeks))
		totalProbability += weightedProb
	}

	return totalProbability, nil
}

// GetMostLikelyTime 活動の最も可能性の高い時間を見つける
// 各クラスタをガウス分布とした場合の頂点（中心）の時刻に重みを付与し、その合計を返す
func GetMostLikelyTime(data []string, weeks int) (int, error) {
	// 時刻文字列を分に変換
	dataMinutes := make([]int, 0, len(data))
	for _, d := range data {
		minutes, err := TimeToMinutes(d)
		if err != nil {
			return 0, err
		}
		dataMinutes = append(dataMinutes, minutes)
	}

	if len(dataMinutes) == 0 {
		return 0, fmt.Errorf("no data provided")
	}

	if len(dataMinutes) == 1 {
		return dataMinutes[0], nil
	}

	// GMMを使用してデータをクラスタリング
	clusters := Clustering(dataMinutes)

	// 各クラスタの中心に重みを付与して合計
	weightedSum := 0.0
	totalWeight := 0.0

	for i := range clusters {
		weight := float64(len(clusters[i].Data)) / float64(len(dataMinutes))
		weightedSum += clusters[i].Center * weight
		totalWeight += weight
	}

	if totalWeight == 0 {
		return 0, fmt.Errorf("clustering failed")
	}

	// 重み付き平均を返す
	return int(weightedSum / totalWeight), nil
}
