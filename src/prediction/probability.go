package prediction

import (
	"fmt"
	"strings"

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
// 3. 確率を合計して返す
func GetProbability(data []string, time string, weeks int) (float64, error) {
	// 時刻文字列を分に変換
	dataMinutes := make([]int, 0, len(data))
	for _, d := range data {
		minutes, err := lib.TimeToMinutes(d)
		if err != nil {
			return 0, err
		}
		dataMinutes = append(dataMinutes, minutes)
	}

	timeMinutes, err := lib.TimeToMinutes(time)
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
	totalProbability := 0.0

	for _, c := range clusters {
		// クラスタ内のデータポイントが1つの場合
		if len(c.Data) == 1 {
			if float64(timeMinutes) >= c.Data[0] {
				totalProbability += 1.0 / float64(weeks)
			}
			continue
		}

		// 2-1. クラスタ平均
		loc := stat.Mean(c.Data, nil)

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

// GetProbabilityByUniqueDate 来訪確率を計算する（日付重複を排除）
// 同じ日に複数のログがある場合、最初の時刻のみを使用する
// data: "2006-01-02 15:04"形式の日付時刻文字列スライス
// time: "HH:MM"形式の時刻文字列
// weeks: 週数
func GetProbabilityByUniqueDate(data []string, time string, weeks int) (float64, error) {
	// 日付ごとに最初の時刻のみを保持
	dateToTime := make(map[string]string)
	for _, d := range data {
		parts := strings.SplitN(d, " ", 2)
		if len(parts) != 2 {
			return 0, fmt.Errorf("invalid datetime format: %s", d)
		}
		date := parts[0]
		timeStr := parts[1]

		// 同じ日付がまだ登録されていない場合のみ追加
		if _, exists := dateToTime[date]; !exists {
			dateToTime[date] = timeStr
		}
	}

	// 重複排除後の時刻リストを作成
	uniqueTimes := make([]string, 0, len(dateToTime))
	for _, t := range dateToTime {
		uniqueTimes = append(uniqueTimes, t)
	}

	// 既存のGetProbability関数を使用して確率を計算
	return GetProbability(uniqueTimes, time, weeks)
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
