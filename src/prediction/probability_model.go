package prediction

import (
	"fmt"
	"strings"

	"github.com/kajiLabTeam/stay-watch-slackbot/lib"
)

// ProbabilityModel はクラスタリング済みの時刻分布を保持する。
//
// Clustering() は 1 回あたり最大 40 回の GMM フィットを行う重い処理なので、
// 同じデータに対して複数の時刻の確率を評価したい場合はモデルを 1 度だけ構築して
// Probability() を繰り返し呼ぶこと。
type ProbabilityModel struct {
	clusters []ClusteringResult
	// singleData はデータ点が 1 つしかない場合の値（分）。isSingle が true のときのみ有効
	singleData int
	isSingle   bool
	isEmpty    bool
}

// NewProbabilityModel は "HH:MM" 形式の時刻リストからモデルを構築する（クラスタリングは 1 回のみ）
func NewProbabilityModel(data []string) (*ProbabilityModel, error) {
	dataMinutes := make([]int, 0, len(data))
	for _, d := range data {
		minutes, err := lib.TimeToMinutes(d)
		if err != nil {
			return nil, err
		}
		dataMinutes = append(dataMinutes, minutes)
	}
	return newProbabilityModelFromMinutes(dataMinutes), nil
}

// NewProbabilityModelFromDatetimes は "2006-01-02 15:04" 形式の日時リストから
// 時刻部分のみを取り出してモデルを構築する（日付重複は排除しない）
func NewProbabilityModelFromDatetimes(data []string) (*ProbabilityModel, error) {
	times, err := extractTimes(data)
	if err != nil {
		return nil, err
	}
	return NewProbabilityModel(times)
}

// NewProbabilityModelByUniqueDate は "2006-01-02 15:04" 形式の日時リストから
// 日付ごとに最初の時刻のみを採用してモデルを構築する
func NewProbabilityModelByUniqueDate(data []string) (*ProbabilityModel, error) {
	times, err := extractUniqueDateTimes(data)
	if err != nil {
		return nil, err
	}
	return NewProbabilityModel(times)
}

func newProbabilityModelFromMinutes(dataMinutes []int) *ProbabilityModel {
	switch len(dataMinutes) {
	case 0:
		return &ProbabilityModel{isEmpty: true}
	case 1:
		return &ProbabilityModel{singleData: dataMinutes[0], isSingle: true}
	default:
		return &ProbabilityModel{clusters: Clustering(dataMinutes)}
	}
}

// Probability は指定時刻（"HH:MM"）における来訪確率を返す
func (m *ProbabilityModel) Probability(time string, weeks int) (float64, error) {
	timeMinutes, err := lib.TimeToMinutes(time)
	if err != nil {
		return 0, err
	}
	return m.probabilityAtMinutes(timeMinutes, weeks), nil
}

// Clusters はモデルが保持するクラスタを返す（データ点が 0 個 / 1 個の場合は nil）
func (m *ProbabilityModel) Clusters() []ClusteringResult {
	return m.clusters
}

func (m *ProbabilityModel) probabilityAtMinutes(timeMinutes int, weeks int) float64 {
	if m.isEmpty {
		return 0
	}

	// データポイントが1つの場合の特別処理
	if m.isSingle {
		if timeMinutes >= m.singleData {
			return 1.0 / float64(weeks)
		}
		return 0
	}

	totalProbability := 0.0
	for _, c := range m.clusters {
		totalProbability += calcClusterProbability(c, timeMinutes, weeks)
	}
	return totalProbability
}

// TotalWeight は時刻によらない全体の確率質量（timeMinutes→∞ でのCDFの極限）を返す。
// 深夜0時をまたぐ時間帯（例: 23:30〜翌00:30）の確率を
// 「全体 - CDF(23:30) + CDF(00:30)」として求める際に使う。
func (m *ProbabilityModel) TotalWeight(weeks int) float64 {
	if m.isEmpty {
		return 0
	}
	if m.isSingle {
		return 1.0 / float64(weeks)
	}

	total := 0.0
	for _, c := range m.clusters {
		total += float64(len(c.Data)) / float64(weeks)
	}
	return total
}

// extractTimes は "2006-01-02 15:04" 形式から時刻部分のみを抽出する（重複排除しない）
func extractTimes(data []string) ([]string, error) {
	times := make([]string, 0, len(data))
	for _, d := range data {
		parts := strings.SplitN(d, " ", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid datetime format: %s", d)
		}
		times = append(times, parts[1])
	}
	return times, nil
}

// extractUniqueDateTimes は日付ごとに最初の時刻のみを保持して時刻リストを返す
func extractUniqueDateTimes(data []string) ([]string, error) {
	dateToTime := make(map[string]string)
	for _, d := range data {
		parts := strings.SplitN(d, " ", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid datetime format: %s", d)
		}
		date := parts[0]
		timeStr := parts[1]

		// 同じ日付がまだ登録されていない場合のみ追加
		if _, exists := dateToTime[date]; !exists {
			dateToTime[date] = timeStr
		}
	}

	uniqueTimes := make([]string, 0, len(dateToTime))
	for _, t := range dateToTime {
		uniqueTimes = append(uniqueTimes, t)
	}
	return uniqueTimes, nil
}
