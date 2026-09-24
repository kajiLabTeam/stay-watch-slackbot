package prediction

import (
	"fmt"
	"math"

	"github.com/kajiLabTeam/stay-watch-slackbot/lib"
	"gonum.org/v1/gonum/stat"
	"gonum.org/v1/gonum/stat/distuv"
)

// GetProbability 来訪確率を計算する
// Python版と同等のロジック:
// 1. GMMを使用してデータをクラスタリング
// 2. 各クラスタの確率を計算
//   - クラスタ中心を平均として使用
//   - クラスタの標準偏差を計算
//   - 正規分布のCDFを使用して確率を計算
//   - 確率を（クラスタサイズ / 週数）で重み付け
//
// 3. 確率を合計して返す
//
// 同じデータで複数の時刻を評価する場合は、クラスタリングを繰り返さずに済む
// NewProbabilityModel() + ProbabilityModel.Probability() を使うこと。
func GetProbability(data []string, time string, weeks int) (float64, error) {
	model, err := NewProbabilityModel(data)
	if err != nil {
		return 0, err
	}
	return model.Probability(time, weeks)
}

// calcClusterProbability 単一クラスタの来訪確率を計算する
func calcClusterProbability(c ClusteringResult, timeMinutes int, weeks int) float64 {
	// クラスタ内のデータポイントが1つの場合
	if len(c.Data) == 1 {
		if float64(timeMinutes) >= c.Data[0] {
			return 1.0 / float64(weeks)
		}
		return 0
	}

	loc := stat.Mean(c.Data, nil)
	scale := stat.StdDev(c.Data, nil)

	// scale = 0（クラスタ内のすべてのデータが同じ）
	if scale == 0 {
		if c.Data[0] == loc && float64(timeMinutes) >= loc {
			return 1.0 * (float64(len(c.Data)) / float64(weeks))
		}
		return 0
	}

	// 正規分布のCDFを使用して確率を計算し、重み付け
	normDist := distuv.Normal{
		Mu:    loc,
		Sigma: scale,
	}
	cdf := normDist.CDF(float64(timeMinutes))
	result := cdf * (float64(len(c.Data)) / float64(weeks))
	if math.IsNaN(result) || math.IsInf(result, 0) {
		return 0
	}
	return result
}

// GetProbabilityByUniqueDate 来訪確率を計算する（日付重複を排除）
// 同じ日に複数のログがある場合、最初の時刻のみを使用する
// data: "2006-01-02 15:04"形式の日付時刻文字列スライス
// time: "HH:MM"形式の時刻文字列
// weeks: 週数
func GetProbabilityByUniqueDate(data []string, time string, weeks int) (float64, error) {
	model, err := NewProbabilityModelByUniqueDate(data)
	if err != nil {
		return 0, err
	}
	return model.Probability(time, weeks)
}

// GetProbabilityFromDatetimes 来訪確率を計算する（日付重複排除なし）
// 同じ日に複数の活動がある場合もすべての時刻を使用する
// data: "2006-01-02 15:04"形式の日付時刻文字列スライス
// time: "HH:MM"形式の時刻文字列
// weeks: 週数
func GetProbabilityFromDatetimes(data []string, time string, weeks int) (float64, error) {
	model, err := NewProbabilityModelFromDatetimes(data)
	if err != nil {
		return 0, err
	}
	return model.Probability(time, weeks)
}

// GetMostLikelyTime 活動の最も可能性の高い時間を見つける
// 各クラスタをガウス分布とした場合の頂点（中心）の時刻に重みを付与し、その合計を返す
func GetMostLikelyTime(data []string, weeks int) (int, error) {
	// 時刻文字列を分に変換
	dataMinutes := make([]int, 0, len(data))
	for _, d := range data {
		minutes, err := lib.TimeToMinutes(d)
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
