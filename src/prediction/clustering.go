// Package prediction provides statistical prediction and clustering algorithms.
package prediction

import (
	"math"
	"math/rand"

	"gonum.org/v1/gonum/stat"
	"gonum.org/v1/gonum/stat/distuv"
)

// ClusteringResult はクラスタリング結果を表す
type ClusteringResult struct {
	Data   []float64 `json:"data"`
	Center float64   `json:"center"`
}

// GaussianMixture は1次元データに対するガウス混合モデル（GMM）を実装する
type GaussianMixture struct {
	NComponents int
	Means       []float64
	Variances   []float64
	Weights     []float64
	MaxIter     int
	Tolerance   float64
}

// NewGaussianMixture 新しいGMMを作成する
func NewGaussianMixture(nComponents int) *GaussianMixture {
	return &GaussianMixture{
		NComponents: nComponents,
		MaxIter:     100,
		Tolerance:   1e-6,
	}
}

// Fit EMアルゴリズムを使用してGMMを学習する
func (gmm *GaussianMixture) Fit(data []float64) {
	n := len(data)
	k := gmm.NComponents

	gmm.initializeKMeansPlusPlus(data)
	gmm.initializeVariancesAndWeights(data)

	responsibilities := make([][]float64, n)
	for i := range responsibilities {
		responsibilities[i] = make([]float64, k)
	}
	prevLogLikelihood := math.Inf(-1)

	for iter := 0; iter < gmm.MaxIter; iter++ {
		gmm.eStep(data, responsibilities)
		gmm.mStep(data, responsibilities)

		logLikelihood := gmm.logLikelihood(data)
		if math.Abs(logLikelihood-prevLogLikelihood) < gmm.Tolerance {
			break
		}
		prevLogLikelihood = logLikelihood
	}
}

// initializeVariancesAndWeights 分散と重みを初期化する
func (gmm *GaussianMixture) initializeVariancesAndWeights(data []float64) {
	k := gmm.NComponents
	gmm.Variances = make([]float64, k)
	gmm.Weights = make([]float64, k)
	dataVar := stat.Variance(data, nil)
	if dataVar < 1e-6 {
		dataVar = 1.0
	}
	for i := 0; i < k; i++ {
		gmm.Variances[i] = dataVar
		gmm.Weights[i] = 1.0 / float64(k)
	}
}

// eStep Eステップ: 各データポイントの負担率を計算する
func (gmm *GaussianMixture) eStep(data []float64, responsibilities [][]float64) {
	k := gmm.NComponents
	for i, x := range data {
		probs := make([]float64, k)
		var total float64
		for j := 0; j < k; j++ {
			normDist := distuv.Normal{
				Mu:    gmm.Means[j],
				Sigma: math.Sqrt(gmm.Variances[j]),
			}
			probs[j] = gmm.Weights[j] * normDist.Prob(x)
			total += probs[j]
		}
		for j := 0; j < k; j++ {
			if total > 0 {
				responsibilities[i][j] = probs[j] / total
			} else {
				responsibilities[i][j] = 1.0 / float64(k)
			}
		}
	}
}

// mStep Mステップ: パラメータを更新する
func (gmm *GaussianMixture) mStep(data []float64, responsibilities [][]float64) {
	n := len(data)
	for j := 0; j < gmm.NComponents; j++ {
		gmm.updateComponent(j, data, responsibilities, n)
	}
}

// updateComponent 1コンポーネント分のパラメータを更新する
func (gmm *GaussianMixture) updateComponent(j int, data []float64, responsibilities [][]float64, n int) {
	var nk, meanSum float64
	for i := 0; i < n; i++ {
		r := responsibilities[i][j]
		nk += r
		meanSum += r * data[i]
	}
	if nk <= 1e-10 {
		return
	}
	gmm.Means[j] = meanSum / nk
	var varSum float64
	for i := 0; i < n; i++ {
		r := responsibilities[i][j]
		diff := data[i] - gmm.Means[j]
		varSum += r * diff * diff
	}
	gmm.Variances[j] = math.Max(varSum/nk, 1e-6)
	gmm.Weights[j] = nk / float64(n)
}

// initializeKMeansPlusPlus K-means++を使用して初期中心を選択する
func (gmm *GaussianMixture) initializeKMeansPlusPlus(data []float64) {
	n := len(data)
	k := gmm.NComponents
	gmm.Means = make([]float64, k)

	if n == 0 {
		return
	}

	gmm.Means[0] = data[rand.Intn(n)]

	for i := 1; i < k; i++ {
		distances, totalDist := calcMinDistances(data, gmm.Means[:i])
		gmm.Means[i] = selectByDistance(data, distances, totalDist)
	}
}

// calcMinDistances 各データポイントから最も近い中心までの距離の二乗を計算する
func calcMinDistances(data []float64, centers []float64) ([]float64, float64) {
	distances := make([]float64, len(data))
	var totalDist float64
	for j, x := range data {
		minDist := math.Inf(1)
		for _, c := range centers {
			dist := (x - c) * (x - c)
			if dist < minDist {
				minDist = dist
			}
		}
		distances[j] = minDist
		totalDist += minDist
	}
	return distances, totalDist
}

// selectByDistance 距離の二乗に比例した確率でデータポイントを選択する
func selectByDistance(data []float64, distances []float64, totalDist float64) float64 {
	if totalDist <= 0 {
		return data[rand.Intn(len(data))]
	}
	r := rand.Float64() * totalDist
	var cumSum float64
	for j, d := range distances {
		cumSum += d
		if cumSum >= r {
			return data[j]
		}
	}
	return data[len(data)-1]
}

// logLikelihood 対数尤度を計算する
func (gmm *GaussianMixture) logLikelihood(data []float64) float64 {
	var ll float64
	for _, x := range data {
		var prob float64
		for j := 0; j < gmm.NComponents; j++ {
			normDist := distuv.Normal{
				Mu:    gmm.Means[j],
				Sigma: math.Sqrt(gmm.Variances[j]),
			}
			prob += gmm.Weights[j] * normDist.Prob(x)
		}
		if prob > 0 {
			ll += math.Log(prob)
		}
	}
	return ll
}

// Predict 各データポイントのクラスタを予測する
func (gmm *GaussianMixture) Predict(data []float64) []int {
	labels := make([]int, len(data))
	for i, x := range data {
		maxProb := 0.0
		maxLabel := 0
		for j := 0; j < gmm.NComponents; j++ {
			normDist := distuv.Normal{
				Mu:    gmm.Means[j],
				Sigma: math.Sqrt(gmm.Variances[j]),
			}
			prob := gmm.Weights[j] * normDist.Prob(x)
			if prob > maxProb {
				maxProb = prob
				maxLabel = j
			}
		}
		labels[i] = maxLabel
	}
	return labels
}

// BIC ベイズ情報量基準（BIC）を計算する
func (gmm *GaussianMixture) BIC(data []float64) float64 {
	n := float64(len(data))
	k := float64(gmm.NComponents)
	// パラメータ数: k個の平均 + k個の分散 + (k-1)個の重み
	numParams := 3*k - 1
	ll := gmm.logLikelihood(data)
	return -2*ll + numParams*math.Log(n)
}

// Clustering データをクラスタリングする（元のPython実装と同等）
func Clustering(data []int) []ClusteringResult {
	// intをfloat64に変換
	floatData := make([]float64, len(data))
	for i, v := range data {
		floatData[i] = float64(v)
	}

	// BICを使用して最適なクラスタ数を見つける
	maxClusters := 4
	if len(floatData) < maxClusters {
		maxClusters = len(floatData)
	}

	var bicValues []float64
	var gmms []*GaussianMixture

	for nClusters := 1; nClusters <= maxClusters; nClusters++ {
		gmm := NewGaussianMixture(nClusters)
		gmm.Fit(floatData)
		bicValues = append(bicValues, gmm.BIC(floatData))
		gmms = append(gmms, gmm)
	}

	// 最小BICのクラスタ数を選択
	optimalIdx := 0
	minBIC := bicValues[0]
	for i, bic := range bicValues {
		if bic < minBIC {
			minBIC = bic
			optimalIdx = i
		}
	}

	optimalGMM := gmms[optimalIdx]
	return makeResultsList(floatData, optimalGMM)
}

// makeResultsList クラスタリング結果をリストに変換する
func makeResultsList(data []float64, gmm *GaussianMixture) []ClusteringResult {
	labels := gmm.Predict(data)
	nClusters := gmm.NComponents

	// クラスタごとにデータを分割
	clusters := make([][]float64, nClusters)
	for i := 0; i < nClusters; i++ {
		clusters[i] = []float64{}
	}

	for i, label := range labels {
		clusters[label] = append(clusters[label], data[i])
	}

	// 結果を構築
	results := make([]ClusteringResult, nClusters)
	for i := 0; i < nClusters; i++ {
		results[i] = ClusteringResult{
			Data:   clusters[i],
			Center: gmm.Means[i],
		}
	}

	return results
}
